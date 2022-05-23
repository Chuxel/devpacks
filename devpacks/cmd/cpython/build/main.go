package main

import (
	"os"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/buildpacks/cpython"
)

func main() {
	args := []string{"build"}
	args = append(args, os.Args[1:]...)
	libcnb.Main(nil, cpython.CPythonBuilder{}, libcnb.WithArguments(args))
}
