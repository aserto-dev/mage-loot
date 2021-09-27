package common

import (
	"path/filepath"

	"github.com/aserto-dev/mage-loot/deps"
)

// Test runs all tests in the project using gotestsum.
func Test(args ...string) error {
	UI.Normal().Msg("Running tests.")

	return deps.GoDep("gotestsum")(
		append([]string{"--format", "short-verbose", "--"},
			append(args,
				"-count=1", "-v", filepath.Join(cwd, "..."), "-coverprofile=cover.out", "-coverpkg=./...", filepath.Join(cwd, "..."))...,
		)...,
	)
}

// Lint runs linting against the project.
func Lint(args ...string) error {
	UI.Normal().Msg("Running lint.")

	return deps.GoDep("golangci-lint")(
		append([]string{"run"}, args...)...,
	)
}
