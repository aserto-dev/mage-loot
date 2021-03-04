package deps

import (
	"io/ioutil"

	"github.com/aserto-dev/mage-loot/fsutil"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type depsConfig struct {
	Go  map[string]func(...string) error
	Bin map[string]func(...string) error
}

const (
	configFile = "Depfile"
)

var (
	config = initConfig()
)

func initConfig() depsConfig {
	result := depsConfig{
		Go:  map[string]func(...string) error{},
		Bin: map[string]func(...string) error{},
	}

	if exists, _ := fsutil.FileExists(configFile); !exists {
		return result
	}

	configs := &struct {
		Go map[string]struct {
			ImportPath string `yaml:"importPath"`
			Version    string `yaml:"version"`
		} `yaml:"go"`
		Bin map[string]struct {
			Version string `yaml:"version"`
			URL     string `yaml:"url"`
			SHA     string `yaml:"sha"`
			ZipPath string `yaml:"zipPath"`
			TGzPath string `yaml:"tgzPath"`
		} `yaml:"bin"`
		Lib map[string]struct {
			Version string `yaml:"version"`
			URL     string `yaml:"url"`
			SHA     string `yaml:"sha"`
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

	// process binaries
	for name, bin := range configs.Bin {
		options := []Option{}
		if bin.ZipPath != "" {
			options = append(options, WithZipPath(bin.ZipPath))
		}
		if bin.TGzPath != "" {
			options = append(options, WithTGzPath(bin.TGzPath))
		}

		binDep := DefBinDep(name, bin.URL, bin.Version, bin.SHA, options...)
		result.Bin[name] = binDep
	}

	for name, goBin := range configs.Go {
		goBinDep := DefGoDep(name, goBin.ImportPath, goBin.Version)
		result.Go[name] = goBinDep
	}

	return result
}
