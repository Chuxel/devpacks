package npminstall

import (
	"log"
	"os"
	"path"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/buildpacks/base"
	"github.com/chuxel/devpacks/internal/buildpacks/nodejs"
	"github.com/chuxel/devpacks/internal/common"
)

type NpmInstallDetector struct {
	// Implements base.DefaultDetector

	// Detect(context libcnb.DetectContext) (libcnb.DetectResult, error)
	// DoDetect(context libcnb.DetectContext) (bool, map[string]interface{}, error)
	// Name() string
	// AlwaysPass() bool
}

func (detector NpmInstallDetector) Detect(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	return base.DefaultDetect(detector, context)
}

func (detector NpmInstallDetector) Name() string {
	return BUILDPACK_NAME
}

func (detector NpmInstallDetector) AlwaysPass() bool {
	return false
}

func (detector NpmInstallDetector) DoDetect(context libcnb.DetectContext) (bool, []libcnb.BuildPlanRequire, map[string]interface{}, error) {
	if common.ContainerImageBuildMode() == "devcontainer" {
		log.Println("Skipping. Detected devcontainer build mode.")
		return false, nil, nil, nil
	}

	// This buildpack always requires nodejs
	reqs := []libcnb.BuildPlanRequire{{Name: nodejs.BUILDPACK_NAME, Metadata: map[string]interface{}{
		"build":  true,
		"launch": true,
	}}}

	// Look for package json in the root
	if _, err := os.Stat(path.Join(context.Application.Path, "package-lock.json")); err != nil {
		log.Println("No package-lock.json found in ", context.Application.Path)
		return false, reqs, nil, nil
	}

	log.Println("Detection passed.")
	return true, reqs, nil, nil
}
