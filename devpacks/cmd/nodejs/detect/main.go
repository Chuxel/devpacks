package main

import (
	"os"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/buildpacks/nodejs"
)

func main() {
	args := []string{"detect"}
	args = append(args, os.Args[1:]...)
	libcnb.Main(nodejs.NodeJsRuntimeDetector{}, nil, libcnb.WithArguments(args))
}
