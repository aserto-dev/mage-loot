package openapi_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aserto-dev/mage-loot/openapi"
	magetesting "github.com/aserto-dev/mage-loot/openapi/testing"
)

const (
	genTypeGoGinServer = "go-gin-server"
	genPackageName     = "deadbeef"
	openapiTestVersion = "v4.3.1"
)

func TestDefinitionOutsideCurrentDir(t *testing.T) {
	assert := require.New(t)

	definitionPath := filepath.Join(os.TempDir(), "foo.yaml")
	outputDir := os.TempDir()
	packageName := genPackageName
	generatorType := genTypeGoGinServer

	err := openapi.GenerateOpenAPI(openapiTestVersion, definitionPath, packageName, outputDir, generatorType)
	expected := fmt.Sprintf("file '%s' doesn't exist", filepath.Join(outputDir, "foo.yaml"))
	assert.ErrorContains(err, expected)
}

func TestDefinitionDoesNotExist(t *testing.T) {
	assert := require.New(t)

	definitionFile, err := os.CreateTemp("", "mageloot-test")
	assert.NoError(err)
	defer func() {
		err = os.Remove(definitionFile.Name())
		assert.NoError(err)
	}()
	outputDir := os.TempDir()
	packageName := genPackageName
	generatorType := genTypeGoGinServer

	err = openapi.GenerateOpenAPI(openapiTestVersion, definitionFile.Name(), packageName, outputDir, generatorType)
	assert.Error(err)

	expected := fmt.Sprintf("path '%s' is outside the current directory", filepath.Join(outputDir, "mageloot-test.+?"))
	assert.Regexp(expected, err.Error())
}

func TestDefinitionIsADir(t *testing.T) {
	assert := require.New(t)

	definitionPath := magetesting.AssetOpenAPIOutputDir()
	outputDir := os.TempDir()
	packageName := genPackageName
	generatorType := genTypeGoGinServer

	err := openapi.GenerateOpenAPI(openapiTestVersion, definitionPath, packageName, outputDir, generatorType)
	assert.Error(err)

	assert.Regexp("failed to determine if file '.+?' exists: not a file", err.Error())
}

func TestOutputOutsideCurrentDir(t *testing.T) {
	assert := require.New(t)

	definitionPath := magetesting.AssetWorkingOpenAPIYaml()
	outputDir := os.TempDir()
	packageName := genPackageName
	generatorType := genTypeGoGinServer

	err := openapi.GenerateOpenAPI(openapiTestVersion, definitionPath, packageName, outputDir, generatorType)

	expected := fmt.Sprintf("path '%s' is outside the current directory", strings.TrimSuffix(outputDir, "/"))
	assert.ErrorContains(err, expected)
}

func TestOutputDoesNotExist(t *testing.T) {
	assert := require.New(t)

	definitionPath := magetesting.AssetWorkingOpenAPIYaml()
	outputDir := filepath.Join(os.TempDir(), "thispathshouldnotexist")
	packageName := genPackageName
	generatorType := genTypeGoGinServer

	err := openapi.GenerateOpenAPI(openapiTestVersion, definitionPath, packageName, outputDir, generatorType)

	expected := fmt.Sprintf("dir '%s' doesn't exist", outputDir)
	assert.ErrorContains(err, expected)
}

func TestOutputIsAFile(t *testing.T) {
	assert := require.New(t)

	definitionPath := magetesting.AssetWorkingOpenAPIYaml()
	outputDir := magetesting.AssetWorkingOpenAPIYaml()
	packageName := genPackageName
	generatorType := genTypeGoGinServer

	err := openapi.GenerateOpenAPI(openapiTestVersion, definitionPath, packageName, outputDir, generatorType)
	assert.Error(err)

	assert.Regexp("failed to determine if dir '.+?' exists: not a dir", err.Error())
}

func TestWorkingOpenAPI(t *testing.T) {
	assert := require.New(t)

	definitionPath := magetesting.AssetWorkingOpenAPIYaml()
	outputDir, err := os.MkdirTemp(magetesting.AssetOpenAPIOutputDir(), "outdir")
	assert.NoError(err)
	defer func() {
		err = os.RemoveAll(outputDir)
		assert.NoError(err)
	}()
	packageName := genPackageName
	generatorType := genTypeGoGinServer

	err = openapi.GenerateOpenAPI(openapiTestVersion, definitionPath, packageName, outputDir, generatorType)
	assert.NoError(err)
}

func TestBadOpenAPI(t *testing.T) {
	assert := require.New(t)

	definitionPath := magetesting.AssetBadOpenAPIYaml()
	outputDir, err := os.MkdirTemp(magetesting.AssetOpenAPIOutputDir(), "outdir")
	assert.NoError(err)
	defer func() {
		err = os.RemoveAll(outputDir)
		assert.NoError(err)
	}()
	packageName := genPackageName
	generatorType := genTypeGoGinServer

	// Don't print to stderr for this test, to avoid confusion
	// (we expect things to fail)
	currentErrOutput := os.Stderr
	os.Stderr = os.NewFile(uintptr(syscall.Stdin), os.DevNull)
	defer func() {
		os.Stderr = currentErrOutput
	}()
	err = openapi.GenerateOpenAPI(openapiTestVersion, definitionPath, packageName, outputDir, generatorType)
	assert.ErrorContains(err, "failed to run docker container to generate an OpenAPI go server")
}
