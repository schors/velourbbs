package bbs

import (
	"maps"
	"slices"

	"github.com/schors/velourbbs/internal/bbt"
	"github.com/schors/velourbbs/internal/config"
	"github.com/schors/velourbbs/internal/screen"
)

// pair is a language and a screen name: a node of the walk (menu-nav, section 6.1).
type pair struct{ lang, name string }

// unreachable adds the warning of each screen that a user cannot reach to sd: the walk
// of menu-nav, section 6.1, from the start screen in cfg.Lang. sets is the result
// of the checks of section 6 and of session-language, section 6, without errors.
func unreachable(cfg config.Config, sets bbt.Sets, sd *bbt.SetDiags) {
	next := map[pair][]pair{} // the pairs that the walk reaches from a pair
	type call struct{ from, to pair }
	var calls []call // each gosub
	for _, lang := range cfg.Langs {
		for name, s := range sets[lang] {
			from := pair{lang, name}
			for _, a := range actions(s) {
				switch a.Op {
				case screen.Goto, screen.Gosub:
					to := pair{lang, a.Screen}
					if a.Lang != "" {
						to.lang = a.Lang
					}
					next[from] = append(next[from], to)
					if a.Op == screen.Gosub {
						calls = append(calls, call{from, to})
					}
				case screen.Lang:
					next[from] = append(next[from], pair{a.Lang, name})
				}
			}
		}
	}
	// A gosub reaches its screen in each language that the walk from its target reaches.
	// The walk from a target uses this rule too, so repeat it until it adds no pair.
	for changed := true; changed; {
		changed = false
		for _, c := range calls {
			for p := range walk(next, c.to) {
				back := pair{p.lang, c.from.name}
				if !slices.Contains(next[c.from], back) {
					next[c.from] = append(next[c.from], back)
					changed = true
				}
			}
		}
	}
	reached := walk(next, pair{cfg.Lang, cfg.Start})
	for _, lang := range cfg.Langs {
		var ds screen.Diags
		for _, name := range slices.Sorted(maps.Keys(sets[lang])) {
			if name != "goodbye" && name != "shutdown" && !reached[pair{lang, name}] {
				ds.Warnf(sets[lang][name].File, 1, 1, "a user cannot reach this screen")
			}
		}
		sd.Add(lang, ds)
	}
}

// actions gives the action of each item of s and the action of its select.
func actions(s *screen.Screen) []screen.Action {
	var as []screen.Action
	for _, it := range s.Items {
		as = append(as, it.Do)
	}
	if s.List != nil && s.List.Select != nil {
		as = append(as, s.List.Select.Do)
	}
	return as
}

// walk gives the pairs that next reaches from start, with start.
func walk(next map[pair][]pair, start pair) map[pair]bool {
	seen := map[pair]bool{start: true}
	queue := []pair{start}
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		for _, q := range next[p] {
			if !seen[q] {
				seen[q] = true
				queue = append(queue, q)
			}
		}
	}
	return seen
}
