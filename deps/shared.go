package deps

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

const (
	externalDir = ".ext"
	binDir      = "bin"
	libDir      = "lib"
	tmpDir      = "tmp"
)

var (
	currentDir string
)

type depOptions struct {
	zipPath string
	tgzPath string
}

// Option is a setting that changes the behavior
// of downloading and configuring a binary or a library
type Option func(*depOptions)

func init() {
	var err error
	currentDir, err = os.Getwd()
	if err != nil {
		panic(errors.Wrap(err, "failed to get working directory"))
	}
}

// downloadFile will download a url to a local file
func downloadFile(filePath string, url string) error {
	dir := filepath.Dir(filePath)
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		panic(errors.Wrapf(err, "failed to create dir '%s'", dir))
	}

	resp, err := http.Get(url)
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

func extTmpDir() string {
	return filepath.Join(currentDir, externalDir, tmpDir)
}

func tmpFile(name string) string {
	err := os.MkdirAll(extTmpDir(), 0700)
	if err != nil {
		panic(errors.Wrap(err, "failed to setup .ext/tmp dir"))
	}

	dir, err := os.MkdirTemp(extTmpDir(), "mageloot*")
	if err != nil {
		panic(errors.Wrap(err, "failed to setup temp file"))
	}

	return filepath.Join(dir, name)
}

func getTmpDir() string {
	err := os.MkdirAll(extTmpDir(), 0700)
	if err != nil {
		panic(errors.Wrap(err, "failed to setup .ext/tmp dir"))
	}

	dir, err := os.MkdirTemp(extTmpDir(), "mageloot*")
	if err != nil {
		panic(errors.Wrap(err, "failed to setup temp dir"))
	}

	return dir
}

func getDownloadURL(url, version string) string {
	type Version struct {
		Version string
	}
	v := Version{version}
	t := template.Must(template.New("url").Parse(url))

	var buf bytes.Buffer
	err := t.Execute(&buf, v)
	if err != nil {
		panic(errors.Wrap(err, "failed to render url template with version"))
	}

	return buf.String()
}