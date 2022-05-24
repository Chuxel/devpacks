package npmbuild

import (
	_ "embed"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/buildpacks/base"
	"github.com/chuxel/devpacks/internal/common/utils"
)

type NpmBuildBuilder struct {
	// Implements base.DefaultBuilder

	// Build(context libcnb.BuildContext) (libcnb.BuildResult, error)
	// Name() string
	// NewLayerContributor(buildMode string, layerTypes libcnb.LayerTypes, context libcnb.BuildContext) libcnb.BaseLayerContributor
}

type NpmBuildLayerContributor struct {
	// Implements libcnb.LayerContributor

	// Contribute(context libcnb.ContributeContext) (libcnb.Layer, error)
	// Name() string

	LayerTypes libcnb.LayerTypes
	Context    libcnb.BuildContext
	BuildMode  string
}

func (builder NpmBuildBuilder) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	return base.DefaultBuild(builder, context)
}

// Implementation of base.BaseBuilder.Name
func (builder NpmBuildBuilder) Name() string {
	return BUILDPACK_NAME
}

// Implementation of base.BaseBuilder.NewLayerContributor
func (builder NpmBuildBuilder) NewLayerContributor(buildMode string, layerTypes libcnb.LayerTypes, context libcnb.BuildContext) libcnb.LayerContributor {
	return NpmBuildLayerContributor{BuildMode: buildMode, LayerTypes: layerTypes, Context: context}
}

// Implementation of libcnb.LayerContributor.Name
func (contrib NpmBuildLayerContributor) Name() string {
	return BUILDPACK_NAME
}

// Implementation of libcnb.LayerContributor.Contribute
func (contrib NpmBuildLayerContributor) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	// TODO: Implement caching scheme, archive off copies - right now this is dumb and just invokes npm run build

	// Execute npm install
	utils.ExecCmd(contrib.Context.Application.Path, false, "npm", "run", "build")

	// Only keep the layer around for caching purposes since the
	// node_modules folder is in the workspace folder in this scenario
	layer.LayerTypes = libcnb.LayerTypes{
		Build:  true,
		Cache:  true,
		Launch: true,
	}

	return layer, nil
}
