package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"strconv"
	"strings"
	"unicode"

	"github.com/schors/velourbbs/internal/config"
	"github.com/schors/velourbbs/internal/key"
	"github.com/schors/velourbbs/internal/menu"
	"github.com/schors/velourbbs/internal/screen"
	"github.com/schors/velourbbs/internal/term"
)

// preview runs velourbbs preview (screen-render, section 3.2): it renders a screen and writes its cell dump
// to stdout. With --keys, it runs a session (menu-nav, section 7).
// The flags can come before and after the name.
func preview(args []string, fsys fs.FS, stdout, stderr io.Writer) int {
	term := screen.Term{Type: "xterm", Cols: 80, Rows: 24, Charset: "utf-8", Colors: screen.Colors256}
	fl := flag.NewFlagSet("velourbbs preview", flag.ContinueOnError)
	fl.SetOutput(stderr)
	fl.Usage = func() { _, _ = io.WriteString(stderr, usage) }
	fl.Func("size", "", func(v string) error { return parseSize(v, &term) })
	colorsSet := false
	fl.Func("set", "", func(v string) error { return setField(v, &term, &colorsSet) })
	var keys []string
	session := false
	fl.Func("keys", "", func(v string) error {
		session = true
		var err error
		keys, err = parseKeys(v)
		return err
	})

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
		_, _ = io.WriteString(stderr, "velourbbs preview: expected one screen name\n"+usage)
		return exitUsage
	}
	name := names[0]
	if err := term.Check(); err != nil {
		_, _ = fmt.Fprintf(stderr, "velourbbs preview: %v\n", err)
		return exitUsage
	}

	cfg, screens, ds := load(fsys)
	writeDiags(stderr, ds)
	if ds.HasErrors() {
		return exitError
	}
	if session {
		if err := menu.CheckTarget(screens, name); err != nil {
			_, _ = fmt.Fprintf(stderr, "velourbbs preview: %v\n", err)
			return exitUsage
		}
		return runKeys(&keysTerm{term: term, keys: keys, w: stdout}, cfg, screens, name, stderr)
	}
	s, ok := screens[name]
	if !ok {
		_, _ = fmt.Fprintf(stderr, "velourbbs preview: unknown screen %q\n", name)
		return exitUsage
	}
	env, _ := menu.Enabled(s, term)
	if err := screen.WriteDump(stdout, screen.Render(s, env, cfg.SessionTheme(term.Colors), cfg.More)); err != nil {
		_, _ = fmt.Fprintf(stderr, "velourbbs preview: %v\n", err)
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

// setField sets term.type or term.colors of t from "<path>=<value>".
// term.type also gives term.colors, unless colorsSet: an explicit term.colors overrides it.
func setField(v string, t *screen.Term, colorsSet *bool) error {
	path, value, ok := strings.Cut(v, "=")
	switch {
	case !ok:
		return fmt.Errorf("expected <path>=<value>")
	case path == "term.type":
		typ, colors := term.TypeColors(value)
		t.Type = typ
		if !*colorsSet {
			t.Colors = colors
		}
		return nil
	case path == "term.colors":
		n, err := decimal(value)
		if err != nil {
			return err
		}
		t.Colors, *colorsSet = n, true
		return nil
	}
	return fmt.Errorf("cannot set %q: only term.type and term.colors", path)
}

// parseKeys parses the key list of --keys (menu-nav, section 7). An element is a key name
// except ANY, or one other printable character except the comma and the space.
func parseKeys(v string) ([]string, error) {
	if v == "" {
		return nil, nil
	}
	var keys []string
	for _, e := range strings.Split(v, ",") {
		k, ok := key.Parse(e)
		if !ok || k == key.Any {
			r := []rune(e)
			if len(r) != 1 || r[0] == ' ' || !unicode.IsPrint(r[0]) {
				return nil, fmt.Errorf("invalid key %q", e)
			}
			k = e
		}
		keys = append(keys, k)
	}
	return keys, nil
}

// errNoKeys is the error of keysTerm.Key at the end of the key list.
var errNoKeys = errors.New("no more keys")

// keysTerm is the terminal of velourbbs preview --keys. It takes the keys from a list
// and writes the events of menu-nav, section 7, to w.
type keysTerm struct {
	term screen.Term
	keys []string
	w    io.Writer
	page int // the number of the last page of the screen
}

func (t *keysTerm) Term() screen.Term { return t.term }

func (t *keysTerm) Key() (string, error) {
	if len(t.keys) == 0 {
		return "", errNoKeys
	}
	k := t.keys[0]
	t.keys = t.keys[1:]
	return k, nil
}

func (t *keysTerm) Page(p screen.Page) error {
	t.page++
	return screen.WritePage(t.w, t.page, p)
}

func (t *keysTerm) Screen(name string) error {
	t.page = 0
	_, err := fmt.Fprintf(t.w, "screen %s\n", name)
	return err
}

func (t *keysTerm) KeyEvent(k string, ignored bool) error {
	if ignored {
		k += " ignored"
	}
	_, err := fmt.Fprintf(t.w, "key %s\n", k)
	return err
}

func (t *keysTerm) Logoff() error {
	_, err := io.WriteString(t.w, "logoff\n")
	return err
}

// runKeys runs the session of velourbbs preview --keys from the screen start
// and gives the exit status. The end of the key list ends the session.
// A key after the end of the session at logoff is an error.
func runKeys(t *keysTerm, cfg config.Config, screens map[string]*screen.Screen, start string, stderr io.Writer) int {
	err := menu.Run(t, cfg, screens, start)
	switch {
	case errors.Is(err, errNoKeys):
		return 0
	case err != nil:
		_, _ = fmt.Fprintf(stderr, "velourbbs preview: %v\n", err)
		return exitError
	case len(t.keys) > 0:
		_, _ = fmt.Fprintf(stderr, "velourbbs preview: key %q after logoff\n", key.Map(t.keys[0]))
		return exitError
	}
	return 0
}
