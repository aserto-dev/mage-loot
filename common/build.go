package common

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"time"

	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
)

const (
	osLinux             = "linux"
	osWindows           = "windows"
	osDarwin            = "darwin"
	windowsBinExtension = ".exe"
)

var (
	// Architectures is a list of architectures to build binaries for.
	Architectures = []string{"amd64", "arm"}
	// OSList is a list of all OSes to build binaries for.
	OSList = []string{osLinux, osWindows, osDarwin}
)

// BuildAll builds all binaries for all OSes and architectures.
func BuildAll(args ...string) error {
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
		Msgf("Will build all commands.")

	cmds, err := ioutil.ReadDir("cmd")
	if err != nil {
		return errors.Wrap(err, "failed to read contents of './cmd' dir")
	}

	for _, c := range cmds {
		for _, a := range Architectures {
			for _, o := range OSList {
				ui.Normal().
					WithStringValue("os", o).
					WithStringValue("arch", a).
					WithStringValue("cmd", c.Name()).
					Msg("Building.")

				out := filepath.Join(cwd, "bin", fmt.Sprintf("%s-%s", o, a), c.Name())
				if o == osWindows {
					out += windowsBinExtension
				}

				err := sh.RunWithV(map[string]string{
					"GOOS":   o,
					"GOARCH": a,
				},
					"go",
					append(
						append([]string{"build"}, args...),
						[]string{"-o", out, filepath.Join(cwd, "cmd", c.Name())}...,
					)...,
				)

				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Build builds the project.
func Build(args ...string) error {
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
		Msgf("Building.")

	cmds, err := ioutil.ReadDir("cmd")
	if err != nil {
		return errors.Wrap(err, "failed to read contents of './cmd' dir")
	}

	for _, c := range cmds {
		out := filepath.Join(cwd, "bin", fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH), c.Name())
		if runtime.GOOS == osWindows {
			out += windowsBinExtension
		}

		err := sh.RunV(
			"go",
			append(
				append([]string{"build"}, args...),
				[]string{"-o", out, filepath.Join(cwd, "cmd", c.Name())}...,
			)...,
		)

		if err != nil {
			return err
		}
	}

	return nil
}
