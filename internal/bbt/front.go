package bbt

import (
	"cmp"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"unicode/utf8"

	"newbbs/internal/screen"
	"newbbs/internal/tomlpos"
)

// front is the front matter of a screen (section 3.2).
type front struct {
	width  *screen.WidthLimits // nil when width is not set
	narrow string
	fields screen.Fields
	items  []screen.Item
	doc    *tomlpos.Doc
	lines  []string
	file   string
}

// parseFront parses the front matter lines. The first line of lines is line 2 of the file.
// The result is nil when the front matter has an error.
func parseFront(d *screen.FileDiags, lines []string) *front {
	fr := &front{lines: lines, file: d.File, fields: screen.Fields{}}
	doc, err := tomlpos.Parse([]byte(strings.Join(lines, "\n")))
	if err != nil {
		var e *tomlpos.Error
		errors.As(err, &e)
		p := fr.convert(e.Pos)
		d.Errorf(p.Line, p.Col, "%s", e.Msg)
		return nil
	}
	fr.doc = doc
	ok := true
	errorf := func(path []string, format string, args ...any) {
		p := fr.pos(path...)
		d.Errorf(p.Line, p.Col, "%s", fmt.Sprintf(format, args...))
		ok = false
	}
	if v, has := doc.Values["fields"]; has {
		fr.parseFields(v, errorf)
	}
	for _, key := range fr.keys(doc.Values) {
		v := doc.Values[key]
		switch key {
		case "fields":
		case "width":
			w, err := screen.ParseWidth(v)
			if err != nil {
				errorf([]string{key}, "width: %v", err)
				continue
			}
			fr.width = &w
		case "narrow":
			s, isStr := v.(string)
			switch {
			case !isStr:
				errorf([]string{key}, "narrow: must be a string")
			case !isName(s):
				errorf([]string{key}, "narrow: invalid screen name %q", s)
			default:
				fr.narrow = s
			}
		case "item":
			fr.parseItems(v, errorf)
		default:
			errorf([]string{key}, "unknown key %q", key)
		}
	}
	if !ok {
		return nil
	}
	return fr
}

type errorFunc func(path []string, format string, args ...any)

func (fr *front) parseFields(v any, errorf errorFunc) {
	t, ok := v.(map[string]any)
	if !ok {
		errorf([]string{"fields"}, "fields: must be a table")
		return
	}
	for _, name := range fr.keys(t, "fields") {
		path, ok := t[name].(string)
		if !ok {
			errorf([]string{"fields", name}, "fields.%s: must be a string", name)
			continue
		}
		ref, err := screen.NewAlias(name, path)
		if err != nil {
			errorf([]string{"fields", name}, "fields: %v", err)
			continue
		}
		fr.fields[name] = ref
	}
}

func (fr *front) parseItems(v any, errorf errorFunc) {
	a, ok := v.([]any)
	if !ok {
		errorf([]string{"item"}, "item: must be an array of tables")
		return
	}
	for i, el := range a {
		idx := fmt.Sprint(i)
		t, ok := el.(map[string]any)
		if !ok {
			errorf([]string{"item", idx}, "item: must be an array of tables")
			continue
		}
		text := map[string]string{}
		for _, key := range fr.keys(t, "item", idx) {
			v := t[key]
			if key != "key" && key != "do" && key != "when" {
				errorf([]string{"item", idx, key}, "unknown key \"item.%s\"", key)
				continue
			}
			s, ok := v.(string)
			if !ok {
				errorf([]string{"item", idx, key}, "item %d: %s: must be a string", i+1, key)
				continue
			}
			text[key] = s
		}
		key := text["key"]
		label := fmt.Sprintf("item %q", key)
		if _, ok := text["key"]; !ok {
			label = fmt.Sprintf("item %d", i+1)
		}
		for _, k := range []string{"key", "do"} {
			if _, ok := t[k]; !ok {
				errorf([]string{"item", idx}, "%s: %s is missing", label, k)
			}
		}
		item := screen.Item{Key: key, Do: text["do"]}
		if when, ok := text["when"]; ok {
			c, err := screen.ParseCond(when, fr.fields, false)
			if err != nil {
				errorf([]string{"item", idx, "when"}, "%s: when: %v", label, err)
			}
			item.When = c
		}
		fr.items = append(fr.items, item)
	}
}

// keys gives the keys of the table t at the path in file order,
// so the diagnostics up to the limit of section 10 do not depend on the map order.
func (fr *front) keys(t map[string]any, path ...string) []string {
	pos := func(key string) tomlpos.Pos { return fr.doc.Pos(append(slices.Clip(path), key)...) }
	return slices.SortedFunc(maps.Keys(t), func(a, b string) int {
		pa, pb := pos(a), pos(b)
		return cmp.Or(cmp.Compare(pa.Line, pb.Line), cmp.Compare(pa.Col, pb.Col), strings.Compare(a, b))
	})
}

// pos gives the file position of the key with the path.
func (fr *front) pos(path ...string) screen.Pos {
	return fr.convert(fr.doc.Pos(path...))
}

// convert gives the file position of a front matter position:
// the line after "+++" and a column in code points.
func (fr *front) convert(p tomlpos.Pos) screen.Pos {
	col := p.Col
	if 1 <= p.Line && p.Line <= len(fr.lines) {
		l := fr.lines[p.Line-1]
		b := min(max(p.Col-1, 0), len(l))
		col = utf8.RuneCountInString(l[:b]) + 1
	}
	return screen.Pos{File: fr.file, Line: p.Line + 1, Col: col}
}
