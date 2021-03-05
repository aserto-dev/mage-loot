// +build mage

package main

import (
	"os"

	"github.com/aserto-dev/mage-loot/common"
	"github.com/aserto-dev/mage-loot/deps"
	"github.com/magefile/mage/mg"
)

func init() {
	// Set private repositories
	os.Setenv("GOPRIVATE", "github.com/aserto-dev")
}

// Deps installs dependency tools for the project
func Deps() {
	deps.GetAllDeps()
}

// Test runs all tests in the project using gotestsum.
func Test() error {
	return common.Test()
}

// Lint runs linting against the project.
func Lint() error {
	return common.Lint()
}

// All runs all targets in the appropriate order.
// The targets are run in the following order:
// deps, lint, test
func All() error {
	mg.SerialDeps(Deps, Lint, Test)
	return nil
}
