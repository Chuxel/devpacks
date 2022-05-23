package base

import (
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/common/devcontainer"
)

type DefaultBuilder interface {
	libcnb.Builder

	Name() string
	NewLayerContributor(buildMode string, layerTypes libcnb.LayerTypes, context libcnb.BuildContext) libcnb.LayerContributor
}

func DefaultBuild(builder DefaultBuilder, context libcnb.BuildContext) (libcnb.BuildResult, error) {
	buildMode := devcontainer.ContainerImageBuildMode()
	log.Println("Devpack path:", context.Buildpack.Path)
	log.Println("Application path:", context.Application.Path)
	log.Println("Build mode:", buildMode)
	log.Println("Number of plan entries:", len(context.Plan.Entries))
	log.Println("Env:", os.Environ())

	result := libcnb.NewBuildResult()

	hasEntry := false
	overrideLayerTypes := map[string]bool{}
	for _, entry := range context.Plan.Entries {
		if entry.Name == builder.Name() {
			// If the entry is for this buildpack, merge values of any layer type overrides set in the entry's metadata
			for _, key := range []string{"build", "launch", "cache"} {
				entryValue, containsKey := entry.Metadata[key]
				if containsKey {
					existingValue, hasExistingValue := overrideLayerTypes[key]
					if hasExistingValue {
						overrideLayerTypes[key] = existingValue || entryValue.(bool)
					} else {
						overrideLayerTypes[key] = entryValue.(bool)
					}
				}
			}
			hasEntry = true
		} else {
			// Otherwise consider the requirement unmet
			result.Unmet = append(result.Unmet, libcnb.UnmetPlanEntry{Name: entry.Name})
		}
	}
	// There could be scenarios where the capability is provided, but nothing
	// requires it. In this case, the buildpack should not execute its steps.
	if !hasEntry {
		log.Println("Buildpack", builder.Name(), "not required. Skipping.")
		return result, nil
	}

	// Override defaults as appropriate
	layerTypes := libcnb.LayerTypes{Build: true, Launch: buildMode == "devcontainer", Cache: true}
	for key, value := range overrideLayerTypes {
		field := reflect.ValueOf(&layerTypes).Elem().FieldByName(strings.ToUpper(key[0:1]) + key[1:])
		field.Set(reflect.ValueOf(value))
	}

	// Use reflection to create a contributor based on the type assigned to the builder
	result.Layers = append(result.Layers, builder.NewLayerContributor(buildMode, layerTypes, context))

	log.Printf("Number of layer contributors: %d", len(result.Layers))
	log.Printf("Unmet entries: %d", len(result.Unmet))

	return result, nil
}
