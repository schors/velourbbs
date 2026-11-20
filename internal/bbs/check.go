package bbs

import (
	"fmt"
	"maps"
	"slices"

	"github.com/schors/velourbbs/internal/bbt"
	"github.com/schors/velourbbs/internal/config"
	"github.com/schors/velourbbs/internal/screen"
	"github.com/schors/velourbbs/internal/tomlcheck"
)

// Check does the checks of menu-nav, section 6, between files in the screen set of each language:
// the goto and gosub actions and their args, the values of the args, the screens without items,
// the start screen and the arguments of goodbye and shutdown. It also checks that each set has the screen names
// and the arguments of each other set (session-language, section 6). Without an error, it warns
// about each screen that a user cannot reach (menu-nav, section 6.1).
// sets is the result of a load without errors. Check merges the diagnostics of the sets with ds
// (bbt.SetDiags) and gives the sorted diagnostics.
func Check(cfg config.Config, sets bbt.Sets, ds screen.Diags) screen.Diags {
	sd := bbt.NewSetDiags(cfg.Langs)
	sd.Add("", ds)
	for _, lang := range cfg.Langs {
		var ls screen.Diags
		screens := sets[lang]
		for _, name := range slices.Sorted(maps.Keys(screens)) {
			s := screens[name]
			end := name == "goodbye" || name == "shutdown"
			if len(s.Items) == 0 && !end {
				ls.Errorf(s.File, 1, 1, "screen has no items")
			}
			if end && len(s.Args) > 0 {
				ls.Errorf(s.File, 1, 1, "the %s screen cannot have arguments", name)
			}
			// An error between an item and another screen has the position of its key path
			// (screen-render, section 3.1).
			errorf := func(at tomlcheck.Place, format string, args ...any) {
				p := at.Value()
				ls.Errorf(s.File, p.Line, p.Col, format, args...)
			}
			for i, it := range s.Items {
				if !it.Do.Op.HasTarget() {
					continue
				}
				at := bbt.ItemPlace(s.Front, i)
				if err := CheckTarget(sets, lang, it.Do); err != nil {
					errorf(at.At("do"), "item %q: do: %v", it.Key, err)
					continue
				}
				for _, err := range CheckArgs(Target(sets, lang, it.Do), it.Do.Target(), it.Do.Names()) {
					errorf(at.At("args"), "item %q: args: %v", it.Key, err)
				}
			}
			if s.List != nil && s.List.Select != nil {
				checkSelect(sets, lang, s.List.Select, s.Front.At("select"), errorf)
			}
		}
		start := screen.Action{Screen: cfg.Start}
		if err := CheckTarget(sets, lang, start); err != nil {
			ls.Errorf(config.ScreensFile, cfg.StartPos.Line, cfg.StartPos.Col, "menu: start: %v", err)
		} else if len(Target(sets, lang, start).Args) > 0 {
			ls.Errorf(config.ScreensFile, cfg.StartPos.Line, cfg.StartPos.Col, "menu: start: screen %q has arguments", cfg.Start)
		}
		sd.Add(lang, ls)
	}
	checkSets(cfg.Langs, sets, sd)
	if !sd.All().HasErrors() {
		sd.Add("", checkReach(cfg.Langs, sets))
	}
	if !sd.All().HasErrors() {
		unreachable(cfg, sets, sd)
	}
	return sd.Diags()
}

// checkSelect checks the goto or the gosub of select in a screen of the set lang: its target
// and its arguments (list-view, section 8). at is the place of select.
func checkSelect(sets bbt.Sets, lang string, sel *screen.Select, at tomlcheck.Place,
	errorf func(at tomlcheck.Place, format string, args ...any)) {
	a := sel.Action(nil, 1) // each argument with a value
	if !a.Op.HasTarget() {
		return
	}
	if err := CheckTarget(sets, lang, a); err != nil {
		errorf(at.At("do"), "select: do: %v", err)
		return
	}
	for _, err := range CheckArgs(Target(sets, lang, a), a.Target(), a.Names()) {
		errorf(at.At("args"), "select: args: %v", err)
	}
}

