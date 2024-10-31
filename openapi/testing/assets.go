package testing

import (
	"path/filepath"
	"runtime"
)

// AssetsDir returns the directory containing test assets.
func AssetsDir() string {
	_, filename, _, _ := runtime.Caller(0) // nolint: dogsled

	return filepath.Join(filepath.Dir(filename), "assets")
}

// AssetWorkingOpenAPIYaml returns a path to a working OpenAPI yaml definition.
func AssetWorkingOpenAPIYaml() string {
	return filepath.Join(AssetsDir(), "working-openapi.yaml")
}

// AssetBadOpenAPIYaml returns a path to an OpenAPI yaml definition that's incorrect.
func AssetBadOpenAPIYaml() string {
	return filepath.Join(AssetsDir(), "bad-openapi.yaml")
}

// AssetOpenAPIOutputDir returns a path to a directory where we can generate Open API code.
func AssetOpenAPIOutputDir() string {
	return filepath.Join(AssetsDir(), "openapi_output")
}
