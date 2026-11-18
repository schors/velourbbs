package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/schors/velourbbs/internal/config"
	"github.com/schors/velourbbs/internal/content"
	"github.com/schors/velourbbs/internal/key"
	"github.com/schors/velourbbs/internal/menu"
	"github.com/schors/velourbbs/internal/screen"
	"github.com/schors/velourbbs/internal/term"
)

// previewTime is the default time of bbs.time in velourbbs preview (session-fields, section 4).
var previewTime = time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)

// preview runs velourbbs preview (screen-render, section 3.2): it renders a screen and writes its cell dump
// to stdout. With --keys, it runs a session (menu-nav, section 7).
// The flags can come before and after the name.
func preview(args []string, fsys fs.FS, stdout, stderr io.Writer) int {
	term := screen.Term{Type: "xterm", Cols: 80, Rows: 24, Charset: "utf-8", Colors: screen.Colors256, Ambiguous: 1}
	// The values of session-fields, section 4. bbs.name and bbs.time come from the config.
	sv := screen.SessionValues{IP: "127.0.0.1", Online: "00:00:00", Version: version(), Sessions: 1, Uptime: "00:00:00"}
	svSet := map[string]bool{} // the session and BBS fields of --set
	fl := flag.NewFlagSet("velourbbs preview", flag.ContinueOnError)
	fl.SetOutput(stderr)
	fl.Usage = func() { _, _ = io.WriteString(stderr, usage) }
	fl.Func("size", "", func(v string) error { return parseSize(v, &term) })
	var set termSet
	argVals := map[string]string{} // the arguments of --set arg.<name>
	fl.Func("set", "", func(v string) error {
		path, value, ok := strings.Cut(v, "=")
		if ok {
			if set, err := setSession(&sv, path, value); set {
				svSet[path] = true
				return err
			}
		}
		return setField(v, &term, &set, argVals)
	})
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
	if !svSet["bbs.name"] {
		sv.Name = cfg.Name
	}
	if !svSet["bbs.time"] {
		sv.Time = previewTime.Format(cfg.Time)
	}
	s, ok := screens[name]
	if session {
		if err := menu.CheckTarget(screens, name); err != nil {
			_, _ = fmt.Fprintf(stderr, "velourbbs preview: %v\n", err)
			return exitUsage
		}
	} else if !ok {
		_, _ = fmt.Fprintf(stderr, "velourbbs preview: unknown screen %q\n", name)
		return exitUsage
	}
	if errs := menu.CheckArgs(s, argVals); len(errs) > 0 {
		for _, err := range errs {
			_, _ = fmt.Fprintf(stderr, "velourbbs preview: %v\n", err)
		}
		return exitUsage
	}
	if f := s.FileScreen; f != nil && f.Arg != "" && f.Contents[argVals[f.Arg]] == nil {
		ns, status := argFile(fsys, s, argVals[f.Arg], stderr)
		if ns == nil {
			return status
		}
		screens = maps.Clone(screens)
		s, screens[name] = ns, ns
	}
	if session {
		return runKeys(&keysTerm{term: term, keys: keys, w: stdout}, cfg, screens, name, argVals, sv, stderr)
	}
	if err := screen.WriteDump(stdout, menu.Pages(cfg, s, term, argVals, sv)); err != nil {
		_, _ = fmt.Fprintf(stderr, "velourbbs preview: %v\n", err)
		return exitError
	}
	return 0
}

