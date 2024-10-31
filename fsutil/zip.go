package fsutil

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

const (
	maxZippedFileSize = 5 * 1024 * 1024 * 1024
)

var (
	ErrIllegalFilePath = errors.New("illegal file path")
	ErrFileTooLarge    = errors.New("file too large")
)

// Unzip unzips an archive to a destination directory.
func Unzip(src, dest string) ([]string, error) {
	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name) // nolint:gosec // check ZipSlip below

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, errors.Wrap(ErrIllegalFilePath, fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			err = os.MkdirAll(fpath, os.ModePerm)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to create directory '%s'", fpath)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		if f.UncompressedSize64 > maxZippedFileSize {
			return nil, errors.Wrapf(ErrFileTooLarge, "max size: %d", maxZippedFileSize)
		}

		_, err = io.Copy(outFile, rc) // nolint:gosec // max file size checked above

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}

	return filenames, nil
}
