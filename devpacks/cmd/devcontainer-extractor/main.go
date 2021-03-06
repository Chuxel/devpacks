package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/chuxel/devpacks/internal/common/devcontainer"
	"github.com/chuxel/devpacks/internal/common/utils"
)

const LABEL_METADATA_TEMPLATE = "{{ index .Config.Labels \"" + devcontainer.DEVCONTAINER_JSON_LABEL_NAME + "\"}}"

func main() {
	if len(os.Args) < 2 {
		log.Println("Usage: devcontainer-extractor <image name> [project path] [merged devcontainer.json file output path]")
	}
	imageName := os.Args[1]

	projectPath, _ := os.Getwd()
	if len(os.Args) > 2 {
		projectPath = os.Args[2]
	}

	outPath := projectPath
	if len(os.Args) > 3 {
		outPath = os.Args[3]
	}

	imageLabelBytes := utils.ExecCmd("", true, "docker", "inspect", "-f", LABEL_METADATA_TEMPLATE, imageName)
	imageDevContainerJson := devcontainer.DevContainer{Properties: make(map[string]interface{})}

	if err := json.Unmarshal(imageLabelBytes, &imageDevContainerJson.Properties); err != nil {
		log.Fatal("Failed to parse devcontainer.json content from image label", err)
	}

	localDevContainerJson := devcontainer.DevContainer{Properties: make(map[string]interface{})}
	if devContainerJsonPath := localDevContainerJson.Load(projectPath); devContainerJsonPath == "" {
		log.Println("No devcontainer.json found in current directory. Skipping load.")
	}

	// Merge files together, handle unsupported properties
	localDevContainerJson.MergePropertyMap(imageDevContainerJson.Properties)
	localDevContainerJson.ConvertUnsupportedPropertiesToRunArgs()

	// Add an image reference if the file has neither a Dockerfile or Docker Compose file reference
	_, hasBuild := localDevContainerJson.Properties["build"]
	_, hasDockerCompose := localDevContainerJson.Properties["dockerComposeFile"]
	if !hasBuild && !hasDockerCompose {
		localDevContainerJson.Properties["image"] = imageName
	}

	if localDevContainerJsonBytes, err := json.MarshalIndent(localDevContainerJson.Properties, "", "\t"); err == nil {
		// Write the devcontainer.json.merged file
		if err := os.MkdirAll(outPath, 0755); err != nil {
			log.Fatal("Failed to create output directory: ", outPath)
		}
		if err := utils.WriteFile(filepath.Join(outPath, "devcontainer.json.merged"), localDevContainerJsonBytes); err != nil {
			log.Fatal("Failed to write devcontainer.json.merged file: ", err)
		}
		log.Println("devcontainer.json.merged file written to current directory!")
	} else {
		log.Fatal("Failed to marshal map to devcontainer.json.merged", err)
	}
}
