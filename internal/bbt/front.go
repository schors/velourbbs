package bbt

import (
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
	Margin  *[]int64 `toml:"margin"`
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
			if item.Do.Op != 0 && !item.Do.Op.HasTarget() { // Op 0: the do has an error
				errorf("%s: args: only goto and gosub take arguments", label)
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

// parseFileKeys parses the keys file, lines, numbers and margin (file-view, section 3).
// It gives nil without file. The loader checks margin against the width limits.
func parseFileKeys(doc *frontDoc, errorf func(format string, args ...any)) *screen.FileSpec {
	if doc.File == nil {
		if doc.Lines != nil {
			errorf("lines: the screen has no file")
		}
		if doc.Numbers != nil {
			errorf("numbers: the screen has no file")
		}
		if doc.Margin != nil {
			errorf("margin: the screen has no file")
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
	if p := doc.Margin; p != nil {
		switch m := *p; {
		case len(m) != 2:
			errorf("margin: must be an array of 2 integers")
		case m[0] < 0 || m[1] < 0:
			errorf("margin: [%d, %d] has a value less than 0", m[0], m[1])
		default:
			f.Left, f.Right = int(m[0]), int(m[1])
		}
	}
	return f
}

// parseAction parses the do of an item (section 8.3).
func parseAction(do string) (screen.Action, error) {
	word, arg, hasArg := strings.Cut(do, " ")
	switch word {
	case "goto", "gosub":
		if !hasArg {
			return screen.Action{}, fmt.Errorf("%s needs a screen name", word)
		}
		if !screen.IsName(arg) {
			return screen.Action{}, fmt.Errorf("invalid screen name %q", arg)
		}
		return screen.Action{Op: targetOps[word], Screen: arg}, nil
	case "logoff", "back", "top":
		if hasArg {
			return screen.Action{}, fmt.Errorf("%s takes no argument", word)
		}
		return screen.Action{Op: plainOps[word]}, nil
	case "page":
		if op, ok := pageOps[arg]; ok && hasArg {
			return screen.Action{Op: op}, nil
		}
		return screen.Action{}, fmt.Errorf("invalid page action %q", do)
	case "theme":
		if hasArg && (arg == "dark" || arg == "light" || arg == "auto") {
			return screen.Action{Op: screen.Theme, Theme: arg}, nil
		}
		return screen.Action{}, fmt.Errorf("invalid theme action %q", do)
	}
	return screen.Action{}, fmt.Errorf("unknown action %q", word)
}

// targetOps is the actions with a screen name, by their word.
var targetOps = map[string]screen.Op{"goto": screen.Goto, "gosub": screen.Gosub}

// plainOps is the actions without an argument, by their word.
var plainOps = map[string]screen.Op{"logoff": screen.Logoff, "back": screen.Back, "top": screen.Top}

// pageOps is the page actions of file-view, section 3.3, by their word after "page".
var pageOps = map[string]screen.Op{
	"next": screen.PageNext, "prev": screen.PagePrev, "first": screen.PageFirst, "last": screen.PageLast,
}
