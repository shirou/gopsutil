// +build linux

package so

import (
	"os"
	"path/filepath"
	"strconv"
)

type callback func(path string, info os.FileInfo, mntNS ns)

// walker traverses all /proc/<PID> subdirectories executing the given callbackFn for each
type walker struct {
	procRoot   string
	callbackFn callback
}

func newWalker(procRoot string, callbackFn callback) *walker {
	return &walker{
		procRoot:   procRoot,
		callbackFn: callbackFn,
	}
}

func (w *walker) walk(path string, info os.FileInfo, err error) error {
	if err != nil {
		return filepath.SkipDir
	}

	if !info.IsDir() {
		return nil
	}

	// Check if we're in a /proc/<PID> directory
	_, err = strconv.Atoi(info.Name())
	if err != nil {
		/* We want to continue walking from /proc */
		if path == w.procRoot {
			return nil
		}
		return filepath.SkipDir
	}

	// Execute callback for this /proc/<PID> entry
	w.callbackFn(path, info, getMntNS(path))

	return filepath.SkipDir
}
