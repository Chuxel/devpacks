package main

import (
	"os"

	"github.com/buildpacks/libcnb"
	"github.com/chuxel/devpacks/internal/buildpacks/pythonutils"
)

func main() {
	args := []string{"build"}
	args = append(args, os.Args[1:]...)
	libcnb.Main(nil, pythonutils.PythonUtilsBuilder{}, libcnb.WithArguments(args))
}
