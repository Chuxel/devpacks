package devcontainer

const DEFAULT_CONTAINER_IMAGE_BUILD_MODE = "production"

// Label and metadata keys
const METADATA_ID_PREFIX = "dev.containers"
const DEVCONTAINER_JSON_LABEL_NAME = METADATA_ID_PREFIX + ".metadata"

// ENV variables
const CONTAINER_IMAGE_BUILD_MODE_ENV_VAR_NAME = "BP_DCNB_BUILD_MODE"
const FINALIZE_JSON_SEARCH_PATH_ENV_VAR_NAME = "FINALIZE_JSON_SEARCH_PATH"

// Paths and filenames
const DEVCONTAINER_CONFIG_RELATIVE_ROOT = "/etc/dev-container-features"
const CONTAINER_IMAGE_BUILD_MODE_MARKER_PATH = "/usr/local" + DEVCONTAINER_CONFIG_RELATIVE_ROOT + "/dcnb-build-mode"
