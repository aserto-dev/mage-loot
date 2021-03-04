package deps

import (
	"fmt"
	"sync"

	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
)

// DefGoDep defines a go dependency that can be installed using
// a command like `go install github.com/aserto-dev/foo@v1.2.3`
func DefGoDep(name, importPath, version string) func(...string) error {
	cmdRegisterMutex.Lock()
	defer cmdRegisterMutex.Unlock()

	if _, ok := cmdDownloadSync[name]; !ok {
		cmdDownloadSync[name] = &sync.Once{}
	}

	return func(args ...string) error {
		cmdDownloadSync[name].Do(func() {
			installGoBin(importPath, version)
		})

		return sh.RunV(name, args...)
	}
}

// GoDep returns a go dependency loaded from a Depfile
func GoDep(name string) func(...string) error {
	goBin := config.Go[name]

	if goBin == nil {
		panic(errors.Errorf("didn't find a go binary dependency named '%s'", name))
	}

	return goBin
}

func installGoBin(importPath, version string) {
	err := sh.RunV("go", "install", fmt.Sprintf("%s@%s", importPath, version))
	if err != nil {
		panic(errors.Wrap(err, "failed to install go dependency"))
	}
}
