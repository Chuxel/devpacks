package base

import (
	"log"
	"os"

	"github.com/buildpacks/libcnb"
)

type BaseDetector struct {
	// Implements libcnb.Detector
	// Detect(context libcnb.DetectContext) (libcnb.DetectResult, error)

	// Overridable methods
	// DoDetect(context libcnb.DetectContext) (bool, map[string]interface{}, error)
	// Name() string
	// AlwaysPass() bool
}

// Implementation of libcnb.Detector.Detect
func (detector BaseDetector) Detect(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	log.Println("Devpack path:", context.Buildpack.Path)
	log.Println("Application path:", context.Application.Path)
	log.Println("Env:", os.Environ())

	var result libcnb.DetectResult
	detected, metadata, err := detector.DoDetect(context)
	if err != nil {
		return result, err
	}
	// Add a plan that either provides nothing (so we don't get an error about
	// no require for something provided) and one that provides itself.
	result.Plans = []libcnb.BuildPlan{
		libcnb.BuildPlan{
			Provides: []libcnb.BuildPlanProvide{},
			Requires: []libcnb.BuildPlanRequire{},
		},
	}
	if detected {
		result.Plans = append(result.Plans,
			libcnb.BuildPlan{
				Provides: []libcnb.BuildPlanProvide{libcnb.BuildPlanProvide{Name: detector.Name()}},
				Requires: []libcnb.BuildPlanRequire{libcnb.BuildPlanRequire{Name: detector.Name(), Metadata: metadata}},
			},
		)

	} else if detector.AlwaysPass() {
		result.Plans = append(result.Plans,
			libcnb.BuildPlan{
				Provides: []libcnb.BuildPlanProvide{libcnb.BuildPlanProvide{Name: detector.Name()}},
				Requires: []libcnb.BuildPlanRequire{},
			},
		)
	}

	result.Pass = detector.AlwaysPass() || detected

	return result, nil
}

// Intended to be overridden
func (detector BaseDetector) Name() string {
	return "base"
}

// Intended to be overridden
func (detector BaseDetector) AlwaysPass() bool {
	return true
}

// Intended to be overridden
func (detector BaseDetector) DoDetect(context libcnb.DetectContext) (bool, map[string]interface{}, error) {
	var metadata map[string]interface{}
	return true, metadata, nil
}
