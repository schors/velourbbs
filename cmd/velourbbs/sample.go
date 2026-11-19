package main

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

// sampleSet is one sample set of sample.toml (screen-render, section 3.5).
type sampleSet struct {
	fields map[string]string
	rows   []map[string]any // each value is a string or an int64; nil without the key rows
}

// loadSample reads the sample set name of sample.toml in fsys. It checks the syntax, the set names
// and the keys of each set, and the types of the values. Each error is an invalid argument.
func loadSample(fsys fs.FS, name string) (*sampleSet, []string) {
	data, err := fs.ReadFile(fsys, sampleFile)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, []string{sampleFile + " does not exist"}
	}
	if err != nil {
		return nil, []string{err.Error()}
	}
	doc, terrs := tomlcheck.Map(data)
	if terrs != nil {
		var msgs []string
		for _, e := range terrs {
			msgs = append(msgs, fmt.Sprintf("%s:%d:%d: %s", sampleFile, e.Line, e.RuneCol(data), e.Msg))
		}
		return nil, msgs
	}
	var msgs []string
	sets := map[string]*sampleSet{}
	for _, n := range slices.Sorted(maps.Keys(doc)) {
		set, errs := sampleTable(doc[n])
		if !isSetName(n) {
			errs = append([]string{"invalid set name"}, errs...)
		}
		for _, e := range errs {
			msgs = append(msgs, fmt.Sprintf("%s: set %q: %s", sampleFile, n, e))
		}
		sets[n] = set
	}
	if msgs != nil {
		return nil, msgs
	}
	set, ok := sets[name]
	if !ok {
		return nil, []string{fmt.Sprintf("%s: unknown set %q", sampleFile, name)}
	}
	return set, nil
}

// paths gives the paths of the fields of the set in sorted order. A nil set has none.
func (set *sampleSet) paths() []string {
	if set == nil {
		return nil
	}
	return slices.Sorted(maps.Keys(set.fields))
}

// isSetName reports whether s matches [a-z][a-z0-9_-]*.
func isSetName(s string) bool {
	for i := range len(s) {
		c := s[i]
		lower := 'a' <= c && c <= 'z'
		other := '0' <= c && c <= '9' || c == '_' || c == '-'
		if !lower && (i == 0 || !other) {
			return false
		}
	}
	return s != ""
}

// sampleTable checks the value v of a set and gives the set.
func sampleTable(v any) (*sampleSet, []string) {
	t, ok := v.(map[string]any)
	if !ok {
		return nil, []string{"must be a table"}
	}
	set := &sampleSet{}
	var errs []string
	for _, key := range slices.Sorted(maps.Keys(t)) {
		switch key {
		case "fields":
			f, ok := t[key].(map[string]any)
			if !ok {
				errs = append(errs, "fields: must be a table of strings")
				continue
			}
			set.fields = map[string]string{}
			for _, path := range slices.Sorted(maps.Keys(f)) {
				s, ok := f[path].(string)
				if !ok {
					errs = append(errs, fmt.Sprintf("fields: %q: must be a string", path))
				}
				set.fields[path] = s
			}
		case "rows":
			rows, ok := t[key].([]any)
			if !ok {
				errs = append(errs, "rows: must be an array of tables")
				continue
			}
			set.rows = []map[string]any{}
			for i, r := range rows {
				row, ok := r.(map[string]any)
				if !ok {
					errs = append(errs, fmt.Sprintf("rows: row %d: must be a table", i+1))
					continue
				}
				for _, f := range slices.Sorted(maps.Keys(row)) {
					switch row[f].(type) {
					case string, int64:
					default:
						errs = append(errs, fmt.Sprintf("rows: row %d: %q: must be a string or an integer", i+1, f))
					}
				}
				set.rows = append(set.rows, row)
			}
		default:
			errs = append(errs, fmt.Sprintf("unknown key %q", key))
		}
	}
	return set, errs
}

// listRows gives the rows of the set for a list screen of the provider from. Each error is an
// error of the preview: a field that the provider does not have, or a value of the wrong type.
// A missing field is empty or 0.
func (set *sampleSet) listRows(from string) ([]screen.ListRow, []string) {
	p, _ := provider.Lookup(from)
	rows := make([]screen.ListRow, 0, len(set.rows))
	var errs []string
	for i, r := range set.rows {
		row := screen.ListRow{}
		for _, f := range p.Fields {
			if screen.RowRef(f).Type() == screen.TypeInt {
				row[f] = 0
			} else {
				row[f] = ""
			}
		}
		for _, f := range slices.Sorted(maps.Keys(r)) {
			if !slices.Contains(p.Fields, f) {
				errs = append(errs, fmt.Sprintf("%s: rows: row %d: provider %q has no field %q", sampleFile, i+1, from, f))
				continue
			}
			switch v := r[f].(type) {
			case int64:
				if screen.RowRef(f).Type() != screen.TypeInt {
					errs = append(errs, fmt.Sprintf("%s: rows: row %d: %q must be a string", sampleFile, i+1, f))
					continue
				}
				row[f] = int(v)
			case string:
				if screen.RowRef(f).Type() != screen.TypeString {
					errs = append(errs, fmt.Sprintf("%s: rows: row %d: %q must be an integer", sampleFile, i+1, f))
					continue
				}
				row[f] = v
			}
		}
		rows = append(rows, row)
	}
	return rows, errs
}
