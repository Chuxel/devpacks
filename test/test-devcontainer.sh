#!/bin/bash
set -e
cd "$(dirname "${BASH_SOURCE[0]}")"

# Run test
pack build test_devcontainer_image \
    --pull-policy if-not-present \
    --builder ghcr.io/chuxel/devpacks/builder-devcontainer-full \
    --trust-builder \
    --path test-project

arch=$(uname -m)
if [ "$arch" = "x86_64" ]; then
    arch="amd64"
fi
os_arch="linux-${arch}"
if [ "$(uname -s)" = "Darwin" ]; then
    os_arch="darwin-${arch}"
fi

../devpacks/bin/devcontainer-extractor/devcontainer-extractor-$os_arch test_devcontainer_image
echo "Renaming to .devcontainer.json..."
mv -f devcontainer.json.merged .devcontainer.json