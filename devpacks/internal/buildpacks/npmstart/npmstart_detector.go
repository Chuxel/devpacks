package npmstart

import (
	"log"
	"os"

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

	var result libcnb.DetectResult
	if buildMode == "devcontainer" {
		log.Println("Skipping. Detected devcontainer build mode.")
		result.Pass = false
	} else {
		result.Plans = []libcnb.BuildPlan{
			{
				Provides: []libcnb.BuildPlanProvide{{Name: BUILDPACK_NAME}},
				Requires: []libcnb.BuildPlanRequire{
					{Name: BUILDPACK_NAME},
					{Name: nodejs.BUILDPACK_NAME, Metadata: map[string]interface{}{"build": true, "launch": true}},
					{Name: npminstall.BUILDPACK_NAME},
				},
			},
		}
		result.Pass = true
		log.Println("Detection passed.")
	}
	return result, nil
}
