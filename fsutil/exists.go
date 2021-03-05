package fsutil

import (
	"os"

	"github.com/pkg/errors"
)

// FileExists checks if a file exists
func FileExists(path string) (bool, error) {
	if info, err := os.Stat(path); err == nil {
		if info.IsDir() {
			return false, errors.New("not a file")
		}
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, errors.Wrapf(err, "failed to stat '%s'", path)
	}
}

// DirExists checks if a directory exists
func DirExists(path string) (bool, error) {
	if info, err := os.Stat(path); err == nil {
		if !info.IsDir() {
			return false, errors.New("not a dir")
		}
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, errors.Wrapf(err, "failed to stat '%s'", path)
	}
}

// EnsureDir makes sure a directory exists.
// It panics if something goes wrong.
func EnsureDir(path string) {
	err := os.MkdirAll(path, 0700)
	if err != nil {
		panic(errors.Wrapf(err, "failed to create dir '%s'", path))
	}
}
