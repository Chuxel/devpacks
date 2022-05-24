package main

import (
	"os"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/buildpacks/procfile"
)

func main() {
	args := []string{"build"}
	args = append(args, os.Args[1:]...)
	libcnb.Main(nil, procfile.ProcfileBuilder{}, libcnb.WithArguments(args))
}
