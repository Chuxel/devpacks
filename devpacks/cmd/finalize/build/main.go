package main

import (
	"os"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/buildpacks/finalize"
)

func main() {
	args := []string{"build"}
	args = append(args, os.Args[1:]...)
	libcnb.Main(nil, finalize.FinalizeBuilder{}, libcnb.WithArguments(args))
}
