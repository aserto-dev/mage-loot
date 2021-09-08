package common

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	"github.com/aserto-dev/mage-loot/fsutil"
	"github.com/imdario/mergo"
)

func MergeOpenAPI(repo, outfile string, subServices []string) error {
	fsutil.EnsureDir(filepath.Dir(outfile))

	files := []string{}

	files = append(files, subServices...)

	sort.Strings(files)

	UI.Normal().
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
		UI.Normal().Compact().Msgf("<= %s", file)

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

	UI.Normal().Msgf("=> %s", outfile)

	err = enc.Encode(dst)
	return err
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
