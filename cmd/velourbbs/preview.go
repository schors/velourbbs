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
	return screen.Term{Type: "xterm", Cols: 80, Rows: 24, Colors: screen.Colors256, Ambiguous: 1}
}

// previewSession gives the values of session-fields, section 4, without the values of the config:
// bbs.name, bbs.time, session.theme and session.lang.
func previewSession() screen.SessionValues {
	return screen.SessionValues{IP: "127.0.0.1", Online: "00:00:00", Version: version(), Sessions: 1, Uptime: "00:00:00"}
}

// preview runs velourbbs preview (screen-render, section 3.2): it renders a screen and writes its cell dump
// to stdout, or with --view its line form (section 3.7). With --keys, it runs a session (menu-nav, section 7).
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
	view := fl.Bool("view", false, "")
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
	if errs := bbs.CheckArgs(s, name, slices.Sorted(maps.Keys(argVals))); len(errs) > 0 {
		for _, err := range errs {
			_, _ = fmt.Fprintf(stderr, "velourbbs preview: %v\n", err)
		}
		return exitUsage
	}
	if m = argFiles(fsys, m, name, argVals, sample, session, stderr); m == nil {
		return exitError
	}
	s = m.Sets[sv.Lang][name]
	writePage := pageWriter(*view, term.Colors)
	if session {
		t := &keysTerm{term: term, keys: keys, w: stdout, write: writePage}
		ss := menu.New(m, t, name, argVals, choice, sv.Lang, func(*menu.Session) screen.SessionValues { return sv })
		return runKeys(t, ss, stderr)
	}
	cfg := m.Config
	cfg.ThemeChoice = choice
	for p := range menu.Pages(cfg, s, term, argVals, sv) {
		if err := writePage(stdout, p); err != nil {
			_, _ = fmt.Fprintf(stderr, "velourbbs preview: %v\n", err)
			return exitError
		}
	}
	return 0
}

// pageWriter gives the writer of a page of velourbbs preview: the cell dump (screen-render, section 3.3),
// or with view the line form for term.colors = colors (section 3.7).
func pageWriter(view bool, colors int) func(io.Writer, screen.Page) error {
	if !view {
		return screen.WritePage
	}
	return func(w io.Writer, p screen.Page) error {
		_, err := w.Write(term.AppendLines(nil, p, colors))
		return err
	}
}

// argFiles loads the rows and the content files of each screen that the values of the arguments args
// of the screen name reach, in the screen set of each language, also through the {arg.<name>}
// of other screens (screen-render, section 3.2). The rows of the sample set, when it has rows,
// replace the rows of the screen name, or of each list screen when all is true (screen-render, section 3.5).
// Then argFiles loads the content files that the values of the rows reach.
// It merges the diagnostics and the errors of the sets as bbt.SetDiags does. It gives a new model
// with the screens that can show the files and the rows. It writes the errors and gives nil after an error.
// The model of the load stays immutable (session, section 4).
func argFiles(fsys fs.FS, m *bbs.Model, name string, args map[string]string, sample *sampleSet, all bool,
	stderr io.Writer) *bbs.Model {
	langs := m.Config.Langs
	nm := &bbs.Model{Config: m.Config, Sets: cloneSets(m.Sets)}
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
	reportValues := func(errs []bbt.ValueError) {
		for _, e := range errs {
			o := e.Origin
			if o.File == "" { // a value of --set
				report(e.Lang, []string{fmt.Sprintf("arg.%s: %v", o.Arg, e.Err)})
			} else {
				report(e.Lang, []string{fmt.Sprintf("%s: args: %s: %v", o.Label, o.Arg, e.Err)})
			}
		}
	}
	load := func(lang, p string) (*screen.Content, error) {
		c, ds, err := content.Load(fsys, lang, p)
		sd.Add(lang, ds)
		return c, err
	}
	seeds := bbt.Reached{name: {}}
	for arg, v := range args {
		seeds[name][arg] = []bbt.Origin{{Arg: arg, Text: v}}
	}
	r := bbt.Reach(nm.Sets, langs, seeds)
	reportValues(bbt.LoadRows(fsys, nm.Sets, langs, r, load))
	failed := len(msgs) > 0 // each set runs its checks after an error of another set
	if sample != nil && sample.rows != nil {
		for _, lang := range langs {
			names := []string{name}
			if all {
				names = slices.Sorted(maps.Keys(nm.Sets[lang]))
			}
			for _, n := range names {
				ls := nm.Sets[lang][n].List
				if ls == nil {
					continue
				}
				rows, errs := sample.listRows(ls.From)
				if errs != nil {
					report(lang, errs)
					failed = true
					continue
				}
				for key := range ls.Lists {
					ls.Lists[key] = rows
				}
			}
		}
	}
	reportValues(bbt.LoadFiles(nm.Sets, langs, r, load))
	ds := sd.Diags()
	writeDiags(stderr, ds)
	failed = failed || len(msgs) > 0 || ds.HasErrors() // a file with errors in it gives no content and no message
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

// cloneSets gives a copy of sets whose screens have their own maps of content files and of lists.
func cloneSets(sets bbt.Sets) bbt.Sets {
	out := bbt.Sets{}
	for lang, set := range sets {
		ns := make(map[string]*screen.Screen, len(set))
		for name, s := range set {
			c := *s
			if f := s.FileScreen; f != nil {
				nf := *f
				nf.Contents = maps.Clone(f.Contents)
				c.FileScreen = &nf
			}
			if l := s.List; l != nil {
				nl := *l
				nl.Lists = maps.Clone(l.Lists)
				c.List = &nl
			}
			ns[name] = &c
		}
		out[lang] = ns
	}
	return out
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
	case path == "term.ambiguous":
		if value != "1" && value != "2" {
			return fmt.Errorf("term.ambiguous is %q, must be 1 or 2", value)
		}
		t.Ambiguous = int(value[0] - '0')
		return nil
	case path == "term.bg_color":
		if value != "" && !isRGB(value) {
			return fmt.Errorf("term.bg_color is %q, must be #rrggbb or empty", value)
		}
		t.BgColor = strings.ToLower(value)
		if !set.background {
			t.Background = term.Background(t.BgColor)
		}
		return nil
	case path == "term.background":
		if value != "" && value != term.BackgroundDark && value != term.BackgroundLight {
			return fmt.Errorf("term.background is %q, must be dark, light or empty", value)
		}
		t.Background, set.background = value, true
		return nil
	}
	return fmt.Errorf("cannot set %q: only term.type, term.colors, term.name, term.ambiguous, "+
		"term.bg_color, term.background, arg.<name>, session.* and bbs.*", path)
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
// and writes the events of menu-nav, section 7, to w. write writes a page.
type keysTerm struct {
	term  screen.Term
	keys  []string
	w     io.Writer
	write func(io.Writer, screen.Page) error
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
	return t.write(t.w, p)
}

// Update writes the whole page (menu-nav, section 7).
func (t *keysTerm) Update(p screen.Page) error {
	return t.write(t.w, p)
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
