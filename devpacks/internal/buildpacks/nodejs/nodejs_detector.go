package nodejs

import (
	"log"
	"os"
	"path"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/buildpacks/base"
)

type NodeJsRuntimeDetector struct {
	// Implements base.DefaultDetector

	// Detect(context libcnb.DetectContext) (libcnb.DetectResult, error)
	// DoDetect(context libcnb.DetectContext) (bool, map[string]interface{}, error)
	// Name() string
	// AlwaysPass() bool
}

func (detector NodeJsRuntimeDetector) Detect(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	return base.DefaultDetect(detector, context)
}

func (detector NodeJsRuntimeDetector) Name() string {
	return BUILDPACK_NAME
}

func (detector NodeJsRuntimeDetector) AlwaysPass() bool {
	return true
}

func (detector NodeJsRuntimeDetector) DoDetect(context libcnb.DetectContext) (bool, []libcnb.BuildPlanRequire, map[string]interface{}, error) {
	// Can be specified in project.toml or pack command line
	if os.Getenv("BP_NODE_VERSION") != "" {
		return true, nil, nil, nil
	}

	// Look for package json in the root
	if _, err := os.Stat(path.Join(context.Application.Path, "package.json")); err != nil {
		log.Println("No package.json found in ", context.Application.Path)
		return false, nil, nil, nil
	}

	log.Println("Detection passed.")
	return true, nil, nil, nil
}
