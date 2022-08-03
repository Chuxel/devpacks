#!/bin/bash
set -e
cd "$(dirname "${BASH_SOURCE[0]}")"

# Run test
pack build test_prod_image \
    -v \
    --pull-policy if-not-present \
    --builder ghcr.io/chuxel/devpacks/builder-prod-full \
    --trust-builder \
    --path test-project

