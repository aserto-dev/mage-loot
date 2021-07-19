package dotnet

import (
	"fmt"
	"strings"

	"github.com/aserto-dev/clui"
	"github.com/aserto-dev/mage-loot/deps"
	"github.com/magefile/mage/sh"
)

type dotnetArgs struct {
	args    []string
	project string
}

// Arg represents a dotnet CLI argument
type Arg func(*dotnetArgs)

var (
	ui = clui.NewUI()
)

// Run runs the dotnet CLI
func Run(args ...Arg) error {
	dotnetArgs := &dotnetArgs{}

	for _, arg := range args {
		arg(dotnetArgs)
	}

	finalArgs := []string{}

	finalArgs = append(finalArgs, dotnetArgs.args...)
	finalArgs = append(finalArgs, dotnetArgs.project)

	ui.Normal().
		WithStringValue("command", "dotnet "+strings.Join(dotnetArgs.args, "\n")).
		Msg(">>> executing dotnet")

	return sh.RunV(deps.BinPath("dotnet"), finalArgs...)
}

// Add adds a new "name value" style argument.
// e.g. dotnet
func Add(name, value string) func(*dotnetArgs) {
	return AddArg(fmt.Sprintf("%s %s", name, value))
}

// AddArg adds a simple argument.
// e.g. --help
func AddArg(arg string) func(*dotnetArgs) {
	return func(o *dotnetArgs) {
		o.args = append(o.args, arg)
	}
}

// Help - Show this text and exit.
func Help() Arg { return AddArg("--help") }
