package bbt

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/schors/velourbbs/internal/key"
	"github.com/schors/velourbbs/internal/screen"
	"github.com/schors/velourbbs/internal/tomlcheck"
)

// front is the front matter of a screen (section 3.2).
type front struct {
	width *screen.WidthLimits // nil when width is not set
	items []screen.Item
}

// frontDoc is the plain decode of a front matter. A nil pointer is a key that is not set.
type frontDoc struct {
	Width *[]int64 `toml:"width"`
	Item  []struct {
		Key  *string `toml:"key"`
		Do   *string `toml:"do"`
		When *string `toml:"when"`
	} `toml:"item"`
}

// parseFront parses the front matter lines. The first line of lines is line 2 of the file.
// fields is the aliases of bbs.toml. The result is nil when the front matter has an error.
func parseFront(d *screen.FileDiags, lines []string, fields screen.Fields) *front {
	var doc frontDoc
	errs := tomlcheck.Decode([]byte(strings.Join(lines, "\n")), &doc)
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
	fr := &front{}
	if doc.Width != nil {
		w, err := screen.NewWidth(*doc.Width)
		if err != nil {
			errorf("width: %v", err)
		} else {
			fr.width = &w
		}
	}
	seen := map[string]bool{}
	for i, it := range doc.Item {
		// The label of menu-nav, section 7: the valid key, else the number.
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
		} else {
			item.Do = a
		}
		if it.When != nil {
			c, err := screen.ParseCond(*it.When, fields, false)
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
