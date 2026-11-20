package bbs

import (
	"maps"
	"slices"

	"github.com/schors/velourbbs/internal/screen"
)

// Samples gives the sample values of the arguments of the screens (screen-render, section 3.4).
// The screen set of [screen] lang gives the values.
type Samples struct {
	actions []sampleAction // the goto and gosub actions in the order of the values
	// sources holds each argument with a literal value: the walk needs no other argument.
	sources map[argRef]bool
	// links holds the arguments whose {arg.<name>} values give each argument.
	links map[argRef][]argRef
}

// argRef is an argument of a screen.
type argRef struct{ screen, arg string }

// sampleAction is a goto or a gosub of the screen from. For select, Args has the values of the first row.
type sampleAction struct {
	from string
	a    screen.Action
}

// NewSamples gives the sample values of the model m: the screens of the set of [screen] lang
// in name order, the items in source order, then select with the first row of the screen:
// the first row of its first list with rows, in the order of the list keys.
func NewSamples(m *Model) *Samples {
	set := m.Sets[m.Config.Lang]
	s := &Samples{sources: map[argRef]bool{}, links: map[argRef][]argRef{}}
	add := func(from string, a screen.Action) {
		if !a.Op.HasTarget() {
			return
		}
		s.actions = append(s.actions, sampleAction{from, a})
		for name := range a.Args {
			s.sources[argRef{a.Screen, name}] = true
		}
		for name, src := range a.ArgRefs {
			to := argRef{a.Screen, name}
			s.links[to] = append(s.links[to], argRef{from, src})
		}
	}
	for _, name := range slices.Sorted(maps.Keys(set)) {
		sc := set[name]
		for _, it := range sc.Items {
			add(name, it.Do)
		}
		if sc.List == nil || sc.List.Select == nil {
			continue
		}
		for _, key := range slices.Sorted(maps.Keys(sc.List.Lists)) {
			if rows := sc.List.Lists[key]; len(rows) > 0 {
				add(name, sc.List.Select.Action(rows[0], 1))
				break
			}
		}
	}
	return s
}

// Arg gives the sample value of the argument arg of the screen name: the value of the first action
// that gives it. A {arg.<name>} value gives the sample value of the argument of its own screen
// by the same rule. An argument that is already on the path gives no value: the action does not count.
// ok is false when no action counts.
func (s *Samples) Arg(name, arg string) (v string, ok bool) {
	return s.arg(argRef{name, arg}, []argRef{{name, arg}})
}

// Fields gives the sample value of each field arg.<name>: the first value that an action gives
// to an argument <name> of a screen. An argument without a value is absent.
func (s *Samples) Fields() map[string]string {
	out := map[string]string{}
	for _, sa := range s.actions {
		for _, name := range sa.a.Names() {
			if _, ok := out[name]; ok {
				continue
			}
			to := argRef{sa.a.Screen, name}
			if v, ok := s.value(sa, name, []argRef{to}); ok {
				out[name] = v
			}
		}
	}
	return out
}

// arg gives the sample value of the argument r. path is the arguments of the walk, with r.
func (s *Samples) arg(r argRef, path []argRef) (string, bool) {
	for _, sa := range s.actions {
		if sa.a.Screen != r.screen {
			continue
		}
		if v, ok := s.value(sa, r.arg, path); ok {
			return v, true
		}
	}
	return "", false
}

// value gives the value of the argument name of the action sa. path is the arguments of the walk.
func (s *Samples) value(sa sampleAction, name string, path []argRef) (string, bool) {
	if v, ok := sa.a.Args[name]; ok {
		return v, true
	}
	src, ok := sa.a.ArgRefs[name]
	if !ok {
		return "", false
	}
	from := argRef{sa.from, src}
	// The check of a value off the path first: the walk does not try each path of a cycle.
	if slices.Contains(path, from) || !s.hasValue(from, path) {
		return "", false
	}
	return s.arg(from, append(slices.Clone(path), from))
}

// hasValue reports whether a walk from the argument r that does not pass the arguments of path
// reaches an argument with a literal value.
func (s *Samples) hasValue(r argRef, path []argRef) bool {
	seen := map[argRef]bool{r: true}
	for _, p := range path {
		seen[p] = true
	}
	queue := []argRef{r}
	for len(queue) > 0 {
		q := queue[0]
		queue = queue[1:]
		if s.sources[q] {
			return true
		}
		for _, n := range s.links[q] {
			if !seen[n] {
				seen[n] = true
				queue = append(queue, n)
			}
		}
	}
	return false
}
