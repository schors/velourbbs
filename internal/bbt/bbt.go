// Package bbt loads the BBS templates of screens/ into the screen model of internal/screen.
// The file pass parses each file once and resolves its fields with the aliases of screens/screens.toml.
// The screen pass expands the includes and the art inserts of each screen.
// Specification: docs/specs/bbt-format.md.
package bbt

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"path"
	"slices"
	"strings"

	"github.com/schors/velourbbs/internal/art"
	"github.com/schors/velourbbs/internal/content"
	"github.com/schors/velourbbs/internal/provider"
	"github.com/schors/velourbbs/internal/screen"
)

// MaxBodyLines is the line limit of the body of a screen after includes (section 9).
const MaxBodyLines = 10000

// Settings is the settings of screens/screens.toml that the load uses.
type Settings struct {
	// Theme is the base theme. The load uses only its names (section 7.3).
	Theme  map[string]screen.Style
	Layout screen.Layout // the layout of [screen]
	Fields screen.Fields // the aliases (section 4.2)
	// Langs is the languages of the BBS, sorted (session-language, section 2). nil gives "en".
	Langs []string
	// LangTables is true when screens/screens.toml has [lang] tables (session-language, section 3).
	LangTables bool
}

// Sets maps the code of each language of the BBS to its screen set: the screens by name
// (session-language, section 4).
type Sets map[string]map[string]*screen.Screen

// Load loads the screen set of each language of the BBS from fsys/screens, with the fragments,
// the art and the content files that the screens use (session-language, sections 4 and 6).
// fsys is the BBS directory. Load merges the diagnostics of the sets with SetDiags.
// The sets are nil when the diagnostics have an error.
func Load(fsys fs.FS, st Settings) (Sets, screen.Diags) {
	if st.Langs == nil {
		st.Langs = []string{"en"}
	}
	sd := NewSetDiags(st.Langs)
	var loaders []*loader
	used := map[string]bool{}
	for _, lang := range st.Langs {
		l := &loader{
			fsys: fsys, lang: lang, st: st,
			diags: map[string]*screen.FileDiags{}, files: map[string]*file{}, arts: map[string]*screen.Art{},
			missing: map[string]bool{}, reported: map[screen.Pos]bool{}, contents: map[string]loaded{},
			found: map[string]string{},
		}
		l.load(used)
		loaders = append(loaders, l)
	}
	loadValues(fsys, loaders, st.Langs)
	for _, l := range loaders {
		l.checkItems(l.units)
		var ds screen.Diags
		for _, d := range l.diags {
			ds = append(ds, d.Diags...)
		}
		ds.Sort()
		sd.Add(l.lang, ds)
	}
	sd.Add("", unused(fsys, st, used))
	ds := sd.Diags()
	if ds.HasErrors() {
		return nil, ds
	}
	sets := Sets{}
	for _, l := range loaders {
		screens := make(map[string]*screen.Screen, len(l.units))
		for _, u := range l.units {
			screens[u.s.Name] = u.s
		}
		sets[l.lang] = screens
	}
	return sets, ds
}

// load loads the screens of the screen set of l.lang and adds the files that it reads to used.
// A screen of the set is a screen file of screens/<lang>/, else of screens/ (session-language, section 4).
func (l *loader) load(used map[string]bool) {
	byName := map[string]string{} // the file of each screen name
	for _, dir := range []string{"screens", "screens/" + l.lang} {
		entries, err := fs.ReadDir(l.fsys, dir)
		if err != nil && (dir == "screens" || !errors.Is(err, fs.ErrNotExist)) {
			l.fd(dir).Errorf(1, 1, "%v", err)
		}
		for _, e := range entries {
			base, ok := strings.CutSuffix(e.Name(), ".bbt")
			if !ok || strings.HasPrefix(base, "_") {
				continue
			}
			name := path.Join(dir, e.Name())
			if e.IsDir() {
				// A directory of the language takes the place of the common screen: unused reports it.
				delete(byName, base)
				continue
			}
			if !screen.IsName(base) {
				used[name] = true // the error is its diagnostic (screen-render, section 3.1)
				l.fd(name).Errorf(1, 1, "invalid screen name %q", base)
				continue
			}
			// An entry of the language takes the place of the common screen.
			byName[base] = name
		}
	}
	// The order of the file names: the diagnostics at one position keep the order of the screens.
	for _, base := range slices.SortedFunc(maps.Keys(byName), func(a, b string) int {
		return strings.Compare(a+".bbt", b+".bbt")
	}) {
		f, missing := l.file(byName[base], false)
		if missing { // a broken symlink (bbt-format, section 2)
			l.fd(byName[base]).Errorf(1, 1, "not a regular file")
		}
		if f != nil {
			l.units = append(l.units, &unit{s: &screen.Screen{Name: base, File: byName[base]}, f: f})
		}
	}
	for _, u := range l.units {
		l.screen(u)
	}
	for name := range l.files {
		used[name] = true
	}
	for name := range l.arts {
		used[name] = true
	}
}

