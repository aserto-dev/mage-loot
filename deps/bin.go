package deps

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/aserto-dev/mage-loot/fsutil"
	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
)

// DefBinDep makes sure a dependency is downloaded and makes it available as
// a runnable command.
func DefBinDep(name, url, version, sha string, options ...Option) {
	cmdRegisterMutex.Lock()
	defer cmdRegisterMutex.Unlock()

	if _, ok := config.Bin[name]; !ok {
		config.Bin[name] = &depDetails{Once: &sync.Once{}}
	}

	var ops depOptions

	config.Bin[name].Procure = func() {
		for _, o := range options {
			o(&ops)
		}

		binPath := binFilePath(name, version)
		config.Bin[name].Path = binPath

		exists, err := fsutil.FileExists(binPath)
		if err != nil {
			panic(errors.Wrapf(err, "failed to determine if bin '%s' exists", binPath))
		}
		if exists {
			return
		}

		config.Bin[name].Once.Do(func() {
			if len(ops.zipPaths) != 0 {
				downloadZippedBin(name, url, version, sha, ops.zipPaths)
				return
			}

			if len(ops.tgzPaths) != 0 {
				downloadTgzBin(name, url, version, sha, ops.tgzPaths)
				return
			}

			// Default to a simple binary
			downloadBinary(name, url, version, sha)
		})
	}
}

// BinDep returns a command for running a binary dependency.
// Its output is sent to stdout.
func BinDep(name string) func(...string) error {
	def := config.Bin[name]

	if def == nil {
		panic(errors.Errorf("didn't find a binary dependency named '%s'", name))
	}

	return func(args ...string) error {
		def.Procure()
		return sh.RunV(def.Path, args...)
	}
}

// BinDepWithEnv returns a command for running a binary dependency.
// It accepts an env map for the new process. Its output is sent to stdout.
func BinDepWithEnv(env map[string]string, name string) func(...string) error {
	def := config.Bin[name]

	if def == nil {
		panic(errors.Errorf("didn't find a binary dependency named '%s'", name))
	}

	return func(args ...string) error {
		def.Procure()
		return sh.RunWithV(env, name, args...)
	}
}

// BinDepOut returns a command for running a binary dependency.
// Its output is returned.
func BinDepOut(name string) func(...string) (string, error) {
	def := config.Bin[name]

	if def == nil {
		panic(errors.Errorf("didn't find a binary dependency named '%s'", name))
	}

	return func(args ...string) (string, error) {
		def.Procure()
		return sh.Output(def.Path, args...)
	}
}

func BinPath(name string) string {
	def := config.Bin[name]

	if def == nil {
		panic(errors.Errorf("didn't find a binary dependency named '%s'", name))
	}

	def.Procure()
	return def.Path
}

func downloadTgzBin(name, url, version, sha string, patterns []string) {
	downloadBin(name, url, version, sha, "tgz", patterns)
}

func downloadZippedBin(name, url, version, sha string, patterns []string) {
	downloadBin(name, url, version, sha, "zip", patterns)
}

func downloadBin(name, url, version, sha, extension string, patterns []string) {
	filePath := tmpFile(name + "." + extension)
	defer os.RemoveAll(filepath.Dir(filePath))
	versionedURL := versionTemplate(url, version)

	ui.Note().WithStringValue(extension, name).WithStringValue("url", versionedURL).Msg("Downloading ...")
	err := downloadFile(filePath, versionedURL)
	if err != nil {
		panic(errors.Wrap(err, "failed to download file"))
	}

	ui.Note().WithStringValue(extension, name).Msg("Checking signature ...")
	verifyFile(filePath, sha)

	unpackDir := mkTmpDir()
	defer os.RemoveAll(unpackDir)

	err = fsutil.Extract(extension, filePath, unpackDir)
	if err != nil {
		panic(errors.Wrapf(err, "failed to unpack '%s'", filePath))
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(unpackDir, pattern))
		if err != nil {
			panic(errors.Wrapf(err, "failed to glob using pattern '%s'", pattern))
		}

		for _, m := range matches {
			binPath := binFilePath(name, version)
			binDir := filepath.Dir(binPath)
			err = os.MkdirAll(binDir, 0700)
			if err != nil {
				panic(errors.Wrapf(err, "failed to create directory '%s'", binDir))
			}
			err = os.Rename(m, binPath)
			if err != nil {
				panic(errors.Wrapf(err, "failed to move binary '%s' to final location", m))
			}

			makeExe(binPath)
		}
	}
}

func downloadBinary(name, url, version, sha string) {
	filePath := binFilePath(name, version)
	versionedURL := versionTemplate(url, version)

	ui.Note().WithStringValue("bin", name).WithStringValue("url", versionedURL).Msg("Downloading ...")
	err := downloadFile(filePath, versionedURL)
	if err != nil {
		panic(errors.Wrap(err, "failed to download file"))
	}

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
