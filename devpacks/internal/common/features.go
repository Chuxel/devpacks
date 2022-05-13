package common

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type FeatureMount struct {
	Source string `json:"source,omitempty"`
	Target string `json:"target,omitempty"`
	Type   string `json:"type,omitempty"`
}

type FeatureOption struct {
	Type        string      `json:"type,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
	Proposals   []string    `json:"proposals,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Description string      `json:"description"`
}

type FeatureConfig struct {
	Id           string                   `json:"id,omitempty"`
	Name         string                   `json:"name,omitempty"`
	Options      map[string]FeatureOption `json:"options,omitempty"`
	Extensions   []string                 `json:"extensions,omitempty"`
	Settings     map[string]interface{}   `json:"settings,omitempty"`
	Entrypoint   string                   `json:"entrypoint,omitempty"`
	Privileged   bool                     `json:"privileged,omitempty"`
	Init         bool                     `json:"init,omitempty"`
	ContainerEnv map[string]string        `json:"containerEnv,omitempty"`
	Mounts       []FeatureMount           `json:"mounts,omitempty"`
	CapAdd       []string                 `json:"capAdd,omitempty"`
	SecurityOpt  []string                 `json:"securityOpt,omitempty"`
	BuildArg     string                   `json:"buildArg,omitempty"`

	// SetProperties(propertyMap map[string]interface{})
	// BuildEnvironment(optionSelections map[string]string, additionalVariables map[string]string) []string
	// OptionEnvVarName(prefix string, optionId string) string
	// ScriptPath(buidpackPath string, script string) string
	// LoadBytes(featureBytes []byte)
}

type FeaturesJson struct {
	Features []FeatureConfig `json:"features"`

	// Load(featuresPath string)
}

type LayerFeatureMetadata struct {
	Id               string
	Version          string
	Config           FeatureConfig
	OptionSelections map[string]string
}

func (featureConfig *FeatureConfig) LoadBytes(featureBytes []byte) {
	if err := json.Unmarshal(featureBytes, featureConfig); err != nil {
		log.Fatal(err)
	}
}

func (feature *FeatureConfig) SetProperties(propertyMap map[string]interface{}) {
	for property, value := range propertyMap {
		if value != nil {
			switch property {
			case "Mounts":
				out := []FeatureMount{}
				inputInterfaceArray := value.([]map[string]interface{})
				for _, value := range inputInterfaceArray {
					obj := PropertyMapToInterface(value, reflect.TypeOf(FeatureMount{})).(*FeatureMount)
					out = append(out, *obj)
				}
				feature.Mounts = out
			case "Options":
				// Convert map[string]interface{} to map[string]FeatureOption
				out := make(map[string]FeatureOption)
				inputInterfaceMap := value.(map[string]interface{})
				for key, value := range inputInterfaceMap {
					valuePropertyMap := value.(map[string]interface{})
					obj := PropertyMapToInterface(valuePropertyMap, reflect.TypeOf(FeatureOption{})).(*FeatureOption)
					out[key] = *obj
				}
				feature.Options = out
			default:
				field := reflect.ValueOf(feature).Elem().FieldByName(property)
				SetFieldValue(field, value)
			}
		}
	}
}

func (feature *FeatureConfig) BuildEnvironment(optionSelections map[string]string, additionalVariables map[string]string) []string {
	// Create environment that includes feature build args
	env := append(os.Environ(),
		feature.OptionEnvVarName(OPTION_SELECTION_ENV_VAR_PREFIX, "")+"=true")
	for optionId, selection := range optionSelections {
		if selection != "" {
			env = append(env, feature.OptionEnvVarName(OPTION_SELECTION_ENV_VAR_PREFIX, optionId)+"="+selection)
		}
	}
	for varName, varValue := range additionalVariables {
		env = append(env, feature.OptionEnvVarName(OPTION_SELECTION_ENV_VAR_PREFIX, varName)+"="+varValue)
	}
	return env
}

func (feature *FeatureConfig) OptionEnvVarName(prefix string, optionId string) string {
	if prefix == "" {
		prefix = OPTION_SELECTION_ENV_VAR_PREFIX
	}
	featureIdSafe := strings.ReplaceAll(strings.ToUpper(feature.Id), "-", "_")
	name := prefix + featureIdSafe
	if optionId != "" {
		optionIdSafe := strings.ReplaceAll(strings.ToUpper(optionId), "-", "_")
		name = prefix + featureIdSafe + "_" + strings.ToUpper(strings.ReplaceAll(optionIdSafe, "-", "_"))
	}
	return name
}

func (feature *FeatureConfig) ScriptPath(buidpackPath string, script string) string {
	return filepath.Join(buidpackPath, "features", feature.Id, "bin", script)
}

func (featuresJson *FeaturesJson) Load(featuresPath string) {
	// Load devcontainer-features.json or features.json
	if featuresPath == "" {
		featuresPath = os.Getenv(BUILDPACK_DIR_ENV_VAR_NAME)
	}
	content, err := ioutil.ReadFile(filepath.Join(featuresPath, "devcontainer-features.json"))
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(content, featuresJson)
	if err != nil {
		log.Fatal(err)
	}
}

func (lfm *LayerFeatureMetadata) SetProperties(propertyMap map[string]interface{}) {
	for property, value := range propertyMap {
		if value != nil {
			if property == "Config" {
				out := FeatureConfig{}
				inputInterfaceMap := value.(map[string]interface{})
				out.SetProperties(inputInterfaceMap)
				lfm.Config = out
			} else {
				field := reflect.ValueOf(lfm).Elem().FieldByName(property)
				SetFieldValue(field, value)
			}
		}
	}
}

func GetOptionMetadataKey(optionId string) string {
	return OPTION_METADATA_KEY_PREFIX + strings.ToLower(strings.ReplaceAll(optionId, "-", "_"))
}
