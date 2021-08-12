package fsutil

import (
	"os"
	"path/filepath"
	"strings"
)

// Inspired by https://github.com/yargevad/filepathx

// Globs represents one filepath glob, with its elements joined by "**".
type Globs []string

func Glob(pattern, exclude string) ([]string, error) {
	in, err := glob(pattern)
	if err != nil {
		return nil, err
	}

	var out []string
	if exclude != "" {
		out, err = glob(exclude)
		if err != nil {
			return nil, err
		}
	}

	if len(out) == 0 {
		return in, nil
	}

	result := make([]string, 0, len(in))
	for i := range in {
		filtered := false
		for j := range out {
			if in[i] == out[j] {
				filtered = true
				break
			}
		}
		if !filtered {
			result = append(result, in[i])
		}
	}

	return result, nil
}

func glob(pattern string) ([]string, error) {
	if !strings.Contains(pattern, "**") {
		// passthru to core package if no double-star
		return filepath.Glob(pattern)
	}

	return Globs(strings.Split(pattern, "**")).Expand()
}

func (globs Globs) Expand() ([]string, error) {
	var matches = []string{""} // accumulate here
	for _, glob := range globs {
		var hits []string
		var hitMap = map[string]bool{}
		for _, match := range matches {
			paths, err := filepath.Glob(match + glob)
			if err != nil {
				return nil, err
			}
			for _, path := range paths {
				err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					// save deduped match from current iteration
					if _, ok := hitMap[path]; !ok {
						hits = append(hits, path)
						hitMap[path] = true
					}
					return nil
				})
				if err != nil {
					return nil, err
				}
			}
		}
		matches = hits
	}

	// fix up return value for nil input
	if globs == nil && len(matches) > 0 && matches[0] == "" {
		matches = matches[1:]
	}

	return matches, nil
}
