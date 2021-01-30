// +build mage

package main

import (
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var (
	gogo      = sh.RunCmd("go")
	goTestSum = sh.RunCmd("gotestsum")

	cwd = ""
)

func gogoWith(env map[string]string, args ...string) error {
	return sh.RunWithV(env, "go", args...)
}

func dockerWith(env map[string]string, args ...string) error {
	return sh.RunWithV(env, "docker", args...)
}

func init() {
	// We want to use Go 1.11 modules even if the source lives inside GOPATH.
	// The default is "auto".
	os.Setenv("GO111MODULE", "on")

	// Set private repositories
	os.Setenv("GOPRIVATE", "github.com/aserto-dev")

	// Disable cgo
	os.Setenv("CGO_ENABLED", "0")

	// Determine current directory
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err.Error())
	}
	cwd = currentDir
}

// Deps installs dependency tools for the project
func Deps() error {
	color.Cyan("Installing dependencies")

	deps := []string{
		"gotest.tools/gotestsum",
		"github.com/golangci/golangci-lint/cmd/golangci-lint",
	}

	for _, dep := range deps {
		color.Cyan(">> %s", dep)

		err := gogo("install", "-i", dep)
		if err != nil {
			return err
		}
	}

	return nil
}

// Test runs all tests in the project using gotestsum.
func Test() error {
	color.Magenta("Running tests.")

	return goTestSum("--format", "short-verbose",
		"--", "-count=1", "-v", filepath.Join(cwd, "..."),
		"-coverprofile=cover.out", "-coverpkg=./...",
		filepath.Join(cwd, "..."))
}

// Lint runs linting against the project.
func Lint() error {
	color.Blue("Running lint.")

	return sh.RunV("golangci-lint", "run")
}

// All runs all targets in the appropriate order.
// The targets are run in the following order:
// deps, lint, test
func All() error {
	mg.SerialDeps(Deps, Lint, Test)
	return nil
}
