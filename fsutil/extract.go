package fsutil

import "github.com/pkg/errors"

var ErrUnknownExtension = errors.New("unknown file extension")

func Extract(extension, src, dest string) error {
	switch extension {
	case "zip":
		_, err := Unzip(src, dest)
		return err
	case "tgz":
		return ExtractTarGz(src, dest)
	case "txz":
		return ExtractTarXz(src, dest)
	default:
		return errors.Wrap(ErrUnknownExtension, extension)
	}
}
