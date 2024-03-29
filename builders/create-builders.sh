#!/bin/bash
set -e
export DOCKER_BUILDKIT=1
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
builder_name="${1:-"empty"}"
publish="${2:-false}"

publisher="chuxel"
repository="devpacks"
version="0.0.8"

mkdir -p /tmp/builder-tmp

create_builder() {
    local builder_type=$1
    local toml_dir="${script_dir}/${builder_name}"
    local toml="$(cat "${toml_dir}/builder-${builder_type}.toml")"
    toml="${toml//\${publisher\}/${publisher}}"
    toml="${toml//\${repository\}/${repository}}"
    toml="${toml//\${version\}/${version}}"
    toml="${toml//\${toml_dir\}/${toml_dir}}"
    echo "${toml}" > /tmp/builder-tmp/builder-${builder_type}.toml
    local uri="ghcr.io/${publisher}/${repository}/builder-${builder_type}-${builder_name}"
    pack builder create "${uri}" --pull-policy if-not-present -c /tmp/builder-tmp/builder-${builder_type}.toml
    if [ "${publish}" = "true" ]; then
        echo "(*) Publishing..."
        docker push "${uri}"
    fi
}

echo "(*) Creating ${builder_name} devcontainer builder..."
create_builder devcontainer
echo "(*) Creating ${builder_name} prod builder..."
create_builder prod

rm -rf /tmp/builder-tmp
