package main

import (
	"io"
	"io/fs"
	"slices"

	"github.com/schors/velourbbs/internal/bbs"
	"github.com/schors/velourbbs/internal/config"
	"github.com/schors/velourbbs/internal/screen"
	"github.com/schors/velourbbs/internal/store"
)

// check runs velourbbs check (screen-render, section 3.1): it writes each error and warning to stderr.
// It reads the data file when it exists and bbs.toml has no error (data-store, section 7).
func check(args []string, dir string, fsys fs.FS, stderr io.Writer) int {
	if len(args) > 0 {
		_, _ = io.WriteString(stderr, "velourbbs check: no arguments expected\n"+usage)
		return exitUsage
	}
	_, ds := bbs.Load(fsys)
	if !slices.ContainsFunc(ds, func(d screen.Diag) bool { return d.File == config.File && !d.Warning }) {
		// The model is nil after an error in another file, so the path comes from a second load.
		cfg, _ := config.Load(fsys)
		ds = append(ds, store.Check(storePath(dir, cfg.Store.Path), cfg.Store.Path)...)
		ds.Sort()
	}
	writeDiags(stderr, ds)
	if ds.HasErrors() {
		return exitError
	}
	return 0
}
