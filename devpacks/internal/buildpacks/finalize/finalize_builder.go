package finalize

import (
	_ "embed"
	"encoding/json"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/common/devcontainer"
)

type FinalizeBuilder struct {
	// Implements libcnb.Builder
	// Build(context libcnb.BuildContext) (libcnb.BuildResult, error)
}

func (builder FinalizeBuilder) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	buildMode := devcontainer.ContainerImageBuildMode()
	log.Println("Devpack path:", context.Buildpack.Path)
	log.Println("Application path:", context.Application.Path)
	log.Println("Build mode:", buildMode)
	log.Println("Number of plan entries:", len(context.Plan.Entries))
	log.Println("Env:", os.Environ())

	result := libcnb.NewBuildResult()

	// This implementation assumes https://github.com/devcontainers/spec/issues/2 is done. We'll
	// create a small utility to do the property conversion assuming these properties are missing.
	devcontainerJsonSearchPath := os.Getenv(devcontainer.FINALIZE_JSON_SEARCH_PATH_ENV_VAR_NAME)
	devcontainerJsonLocs := filepath.SplitList(devcontainerJsonSearchPath)
	labelContents := make([]map[string]interface{}, 0)
	// For each path in search list and merge properties
	for _, loc := range devcontainerJsonLocs {
		if _, err := os.Stat(path.Join(loc, "devcontainer.json")); err == nil {
			labelContents = append(labelContents, devcontainer.NewDevContainer(loc).Properties)
		}
	}
	// Force userEnvProbe to something other than "none" - needed so env vars are picked up
	labelContents = append(labelContents, map[string]interface{}{"userEnvProbe": "loginInteractiveShell"})

	// Add the result to the label
	log.Println("Adding dev container metadata content to label ", devcontainer.DEVCONTAINER_JSON_LABEL_NAME)
	devContainerJsonBytes, err := json.Marshal(labelContents)
	if err != nil {
		log.Fatal(err)
	}
	result.Labels = []libcnb.Label{
		{
			Key:   devcontainer.DEVCONTAINER_JSON_LABEL_NAME,
			Value: string(devContainerJsonBytes),
		},
	}

	// Set the default process to bash
	log.Println("Overriding default launch process...")
	result.Processes = append(result.Processes, libcnb.Process{
		Type:      "devcontainer",
		Command:   "/bin/bash",
		Arguments: []string{"-c", "echo 'Dev container now started and waiting for connection.' && while true; do sleep 100; done"},
		Default:   true,
		Direct:    true,
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
