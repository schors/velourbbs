package bbt

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/schors/velourbbs/internal/content"
	"github.com/schors/velourbbs/internal/key"
	"github.com/schors/velourbbs/internal/provider"
	"github.com/schors/velourbbs/internal/screen"
	"github.com/schors/velourbbs/internal/tomlcheck"
)

// front is the front matter of a screen (section 3.2).
type front struct {
	layout screen.LayoutKeys
	items  []screen.Item
	file   *screen.FileSpec // the keys of a file screen, else nil (file-view, section 3)
	list   *screen.ListSpec // the keys of a list screen, else nil (list-view, section 3)
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
	Data    *struct {
		From *string            `toml:"from"`
		Args *map[string]string `toml:"args"`
	} `toml:"data"`
	Rows *int64 `toml:"rows"`
	Item []struct {
		Key  *string            `toml:"key"`
		Do   *string            `toml:"do"`
		When *string            `toml:"when"`
		Args *map[string]string `toml:"args"`
	} `toml:"item"`
}

// parseFront parses the front matter lines. The first line of lines is line 2 of the file.
// fields is the aliases of screens/screens.toml. The result is nil when the front matter has an error.
// langs is the languages of the BBS.
func parseFront(d *screen.FileDiags, lines []string, fields screen.Fields, langs []string) *front {
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
	fr.list = parseListKeys(&doc, errorf)
	pages := doc.File != nil || doc.Data != nil // a page action is valid (file-view, section 3.3)
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
		} else if a, err := parseAction(*it.Do, langs); err != nil {
			errorf("%s: do: %v", label, err)
		} else if a.Op.IsPage() && !pages {
			errorf("%s: do: %q is valid only in a file screen or a list screen", label, *it.Do)
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
			} else if f := c.ListField(); f != "" && doc.Data == nil {
				errorf("%s: when: field %q is valid only in a list screen", label, f)
			} else if r := rowField(c); r != "" {
				errorf("%s: when: field %q is valid only in a list line", label, r)
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
	if name, token := argToken(v); token {
		if name == "" {
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

// rowField gives the path of the first row.<field> of c, else "" (list-view, section 6).
func rowField(c *screen.Cond) string {
	for _, r := range c.Refs() {
		if r.Row() != "" {
			return r.String()
		}
	}
	return ""
}

// argToken reports whether the value v starts with "{": a token. name is the <name> of a value
// that is exactly one token {arg.<name>}, else "" (bbt-format, section 8.5).
func argToken(v string) (name string, token bool) {
	t, ok := strings.CutPrefix(v, "{")
	if !ok {
		return "", false
	}
	name, ok = strings.CutPrefix(strings.TrimSuffix(t, "}"), "arg.")
	if !ok || !strings.HasSuffix(t, "}") || !screen.IsIdent(name) {
		return "", true
	}
	return name, true
}

// The range of the key rows (list-view, section 3).
const (
	minListRows = 1
	maxListRows = 100
)

// parseListKeys parses the keys data and rows (list-view, sections 3 and 4).
// It gives nil without data. The loader reads the rows of the provider.
func parseListKeys(doc *frontDoc, errorf func(format string, args ...any)) *screen.ListSpec {
	if doc.Data == nil {
		if doc.Rows != nil {
			errorf("rows: the screen has no data")
		}
		return nil
	}
	if doc.File != nil {
		errorf("data: a list screen cannot have file")
	}
	l := &screen.ListSpec{Args: map[string]string{}, ArgRefs: map[string]string{}, Lists: map[string][]screen.ListRow{}}
	var args map[string]string
	if doc.Data.Args != nil {
		args = *doc.Data.Args
	}
	var p *provider.Provider
	if doc.Data.From == nil {
		errorf("data: from is missing")
	} else if found, ok := provider.Lookup(*doc.Data.From); !ok {
		errorf("data: from: unknown provider %q", *doc.Data.From)
	} else {
		p, l.From = found, found.Name
		for _, err := range p.CheckArgs(slices.Sorted(maps.Keys(args))) {
			errorf("data: args: %v", err)
		}
	}
	for _, name := range slices.Sorted(maps.Keys(args)) {
		v := args[name]
		if arg, token := argToken(v); token {
			if arg == "" {
				errorf("data: args: %s: invalid token %q, want a value or {arg.<name>}", name, v)
			}
			l.ArgRefs[name] = arg
			continue
		}
		if p != nil && slices.ContainsFunc(p.Args, func(a provider.Arg) bool { return a.Name == name }) {
			if err := p.Check(name, v); err != nil {
				errorf("data: args: %s: %v", name, err)
			}
		}
		l.Args[name] = v
	}
	switch {
	case doc.Rows == nil:
		errorf("rows is missing")
	case *doc.Rows < minListRows || *doc.Rows > maxListRows:
		errorf("rows: %d, must be from %d to %d", *doc.Rows, minListRows, maxListRows)
	default:
		l.Rows = int(*doc.Rows)
	}
	return l
}

// parseAction parses the do of an item (section 8.3). langs is the languages of the BBS
// (session-language, section 2).
func parseAction(do string, langs []string) (screen.Action, error) {
	word, arg, hasArg := strings.Cut(do, " ")
	switch word {
	case "goto", "gosub":
		if !hasArg {
			return screen.Action{}, fmt.Errorf("%s needs a screen name", word)
		}
		code, name, hasCode := strings.Cut(arg, "/")
		if !hasCode {
			code, name = "", arg
		}
		if !screen.IsName(name) || hasCode && !screen.IsLang(code) {
			return screen.Action{}, fmt.Errorf("invalid screen name %q", arg)
		}
		if hasCode && !slices.Contains(langs, code) {
			return screen.Action{}, fmt.Errorf("%q is not a language of the BBS", code)
		}
		return screen.Action{Op: targetOps[word], Screen: name, Lang: code}, nil
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
	case "lang":
		switch {
		case !hasArg || !screen.IsLang(arg):
			return screen.Action{}, fmt.Errorf("invalid lang action %q", do)
		case !slices.Contains(langs, arg):
			return screen.Action{}, fmt.Errorf("%q is not a language of the BBS", arg)
		}
		return screen.Action{Op: screen.Lang, Lang: arg}, nil
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
