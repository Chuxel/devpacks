package finalize

import (
	_ "embed"
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/common"
	"github.com/tailscale/hujson"
)

type FinalizeBuilder struct {
	// Implements libcnb.Builder
	// Build(context libcnb.BuildContext) (libcnb.BuildResult, error)
}

func (builder FinalizeBuilder) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	buildMode := common.ContainerImageBuildMode()
	log.Println("Devpack path:", context.Buildpack.Path)
	log.Println("Application path:", context.Application.Path)
	log.Println("Build mode:", buildMode)
	log.Println("Number of plan entries:", len(context.Plan.Entries))
	log.Println("Env:", os.Environ())

	result := libcnb.NewBuildResult()

	// This implementation assumes https://github.com/devcontainers/spec/issues/2 is done. We'll
	// create a small utility to do the property conversion assuming these properties are missing.
	mergedDevContainerJson := common.DevContainer{Properties: make(map[string]interface{})}
	featureJsonSearchPath := os.Getenv(common.FINALIZE_JSON_SEARCH_PATH_ENV_VAR_NAME)
	featureJsonLocs := filepath.SplitList(featureJsonSearchPath)
	// For each path in search list
	for _, loc := range featureJsonLocs {
		filename := filepath.Join(loc, "feature.json")
		log.Println("Processing ", filename)
		// Load jsonc file
		featureConfigBytes, err := os.ReadFile(filename)
		if err != nil {
			log.Fatal(err)
		}
		ast, err := hujson.Parse(featureConfigBytes)
		if err != nil {
			log.Fatal(err)
		}
		ast.Standardize()
		content := ast.Pack()
		inMap := make(map[string]interface{})
		if err := json.Unmarshal(content, &inMap); err != nil {
			log.Fatal(err)
		}

		// Merge content
		mergedDevContainerJson.MergePropertyMap(inMap)
	}

	// Add the result to the label
	log.Println("Applying merged json content to label ", common.DEVCONTAINER_JSON_LABEL_NAME)
	devContainerJsonBytes, err := json.Marshal(mergedDevContainerJson.Properties)
	if err != nil {
		log.Fatal(err)
	}
	result.Labels = []libcnb.Label{
		{
			Key:   common.DEVCONTAINER_JSON_LABEL_NAME,
			Value: string(devContainerJsonBytes),
		},
	}

	// Set the default process to bash
	log.Println("Overriding default launch process...")
	result.Processes = append(result.Processes, libcnb.Process{
		Type:    "devcontainer",
		Command: "/cnb/lifecycle/launcher",
		Default: true,
		Direct:  true,
	})

	// Clear out workspace folder since we'll bind mount these contents
	log.Println("Removing workspace contents from image...")
	files, err := os.ReadDir(context.Application.Path)
	if err != nil {
		log.Fatal("Failed to get directory contents in", context.Application.Path, "-", err)
	}
	for _, file := range files {
		if err := os.RemoveAll(filepath.Join(context.Application.Path, file.Name())); err != nil {
			log.Fatal(err)
		}
	}

	// Handle unmets
	for _, entry := range context.Plan.Entries {
		if entry.Name != BUILDPACK_NAME {
			result.Unmet = append(result.Unmet, libcnb.UnmetPlanEntry{Name: entry.Name})
		}
	}

	log.Printf("Unmet entries: %d", len(result.Unmet))

	return result, nil

}
