#!/bin/bash
set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

clear_cache_flag=""
if [ "${1:-false}" = "true" ]; then
    clear_cache_flag="--clear-cache"
fi

# Create buildpacks and builders
cd "${SCRIPT_DIR}/../devpacks"
make
cd "${SCRIPT_DIR}/../builders"
./create-builders.sh full
cd "${SCRIPT_DIR}/test-project"

# Run test
pack build test_image \
    -v \
    --pull-policy if-not-present \
    --builder ghcr.io/chuxel/devpacks/builder-devcontainer-full \
    ${clear_cache_flag} \
    --trust-builder
