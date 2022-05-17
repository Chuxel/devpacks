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

type FinalizeLayerContributor struct {
	// Implements libcnb.LayerContributor
	// Contribute(context libcnb.ContributeContext) (libcnb.Layer, error)
	// Name() string
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
		// Load jsonc file
		featureConfigBytes, err := os.ReadFile(filepath.Join(loc, "feature.json"))
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
	devContainerJsonBytes, err := json.Marshal(mergedDevContainerJson)
	if err != nil {
		log.Fatal(err)
	}
	result.Labels = []libcnb.Label{
		{
			Key:   common.DEVCONTAINER_JSON_LABEL_NAME,
			Value: string(devContainerJsonBytes),
		},
	}

	//result.Layers = append(result.Layers, FinalizeLayerContributor{})

	// Handle unmets
	for _, entry := range context.Plan.Entries {
		if entry.Name != FINALIZE_BUILDPACK_NAME {
			result.Unmet = append(result.Unmet, libcnb.UnmetPlanEntry{Name: entry.Name})
		}
	}

	log.Printf("Unmet entries: %d", len(result.Unmet))

	return result, nil

}

// Implementation of libcnb.LayerContributor.Name
func (contrib FinalizeLayerContributor) Name() string {
	return FINALIZE_BUILDPACK_NAME
}

// Implementation of libcnb.LayerContributor.Contribute
func (contrib FinalizeLayerContributor) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	return layer, nil
}
