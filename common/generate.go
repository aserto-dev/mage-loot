package common

import (
	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
)

// Generate runs go generate for the specified paths.
// If no paths are used, it generates for './...'.
func Generate(paths ...string) error {
	ui.Normal().Msg("Generating code.")

	if len(paths) == 0 {
		if err := sh.RunV("go", "generate", "./..."); err != nil {
			return errors.Wrap(err, "failed to run go generate for path './...'")
		}

		return nil
	}

	for _, path := range paths {
		if err := sh.RunV("go", "generate", path); err != nil {
			return errors.Wrapf(err, "failed to run go generate '%s'", path)
		}
	}

	return nil
}
