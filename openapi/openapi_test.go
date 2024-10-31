package openapi_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aserto-dev/mage-loot/openapi"
	magetesting "github.com/aserto-dev/mage-loot/openapi/testing"
)

const (
	genTypeGoGinServer = "go-gin-server"
	genPackageName     = "deadbeef"
	openapiTestVersion = "v4.3.1"
)

func TestDefinitionOutsideCurrentDir(t *testing.T) {
	g := NewGomegaWithT(t)

	definitionPath := filepath.Join(os.TempDir(), "foo.yaml")
	outputDir := os.TempDir()
	packageName := genPackageName
	generatorType := genTypeGoGinServer

	err := openapi.GenerateOpenAPI(openapiTestVersion, definitionPath, packageName, outputDir, generatorType)

	g.Expect(err).To(HaveOccurred())
	expected := fmt.Sprintf("file '%sfoo.yaml' doesn't exist", outputDir)
	g.Expect(err.Error()).To(Equal(expected))
}

func TestDefinitionDoesNotExist(t *testing.T) {
	g := NewGomegaWithT(t)

	definitionFile, err := os.CreateTemp("", "mageloot-test")
	g.Expect(err).ToNot(HaveOccurred())
	defer func() {
		err = os.Remove(definitionFile.Name())
		g.Expect(err).ToNot(HaveOccurred())
	}()
	outputDir := os.TempDir()
	packageName := genPackageName
	generatorType := genTypeGoGinServer

	err = openapi.GenerateOpenAPI(openapiTestVersion, definitionFile.Name(), packageName, outputDir, generatorType)

	g.Expect(err).To(HaveOccurred())
	expected := fmt.Sprintf("path '%smageloot-test.+?' is outside the current directory", outputDir)
	g.Expect(err.Error()).To(MatchRegexp(expected))
}

func TestDefinitionIsADir(t *testing.T) {
	g := NewGomegaWithT(t)

	definitionPath := magetesting.AssetOpenAPIOutputDir()
	outputDir := os.TempDir()
	packageName := genPackageName
	generatorType := genTypeGoGinServer

	err := openapi.GenerateOpenAPI(openapiTestVersion, definitionPath, packageName, outputDir, generatorType)

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(MatchRegexp("failed to determine if file '.+?' exists: not a file"))
}

func TestOutputOutsideCurrentDir(t *testing.T) {
	g := NewGomegaWithT(t)

	definitionPath := magetesting.AssetWorkingOpenAPIYaml()
	outputDir := os.TempDir()
	packageName := genPackageName
	generatorType := genTypeGoGinServer

	err := openapi.GenerateOpenAPI(openapiTestVersion, definitionPath, packageName, outputDir, generatorType)

	g.Expect(err).To(HaveOccurred())
	expected := fmt.Sprintf("path '%s' is outside the current directory", strings.TrimSuffix(outputDir, "/"))
	g.Expect(err.Error()).To(Equal(expected))
}

func TestOutputDoesNotExist(t *testing.T) {
	g := NewGomegaWithT(t)

	definitionPath := magetesting.AssetWorkingOpenAPIYaml()
	outputDir := filepath.Join(os.TempDir(), "thispathshouldnotexist")
	packageName := genPackageName
	generatorType := genTypeGoGinServer

	err := openapi.GenerateOpenAPI(openapiTestVersion, definitionPath, packageName, outputDir, generatorType)

	g.Expect(err).To(HaveOccurred())
	expected := fmt.Sprintf("dir '%s' doesn't exist", outputDir)
	g.Expect(err.Error()).To(Equal(expected))
}

func TestOutputIsAFile(t *testing.T) {
	g := NewGomegaWithT(t)

	definitionPath := magetesting.AssetWorkingOpenAPIYaml()
	outputDir := magetesting.AssetWorkingOpenAPIYaml()
	packageName := genPackageName
	generatorType := genTypeGoGinServer

	err := openapi.GenerateOpenAPI(openapiTestVersion, definitionPath, packageName, outputDir, generatorType)

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(MatchRegexp("failed to determine if dir '.+?' exists: not a dir"))
}

func TestWorkingOpenAPI(t *testing.T) {
	g := NewGomegaWithT(t)

	definitionPath := magetesting.AssetWorkingOpenAPIYaml()
	outputDir, err := os.MkdirTemp(magetesting.AssetOpenAPIOutputDir(), "outdir")
	g.Expect(err).ToNot(HaveOccurred())
	defer func() {
		err = os.RemoveAll(outputDir)
		g.Expect(err).ToNot(HaveOccurred())
	}()
	packageName := genPackageName
	generatorType := genTypeGoGinServer

	err = openapi.GenerateOpenAPI(openapiTestVersion, definitionPath, packageName, outputDir, generatorType)

	g.Expect(err).ToNot(HaveOccurred())
}

func TestBadOpenAPI(t *testing.T) {
	g := NewGomegaWithT(t)

	definitionPath := magetesting.AssetBadOpenAPIYaml()
	outputDir, err := os.MkdirTemp(magetesting.AssetOpenAPIOutputDir(), "outdir")
	g.Expect(err).ToNot(HaveOccurred())
	defer func() {
		err = os.RemoveAll(outputDir)
		g.Expect(err).ToNot(HaveOccurred())
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

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("failed to run docker container to generate an OpenAPI go server"))
}
