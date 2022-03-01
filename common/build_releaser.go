package common

import (
	"github.com/aserto-dev/mage-loot/deps"
)

// BuildAllReleaser builds all binaries for all OSes and architectures, in preparation for a release.
func BuildAllReleaser(args ...string) error {
	err := GitleaksCheck()
	if err != nil {
		return err
	}
	return deps.GoDep("goreleaser")(append([]string{"build", "--rm-dist"}, args...)...)
}

// BuildReleaser builds the project.
func BuildReleaser(args ...string) error {
	err := GitleaksCheck()
	if err != nil {
		return err
	}

	return deps.GoDep("goreleaser")(append([]string{"build", "--rm-dist", "--snapshot", "--single-target"}, args...)...)
}

// Release releases the project.
func Release(args ...string) error {
	return deps.GoDep("goreleaser")(append([]string{"release"}, args...)...)
}
