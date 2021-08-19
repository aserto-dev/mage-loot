package common

import (
	"os"
	"path/filepath"

	"github.com/aserto-dev/mage-loot/fsutil"
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

func SetupGitConfig(email, name string) error {
	sshDir := filepath.Join(os.ExpandEnv("$HOME"), ".ssh")
	err := os.MkdirAll(sshDir, 0700)
	if err != nil {
		return err
	}

	knownHostsFile := filepath.Join(sshDir, "known_hosts")
	knownHostsExists, err := fsutil.FileExists(knownHostsFile)
	if err != nil {
		return err
	}

	if !knownHostsExists {
		githubKeyscan, err := sh.Output("ssh-keyscan", "github.com")
		if err != nil {
			return err
		}
		err = os.WriteFile(knownHostsFile, []byte(githubKeyscan), 0600)
		if err != nil {
			return err
		}
	} else {
		UI.Normal().Msg("Known hosts SSH file already exists")
	}

	idRSAFile := filepath.Join(sshDir, "id_rsa")
	idRSAExists, err := fsutil.FileExists(idRSAFile)
	if err != nil {
		return err
	}

	if !idRSAExists {
		value, ok := os.LookupEnv("SSH_PRIVATE_KEY")
		if ok {
			UI.Normal().Msg("Setting up git SSH key.")

			err = os.WriteFile(idRSAFile, []byte(value), 0600)
			if err != nil {
				return err
			}

			err = sh.Run("ssh-add", idRSAFile)
			if err != nil {
				return err
			}
		}
	} else {
		UI.Normal().Msg("id_rsa file already exists")
	}

	existingEmailConfig, err := sh.Output("git", "config", "--global", "user.email")
	if err != nil {
		return err
	}
	if existingEmailConfig == "" {
		UI.Normal().Msg("Setting git config for email, name and ssh insteadOf http.")

		err := sh.Run("git", "config", "--global", "user.email", email)
		if err != nil {
			return err
		}
		err = sh.Run("git", "config", "--global", "user.email", name)
		if err != nil {
			return err
		}
		err = sh.Run("git", "config", "--global", `url."git@github.com:".insteadOf`, "https://github.com/")
		if err != nil {
			return err
		}

	} else {
		UI.Normal().Msg("Git already configured.")
	}

	return nil
}
