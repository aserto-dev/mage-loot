package fsutil

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/ulikunitz/xz"
)

const currentDirHeader string = "./"

func ExtractTarXz(src, dest string) error { //nolint:funlen,gocyclo // to be refactored
	hardLinks := make(map[string]string)

	xzStream, err := os.Open(src)
	if err != nil {
		return errors.Wrap(err, "failed to open file")
	}
	defer xzStream.Close()

	uncompressedStream, err := xz.NewReader(xzStream)
	if err != nil {
		return errors.Wrap(err, "failed to read xz stream")
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

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(fpath, 0755); err != nil {
				return errors.Wrap(err, "failed to create dir")
			}
			continue

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
			continue

		case tar.TypeSymlink:
			linkPath := filepath.Join(dest, header.Name) //nolint:gosec // required to establish links from archive
			if err := os.Symlink(header.Linkname, linkPath); err != nil {
				if os.IsExist(err) {
					continue
				}
				return err
			}
			continue

		case tar.TypeLink:
			/* Store details of hard links, which we process finally */
			linkPath := filepath.Join(dest, header.Linkname) //nolint:gosec // required to establish symlinks from archive
			linkPath2 := filepath.Join(dest, header.Name)    //nolint:gosec // required to establish symlinks from archive
			hardLinks[linkPath2] = linkPath
			continue

		default:
			return errors.Errorf(
				"unknown type: %s in %s", string(header.Typeflag), header.Name)
		}
	}

	for k, v := range hardLinks {
		if err := os.Link(v, k); err != nil {
			return err
		}
	}

	return nil
}
