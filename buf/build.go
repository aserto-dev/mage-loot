package buf

import (
	"github.com/aserto-dev/mage-loot/fsutil"
)

// Generate proto artifacts.
func Generate(binFile string, protoPluginPaths []string) error {
	if err := Lint(); err != nil {
		return err
	}
	if err := Build(binFile); err != nil {
		return err
	}

	path, err := getBufPath(protoPluginPaths)
	if err != nil {
		return err
	}

	return RunWithEnv(
		map[string]string{
			"PATH": path,
		},
		WithLogin(),
		AddArg("generate"),
	)
}

func Lint(bufConfigs ...string) error {
	if len(bufConfigs) == 0 {
		return Run(
			AddArg("lint"),
		)
	}

	for _, c := range bufConfigs {
		if err := Run(
			AddArg("lint"),
			AddArg("--config"),
			AddArg(c),
		); err != nil {
			return err
		}
	}

	return nil
}

func LintBin(binFile string, bufConfigs ...string) error {
	if len(bufConfigs) == 0 {
		return Run(
			AddArg("lint"),
			AddArg(binFile),
		)
	}

	for _, c := range bufConfigs {
		if err := Run(
			AddArg("lint"),
			AddArg("--config"),
			AddArg(c),
			AddArg(binFile),
		); err != nil {
			return err
		}
	}

	return nil
}

func BuildDir(dir, binFile string) error {
	fsutil.EnsureDir("bin")

	return Run(
		AddArg("build"),
		AddArg("--output"),
		AddArg(binFile),
		AddArg(dir),
	)
}

func Build(binFile string, bufConfigs ...string) error {
	fsutil.EnsureDir("bin")

	if len(bufConfigs) == 0 {
		return Run(
			AddArg("build"),
			AddArg("--output"),
			AddArg(binFile),
		)
	}

	for _, c := range bufConfigs {
		if err := Run(
			AddArg("build"),
			AddArg("--output"),
			AddArg(binFile),
			AddArg("--config"),
			AddArg(c),
		); err != nil {
			return err
		}
	}

	return nil
}

func ModUpdate(dirs ...string) error {
	if len(dirs) == 0 {
		return Run(
			AddArg("mod"),
			AddArg("update"),
		)
	}

	for _, d := range dirs {
		err := Run(
			AddArg("mod"),
			AddArg("update"),
			AddArg(d),
		)
		if err != nil {
			return err
		}
	}

	return nil
}
