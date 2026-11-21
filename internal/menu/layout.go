package menu

import (
	"fmt"
	"maps"
	"slices"

	"github.com/schors/velourbbs/internal/bbs"
	"github.com/schors/velourbbs/internal/bbt"
	"github.com/schors/velourbbs/internal/screen"
)

// Layout renders each screen of each screen set of m at each terminal of terms and gives
// the layout warnings (screen-render, section 3.6). sv holds the session and BBS fields of velourbbs preview;
// Layout sets session.lang. The warnings have no limit of screen.MaxFileDiags.
func Layout(m *bbs.Model, terms []screen.Term, sv screen.SessionValues) screen.Diags {
	args := screenArgs(m)
	sd := bbt.NewSetDiags(m.Config.Langs)
	for _, lang := range m.Config.Langs {
		sv.Lang = lang
		set := m.Sets[lang]
		var ds screen.Diags
		for _, name := range slices.Sorted(maps.Keys(set)) {
			s := set[name]
			a, ok := args[name]
			if !ok && len(s.Args) > 0 { // no action shows the screen
				continue
			}
			for _, term := range terms {
				ds = append(ds, screenLayout(m, s, term, a, sv)...)
			}
		}
		sd.Add(lang, ds)
	}
	return sd.All()
}

// screenLayout renders each page of the screen s and gives its layout warnings:
// one for each position and kind, with the values of the first page where it happens.
func screenLayout(m *bbs.Model, s *screen.Screen, term screen.Term, args map[string]string,
	sv screen.SessionValues) screen.Diags {
	log := &screen.LayoutLog{}
	for range pages(m.Config, s, term, args, sv, log) {
	}
	type key struct {
		kind screen.LayoutKind
		pos  screen.Pos
	}
	seen := map[key]bool{}
	var ds screen.Diags
	for _, n := range log.Notes {
		k := key{n.Kind, n.Pos}
		if seen[k] {
			continue
		}
		seen[k] = true
		msg := fmt.Sprintf("at %dx%d: %s", term.Cols, term.Rows, n.Message())
		if n.Pos.File != s.File {
			msg = fmt.Sprintf("screen %q: %s", s.Name, msg)
		}
		ds.Warnf(n.Pos.File, n.Pos.Line, n.Pos.Col, "%s", msg)
	}
	return ds
}

// screenArgs gives the arguments of each screen name that an action shows: the arguments
// of the first action that shows it (screen-render, section 3.6). The order is the screen set
// of [screen] lang, then the other sets in the order of the codes; in each set, the screens
// in name order, the items in source order, then select with the first row of the list.
// A list screen with arguments has a list only after an action shows it.
// A {arg.<name>} value gives the sample value of the argument (bbs.Samples). {in} gives an empty value.
// An action whose {arg.<name>} gives no value does not count.
func screenArgs(m *bbs.Model) map[string]map[string]string {
	out := map[string]map[string]string{}
	samples := bbs.NewSamples(m)
	var from string // the screen of the action
	add := func(a screen.Action) {
		if _, ok := out[a.Screen]; ok {
			return
		}
		args := maps.Clone(a.Args)
		for name, src := range a.ArgRefs {
			v, ok := samples.Arg(from, src)
			if !ok {
				return
			}
			if args == nil {
				args = map[string]string{}
			}
			args[name] = v
		}
		for _, name := range a.In { // {in} gives an empty value
			if args == nil {
				args = map[string]string{}
			}
			args[name] = ""
		}
		out[a.Screen] = args
	}
	langs := []string{m.Config.Lang}
	for _, lang := range m.Config.Langs {
		if lang != m.Config.Lang {
			langs = append(langs, lang)
		}
	}
	for _, lang := range langs {
		set := m.Sets[lang]
		for _, name := range slices.Sorted(maps.Keys(set)) {
			s := set[name]
			from = name
			for _, it := range s.Items {
				if it.Do.Op.HasTarget() {
					add(it.Do)
				}
			}
			if s.List == nil || s.List.Select == nil || !s.List.Select.Do.Op.HasTarget() {
				continue
			}
			a, ok := out[name]
			if !ok && len(s.Args) > 0 {
				continue
			}
			if rows := s.List.List(a); len(rows) > 0 {
				add(s.List.Select.Action(rows[0], 1))
			}
		}
	}
	return out
}
