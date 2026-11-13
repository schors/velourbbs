// Package menu checks the menus between screens and runs the session.
// Specification: docs/specs/menu-nav.md, sections 6 and 7.
package menu

import (
	"fmt"
	"maps"
	"slices"

	"newbbs/internal/config"
	"newbbs/internal/screen"
)

// Check does the checks of section 7 between files: the goto actions, the screens
// without items and the start screen. screens is the result of a load without errors.
// Check adds the errors to ds with the limit of screen.MaxFileDiags for each file
// and gives the sorted diagnostics.
func Check(cfg config.Config, screens map[string]*screen.Screen, ds screen.Diags) screen.Diags {
	files := map[string]*screen.FileDiags{}
	fd := func(file string) *screen.FileDiags {
		d, ok := files[file]
		if !ok {
			d = &screen.FileDiags{File: file}
			files[file] = d
		}
		return d
	}
	for _, d := range ds {
		fd(d.File).Diags = append(fd(d.File).Diags, d)
	}
	narrow := map[string]bool{}
	for _, s := range screens {
		if s.Narrow != nil {
			narrow[s.Narrow.Name] = true
		}
	}
	for _, name := range slices.Sorted(maps.Keys(screens)) {
		s := screens[name]
		if len(s.Items) == 0 && name != "goodbye" && !narrow[name] {
			fd(s.File).Errorf(1, 1, "screen has no items")
		}
		for _, it := range s.Items {
			if it.Do.Op != screen.Goto {
				continue
			}
			if err := CheckTarget(screens, it.Do.Screen); err != nil {
				fd(it.DoPos.File).Errorf(it.DoPos.Line, it.DoPos.Col, "item %q: do: %v", it.Key, err)
			}
		}
	}
	if err := CheckTarget(screens, cfg.Start); err != nil {
		p := cfg.StartPos
		fd(p.File).Errorf(p.Line, p.Col, "menu: start: %v", err)
	}
	var out screen.Diags
	for _, d := range files {
		out = append(out, d.Diags...)
	}
	out.Sort()
	return out
}

// CheckTarget checks the screen of a goto or of the start: it must exist and have items.
func CheckTarget(screens map[string]*screen.Screen, name string) error {
	s, ok := screens[name]
	switch {
	case !ok:
		return fmt.Errorf("unknown screen %q", name)
	case len(s.Items) == 0:
		return fmt.Errorf("screen %q has no items", name)
	}
	return nil
}
