package sqlboiler

import (
	"path"
	"strings"

	"github.com/aserto-dev/clui"
	"github.com/aserto-dev/mage-loot/deps"
	"github.com/magefile/mage/sh"
)

type sqlboilerArgs struct {
	args []string
}

// Arg represents a sqlboiler CLI argument
type Arg func(*sqlboilerArgs)

var (
	ui = clui.NewUI()
)

// Run runs the sqlboiler CLI
func Run(args ...Arg) error {
	sqlboilerArgs := &sqlboilerArgs{}

	for _, arg := range args {
		arg(sqlboilerArgs)
	}

	finalArgs := []string{}

	finalArgs = append(finalArgs, sqlboilerArgs.args...)

	psqlPluginPath := path.Dir(deps.GoBinPath("sqlboiler-psql"))

	ui.Normal().
		WithStringValue("command", "sqlboiler "+strings.Join(sqlboilerArgs.args, "\n")).
		WithStringValue("psql plugin path", psqlPluginPath).
		WithStringValue("args", strings.Join(finalArgs, " ")).
		Msg(">>> executing sqlboiler")

	return sh.RunWithV(
		map[string]string{
			"PATH": psqlPluginPath,
		},
		deps.GoBinPath("sqlboiler"),
		finalArgs...)
}

// AddArg adds a simple argument.
// e.g. --foo
func AddArg(arg string) func(*sqlboilerArgs) {
	return func(o *sqlboilerArgs) {
		o.args = append(o.args, arg)
	}
}
