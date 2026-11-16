package bbt

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/schors/velourbbs/internal/content"
	"github.com/schors/velourbbs/internal/key"
	"github.com/schors/velourbbs/internal/screen"
	"github.com/schors/velourbbs/internal/tomlcheck"
)

// front is the front matter of a screen (section 3.2).
type front struct {
	layout screen.LayoutKeys
	items  []screen.Item
	file   *screen.FileSpec // the keys of a file screen, else nil (file-view, section 3)
}

// frontDoc is the plain decode of a front matter. A nil pointer is a key that is not set.
type frontDoc struct {
	Width   *[]int64 `toml:"width"`
	Align   *string  `toml:"align"`
	VAlign  *string  `toml:"valign"`
	Fill    *string  `toml:"fill"`
	File    *string  `toml:"file"`
	Lines   *[]int64 `toml:"lines"`
	Numbers *bool    `toml:"numbers"`
	Item    []struct {
		Key  *string            `toml:"key"`
		Do   *string            `toml:"do"`
		When *string            `toml:"when"`
		Args *map[string]string `toml:"args"`
	} `toml:"item"`
}

// parseFront parses the front matter lines. The first line of lines is line 2 of the file.
// fields is the aliases of screens/screens.toml. The result is nil when the front matter has an error.
func parseFront(d *screen.FileDiags, lines []string, fields screen.Fields) *front {
	var doc frontDoc
	data := []byte(strings.Join(lines, "\n"))
	errs := tomlcheck.Decode(data, &doc)
	for _, e := range errs {
		// The line after "+++" and a column in code points.
		d.Errorf(e.Line+1, e.RuneCol(data), "%s", e.Msg)
	}
	if len(errs) > 0 {
		return nil
	}
	// Each other error has the position 2:1 (screen-render, section 3.1).
	ok := true
	errorf := func(format string, args ...any) {
		d.Errorf(2, 1, format, args...)
		ok = false
	}
	fr := &front{layout: screen.ParseLayoutKeys(doc.Width, doc.Align, doc.VAlign, doc.Fill,
		func(key string, err error) { errorf("%s: %v", key, err) })}
	fr.file = parseFileKeys(&doc, errorf)
	seen := map[string]bool{}
	for i, it := range doc.Item {
		// The label of menu-nav, section 6: the valid key, else the number.
		label := fmt.Sprintf("item %d", i+1)
		var item screen.Item
		if it.Key == nil {
			errorf("%s: key is missing", label)
		} else if k, valid := key.Parse(*it.Key); !valid {
			errorf("%s: key: invalid key %q", label, *it.Key)
		} else {
			item.Key = k
			label = fmt.Sprintf("item %q", k)
			if seen[k] {
				errorf("%s: key: duplicate key", label)
			}
			seen[k] = true
		}
		if it.Do == nil {
			errorf("%s: do is missing", label)
		} else if a, err := parseAction(*it.Do); err != nil {
			errorf("%s: do: %v", label, err)
		} else if a.Op.IsPage() && doc.File == nil {
			errorf("%s: do: %q is valid only in a file screen", label, *it.Do)
		} else {
			item.Do = a
		}
		if it.Args != nil {
			if item.Do.Op != 0 && item.Do.Op != screen.Goto { // Op 0: the do has an error
				errorf("%s: args: only goto takes arguments", label)
			}
			for _, name := range slices.Sorted(maps.Keys(*it.Args)) {
				if err := screen.CheckArg(name, (*it.Args)[name]); err != nil {
					errorf("%s: args: %v", label, err)
				}
			}
			item.Do.Args = *it.Args
		}
		if it.When != nil {
			c, err := screen.ParseCond(*it.When, fields, false)
			if err != nil {
				errorf("%s: when: %v", label, err)
			} else if f := c.FileField(); f != "" && doc.File == nil {
				errorf("%s: when: field %q is valid only in a file screen", label, f)
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

// parseFileKeys parses the keys file, lines and numbers (file-view, section 3).
// It gives nil without file.
func parseFileKeys(doc *frontDoc, errorf func(format string, args ...any)) *screen.FileSpec {
	if doc.File == nil {
		if doc.Lines != nil {
			errorf("lines: the screen has no file")
		}
		if doc.Numbers != nil {
			errorf("numbers: the screen has no file")
		}
		return nil
	}
	f := &screen.FileSpec{Contents: map[string]*screen.Content{}}
	v := *doc.File
	if token, ok := strings.CutPrefix(v, "{"); ok {
		name, ok := strings.CutPrefix(strings.TrimSuffix(token, "}"), "arg.")
		if !ok || !strings.HasSuffix(token, "}") || !screen.IsIdent(name) {
			errorf("file: invalid token %q, want a content path or {arg.<name>}", v)
		}
		f.Arg = name
	} else if err := content.CheckPath(v); err != nil {
		errorf("file: %v", err)
	} else {
		f.Path = v
	}
	if p := doc.Lines; p != nil {
		switch l := *p; {
		case len(l) != 2:
			errorf("lines: must be an array of 2 integers")
		case l[0] < 1 || l[0] > l[1]:
			errorf("lines: [%d, %d] is not in 1 <= first <= last", l[0], l[1])
		default:
			f.First, f.Last = int(l[0]), int(l[1])
		}
	}
	if doc.Numbers != nil {
		f.Numbers, f.NumbersKey = *doc.Numbers, true
	}
	return f
}

// parseAction parses the do of an item (section 8.3).
func parseAction(do string) (screen.Action, error) {
	word, arg, hasArg := strings.Cut(do, " ")
	switch word {
	case "goto":
		if !hasArg {
			return screen.Action{}, errors.New("goto needs a screen name")
		}
		if !screen.IsName(arg) {
			return screen.Action{}, fmt.Errorf("invalid screen name %q", arg)
		}
		return screen.Action{Op: screen.Goto, Screen: arg}, nil
	case "logoff":
		if hasArg {
			return screen.Action{}, errors.New("logoff takes no argument")
		}
		return screen.Action{Op: screen.Logoff}, nil
	case "page":
		if op, ok := pageOps[arg]; ok && hasArg {
			return screen.Action{Op: op}, nil
		}
		return screen.Action{}, fmt.Errorf("invalid page action %q", do)
	}
	return screen.Action{}, fmt.Errorf("unknown action %q", word)
}

// pageOps is the page actions of file-view, section 3.3, by their word after "page".
var pageOps = map[string]screen.Op{
	"next": screen.PageNext, "prev": screen.PagePrev, "first": screen.PageFirst, "last": screen.PageLast,
}
