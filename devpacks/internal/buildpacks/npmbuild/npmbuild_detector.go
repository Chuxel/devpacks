package npmbuild

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/buildpacks/base"
	"github.com/chuxel/devpacks/internal/buildpacks/nodejs"
	"github.com/chuxel/devpacks/internal/buildpacks/npminstall"
	"github.com/chuxel/devpacks/internal/common/devcontainer"
)

type NpmBuildDetector struct {
	// Implements base.DefaultDetector

	// Detect(context libcnb.DetectContext) (libcnb.DetectResult, error)
	// DoDetect(context libcnb.DetectContext) (bool, map[string]interface{}, error)
	// Name() string
	// AlwaysPass() bool
}

func (detector NpmBuildDetector) Detect(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	return base.DefaultDetect(detector, context)
}

func (detector NpmBuildDetector) Name() string {
	return BUILDPACK_NAME
}

func (detector NpmBuildDetector) AlwaysPass() bool {
	return false
}

func (detector NpmBuildDetector) DoDetect(context libcnb.DetectContext) (bool, []libcnb.BuildPlanRequire, map[string]interface{}, error) {
	if devcontainer.ContainerImageBuildMode() == "devcontainer" || !detector.hasNpmBuild(context.Application.Path) {
		log.Println("Skipping. Detected devcontainer build mode.")
		return false, nil, nil, nil
	}

	// This buildpack always requires nodejs
	reqs := []libcnb.BuildPlanRequire{
		{Name: nodejs.BUILDPACK_NAME, Metadata: map[string]interface{}{
			"build":  true,
			"launch": true,
		}}, {Name: npminstall.BUILDPACK_NAME, Metadata: map[string]interface{}{
			"build":  true,
			"launch": true,
		}}}

	log.Println("Detection passed.")
	return true, reqs, nil, nil
}

func (contrib NpmBuildDetector) hasNpmBuild(appPath string) bool {
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
		_, hasKey := packageJson.Scripts["build"]
		return hasKey
	}

	return false
}
