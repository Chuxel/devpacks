package cpython

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/buildpacks/base"
	"github.com/chuxel/devpacks/internal/common/actions"
	"github.com/chuxel/devpacks/internal/common/devcontainer"
	"github.com/chuxel/devpacks/internal/common/utils"
)

//go:embed assets/devcontainer.json
var devcontainerJsonBytes []byte

type CPythonBuilder struct {
	// Implements base.DefaultBuilder

	// Build(context libcnb.BuildContext) (libcnb.BuildResult, error)
	// Name() string
	// NewLayerContributor(buildMode string, layerTypes libcnb.LayerTypes, context libcnb.BuildContext) libcnb.BaseLayerContributor
}

type CPythonLayerContributor struct {
	// Implements libcnb.LayerContributor

	// Contribute(context libcnb.ContributeContext) (libcnb.Layer, error)
	// Name() string

	LayerTypes libcnb.LayerTypes
	Context    libcnb.BuildContext
	BuildMode  string
}

func (builder CPythonBuilder) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	return base.DefaultBuild(builder, context)
}

// Implementation of base.BaseBuilder.Name
func (builder CPythonBuilder) Name() string {
	return BUILDPACK_NAME
}

// Implementation of base.BaseBuilder.NewLayerContributor
func (builder CPythonBuilder) NewLayerContributor(buildMode string, layerTypes libcnb.LayerTypes, context libcnb.BuildContext) libcnb.LayerContributor {
	return CPythonLayerContributor{BuildMode: buildMode, LayerTypes: layerTypes, Context: context}
}

// Implementation of libcnb.LayerContributor.Name
func (contrib CPythonLayerContributor) Name() string {
	return BUILDPACK_NAME
}

// Implementation of libcnb.LayerContributor.Contribute
func (contrib CPythonLayerContributor) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {

	// Version to download
	requestedVersion := "latest"

	// Can be specified in project.toml or pack command line
	if os.Getenv("BP_PYTHON_VERSION") != "" {
		requestedVersion = os.Getenv("BP_CPYTHON_VERSION")
	} else {
		// Otherwise look for version in a few common files
		var candidateVersion string
		var found bool
		if candidateVersion, found = contrib.versionInFile("runtime.txt", "python-"); found {
			requestedVersion = candidateVersion
		}
		/* else if candidateVersion, found = contrib.versionInFile(".node-version"); found {
			requestedVersion = candidateVersion
		}*/
	}

	// Determine real node version to acquire (since requested could be a semver range)
	manifest := actions.NewVersionManifestFromUrl("https://raw.githubusercontent.com/actions/python-versions/main/versions-manifest.json")
	version := manifest.FindVersion(requestedVersion, true)

	install := true
	// Check to see if a cached layer has already been restored and compare the version to see if we should recreate it
	if layer.Metadata["python_version"] != nil {
		if version != fmt.Sprint(layer.Metadata["python_version"]) {
			if err := os.RemoveAll(layer.Path); err != nil {
				log.Fatal("Unable to remove ", layer.Path, ". ", err)
			}
			install = true
		} else {
			log.Println("Reusing cached layer.")
			install = false
		}
	}

	if install {
		dlUrl := manifest.FindDownloadUrl(version)
		tgzBytes := utils.DownloadBytesFromUrl(dlUrl)
		utils.UntarBytes(tgzBytes, layer.Path, 0)
		// Delete source tarball
		if err := os.Remove(filepath.Join(layer.Path, "Python-"+version+".tgz")); err != nil {
			log.Fatal("Unable to remove Python source code tgz. ", err)
		}

		// Fix #! paths for scripts - These are hard coded to expected Actions spot
		regexp := regexp.MustCompile("#!/opt/hostedtoolcache/Python/.*")
		bins, err := os.ReadDir(filepath.Join(layer.Path, "bin"))
		if err != nil {
			log.Fatal("Failed to read contents of bin folder. ", err)
		}
		for _, bin := range bins {
			binPath := filepath.Join(layer.Path, "bin", bin.Name())
			contents, err := os.ReadFile(binPath)
			if err != nil {
				log.Fatal("Failed to read file ", binPath, ". ", err)
			}
			if regexp.Match(contents) {
				updated := regexp.ReplaceAll(contents, []byte("#!"+filepath.Join(layer.Path, "bin", "python3")))
				utils.WriteFile(binPath, updated)
				if err != nil {
					log.Fatal("Failed to write file ", binPath, ". ", err)
				}
			}
		}

		// Add PYTHON_VERSION env var
		layer.SharedEnvironment.Default("PYTHON_VERSION", version)
		// Update lookup feature.json search path for finalize buildpack
		layer.BuildEnvironment.Append(devcontainer.FINALIZE_JSON_SEARCH_PATH_ENV_VAR_NAME, string(filepath.ListSeparator), layer.Path)
	}

	// Set the layer types based on what was set for the contributor
	layer.LayerTypes = contrib.LayerTypes
	layer.Metadata = map[string]interface{}{
		"python_version": version,
	}
	// Write feature.json in all cases since its quick and we can avoid doing a checksum when caching
	utils.WriteFile(path.Join(layer.Path, "devcontainer.json"), devcontainerJsonBytes)

	return layer, nil
}

func (contrib CPythonLayerContributor) versionInFile(name string, prefix string) (string, bool) {
	versionFilePath := filepath.Join(contrib.Context.Application.Path, name)
	// Get engine value for nodejs if it exists in package.json
	if _, err := os.Stat(versionFilePath); err == nil {
		content, err := os.ReadFile(versionFilePath)
		if err != nil {
			log.Fatal("Failed to read ", name, ". ", err)
		}
		lines := strings.Split(fmt.Sprint(content), "\n")
		for _, line := range lines {
			line := strings.TrimSpace(line)
			if strings.HasPrefix(line, prefix) {
				return strings.TrimPrefix(line, prefix), true
			}
		}
	}

	return "", false
}
