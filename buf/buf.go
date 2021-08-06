package buf

import (
	"strings"

	"github.com/aserto-dev/clui"
	"github.com/aserto-dev/mage-loot/deps"
)

type bufArgs struct {
	args []string
}

// Arg represents a protoc CLI argument
type Arg func(*bufArgs)

var (
	ui = clui.NewUI()
)

// Run runs the protoc CLI
func Run(args ...Arg) error {
	bufArgs := &bufArgs{}

	for _, arg := range args {
		arg(bufArgs)
	}

	finalArgs := []string{}

	finalArgs = append(finalArgs, bufArgs.args...)

	ui.Normal().
		Msg(">>> executing buf " + strings.Join(bufArgs.args, " "))

	return deps.GoDep("buf")(finalArgs...)
}

// AddArg adds a simple argument.
func AddArg(arg string) func(*bufArgs) {
	return func(o *bufArgs) {
		o.args = append(o.args, arg)
	}
}

func AddPaths(paths []string) func(*bufArgs) {
	return func(o *bufArgs) {
		for _, p := range paths {
			o.args = append(o.args, "--path")
			o.args = append(o.args, p)
		}
	}
}
