package common

import (
	"encoding/json"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"

	"github.com/aserto-dev/mage-loot/fsutil"
	"github.com/imdario/mergo"
	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

func CreateOpenAPI(repo, outfile string, subServices []string) error {
	fsutil.EnsureDir(filepath.Dir(outfile))

	jsonStr, err := sh.Output("go", "mod", "edit", "-json")
	if err != nil {
		return err
	}

	version := gjson.Get(jsonStr, "Require.#(Path==\""+repo+"\").Version")

	goModCache, err := sh.Output("go", "env", "GOMODCACHE")
	if err != nil {
		return err
	}

	files := []string{}
	for _, s := range subServices {
		files = append(files, path.Join(goModCache, repo+"@"+version.String(), s))
	}

	sort.Strings(files)

	ui.Normal().
		Msg("openapi-create")

	return merge(files, outfile)
}

func MergeOpenAPI(repo, outfile string, subServices []string) error {
	fsutil.EnsureDir(filepath.Dir(outfile))

	files := []string{}

	files = append(files, subServices...)

	sort.Strings(files)

	ui.Normal().
		Msg("openapi-merge")

	return merge(files, outfile)
}

func merge(files []string, outfile string) error {
	var (
		src map[string]interface{}
		dst = make(map[string]interface{})
		err error
	)

	for _, file := range files {
		ui.Normal().Compact().Msgf("<= %s", file)

		src, err = loadFile(file)
		if err != nil {
			return err
		}

		if err := mergo.Merge(&dst, src); err != nil {
			return err
		}
	}

	w, err := os.Create(outfile)
	if err != nil {
		return err
	}
	defer w.Close()

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(true)

	ui.Normal().Msgf("=> %s", outfile)

	if err := enc.Encode(dst); err != nil {
		return err
	}
	return nil
}

func loadFile(filePath string) (map[string]interface{}, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var data map[string]interface{}

	dec := json.NewDecoder(f)

	if err := dec.Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

func CopyOpenAPI(repo, service, outfile string) error {
	jsonStr, err := sh.Output("go", "mod", "edit", "-json")
	if err != nil {
		return err
	}

	version := gjson.Get(jsonStr, "Require.#(Path==\""+repo+"\").Version")

	goModCache, err := sh.Output("go", "env", "GOMODCACHE")
	if err != nil {
		return err
	}

	filepath := path.Join(goModCache, repo+"@"+version.String(), "openapi", service, "openapi.json")

	return copyFile(filepath, outfile, true)
}

func copyFile(src, dst string, overwrite bool) error {
	const bufferSize int64 = 1024

	srcFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !srcFileStat.Mode().IsRegular() {
		return errors.Errorf("%s is not a regular file", src)
	}

	reader, err := os.Open(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	_, err = os.Stat(dst)
	if err == nil && !overwrite {
		return errors.Errorf("File %s already exists", dst)
	}

	writer, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer writer.Close()

	buf := make([]byte, bufferSize)
	for {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		if _, err := writer.Write(buf[:n]); err != nil {
			return err
		}
	}
	return err
}
