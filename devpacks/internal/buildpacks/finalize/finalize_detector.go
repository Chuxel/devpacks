package finalize

import (
	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/common"
)

type FinalizeDetector struct {
	// Implements libcnb.Detector
	// Detect(context libcnb.DetectContext) (libcnb.DetectResult, error)
}

func (detector FinalizeDetector) Detect(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	var result libcnb.DetectResult
	if common.ContainerImageBuildMode() == "devcontainer" {
		result.Plans = []libcnb.BuildPlan{
			{
				Provides: []libcnb.BuildPlanProvide{{Name: FINALIZE_BUILDPACK_NAME}},
				Requires: []libcnb.BuildPlanRequire{{Name: FINALIZE_BUILDPACK_NAME}},
			},
		}
	} else {
		result.Pass = false
	}
	return result, nil
}
