package npmstart

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/buildpacks/nodejs"
	"github.com/chuxel/devpacks/internal/buildpacks/npminstall"
	"github.com/chuxel/devpacks/internal/common"
)

type NpmStartDetector struct {
	// Implements libcnb.Detector
	// Detect(context libcnb.DetectContext) (libcnb.DetectResult, error)
}

func (detector NpmStartDetector) Detect(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	buildMode := common.ContainerImageBuildMode()
	log.Println("Devpack path:", context.Buildpack.Path)
	log.Println("Application path:", context.Application.Path)
	log.Println("Build mode:", buildMode)
	log.Println("Env:", os.Environ())

	if buildMode == "devcontainer" {
		log.Println("Skipping. Detected devcontainer build mode.")
		return libcnb.DetectResult{Pass: false}, nil
	} else if detector.hasNpmStart(context.Application.Path) {
		log.Println("Detection passed.")
		return libcnb.DetectResult{
			Pass: true,
			Plans: []libcnb.BuildPlan{
				{
					Provides: []libcnb.BuildPlanProvide{{Name: BUILDPACK_NAME}},
					Requires: []libcnb.BuildPlanRequire{
						{Name: BUILDPACK_NAME},
						{Name: nodejs.BUILDPACK_NAME, Metadata: map[string]interface{}{"build": true, "launch": true}},
						{Name: npminstall.BUILDPACK_NAME, Metadata: map[string]interface{}{"build": true, "launch": true}},
					},
				},
			},
		}, nil
	}
	log.Println("Skipping. Did not find npm start.")
	return libcnb.DetectResult{Pass: false}, nil
}

func (contrib NpmStartDetector) hasNpmStart(appPath string) bool {
	packageJsonPath := filepath.Join(appPath, "package.json")
	// Get engine value for nodejs if it exists in package.json
	if _, err := os.Stat(packageJsonPath); err == nil {
		type PackageJson struct {
			Scripts map[string]string
		}
		var packageJson PackageJson
		content, err := os.ReadFile(packageJsonPath)
		if err != nil {
			log.Fatal("Failed to read package.json", err)
		}
		if err := json.Unmarshal(content, &packageJson); err != nil {
			log.Fatal("Failed to parse package.json", err)
		}
		_, hasKey := packageJson.Scripts["start"]
		return hasKey
	}

	return false
}
