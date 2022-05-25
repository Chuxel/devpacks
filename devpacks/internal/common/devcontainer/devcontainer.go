package devcontainer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"

	"github.com/chuxel/devpacks/internal/common/utils"
	"github.com/tailscale/hujson"
)

// Pull in json as a simple map of maps given the structure
type DevContainer struct {
	Properties map[string]interface{}
	Path       string
}

func NewEmptyDevContainer() DevContainer {
	return DevContainer{
		Properties: make(map[string]interface{}),
		Path:       "",
	}
}

func NewDevContainer(applicationFolder string) DevContainer {
	devcontainer := NewEmptyDevContainer()
	devcontainer.Load(applicationFolder)
	if devcontainer.Path == "" {
		log.Fatal("Unable to find devcontainer.json file in ", applicationFolder)
	}
	return devcontainer
}

func (devContainer *DevContainer) Load(applicationFolder string) string {
	content, devContainerJsonPath := loadDevContainerJsonContent(applicationFolder)
	if devContainerJsonPath != "" {
		err := json.Unmarshal(content, &devContainer.Properties)
		if err != nil {
			log.Fatal("Failed to unmarshal devcontainer.json contents. ", err)
		}

	}
	devContainer.Path = devContainerJsonPath
	return devContainerJsonPath
}

func (devContainer *DevContainer) Merge(inDevContainer DevContainer) {
	devContainer.MergePropertyMap(inDevContainer.Properties)
}

func (devContainer *DevContainer) MergePropertyMap(inMap map[string]interface{}) {
	// Special processing for lifecycle commands - append rather than replace
	// TODO: Support array syntax... only string is supported for now
	lifecyclePropNames := []string{"initializeCommand", "onCreateCommand", "updateContentCommand", "postCreateCommand", "postStartCommand", "postAttachCommand"}
	mergedLifecycleProps := make(map[string]string)
	for _, prop := range lifecyclePropNames {
		val, hasKey := devContainer.Properties[prop]
		inVal, inHasKey := inMap[prop]
		if hasKey {
			if inHasKey {
				mergedLifecycleProps[prop] = fmt.Sprint(val) + "; " + fmt.Sprint(inVal)
			} else {
				mergedLifecycleProps[prop] = fmt.Sprint(val)
			}
		} else if inHasKey {
			mergedLifecycleProps[prop] = fmt.Sprint(inVal)
		}
	}

	// Handle other properties
	result := utils.MergeProperties(devContainer.Properties, inMap)

	// Update object with result
	devContainer.Properties = make(map[string]interface{})
	itr := reflect.ValueOf(result).MapRange()
	for itr.Next() {
		key := itr.Key().String()
		if utils.SliceContainsString(lifecyclePropNames, key) {
			devContainer.Properties[key] = reflect.ValueOf(mergedLifecycleProps[key]).Interface()
		} else {
			devContainer.Properties[key] = itr.Value().Interface()
		}
	}
}

func LoadDevContainerJsonAsMap(applicationFolder string) (map[string]json.RawMessage, string) {
	jsonMap := make(map[string]json.RawMessage)
	content, devContainerJsonPath := loadDevContainerJsonContent(applicationFolder)
	if devContainerJsonPath != "" {
		err := json.Unmarshal(content, &jsonMap)
		if err != nil {
			log.Fatal(err)
		}
	}
	return jsonMap, devContainerJsonPath
}

func (devContainer *DevContainer) ConvertUnsupportedPropertiesToRunArgs() {
	var runArgs []string
	if inter, hasKey := devContainer.Properties["runArgs"]; hasKey {
		runArgs = utils.InterfaceToStringSlice(inter)
	} else {
		runArgs = make([]string, 0)
	}

	if _, hasKey := devContainer.Properties["privileged"]; hasKey {
		runArgs = utils.AddToSliceIfUnique(runArgs, "--privileged")
		devContainer.Properties["privileged"] = nil
	}

	if _, hasKey := devContainer.Properties["init"]; hasKey {
		runArgs = utils.AddToSliceIfUnique(runArgs, "--init")
		devContainer.Properties["init"] = nil
	}

	if inter, hasKey := devContainer.Properties["capAdd"]; hasKey {
		capAdd := utils.InterfaceToStringSlice(inter)
		for _, cap := range capAdd {
			runArgs = utils.AddToSliceIfUnique(runArgs, "--cap-add="+cap)
			devContainer.Properties["capAdd"] = nil
		}
	}

	if inter, hasKey := devContainer.Properties["securityOpt"]; hasKey {
		capAdd := utils.InterfaceToStringSlice(inter)
		for _, cap := range capAdd {
			runArgs = utils.AddToSliceIfUnique(runArgs, "--security-op="+cap)
			devContainer.Properties["securityOpt"] = nil
		}
	}

	devContainer.Properties["runArgs"] = runArgs
}

func loadDevContainerJsonContent(applicationFolder string) ([]byte, string) {
	devContainerJsonPath := findDevContainerJson(applicationFolder)
	if devContainerJsonPath == "" {
		return []byte{}, devContainerJsonPath
	}
	content, err := ioutil.ReadFile(devContainerJsonPath)
	if err != nil {
		log.Fatal(err)
	}
	// Strip out comments to enable parsing
	ast, err := hujson.Parse(content)
	if err != nil {
		log.Fatal(err)
	}
	ast.Standardize()
	content = ast.Pack()

	return content, devContainerJsonPath
}

func findDevContainerJson(applicationFolder string) string {
	// Load devcontainer.json
	if applicationFolder == "" {
		var err error
		if applicationFolder, err = os.Getwd(); err != nil {
			log.Fatal("Failed to get current working directory. ", err)
		}
	}

	possiblePaths := []string{filepath.Join(applicationFolder, ".devcontainer", "devcontainer.json"), filepath.Join(applicationFolder, ".devcontainer.json"), filepath.Join(applicationFolder, "devcontainer.json")}
	for _, expectedPath := range possiblePaths {
		if _, err := os.Stat(expectedPath); err == nil {
			return expectedPath
		} else if !os.IsNotExist(err) {
			log.Fatal("Stat error for path. ", err)
		}
	}

	return ""
}