// checkReach reports each {row.<field>} value that reaches the data of a screen, also through
// the {arg.<name>} values of other screens (bbt-format, section 8.5). It needs the screens
// of all sets of langs and their checks without errors.
func checkReach(langs []string, sets bbt.Sets) screen.Diags {
	r := bbt.Reach(sets, langs, nil)
	var ds screen.Diags
	type key struct {
		o    bbt.Origin
		name string
	}
	seen := map[key]bool{}
	for _, lang := range langs {
		for _, name := range slices.Sorted(maps.Keys(sets[lang])) {
			ls := sets[lang][name].List
			if ls == nil {
				continue
			}
			for _, pa := range slices.Sorted(maps.Keys(ls.ArgRefs)) {
				for _, o := range r[name][ls.ArgRefs[pa]] {
					if o.Row == "" || seen[key{o, name}] {
						continue
					}
					seen[key{o, name}] = true
					p := o.Pos()
					ds.Errorf(p.File, p.Line, p.Col, "%s: args: %s: the value reaches the data of screen %q", o.Label, o.Arg, name)
				}
			}
		}
	}
	return ds
}

// checkSets reports each screen name of a set that is not in the set of each other language,
// or that has other arguments there. The reference set of a name is the first set that holds it.
// Each other set finds the error at 1:1 of the screen file in the reference set (session-language, section 6).
func checkSets(langs []string, sets bbt.Sets, sd *bbt.SetDiags) {
	names := map[string]bool{}
	for _, lang := range langs {
		for name := range sets[lang] {
			names[name] = true
		}
	}
	for _, name := range slices.Sorted(maps.Keys(names)) {
		i := slices.IndexFunc(langs, func(lang string) bool { return sets[lang][name] != nil })
		ref := sets[langs[i]][name]
		for _, lang := range langs {
			s, ok := sets[lang][name]
			var ds screen.Diags
			switch {
			case lang == langs[i]:
				continue
			case !ok:
				ds.Errorf(ref.File, 1, 1, "no screen %q", name)
			case !slices.Equal(ref.Args, s.Args):
				ds.Errorf(ref.File, 1, 1, "arguments differ")
			}
			sd.Add(lang, ds)
		}
	}
}

// CheckArgs reports each argument of s that the sorted names do not give
// and each name that is not an argument of s (bbt-format, section 8.5).
// The errors name s as target: "<name>" or "<code>/<name>" (session-language, section 6).
func CheckArgs(s *screen.Screen, target string, names []string) []error {
	var errs []error
	for _, name := range s.Args {
		if !slices.Contains(names, name) {
			errs = append(errs, fmt.Errorf("missing argument %q of screen %q", name, target))
		}
	}
	for _, name := range names {
		if !slices.Contains(s.Args, name) {
			errs = append(errs, fmt.Errorf("screen %q has no argument %q", target, name))
		}
	}
	return errs
}

// Target gives the screen of the goto or the gosub a in a screen of the set lang, else nil:
// a screen of the set of the code of a, else of the set lang (session-language, section 6).
func Target(sets bbt.Sets, lang string, a screen.Action) *screen.Screen {
	if a.Lang != "" {
		lang = a.Lang
	}
	return sets[lang][a.Screen]
}

// CheckTarget checks the target of a goto, of a gosub or of the start a in a screen of the set lang:
// it must exist and have items. The error names the target as the do names it.
func CheckTarget(sets bbt.Sets, lang string, a screen.Action) error {
	s := Target(sets, lang, a)
	switch {
	case s == nil:
		return fmt.Errorf("unknown screen %q", a.Target())
	case len(s.Items) == 0:
		return fmt.Errorf("screen %q has no items", a.Target())
	}
	return nil
}
