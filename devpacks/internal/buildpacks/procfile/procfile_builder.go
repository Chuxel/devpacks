package procfile

import (
	_ "embed"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/common/devcontainer"
)

type ProcfileBuilder struct {
	// Implements libcnb.Builder
	// Build(context libcnb.BuildContext) (libcnb.BuildResult, error)
}

func (builder ProcfileBuilder) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	buildMode := devcontainer.ContainerImageBuildMode()
	log.Println("Devpack path:", context.Buildpack.Path)
	log.Println("Application path:", context.Application.Path)
	log.Println("Build mode:", buildMode)
	log.Println("Number of plan entries:", len(context.Plan.Entries))
	log.Println("Env:", os.Environ())

	process := os.Getenv("BP_PROCESS_TYPE")
	if process == "" {
		process = "web"
	}

	content, err := os.ReadFile(filepath.Join(context.Application.Path, "Procfile"))
	if err != nil {
		log.Fatal("Failed to read Procfile:", err)
	}
	command := ""
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, process+":") {
			command = strings.TrimSpace(strings.TrimPrefix(line, process+":"))
		}
	}

	// TODO: Apply environment variables to command if web (see https://devcenter.heroku.com/articles/procfile#the-web-process-type)
	result := libcnb.NewBuildResult()
	result.Processes = append(result.Processes, libcnb.Process{
		Type:      "web",
		Command:   "bash",
		Arguments: []string{"-c", command},
		Default:   true,
	})

	return result, nil
}
