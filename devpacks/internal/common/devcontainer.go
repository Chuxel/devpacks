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

func loadDevContainerJsonConent(applicationFolder string) ([]byte, string) {
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
	content, devContainerJsonPath := loadDevContainerJsonConent(applicationFolder)
	if devContainerJsonPath != "" {
		err := json.Unmarshal(content, &devContainer.Properties)
		if err != nil {
			log.Fatal(err)
		}

	}
	return devContainerJsonPath
}

func (devContainer *DevContainer) MergePropertyMap(inMap map[string]interface{}) {
	result := mergePropertyMap(devContainer.Properties, inMap)
	devContainer.Properties = make(map[string]interface{})
	itr := reflect.ValueOf(result).MapRange()
	for itr.Next() {
		devContainer.Properties[itr.Key().String()] = itr.Value()
	}

}

func LoadDevContainerJsonAsMap(applicationFolder string) (map[string]json.RawMessage, string) {
	jsonMap := make(map[string]json.RawMessage)
	content, devContainerJsonPath := loadDevContainerJsonConent(applicationFolder)
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

func mergePropertyMap(existingVal interface{}, inVal interface{}) interface{} {
	typ := reflect.TypeOf(inVal).Kind()

	if typ == reflect.Slice || typ == reflect.Array {
		outVal := make([]interface{}, 0)
		if existingVal != nil {
			rExVal := reflect.ValueOf(existingVal)
			for i := 0; i < rExVal.Len(); i++ {
				outVal = append(outVal, rExVal.Index(i).Interface())
			}
		}
		rInVal := reflect.ValueOf(inVal)
		for i := 0; i < rInVal.Len(); i++ {
			outVal = append(outVal, rInVal.Index(i).Interface())
		}
		return outVal

	} else if typ == reflect.Map {
		if existingVal == nil {
			return inVal
		}

		outVal := make(map[string]interface{})
		rExVal := reflect.ValueOf(existingVal)
		rInVal := reflect.ValueOf(inVal)
		itr := rInVal.MapRange()
		for itr.Next() {
			outVal[itr.Key().String()] = mergePropertyMap(rExVal.MapIndex(itr.Key()).Interface(), itr.Value().Interface())
		}
		return outVal

	} else {
		return inVal
	}
}

/*
func mergeFeatureConfigToDevContainerJson(features []FeatureConfig, devContainerJsonMap map[string]json.RawMessage) map[string]json.RawMessage {
	finalizeFeatureConfig := generateFinalizeFeatureConfig(postProcessingConfig)
	var runArgs []string
	if devContainerJsonMap["runArgs"] != nil {
		if err := json.Unmarshal(devContainerJsonMap["runArgs"], &runArgs); err != nil {
			log.Fatal("Failed to unmarshal runArgs from devcontainer.json: ", err)
		}
	}
	if finalizeFeatureConfig.Privileged {
		runArgs = AddToSliceIfUnique(runArgs, "--privileged")
	}
	if finalizeFeatureConfig.Init {
		runArgs = AddToSliceIfUnique(runArgs, "--init")
	}
	if finalizeFeatureConfig.CapAdd != nil {
		for _, cap := range finalizeFeatureConfig.CapAdd {
			runArgs = AddToSliceIfUnique(runArgs, "--cap-add="+cap)
		}
	}
	if finalizeFeatureConfig.SecurityOpt != nil {
		for _, opt := range finalizeFeatureConfig.SecurityOpt {
			runArgs = AddToSliceIfUnique(runArgs, "--security-opt="+opt)
		}
	}
	devContainerJsonMap["runArgs"] = ToJsonRawMessage(runArgs)

	if finalizeFeatureConfig.Extensions != nil {
		var extensions []string
		if devContainerJsonMap["extensions"] != nil {
			if err := json.Unmarshal(devContainerJsonMap["extensions"], &extensions); err != nil {
				log.Fatal("Failed to unmarshal extensions from devcontainer.json: ", err)
			}
		}
		extensions = SliceUnion(extensions, finalizeFeatureConfig.Extensions)
		devContainerJsonMap["extensions"] = ToJsonRawMessage(extensions)
	}
	if finalizeFeatureConfig.Settings != nil {
		var settings map[string]interface{}
		if devContainerJsonMap["settings"] != nil {
			if err := json.Unmarshal(devContainerJsonMap["settings"], &settings); err != nil {
				log.Fatal("Failed to unmarshal extensions from devcontainer.json: ", err)
			}
		}
		//TODO: Settings merge
		devContainerJsonMap["settings"] = ToJsonRawMessage(settings)
	}
	if finalizeFeatureConfig.Mounts != nil {
		var mounts []string
		if devContainerJsonMap["mounts"] != nil {
			if err := json.Unmarshal(devContainerJsonMap["mounts"], &mounts); err != nil {
				log.Fatal("Failed to unmarshal mounts from devcontainer.json: ", err)
			}
		}
		for _, mount := range finalizeFeatureConfig.Mounts {
			mounts = AddToSliceIfUnique(runArgs, "source="+mount.Source+",target="+mount.Target+",type="+mount.Type)
		}
		devContainerJsonMap["mounts"] = ToJsonRawMessage(mounts)
	}
	return devContainerJsonMap
}
*/
