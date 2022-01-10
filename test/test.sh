#!/bin/bash
set -e
buildpack_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
docker build -t devcontainer-buildpack-run -f "${buildpack_root}/container-images/run.Dockerfile" .
pack build -v test_image \
    --builder paketobuildpacks/builder:full \
    --run-image devcontainer-buildpack-run \
    --buildpack "paketo-buildpacks/cpython" \
    --buildpack "${buildpack_root}/buildpacks/python-utils" 