// unused reports each file of screens/ and of the directory of each language that no screen set uses,
// and in a BBS with [lang] tables, each directory of screens/ that is not a language
// (screen-render, section 3.1, and session-language, section 3). used is the files that the sets read.
func unused(fsys fs.FS, st Settings, used map[string]bool) screen.Diags {
	var ds screen.Diags
	for _, dir := range append([]string{"screens"}, st.Langs...) {
		if dir != "screens" {
			dir = "screens/" + dir
		}
		entries, _ := fs.ReadDir(fsys, dir) // the load of the sets reports the error
		for _, e := range entries {
			name := path.Join(dir, e.Name())
			switch {
			case !strings.HasSuffix(e.Name(), ".bbt") && !strings.HasSuffix(e.Name(), ".ans"):
				if dir == "screens" && e.IsDir() && st.LangTables && !slices.Contains(st.Langs, e.Name()) {
					ds.Warnf(name, 1, 1, "%q is not a language of [lang]", e.Name())
				}
				continue
			case used[name]:
				continue // its load reports a file that is not a regular file (section 2)
			}
			if info, err := fs.Stat(fsys, name); err == nil && !info.Mode().IsRegular() {
				ds.Errorf(name, 1, 1, "not a regular file")
			} else {
				ds.Warnf(name, 1, 1, "no screen uses this file")
			}
		}
	}
	return ds
}

// loader loads the screen set of one language.
type loader struct {
	fsys  fs.FS
	lang  string // the language of the screen set
	st    Settings
	units []*unit // the screens of the set
	diags map[string]*screen.FileDiags
	files map[string]*file       // nil for a file that cannot be read
	arts  map[string]*screen.Art // nil for art that cannot be loaded
	// missing holds the .bbt and .ans files that do not exist.
	missing map[string]bool
	// reported holds the include and art insert positions with an error:
	// the screens that use a fragment report its errors once.
	reported map[screen.Pos]bool
	contents map[string]loaded // the content files by path
	found    map[string]string // the path of each name of a fragment or an art file (session-language, section 4)
}

// find gives the path of the fragment or art file name in the language of the set.
func (l *loader) find(name string) string {
	p, ok := l.found[name]
	if !ok {
		p = screen.Find(l.fsys, "screens", l.lang, name)
		l.found[name] = p
	}
	return p
}

// loaded is the result of the load of a content file: the content, else nil, and the error
// of the whole file.
type loaded struct {
	c   *screen.Content
	err error
}

// fd gives the diagnostics of the file name.
func (l *loader) fd(name string) *screen.FileDiags {
	d, ok := l.diags[name]
	if !ok {
		d = &screen.FileDiags{File: name}
		l.diags[name] = d
	}
	return d
}

// file reads and parses a .bbt file once. missing is true when the file does not exist;
// the caller reports it. Each other read error is an error at 1:1 of the file.
func (l *loader) file(name string, fragment bool) (f *file, missing bool) {
	if f, ok := l.files[name]; ok {
		return f, l.missing[name]
	}
	data, err := screen.ReadFile(l.fsys, name)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		l.files[name] = nil
		l.missing[name] = true
		return nil, true
	case err != nil:
		l.fd(name).Errorf(1, 1, "%v", err)
		l.files[name] = nil
		return nil, false
	}
	f = parseFile(l.fd(name), data, l.st.Theme, l.st.Fields, l.st.Langs, fragment)
	l.files[name] = f
	return f, false
}

