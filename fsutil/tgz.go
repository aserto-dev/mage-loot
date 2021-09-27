package fsutil

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

func ExtractTarGz(src, dest string) error {
	gzipStream, err := os.Open(src)
	if err != nil {
		return errors.Wrap(err, "failed to open file")
	}
	defer gzipStream.Close()

	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return errors.Wrap(err, "failed to read gzip stream")
	}

	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		// Tars have a header for the current directory. Skip it.
		if header.Name == currentDirHeader {
			continue
		}

		if err != nil {
			return errors.Wrap(err, "failed to read for tar stream")
		}

		fpath := filepath.Join(dest, filepath.Clean(header.Name)) // nolint:gosec // check ZipSlip below

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", fpath)
		}

		err = createTarResource(header, fpath, tarReader)
		if err != nil {
			return err
		}

	}

	return nil
}

func createTarResource(header *tar.Header, fpath string, tarReader *tar.Reader) error {
	if err := os.MkdirAll(path.Dir(fpath), 0755); err != nil {
		return errors.Wrap(err, "failed to create dir")
	}
	switch header.Typeflag {
	case tar.TypeDir:
		if err := os.MkdirAll(fpath, 0755); err != nil {
			return errors.Wrap(err, "failed to create dir")
		}
	case tar.TypeReg:

		if err := writeTarFile(tarReader, fpath, header.Mode); err != nil {
			return err
		}

	case tar.TypeSymlink:
		if err := os.Symlink(header.Linkname, fpath); err != nil {
			return errors.Wrapf(err, "failed to create symlink %s to file %s", fpath, header.Linkname)
		}

	default:
		return errors.Errorf(
			"unknown type: %s in %s", string(header.Typeflag), header.Name)
	}
	return nil
}

func writeTarFile(tarReader *tar.Reader, fpath string, fileMode int64) error {
	outFile, err := os.Create(fpath)
	if err != nil {
		return errors.Wrap(err, "failed to create file")
	}
	err = outFile.Chmod(fs.FileMode(fileMode))
	if err != nil {
		return errors.Wrapf(err, "cannot change mode of file: %s", fpath)
	}

	totalRead := int64(0)
	for {
		n, err := io.CopyN(outFile, tarReader, 1024)
		totalRead += n
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}

	outFile.Close()
	return nil
}
