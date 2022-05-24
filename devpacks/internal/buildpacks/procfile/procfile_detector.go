package procfile

import (
	"log"
	"os"
	"path/filepath"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/common/devcontainer"
)

type ProcfileDetector struct {
	// Implements libcnb.Detector
	// Detect(context libcnb.DetectContext) (libcnb.DetectResult, error)
}

func (detector ProcfileDetector) Detect(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	buildMode := devcontainer.ContainerImageBuildMode()
	log.Println("Devpack path:", context.Buildpack.Path)
	log.Println("Application path:", context.Application.Path)
	log.Println("Build mode:", buildMode)
	log.Println("Env:", os.Environ())

	if buildMode == "devcontainer" {
		log.Println("Skipping. Detected devcontainer build mode.")
		return libcnb.DetectResult{Pass: false}, nil
	}
	if _, err := os.Stat(filepath.Join(context.Application.Path, "Procfile")); err != nil {
		log.Println("Skipping. Did not find Procfile.")
		return libcnb.DetectResult{Pass: false}, nil
	}
	// TODO Allow other process types
	log.Println("Detection passed.")
	return libcnb.DetectResult{
		Pass: true,
		Plans: []libcnb.BuildPlan{
			{
				Provides: []libcnb.BuildPlanProvide{{Name: BUILDPACK_NAME}},
				Requires: []libcnb.BuildPlanRequire{{Name: BUILDPACK_NAME}},
			},
		},
	}, nil
}
