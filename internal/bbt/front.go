package bbt

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"unicode/utf8"

	"newbbs/internal/key"
	"newbbs/internal/screen"
	"newbbs/internal/tomlcheck"
)

// front is the front matter of a screen (section 3.2).
type front struct {
	width  *screen.WidthLimits // nil when width is not set
	fields screen.Fields
	items  []screen.Item
}

// frontDoc is the plain decode of a front matter. A nil pointer is a key that is not set.
type frontDoc struct {
	Width  *[]int64          `toml:"width"`
	Fields map[string]string `toml:"fields"`
	Item   []struct {
		Key  *string `toml:"key"`
		Do   *string `toml:"do"`
		When *string `toml:"when"`
	} `toml:"item"`
}

// frontValues is the decode of the values with a check (section 9.1).
type frontValues struct {
	Item []struct {
		Key itemKey    `toml:"key"`
		Do  itemAction `toml:"do"`
	} `toml:"item"`
}

type itemKey struct{ k string }

func (k *itemKey) UnmarshalText(text []byte) error {
	var ok bool
	if k.k, ok = key.Parse(string(text)); !ok {
		return fmt.Errorf("invalid key %q", text)
	}
	return nil
}

type itemAction struct{ a screen.Action }

func (a *itemAction) UnmarshalText(text []byte) error {
	var err error
	a.a, err = parseAction(string(text))
	return err
}

// parseFront parses the front matter lines. The first line of lines is line 2 of the file.
// The result is nil when the front matter has an error.
func parseFront(d *screen.FileDiags, lines []string) *front {
	var doc frontDoc
	var vals frontValues
	errs := tomlcheck.Decode([]byte(strings.Join(lines, "\n")), &doc, &vals)
	for _, e := range errs {
		// The line after "+++" and a column in code points.
		col := e.Col
		if 1 <= e.Line && e.Line <= len(lines) {
			l := lines[e.Line-1]
			col = utf8.RuneCountInString(l[:min(max(e.Col-1, 0), len(l))]) + 1
		}
		d.Errorf(e.Line+1, col, "%s", e.Msg)
	}
	if len(errs) > 0 {
		return nil
	}
	// Each other error has the position 2:1 (section 9.1).
	ok := true
	errorf := func(format string, args ...any) {
		d.Errorf(2, 1, format, args...)
		ok = false
	}
	fr := &front{fields: screen.Fields{}}
	if doc.Width != nil {
		w, err := screen.NewWidth(*doc.Width)
		if err != nil {
			errorf("width: %v", err)
		} else {
			fr.width = &w
		}
	}
	for _, name := range slices.Sorted(maps.Keys(doc.Fields)) {
		ref, err := screen.NewAlias(name, doc.Fields[name])
		if err != nil {
			errorf("fields: %v", err)
			continue
		}
		fr.fields[name] = ref
	}
	seen := map[string]bool{}
	for i, it := range doc.Item {
		v := vals.Item[i]
		// The label of menu-nav, section 7: the key, else the number.
		label := fmt.Sprintf("item %d", i+1)
		if it.Key != nil {
			label = fmt.Sprintf("item %q", v.Key.k)
		}
		item := screen.Item{Key: v.Key.k, Do: v.Do.a}
		switch {
		case it.Key == nil:
			errorf("%s: key is missing", label)
		case seen[item.Key]:
			errorf("%s: key: duplicate key", label)
		}
		seen[item.Key] = true
		if it.Do == nil {
			errorf("%s: do is missing", label)
		}
		if it.When != nil {
			c, err := screen.ParseCond(*it.When, fr.fields, false)
			if err != nil {
				errorf("%s: when: %v", label, err)
			}
			item.When = c
		}
		fr.items = append(fr.items, item)
	}
	if !ok {
		return nil
	}
	return fr
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
