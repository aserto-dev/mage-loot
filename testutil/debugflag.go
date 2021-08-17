package testutil

import "flag"

var (
	flagDebug = flag.Bool("debug", false, "Output server logs to stdout")
)

func DebugFlagSet() bool {
	return *flagDebug
}
