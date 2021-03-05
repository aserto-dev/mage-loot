package deps

import (
	"io/ioutil"
	"os"
	"sync"

	"github.com/aserto-dev/clui"
	"github.com/aserto-dev/mage-loot/fsutil"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type depsConfig struct {
	Go  map[string]*depDetails
	Bin map[string]*depDetails
	Lib map[string]*depDetails
}

type depDetails struct {
	Procure func()
	Once    *sync.Once
}

const (
	configFile = "Depfile"
)

var (
	config = depsConfig{
		Go:  map[string]*depDetails{},
		Bin: map[string]*depDetails{},
		Lib: map[string]*depDetails{},
	}

	cmdRegisterMutex = &sync.Mutex{}
	ui               = clui.NewUI()
)

// GetAllDeps explicitly goes through all dependencies
// and downloads them, even if they might not be used.
func GetAllDeps() {
	for name, bin := range config.Bin {
		ui.Normal().Msgf("Procuring bin '%s'", name)
		bin.Procure()
	}
	for name, goBin := range config.Go {
		ui.Normal().Msgf("Procuring go bin '%s'", name)
		goBin.Procure()
	}

	ui.Exclamation().Msg("Cleaning lib dir.")
	err := os.RemoveAll(LibDir())
	if err != nil {
		panic(errors.Wrap(err, "failed to clean lib dir"))
	}
	for name, lib := range config.Lib {
		ui.Normal().Msgf("Procuring lib '%s'", name)
		lib.Procure()
	}
}

func init() {
	if exists, _ := fsutil.FileExists(configFile); !exists {
		return
	}

	configs := &struct {
		Go map[string]struct {
			ImportPath string `yaml:"importPath"`
			Version    string `yaml:"version"`
		} `yaml:"go"`
		Bin map[string]struct {
			Version  string   `yaml:"version"`
			URL      string   `yaml:"url"`
			SHA      string   `yaml:"sha"`
			ZipPaths []string `yaml:"zipPaths"`
			TGzPaths []string `yaml:"tgzPaths"`
		} `yaml:"bin"`
		Lib map[string]struct {
			Version   string   `yaml:"version"`
			URL       string   `yaml:"url"`
			SHA       string   `yaml:"sha"`
			ZipPaths  []string `yaml:"zipPaths"`
			TGzPaths  []string `yaml:"tgzPaths"`
			LibPrefix string   `yaml:"libPrefix"`
		} `yaml:"lib"`
	}{}

	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		panic(errors.Wrapf(err, "failed to read %s", configFile))
	}
	err = yaml.Unmarshal(yamlFile, configs)
	if err != nil {
		panic(errors.Wrapf(err, "failed to unmarshal %s", configFile))
	}

	for name, bin := range configs.Bin {
		options := []Option{}
		if len(bin.ZipPaths) != 0 {
			options = append(options, WithZipPaths(bin.ZipPaths...))
		}
		if len(bin.TGzPaths) != 0 {
			options = append(options, WithTGzPaths(bin.TGzPaths...))
		}

		DefBinDep(name, bin.URL, bin.Version, bin.SHA, options...)
	}

	for name, lib := range configs.Lib {
		options := []Option{}
		if len(lib.ZipPaths) != 0 {
			options = append(options, WithZipPaths(lib.ZipPaths...))
		}
		if len(lib.TGzPaths) != 0 {
			options = append(options, WithTGzPaths(lib.TGzPaths...))
		}
		if lib.LibPrefix != "" {
			options = append(options, WithLibPrefix(lib.LibPrefix))
		}

		DefLibDep(name, lib.URL, lib.Version, lib.SHA, options...)
	}

	for name, goBin := range configs.Go {
		DefGoDep(name, goBin.ImportPath, goBin.Version)
	}
}
