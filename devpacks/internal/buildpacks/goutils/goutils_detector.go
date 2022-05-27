package goutils

import (
	"log"
	"os"
	"path/filepath"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/buildpacks/base"
	"github.com/chuxel/devpacks/internal/common/devcontainer"
)

type GoUtilsDetector struct {
	// Implements base.DefaultDetector

	// Detect(context libcnb.DetectContext) (libcnb.DetectResult, error)
	// DoDetect(context libcnb.DetectContext) (bool, map[string]interface{}, error)
	// Name() string
	// AlwaysPass() bool
}

func (detector GoUtilsDetector) Detect(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	return base.DefaultDetect(detector, context)
}

func (detector GoUtilsDetector) Name() string {
	return BUILDPACK_NAME
}

func (detector GoUtilsDetector) AlwaysPass() bool {
	return true
}

func (detector GoUtilsDetector) DoDetect(context libcnb.DetectContext) (bool, []libcnb.BuildPlanRequire, map[string]interface{}, error) {
	if devcontainer.ContainerImageBuildMode() != "devcontainer" {
		log.Println("Skipping since not in devcontainer mode.")
		return false, nil, nil, nil
	}

	// This buildpack always requires the "go" buildpack
	reqs := []libcnb.BuildPlanRequire{{Name: "go", Metadata: map[string]interface{}{
		"build":  true,
		"launch": true,
	}}}

	// Can be specified in project.toml or pack command line
	if os.Getenv("BP_GO_VERSION") != "" || os.Getenv("BP_GO_UTILS") != "" {
		return true, reqs, nil, nil
	}

	// Look for go.mod in the root - TODO: Others?
	filesToCheck := []string{"go.mod"}
	for _, file := range filesToCheck {
		if _, err := os.Stat(filepath.Join(context.Application.Path, file)); err == nil {
			log.Println("Detection passed.")
			return true, reqs, nil, nil
		}
	}

	log.Println("Go not detected.")
	return false, nil, nil, nil
}
