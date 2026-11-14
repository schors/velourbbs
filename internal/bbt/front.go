package bbt

import (
	"cmp"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"unicode/utf8"

	"newbbs/internal/key"
	"newbbs/internal/screen"
	"newbbs/internal/tomlpos"
)

// front is the front matter of a screen (section 3.2).
type front struct {
	width  *screen.WidthLimits // nil when width is not set
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
	seen := map[string]bool{}
	for i, el := range a {
		idx := fmt.Sprint(i)
		t, ok := el.(map[string]any)
		if !ok {
			errorf([]string{"item", idx}, "item: must be an array of tables")
			continue
		}
		// The label of menu-nav, section 7: the key when it is valid, else the number.
		raw, _ := t["key"].(string)
		k, valid := key.Parse(raw)
		label := fmt.Sprintf("item %d", i+1)
		if valid {
			label = fmt.Sprintf("item %q", k)
		}
		item := screen.Item{Key: k}
		for _, name := range fr.keys(t, "item", idx) {
			path := []string{"item", idx, name}
			if name != "key" && name != "do" && name != "when" {
				errorf(path, "%s: unknown key %q", label, name)
				continue
			}
			s, ok := t[name].(string)
			if !ok {
				errorf(path, "%s: %s: must be a string", label, name)
				continue
			}
			switch name {
			case "key":
				switch {
				case !valid:
					errorf(path, "%s: key: invalid key %q", label, s)
				case seen[k]:
					errorf(path, "%s: key: duplicate key", label)
				}
				seen[k] = valid
			case "do":
				act, err := parseAction(s)
				if err != nil {
					errorf(path, "%s: do: %v", label, err)
				}
				item.Do = act
				item.DoPos = fr.pos(path...)
			case "when":
				c, err := screen.ParseCond(s, fr.fields, false)
				if err != nil {
					errorf(path, "%s: when: %v", label, err)
				}
				item.When = c
			}
		}
		for _, name := range []string{"key", "do"} {
			if _, ok := t[name]; !ok {
				errorf([]string{"item", idx}, "%s: %s is missing", label, name)
			}
		}
		fr.items = append(fr.items, item)
	}
}

// parseAction parses the do of an item (menu-nav, section 5).
func parseAction(do string) (screen.Action, error) {
	word, arg, hasArg := strings.Cut(do, " ")
	switch word {
	case "goto":
		if !hasArg {
			return screen.Action{}, errors.New("goto needs a screen name")
		}
		if !isName(arg) {
			return screen.Action{}, fmt.Errorf("invalid screen name %q", arg)
		}
		return screen.Action{Op: screen.Goto, Screen: arg}, nil
	case "logoff":
		if hasArg {
			return screen.Action{}, errors.New("logoff takes no argument")
		}
		return screen.Action{Op: screen.Logoff}, nil
	}
	return screen.Action{}, fmt.Errorf("unknown action %q", word)
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
