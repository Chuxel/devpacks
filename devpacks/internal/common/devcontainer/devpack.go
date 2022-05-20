package devcontainer

import (
	"os"
	"strings"
)

var cachedContainerImageBuildMode = ""

func ContainerImageBuildMode() string {
	if cachedContainerImageBuildMode != "" {
		return cachedContainerImageBuildMode
	}
	cachedContainerImageBuildMode := os.Getenv(CONTAINER_IMAGE_BUILD_MODE_ENV_VAR_NAME)
	if cachedContainerImageBuildMode == "" {
		if _, err := os.Stat(CONTAINER_IMAGE_BUILD_MODE_MARKER_PATH); err != nil {
			cachedContainerImageBuildMode = DEFAULT_CONTAINER_IMAGE_BUILD_MODE
		} else {
			fileBytes, err := os.ReadFile(CONTAINER_IMAGE_BUILD_MODE_MARKER_PATH)
			if err != nil {
				cachedContainerImageBuildMode = DEFAULT_CONTAINER_IMAGE_BUILD_MODE
			} else {
				cachedContainerImageBuildMode = strings.TrimSpace(string(fileBytes))
			}
		}
	}
	return cachedContainerImageBuildMode
}
