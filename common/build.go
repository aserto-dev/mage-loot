package common

import "github.com/aserto-dev/mage-loot/deps"

// BuildAll builds all binaries for all OSes and architectures, in preparation for a release.
func BuildAll(args ...string) error {
	return deps.GoDep("goreleaser")(append([]string{"build", "--rm-dist"}, args...)...)
}

// Build builds the project.
func Build(args ...string) error {
	return deps.GoDep("goreleaser")(append([]string{"build", "--rm-dist", "--snapshot", "--single-target"}, args...)...)
}
