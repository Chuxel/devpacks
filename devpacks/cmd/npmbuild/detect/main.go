package main

import (
	"os"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/buildpacks/npmbuild"
)

func main() {
	args := []string{"detect"}
	args = append(args, os.Args[1:]...)
	libcnb.Main(npmbuild.NpmBuildDetector{}, nil, libcnb.WithArguments(args))
}
