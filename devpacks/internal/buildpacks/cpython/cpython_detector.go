package cpython

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/buildpacks/base"
)

type CPythonDetector struct {
	// Implements base.DefaultDetector

	// Detect(context libcnb.DetectContext) (libcnb.DetectResult, error)
	// DoDetect(context libcnb.DetectContext) (bool, map[string]interface{}, error)
	// Name() string
	// AlwaysPass() bool
}

func (detector CPythonDetector) Detect(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	return base.DefaultDetect(detector, context)
}

func (detector CPythonDetector) Name() string {
	return BUILDPACK_NAME
}

func (detector CPythonDetector) AlwaysPass() bool {
	return true
}

func (detector CPythonDetector) DoDetect(context libcnb.DetectContext) (bool, []libcnb.BuildPlanRequire, map[string]interface{}, error) {
	// Can be specified in project.toml or pack command line
	if os.Getenv("BP_CPYTHON_VERSION") != "" {
		return true, nil, nil, nil
	}

	// Look for requirements.txt in the root - TODO: Others? e.g. any .py file?
	filesToCheck := []string{"requirements.txt"}
	for _, file := range filesToCheck {
		if _, err := os.Stat(filepath.Join(context.Application.Path, file)); err == nil {
			log.Println("Detection passed.")
			return true, nil, nil, nil
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
			return true, nil, nil, nil
		}
	}

	log.Println("Python not detected.")
	return false, nil, nil, nil
}
