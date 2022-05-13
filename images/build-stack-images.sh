#!/bin/bash
set -e
export DOCKER_BUILDKIT=1
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
devcontainer_features_dir="${script_dir}/../devcontainer-features"
publish="${1:-false}"

publisher="$(jq -r '.publisher' "${devcontainer_features_dir}"/devpack-settings.json)"
featureset_name="$(jq -r '.featureSet' "${devcontainer_features_dir}"/devpack-settings.json)"
version="$(jq -r '.version' "${devcontainer_features_dir}"/devpack-settings.json)"
uri_prefix="ghcr.io/${publisher}/${featureset_name}"

build_stack_images() {
    prefix=""
    if [ ! -z "${1}" ]; then
        prefix="${1}-"
    fi
    docker build -t "${uri_prefix}/stack-${prefix}build-image" --cache-from "${uri_prefix}/stack-${prefix}build-image" --target ${prefix}build "${script_dir}"
    docker build -t "${uri_prefix}/stack-${prefix}run-image" --cache-from "${uri_prefix}/stack-${prefix}run-image" --target ${prefix}run "${script_dir}"
    if [ "${publish}" = "true" ]; then
        echo "(*) Publishing..."
        docker push "${uri_prefix}/stack-${prefix}build-image"
        docker push "${uri_prefix}/stack-${prefix}run-image"
    fi
}

# Create two stacks - normal, devcontainer
echo "(*) Building stack images..."
build_stack_images
echo "(*) Building devcontainer stack images..."
build_stack_images devcontainer