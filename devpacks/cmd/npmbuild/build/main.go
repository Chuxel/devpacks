package main

import (
	"os"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/buildpacks/npmbuild"
)

func main() {
	args := []string{"build"}
	args = append(args, os.Args[1:]...)
	libcnb.Main(nil, npmbuild.NpmBuildBuilder{}, libcnb.WithArguments(args))
}
