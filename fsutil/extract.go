package fsutil

import "fmt"

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
		return fmt.Errorf("unknown file extension: %s", extension)
	}
}
