package buf

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/aserto-dev/mage-loot/fsutil"
	"gopkg.in/yaml.v2"
)

func GenerateDev(binFile string, protoDirs, deps []string) error {
	cleanup, err := CreateDevWorkspace(protoDirs, deps)
	if cleanup != nil {
		defer cleanup()
	}
	if err != nil {
		return err
	}

	if err := Lint(); err != nil {
		return err
	}

	if err := Build(binFile); err != nil {
		return err
	}

	if len(protoDirs) == 0 {
		return Run(AddArg("generate"))
	}

	for _, d := range protoDirs {
		err := Run(AddArg("generate"), AddArg("--path"), AddArg(d))
		if err != nil {
			return err
		}
	}

	return nil
}

func CreateDevWorkspace(protoDirs, dependencies []string) (func() error, error) {
	ui.Note().WithStringValue("local-deps-dir", "buf-tmp").Msg("Creating temporary workspace")

	devTmpDir := "buf-tmp"
	err := os.MkdirAll(devTmpDir, 0755)
	if err != nil {
		return nil, err
	}

	workspaceFileExists, err := fsutil.FileExists("buf.work.yaml")
	if err != nil {
		return nil, err
	}

	cleanup := func() error {
		ui.Note().WithStringValue("local-deps-dir", "buf-tmp").Msg("Cleaning up temporary workspace")

		err := os.RemoveAll(devTmpDir)
		if err != nil {
			return err
		}

		// If the workspace file did not exist before, remove it.
		if !workspaceFileExists {
			ui.Note().Msg("Removing buf.work.yaml")

			err = os.Remove("buf.work.yaml")
			if err != nil {
				return err
			}
		} else {
			// Otherwise, restore the original file.
			ui.Note().Msg("Restoring buf.work.yaml")
			err = os.Rename("buf.work.yaml.orig", "buf.work.yaml")
			if err != nil {
				return err
			}
		}

		return nil
	}

	// If the workspace file exists, move it to a backup.
	if workspaceFileExists {
		ui.Note().Msg("Backing up buf.work.yaml")
		err = os.Rename("buf.work.yaml", "buf.work.yaml.orig")
		if err != nil {
			return nil, err
		}
	}

	workspaces := struct {
		Version     string   `yaml:"version"`
		Directories []string `yaml:"directories"`
	}{
		"v1",
		[]string{},
	}

	// If the workspace file exists, load it
	if workspaceFileExists {
		ui.Note().Msg("Loading buf.work.yaml")

		workspaceContents, err := ioutil.ReadFile("buf.work.yaml.orig")
		if err != nil {
			return cleanup, err
		}
		err = yaml.Unmarshal(workspaceContents, &workspaces)
		if err != nil {
			return cleanup, err
		}
	} else {
		workspaces.Directories = append(workspaces.Directories, protoDirs...)
	}

	for _, d := range dependencies {
		tmpDir := filepath.Join(devTmpDir, filepath.Base(d))

		ui.Note().
			WithStringValue("local-deps-dir", d).
			WithStringValue("tmp-dir", tmpDir).
			Msg("Adding dependency to workspace")

		// Copy directory in a local tmp folder
		if err := fsutil.CopyDir(d, tmpDir); err != nil {
			return cleanup, err
		}

		// Add directory to workspaces
		workspaces.Directories = append(workspaces.Directories, tmpDir)
	}

	// Write workspaces to file
	ui.Note().Msg("Writing buf.work.yaml")
	workspacesFile, err := os.OpenFile("buf.work.yaml", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return cleanup, err
	}

	if err := yaml.NewEncoder(workspacesFile).Encode(workspaces); err != nil {
		return cleanup, err
	}

	return cleanup, nil
}

func GetCommonProtoRepo() string {
	protoRepo := os.Getenv("PROTO_REPO")
	if protoRepo == "" {
		protoRepo = "../proto"
	}
	return protoRepo
}
