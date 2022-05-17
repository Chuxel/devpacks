package finalize

import (
	"log"
	"os"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/common"
)

type FinalizeDetector struct {
	// Implements libcnb.Detector
	// Detect(context libcnb.DetectContext) (libcnb.DetectResult, error)
}

func (detector FinalizeDetector) Detect(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	buildMode := common.ContainerImageBuildMode()
	log.Println("Devpack path:", context.Buildpack.Path)
	log.Println("Application path:", context.Application.Path)
	log.Println("Build mode:", buildMode)
	log.Println("Env:", os.Environ())

	var result libcnb.DetectResult
	if buildMode == "devcontainer" {
		result.Plans = []libcnb.BuildPlan{
			{
				Provides: []libcnb.BuildPlanProvide{{Name: FINALIZE_BUILDPACK_NAME}},
				Requires: []libcnb.BuildPlanRequire{{Name: FINALIZE_BUILDPACK_NAME}},
			},
		}
		result.Pass = true
		log.Println("Detection passed.")
	} else {
		log.Println("Skipping since not in devcontainer mode.")
		result.Pass = false
	}
	return result, nil
}
