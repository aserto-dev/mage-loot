package dotnetrestore

import (
	"strings"

	"github.com/aserto-dev/clui"
	"github.com/aserto-dev/mage-loot/deps"
	"github.com/magefile/mage/sh"
)

type dotnetRestoreArgs struct {
	args     []string
	solution string
	project  string
}

// Arg represents a dotnet build CLI argument.
type Arg func(*dotnetRestoreArgs)

var (
	ui = clui.NewUI()
)

// Run runs the dotnet CLI.
func Run(args ...Arg) error {
	dotnetRestoreArgs := &dotnetRestoreArgs{}

	for _, arg := range args {
		arg(dotnetRestoreArgs)
	}

	finalArgs := []string{}

	finalArgs = append(finalArgs, "restore")
	finalArgs = append(finalArgs, dotnetRestoreArgs.args...)
	finalArgs = append(finalArgs, dotnetRestoreArgs.project, dotnetRestoreArgs.solution)

	ui.Normal().
		WithStringValue("command", "dotnet "+strings.Join(finalArgs, "\n")).
		Msg(">>> executing dotnet")

	return sh.RunV(deps.BinPath("dotnet"), finalArgs...)
}

// Add adds a new "name value" style argument (e.g. --v m).
func Add(name, value string) func(*dotnetRestoreArgs) {
	return func(o *dotnetRestoreArgs) {
		o.args = append(o.args, name, value)
	}
}

// AddArg adds a simple argument (e.g. --help).
func AddArg(arg string) func(*dotnetRestoreArgs) {
	return func(o *dotnetRestoreArgs) {
		o.args = append(o.args, arg)
	}
}

// Help - Show this text and exit.
func Help() Arg { return AddArg("--help") }

// The NuGet package source to use for the restore.
func Source(source string) Arg { return Add("-s", source) }

// The target runtime to restore packages for.
func Runtime(runtime string) Arg { return Add("-r", runtime) }

// The directory to restore packages to.
func Packages(packagesDir string) Arg { return Add("--packages", packagesDir) }

// Prevent restoring multiple projects in parallel.
func DisableParallel() Arg { return AddArg("--disable-parallel") }

// The NuGet configuration file to use.
func ConfigFile(configFile string) Arg { return Add("--configfile", configFile) }

// Do not cache packages and http requests.
func NoCache() Arg { return AddArg("--no-cache") }

// Treat package source failures as warnings.
func IgnoreFailedSources() Arg { return AddArg("Treat package source failures as warnings.") }

// Do not restore project-to-project references and only restore the specified project.
func NoDependencies() Arg { return AddArg(" -no-dependencies") }

// Force all dependencies to be resolved even if the last restore was successful.
// This is equivalent to deleting project.assets.json.
func Force() Arg { return AddArg("-f") }

// Set the MSBuild verbosity level. Allowed values are q[uiet], m[inimal], n[ormal], d[etailed], and diag[nostic].
func Verbosity(level string) Arg { return Add("-v", level) }

// Allows the command to stop and wait for user input or action (for example to complete authentication).
func Interactive() Arg { return AddArg("--interactive") }

// Don't allow updating project lock file.
func LockedMode() Arg { return AddArg("--locked-mode") }

// Output location where project lock file is written. By default, this is 'PROJECT_ROOT\packages.lock.json'.
func LockFilePath(lockFilePath string) Arg { return Add("--lock-file-path", lockFilePath) }

// Forces restore to reevaluate all dependencies even if a lock file already exists.
func ForceElevate() Arg { return AddArg("--force-evaluate") }

// Solution file (.sln) to build.
func Solution(file string) func(*dotnetRestoreArgs) {
	return func(o *dotnetRestoreArgs) {
		o.solution = file
	}
}

// Project file (.csproj) to build.
func Project(file string) func(*dotnetRestoreArgs) {
	return func(o *dotnetRestoreArgs) {
		o.project = file
	}
}
