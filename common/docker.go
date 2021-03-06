package common

import (
	"time"

	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
)

// DockerImage builds the docker image for the project.
func DockerImage(repositoryAndTag string, args ...string) error {
	version, err := Version()
	if err != nil {
		return err
	}
	commit, err := Commit()
	if err != nil {
		return err
	}
	date := time.Now().UTC().Format(time.RFC3339)

	ui.Normal().
		WithStringValue("version", version).
		WithStringValue("commit", commit).
		WithStringValue("date", date).
		Msgf("Building docker image.")

	if repositoryAndTag == "" {
		return errors.Errorf("docker image repository and tag can't be empty")
	}

	return sh.RunWithV(map[string]string{
		"COMMIT":  commit,
		"VERSION": version,
	},
		"docker",
		append(
			append([]string{
				"build",
				"--ssh", "default", ".",
				"--build-arg", "COMMIT",
				"--build-arg", "GO_VERSION",
				"--build-arg", "VERSION",
			},
				args...),
			"-t", repositoryAndTag,
		)...,
	)
}
