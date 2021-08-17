package common

import (
	"strings"

	"github.com/aserto-dev/mage-loot/deps"
	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
)

func Commit() (string, error) {
	out, err := sh.Output("git", "rev-parse", "--short", "HEAD")
	if err != nil {
		return "", errors.Wrap(err, "please make sure this is a git repo - failed to determine commit")
	}

	return out, nil
}

func Version() (string, error) {
	out, err := deps.GoDepOutput("sver")()
	if err != nil {
		return "", errors.Wrap(err, "please make sure you have a valid tag - failed to determine version")
	}

	return out, nil
}

func NextVersion(part string) (string, error) {
	out, err := deps.GoDepOutput("sver")("--next", part)
	if err != nil {
		return "", errors.Wrap(err, "please make sure you have a valid tag - failed to determine version")
	}

	return out, nil
}

func IsDirty(version string) bool {
	return strings.Contains(version, "-dirty")
}
