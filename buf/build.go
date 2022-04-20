package buf

import (
	"io"
	"os/exec"
	"strings"

	"github.com/aserto-dev/mage-loot/deps"
	"github.com/aserto-dev/mage-loot/fsutil"
	"github.com/aserto-dev/mage-loot/testutil"
)

// Generate proto artifacts
func Generate(binFile string) error {
	if err := Login(); err != nil {
		return err
	}
	if err := Lint(); err != nil {
		return err
	}
	if err := Build(binFile); err != nil {
		return err
	}

	return Run(
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

func Login() error {
	bufToken := testutil.VaultValue("buf.build", "ASERTO_BUF_TOKEN")
	bufUser := testutil.VaultValue("buf.build", "ASERTO_BUF_USER")

	args := []string{"registry", "login", "--username", bufUser, "--token-stdin"}

	cmd := exec.Command(deps.GoBinPath("buf"), args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, bufToken)
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	ui.Normal().
		Msg(">>> executing buf " + strings.Join(args, " "))
	ui.Normal().
		Msg(string(out))

	return err
}
