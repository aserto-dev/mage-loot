package openapi_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
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
	g.Expect(err.Error()).To(Equal("file '/tmp/foo.yaml' doesn't exist"))
}

func TestDefinitionDoesNotExist(t *testing.T) {
	g := NewGomegaWithT(t)

	definitionFile, err := ioutil.TempFile("", "mageloot-test")
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
	g.Expect(err.Error()).To(MatchRegexp("path '/tmp/mageloot-test.+?' is outside the current directory"))
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
	g.Expect(err.Error()).To(Equal("path '/tmp' is outside the current directory"))
}

func TestOutputDoesNotExist(t *testing.T) {
	g := NewGomegaWithT(t)

	definitionPath := magetesting.AssetWorkingOpenAPIYaml()
	outputDir := filepath.Join(os.TempDir(), "thispathshouldnotexist")
	packageName := genPackageName
	generatorType := genTypeGoGinServer

	err := openapi.GenerateOpenAPI(openapiTestVersion, definitionPath, packageName, outputDir, generatorType)

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(Equal("dir '/tmp/thispathshouldnotexist' doesn't exist"))
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
	outputDir, err := ioutil.TempDir(magetesting.AssetOpenAPIOutputDir(), "outdir")
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
	outputDir, err := ioutil.TempDir(magetesting.AssetOpenAPIOutputDir(), "outdir")
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
