package base

import (
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/common"
)

const NODEJS_RUNTIME_BUILDPACK_NAME = "nodejs_runtime"

type BaseBuilder struct {
	// Implements libcnb.Builder
	// Build(context libcnb.BuildContext) (libcnb.BuildResult, error)

	ContributorType reflect.Type
}

type BaseLayerContributor interface {
	libcnb.LayerContributor
	// Contribute(context libcnb.ContributeContext) (libcnb.Layer, error)
	// Name() string

	ApplyBuilderSettings(layerTypes libcnb.LayerTypes, context libcnb.BuildContext)
}

// Implementation of libcnb.Builder.Build
func (builder BaseBuilder) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	buildMode := common.GetContainerImageBuildMode()
	log.Println("Devpack path:", context.Buildpack.Path)
	log.Println("Application path:", context.Application.Path)
	log.Println("Build mode:", buildMode)
	log.Println("Number of plan entries:", len(context.Plan.Entries))
	log.Println("Env:", os.Environ())

	var result libcnb.BuildResult

	overrideLayerTypes := map[string]bool{}
	for _, entry := range context.Plan.Entries {
		if entry.Name != NODEJS_RUNTIME_BUILDPACK_NAME {
			// If the entry is for this buildpack, merge values of any layer type overrides set in the entry's metadata
			for _, key := range []string{"Build", "Launch", "Cache"} {
				entryValue, containsKey := entry.Metadata[strings.ToLower(key)]
				if containsKey {
					existingValue, hasExistingValue := overrideLayerTypes[key]
					if hasExistingValue {
						overrideLayerTypes[key] = existingValue || entryValue.(bool)
					} else {
						overrideLayerTypes[key] = entryValue.(bool)
					}
				}
			}
		} else {
			// Otherwise consider the requirement unmet
			result.Unmet = append(result.Unmet, libcnb.UnmetPlanEntry{Name: entry.Name})
		}
	}
	// Override defaults as appropriate
	layerTypes := libcnb.LayerTypes{Build: true, Launch: true, Cache: true}
	for key, value := range overrideLayerTypes {
		field := reflect.ValueOf(&layerTypes).Elem().FieldByName(key)
		field.Set(reflect.ValueOf(value))
	}

	layerContributorPtr := reflect.New(builder.ContributorType)
	layerContributor := layerContributorPtr.Elem().Interface().(BaseLayerContributor)
	layerContributor.ApplyBuilderSettings(layerTypes, context)
	result.Layers = append(result.Layers, layerContributor)

	log.Printf("Number of layer contributors: %d", len(result.Layers))
	log.Printf("Unmet entries: %d", len(result.Unmet))

	return result, nil
}
