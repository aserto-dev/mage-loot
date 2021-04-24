package fsutil

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"

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

		if err != nil {
			return errors.Wrap(err, "failed to read for tar stream")
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(filepath.Join(dest, header.Name), 0755); err != nil {
				return errors.Wrap(err, "failed to create dir")
			}
		case tar.TypeReg:
			outFile, err := os.Create(filepath.Join(dest, header.Name))
			if err != nil {
				return errors.Wrap(err, "failed to create file")
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return errors.Wrap(err, "failed write to file")
			}
			outFile.Close()
		default:
			return errors.Errorf(
				"unknown type: %s in %s", header.Typeflag, header.Name)
		}
	}

	return nil
}
