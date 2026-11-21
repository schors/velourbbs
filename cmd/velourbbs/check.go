package main

import (
	"flag"
	"io"
	"io/fs"
	"slices"

	"github.com/schors/velourbbs/internal/bbs"
	"github.com/schors/velourbbs/internal/config"
	"github.com/schors/velourbbs/internal/menu"
	"github.com/schors/velourbbs/internal/screen"
	"github.com/schors/velourbbs/internal/store"
)

// check runs velourbbs check (screen-render, section 3.1): it writes each error and warning to stderr.
// It reads the data file when it exists and bbs.toml has no error (data-store, section 7).
// With --size, it also gives the layout warnings after a check without errors (screen-render, section 3.6).
func check(args []string, dir string, fsys fs.FS, stderr io.Writer) int {
	var terms []screen.Term // the sizes, each once
	fl := flag.NewFlagSet("velourbbs check", flag.ContinueOnError)
	fl.SetOutput(stderr)
	fl.Usage = func() { _, _ = io.WriteString(stderr, usage) }
	fl.Func("size", "", func(v string) error {
		t := previewTerm()
		if err := parseSize(v, &t); err != nil {
			return err
		}
		if err := t.Check(); err != nil {
			return err
		}
		if !slices.Contains(terms, t) {
			terms = append(terms, t)
		}
		return nil
	})
	if err := fl.Parse(args); err != nil {
		return exitUsage
	}
	if fl.NArg() > 0 {
		_, _ = io.WriteString(stderr, "velourbbs check: no arguments expected\n"+usage)
		return exitUsage
	}
	m, ds := bbs.Load(fsys)
	if !slices.ContainsFunc(ds, func(d screen.Diag) bool { return d.File == config.File && !d.Warning }) {
		// The model is nil after an error in another file, so the path comes from a second load.
		cfg, _ := config.Load(fsys)
		ds = append(ds, store.Check(storePath(dir, cfg.Store.Path), cfg.Store.Path)...)
		ds.Sort()
	}
	if len(terms) > 0 && !ds.HasErrors() {
		sv := previewSession()
		sv.Name = m.Config.Name
		previewClock(&sv, m.Config, false)
		ds = append(ds, menu.Layout(m, terms, sv)...)
		ds.Sort()
	}
	writeDiags(stderr, ds)
	if ds.HasErrors() {
		return exitError
	}
	return 0
}
