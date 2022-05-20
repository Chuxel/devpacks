package nodejs

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/blang/semver/v4"
	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/buildpacks/base"
	"github.com/chuxel/devpacks/internal/common"
)

//go:embed assets/feature.json
var featureJsonBytes []byte

// Copy of https://nodejs.org/download/release/index.json
//go:embed assets/index.json
var nodeIndexJsonBytes []byte

type NodeJsRuntimeBuilder struct {
	// Implements base.DefaultBuilder

	// Build(context libcnb.BuildContext) (libcnb.BuildResult, error)
	// Name() string
	// NewLayerContributor(buildMode string, layerTypes libcnb.LayerTypes, context libcnb.BuildContext) libcnb.BaseLayerContributor
}

type NodeJsRuntimeLayerContributor struct {
	// Implements libcnb.LayerContributor

	// Contribute(context libcnb.ContributeContext) (libcnb.Layer, error)
	// Name() string

	LayerTypes libcnb.LayerTypes
	Context    libcnb.BuildContext
	BuildMode  string
}

func (builder NodeJsRuntimeBuilder) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	return base.DefaultBuild(builder, context)
}

// Implementation of base.BaseBuilder.Name
func (builder NodeJsRuntimeBuilder) Name() string {
	return BUILDPACK_NAME
}

// Implementation of base.BaseBuilder.NewLayerContributor
func (builder NodeJsRuntimeBuilder) NewLayerContributor(buildMode string, layerTypes libcnb.LayerTypes, context libcnb.BuildContext) libcnb.LayerContributor {
	return NodeJsRuntimeLayerContributor{BuildMode: buildMode, LayerTypes: layerTypes, Context: context}
}

// Implementation of libcnb.LayerContributor.Name
func (contrib NodeJsRuntimeLayerContributor) Name() string {
	return BUILDPACK_NAME
}

// Implementation of libcnb.LayerContributor.Contribute
func (contrib NodeJsRuntimeLayerContributor) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {

	// Version of Node.js to download
	requestedVersion := "^18.1.0"

	// Can be specified in project.toml or pack command line
	if os.Getenv("BP_NODE_VERSION") != "" {
		requestedVersion = os.Getenv("BP_NODE_VERSION")
	} else {
		// Otherwise look for version in a few common files
		var candidateVersion string
		var found bool
		if candidateVersion, found = contrib.packageJsonVersion(); found {
			requestedVersion = candidateVersion
		} else if candidateVersion, found = contrib.versionInFile(".nvmrc"); found {
			requestedVersion = candidateVersion
		} else if candidateVersion, found = contrib.versionInFile(".node-version"); found {
			requestedVersion = candidateVersion
		}
	}

	// Determine real node version to acquire (since requested could be a semver range)
	nodeVersion := findRealNodeVersion(requestedVersion)

	installNode := true
	// Check to see if a cached layer has already been restored and compare the version to see if we should recreate it
	if layer.Metadata["node_version"] != nil {
		if nodeVersion != fmt.Sprint(layer.Metadata["node_version"]) {
			if err := os.RemoveAll(layer.Path); err != nil {
				log.Fatal("Unable to remove ", layer.Path, ". ", err)
			}
			installNode = true
		} else {
			log.Println("Reusing cached layer.")
			installNode = false
		}
	}

	if installNode {
		downloadAndUntarNode(nodeVersion, layer.Path)
		// Add NODE_VERSION env var
		layer.SharedEnvironment.Default("NODE_VERSION", nodeVersion)
		// Update lookup feature.json search path for finalize buildpack
		layer.BuildEnvironment.Append(common.FINALIZE_JSON_SEARCH_PATH_ENV_VAR_NAME, string(filepath.ListSeparator), layer.Path)
		// Set the layer types based on what was set for the contributor
	}

	layer.LayerTypes = contrib.LayerTypes
	layer.Metadata = map[string]interface{}{
		"node_version": nodeVersion,
	}
	// Write feature.json in all cases since its quick and we can avoid doing a checksum when caching
	common.WriteFile(path.Join(layer.Path, "feature.json"), featureJsonBytes)

	return layer, nil
}

func downloadAndUntarNode(nodeVersion string, targetPath string) {
	// Make sure target path exists
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		log.Fatal(err)
	}

	// Download file into memory so we can do a checksum
	dl_arch := runtime.GOARCH
	if dl_arch == "amd64" {
		dl_arch = "x64"
	}
	dl_url := "https://nodejs.org/download/release/v" + nodeVersion + "/node-v" + nodeVersion + "-linux-" + dl_arch + ".tar.gz"
	response, err := http.Get(dl_url)
	if err != nil {
		log.Fatal(err)
	}
	if response.StatusCode != 200 {
		log.Fatal("Got status code ", response.StatusCode, " for ", dl_url)
	}
	tgzBytes, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	// TODO: Verify checksum and signature -- download SHASUM256.txt from the same spot

	// Untar into the target location
	common.UntarBytes(tgzBytes, targetPath, 1)
}

func findRealNodeVersion(requestedVersion string) string {
	// Parse copy of https://nodejs.org/download/release/index.json
	type NodeIndexVersion struct {
		Version string
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
		expectedRange := common.NewSemverRange(requestedVersion)
		// Sorted in ascending order, so run through in reverse order to get the latest matching
		for i := len(nodeVersions) - 1; i >= 0; i-- {
			nodeVersion := nodeVersions[i]
			if expectedRange(nodeVersion) {
				return nodeVersion.FinalizeVersion()
			}
		}

		log.Fatal("Unable to match node version", requestedVersion)
	}

	return nodeVersions[nodeVersions.Len()-1].FinalizeVersion()
}

func (contrib NodeJsRuntimeLayerContributor) packageJsonVersion() (string, bool) {
	packageJsonPath := filepath.Join(contrib.Context.Application.Path, "package.json")
	// Get engine value for nodejs if it exists in package.json
	if _, err := os.Stat(packageJsonPath); err == nil {
		type PackageJson struct {
			Engines map[string]string
		}
		var packageJson PackageJson

		content, err := os.ReadFile(packageJsonPath)
		if err != nil {
			log.Fatal("Failed to read package.json. ", err)
		}
		if err := json.Unmarshal(content, &packageJson); err != nil {
			log.Fatal("Failed to parse package.json. ", err)
		}
		version, hasKey := packageJson.Engines["node"]
		return version, hasKey
	}

	return "", false
}

func (contrib NodeJsRuntimeLayerContributor) versionInFile(name string) (string, bool) {
	versionFilePath := filepath.Join(contrib.Context.Application.Path, name)
	// Get engine value for nodejs if it exists in package.json
	if _, err := os.Stat(versionFilePath); err == nil {
		content, err := os.ReadFile(versionFilePath)
		if err != nil {
			log.Fatal("Failed to read ", name, ". ", err)
		}
		if content[0] == 'v' {
			return fmt.Sprint(content[1:]), true
		}
		return fmt.Sprint(content), true
	}

	return "", false
}
