package common

import (
	"fmt"
	"os"

	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
)

// Generate runs go generate for the specified paths.
// If no paths are used, it generates for './...'.
func Generate(paths ...string) error {
	return GenerateWith(nil, paths...)
}

// Generate runs go generate for the specified paths using the specified generation tools.
// If no paths are used, it generates for './...'.
func GenerateWith(tools []string, paths ...string) error {
	UI.Normal().Msg("Generating code.")

	var env map[string]string

	if len(tools) > 0 {
		env = make(map[string]string)
		env["PATH"] = os.Getenv("PATH")
		for _, tool := range tools {
			env["PATH"] = fmt.Sprintf("%s%s%s", tool, string(os.PathListSeparator), env["PATH"])
		}
	}

	if len(paths) == 0 {
		if err := sh.RunWithV(env, "go", "generate", "./..."); err != nil {
			return errors.Wrap(err, "failed to run go generate for path './...'")
		}

		return nil
	}

	for _, path := range paths {
		if err := sh.RunWithV(env, "go", "generate", path); err != nil {
			return errors.Wrapf(err, "failed to run go generate '%s'", path)
		}
	}

	return nil
}
