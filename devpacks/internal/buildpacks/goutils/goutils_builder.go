package goutils

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
	"strings"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/buildpacks/base"
	"github.com/chuxel/devpacks/internal/common/devcontainer"
	"github.com/chuxel/devpacks/internal/common/utils"
)

//go:embed assets/devcontainer.json
var devcontainerJsonBytes []byte

type GoUtilsBuilder struct {
	// Implements base.DefaultBuilder

	// Build(context libcnb.BuildContext) (libcnb.BuildResult, error)
	// Name() string
	// NewLayerContributor(buildMode string, layerTypes libcnb.LayerTypes, context libcnb.BuildContext) libcnb.BaseLayerContributor
}

type GoUtilsLayerContributor struct {
	// Implements libcnb.LayerContributor

	// Contribute(context libcnb.ContributeContext) (libcnb.Layer, error)
	// Name() string

	LayerTypes libcnb.LayerTypes
	Context    libcnb.BuildContext
	BuildMode  string
}

func (builder GoUtilsBuilder) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	return base.DefaultBuild(builder, context)
}

// Implementation of base.BaseBuilder.Name
func (builder GoUtilsBuilder) Name() string {
	return BUILDPACK_NAME
}

// Implementation of base.BaseBuilder.NewLayerContributor
func (builder GoUtilsBuilder) NewLayerContributor(buildMode string, layerTypes libcnb.LayerTypes, context libcnb.BuildContext) libcnb.LayerContributor {
	return GoUtilsLayerContributor{BuildMode: buildMode, LayerTypes: layerTypes, Context: context}
}

// Implementation of libcnb.LayerContributor.Name
func (contrib GoUtilsLayerContributor) Name() string {
	return BUILDPACK_NAME
}

// Implementation of libcnb.LayerContributor.Contribute
func (contrib GoUtilsLayerContributor) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	var modList []string
	if os.Getenv("BP_GO_UTILS") != "" {
		modList = strings.Split(os.Getenv("BP_GO_UTILS"), " ")
	} else {
		modList = strings.Split(DEFAULT_GO_UTILS, " ")
	}

	// Generate and verify checksum
	hashCheckBytes := []byte(fmt.Sprint(os.Getenv("GO_VERSION"), " ", modList))
	hashGen := sha256.New()
	currentHash := base64.StdEncoding.EncodeToString(hashGen.Sum(hashCheckBytes))
	if layer.Metadata["sha256"] != nil {
		if currentHash == fmt.Sprint(layer.Metadata["sha256"]) {
			layer.LayerTypes = contrib.LayerTypes
			return layer, nil
		}
	}
	// Clean out layer folder in the event we invalidated the cache
	if err := os.RemoveAll(layer.Path); err != nil {
		log.Fatal("Unable to remove ", layer.Path, ". ", err)
	}

	// Make sure target path exists
	if err := os.MkdirAll(filepath.Join(layer.Path, "bin"), 0755); err != nil {
		log.Fatal("Unable to create layer folder: ", layer.Path, err)
	}
	// Write devcontainer.json in all cases since its quick and we can avoid doing a checksum when caching
	updatedBytes := bytes.ReplaceAll(devcontainerJsonBytes, []byte("{{layerDir}}"), []byte(layer.Path))
	if err := utils.WriteFile(path.Join(layer.Path, "devcontainer.json"), updatedBytes); err != nil {
		log.Fatal("Failed to write devcontainer.json: ", err)
	}

	// Install tools
	goTmp := filepath.Join("/tmp", "tool-tmp")
	os.Setenv("GOPATH", goTmp)
	os.Setenv("GOCACHE", filepath.Join(goTmp, "cache"))
	for _, mod := range modList {
		utils.ExecCmd(layer.Path, false, "go", "install", mod)
	}
	// Move binaries (only)
	utils.CpR(filepath.Join(goTmp, "bin"), layer.Path)

	// Update devcontainer.json search path for finalize buildpack to pull in properties
	layer.BuildEnvironment.Append(devcontainer.FINALIZE_JSON_SEARCH_PATH_ENV_VAR_NAME, string(filepath.ListSeparator), layer.Path)

	// Set the layer types based on what was set for the contributor
	layer.LayerTypes = contrib.LayerTypes
	layer.Metadata = map[string]interface{}{
		"sha256": currentHash,
	}

	return layer, nil
}
