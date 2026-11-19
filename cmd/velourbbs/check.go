package main

import (
	"io"
	"io/fs"

	"github.com/schors/velourbbs/internal/bbs"
)

// check runs velourbbs check (screen-render, section 3.1): it writes each error and warning to stderr.
func check(args []string, fsys fs.FS, stderr io.Writer) int {
	if len(args) > 0 {
		_, _ = io.WriteString(stderr, "velourbbs check: no arguments expected\n"+usage)
		return exitUsage
	}
	_, ds := bbs.Load(fsys)
	writeDiags(stderr, ds)
	if ds.HasErrors() {
		return exitError
	}
	return 0
}
