package deps

import (
	"fmt"
	"sync"

	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
)

// DefGoDep defines a go dependency that can be installed using
// a command like `go install github.com/aserto-dev/foo@v1.2.3`
func DefGoDep(name, importPath, version string) {
	cmdRegisterMutex.Lock()
	defer cmdRegisterMutex.Unlock()

	if _, ok := config.Go[name]; !ok {
		config.Go[name] = &depDetails{Once: &sync.Once{}}
	}

	config.Go[name].Procure = func() {
		config.Go[name].Once.Do(func() {
			installGoBin(importPath, version)
		})
	}
}

// GoDepOutput returns a command for running a go dependency.
// Its output is returned.
func GoDepOutput(name string) func(...string) (string, error) {
	def := config.Go[name]

	if def == nil {
		panic(errors.Errorf("didn't find a go binary dependency named '%s'", name))
	}

	return func(args ...string) (string, error) {
		def.Procure()
		return sh.Output(name, args...)
	}
}

// GoDepOutputWith returns a command for running a go dependency with env vars.
// Its output is returned.
func GoDepOutputWith(name string) func(map[string]string, ...string) (string, error) {
	def := config.Go[name]

	if def == nil {
		panic(errors.Errorf("didn't find a go binary dependency named '%s'", name))
	}

	return func(env map[string]string, args ...string) (string, error) {
		def.Procure()
		return sh.OutputWith(env, name, args...)
	}
}

// GoDep returns a command for running a go dependency.
// Its output is sent to stdout.
func GoDep(name string) func(...string) error {
	def := config.Go[name]

	if def == nil {
		panic(errors.Errorf("didn't find a go binary dependency named '%s'", name))
	}

	return func(args ...string) error {
		def.Procure()
		return sh.RunV(name, args...)
	}
}

func installGoBin(importPath, version string) {
	err := sh.RunV("go", "install", fmt.Sprintf("%s@%s", importPath, version))
	if err != nil {
		panic(errors.Wrap(err, "failed to install go dependency"))
	}
}
