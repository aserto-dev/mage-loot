package deps

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/aserto-dev/clui"
	"github.com/aserto-dev/mage-loot/fsutil"
	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
)

var (
	cmdDownloadSync  = map[string]*sync.Once{}
	cmdRegisterMutex = &sync.Mutex{}
	ui               = clui.NewUI()
)

// WithZipPath tells us the binary or lib lives inside
// a zip archive
func WithZipPath(path string) Option {
	return func(o *depOptions) {
		o.zipPath = path
	}
}

// WithTGzPath tells us the binary or lib lives inside
// a tarred and gzipped archive
func WithTGzPath(path string) Option {
	return func(o *depOptions) {
		o.tgzPath = path
	}
}

// BinDep returns a binary dependency loaded from a Depfile
func BinDep(name string) func(...string) error {
	bin := config.Bin[name]

	if bin == nil {
		panic(errors.Errorf("didn't find a binary dependency named '%s'", name))
	}

	return bin
}

// DefBinDep makes sure a dependency is downloaded and makes it available as
// a runnable command.
func DefBinDep(name, url, version, sha string, options ...Option) func(...string) error {
	cmdRegisterMutex.Lock()
	defer cmdRegisterMutex.Unlock()

	if _, ok := cmdDownloadSync[name]; !ok {
		cmdDownloadSync[name] = &sync.Once{}
	}

	var ops depOptions
	return func(args ...string) error {
		for _, o := range options {
			o(&ops)
		}

		binPath := binFilePath(name, version)
		exists, err := fsutil.FileExists(binPath)
		if err != nil {
			panic(errors.Wrapf(err, "failed to determine if bin '%s' exists", binPath))
		}
		if exists {
			return sh.RunV(binPath, args...)
		}

		cmdDownloadSync[name].Do(func() {
			if ops.zipPath != "" {
				ui.Note().WithStringValue("Path in archive", ops.zipPath).Msg("Looking for binary inside zip archive...")
				downloadZip(name, url, version, sha, ops.zipPath)
				return
			}

			// Default to a simple binary
			downloadBinary(name, url, version, sha)
		})

		return sh.RunV(binPath, args...)
	}
}

func downloadZip(name, url, version, sha, zipPath string) {
	filePath := tmpFile(name + ".zip")
	defer os.RemoveAll(filepath.Dir(filePath))
	versionedURL := getDownloadURL(url, version)

	ui.Note().WithStringValue("zip", name).WithStringValue("url", versionedURL).Msg("Downloading ...")
	downloadFile(filePath, versionedURL)

	ui.Note().WithStringValue("zip", name).Msg("Checking signature ...")
	verifyFile(filePath, sha)

	unzipDir := getTmpDir()
	defer os.RemoveAll(unzipDir)

	_, err := fsutil.Unzip(filePath, unzipDir)
	if err != nil {
		panic(errors.Wrapf(err, "failed to unzip '%s'", filePath))
	}

	src := filepath.Join(unzipDir, zipPath)
	binPath := binFilePath(name, version)
	binDir := filepath.Dir(binPath)
	err = os.MkdirAll(binDir, 0700)
	if err != nil {
		panic(errors.Wrapf(err, "failed to create directory '%s'", binDir))
	}
	err = os.Rename(src, binPath)
	if err != nil {
		panic(errors.Wrapf(err, "failed to move binary '%s' to final location", src))
	}
	makeExe(binPath)
}

func downloadBinary(name, url, version, sha string) {
	filePath := binFilePath(name, version)
	versionedURL := getDownloadURL(url, version)

	ui.Note().WithStringValue("bin", name).WithStringValue("url", versionedURL).Msg("Downloading ...")
	downloadFile(filePath, versionedURL)

	ui.Note().WithStringValue("bin", name).Msg("Checking signature ...")
	verifyFile(filePath, sha)

	makeExe(filePath)
}

func binFilePath(name, version string) string {
	return filepath.Join(BinDir(), name+"-"+version)
}

func makeExe(path string) {
	err := os.Chmod(path, 0700)
	if err != nil {
		panic(errors.Wrapf(err, "failed to chmod file '%s'", path))
	}
}
