//go:build windows

package common

import (
	"io/fs"
	"os"
)

// Windows lacks syscall.Stat_t and will throw an error with chown, so skip it
func SyncUIDGID(targetFile *os.File, sourceFileInfo fs.FileInfo) {
	return
}
