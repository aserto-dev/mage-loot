package deps

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/aserto-dev/mage-loot/fsutil"
	"github.com/pkg/errors"
)

// DefLibDep makes sure a lib dependency is downloaded and unpacks it
func DefLibDep(name, url, sha string, options ...Option) {
	cmdRegisterMutex.Lock()
	defer cmdRegisterMutex.Unlock()

	if _, ok := config.Lib[name]; !ok {
		config.Lib[name] = &depDetails{Once: &sync.Once{}}
	}

	var ops depOptions

	config.Lib[name].Procure = func() {
		config.Lib[name].Once.Do(func() {
			for _, o := range options {
				o(&ops)
			}

			if len(ops.zipPaths) != 0 {
				downloadZippedLib(name, url, sha, ops.libPrefix, ops.zipPaths)
				return
			}

			if len(ops.tgzPaths) != 0 {
				downloadTgzLib(name, url, sha, ops.libPrefix, ops.tgzPaths)
				return
			}
		})
	}
}

func downloadZippedLib(name, url, sha, prefix string, patterns []string) {
	downloadLib(name, url, sha, "zip", prefix, patterns)
}

func downloadTgzLib(name, url, sha, prefix string, patterns []string) {
	downloadLib(name, url, sha, "tgz", prefix, patterns)
}

func downloadLib(name, url, sha, extension, prefix string, patterns []string) {
	filePath := tmpFile(name + "." + extension)
	defer os.RemoveAll(filepath.Dir(filePath))

	ui.Note().WithStringValue(extension, name).WithStringValue("url", url).Msg("Downloading ...")
	err := downloadFile(filePath, url)
	if err != nil {
		panic(errors.Wrap(err, "failed to download file"))
	}

	ui.Note().WithStringValue(extension, name).Msg("Checking signature ...")
	verifyFile(filePath, sha)

	libPath := LibDir()
	err = os.MkdirAll(libPath, 0700)
	if err != nil {
		panic(errors.Wrapf(err, "failed to create directory '%s'", libPath))
	}

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
			ui.Note().WithStringValue("  match", m).Msg("> lib file")
			relPath, err := filepath.Rel(unpackDir, m)
			if err != nil {
				panic(errors.Wrapf(err, "failed to get relative path for '%s'", m))
			}

			if prefix != "" {
				relPath, err = filepath.Rel(prefix, relPath)
				if err != nil {
					panic(errors.Wrapf(err, "failed to calculate relative path using prefix '%s' for path '%s'", prefix, relPath))
				}
			}

			dst := filepath.Join(libPath, relPath)
			dstDir := filepath.Dir(dst)
			err = os.MkdirAll(dstDir, 0700)
			if err != nil {
				panic(errors.Wrapf(err, "failed to create dir '%s'", dstDir))
			}

			err = os.Rename(m, dst)
			if err != nil {
				panic(errors.Wrapf(err, "failed to move '%s' to '%s'", m, dst))
			}
		}
	}
}
