package common

import (
	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
)

func GitTag(tag string) error {
	if err := sh.RunV("git", "tag", tag); err != nil {
		return errors.Wrap(err, "failed to tag")
	}

	return nil
}

func GitPushTag(tag, remote string) error {
	if err := sh.RunV("git", "push", remote, tag); err != nil {
		return errors.Wrapf(err, "failed to push tag [%s] to remote [%s]", tag, remote)
	}

	return nil
}