// argFile loads the content file p of the argument of the file screen s with the checks
// of file-view, section 7. It gives a copy of s that can show p, or nil and the exit status
// after an error. The model of the load stays immutable.
func argFile(fsys fs.FS, s *screen.Screen, p string, stderr io.Writer) (*screen.Screen, int) {
	f := s.FileScreen
	c, ds, err := content.Load(fsys, p)
	writeDiags(stderr, ds)
	var errs []error
	if err != nil {
		errs = []error{err}
	} else if c != nil {
		errs = f.Check(p, c)
	}
	for _, err := range errs {
		_, _ = fmt.Fprintf(stderr, "velourbbs preview: arg.%s: %v\n", f.Arg, err)
	}
	if c == nil || len(errs) > 0 {
		return nil, exitError
	}
	nf := *f
	nf.Contents = maps.Clone(f.Contents)
	nf.Contents[p] = c
	ns := *s
	ns.FileScreen = &nf
	return &ns, 0
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

// termSet tells which term.* fields an explicit --set gave. Such a field is not
// derived from another field.
type termSet struct {
	colors     bool // term.colors
	background bool // term.background
}

// setField sets a term.* field of t, or an argument of args, from "<path>=<value>".
// term.type also gives term.colors and term.bg_color gives term.background,
// unless set tells that an explicit --set gave them.
func setField(v string, t *screen.Term, set *termSet, args map[string]string) error {
	path, value, ok := strings.Cut(v, "=")
	name, isArg := strings.CutPrefix(path, "arg.")
	switch {
	case !ok:
		return fmt.Errorf("expected <path>=<value>")
	case isArg:
		if err := screen.CheckArg(name, value); err != nil {
			return err
		}
		args[name] = value
		return nil
	case path == "term.type":
		typ, colors := term.TypeColors(value)
		t.Type = typ
		if !set.colors {
			t.Colors = colors
		}
		return nil
	case path == "term.colors":
		n, err := decimal(value)
		if err != nil {
			return err
		}
		t.Colors, set.colors = n, true
		return nil
	case path == "term.name":
		t.Name = term.CleanName(value)
		return nil
	case path == "term.images":
		if value != "" && value != screen.ImagesSixel {
			return fmt.Errorf("term.images is %q, must be %s or empty", value, screen.ImagesSixel)
		}
		t.Images = value
		return nil
	case path == "term.ambiguous":
		if value != "1" && value != "2" {
			return fmt.Errorf("term.ambiguous is %q, must be 1 or 2", value)
		}
		t.Ambiguous = int(value[0] - '0')
		return nil
	case path == "term.bg_color" || path == "term.fg_color":
		if value != "" && !isRGB(value) {
			return fmt.Errorf("%s is %q, must be #rrggbb or empty", path, value)
		}
		value = strings.ToLower(value)
		if path == "term.fg_color" {
			t.FgColor = value
			return nil
		}
		t.BgColor = value
		if !set.background {
			t.Background = term.Background(value)
		}
		return nil
	case path == "term.background":
		if value != "" && value != term.BackgroundDark && value != term.BackgroundLight {
			return fmt.Errorf("term.background is %q, must be dark, light or empty", value)
		}
		t.Background, set.background = value, true
		return nil
	}
	return fmt.Errorf("cannot set %q: only term.type, term.colors, term.name, term.images, term.ambiguous, "+
		"term.bg_color, term.fg_color, term.background, arg.<name>, session.* and bbs.*", path)
}

// isRGB reports whether s is #rrggbb with hex digits in each case.
func isRGB(s string) bool {
	hex, ok := strings.CutPrefix(s, "#")
	return ok && len(hex) == 6 && strings.Trim(hex, "0123456789abcdefABCDEF") == ""
}

// setSession sets the session or BBS field path of sv to value (session-fields, section 4).
// It reports whether path is such a field.
func setSession(sv *screen.SessionValues, path, value string) (bool, error) {
	var p *string
	switch path {
	case "bbs.sessions":
		n, err := decimal(value)
		if err != nil {
			return true, fmt.Errorf("%s: %w", path, err)
		}
		sv.Sessions = n
		return true, nil
	case "session.ip":
		p = &sv.IP
	case "session.online":
		p = &sv.Online
	case "bbs.name":
		p = &sv.Name
	case "bbs.version":
		p = &sv.Version
	case "bbs.time":
		p = &sv.Time
	case "bbs.uptime":
		p = &sv.Uptime
	default:
		return false, nil
	}
	if err := screen.CheckValue(value); err != nil {
		return true, fmt.Errorf("%s: %w", path, err)
	}
	*p = value
	return true, nil
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
	return screen.WritePage(t.w, p)
}

func (t *keysTerm) Screen(name string) error {
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
// with the arguments args and the session and BBS fields sv, and gives the exit status.
// The end of the key list ends the session. A key after the end of the session at logoff is an error.
func runKeys(t *keysTerm, cfg config.Config, screens map[string]*screen.Screen, start string, args map[string]string,
	sv screen.SessionValues, stderr io.Writer) int {
	err := menu.Run(t, cfg, screens, start, args, func() screen.SessionValues { return sv })
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
