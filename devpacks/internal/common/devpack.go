package common

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var cachedContainerImageBuildMode = ""

// Required configuration for processing
type DevpackSettings struct {
	Publisher  string   // aka GitHub Org
	FeatureSet string   // aka GitHub Repository
	Version    string   // Used for version pinning
	ApiVersion string   // Buildpack API version to target
	Stacks     []string // Array of stacks that the buildpack should support

	//func (dp *DevpackSettings) Load(featuresPath string)
}

func (dp *DevpackSettings) Load(featuresPath string) {
	if featuresPath == "" {
		featuresPath = os.Getenv(BUILDPACK_DIR_ENV_VAR_NAME)
	}
	content, err := ioutil.ReadFile(filepath.Join(featuresPath, DEVPACK_SETTINGS_FILENAME))
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(content, dp)
	if err != nil {
		log.Fatal(err)
	}
}

func GetContainerImageBuildMode() string {
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
