package dotnetbuild

import (
	"strings"

	"github.com/aserto-dev/clui"
	"github.com/aserto-dev/mage-loot/deps"
	"github.com/magefile/mage/sh"
)

type dotnetBuildArgs struct {
	args     []string
	solution string
	project  string
}

// Arg represents a dotnet build CLI argument
type Arg func(*dotnetBuildArgs)

var (
	ui = clui.NewUI()
)

// Run runs the dotnet CLI
func Run(args ...Arg) error {
	dotnetBuildArgs := &dotnetBuildArgs{}

	for _, arg := range args {
		arg(dotnetBuildArgs)
	}

	finalArgs := []string{}

	finalArgs = append(finalArgs, "build")
	finalArgs = append(finalArgs, dotnetBuildArgs.args...)
	finalArgs = append(finalArgs, dotnetBuildArgs.project, dotnetBuildArgs.solution)

	ui.Normal().
		WithStringValue("command", "dotnet "+strings.Join(finalArgs, "\n")).
		Msg(">>> executing dotnet")

	return sh.RunV(deps.BinPath("dotnet"), finalArgs...)
}

// Add adds a new "name value" style argument.
// e.g. --v m
func Add(name, value string) func(*dotnetBuildArgs) {
	return func(o *dotnetBuildArgs) {
		o.args = append(o.args, name, value)
	}
}

// AddArg adds a simple argument.
// e.g. --help
func AddArg(arg string) func(*dotnetBuildArgs) {
	return func(o *dotnetBuildArgs) {
		o.args = append(o.args, arg)
	}
}

// Help - Show this text and exit.
func Help() Arg { return AddArg("--help") }

// The output directory to place built artifacts in.
func Output(output string) Arg { return Add("-o", output) }

// The target framework to build for. The target framework must also be specified in the project file.
func Framework(framework string) Arg { return Add("-f", framework) }

// The configuration to use for building the project. The default for most projects is 'Debug'.
func Configuration(configuration string) Arg { return Add("-c", configuration) }

// The target runtime to build for.
func Runtime(runtime string) Arg { return Add("-r", runtime) }

// Set the value of the $(VersionSuffix) property to use when building the project.
func VersionSuffix(versionSufix string) Arg { return Add("-version-suffix", versionSufix) }

// Do not use incremental building.
func NoIncremental() Arg { return AddArg("--no-incremental") }

// Do not build project-to-project references and only build the specified project.
func NoDependencies() Arg { return AddArg("--no-dependencies") }

// Do not display the startup banner or the copyright message.
func NoLogo() Arg { return AddArg("--nologo") }

// Do not restore the project before building.
func NoRestore() Arg { return AddArg("--no-restore") }

// Allows the command to stop and wait for user input or action (for example to complete authentication)
func Interactive() Arg { return AddArg("--interactive") }

// Set the MSBuild verbosity level. Allowed values are q[uiet], m[inimal], n[ormal], d[etailed], and diag[nostic].
func Verbosity(level string) Arg { return Add("-v", level) }

// Force all dependencies to be resolved even if the last restore was successful.
func Force() Arg { return AddArg("--force") }

// Solution file (.sln) to build
func Solution(file string) func(*dotnetBuildArgs) {
	return func(o *dotnetBuildArgs) {
		o.solution = file
	}
}

// Project file (.csproj) to build
func Project(file string) func(*dotnetBuildArgs) {
	return func(o *dotnetBuildArgs) {
		o.project = file
	}
}
