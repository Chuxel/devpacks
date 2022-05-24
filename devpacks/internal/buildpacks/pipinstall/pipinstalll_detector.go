package pipinstall

import (
	"log"
	"os"
	"path/filepath"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/buildpacks/base"
	"github.com/chuxel/devpacks/internal/buildpacks/cpython"
	"github.com/chuxel/devpacks/internal/common/devcontainer"
)

type PipInstallDetector struct {
	// Implements base.DefaultDetector

	// Detect(context libcnb.DetectContext) (libcnb.DetectResult, error)
	// DoDetect(context libcnb.DetectContext) (bool, map[string]interface{}, error)
	// Name() string
	// AlwaysPass() bool
}

func (detector PipInstallDetector) Detect(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	return base.DefaultDetect(detector, context)
}

func (detector PipInstallDetector) Name() string {
	return BUILDPACK_NAME
}

func (detector PipInstallDetector) AlwaysPass() bool {
	return true
}

func (detector PipInstallDetector) DoDetect(context libcnb.DetectContext) (bool, []libcnb.BuildPlanRequire, map[string]interface{}, error) {
	if devcontainer.ContainerImageBuildMode() == "devcontainer" {
		log.Println("Skipping. Detected devcontainer build mode.")
		return false, nil, nil, nil
	}

	// This buildpack always requires cpython
	reqs := []libcnb.BuildPlanRequire{{Name: cpython.BUILDPACK_NAME, Metadata: map[string]interface{}{
		"build":  true,
		"launch": true,
	}}}

	// Check for requirements.txt - can't pip install otherwise
	if _, err := os.Stat(filepath.Join(context.Application.Path, "requirements.txt")); err == nil {
		log.Println("Detection passed.")
		return true, reqs, nil, nil
	}

	log.Println("requirements.txt not detected.")
	return false, nil, nil, nil
}
