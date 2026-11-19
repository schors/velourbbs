package bbs

import (
	"fmt"
	"maps"
	"slices"

	"github.com/schors/velourbbs/internal/config"
	"github.com/schors/velourbbs/internal/screen"
)

// Check does the checks of menu-nav, section 6, between files: the goto and gosub actions and their args,
// the screens without items, the start screen and the arguments of goodbye and shutdown. screens is the result of a load without errors.
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
	for _, name := range slices.Sorted(maps.Keys(screens)) {
		s := screens[name]
		end := name == "goodbye" || name == "shutdown"
		if len(s.Items) == 0 && !end {
			fd(s.File).Errorf(1, 1, "screen has no items")
		}
		if end && len(s.Args) > 0 {
			fd(s.File).Errorf(1, 1, "the %s screen cannot have arguments", name)
		}
		for _, it := range s.Items {
			if !it.Do.Op.HasTarget() {
				continue
			}
			// An error between an item and another screen has the position 2:1.
			if err := CheckTarget(screens, it.Do.Screen); err != nil {
				fd(s.File).Errorf(2, 1, "item %q: do: %v", it.Key, err)
				continue
			}
			for _, err := range CheckArgs(screens[it.Do.Screen], it.Do.Args) {
				fd(s.File).Errorf(2, 1, "item %q: args: %v", it.Key, err)
			}
		}
	}
	if err := CheckTarget(screens, cfg.Start); err != nil {
		fd(config.ScreensFile).Errorf(1, 1, "menu: start: %v", err)
	} else if len(screens[cfg.Start].Args) > 0 {
		fd(config.ScreensFile).Errorf(1, 1, "menu: start: screen %q has arguments", cfg.Start)
	}
	var out screen.Diags
	for _, d := range files {
		out = append(out, d.Diags...)
	}
	out.Sort()
	return out
}

// CheckArgs reports each argument of s that args does not give
// and each name of args that is not an argument of s (bbt-format, section 8.5).
func CheckArgs(s *screen.Screen, args map[string]string) []error {
	var errs []error
	for _, name := range s.Args {
		if _, ok := args[name]; !ok {
			errs = append(errs, fmt.Errorf("missing argument %q of screen %q", name, s.Name))
		}
	}
	for _, name := range slices.Sorted(maps.Keys(args)) {
		if !slices.Contains(s.Args, name) {
			errs = append(errs, fmt.Errorf("screen %q has no argument %q", s.Name, name))
		}
	}
	return errs
}

// CheckTarget checks the screen of a goto, of a gosub or of the start: it must exist and have items.
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
