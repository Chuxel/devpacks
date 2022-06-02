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

Right now it supports basic Node.js apps with a `start` script in `package.json`, basic Python 3 applications that use `pip` (and thus have a `requirements.txt` file), building Go apps/services. The Go and Python apps need to include a [`Procfile`](https://devcenter.heroku.com/articles/procfile) with a `web` entry to specify the startup command.

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

### Buildpack information

Each buildpack in this repository demos something slightly different.

- `nodejs` - Demos installing Node.js, supporting different layering requirements, and adding devcontainer.json metadata.
    - `npminstall` - Demos a dual-mode buildpack that executes `npm install` in prod mode, but adds a `postCreateCommand` instead in devcontainer mode. Also "requires" `nodejs`.
    - `npmbuild` - Demos an optional, prod-only buildpack.
    - `npmstart` - Demos adding a prod-only launch config.
- `cpython` - Demos installing cpython using [GitHub Action's python-versions builds](https://github.com/actions/python-versions) and parsing its `versions-manifest.json` file to find the right download. (This model should extend to other Actions "versions" repositories. ) Also add devcontainer.json metadata.
    - `pipinstall` - Another dual-mode buildpack like `npminstall`, but for pip3.
    - `pythonutils` - Demonstrates a devcontainer mode only step to install tools like `pylint` that you would not want in prod mode.
- `goutils` - Demonstrates a devcontainer mode only buildpack that can depend on a [completely external Paketo buildpack](https://github.com/paketo-buildpacks/go-dist) to acquire Go itself, then install tools needed for developing. This buildpack also adds all needed devcontainer.json metadata for go development including setting the ptrace capability for debugging. The [full Go Paketo buildpack set](https://github.com/paketo-buildpacks/go) is then used in the prod builder.
- `procfile` - Demos creating a launch command while in production mode from a [`Procfile`](https://devcenter.heroku.com/articles/procfile).
- `finalize` - Demonstrates processing of accumulating devcontainer.json metadata from multiple buildpacks, placing it in a label, cleaning out the source tree, and adding a launch command that prevents the container from terminating by default.

## How it works

The Buildpacks are written in Go and take advantage of [libcnb](https://pkg.go.dev/github.com/buildpacks/libcnb) to simplify interop with the buildpack spec. Here's how the different pieces in this repository work together:

1. The base images for the `prod` and `devcontainer` builders are constructed using a multi-stage Dockerfile with later dev container stages adding more base content - they are therefore a superset of the prod images. The main reason for two sets of images is that the devcontainer image includes a number of utilities like htop, ps, zsh, etc. Installing these OS utilities requires root access, which is not allowed today. (However, a [proposal](https://github.com/buildpacks/spec/pull/307) could help with this long term so that these become part of a buildpack instead.)

2. The devcontainer base images also include updates to rc/profile files to handle the fact that any Buildpack injected environment variables are not available to "docker exec" (or other CLI). Only the sub-processes of the entrypoint get the environmant variables buildpacks add by default, and interacting with the dev container is typically done using commands like exec. See [launcher-hack.sh](images/scripts/launcher-hack.sh) for details. This is **critical** to ensuring things work in the dev container context. Here again, this is in the image since buildpacks cannot modify contents outside of their specific layer folders either.

1. A "build mode" allows for dual-purpose buildpacks that can either alter behaviors or simply not be detected when in one mode or the other. For example, a `pythonutils` buildpack that injects tools like `pylint` only executes in devcontainer mode, while others like `nodejs` or `cpython` execute in both modes. A file placed in a known location in the Dockerfile from step 1 indicates the mode for the build - though you can also set this mode using the `BP_DCNB_BUILD_MODE` environment variable.

2. Base buildpacks like `nodejs` and `cpython` are set up so that downstream buildpacks like `npminstall` and `pythoninstall` can add requirements that affect whether they are available in the build image, launch image (resulting output) or both through metadata. Setting `build=true` causes the `nodejs` or `python` to place the contents in the build image while `launch=true` causes it to be in the launch image. The union of all requirements is considered for the final result. As a result, these two buildpacks are set up to always "pass" detection, and instead only "provide" the capability for others to require in the event of a failed detection. Where this dynamic behavior is important for this use case is this enables a downstream buildpack to say something should be in the launch image, but not in the build image in one specific mode without having to alter the original. ([Paketo buildpacks use a similar trick](https://github.com/paketo-buildpacks/cpython#integration) so that runtimes can be used for tools in the build image even if they aren't in the output - but have the same benefits. See the `goutils` buildpack for a reuse example.)

4. The buildpacks can optionally place a `devcontainer.json` snippet file in their layers and add the path to it in a common `FINALIZE_JSON_SEARCH_PATH` build-time environment variable for the layer. These devcontainer.json files can include tooling settings, runtime settings like adding capabilities (e.g. ptrace or privileged), or even lifecycle commands.

5. A `finalize` buildpack merges all devcontainer.json snippets from the `FINALIZE_JSON_SEARCH_PATH` and adds a `dev.containers.json` label on the image with their contents.

6. The `finalize` buildpack also removes the source code since this is expected to be mounted into the container when the image is used. As a result, `finalize` will fail detection in production mode and is last in the ordering in the devcontainer builder. It also overrides the default launch step to one that sleeps infinitely to prevent it from shutting down (though this last part is technically optional).

6. Since the dev container CLI (and related products like VS Code Remote - Containers and Codespaces) do not support consuming dev container metadata from an image label yet, an extractor utility (see `devpacks/cmd/devcontainer-extractor`) extracts metadata from the image and merges it with a local devcontainer.json file if one is found (creating `devcontainer.json.merged`). 

7. The extractor also translates properties proposed in https://github.com/devcontainers/spec/issues/2 to `runArgs` equivalents as a stop gap.

That's the scoop!

## Notes and problems not solved

1. Buildpacks cannot install anything that requires root access or modify contents outside of the specified layer folder (which isn't a Docker layer in and of itself). There's a [image extension/Dockerfile proposal](https://github.com/buildpacks/spec/pull/307) that could enable it.

2. Furthermore, if the [image extension/Dockerfile proposal](https://github.com/buildpacks/spec/pull/307) goes through, a single "builder" could be used rather than separate ones for devcontainers and production. The primary reason for separate builders today is base image contents because that you cannot install common utilities without root access or access to folders outside of the layer folder.  This would work by using different sets of `[[order.group]]` entries in the builder. Dev container focused sets would include then need to include a "mode" buildpack at the start that only passes if the `BP_DCNB_BUILD_MODE` environment variable is set to "devcontainer". The steps in the `common-debian.sh` and `launcher-hack.sh` referenced in the Dockerfile could be contained inside this mode buildpack.

2. Given the way [Paketo buildpacks are set up](https://github.com/paketo-buildpacks/rfcs/blob/main/text/python/0001-restructure.md), it would be possible to reuse their `cpython` or `nodejs` buildpacks. To do so for Python, the pythonutils buildpack in this repository would need to be modified to add all needed devcontainer.json contents, and then add a requirement specifying `build=true` and `launch=true` in the metadata. However, dev container mode would not be able to reuse their npm install or pip install buildpacks. A secondary buildpack would be needed to add devcontainer.json metadata in those cases. The `goutils` buildpack is a simplified example of this model.
