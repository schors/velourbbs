package bbt

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"slices"

	"github.com/schors/velourbbs/internal/provider"
	"github.com/schors/velourbbs/internal/screen"
	"github.com/schors/velourbbs/internal/tomlcheck"
)

// Origin is the literal, the {row.<field>} or the {in} that gives the values of an argument,
// or a value of velourbbs preview --set (bbt-format, section 8.5).
type Origin struct {
	// File, Screen and Label are the screen file, the screen name and `item "<KEY>"` or `select`.
	// At is the place of the item or of select. A value of --set has none of them.
	File, Screen, Label string
	At                  tomlcheck.Place
	Arg                 string // the argument name in the args of the origin
	Row                 string // the <field> of a {row.<field>} value, else ""
	In                  bool   // a value {in}: an open value
	Text                string // the value of a literal or of --set
}

// Pos gives the position of the value args.<Arg> of the item or of select of o.
func (o Origin) Pos() screen.Pos {
	return filePos(o.File, o.At.At("args", o.Arg).Value())
}

// Reached maps a screen name to the origins of the values of each argument of the screen.
type Reached map[string]map[string][]Origin

// add adds the origin o to the argument arg of the screen name. It reports whether o is new.
func (r Reached) add(name, arg string, o Origin) bool {
	args := r[name]
	if args == nil {
		args = map[string][]Origin{}
		r[name] = args
	}
	if slices.Contains(args[arg], o) {
		return false
	}
	args[arg] = append(args[arg], o)
	return true
}

// Reach gives the origins of the values of each argument of each screen of sets (bbt-format, section 8.5):
// the values of the goto and gosub actions of the items and of select in all sets, also through
// the {arg.<name>} values of other screens. seeds holds more origins by screen name and argument.
// langs is the languages of the BBS.
func Reach(sets Sets, langs []string, seeds Reached) Reached {
	r := Reached{}
	for name, args := range seeds {
		for arg, os := range args {
			for _, o := range os {
				r.add(name, arg, o)
			}
		}
	}
	// link is a value {arg.<from>} of the screen src that gives the argument to of the screen dst.
	type link struct{ src, from, dst, to string }
	var links []link
	for _, lang := range langs {
		for _, name := range slices.Sorted(maps.Keys(sets[lang])) {
			s := sets[lang][name]
			// rows holds the row field of each value {row.<field>} of select.
			add := func(label string, at tomlcheck.Place, a screen.Action, rows map[string]string) {
				if !a.Op.HasTarget() {
					return
				}
				o := Origin{File: s.File, Screen: s.Name, Label: label, At: at}
				for _, arg := range slices.Sorted(maps.Keys(a.Args)) {
					o.Arg, o.Text = arg, a.Args[arg]
					r.add(a.Screen, arg, o)
				}
				o.Text, o.In = "", true
				for _, arg := range a.In {
					o.Arg = arg
					r.add(a.Screen, arg, o)
				}
				o.In = false
				for _, arg := range slices.Sorted(maps.Keys(rows)) {
					o.Arg, o.Row = arg, rows[arg]
					r.add(a.Screen, arg, o)
				}
				for _, arg := range slices.Sorted(maps.Keys(a.ArgRefs)) {
					l := link{s.Name, a.ArgRefs[arg], a.Screen, arg}
					if !slices.Contains(links, l) {
						links = append(links, l)
					}
				}
			}
			for i, it := range s.Items {
				add(itemLabel(it.Key), ItemPlace(s.Front, i), it.Do, nil)
			}
			if s.List != nil && s.List.Select != nil {
				add("select", s.Front.At("select"), s.List.Select.Do, s.List.Select.Refs)
			}
		}
	}
	// A cycle of screens gives no new origins, so the loop ends.
	for changed := true; changed; {
		changed = false
		for _, l := range links {
			for _, o := range r[l.src][l.from] {
				changed = r.add(l.dst, l.to, o) || changed
			}
		}
	}
	return r
}

// itemLabel gives the label of the item with the key k in the errors (menu-nav, section 6).
func itemLabel(k string) string { return fmt.Sprintf("item %q", k) }

// Texts gives the values of the origin o: the text of a literal or of --set, or the field of each row
// of each list of the select of o in the sets of langs (list-view, section 8).
// An open value gives no text: the checks between the screens report it at file and data.
func Texts(sets Sets, langs []string, o Origin) []string {
	if o.In {
		return nil
	}
	if o.Row == "" {
		return []string{o.Text}
	}
	var out []string
	for _, lang := range langs {
		s := sets[lang][o.Screen]
		if s == nil || s.File != o.File || s.List == nil {
			continue
		}
		for _, key := range slices.Sorted(maps.Keys(s.List.Lists)) {
			for i, row := range s.List.Lists[key] {
				if v := screen.RowValue(row, i+1, o.Row); !slices.Contains(out, v) {
					out = append(out, v)
				}
			}
		}
	}
	return out
}

// ValueError is an error of a value at its origin (bbt-format, section 8.5).
type ValueError struct {
	Lang   string // the screen set that finds the error
	Origin Origin
	Err    error
}

