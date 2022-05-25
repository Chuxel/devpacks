package pipinstall

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/buildpacks/base"
	"github.com/chuxel/devpacks/internal/common/devcontainer"
	"github.com/chuxel/devpacks/internal/common/utils"
)

//go:embed assets/devcontainer.json
var devcontainerJsonBytes []byte

type PipInstallBuilder struct {
	// Implements base.DefaultBuilder

	// Build(context libcnb.BuildContext) (libcnb.BuildResult, error)
	// Name() string
	// NewLayerContributor(buildMode string, layerTypes libcnb.LayerTypes, context libcnb.BuildContext) libcnb.BaseLayerContributor
}

type PipInstallLayerContributor struct {
	// Implements libcnb.LayerContributor

	// Contribute(context libcnb.ContributeContext) (libcnb.Layer, error)
	// Name() string

	LayerTypes libcnb.LayerTypes
	Context    libcnb.BuildContext
	BuildMode  string
}

func (builder PipInstallBuilder) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	return base.DefaultBuild(builder, context)
}

// Implementation of base.BaseBuilder.Name
func (builder PipInstallBuilder) Name() string {
	return BUILDPACK_NAME
}

// Implementation of base.BaseBuilder.NewLayerContributor
func (builder PipInstallBuilder) NewLayerContributor(buildMode string, layerTypes libcnb.LayerTypes, context libcnb.BuildContext) libcnb.LayerContributor {
	return PipInstallLayerContributor{BuildMode: buildMode, LayerTypes: layerTypes, Context: context}
}

// Implementation of libcnb.LayerContributor.Name
func (contrib PipInstallLayerContributor) Name() string {
	return BUILDPACK_NAME
}

// Implementation of libcnb.LayerContributor.Contribute
func (contrib PipInstallLayerContributor) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	// Just add a post create command in the devcontainer mode
	if devcontainer.ContainerImageBuildMode() == "devcontainer" {
		log.Println("Detected devcontainer build mode - adding devcontainer.json contents.")
		if err := os.MkdirAll(layer.Path, 0755); err != nil {
			log.Fatal("Unable to create layer folder. ", err)
		}
		updatedBytes := bytes.ReplaceAll(devcontainerJsonBytes, []byte("{{layerDir}}"), []byte(layer.Path))
		if err := utils.WriteFile(path.Join(layer.Path, "devcontainer.json"), updatedBytes); err != nil {
			log.Fatal("Unable to write devcontainer.json. ", err)
		}
		// Update devcontainer.json search path for finalize buildpack to pull in properties
		layer.BuildEnvironment.Append(devcontainer.FINALIZE_JSON_SEARCH_PATH_ENV_VAR_NAME, string(filepath.ListSeparator), layer.Path)
		layer.LayerTypes = libcnb.LayerTypes{
			Build:  true,
			Cache:  false,
			Launch: false,
		}
		return layer, nil
	}

	// Determine sha256 of requirements.txt
	requirementsTxtBytes, err := os.ReadFile(filepath.Join(contrib.Context.Application.Path, "requirements.txt"))
	if err != nil {
		log.Fatal("Failed to load requirements.txt. Be sure this file is in your repository. ", err)
	}
	hashGen := sha256.New()
	currentHash := base64.StdEncoding.EncodeToString(hashGen.Sum(requirementsTxtBytes))
	// Use sha256 to see if layer already exists and is the same so we can reuse
	if layer.Metadata["sha256"] != nil {
		if currentHash == fmt.Sprint(layer.Metadata["sha256"]) {
			log.Println("Reusing cached layer.")
			layer.LayerTypes = libcnb.LayerTypes{
				Build:  true,
				Cache:  true,
				Launch: true,
			}
			return layer, nil
		} else {
			// Otherwise remove layer node_modules since we'll need to recreate
			if err := os.RemoveAll(layer.Path); err != nil {
				log.Fatal("Failed to remove ", layer.Path, ". ", err)
			}
			if err := os.MkdirAll(layer.Path, 0755); err != nil {
				log.Fatal("Unable to create layer folder. ", err)
			}
		}
	}

	// Execute pip install
	cacheTmp := filepath.Join(layer.Path, "tmp-cache")
	os.Setenv("PYTHONUSERBASE", layer.Path)
	os.Setenv("PIP_CACHE_DIR", cacheTmp)
	utils.ExecCmd(contrib.Context.Application.Path, false, "pip3", "install", "--user", "-r", "requirements.txt")
	if err := os.RemoveAll(cacheTmp); err != nil {
		log.Fatal("Unable to remove tmp folder: ", cacheTmp, err)
	}

	// Add layer metadata (e.g. hash)
	layer.SharedEnvironment.Override("PYTHONUSERBASE", layer.Path)
	layer.Metadata = map[string]interface{}{
		"sha256": currentHash,
	}
	layer.LayerTypes = libcnb.LayerTypes{
		Build:  true,
		Cache:  true,
		Launch: true,
	}

	return layer, nil
}
