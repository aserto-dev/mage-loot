package fsutil

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
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

		if err != nil {
			return errors.Wrap(err, "failed to read for tar stream")
		}

		fpath := filepath.Join(dest, header.Name) // nolint:gosec // check ZipSlip below

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", fpath)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(fpath, 0755); err != nil {
				return errors.Wrap(err, "failed to create dir")
			}
		case tar.TypeReg:
			outFile, err := os.Create(fpath)
			if err != nil {
				return errors.Wrap(err, "failed to create file")
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
		default:
			return errors.Errorf(
				"unknown type: %s in %s", string(header.Typeflag), header.Name)
		}
	}

	return nil
}
