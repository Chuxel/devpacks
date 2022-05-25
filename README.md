# Chuxel's development buildpack sandbox

*Everything here will change - a lot. Don't depend on any of it for anything you're doing. Good for demos, that's it.*

## Development setup

1. Open this repository in GitHub Codespaces or Remote - Containers using VS Code
2. File > Open workspace and select `workspace.code-workspace`

## What is here

This repo demonstrates the value of https://github.com/devcontainers/spec/issues/18 (and https://github.com/devcontainers/spec/issues/2) by integrating development container metadata into [Cloud Native Buildpacks](https://buildpacks.io/). 

Repo contents:

1. **[devpacks](devpacks)**: Code to create a set of buildpacks (e.g., see `devpacks/internal/buildpacks`). Build using `make`.
2. **[images](images)**: A Dockerfile and related content to generate a set of [stack images](https://buildpacks.io/docs/operator-guide/create-a-stack/).
3. **[builders](builders)**: Config needed to create two [builders](https://buildpacks.io/docs/operator-guide/create-a-builder/) that include (1) and (2).
4. **[devpacks/cmd/devcontainer-extractor](devpacks/cmd/devcontainer-extractor)** A utility to generate a `devcontainer.json` file from the output of one of the builders

The resulting `ghcr.io/chuxel/devpacks/builder-prod-full` builder behaves like a typical buildpack, while `ghcr.io/chuxel/devpacks/builder-devcontainer-full` instead focuses on creating a dev container image that is similar to production.

These builders can be used with the [`pack` CLI](https://buildpacks.io/docs/tools/pack/) or other CNB v3 compliant tools. 

An extractor utility (see [releases](https://github.com/Chuxel/devpacks/releases)) can then be used to generate `devcontainer.json` or (merge contents with an existing file) for images output by the `builder-devcontainer-full` builder. The resulting devcontainer.json will contain any needed tooling or runtime metadata - including relevant lifecycle script properties like `postCreateCommand` (e.g. `npm install`). It will also add a reference to the specified image to the json file as long as no Dockerfile or Docker Compose file is referenced - as long as no existing Dockerfile or Docker Compose file is referenced in the local `devcontainer.json` file. This allows developers to add additional metadata and settings to the devcontainer.json file not supported by the buildpack (including dev container features).

Right now it supports basic Node.js apps with a `start` script in `package.json`, and basic Python 3 applications that use `pip` (and thus have a `requirements.txt` file) and include a [`Procfile`](https://devcenter.heroku.com/articles/procfile) with a `web` entry to specify the startup command.

### Usage
This:
```
$ pack build devcontainer_image --trust-builder --builder ghcr.io/chuxel/devpacks/builder-devcontainer-full
$ devcontainer-extractor devcontainer_image
```
...will use the contents of the current folder to create an image called `devcontainer_image` and a related `devcontainer.json.merged` file. Removing `.merged` from the filename would allow you to test it in the VS Code Remote - Containers extension.

And this:
```
$ pack build devcontainer_image --trust-builder --builder ghcr.io/chuxel/devpacks/builder-prod-full
```
... will generate a production version of the image with the application inside it instead.

## How it works

The Buildpacks are written in Go and take advantage of [libcnb](https://pkg.go.dev/github.com/buildpacks/libcnb) to simplify interop with the buildpack spec. Here's how the different pieces in this repository work together:

1. The base images for the `prod` and `devcontainer` builders are constructed using a multi-stage Dockerfile with later dev container stages adding more base content - they are therefore a superset of the prod images. The main reason for two sets of images is that the devcontainer image includes a number of utilities like htop, ps, zsh, etc. Installing these OS utilities requires root access, which is not allowed today. (However, a [proposal](https://github.com/buildpacks/spec/pull/307) could help with this long term so that these become part of a buildpack instead.)

2. The devcontainer base images include also include updates to rc/profile files to handle the fact that any Buildpack injected environment variables are not available to "docker exec" (or other CLI). Only the sub-processes of the entrypoint get the environmant variables buildpacks add by default, and interacting with the dev container is typically done using commands like exec. See [launcher-hack.sh](images/scripts/launcher-hack.sh) for details. This is **critical** to ensuring things work in the dev container context. Here again, this is in the image since buildpacks cannot modify contents outside of their specific layer folders either.

3. A "build mode" allows for dual-purpose buildpacks that can either alter behaviors or simply not be detected when in one mode or the other. For example, a `pythonutils` buildpack that injects tools like `pylint` only executes in devcontainer mode, while others like `nodejs` or `cpython` execute in both modes. A file placed in a known location in the Dockerfile fom step 1 indicates the mode for the build - though you can also sett this mode using the `BP_DCNB_BUILD_MODE` environment variable.

4. The buildpacks can optionally place a `devcontainer.json` snippet file in their layers and add the path to it in a common `FINALIZE_JSON_SEARCH_PATH` build-time environment variable for the layer. These devcontainer.json files can include tooling settings, runtime settings like adding capabilities (e.g. ptrace or privileged), or even lifecycle commands.

5. A `finalize` buildpack merges all devcontainer.json snippets from the `FINALIZE_JSON_SEARCH_PATH` and adds a `dev.containers.json` label on the image with their contents.

6. The `finalize` buildpack also removes the source code since this is expected to be mounted into the container when the image is used. As a result, `finalize` will fail detection in production mode and is last in the ordering in the devcontainer builder. It also overrides the default launch step to one that sleeps infinitely to prevent it from shutting down (though this last part is technically optional).

6. Since the dev container CLI (and related products like VS Code Remote - Containers and Codespaces) do not support consuming dev container metadata from an image label yet, an extractor utility (see `devpacks/cmd/devcontainer-extractor`) extracts metadata from the image and merges it with a local devcontainer.json file if one is found (creating `devcontainer.json.merged`). 

7. The extractor also translates properties proposed in https://github.com/devcontainers/spec/issues/2 to `runArgs` equivalents as a stop gap.

That's the scoop!

## Problems not solved

1. Buildpacks cannot install anything that requires root access or modify contents outside of the specified layer folder (which isn't a Docker layer in and of itself). There's a [proposal](https://github.com/buildpacks/spec/pull/307) that could enable it.