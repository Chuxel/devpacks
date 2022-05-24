package main

import (
	"os"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/buildpacks/procfile"
)

func main() {
	args := []string{"detect"}
	args = append(args, os.Args[1:]...)
	libcnb.Main(procfile.ProcfileDetector{}, nil, libcnb.WithArguments(args))
}
