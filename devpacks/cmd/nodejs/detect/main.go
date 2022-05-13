package main

import (
	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/buildpacks/nodejs"
)

func main() {
	libcnb.Main(nodejs.NodeJsRuntimeDetector{}, nil)
}
