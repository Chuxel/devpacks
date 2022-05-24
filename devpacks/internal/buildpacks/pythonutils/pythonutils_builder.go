package pythonutils

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

type PythonUtilsBuilder struct {
	// Implements base.DefaultBuilder

	// Build(context libcnb.BuildContext) (libcnb.BuildResult, error)
	// Name() string
	// NewLayerContributor(buildMode string, layerTypes libcnb.LayerTypes, context libcnb.BuildContext) libcnb.BaseLayerContributor
}

type PythonUtilsLayerContributor struct {
	// Implements libcnb.LayerContributor

	// Contribute(context libcnb.ContributeContext) (libcnb.Layer, error)
	// Name() string

	LayerTypes libcnb.LayerTypes
	Context    libcnb.BuildContext
	BuildMode  string
}

func (builder PythonUtilsBuilder) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	return base.DefaultBuild(builder, context)
}

// Implementation of base.BaseBuilder.Name
func (builder PythonUtilsBuilder) Name() string {
	return BUILDPACK_NAME
}

// Implementation of base.BaseBuilder.NewLayerContributor
func (builder PythonUtilsBuilder) NewLayerContributor(buildMode string, layerTypes libcnb.LayerTypes, context libcnb.BuildContext) libcnb.LayerContributor {
	return PythonUtilsLayerContributor{BuildMode: buildMode, LayerTypes: layerTypes, Context: context}
}

// Implementation of libcnb.LayerContributor.Name
func (contrib PythonUtilsLayerContributor) Name() string {
	return BUILDPACK_NAME
}

// Implementation of libcnb.LayerContributor.Contribute
func (contrib PythonUtilsLayerContributor) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	var pkgList []string
	if os.Getenv("BP_PYTHON_UTILS") != "" {
		pkgList = strings.Split(os.Getenv("BP_PYTHON_UTILS"), " ")
	} else {
		pkgList = strings.Split(DEFAULT_PYTHON_UTILS, " ")
	}

	// Generate and verify checksum
	hashCheckBytes := []byte(fmt.Sprint(os.Getenv("PYTHON_VERSION"), " ", pkgList))
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
	if err := os.MkdirAll(layer.Path, 0755); err != nil {
		log.Fatal("Unable to create layer folder: ", layer.Path, err)
	}
	// Write devcontainer.json in all cases since its quick and we can avoid doing a checksum when caching
	updatedBytes := bytes.ReplaceAll(devcontainerJsonBytes, []byte("{{layerDir}}"), []byte(layer.Path))
	if err := utils.WriteFile(path.Join(layer.Path, "devcontainer.json"), updatedBytes); err != nil {
		log.Fatal("Failed to write devcontainer.json: ", err)
	}

	// Use pip to install pipx in a temporary spot we'll remove later
	pyTmp := filepath.Join(layer.Path, "tmp")
	os.Setenv("PYTHONUSERBASE", pyTmp)
	os.Setenv("PIP_CACHE_DIR", filepath.Join(pyTmp, "cache"))
	os.Setenv("PIPX_HOME", filepath.Join(layer.Path, "pipx"))
	os.Setenv("PIPX_BIN_DIR", filepath.Join(layer.Path, "bin"))
	pipx := filepath.Join(pyTmp, "bin", "pipx")
	utils.ExecCmd(layer.Path, false, "pip3", "install", "--disable-pip-version-check", "--no-cache-dir", "--user", "pipx")
	// Install packages using pipx
	utils.ExecCmd(layer.Path, false, pipx, "install", "--pip-args=--no-cache-dir", "pipx")
	for _, pkg := range pkgList {
		utils.ExecCmd(layer.Path, false, pipx, "install", "--pip-args=--no-cache-dir", pkg)
	}
	// Clear out temp folder
	if err := os.RemoveAll(pyTmp); err != nil {
		log.Fatal("Unable to remove tmp folder: ", pyTmp, err)
	}
	// Update devcontainer.json search path for finalize buildpack to pull in properties
	layer.BuildEnvironment.Append(devcontainer.FINALIZE_JSON_SEARCH_PATH_ENV_VAR_NAME, string(filepath.ListSeparator), layer.Path)

	// Set the layer types based on what was set for the contributor
	layer.LayerTypes = contrib.LayerTypes
	layer.Metadata = map[string]interface{}{
		"sha256": currentHash,
	}

	return layer, nil
}
