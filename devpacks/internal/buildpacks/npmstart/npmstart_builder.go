package npmstart

import (
	_ "embed"
	"log"
	"os"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/common"
)

type NpmStartBuilder struct {
	// Implements libcnb.Builder
	// Build(context libcnb.BuildContext) (libcnb.BuildResult, error)
}

func (builder NpmStartBuilder) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	buildMode := common.ContainerImageBuildMode()
	log.Println("Devpack path:", context.Buildpack.Path)
	log.Println("Application path:", context.Application.Path)
	log.Println("Build mode:", buildMode)
	log.Println("Number of plan entries:", len(context.Plan.Entries))
	log.Println("Env:", os.Environ())

	result := libcnb.NewBuildResult()

	result.Processes = append(result.Processes, libcnb.Process{
		Type:      "web",
		Command:   "npm",
		Arguments: []string{"start"},
		Default:   true,
	})

	return result, nil

}
