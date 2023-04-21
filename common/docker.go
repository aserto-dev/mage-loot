package common

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/aserto-dev/mage-loot/deps"
	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
)

// DockerPush pushes an image
func DockerPush(existingImage, imageToPush string) error {
	UI.Normal().WithStringValue("tag", imageToPush).Msg("Tagging image.")

	err := sh.RunV("docker", "tag", existingImage, imageToPush)
	if err != nil {
		return errors.Wrap(err, "failed to tag image")
	}

	return sh.RunV("docker", "push", imageToPush)
}

// DockerTags uses sver to get a list of tags to be pushed.
// Expects env vars DOCKER_USERNAME and DOCKER_PASSWORD to be set.
func DockerTags(registry, image string) ([]string, error) {
	user := os.ExpandEnv("$DOCKER_USERNAME")
	password := os.ExpandEnv("$DOCKER_PASSWORD")

	out, err := deps.GoDepOutput("sver")("tags", "-u", user, "-p", password, "-s", registry, image)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read tags in registry: %s", out)
	}

	result := []string{}

	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		result = append(result, scanner.Text())
	}

	return result, nil
}

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

	platform := os.Getenv("BUILD_PLATFORM")
	if platform == "" {
		platform = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	}

	UI.Normal().
		WithStringValue("version", version).
		WithStringValue("commit", commit).
		WithStringValue("date", date).
		WithStringValue("platform", platform).
		Msgf("Building docker image.")

	if repositoryAndTag == "" {
		return errors.Errorf("docker image repository and tag can't be empty")
	}

	return sh.RunWithV(map[string]string{
		"COMMIT":  commit,
		"VERSION": version,
		"DATE":    date,
	},
		"docker",
		append(
			append([]string{
				"build",
				"--ssh", "default", ".",
				"--build-arg", "COMMIT",
				"--build-arg", "GO_VERSION",
				"--build-arg", "VERSION",
				"--build-arg", "DATE",
				"--platform", platform,
			},
				args...),
			"-t", repositoryAndTag,
		)...,
	)
}
