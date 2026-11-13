package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"strconv"
	"strings"

	"newbbs/internal/screen"
)

// preview runs bbs preview (section 9.2): it renders a screen and writes its cell dump
// to stdout. The flags can come before and after the name.
func preview(args []string, fsys fs.FS, stdout, stderr io.Writer) int {
	term := screen.Term{Type: "ansi", Cols: 80, Rows: 25, Charset: "utf-8", Colors: 16}
	fl := flag.NewFlagSet("bbs preview", flag.ContinueOnError)
	fl.SetOutput(stderr)
	fl.Usage = func() { _, _ = io.WriteString(stderr, usage) }
	fl.Func("size", "", func(v string) error { return parseSize(v, &term) })
	fl.Func("set", "", func(v string) error { return setField(v, &term) })

	// flag stops at the first argument that is not a flag, so parse again after each name.
	var names []string
	for {
		if err := fl.Parse(args); err != nil {
			return exitUsage
		}
		if fl.NArg() == 0 {
			break
		}
		names = append(names, fl.Arg(0))
		args = fl.Args()[1:]
	}
	if len(names) != 1 || names[0] == "" {
		_, _ = io.WriteString(stderr, "bbs preview: expected one screen name\n"+usage)
		return exitUsage
	}
	name := names[0]
	if err := term.Check(); err != nil {
		_, _ = fmt.Fprintf(stderr, "bbs preview: %v\n", err)
		return exitUsage
	}

	cfg, screens, ds := load(fsys)
	writeDiags(stderr, ds)
	if ds.HasErrors() {
		return exitError
	}
	s, ok := screens[name]
	if !ok {
		_, _ = fmt.Fprintf(stderr, "bbs preview: unknown screen %q\n", name)
		return exitUsage
	}
	env := &screen.Env{Term: term, Items: map[string]bool{}}
	// The items come from the screen also when the render uses its narrow screen (section 6.1).
	for _, it := range s.Items {
		if it.When == nil || it.When.Eval(env) {
			env.Items[it.Key] = true
		}
	}
	pages := screen.Render(s, env, cfg.Theme, cfg.More)
	if err := screen.WriteDump(stdout, term.Cols, term.Rows, pages); err != nil {
		_, _ = fmt.Fprintf(stderr, "bbs preview: %v\n", err)
		return exitError
	}
	return 0
}

// parseSize sets the size of term from "<cols>x<rows>". term.Check checks the ranges.
func parseSize(v string, term *screen.Term) error {
	c, r, _ := strings.Cut(v, "x")
	cols, err := decimal(c)
	if err != nil {
		return fmt.Errorf("cols: %w", err)
	}
	rows, err := decimal(r)
	if err != nil {
		return fmt.Errorf("rows: %w", err)
	}
	term.Cols, term.Rows = cols, rows
	return nil
}

// decimal parses a number of decimal digits only: strconv.Atoi also takes a sign.
func decimal(s string) (int, error) {
	if s == "" || strings.Trim(s, "0123456789") != "" {
		return 0, fmt.Errorf("%q is not a decimal number", s)
	}
	return strconv.Atoi(s)
}

// setField sets term.type or term.colors from "<path>=<value>".
func setField(v string, term *screen.Term) error {
	path, value, ok := strings.Cut(v, "=")
	switch {
	case !ok:
		return fmt.Errorf("expected <path>=<value>")
	case path == "term.type":
		term.Type = value
		return nil
	case path == "term.colors":
		n, err := decimal(value)
		if err != nil {
			return err
		}
		term.Colors = n
		return nil
	}
	return fmt.Errorf("cannot set %q: only term.type and term.colors", path)
}
