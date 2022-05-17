package finalize

import (
	"os"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/common"
)

type FinalizeDetector struct {
	// Implements libcnb.Detector
	// Detect(context libcnb.DetectContext) (libcnb.DetectResult, error)
}

func (detector FinalizeDetector) Detect(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	var result libcnb.DetectResult
	if common.ContainerImageBuildMode() == "devcontainer" && os.Getenv(common.FINALIZE_JSON_SEARCH_PATH_ENV_VAR_NAME) != "" {
		result.Plans = []libcnb.BuildPlan{
			{
				Provides: []libcnb.BuildPlanProvide{{Name: FINALIZE_BUILDPACK_NAME}},
				Requires: []libcnb.BuildPlanRequire{{Name: FINALIZE_BUILDPACK_NAME}},
			},
		}
		result.Pass = true
	} else {
		result.Pass = false
	}
	return result, nil
}
