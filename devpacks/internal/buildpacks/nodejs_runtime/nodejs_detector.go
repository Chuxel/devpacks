package nodejs_runtime

import (
	"log"
	"os"
	"os/exec"
	"reflect"
	"strings"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/common"
	"github.com/joho/godotenv"
)

const DevContainerFeaturesEnvPath = "/tmp/devcontainer-features.env"

type FeatureDetector struct {
	// Implements libcnb.Detector
	// Detect(context libcnb.DetectContext) (libcnb.DetectResult, error)
}

// Implementation of libcnb.Detector.Detect
func (fd FeatureDetector) Detect(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	log.Println("Devpack path:", context.Buildpack.Path)
	log.Println("Application path:", context.Application.Path)
	log.Println("Env:", os.Environ())

	var result libcnb.DetectResult

	// Load features.json, buildpack settings
	devpackSettings := common.DevpackSettings{}
	devpackSettings.Load(context.Buildpack.Path)
	featuresJson := common.FeaturesJson{}
	featuresJson.Load(context.Buildpack.Path)
	log.Println("Number of features in Devpack:", len(featuresJson.Features))

	// Load devcontainer.json if in devcontainer build mode
	var devContainerJson common.DevContainerJson
	if common.GetContainerImageBuildMode() == "devcontainer" {
		devContainerJson.Load(context.Application.Path)
	}

	// See if should provide any features
	var plan libcnb.BuildPlan
	onlyProvided := []libcnb.BuildPlanProvide{}
	for _, feature := range featuresJson.Features {
		detected, provide, require, err := detectFeature(context, devpackSettings, feature, devContainerJson)
		if err != nil {
			return result, err
		}
		if detected {
			log.Printf("- %s detected\n", feature.Id)
			plan.Provides = append(plan.Provides, provide)
			plan.Requires = append(plan.Requires, require)
		} else {
			onlyProvided = append(onlyProvided, provide)
			log.Printf("- %s provided\n", provide.Name)
		}

	}

	result.Plans = append(result.Plans, plan)
	// Generate all permutations where something is just provided
	combinationList := common.GetAllCombinations(len(onlyProvided))
	for _, combination := range combinationList {
		var optionalPlan libcnb.BuildPlan
		copy(optionalPlan.Requires, plan.Requires)
		copy(optionalPlan.Provides, plan.Provides)
		for _, i := range combination {
			optionalPlan.Provides = append(optionalPlan.Provides, onlyProvided[i])
		}
		result.Plans = append(result.Plans, optionalPlan)
	}

	// Always pass since we can provide features even if they're not used by this buildpack
	result.Pass = true
	return result, nil
}

func detectFeature(context libcnb.DetectContext, buildpackSettings common.DevpackSettings, feature common.FeatureConfig, devContainerJson common.DevContainerJson) (bool, libcnb.BuildPlanProvide, libcnb.BuildPlanRequire, error) {
	// e.g. chuxel/devcontainer/features/packcli
	fullFeatureId := feature.FullFeatureId(buildpackSettings, "/")
	provide := libcnb.BuildPlanProvide{Name: fullFeatureId}
	require := libcnb.BuildPlanRequire{Name: fullFeatureId, Metadata: make(map[string]interface{})}

	// Always set build mode
	optionSelections := map[string]string{common.BUILD_MODE_OPTION_NAME: common.GetContainerImageBuildMode()}
	// Add any option selections from BP_CONTAINER_FEATURE_<feature.Id>_<option> env vars and devcontainer.json (in devcontainer mode)
	detected, optionSelections := detectOptionSelections(feature, buildpackSettings, devContainerJson)
	// Always add optionSelections to require metadata
	for optionId, selection := range optionSelections {
		require.Metadata[common.GetOptionMetadataKey(optionId)] = selection
	}

	// Check if detect script for feature exists, return whatever the result of the devcontainer.json and env var detection happens to be
	detectScriptPath := feature.ScriptPath(context.Buildpack.Path, "detect")
	_, err := os.Stat(detectScriptPath)
	if err != nil {
		return detected, provide, require, nil
	}

	// Execute the script - set path to where a resulting devcontainer-features.env should be placed as env var
	log.Printf("- Executing %s\n", detectScriptPath)
	env := feature.BuildEnvironment(optionSelections, map[string]string{
		"SELECTION_ENV_FILE_PATH": DevContainerFeaturesEnvPath,
	})
	logWriter := log.Writer()
	detectCommand := exec.Command(detectScriptPath)
	detectCommand.Env = env
	detectCommand.Stdout = logWriter
	detectCommand.Stderr = logWriter
	if err := detectCommand.Run(); err != nil {
		log.Fatal(err)
	}

	exitCode := detectCommand.ProcessState.ExitCode()
	if exitCode == 0 {
		// Read option selections if any are provided
		if _, err := os.Stat(DevContainerFeaturesEnvPath); err != nil {
			if err := godotenv.Load(DevContainerFeaturesEnvPath); err != nil {
				log.Fatal(err)
			}
			_, optionSelections = mergeOptionSelectionsFromEnv(feature, optionSelections, common.OPTION_SELECTION_ENV_VAR_PREFIX)
			for option, selection := range optionSelections {
				require.Metadata[common.GetOptionMetadataKey(option)] = selection
			}
		}
		return true, provide, require, nil
	}
	// 100 means failed, other error codes mean an error ocurred
	if exitCode == 100 {
		return false, provide, require, nil
	} else {
		return false, provide, require, common.NonZeroExitError{ExitCode: exitCode}
	}
}

func detectOptionSelections(feature common.FeatureConfig, buildpackSettings common.DevpackSettings, devContainerJson common.DevContainerJson) (bool, map[string]string) {
	optionSelections := make(map[string]string)
	detectedDevContainerJson := false
	// If in dev container mode, parse devcontainer.json features (if any)
	if common.GetContainerImageBuildMode() == "devcontainer" {
		fullFeatureId := feature.FullFeatureId(buildpackSettings, "/")
		for featureName, jsonOptionSelections := range devContainerJson.Features {
			if featureName == fullFeatureId || strings.HasPrefix(featureName, fullFeatureId+"@") {
				detectedDevContainerJson = true
				if reflect.TypeOf(jsonOptionSelections).String() == "string" {
					optionSelections["version"] = jsonOptionSelections.(string)
				} else {
					// Use reflection to convert the from a map[string]interface{} to a map[string]string
					mapRange := reflect.ValueOf(jsonOptionSelections).MapRange()
					for mapRange.Next() {
						optionSelections[mapRange.Key().String()] = mapRange.Value().Elem().String()
					}
				}
				break
			}
		}
	}

	// Look for BP_CONTAINER_FEATURE_<feature.Id>_<option> environment variables, convert
	detectedEnv, optionselections := mergeOptionSelectionsFromEnv(feature, optionSelections, common.PROJECT_TOML_OPTION_SELECTION_ENV_VAR_PREFIX)
	return (detectedDevContainerJson || detectedEnv), optionselections
}

func mergeOptionSelectionsFromEnv(feature common.FeatureConfig, optionSelections map[string]string, prefix string) (bool, map[string]string) {
	detected := false
	enabledEnvVarVal := os.Getenv(feature.OptionEnvVarName(prefix, ""))
	if enabledEnvVarVal != "" && enabledEnvVarVal != "false" {
		detected = true
	}
	for optionId := range feature.Options {
		optionValue := os.Getenv(feature.OptionEnvVarName(prefix, optionId))
		if optionValue != "" {
			optionSelections[optionId] = optionValue
		}
	}
	return detected, optionSelections
}
