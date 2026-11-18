// Package bbt loads the BBS templates of screens/ into the screen model of internal/screen.
// The file pass parses each file once and resolves its fields with the aliases of screens/screens.toml.
// The screen pass expands the includes and the art inserts of each screen.
// Specification: docs/specs/bbt-format.md.
package bbt

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/schors/velourbbs/internal/art"
	"github.com/schors/velourbbs/internal/content"
	"github.com/schors/velourbbs/internal/screen"
)

// MaxBodyLines is the line limit of the body of a screen after includes (section 9).
const MaxBodyLines = 10000

// Load loads each screen of fsys/screens with the fragments and the art that it uses.
// fsys is the BBS directory. theme (the base theme), layout and fields come from
// screens/screens.toml. The load uses only the names of theme (bbt-format, section 7.3).
// The map of screens by name is nil when the diagnostics have an error.
func Load(fsys fs.FS, theme map[string]screen.Style, layout screen.Layout, fields screen.Fields) (map[string]*screen.Screen, screen.Diags) {
	l := &loader{
		fsys: fsys, theme: theme, layout: layout, fields: fields,
		diags: map[string]*screen.FileDiags{}, files: map[string]*file{}, arts: map[string]*screen.Art{},
		missing: map[string]bool{}, reported: map[screen.Pos]bool{}, contents: map[string]loaded{},
	}
	entries, err := fs.ReadDir(fsys, "screens")
	if err != nil {
		l.fd("screens").Errorf(1, 1, "%v", err)
	}
	var units []*unit
	for _, e := range entries {
		base, ok := strings.CutSuffix(e.Name(), ".bbt")
		if !ok || e.IsDir() || strings.HasPrefix(base, "_") {
			continue
		}
		name := path.Join("screens", e.Name())
		if !screen.IsName(base) {
			l.fd(name).Errorf(1, 1, "invalid screen name %q", base)
			continue
		}
		if f, _ := l.file(name, false); f != nil {
			units = append(units, &unit{s: &screen.Screen{Name: base, File: name}, f: f})
		}
	}
	for _, u := range units {
		l.screen(u)
	}
	l.targetFiles(units)
	l.checkItems(units)
	for _, e := range entries {
		name := path.Join("screens", e.Name())
		var used bool
		switch {
		case strings.HasPrefix(e.Name(), "_") && strings.HasSuffix(e.Name(), ".bbt"):
			_, used = l.files[name]
		case strings.HasSuffix(e.Name(), ".ans"):
			_, used = l.arts[name]
		case strings.HasSuffix(e.Name(), ".bbt"):
			used = !e.IsDir() // the screen pass reads each screen file
		default:
			continue
		}
		if used {
			continue // its load reports a file that is not a regular file (section 2)
		}
		if info, err := fs.Stat(fsys, name); err == nil && !info.Mode().IsRegular() {
			l.fd(name).Errorf(1, 1, "not a regular file")
		} else {
			l.fd(name).Warnf(1, 1, "no screen uses this file")
		}
	}

	var ds screen.Diags
	for _, d := range l.diags {
		ds = append(ds, d.Diags...)
	}
	ds.Sort()
	if ds.HasErrors() {
		return nil, ds
	}
	screens := make(map[string]*screen.Screen, len(units))
	for _, u := range units {
		screens[u.s.Name] = u.s
	}
	return screens, ds
}

type loader struct {
	fsys   fs.FS
	theme  map[string]screen.Style
	layout screen.Layout // the layout of [screen]
	fields screen.Fields // the aliases (section 4.2)
	diags  map[string]*screen.FileDiags
	files  map[string]*file       // nil for a file that cannot be read
	arts   map[string]*screen.Art // nil for art that cannot be loaded
	// missing holds the .bbt and .ans files that do not exist.
	missing map[string]bool
	// reported holds the include and art insert positions with an error:
	// the screens that use a fragment report its errors once.
	reported map[screen.Pos]bool
	contents map[string]loaded // the content files by path
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
	f = parseFile(l.fd(name), data, l.theme, l.fields, fragment)
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
	a, ds := art.Load(l.fsys, name, l.fields)
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
	c, ds, err := content.Load(l.fsys, p)
	for _, d := range ds {
		l.fd(d.File).Diags = append(l.fd(d.File).Diags, d)
	}
	l.contents[p] = loaded{c, err}
	return c, err
}

// literalFile loads the content file of the key file of the screen u (file-view, section 7).
func (l *loader) literalFile(u *unit, f *screen.FileSpec) {
	pos := screen.Pos{File: u.f.name, Line: 2, Col: 1}
	c, err := l.content(f.Path)
	if err != nil {
		l.errorf(u, pos, false, "file: %v", err)
		return
	}
	if c == nil { // the errors are in the content file
		return
	}
	errs := f.Check(f.Path, c)
	for _, err := range errs {
		l.errorf(u, pos, false, "%v", err)
	}
	if len(errs) == 0 {
		f.Contents[f.Path] = c
	}
}

// targetFiles loads the content file of each goto and gosub to a file screen with {arg.<name>}.
// An error is an error of the item with the goto or the gosub (file-view, section 7).
// The menu checks report a target screen that does not exist and a missing argument.
func (l *loader) targetFiles(units []*unit) {
	byName := map[string]*unit{}
	for _, u := range units {
		byName[u.s.Name] = u
	}
	for _, u := range units {
		if !u.done {
			continue
		}
		for _, it := range u.s.Items {
			t := byName[it.Do.Screen]
			if !it.Do.Op.HasTarget() || t == nil || !t.done || t.s.FileScreen == nil || t.s.FileScreen.Arg == "" {
				continue
			}
			f := t.s.FileScreen
			p, ok := it.Do.Args[f.Arg]
			if !ok {
				continue
			}
			c, err := l.content(p)
			var errs []error
			if err != nil {
				errs = []error{err}
			} else if c != nil {
				errs = f.Check(p, c)
			}
			for _, err := range errs {
				l.fd(u.f.name).Errorf(2, 1, "item %q: args: %s: %v", it.Key, f.Arg, err)
			}
			if c != nil && len(errs) == 0 {
				f.Contents[p] = c
			}
		}
	}
}
