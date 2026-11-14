package main

import (
	"io"
	"io/fs"
)

// check runs velourbbs check (section 9.1): it writes each error and warning to stderr.
func check(args []string, fsys fs.FS, stderr io.Writer) int {
	if len(args) > 0 {
		_, _ = io.WriteString(stderr, "velourbbs check: no arguments expected\n"+usage)
		return exitUsage
	}
	_, _, ds := load(fsys)
	writeDiags(stderr, ds)
	if ds.HasErrors() {
		return exitError
	}
	return 0
}
