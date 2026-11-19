package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/schors/velourbbs/internal/bbs"
	"github.com/schors/velourbbs/internal/bbt"
	"github.com/schors/velourbbs/internal/config"
	"github.com/schors/velourbbs/internal/content"
	"github.com/schors/velourbbs/internal/key"
	"github.com/schors/velourbbs/internal/menu"
	"github.com/schors/velourbbs/internal/provider"
	"github.com/schors/velourbbs/internal/screen"
	"github.com/schors/velourbbs/internal/term"
)

// previewTime gives the default time of bbs.time in velourbbs preview (session-fields, section 4):
// the date and the clock of the reference time of the package time in the zone zone.
// The zone Local gives UTC, so the dump is the same on each machine.
func previewTime(zone *time.Location) time.Time {
	if zone == time.Local {
		zone = time.UTC
	}
	return time.Date(2006, 1, 2, 15, 4, 5, 0, zone)
}

// previewTerm gives the term.* values of velourbbs preview without --size and --set
// (screen-render, section 3.2).
func previewTerm() screen.Term {
	return screen.Term{Type: "xterm", Cols: 80, Rows: 24, Charset: "utf-8", Colors: screen.Colors256, Ambiguous: 1}
}

// previewSession gives the values of session-fields, section 4, without the values of the config:
// bbs.name, bbs.time, session.theme and session.lang.
func previewSession() screen.SessionValues {
	return screen.SessionValues{IP: "127.0.0.1", Online: "00:00:00", Version: version(), Sessions: 1, Uptime: "00:00:00"}
}

