package deps

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
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
	Path    string
}

var (
	config = depsConfig{
		Go:  map[string]*depDetails{},
		Bin: map[string]*depDetails{},
		Lib: map[string]*depDetails{},
	}

	skipProcurement  = false
	cmdRegisterMutex = &sync.Mutex{}
	ui               = clui.NewUI()
)

// GetAllDeps explicitly goes through all dependencies
// and downloads them, even if they might not be used.
func GetAllDeps() {
	for name, bin := range config.Bin {
		ui.Normal().Msgf("Procuring bin '%s'", name)

		if !skipProcurement {
			bin.Procure()
		}
	}
	for name, goBin := range config.Go {
		ui.Normal().Msgf("Procuring go bin '%s'", name)

		if !skipProcurement {
			goBin.Procure()
		}
	}

	ui.Exclamation().Msg("Cleaning lib dir.")
	err := os.RemoveAll(LibDir())
	if err != nil {
		panic(errors.Wrap(err, "failed to clean lib dir"))
	}
	for name, lib := range config.Lib {
		ui.Normal().Msgf("Procuring lib '%s'", name)

		if !skipProcurement {
			lib.Procure()
		}
	}
}

func lookupConfig(dir string) string {
	configFile := filepath.Join(dir, "Depfile")
	if exists, _ := fsutil.FileExists(configFile); exists {
		return configFile
	}

	parent, err := filepath.Abs(filepath.Join(dir, ".."))

	if parent == dir {
		return ""
	}

	if err != nil {
		panic(errors.Wrap(err, "failed to get parent path"))
	}

	return lookupConfig(parent)
}

func init() {
	_, skipProcurement = os.LookupEnv("DEPFILE_SKIP_PROCUREMENT")

	configFile := lookupConfig(".")
	if configFile == "" {
		return
	}

	currentDir = filepath.Dir(configFile)

	configs := &struct {
		Go map[string]struct {
			ImportPath string `yaml:"importPath"`
			Version    string `yaml:"version"`
		} `yaml:"go"`
		Bin map[string]struct {
			Version    string            `yaml:"version"`
			URL        string            `yaml:"url"`
			Entrypoint string            `yaml:"entrypoint"`
			SHA        map[string]string `yaml:"sha"`
			ZipPaths   []string          `yaml:"zipPaths"`
			TGzPaths   []string          `yaml:"tgzPaths"`
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
			zipPaths := parseArrayTemplate(bin.ZipPaths, bin.Version)
			options = append(options, WithZipPaths(zipPaths...))
		}
		if len(bin.TGzPaths) != 0 {
			tgzPaths := parseArrayTemplate(bin.TGzPaths, bin.Version)
			options = append(options, WithTGzPaths(tgzPaths...))
		}

		sha, ok := bin.SHA[runtime.GOOS+"-"+runtime.GOARCH]
		if !ok {
			panic(errors.Errorf("no SHA found for os and arch '%s'", runtime.GOOS+"-"+runtime.GOARCH))
		}
		entrypoint := parseStringTemplate(bin.Entrypoint, bin.Version)
		if bin.Entrypoint == "" {
			entrypoint = name
		}
		url := parseStringTemplate(bin.URL, bin.Version)
		DefBinDep(name, url, bin.Version, sha, entrypoint, options...)
	}

	for name, lib := range configs.Lib {
		options := []Option{}
		if len(lib.ZipPaths) != 0 {
			zipPaths := parseArrayTemplate(lib.ZipPaths, lib.Version)
			options = append(options, WithZipPaths(zipPaths...))
		}
		if len(lib.TGzPaths) != 0 {
			tgzPaths := parseArrayTemplate(lib.TGzPaths, lib.Version)
			options = append(options, WithTGzPaths(tgzPaths...))
		}
		if lib.LibPrefix != "" {
			libPrefix := parseStringTemplate(lib.LibPrefix, lib.Version)
			options = append(options, WithLibPrefix(libPrefix))
		}

		url := parseStringTemplate(lib.URL, lib.Version)

		DefLibDep(name, url, lib.SHA, options...)
	}

	for name, goBin := range configs.Go {
		DefGoDep(name, goBin.ImportPath, goBin.Version)
	}
}