// ContentFunc loads the content file p in the screen set lang (file-view, section 2).
// It reports the errors in the file. The error is an error of the whole file.
// The content is nil after an error.
type ContentFunc func(lang, p string) (*screen.Content, error)

// valueErrors collects the errors of the values. An equal error appears once.
type valueErrors []ValueError

func (ve *valueErrors) add(lang string, os []Origin, err error) {
	for _, o := range os {
		e := ValueError{lang, o, err}
		if !slices.ContainsFunc(*ve, func(x ValueError) bool {
			return x.Lang == e.Lang && x.Origin == e.Origin && x.Err.Error() == e.Err.Error()
		}) {
			*ve = append(*ve, e)
		}
	}
}

// values gives the values of the argument arg of the screen name in r, each with its origins,
// in the order of the origins. A {row.<field>} value cannot give rows: the checks between
// the screens report it, so values skips it when rows is true.
func values(sets Sets, langs []string, r Reached, name, arg string, rows bool) ([]string, map[string][]Origin) {
	var texts []string
	by := map[string][]Origin{}
	for _, o := range r[name][arg] {
		if rows && o.Row != "" {
			continue
		}
		for _, t := range Texts(sets, langs, o) {
			if by[t] == nil {
				texts = append(texts, t)
			}
			by[t] = append(by[t], o)
		}
	}
	return texts, by
}

// LoadRows reads the rows of each list screen of sets with {arg.<name>} in data for each value
// of its arguments in r, in the set of each screen (list-view, section 4). A data with more than
// one {arg.<name>} gives each combination of the values. It puts the rows into the screen
// and gives the errors at the origins of the values. fsys is the BBS directory.
func LoadRows(fsys fs.FS, sets Sets, langs []string, r Reached, content ContentFunc) []ValueError {
	var ve valueErrors
	for _, lang := range langs {
		for _, name := range slices.Sorted(maps.Keys(sets[lang])) {
			ls := sets[lang][name].List
			if ls == nil || len(ls.ArgRefs) == 0 || ls.From == "" {
				continue
			}
			p, _ := provider.Lookup(ls.From)
			var args []string // the screen arguments of data
			for _, a := range ls.ArgRefs {
				if !slices.Contains(args, a) {
					args = append(args, a)
				}
			}
			slices.Sort(args)
			texts := make([][]string, len(args))
			by := make([]map[string][]Origin, len(args))
			for i, a := range args {
				texts[i], by[i] = values(sets, langs, r, name, a, true)
			}
			for combo := range combinations(texts) {
				vals := map[string]string{}
				var all []Origin
				for i, a := range args {
					vals[a] = combo[i]
					all = append(all, by[i][combo[i]]...)
				}
				pv := ls.Values(vals)
				key := screen.ListKey(pv)
				if _, ok := ls.Lists[key]; ok {
					continue
				}
				rows, errs := p.Rows(fsys, lang, pv, func(path string) (*screen.Content, error) { return content(lang, path) })
				if len(errs) == 0 {
					ls.Lists[key] = rows
				}
				for _, err := range errs {
					var ke screen.KeyError
					if errors.As(err, &ke) && ls.ArgRefs[ke.Key] != "" {
						a := ls.ArgRefs[ke.Key]
						ve.add(lang, by[slices.Index(args, a)][vals[a]], ke.Err)
						continue
					}
					ve.add(lang, all, err) // an error of a literal value of data
				}
			}
		}
	}
	return ve
}

// combinations gives each combination of one text of each list of texts, in order.
func combinations(texts [][]string) func(yield func([]string) bool) {
	return func(yield func([]string) bool) {
		combo := make([]string, len(texts))
		var walk func(i int) bool
		walk = func(i int) bool {
			if i == len(texts) {
				return yield(slices.Clone(combo))
			}
			for _, t := range texts[i] {
				combo[i] = t
				if !walk(i + 1) {
					return false
				}
			}
			return true
		}
		walk(0)
	}
}

// LoadFiles loads the content file of each value in r of the argument of each file screen of sets
// with {arg.<name>}, in the set of each screen, with the checks of file-view, section 7.
// It puts the files into the screen and gives the errors at the origins of the values.
func LoadFiles(sets Sets, langs []string, r Reached, content ContentFunc) []ValueError {
	var ve valueErrors
	for _, lang := range langs {
		for _, name := range slices.Sorted(maps.Keys(sets[lang])) {
			f := sets[lang][name].FileScreen
			if f == nil || f.Arg == "" {
				continue
			}
			texts, by := values(sets, langs, r, name, f.Arg, false)
			for _, p := range texts {
				if f.Contents[p] != nil {
					continue
				}
				c, err := content(lang, p)
				var errs []error
				if err != nil {
					errs = []error{err}
				} else if c != nil {
					errs = f.Check(p, c)
				}
				for _, err := range errs {
					ve.add(lang, by[p], err)
				}
				if c != nil && len(errs) == 0 {
					f.Contents[p] = c
				}
			}
		}
	}
	return ve
}
