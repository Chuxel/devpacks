package common

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"

	"github.com/tailscale/hujson"
)

// Pull in json as a simple map of maps given the structure
type DevContainer struct {
	Properties map[string]interface{}
}

func loadDevContainerJsonContent(applicationFolder string) ([]byte, string) {
	devContainerJsonPath := FindDevContainerJson(applicationFolder)
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

func (devContainer *DevContainer) Load(applicationFolder string) string {
	content, devContainerJsonPath := loadDevContainerJsonContent(applicationFolder)
	if devContainerJsonPath != "" {
		err := json.Unmarshal(content, &devContainer.Properties)
		if err != nil {
			log.Fatal("Failed to unmarshal devcontainer.json contents. ", err)
		}

	}
	return devContainerJsonPath
}

func (devContainer *DevContainer) MergePropertyMap(inMap map[string]interface{}) {
	result := MergeProperties(devContainer.Properties, inMap)
	devContainer.Properties = make(map[string]interface{})
	itr := reflect.ValueOf(result).MapRange()
	for itr.Next() {
		devContainer.Properties[itr.Key().String()] = itr.Value().Interface()
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

func FindDevContainerJson(applicationFolder string) string {
	// Load devcontainer.json
	if applicationFolder == "" {
		var err error
		applicationFolder, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
	}

	expectedPath := filepath.Join(applicationFolder, ".devcontainer", "devcontainer.json")
	if _, err := os.Stat(expectedPath); err != nil {
		// if file does not exist, try .devcontainer.json instead
		if os.IsNotExist(err) {
			expectedPath = filepath.Join(applicationFolder, ".devcontainer.json")
			if _, err := os.Stat(expectedPath); err != nil {
				if !os.IsNotExist(err) {
					log.Fatal(err)
				}
				return ""
			}
		} else {
			log.Fatal(err)
		}
	}
	return expectedPath
}

func (devContainer *DevContainer) ConvertUnsupportedPropertiesToRunArgs() {
	var runArgs []string
	if inter, hasKey := devContainer.Properties["runArgs"]; hasKey {
		runArgs = InterfaceToStringSlice(inter)
	} else {
		runArgs = make([]string, 0)
	}

	if _, hasKey := devContainer.Properties["privileged"]; hasKey {
		runArgs = AddToSliceIfUnique(runArgs, "--privileged")
		devContainer.Properties["privileged"] = nil
	}

	if _, hasKey := devContainer.Properties["init"]; hasKey {
		runArgs = AddToSliceIfUnique(runArgs, "--init")
		devContainer.Properties["init"] = nil
	}

	if inter, hasKey := devContainer.Properties["capAdd"]; hasKey {
		capAdd := InterfaceToStringSlice(inter)
		for _, cap := range capAdd {
			runArgs = AddToSliceIfUnique(runArgs, "--cap-add="+cap)
			devContainer.Properties["capAdd"] = nil
		}
	}

	if inter, hasKey := devContainer.Properties["securityOpt"]; hasKey {
		capAdd := InterfaceToStringSlice(inter)
		for _, cap := range capAdd {
			runArgs = AddToSliceIfUnique(runArgs, "--security-op="+cap)
			devContainer.Properties["securityOpt"] = nil
		}
	}

	devContainer.Properties["runArgs"] = runArgs
}
