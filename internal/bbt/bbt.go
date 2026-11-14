// Package bbt loads the BBS templates of screens/ into the screen model of internal/screen.
// Specification: docs/specs/screen-format.md, sections 3, 4.2, 6.1 and 10.
package bbt

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"newbbs/internal/art"
	"newbbs/internal/screen"
)

// MaxBodyLines is the line limit of the body of a screen after includes (section 10).
const MaxBodyLines = 10000

// Load loads each screen of fsys/screens with the fragments and the art that it uses.
// fsys is the BBS directory. theme and width come from bbs.toml.
// The map of screens by name is nil when the diagnostics have an error.
func Load(fsys fs.FS, theme map[string]screen.Style, width screen.WidthLimits) (map[string]*screen.Screen, screen.Diags) {
	l := &loader{
		fsys: fsys, theme: theme, width: width,
		diags: map[string]*screen.FileDiags{}, files: map[string]*file{}, arts: map[string]*screen.Art{},
		reported: map[screen.Pos]bool{},
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
		if !isName(base) {
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
	fsys  fs.FS
	theme map[string]screen.Style
	width screen.WidthLimits
	diags map[string]*screen.FileDiags
	files map[string]*file       // nil for a file that cannot be read
	arts  map[string]*screen.Art // nil for art that cannot be loaded
	// reported holds the include and art insert positions with an error:
	// the screens that use a fragment report its errors once.
	reported map[screen.Pos]bool
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
		return f, f == nil
	}
	data, err := screen.ReadFile(l.fsys, name)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		l.files[name] = nil
		return nil, true
	case err != nil:
		l.fd(name).Errorf(1, 1, "%v", err)
		l.files[name] = nil
		return nil, false
	}
	f = parseFile(l.fd(name), data, l.theme, fragment)
	l.files[name] = f
	return f, false
}

// art loads an .ans file once. missing is true when the file does not exist.
func (l *loader) art(name string) (a *screen.Art, missing bool) {
	if a, ok := l.arts[name]; ok {
		return a, a == nil
	}
	if _, err := fs.Stat(l.fsys, name); errors.Is(err, fs.ErrNotExist) {
		l.arts[name] = nil
		return nil, true
	}
	a, ds := art.Load(l.fsys, name)
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
