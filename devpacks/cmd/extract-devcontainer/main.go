package main

import "github.com/chuxel/devpacks/internal/common"

func main() {
	imageName := "test_image"

	common.ExecCmd("", true, "docker", "inspect", imageName)
}
