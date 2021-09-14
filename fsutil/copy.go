package fsutil

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// CopyDir recursively copies the src directory to the dest directory
func CopyDir(src, dest string) error {

	f, err := os.Open(src)
	if err != nil {
		return err
	}

	file, err := f.Stat()
	if err != nil {
		return err
	}
	if !file.IsDir() {
		return errors.New("source " + file.Name() + " is not a directory")
	}

	if _, err := os.Stat(dest); os.IsNotExist(err) {
		err = os.Mkdir(dest, file.Mode())
		if err != nil {
			return errors.Wrapf(err, "could not create destination directory: %s ", dest)
		}
	}

	files, err := ioutil.ReadDir(src)
	if err != nil {
		return errors.Wrapf(err, "could not read source directory: %s ", src)
	}

	for _, f := range files {

		if f.IsDir() {
			sourceDir := filepath.Join(src, f.Name())
			destDir := filepath.Join(dest, f.Name())
			err = CopyDir(sourceDir, destDir)
			if err != nil {
				return errors.Wrapf(err, "failed to copy %s to %s", sourceDir, destDir)
			}

		}

		if !f.IsDir() {
			sourceFile := filepath.Join(src, f.Name())
			content, err := ioutil.ReadFile(sourceFile)
			if err != nil {
				return errors.Wrapf(err, "could not read source file %s", sourceFile)

			}

			destFile := filepath.Join(dest, f.Name())
			err = ioutil.WriteFile(destFile, content, f.Mode())
			if err != nil {
				return errors.Wrapf(err, "could not write to destination file %s", destFile)

			}

		}

	}

	return nil
}
