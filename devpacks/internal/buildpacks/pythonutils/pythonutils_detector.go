package pythonutils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/buildpacks/base"
	"github.com/chuxel/devpacks/internal/buildpacks/cpython"
	"github.com/chuxel/devpacks/internal/common/devcontainer"
)

type PythonUtilsDetector struct {
	// Implements base.DefaultDetector

	// Detect(context libcnb.DetectContext) (libcnb.DetectResult, error)
	// DoDetect(context libcnb.DetectContext) (bool, map[string]interface{}, error)
	// Name() string
	// AlwaysPass() bool
}

func (detector PythonUtilsDetector) Detect(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	return base.DefaultDetect(detector, context)
}

func (detector PythonUtilsDetector) Name() string {
	return BUILDPACK_NAME
}

func (detector PythonUtilsDetector) AlwaysPass() bool {
	return true
}

func (detector PythonUtilsDetector) DoDetect(context libcnb.DetectContext) (bool, []libcnb.BuildPlanRequire, map[string]interface{}, error) {
	if devcontainer.ContainerImageBuildMode() != "devcontainer" {
		log.Println("Skipping since not in devcontainer mode.")
		return false, nil, nil, nil
	}

	// This buildpack always requires cpython
	reqs := []libcnb.BuildPlanRequire{{Name: cpython.BUILDPACK_NAME, Metadata: map[string]interface{}{
		"build":  true,
		"launch": true,
	}}}

	// Can be specified in project.toml or pack command line
	if os.Getenv("BP_CPYTHON_VERSION") != "" || os.Getenv("BP_PYTHON_UTILS") != "" {
		return true, reqs, nil, nil
	}

	// Look for requirements.txt, environment.yml in the root
	filesToCheck := []string{"requirements.txt"}
	for _, file := range filesToCheck {
		if _, err := os.Stat(filepath.Join(context.Application.Path, file)); err == nil {
			log.Println("Detection passed.")
			return true, reqs, nil, nil
		}
	}

	// See if runtime.txt is present, and contains a python reference
	if _, err := os.Stat(filepath.Join(context.Application.Path, "runtime.txt")); err == nil {
		contents, err := os.ReadFile(filepath.Join(context.Application.Path, "runtime.txt"))
		if err != nil {
			log.Fatal("Failed to read runtime.txt. ", err)
		}
		if strings.Contains(fmt.Sprint(contents), "python-") {
			log.Println("Detection passed.")
			return true, reqs, nil, nil
		}
	}

	log.Println("Python not detected.")
	return false, nil, nil, nil
}
