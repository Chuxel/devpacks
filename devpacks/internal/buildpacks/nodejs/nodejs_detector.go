package nodejs

import (
	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/buildpacks/base"
)

type NodeJsRuntimeDetector struct {
	base.BaseDetector

	// DoDetect(context libcnb.DetectContext) (bool, map[string]interface{}, error)
	// Name() string
	// AlwaysPass() bool
}

func (detector NodeJsRuntimeDetector) Name() string {
	return NODEJS_RUNTIME_BUILDPACK_NAME
}

func (detector NodeJsRuntimeDetector) AlwaysPass() bool {
	return true
}

func (detector NodeJsRuntimeDetector) DoDetect(context libcnb.DetectContext) (bool, map[string]interface{}, error) {
	var metadata map[string]interface{}
	return true, metadata, nil
}
