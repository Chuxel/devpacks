package nodejs_runtime

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"

	"github.com/blang/semver/v4"
	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/common"
)

//go:embed ../../../assets/nodejs_runtime/feature.json
var featureJsonBytes []byte

// Copy of https://nodejs.org/download/release/index.json
//go:embed ../../../assets/nodejs_runtime/index.json
var nodeIndexJsonBytes []byte

const NODEJS_RUNTIME_BUILDPACK_NAME = "nodejs-runtime"

type NodeJsRuntimeLayerContributor struct {
	// Implements libcnb.LayerContributor
	// Contribute(context libcnb.ContributeContext) (libcnb.Layer, error)
	// Name() string

	// Implements BaseLayerContributor
	// ApplyBuilderSettings(feature common.FeatureConfig, layerTypes libcnb.LayerTypes, context libcnb.BuildContext)

	layerTypes libcnb.LayerTypes
	context    libcnb.BuildContext
}

// Implementation of libcnb.LayerContributor.Name
func (contrib NodeJsRuntimeLayerContributor) Name() string {
	return NODEJS_RUNTIME_BUILDPACK_NAME
}

func (contrib NodeJsRuntimeLayerContributor) ApplyBuilderSettings(layerTypes libcnb.LayerTypes, context libcnb.BuildContext) {
	contrib.layerTypes = layerTypes
	contrib.context = context
}

// Implementation of libcnb.LayerContributor.Contribute
func (contrib NodeJsRuntimeLayerContributor) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	// Version of Node.js to download
	// Defaults to Node 18
	requestedVersion := "^18.0.0"

	//TODO: Set nodejs version based on BP_ env var if set

	packageJsonPath := path.Join(contrib.context.Application.Path, "package.json")
	// Get engine value for nodejs if it exists in package.json
	if _, err := os.Stat(packageJsonPath); err == nil {
		type PackageJson struct {
			Engines map[string]string `json:engines`
		}
		var packageJson PackageJson
		content, err := os.ReadFile(packageJsonPath)
		if err != nil {
			log.Fatal(err)
		}
		if err := json.Unmarshal(content, &packageJson); err != nil {
			log.Fatal(err)
		}
		candidateVersion, hasKey := packageJson.Engines["nodejs"]
		if hasKey {
			requestedVersion = candidateVersion
		}
	}

	// Determine real node version to acquire (since requested could be a semver range)
	nodeVersion := findRealNodeVersion(requestedVersion)

	// Check to see if a cached layer has already been restored and compare the version to see if we should recreate it
	cacheCheckFilePath := path.Join(layer.Path, "buildpack_cache_check.txt")
	if _, err := os.Stat(layer.Path); err != nil {
		installedVersionBytes, err := os.ReadFile(cacheCheckFilePath)
		if err != nil {
			log.Fatal(err)
		}
		if string(installedVersionBytes) != nodeVersion {
			donwloadAndUntarNode(nodeVersion, layer.Path)
		}
	} else {
		if err := os.MkdirAll(layer.Path, 0755); err != nil {
			log.Fatal(err)
		}
		donwloadAndUntarNode(nodeVersion, layer.Path)
	}
	common.WriteFile(cacheCheckFilePath, []byte(nodeVersion))

	// Augment and write feature.json file to path
	// **This buildpack doesn't need to modify, so just write**
	// featureConfig := common.FeatureConfig{}
	// featureConfig.LoadBytes(featureJsonBytes)
	common.WriteFile(path.Join(layer.Path, "feature.json"), featureJsonBytes)

	// Update lookup feature.json search path for finalize buildpack
	layer.BuildEnvironment.Append(common.FINALIZE_FEATURE_JSON_SEARCH_PATH_ENV_VAR_NAME, ":", layer.Path)

	// Set the layer types based on what was set for the contributor
	layer.LayerTypes = contrib.layerTypes
	layer.Metadata["build"] = layer.LayerTypes.Build
	layer.Metadata["launch"] = layer.LayerTypes.Launch
	layer.Metadata["cache"] = layer.LayerTypes.Cache
	layer.Metadata["node_version"] = nodeVersion

	return layer, nil
}

func donwloadAndUntarNode(nodeVersion string, targetPath string) {
	// Download file into memory so we can do a checksum
	dl_url := "https://nodejs.org/download/release/v" + nodeVersion + "node-" + nodeVersion + "-linux-" + runtime.GOARCH + ".tar.gz"
	response, err := http.Get(dl_url)
	if err != nil {
		log.Fatal(err)
	}
	if response.StatusCode != 200 {
		log.Fatal("Got status code", response.StatusCode, "for", dl_url)
	}
	tgzBytes, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	// TODO: Verify checksum and signature -- download SHASUM256.txt from the same spot

	// Untar into the target location
	common.UntarBytes(tgzBytes, targetPath)
}

func findRealNodeVersion(requestedVersion string) string {
	// Parse copy of https://nodejs.org/download/release/index.json
	type NodeIndexVersion struct {
		Version string `json:version`
	}
	nodeIndexVersions := []NodeIndexVersion{}
	if err := json.Unmarshal(nodeIndexJsonBytes, &nodeIndexVersions); err != nil {
		log.Fatal(err)
	}
	nodeVersions := semver.Versions{}
	for _, nodeIndexVersion := range nodeIndexVersions {
		version, err := semver.ParseTolerant(nodeIndexVersion.Version)
		if err != nil {
			log.Fatal(err)
		}
		nodeVersions = append(nodeVersions, version)
	}
	semver.Sort(nodeVersions)

	if requestedVersion != "latest" {
		expectedRange := semver.MustParseRange(requestedVersion)
		for _, nodeVersion := range nodeVersions {
			if expectedRange(nodeVersion) {
				return nodeVersion.FinalizeVersion()
			}
		}
		log.Fatal("Invalide node version", requestedVersion)
	}

	return nodeVersions[0].FinalizeVersion()
}
