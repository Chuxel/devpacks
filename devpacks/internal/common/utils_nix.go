//go:build !windows

package common

import (
	"io/fs"
	"os"
	"syscall"
)

func SyncUIDGID(targetFile *os.File, sourceFileInfo fs.FileInfo) {
	targetFile.Chown(int(sourceFileInfo.Sys().(*syscall.Stat_t).Uid), int(sourceFileInfo.Sys().(*syscall.Stat_t).Gid))
}
