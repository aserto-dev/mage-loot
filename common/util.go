package common

import (
	"os"
	"path/filepath"

	"github.com/aserto-dev/clui"
	"github.com/pkg/errors"
)

var (
	UI  = clui.NewUI()
	cwd string
)

func init() {
	var err error
	cwd, err = os.Getwd()
	if err != nil {
		panic(errors.Wrap(err, "failed to determine working directory"))
	}

	cwd, err = filepath.Abs(cwd)
	if err != nil {
		panic(errors.Wrap(err, "failed to determine absolute path for the working directory"))
	}
}

// WorkDir returns the current working directory
func WorkDir() string {
	return cwd
}
