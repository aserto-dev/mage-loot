package deps

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/aserto-dev/mage-loot/fsutil"
	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
)

const (
	zipExt = ".zip"
	tgzExt = ".tgz"
	gzExt  = ".gz"
	txzExt = ".txz"
	xzExt  = ".xz"
)

// DefBinDep makes sure a dependency is downloaded and makes it available as
// a runnable command.
func DefBinDep(name, url, version, sha, entrypoint string, options ...Option) {
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

		// downloading an archive
		entrypointPath := filepath.Join(binPath, entrypoint)

		config.Bin[name].Path = entrypointPath

		exists, err := fsutil.DirExists(binPath)
		if err != nil {
			panic(errors.Wrapf(err, "failed to determine if bin '%s' exists", binPath))
		}
		if exists {
			return
		}

		config.Bin[name].Once.Do(func() {
			if len(ops.zipPaths) != 0 && path.Ext(url) == zipExt {
				downloadZippedBin(name, url, version, sha, ops.zipPaths)
				return
			}

			if len(ops.tgzPaths) != 0 && (path.Ext(url) == tgzExt || path.Ext(url) == gzExt) {
				downloadTgzBin(name, url, version, sha, ops.tgzPaths)
				return
			}

			if len(ops.txzPaths) != 0 && (path.Ext(url) == txzExt || path.Ext(url) == xzExt) {
				downloadTxzBin(name, url, version, sha, ops.txzPaths)
				return
			}

			// Default to a simple binary
			downloadBinary(name, entrypoint, url, version, sha)
		})

		makeExe(entrypointPath)
	}
}

// BinExec returns a command for running a binary dependency.
// Its stdout and stderr are pipeped to the given writers.
func BinExec(name string, stdout, stderr io.Writer) func(...string) error {
	def := config.Bin[name]

	if def == nil {
		panic(errors.Errorf("didn't find a binary dependency named '%s'", name))
	}

	return func(args ...string) error {
		if !skipProcurement {
			def.Procure()
		}

		_, err := sh.Exec(nil, stdout, stderr, name, args...)
		return err
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
		if !skipProcurement {
			def.Procure()
		}

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
		if !skipProcurement {
			def.Procure()
		}

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
		if !skipProcurement {
			def.Procure()
		}

		return sh.Output(def.Path, args...)
	}
}

// BinDepOutWithEnv returns a command for running a binary dependency.
// It accepts an env map for the new process. Its output is returned.
func BinDepOutWithEnv(env map[string]string, name string) func(...string) (string, error) {
	def := config.Bin[name]

	if def == nil {
		panic(errors.Errorf("didn't find a binary dependency named '%s'", name))
	}

	return func(args ...string) (string, error) {
		if !skipProcurement {
			def.Procure()
		}

		return sh.OutputWith(env, def.Path, args...)
	}
}

func BinPath(name string) string {
	def := config.Bin[name]

	if def == nil {
		panic(errors.Errorf("didn't find a binary dependency named '%s'", name))
	}

	if !skipProcurement {
		def.Procure()
	}

	return def.Path
}

func downloadTgzBin(name, url, version, sha string, patterns []string) {
	downloadBin(name, url, version, sha, "tgz", patterns)
}

func downloadTxzBin(name, url, version, sha string, patterns []string) {
	downloadBin(name, url, version, sha, "txz", patterns)
}

func downloadZippedBin(name, url, version, sha string, patterns []string) {
	downloadBin(name, url, version, sha, "zip", patterns)
}

func downloadBin(name, url, version, sha, extension string, patterns []string) {
	filePath := tmpFile(name + "." + extension)
	defer os.RemoveAll(filepath.Dir(filePath))

	ui.Note().WithStringValue(extension, name).WithStringValue("url", url).Msg("Downloading ...")
	err := downloadFile(filePath, url)
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
			binDir := binFilePath(name, version)
			err = os.MkdirAll(binDir, 0700)
			if err != nil {
				panic(errors.Wrapf(err, "failed to create directory '%s'", binDir))
			}
			binPath := filepath.Join(binDir, filepath.Base(m))

			err = os.Rename(m, binPath)
			if err != nil {
				panic(errors.Wrapf(err, "failed to move binary '%s' to final location", m))
			}
		}
	}
}

func downloadBinary(name, entrypoint, url, version, sha string) {
	dirPath := binFilePath(name, version)

	err := os.MkdirAll(dirPath, 0700)
	if err != nil {
		panic(errors.Wrap(err, "failed to create dir for binary"))
	}

	binPath := filepath.Join(dirPath, entrypoint)
	ui.Note().WithStringValue("bin", name).WithStringValue("url", url).Msg("Downloading ...")
	err = downloadFile(binPath, url)
	if err != nil {
		panic(errors.Wrap(err, "failed to download file"))
	}

	ui.Note().WithStringValue("bin", name).Msg("Checking signature ...")
	verifyFile(binPath, sha)

	makeExe(binPath)
}

func binFilePath(name, version string) string {
	return filepath.Join(BinDir(), name+"-"+version)
}

func makeExe(exePath string) {
	err := os.Chmod(exePath, 0700)
	if err != nil {
		panic(errors.Wrapf(err, "failed to chmod file '%s'", exePath))
	}
}
