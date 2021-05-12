package common

import (
	"encoding/json"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/imdario/mergo"
	"github.com/magefile/mage/sh"
	"github.com/tidwall/gjson"
)

func CreateOpenAPI(repo, outfile string, subServices []string) error {
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

	ui.Exclamation().
		WithStringValue("files", strings.Join(files, ", ")).
		WithStringValue("outfile", outfile).
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
