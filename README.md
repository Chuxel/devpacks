# Chuxel's development buildpack sandbox

*Everything here will change - a lot. Don't depend on any of it for anything you're doing. Good for demos, that's it.*

## Development setup

1. Open this repository in GitHub Codespaces or Remote - Containers using VS Code
2. File > Open workspace and select `workspace.code-workspace`

## What is here

This repo demonstrates the value of https://github.com/devcontainers/spec/issues/18 (and https://github.com/devcontainers/spec/issues/2) by integrating development container metadata into [Cloud Native Buildpacks](https://buildpacks.io/). 

Repo contents:

1. A set of buildpacks under `devpacks` (with core logic being in `devpacks/internal/buildpacks`)
2. A set of [stack images](https://buildpacks.io/docs/operator-guide/create-a-stack/) generated via a `Dockerfile` under `images`
3. Two [builders](https://buildpacks.io/docs/operator-guide/create-a-builder/) that include (1) and (2) under `builders`
4. A utility to generate a `devcontainer.json` file from the output of one of the builders (`devpacks/cmd/devcontainer-extractor`)

The `ghcr.io/chuxel/devpacks/builder-prod-full` builder behaves like a typical buildpack, while `ghcr.io/chuxel/devpacks/builder-devcontainer-full` instead focuses on creating a dev container image that is similar to production.

These builders can be used with the [`pack` CLI](https://buildpacks.io/docs/tools/pack/) or other CNB v3 compliant tools. 

An extractor utility (see [releases](https://github.com/Chuxel/devpacks/releases)) can then be used to generate `devcontainer.json` or (merge contents with an existing file) for images output by the `builder-devcontainer-full` builder. The resulting devcontainer.json will contain any needed tooling or runtime metadata - including relevant lifecycle script properties like `postCreateCommand` (e.g. `npm install`). It will also add a reference to the specified image to the json file as long as no Dockerfile or Docker Compose file is referenced - as long as no existing Dockerfile or Docker Compose file is referenced in the local `devcontainer.json` file. This allows developers to add additional metadata and settings to the devcontainer.json file not supported by the buildpack (including dev container features).

Right now it supports basic Node.js apps with a `start` entry in `package.json`, and basic Python applications that use `pip` (and thus have a `requirements.txt` file) and include a [`Procfile`](https://devcenter.heroku.com/articles/procfile) with a `web` entry to specify the startup command.

e.g. This:
```
$ pack build devcontainer_image --trust-builder --builder ghcr.io/chuxel/devpacks/builder-devcontainer-full
$ devcontainer-extractor devcontainer_image
```

Will create an image called `devcontainer_image` and a and output a related `devcontainer.json.merged` file using the based on the contents of the current folder.

## How it works

Buildpacks are written in Go and take advantage of [libcnb](https://pkg.go.dev/github.com/buildpacks/libcnb) to simplify interop with the buildpack spec.

1. The base images for the `prod` and `devcontainer` builders are constructed using a multi-stage Dockerfile with later dev container stages adding more base content, but otherwise the same structure - which ensures consistency.
2. A "build mode" allows for dual-purpose buildpacks that can either alter behaviors or simply not be detected when in one mode or the other. For example, a `pythonutils` buildpack that injects tools like `pylint` only executes in devcontainer mode, while others like `nodejs` or `cpython` execute in both modes. A file in the build image can be used determine the mode, but they also support passing in the mode as a `BP_DCNB_BUILD_MODE` environment variable.
3. Buildpacks like `npminstall` then add "requires" entries on runtime buildpacks like `nodejs` to ensure proper ordering. 
4. The buildpacks can optionally place a `devcontainer.json` snippet file in their layer and add the path to it in a common `FINALIZE_JSON_SEARCH_PATH` environment variable. This can include tooling settings, runtime settings like adding capabilities (e.g. ptrace or privileged), or even lifecycle commands.
4. A `finalize` buildpack merges all devcontainer.json snippets from the `FINALIZE_JSON_SEARCH_PATH` and adds a `dev.containers.json` label on the image with their contents.
5. The `finalize` buildpack also removes the source code since this is expected to be mounted as a volume or bind mounted when the image is used. As a result, `finalize` will fail detection in production mode and is the last in order in the devcontainer builder.
7. Since the dev container CLI (and related products like VS Code Remote - Containers and Codespaces) does not support dev container metadata in an image label, an extractor utility (see `devpacks/cmd/devcontainer-extractor`) extracts metadata from the image and merges it with a local devcontainer.json file if one is found (creating `devcontainer.json.merged`). It also translates properties proposed in https://github.com/devcontainers/spec/issues/2 to `runArgs` equivalents.
8. Finally, the devcontainer base images include code to handle the fact that any Buildpack injected environment variables are not available to "docker exec" (or other CLI) initiated processes since these do not execute from the entrypoint. It instead adds a line in several rc/profile files to detect the scenario and replace the process with a new one that was initiated via the launcher (see `images/scripts/launcher-hack.sh`). This is **critical** to ensuring things work in the dev container context.

That's the scoop!
