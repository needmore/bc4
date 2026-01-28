package utils

import (
	"os"
	"runtime"
)

// AtomicRename renames src to dst atomically where possible.
// On POSIX systems os.Rename replaces an existing dst atomically.
// On Windows os.Rename fails when dst already exists, so we remove
// the destination first and then rename. This leaves a small window
// where dst doesn't exist, but it prevents silent failures that would
// strand the temp file and leave the old config in place.
func AtomicRename(src, dst string) error {
	if runtime.GOOS == "windows" {
		// Remove the destination so os.Rename can succeed.
		// Ignore the error – the file may not exist yet.
		_ = os.Remove(dst)
	}
	return os.Rename(src, dst)
}
