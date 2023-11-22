package mage

import (
	"fmt"
	"os"
	"strings"

	"github.com/aserto-dev/clui"
	"github.com/aserto-dev/mage-loot/common"
	"github.com/magefile/mage/mage"
)

// Arg represents a mage argument.
type Arg func(*mageArgs)

var (
	ui = clui.NewUI()
)

type mageArgs struct {
	args []string
}

// AddArg adds a simple argument.
func AddArg(arg string) func(*mageArgs) {
	return func(o *mageArgs) {
		o.args = append(o.args, arg)
	}
}

// Run mage on the specified directory with the given args.
// the output of the mage command is printed to stdout and the error message to stderr.
func RunDir(dir string, args ...Arg) error {
	return mageRun(dir, dir, args...)
}

// Run mage on the specified directory with the given args.
// the output of the mage command is printed to stdout and the error message to stderr.
func RunDirs(dir, workDir string, args ...Arg) error {
	return mageRun(dir, workDir, args...)
}

// Run mage on the current directory with the given args
// the output of the mage command is printed to stdout and the error message to stderr.
func Run(args ...Arg) error {
	dir := common.WorkDir()
	return mageRun(dir, dir, args...)
}

func mageRun(mageDir, workDir string, args ...Arg) error {
	mageArgs := &mageArgs{}

	for _, arg := range args {
		arg(mageArgs)
	}

	invocation := mage.Invocation{
		Dir:     mageDir,
		Args:    mageArgs.args,
		WorkDir: workDir,
	}

	invocation.Stderr = os.Stderr
	invocation.Stdout = os.Stdout

	ui.Normal().
		Msgf(">>> executing 'mage %s' on %s", strings.Join(mageArgs.args, " "), invocation.Dir)

	exitCode := mage.Invoke(invocation)

	if exitCode != 0 {
		return fmt.Errorf("mage exited with code %d", exitCode)
	}
	return nil
}