// art loads an .ans file once. missing is true when the file does not exist.
func (l *loader) art(name string) (a *screen.Art, missing bool) {
	if a, ok := l.arts[name]; ok {
		return a, l.missing[name]
	}
	if _, err := fs.Stat(l.fsys, name); errors.Is(err, fs.ErrNotExist) {
		l.arts[name] = nil
		l.missing[name] = true
		return nil, true
	}
	a, ds := art.Load(l.fsys, name, l.st.Fields)
	l.fd(name).Diags = append(l.fd(name).Diags, ds...)
	l.arts[name] = a
	return a, false
}

// errorOnce reports an error at pos once for all screens.
func (l *loader) errorOnce(pos screen.Pos, format string, args ...any) {
	if !l.reported[pos] {
		l.reported[pos] = true
		errorAt(l.fd(pos.File), pos, format, args...)
	}
}

func errorAt(d *screen.FileDiags, pos screen.Pos, format string, args ...any) {
	d.Errorf(pos.Line, pos.Col, "%s", fmt.Sprintf(format, args...))
}

// content loads the content file p once (file-view, section 2). Its errors in the file
// appear once. The error is an error of the whole file; the caller reports it.
// The content is nil after an error.
func (l *loader) content(p string) (*screen.Content, error) {
	if r, ok := l.contents[p]; ok {
		return r.c, r.err
	}
	c, ds, err := content.Load(l.fsys, l.lang, p)
	for _, d := range ds {
		l.fd(d.File).Diags = append(l.fd(d.File).Diags, d)
	}
	l.contents[p] = loaded{c, err}
	return c, err
}

// literalFile loads the content file of the key file of the screen u (file-view, section 7).
// An error of lines or numbers has the position of its key.
func (l *loader) literalFile(u *unit, f *screen.FileSpec) {
	c, err := l.content(f.Path)
	if err != nil {
		l.errorf(u, u.keyPos("file"), false, "file: %v", err)
		return
	}
	if c == nil { // the errors are in the content file
		return
	}
	errs := f.Check(f.Path, c)
	for _, err := range errs {
		l.errorf(u, u.keyPos(keyOf(err)), false, "%v", err)
	}
	if len(errs) == 0 {
		f.Contents[f.Path] = c
	}
}

// literalRows reads the rows of the list screen u whose data has no {arg.<name>} value
// (list-view, section 4). An error is an error of the key data, at the value of its argument.
func (l *loader) literalRows(u *unit, ls *screen.ListSpec) {
	for _, err := range l.rows(ls, ls.Args) {
		l.errorf(u, u.keyPos("data", "args", keyOf(err)), false, "data: args: %v", err)
	}
}

// keyOf gives the key of a screen.KeyError, else "".
func keyOf(err error) string {
	var ke screen.KeyError
	if errors.As(err, &ke) {
		return ke.Key
	}
	return ""
}

// rows reads the rows of the provider of ls for the provider arguments values into ls.Lists
// and gives the errors.
func (l *loader) rows(ls *screen.ListSpec, values map[string]string) []error {
	p, _ := provider.Lookup(ls.From)
	rows, errs := p.Rows(l.fsys, l.lang, values, l.content)
	if len(errs) == 0 {
		ls.Lists[screen.ListKey(values)] = rows
	}
	return errs
}

// loadValues loads the content file of each value of an argument of a file screen with {arg.<name>},
// and the rows of each value of an argument of data of a list screen (bbt-format, section 8.5).
// Each set loads and checks each value with its own search. An error is an error at the origin
// of the value in the set that finds it (session-language, section 6).
// The rows come first: the {row.<field>} values of select need them.
// The checks between the screens report a target screen that does not exist, a missing argument
// and a {row.<field>} value that reaches data.
func loadValues(fsys fs.FS, loaders []*loader, langs []string) {
	sets := Sets{}
	byLang := map[string]*loader{}
	for _, l := range loaders {
		byLang[l.lang] = l
		screens := map[string]*screen.Screen{}
		for _, u := range l.units {
			if u.done {
				screens[u.s.Name] = u.s
			}
		}
		sets[l.lang] = screens
	}
	content := func(lang, p string) (*screen.Content, error) { return byLang[lang].content(p) }
	r := Reach(sets, langs, nil)
	errs := LoadRows(fsys, sets, langs, r, content)
	errs = append(errs, LoadFiles(sets, langs, r, content)...)
	for _, e := range errs {
		p := e.Origin.Pos()
		byLang[e.Lang].fd(p.File).Errorf(p.Line, p.Col, "%s: args: %s: %v", e.Origin.Label, e.Origin.Arg, e.Err)
	}
}
