package bbt

import (
	"fmt"
	"slices"
	"strings"

	"github.com/schors/velourbbs/internal/screen"
)

// SetDiags collects the diagnostics that each screen set finds and merges them
// (session-language, section 6).
type SetDiags struct {
	langs []string
	all   screen.Diags            // the diagnostics of all sets
	sets  map[string]screen.Diags // the diagnostics of each set by its language
}

// NewSetDiags gives a SetDiags for the languages of the BBS langs, sorted.
func NewSetDiags(langs []string) *SetDiags {
	return &SetDiags{langs: langs, sets: map[string]screen.Diags{}}
}

// Add adds the diagnostics ds that the set of lang finds. lang "" adds diagnostics of all sets.
func (sd *SetDiags) Add(lang string, ds screen.Diags) {
	if lang == "" {
		sd.all = append(sd.all, ds...)
		return
	}
	sd.sets[lang] = append(sd.sets[lang], ds...)
}

// Diags gives the sorted diagnostics with the limit of screen.MaxFileDiags for each file.
// An equal diagnostic appears once. When each set of its file finds it, it has no suffix.
// Else it appears once for each set that finds it, with the suffix ` (language "<code>")`.
// The sets of a file in screens/<code>/ or content/<code>/ are the set of <code>;
// the sets of each other file are all sets. With one language, there is no suffix.
func (sd *SetDiags) Diags() screen.Diags {
	files := map[string]*screen.FileDiags{}
	for _, d := range sd.merge() {
		fd, ok := files[d.File]
		if !ok {
			fd = &screen.FileDiags{File: d.File}
			files[d.File] = fd
		}
		if d.Warning {
			fd.Warnf(d.Row, d.Col, "%s", d.Msg)
		} else {
			fd.Errorf(d.Row, d.Col, "%s", d.Msg)
		}
	}
	var out screen.Diags
	for _, fd := range files {
		out = append(out, fd.Diags...)
	}
	out.Sort()
	return out
}

// All gives the diagnostics as Diags does, without the limit.
func (sd *SetDiags) All() screen.Diags {
	out := sd.merge()
	out.Sort()
	return out
}

// merge gives the diagnostics of all sets, then the merged diagnostics of each set, in the order they came.
func (sd *SetDiags) merge() screen.Diags {
	var order []screen.Diag
	found := map[screen.Diag][]string{}
	for _, lang := range sd.langs {
		for _, d := range sd.sets[lang] {
			if _, ok := found[d]; !ok {
				order = append(order, d)
			}
			if !slices.Contains(found[d], lang) {
				found[d] = append(found[d], lang)
			}
		}
	}
	out := slices.Clone(sd.all)
	for _, d := range order {
		in := found[d]
		own := sd.fileLangs(d.File)
		if len(sd.langs) == 1 || !slices.ContainsFunc(own, func(l string) bool { return !slices.Contains(in, l) }) {
			out = append(out, d)
			continue
		}
		for _, lang := range in {
			s := d
			s.Msg = fmt.Sprintf("%s (language %q)", d.Msg, lang)
			out = append(out, s)
		}
	}
	return out
}

// fileLangs gives the sets of the file name: the language of a language file, else all languages
// (session-language, section 2).
func (sd *SetDiags) fileLangs(name string) []string {
	for _, dir := range []string{"screens/", "content/"} {
		if rest, ok := strings.CutPrefix(name, dir); ok {
			code, _, _ := strings.Cut(rest, "/")
			if slices.Contains(sd.langs, code) {
				return []string{code}
			}
		}
	}
	return sd.langs
}
