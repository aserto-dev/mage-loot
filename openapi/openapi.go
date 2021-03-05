package openapi

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/aserto-dev/mage-loot/fsutil"
	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
)

const (
	openAPIDockerImage = "openapitools/openapi-generator-cli"
)

// GenerateOpenAPI generates code for the specified Open API definition
// the openAPI definition path must be relative to the current working directory
func GenerateOpenAPI(version, openAPIDefinitionPath, packageName, outputDir, generatorType string, additionalArgs ...string) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "failed to get working directory")
	}

	// Check that the definition path is located somewhere inside the current dir
	openapiAbsPath, err := filepath.Abs(openAPIDefinitionPath)
	if err != nil {
		return errors.Wrapf(err, "failed to determine absolute path for openapi definition '%s'", openAPIDefinitionPath)
	}
	exists, err := fsutil.FileExists(openapiAbsPath)
	if err != nil {
		return errors.Wrapf(err, "failed to determine if file '%s' exists", openapiAbsPath)
	}
	if !exists {
		return errors.Errorf("file '%s' doesn't exist", openapiAbsPath)
	}
	openapiRelPath, err := filepath.Rel(currentDir, openapiAbsPath)
	if err != nil {
		return errors.Wrapf(err, "failed to determine relative path to '%s' from current working dir", openapiAbsPath)
	}
	if strings.HasPrefix(openapiRelPath, "..") {
		return errors.Errorf("path '%s' is outside the current directory", openapiAbsPath)
	}
	openapiContainerPath := filepath.Join("/local", openapiRelPath) // nolint // container will always be unix

	// Check that the output director is located somewhere inside the current dir
	outputAbsDir, err := filepath.Abs(outputDir)
	if err != nil {
		return errors.Wrapf(err, "failed to determine absolute path for output dir '%s'", outputDir)
	}
	exists, err = fsutil.DirExists(outputAbsDir)
	if err != nil {
		return errors.Wrapf(err, "failed to determine if dir '%s' exists", outputAbsDir)
	}
	if !exists {
		return errors.Errorf("dir '%s' doesn't exist", outputAbsDir)
	}
	outputRelDir, err := filepath.Rel(currentDir, outputAbsDir)
	if err != nil {
		return errors.Wrapf(err, "failed to determine relative path to '%s' from current working dir", outputAbsDir)
	}
	if strings.HasPrefix(outputRelDir, "..") {
		return errors.Errorf("path '%s' is outside the current directory", outputAbsDir)
	}
	outputContainerPath := filepath.Join("/local", outputRelDir) // nolint // container will always be unix

	currentUser, err := user.Current()
	if err != nil {
		return errors.Wrap(err, "failed to determine current user")
	}

	err = sh.Run("docker", append([]string{"run", "--rm",
		"-u", fmt.Sprintf("%s:%s", currentUser.Uid, currentUser.Gid),
		"-v", fmt.Sprintf("%s:/local", currentDir),
		fmt.Sprintf("%s:%s", openAPIDockerImage, version),
		"generate", "-i", openapiContainerPath, "-g", generatorType, "-o", outputContainerPath},
		additionalArgs...)...)

	if err != nil {
		return errors.Wrap(err, "failed to run docker container to generate an OpenAPI go server")
	}
	return nil
}
