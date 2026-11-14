// Command bbs checks and previews the screens of a BBS directory, the current directory,
// and serves them over telnet. It also converts ANSI art to the flat form.
// Specification: docs/specs/screen-format.md, section 9, docs/specs/art-import.md, section 2,
// and docs/specs/telnet-term.md, section 10.
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

// The exit statuses: an error in the files (section 9.1) or of listen, and a usage error.
const (
	exitError = 1
	exitUsage = 2
)

const usage = `usage:
  bbs art import <input> <output>
  bbs check
  bbs preview <name> [--size <cols>x<rows>] [--set <path>=<value>]... [--keys <keys>]
  bbs serve
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
	case "art":
		return art(args[1:], stderr)
	case "check":
		return check(args[1:], fsys, stderr)
	case "preview":
		return preview(args[1:], fsys, stdout, stderr)
	case "serve":
		return serve(args[1:], fsys, stderr)
	}
	_, _ = fmt.Fprintf(stderr, "bbs: unknown command %q\n%s", args[0], usage)
	return exitUsage
}

// load reads bbs.toml and the screens of fsys and does the menu checks.
// A missing bbs.toml is an error.
// The screens are nil when the diagnostics have an error.
func load(fsys fs.FS) (config.Config, map[string]*screen.Screen, screen.Diags) {
	var ds screen.Diags
	data, err := screen.ReadFile(fsys, config.File)
	if err != nil {
		ds.Errorf(config.File, 1, 1, "%v", err)
	}
	// Parse also gives the defaults when the read fails, so the screens load with them.
	cfg, cds := config.Parse(data)
	ds = append(ds, cds...)
	screens, sds := bbt.Load(fsys, cfg.Theme, cfg.Width, cfg.Fields)
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
