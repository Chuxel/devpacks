package npminstall

import (
	"crypto/sha256"
	_ "embed"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/buildpacks/base"
)

//go:embed assets/check-symlink.sh
var symlinkScript []byte

type NpmInstallBuilder struct {
	// Implements base.DefaultBuilder

	// Build(context libcnb.BuildContext) (libcnb.BuildResult, error)
	// Name() string
	// NewLayerContributor(buildMode string, layerTypes libcnb.LayerTypes, context libcnb.BuildContext) libcnb.BaseLayerContributor
}

type NpmInstallLayerContributor struct {
	// Implements libcnb.LayerContributor

	// Contribute(context libcnb.ContributeContext) (libcnb.Layer, error)
	// Name() string

	LayerTypes libcnb.LayerTypes
	Context    libcnb.BuildContext
	BuildMode  string
}

func (builder NpmInstallBuilder) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	return base.DefaultBuild(builder, context)
}

// Implementation of base.BaseBuilder.Name
func (builder NpmInstallBuilder) Name() string {
	return BUILDPACK_NAME
}

// Implementation of base.BaseBuilder.NewLayerContributor
func (builder NpmInstallBuilder) NewLayerContributor(buildMode string, layerTypes libcnb.LayerTypes, context libcnb.BuildContext) libcnb.LayerContributor {
	return NpmInstallLayerContributor{BuildMode: buildMode, LayerTypes: layerTypes, Context: context}
}

// Implementation of libcnb.LayerContributor.Name
func (contrib NpmInstallLayerContributor) Name() string {
	return BUILDPACK_NAME
}

// Implementation of libcnb.LayerContributor.Contribute
func (contrib NpmInstallLayerContributor) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	// Determine sha256 of package-lock.json
	packageLockBytes, err := os.ReadFile(filepath.Join(contrib.Context.Application.Path, "package-lock.json"))
	if err != nil {
		log.Fatal("Failed to load package-lock.json. Be sure this file is in your repository. ", err)
	}
	hashGen := sha256.New()
	currentHash := base64.StdEncoding.EncodeToString(hashGen.Sum(packageLockBytes))
	// If layer already exists, skip if sha256 is the same
	layerNodeModules := filepath.Join(layer.Path, "node_modules")
	if layer.Metadata["sha256"] != nil {
		if currentHash == fmt.Sprint(layer.Metadata["sha256"]) {
			return layer, nil
		} else {
			// Otherwise remove layer node_modules since we'll need to recreate
			if err := os.RemoveAll(layerNodeModules); err != nil {
				log.Fatal("Failed to remove ", layerNodeModules, ". ", err)
			}
		}
	}

	// Remove existing node_modules folder if found - we're going to symlink it
	appNodeModules := filepath.Join(contrib.Context.Application.Path, "node_modules")
	if _, err := os.Stat(appNodeModules); err != nil {
		if err := os.RemoveAll(appNodeModules); err != nil {
			log.Fatal("Failed to remove ", appNodeModules, ". ", err)
		}
	}

	// Execute npm install
	cmd := exec.Command("npm", "install")
	cmd.Env = os.Environ()
	writer := log.Writer()
	cmd.Stdout = writer
	cmd.Stderr = writer
	cmd.Dir = contrib.Context.Application.Path
	if err := cmd.Run(); err != nil {
		log.Fatal("npm install failed. ", err)
	} else if cmd.ProcessState.ExitCode() != 0 {
		log.Fatal("npm install failed with exit code ", cmd.ProcessState.ExitCode())
	}

	/*
		// Move node_modules to layer folder, symlink for other build steps
		if err := os.MkdirAll(layer.Path, 0755); err != nil {
			log.Fatal("Unable to create node_modules folder. ", err)
		}
		if err := os.Rename(appNodeModules, layerNodeModules); err != nil {
			log.Fatal("Unable to move node_modules into layer folder. ", err)
		}
		if err := os.Symlink(layerNodeModules, appNodeModules); err != nil {
			log.Fatal("Unable to create node_modules symlink. ", err)
		}
	*/

	// The symlink created here will only apply to the build image and we need it in the launch image.
	// So add a profile.d script that the launcher will fire to verify, but we need to add some paths
	script := fmt.Sprintf("#!/bin/bash\nWORKSPACE_FOLDER=\"%s\"\nNPM_INSTALL_LAYER=\"%s\"\n%s", contrib.Context.Application.Path, layer.Path, symlinkScript)
	layer.Profile.Add("check-symlink", script)

	layer.LayerTypes = libcnb.LayerTypes{
		Build:  true,
		Cache:  true,
		Launch: true,
	}
	// Add layer metadata (e.g. hash)
	layer.Metadata = map[string]interface{}{
		"sha256": currentHash,
	}

	return layer, nil
}