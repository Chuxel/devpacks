package main

import (
	"os"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/buildpacks/cpython"
)

func main() {
	args := []string{"detect"}
	args = append(args, os.Args[1:]...)
	libcnb.Main(cpython.CPythonDetector{}, nil, libcnb.WithArguments(args))
}
