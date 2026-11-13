// Command bbs checks and previews the screens of a BBS directory: the current directory.
// Specification: docs/specs/screen-format.md, section 9.
package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"

	"newbbs/internal/bbt"
	"newbbs/internal/config"
	"newbbs/internal/menu"
	"newbbs/internal/screen"
)

// The exit statuses: an error in the files (section 9.1) and a usage error.
const (
	exitError = 1
	exitUsage = 2
)

const usage = `usage:
  bbs check
  bbs preview <name> [--size <cols>x<rows>] [--set <path>=<value>]... [--keys <keys>]
`

func main() {
	os.Exit(run(os.Args[1:], os.DirFS("."), os.Stdout, os.Stderr))
}

// run runs the command of args on the BBS directory fsys and gives the exit status.
func run(args []string, fsys fs.FS, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		_, _ = io.WriteString(stderr, usage)
		return exitUsage
	}
	switch args[0] {
	case "check":
		return check(args[1:], fsys, stderr)
	case "preview":
		return preview(args[1:], fsys, stdout, stderr)
	}
	_, _ = fmt.Fprintf(stderr, "bbs: unknown command %q\n%s", args[0], usage)
	return exitUsage
}

// load reads bbs.toml and the screens of fsys and does the menu checks.
// A missing bbs.toml is an error.
// The screens are nil when the diagnostics have an error.
func load(fsys fs.FS) (config.Config, map[string]*screen.Screen, screen.Diags) {
	const name = "bbs.toml"
	var ds screen.Diags
	data, err := screen.ReadFile(fsys, name)
	if err != nil {
		ds.Errorf(name, 1, 1, "%v", err)
	}
	// Parse also gives the defaults when the read fails, so the screens load with them.
	cfg, cds := config.Parse(name, data)
	ds = append(ds, cds...)
	screens, sds := bbt.Load(fsys, cfg.Theme, cfg.Width)
	ds = append(ds, sds...)
	ds.Sort()
	if ds.HasErrors() {
		return cfg, nil, ds
	}
	// The menu checks between files need all screens (menu-nav, section 7).
	ds = menu.Check(cfg, screens, ds)
	if ds.HasErrors() {
		screens = nil
	}
	return cfg, screens, ds
}

// writeDiags writes each diagnostic on its own line.
func writeDiags(w io.Writer, ds screen.Diags) {
	for _, d := range ds {
		_, _ = fmt.Fprintln(w, d)
	}
}
