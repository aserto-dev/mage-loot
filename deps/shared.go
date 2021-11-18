package deps

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

const (
	externalDir = ".ext"
	binDir      = "bin"
	goBinDir    = "gobin"
	libDir      = "lib"
	tmpDir      = "tmp"
)

var (
	currentDir string
)

type depOptions struct {
	zipPaths  []string
	tgzPaths  []string
	txzPaths  []string
	libPrefix string
}

// Option is a setting that changes the behavior
// of downloading and configuring a binary or a library
type Option func(*depOptions)

// WithZipPaths tells us the binary or lib lives inside
// a zip archive
func WithZipPaths(paths ...string) Option {
	return func(o *depOptions) {
		o.zipPaths = paths
	}
}

// WithTGzPaths tells us the binary or lib lives inside
// a tarred and gzipped archive
func WithTGzPaths(paths ...string) Option {
	return func(o *depOptions) {
		o.tgzPaths = paths
	}
}

// WithTXzPaths tells us the binary or lib lives inside
// a tarred and xz utility compressed archive
func WithTXzPaths(paths ...string) Option {
	return func(o *depOptions) {
		o.txzPaths = paths
	}
}

// WithLibPrefix tells us we should remove the specified
// prefix from the lib paths.
// This option can use the {{.Version}} template.
func WithLibPrefix(prefix string) Option {
	return func(o *depOptions) {
		o.libPrefix = prefix
	}
}

// downloadFile will download a url to a local file
func downloadFile(filePath, url string) error {
	dir := filepath.Dir(filePath)
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		panic(errors.Wrapf(err, "failed to create dir '%s'", dir))
	}

	resp, err := http.Get(url) // nolint:gosec // urls come from a config file and are verified against a SHA256 signature
	if err != nil {
		panic(errors.Wrap(err, "http get request failed"))
	}
	defer resp.Body.Close()

	out, err := os.Create(filePath)
	if err != nil {
		panic(errors.Wrapf(err, "failed to create file '%s'", filePath))
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func verifyFile(filePath, sha string) {
	f, err := os.Open(filePath)
	if err != nil {
		panic(errors.Wrapf(err, "failed to open file '%s' for calculating sha", filePath))
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		panic(errors.Wrapf(err, "failed to calculate sha for file '%s'", filePath))
	}

	value := hex.EncodeToString(hasher.Sum(nil))

	if value != sha {
		panic(errors.Errorf("expected SHA256 for file '%s' to be '%s', not '%s'", filePath, value, sha))
	}
}

// BinDir returns the absolute path to the bin directory of tools
// that are not go.
func BinDir() string {
	return filepath.Join(currentDir, externalDir, binDir)
}

// LibDir returns the absolute path to the lib dir
func LibDir() string {
	return filepath.Join(currentDir, externalDir, libDir)
}

// LibDir returns the absolute path to the ext tmp dir
func ExtTmpDir() string {
	return filepath.Join(currentDir, externalDir, tmpDir)
}

// GoBinDir returns the absolute path to the bin directory of tools
func GoBinDir() string {
	return filepath.Join(currentDir, externalDir, goBinDir)
}

func tmpFile(name string) string {
	err := os.MkdirAll(ExtTmpDir(), 0700)
	if err != nil {
		panic(errors.Wrap(err, "failed to setup .ext/tmp dir"))
	}

	dir, err := os.MkdirTemp(ExtTmpDir(), "mageloot*")
	if err != nil {
		panic(errors.Wrap(err, "failed to setup temp file"))
	}

	return filepath.Join(dir, name)
}

func mkTmpDir() string {
	err := os.MkdirAll(ExtTmpDir(), 0700)
	if err != nil {
		panic(errors.Wrap(err, "failed to setup .ext/tmp dir"))
	}

	dir, err := os.MkdirTemp(ExtTmpDir(), "mageloot*")
	if err != nil {
		panic(errors.Wrap(err, "failed to setup temp dir"))
	}

	return dir
}
