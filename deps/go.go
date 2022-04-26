package deps

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
)

// DefGoDep defines a go dependency that can be installed using
// a command like `go install github.com/aserto-dev/foo@v1.2.3`
func DefGoDep(name, importPath, version, entrypoint string) {
	cmdRegisterMutex.Lock()
	defer cmdRegisterMutex.Unlock()

	if _, ok := config.Go[name]; !ok {
		config.Go[name] = &depDetails{Once: &sync.Once{}}
	}

	binPath := goBinFilePath(name, version)

	config.Go[name].Procure = func() {
		config.Go[name].Once.Do(func() {
			installGoBin(binPath, importPath, version)
		})
	}

	config.Go[name].Path = filepath.Join(binPath, entrypoint)
}

// GoDepOutput returns a command for running a go dependency.
// Its output is returned.
func GoDepOutput(name string) func(...string) (string, error) {
	def := config.Go[name]

	if def == nil {
		panic(errors.Errorf("didn't find a go binary dependency named '%s'", name))
	}

	return func(args ...string) (string, error) {
		if !skipProcurement {
			def.Procure()
		}

		return sh.Output(def.Path, args...)
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
		if !skipProcurement {
			def.Procure()
		}

		return sh.OutputWith(env, def.Path, args...)
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
		if !skipProcurement {
			def.Procure()
		}

		return sh.RunV(def.Path, args...)
	}
}

// GoDepWithEnv returns a command for running a go dependency.
// It accepts an env map for the new process. Its output is sent to stdout.
func GoDepWithEnv(env map[string]string, name string) func(...string) error {
	def := config.Go[name]

	if def == nil {
		panic(errors.Errorf("didn't find a binary dependency named '%s'", name))
	}

	return func(args ...string) error {
		if !skipProcurement {
			def.Procure()
		}

		return sh.RunWithV(env, def.Path, args...)
	}
}

func GoBinPath(name string) string {
	def := config.Go[name]

	if def == nil {
		panic(errors.Errorf("didn't find a go binary dependency named '%s'", name))
	}

	if !skipProcurement {
		def.Procure()
	}

	return def.Path
}

func installGoBin(binPath, importPath, version string) {
	env := make(map[string]string)
	env["GOBIN"] = binPath
	err := sh.RunWith(env, "go", "install", fmt.Sprintf("%s@%s", importPath, version))
	if err != nil {
		panic(errors.Wrap(err, "failed to install go dependency"))
	}
}

func goBinFilePath(name, version string) string {
	return filepath.Join(GoBinDir(), name+"-"+version)
}
