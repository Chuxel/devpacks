# Chuxel's development buildpack sandbox

*Everything here will change - a lot. Don't depend on any of it for anything you're doing. Good for demos, that's it.*

## Development setup

1. Open this repository in GitHub Codespaces or Remote - Containers using VS Code
2. File > Open workspace and select `workspace.code-workspace`

## What is here

This repo demonstrates the value of https://github.com/devcontainers/spec/issues/18 (and https://github.com/devcontainers/spec/issues/2) by integrating a development container metadata into [Cloud Native Buildpacks](https://buildpacks.io/). Repo contents:

1. A set of buildpacks under `devpacks`
2. A set of stack images under `images`
3. Two builders that include (1) and (2) under `builders`
4. A utility to generate a `devcontainer.json` file from the output of one of the builders

The `ghcr.io/chuxel/devpacks/builder-prod-full` builder behaves like a typical buildpack, while `ghcr.io/chuxel/devpacks/builder-devcontainer-full` instead focuses on a dev container image that is similar to production. 

These builders can be used with the [`pack` CLI](https://buildpacks.io/docs/tools/pack/) or other buildpacks v3 compliant tools. 

An extractor utility (see [releases](https://github.com/Chuxel/devpacks/releases)) can then be used to generate a `devcontainer.json` or merge a file from the output of the `builder-devcontainer-full` builder. It will also add a reference to the specified image if no Dockerfile or Docker Compose file is referenced in the local `devcontainer.json` (and otherwise assumes you've referenced the right image in these files). This allows developers to add additional metadata and settings to the devcontainer.json file not supported by the buildpack (including dev container features).

Right now it supports basic Node.js apps with a `start` entry in `package.json`, and basic Python applications that use `pip` (and thus have a `requirements.txt` file) and include a [`Procfile`](https://devcenter.heroku.com/articles/procfile) with a `web` entry to specify the startup command.

## How it works

Buildpacks are written in Go and take advantage of libcnb to simplify interop with the buildpack spec.

1. The base images for the `prod` and `devcontainer` builders are constructed using a multi-stage Dockerfile with dev later container stages adding more base content, but otherwise the same structure - which ensures consistency.
2. A "build mode" allows for dual-purpose buildpacks that can either alter behaviors or simply not be detected when in one mode or the other. For example, a `pythonutils` buildpack that injects tools like `pylint` only executes in devcontainer mode, while others like `nodejs` or `cpython` execute in both modes. A file in the build image can be used determine the mode, but they also support passing in the mode as a `BP_DCNB_BUILD_MODE` environment variable.
3. Buildpacks like `npminstall` then add "requires" entries on runtime buildpacks like `nodejs` to ensure proper ordering. 
4. The buildpacks can optionally place a `devcontainer.json` snippet file in their layer and add the path to it in a common `FINALIZE_JSON_SEARCH_PATH` environment variable. 
4. A `finalize` buildpack merges all devcontainer.json snippets from the `FINALIZE_JSON_SEARCH_PATH` and adds a `dev.containers.json` label on the image with their contents.
5. The `finalize` buildpack also removes the source code since this is expected to be mounted as a volume or bind mounted when the image is used. As a result, `finalize` will fail detection in production mode and is the last in order in the devcontainer builder.
7. Since the dev container CLI (and related products like VS Code Remote - Containers and Codespaces) does not support dev container metadata in an image label, an extractor utility (see `devpacks/cmd/devcontainer-extractor`) extracts metadata from the image and merges it with a local devcontainer.json file if one is found (creating `devcontainer.json.merged`). It also translates properties proposed in https://github.com/devcontainers/spec/issues/2 to `runArgs` equivalents.

That's the scoop!