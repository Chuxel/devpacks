package main

import (
	"os"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/buildpacks/pythonutils"
)

func main() {
	args := []string{"detect"}
	args = append(args, os.Args[1:]...)
	libcnb.Main(pythonutils.PythonUtilsDetector{}, nil, libcnb.WithArguments(args))
}
