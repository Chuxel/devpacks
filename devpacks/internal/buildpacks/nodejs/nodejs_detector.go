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
	return NODEJS_RUNTIME_BUILDPACK_NAME
}

func (detector NodeJsRuntimeDetector) AlwaysPass() bool {
	return true
}

func (detector NodeJsRuntimeDetector) DoDetect(context libcnb.DetectContext) (bool, map[string]interface{}, error) {
	var metadata map[string]interface{}

	//TODO: Add env var to do detect

	// Look for package json in the root
	if _, err := os.Stat(path.Join(context.Application.Path, "package.json")); err != nil {
		log.Println("No package.json found in ", context.Application.Path)
		return false, metadata, nil
	}

	log.Println("Detection passed.")
	return true, metadata, nil
}