// preview runs velourbbs preview (screen-render, section 3.2): it renders a screen and writes its cell dump
// to stdout. With --keys, it runs a session (menu-nav, section 7).
// The flags can come before and after the name.
func preview(args []string, fsys fs.FS, stdout, stderr io.Writer) int {
	term := previewTerm()
	// bbs.name and bbs.time come from the config.
	sv := previewSession()
	svSet := map[string]bool{} // the session and BBS fields of --set and of the sample set
	fl := flag.NewFlagSet("velourbbs preview", flag.ContinueOnError)
	fl.SetOutput(stderr)
	fl.Usage = func() { _, _ = io.WriteString(stderr, usage) }
	fl.Func("size", "", func(v string) error { return parseSize(v, &term) })
	var set termSet
	argVals := map[string]string{} // the arguments of --set arg.<name>
	given := map[string]bool{}     // the paths of --set
	setValue := func(v string) error {
		path, value, ok := strings.Cut(v, "=")
		if ok {
			given[path] = true
			if set, err := setSession(&sv, path, value); set {
				svSet[path] = true
				return err
			}
		}
		return setField(v, &term, &set, argVals)
	}
	fl.Func("set", "", setValue)
	sampleName, sampleGiven := "", false
	fl.Func("sample", "", func(v string) error {
		sampleName, sampleGiven = v, true
		return nil
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
	// The values of the sample set come before the values of --set (screen-render, section 3.2).
	var sample *sampleSet
	if sampleGiven {
		var msgs []string
		sample, msgs = loadSample(fsys, sampleName)
		for _, path := range sample.paths() {
			v := path + "=" + sample.fields[path]
			err := checkSetValue(v)
			if err == nil && !given[path] {
				err = setValue(v)
			}
			if err != nil {
				msgs = append(msgs, fmt.Sprintf("%s: fields: %v", sampleFile, err))
			}
		}
		for _, msg := range msgs {
			_, _ = fmt.Fprintf(stderr, "velourbbs preview: %s\n", msg)
		}
		if msgs != nil {
			return exitUsage
		}
	}
	if err := term.Check(); err != nil {
		_, _ = fmt.Fprintf(stderr, "velourbbs preview: %v\n", err)
		return exitUsage
	}

	m, ds := bbs.Load(fsys)
	writeDiags(stderr, ds)
	if ds.HasErrors() {
		return exitError
	}
	if !svSet["bbs.name"] {
		sv.Name = m.Config.Name
	}
	if !svSet["bbs.time"] {
		sv.Time = previewTime(m.Config.Zone).Format(m.Config.Time)
	}
	choice := m.Config.ThemeChoice
	if svSet["session.theme"] { // the theme choice at the start
		choice = config.ThemeChoice(sv.Theme)
	}
	if !svSet["session.lang"] {
		sv.Lang = m.Config.Lang
	} else if !slices.Contains(m.Config.Langs, sv.Lang) {
		_, _ = fmt.Fprintf(stderr, "velourbbs preview: session.lang is %q, not a language of the BBS\n", sv.Lang)
		return exitUsage
	}
	s, ok := m.Sets[sv.Lang][name]
	if session {
		if err := bbs.CheckTarget(m.Sets, sv.Lang, screen.Action{Screen: name}); err != nil {
			_, _ = fmt.Fprintf(stderr, "velourbbs preview: %v\n", err)
			return exitUsage
		}
	} else if !ok {
		_, _ = fmt.Fprintf(stderr, "velourbbs preview: unknown screen %q\n", name)
		return exitUsage
	}
	if errs := bbs.CheckArgs(s, name, argVals); len(errs) > 0 {
		for _, err := range errs {
			_, _ = fmt.Fprintf(stderr, "velourbbs preview: %v\n", err)
		}
		return exitUsage
	}
	if m = argFiles(fsys, m, name, argVals, sample, session, stderr); m == nil {
		return exitError
	}
	s = m.Sets[sv.Lang][name]
	if session {
		t := &keysTerm{term: term, keys: keys, w: stdout}
		ss := menu.New(m, t, name, argVals, choice, sv.Lang, func(*menu.Session) screen.SessionValues { return sv })
		return runKeys(t, ss, stderr)
	}
	cfg := m.Config
	cfg.ThemeChoice = choice
	if err := screen.WriteDump(stdout, menu.Pages(cfg, s, term, argVals, sv)); err != nil {
		_, _ = fmt.Fprintf(stderr, "velourbbs preview: %v\n", err)
		return exitError
	}
	return 0
}

// argFiles loads the content file of the argument of the screen name in the screen set
// of each language, when the screen is a file screen with {arg.<name>}, with the checks
// of file-view, section 7 (session-language, section 8). For a list screen with {arg.<name>} in data,
// it reads the rows of the value (screen-render, section 3.2). The rows of the sample set,
// when it has rows, replace the rows of the screen name, or of each list screen when all is true
// (screen-render, section 3.5). It loads the content files of select of the new rows.
// It merges the diagnostics and the errors of the sets as bbt.SetDiags does. It gives m, or a new model
// with the screens that can show the files and the rows. It writes the errors and gives nil after an error.
// The model of the load stays immutable (session, section 4).
func argFiles(fsys fs.FS, m *bbs.Model, name string, args map[string]string, sample *sampleSet, all bool,
	stderr io.Writer) *bbs.Model {
	langs := m.Config.Langs
	nm := &bbs.Model{Config: m.Config, Sets: maps.Clone(m.Sets)}
	sd := bbt.NewSetDiags(langs)
	var msgs []string                            // the errors in the order of the first set that finds them
	found := map[string][]string{}               // the sets that find each error
	report := func(lang string, errs []string) { // a set reports each error once
		for _, msg := range errs {
			if found[msg] == nil {
				msgs = append(msgs, msg)
			}
			if !slices.Contains(found[msg], lang) {
				found[msg] = append(found[msg], lang)
			}
		}
	}
	put := func(lang string, s *screen.Screen) {
		nm.Sets[lang] = maps.Clone(nm.Sets[lang])
		nm.Sets[lang][s.Name] = s
	}
	newRows := map[string]bool{} // the sets with the rows of a --set value
	failed := false              // each set runs its checks after an error of another set
	for _, lang := range langs {
		s := m.Sets[lang][name] // the check of the sets makes sure that each set has the screen
		var ns *screen.Screen
		var errs []string
		if s.List != nil {
			ns, errs = argRows(fsys, lang, s, args, sd)
			newRows[lang] = ns != nil && ns != s
		} else {
			ns, errs = argFile(fsys, lang, s, args, "arg.", sd)
		}
		report(lang, errs)
		switch {
		case ns == nil:
			failed = true
		case ns != s:
			put(lang, ns)
		}
	}
	for _, lang := range langs {
		names := []string{name}
		if all {
			names = slices.Sorted(maps.Keys(nm.Sets[lang]))
		}
		for _, n := range names {
			s := nm.Sets[lang][n]
			if s.List == nil {
				continue
			}
			var rows []screen.ListRow
			switch {
			case sample != nil && sample.rows != nil:
				var errs []string
				if rows, errs = sample.listRows(s.List.From); errs != nil {
					report(lang, errs)
					failed = true
					continue
				}
				s = withRows(s, rows)
				put(lang, s)
			case n == name && newRows[lang]:
				rows = s.List.List(args)
			default:
				continue
			}
			target, errs := selectFiles(fsys, lang, nm, s, rows, sd)
			report(lang, errs)
			switch {
			case errs != nil:
				failed = true
			case target != nil:
				tl := lang
				if a := s.List.Select.Do; a.Lang != "" {
					tl = a.Lang
				}
				put(tl, target)
			}
		}
	}
	ds := sd.Diags()
	writeDiags(stderr, ds)
	failed = failed || ds.HasErrors() // a file with errors in it gives no content and no message
	for _, msg := range msgs {
		if len(found[msg]) == len(langs) {
			_, _ = fmt.Fprintf(stderr, "velourbbs preview: %s\n", msg)
			continue
		}
		for _, lang := range found[msg] {
			_, _ = fmt.Fprintf(stderr, "velourbbs preview: %s (language %q)\n", msg, lang)
		}
	}
	if failed {
		return nil
	}
	return nm
}

// argFile loads the content file of the argument of the file screen s in the set lang.
// It gives s when it needs no file, a new screen with the file, or nil and the errors.
// Each error starts with prefix and the argument name. It adds the diagnostics of the file to sd.
func argFile(fsys fs.FS, lang string, s *screen.Screen, args map[string]string, prefix string,
	sd *bbt.SetDiags) (*screen.Screen, []string) {
	f := s.FileScreen
	if f == nil || f.Arg == "" || f.Contents[args[f.Arg]] != nil {
		return s, nil
	}
	p := args[f.Arg]
	c, ds, err := content.Load(fsys, lang, p)
	sd.Add(lang, ds)
	var errs []error
	if err != nil {
		errs = []error{err}
	} else if c != nil {
		errs = f.Check(p, c)
	}
	var msgs []string
	for _, err := range errs {
		msgs = append(msgs, fmt.Sprintf("%s%s: %v", prefix, f.Arg, err))
	}
	if c == nil || len(errs) > 0 {
		return nil, msgs
	}
	nf := *f
	nf.Contents = maps.Clone(f.Contents)
	nf.Contents[p] = c
	ns := *s
	ns.FileScreen = &nf
	return &ns, nil
}

// argRows reads the rows of the list screen s for the arguments args in the set lang.
// It gives s when it has the rows, a new screen with the rows, or nil and the errors.
// It adds the diagnostics of the content files to sd.
func argRows(fsys fs.FS, lang string, s *screen.Screen, args map[string]string, sd *bbt.SetDiags) (*screen.Screen, []string) {
	l := s.List
	values := l.Values(args)
	key := screen.ListKey(values)
	if _, ok := l.Lists[key]; ok {
		return s, nil
	}
	p, _ := provider.Lookup(l.From)
	diags := false
	rows, errs := p.Rows(fsys, lang, values, func(path string) (*screen.Content, error) {
		c, ds, err := content.Load(fsys, lang, path)
		sd.Add(lang, ds)
		diags = diags || ds.HasErrors()
		return c, err
	})
	var msgs []string
	for _, err := range errs {
		msgs = append(msgs, fmt.Sprintf("data: args: %v", err))
	}
	if len(errs) > 0 || diags {
		return nil, msgs
	}
	nl := *l
	nl.Lists = maps.Clone(l.Lists)
	nl.Lists[key] = rows
	ns := *s
	ns.List = &nl
	return &ns, nil
}

// withRows gives the list screen s with the rows of each provider value replaced with rows.
func withRows(s *screen.Screen, rows []screen.ListRow) *screen.Screen {
	nl := *s.List
	nl.Lists = map[string][]screen.ListRow{}
	for key := range s.List.Lists {
		nl.Lists[key] = rows
	}
	ns := *s
	ns.List = &nl
	return &ns
}

// selectFiles loads the content file of the value of select of each row of rows of the list screen s,
// when the target of select is a file screen (list-view, section 8, and screen-render, sections 3.2
// and 3.5). rows are rows that the load did not read. It gives the target in m with the new files,
// else nil, and the errors. It adds the diagnostics of the files to sd.
func selectFiles(fsys fs.FS, lang string, m *bbs.Model, s *screen.Screen, rows []screen.ListRow,
	sd *bbt.SetDiags) (*screen.Screen, []string) {
	sel := s.List.Select
	if sel == nil || !sel.Do.Op.HasTarget() {
		return nil, nil
	}
	t := bbs.Target(m.Sets, lang, sel.Do) // the check of the sets makes sure that it exists
	// The target is in the set of its language: its files too.
	if sel.Do.Lang != "" {
		lang = sel.Do.Lang
	}
	nt := t
	var msgs []string
	for i, r := range rows {
		next, errs := argFile(fsys, lang, nt, sel.Action(r, i+1).Args, "select: args: ", sd)
		msgs = append(msgs, errs...)
		if next != nil {
			nt = next
		}
	}
	if len(msgs) > 0 {
		return nil, msgs
	}
	if nt == t {
		return nil, nil
	}
	return nt, nil
}

// checkSetValue checks the value of --set "<path>=<value>" with the default values of the preview,
// without a change of the values of the preview. The checks of the arguments of the screen
// come after the load.
func checkSetValue(v string) error {
	term, sv := previewTerm(), previewSession()
	path, value, ok := strings.Cut(v, "=")
	if ok {
		if set, err := setSession(&sv, path, value); set {
			return err
		}
	}
	if err := setField(v, &term, &termSet{}, map[string]string{}); err != nil {
		return err
	}
	return term.Check()
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
// session.theme sets the theme choice at the start: preview gives it to the config.
// session.lang sets the session language at the start.
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
	case "session.theme":
		switch c := config.ThemeChoice(value); c {
		case config.ThemeDark, config.ThemeLight, config.ThemeAuto:
			sv.Theme = value
			return true, nil
		}
		return true, fmt.Errorf("session.theme is %q, must be dark, light or auto", value)
	case "session.lang": // preview checks the languages of the BBS after the load
		if !screen.IsLang(value) {
			return true, fmt.Errorf("session.lang is %q, must be a language code", value)
		}
		sv.Lang = value
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

// Key ignores the context of the session (session, section 6).
func (t *keysTerm) Key(context.Context) (string, error) {
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

// Update writes the whole page: the dump of menu-nav, section 7.
func (t *keysTerm) Update(p screen.Page) error {
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

// runKeys runs the session ss of velourbbs preview --keys on t and gives the exit status.
// The end of the key list ends the session. A key after the end of the session at logoff is an error.
func runKeys(t *keysTerm, ss *menu.Session, stderr io.Writer) int {
	err := ss.Run(context.Background())
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
