package base

import (
	"log"
	"os"

	"github.com/buildpacks/libcnb"
)

type DefaultDetector interface {
	libcnb.Detector

	Name() string
	AlwaysPass() bool
	DoDetect(context libcnb.DetectContext) (bool, []libcnb.BuildPlanRequire, map[string]interface{}, error)
}

// Implementation of libcnb.Detector.Detect
func DefaultDetect(detector DefaultDetector, context libcnb.DetectContext) (libcnb.DetectResult, error) {
	log.Println("Devpack path:", context.Buildpack.Path)
	log.Println("Application path:", context.Application.Path)
	log.Println("Env:", os.Environ())

	var result libcnb.DetectResult
	detected, reqs, metadata, err := detector.DoDetect(context)
	if err != nil {
		return result, err
	}
	if reqs == nil {
		reqs = []libcnb.BuildPlanRequire{}
	}

	result.Plans = []libcnb.BuildPlan{}
	if detected {
		result.Plans = append(result.Plans, libcnb.BuildPlan{
			Provides: []libcnb.BuildPlanProvide{{Name: detector.Name()}},
			Requires: append(reqs, libcnb.BuildPlanRequire{Name: detector.Name(), Metadata: metadata}),
		})
	} else if detector.AlwaysPass() {
		// Add a plan that either provides nothing (so we don't get an error about
		// no require for something provided) and one that provides itself.
		result.Plans = append(result.Plans,
			libcnb.BuildPlan{
				Provides: []libcnb.BuildPlanProvide{},
				Requires: []libcnb.BuildPlanRequire{},
			},
			libcnb.BuildPlan{
				Provides: []libcnb.BuildPlanProvide{{Name: detector.Name()}},
				Requires: reqs,
			},
		)
	}
	result.Pass = detector.AlwaysPass() || detected

	return result, nil
}
