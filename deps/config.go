package deps

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/aserto-dev/clui"
	"github.com/aserto-dev/mage-loot/fsutil"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

type depFile struct {
	Go  map[string]goConfig  `yaml:"go"`
	Bin map[string]binConfig `yaml:"bin"`
	Lib map[string]libConfig `yaml:"lib"`
}

type goConfig struct {
	ImportPath string `yaml:"importPath"`
	Version    string `yaml:"version"`
	Entrypoint string `yaml:"entrypoint"`
}

type binConfig struct {
	Version    string            `yaml:"version"`
	URL        string            `yaml:"url"`
	Entrypoint string            `yaml:"entrypoint"`
	SHA        map[string]string `yaml:"sha"`
	ZipPaths   []string          `yaml:"zipPaths"`
	TGzPaths   []string          `yaml:"tgzPaths"`
	TXzPaths   []string          `yaml:"txzPaths"`
}

type libConfig struct {
	Version   string   `yaml:"version"`
	URL       string   `yaml:"url"`
	OutputDir string   `yaml:"outputDir"`
	SHA       string   `yaml:"sha"`
	ZipPaths  []string `yaml:"zipPaths"`
	TGzPaths  []string `yaml:"tgzPaths"`
	TXzPaths  []string `yaml:"txzPaths"`
	LibPrefix string   `yaml:"libPrefix"`
}

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
		configFilePath, err := filepath.Abs(configFile)
		if err != nil {
			panic(errors.Wrap(err, "failed to get absolute path of config file"))
		}
		return configFilePath
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

	configs := &depFile{}

	yamlFile, err := os.ReadFile(configFile)
	if err != nil {
		panic(errors.Wrapf(err, "failed to read %s", configFile))
	}
	err = yaml.Unmarshal(yamlFile, configs)
	if err != nil {
		panic(errors.Wrapf(err, "failed to unmarshal %s", configFile))
	}

	buildBinDep(configs.Bin)

	buildLibDep(configs.Lib)

	buildGoDep(configs.Go)
}

func buildBinDep(binConfigs map[string]binConfig) {
	for name, bin := range binConfigs { //nolint:gocritic // TODO refactor
		options := []Option{}

		if len(bin.ZipPaths) != 0 {
			zipPaths := parseArrayTemplate(bin.ZipPaths, bin.Version)
			options = append(options, WithZipPaths(zipPaths...))
		}
		if len(bin.TGzPaths) != 0 {
			tgzPaths := parseArrayTemplate(bin.TGzPaths, bin.Version)
			options = append(options, WithTGzPaths(tgzPaths...))
		}
		if len(bin.TXzPaths) != 0 {
			txzPaths := parseArrayTemplate(bin.TXzPaths, bin.Version)
			options = append(options, WithTXzPaths(txzPaths...))
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
}

func buildLibDep(libConfigs map[string]libConfig) {
	for name, lib := range libConfigs { //nolint:gocritic // TODO refactor
		options := []Option{}
		if len(lib.ZipPaths) != 0 {
			zipPaths := parseArrayTemplate(lib.ZipPaths, lib.Version)
			options = append(options, WithZipPaths(zipPaths...))
		}
		if len(lib.TGzPaths) != 0 {
			tgzPaths := parseArrayTemplate(lib.TGzPaths, lib.Version)
			options = append(options, WithTGzPaths(tgzPaths...))
		}
		if len(lib.TXzPaths) != 0 {
			txzPaths := parseArrayTemplate(lib.TXzPaths, lib.Version)
			options = append(options, WithTGzPaths(txzPaths...))
		}

		if lib.LibPrefix != "" {
			libPrefix := parseStringTemplate(lib.LibPrefix, lib.Version)
			options = append(options, WithLibPrefix(libPrefix))
		}

		url := parseStringTemplate(lib.URL, lib.Version)

		DefLibDep(name, url, lib.SHA, lib.OutputDir, options...)
	}
}

func buildGoDep(goConfigs map[string]goConfig) {
	for name, goBin := range goConfigs {
		entrypoint := parseStringTemplate(goBin.Entrypoint, goBin.Version)
		if goBin.Entrypoint == "" {
			entrypoint = name
		}
		DefGoDep(name, goBin.ImportPath, goBin.Version, entrypoint)
	}
}
