package bbt

import (
	"fmt"
	"maps"
	"math"
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
	at     tomlcheck.Place // the root table, with positions in the screen file
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
	Gutter  *int64   `toml:"gutter"`
	File    *string  `toml:"file"`
	Lines   *[]int64 `toml:"lines"`
	Numbers *bool    `toml:"numbers"`
	Margin  *[]int64 `toml:"margin"`
	Data    *struct {
		From *string            `toml:"from"`
		Args *map[string]string `toml:"args"`
	} `toml:"data"`
	Rows   *int64             `toml:"rows"`
	Fields *map[string]string `toml:"fields"`
	Select *struct {
		Do   *string            `toml:"do"`
		Args *map[string]string `toml:"args"`
	} `toml:"select"`
	Item []struct {
		Key    *string            `toml:"key"`
		Do     *string            `toml:"do"`
		When   *string            `toml:"when"`
		Args   *map[string]string `toml:"args"`
		Ask    *string            `toml:"ask"`
		AskMax *int64             `toml:"ask_max"`
	} `toml:"item"`
}

// parseFront parses the front matter lines. The first line of lines is line 2 of the file.
// fields is the aliases of screens/screens.toml. The front is nil when the front matter has an error.
// own is the valid aliases of [fields] over fields, also after an error, else nil without [fields]:
// the front matter and the body use them (section 4.2). langs is the languages of the BBS.
func parseFront(d *screen.FileDiags, lines []string, fields screen.Fields, langs []string) (fr *front, own screen.Fields) {
	var doc frontDoc
	data := []byte(strings.Join(lines, "\n"))
	root, errs := tomlcheck.DecodeFront(data, &doc)
	for _, e := range errs {
		d.Errorf(e.Line, e.Col, "%s", e.Msg)
	}
	if len(errs) > 0 {
		// The document gives only these errors (screen-render, section 3.1): the aliases
		// that it decodes stay for the body without their errors.
		return nil, ownFields(doc.Fields, root, fields, func(tomlcheck.Pos, string, ...any) {})
	}
	// Each other error has the position of the key path of its message (screen-render, section 3.1).
	ok := true
	errorf := func(at tomlcheck.Pos, format string, args ...any) {
		d.Errorf(at.Line, at.Col, format, args...)
		ok = false
	}
	fr = &front{at: root, layout: screen.ParseLayoutKeys(doc.Width, doc.Align, doc.VAlign, doc.Fill, doc.Gutter,
		func(key string, err error) { errorf(root.At(key).Value(), "%s: %v", key, err) })}
	if own = ownFields(doc.Fields, root, fields, errorf); own != nil {
		fields = own
	}
	fr.file = parseFileKeys(&doc, root, errorf)
	fr.list = parseListKeys(&doc, root, langs, errorf)
	pages := doc.File != nil || doc.Data != nil // a page action is valid (file-view, section 3.3)
	seen := map[string]bool{}
	for i, it := range doc.Item {
		at := ItemPlace(root, i)
		// The label of menu-nav, section 6: the valid key, else the number.
		label := fmt.Sprintf("item %d", i+1)
		var item screen.Item
		if it.Key == nil {
			errorf(at.Value(), "%s: key is missing", label)
		} else if k, valid := key.Parse(*it.Key); !valid {
			errorf(at.At("key").Value(), "%s: key: invalid key %q", label, *it.Key)
		} else {
			item.Key = k
			label = fmt.Sprintf("item %q", k)
			if seen[k] {
				errorf(at.At("key").Value(), "%s: key: duplicate key", label)
			}
			if doc.Select != nil && key.IsCursor(k) {
				errorf(at.At("key").Value(), "%s: key: %q is a key of select", label, k)
			}
			seen[k] = true
		}
		if it.Do == nil {
			errorf(at.Value(), "%s: do is missing", label)
		} else if a, err := parseAction(*it.Do, langs); err != nil {
			errorf(at.At("do").Value(), "%s: do: %v", label, err)
		} else if a.Op.IsPage() && !pages {
			errorf(at.At("do").Value(), "%s: do: %q is valid only in a file screen or a list screen", label, *it.Do)
		} else {
			item.Do = a
		}
		if it.Args != nil {
			if item.Do.Op != 0 && !item.Do.Op.HasTarget() { // Op 0: the do has an error
				errorf(at.At("args").Value(), "%s: args: only goto and gosub take arguments", label)
			}
			item.Do.Args = map[string]string{}
			for _, name := range slices.Sorted(maps.Keys(*it.Args)) {
				v := (*it.Args)[name]
				pos := argPos(at.At("args"), name)
				kind, ref, token := tokenValue(v)
				switch {
				case !token:
					if err := screen.CheckArg(name, v); err != nil {
						errorf(pos, "%s: args: %v", label, err)
					}
					item.Do.Args[name] = v
				case !screen.IsIdent(name):
					errorf(pos, "%s: args: invalid argument name %q", label, name)
				case kind == "in" && it.Ask == nil:
					errorf(pos, "%s: args: %s: {in} is valid only in an item with ask", label, name)
				case kind == "in":
					item.Do.In = append(item.Do.In, name) // sorted: the names are in order
				case kind != "arg":
					errorf(pos, "%s: args: %s: invalid token %q, want a value, {arg.<name>} or {in}", label, name, v)
				default:
					if item.Do.ArgRefs == nil {
						item.Do.ArgRefs = map[string]string{}
					}
					item.Do.ArgRefs[name] = ref
				}
			}
		}
		item.Ask = parseAsk(it.Ask, it.AskMax, item.Do, at, label, errorf)
		if it.When != nil {
			when := at.At("when").Value()
			c, err := screen.ParseCond(*it.When, fields, false)
			if err != nil {
				errorf(when, "%s: when: %v", label, err)
			} else if f := c.FileField(); f != "" && doc.File == nil {
				errorf(when, "%s: when: field %q is valid only in a file screen", label, f)
			} else if f := c.ListField(); f != "" && doc.Data == nil {
				errorf(when, "%s: when: field %q is valid only in a list screen", label, f)
			} else if r := rowField(c); r != "" {
				errorf(when, "%s: when: field %q is valid only in a list line", label, r)
			}
			item.When = c
		}
		fr.items = append(fr.items, item)
	}
	if !ok {
		return nil, own
	}
	return fr, own
}

// parseAsk parses the keys ask and ask_max of the item at the place at with the action a
// and the label label (section 8.6). It gives nil without ask. The loader checks the free cells.
func parseAsk(ask *string, max *int64, a screen.Action, at tomlcheck.Place, label string, errorf errorFunc) *screen.Ask {
	if ask == nil {
		if max != nil {
			errorf(at.At("ask_max").Value(), "%s: ask_max: the item has no ask", label)
		}
		return nil
	}
	s := &screen.Ask{Prompt: *ask}
	if err := screen.CheckText(*ask); err != nil {
		errorf(at.At("ask").Value(), "%s: ask: %v", label, err)
	}
	if max != nil {
		if *max < 1 {
			errorf(at.At("ask_max").Value(), "%s: ask_max: %d, must be at least 1", label, *max)
		} else {
			s.Max = int(min(*max, math.MaxInt32))
		}
	}
	// Op 0: the do has an error.
	if a.Op != 0 && !a.Op.HasTarget() {
		errorf(at.At("ask").Value(), "%s: ask: the action must be goto or gosub", label)
	} else if a.Op != 0 && len(a.In) == 0 {
		errorf(at.At("ask").Value(), "%s: ask: args has no value {in}", label)
	}
	return s
}

// errorFunc reports an error of a front matter at a position in the screen file.
type errorFunc func(at tomlcheck.Pos, format string, args ...any)

// ownFields gives the valid aliases of the [fields] table over fields, else nil without the table
// (section 4.2). root is the root table of the front matter. errorf reports the errors of the aliases.
func ownFields(table *map[string]string, root tomlcheck.Place, fields screen.Fields, errorf errorFunc) screen.Fields {
	if table == nil {
		return nil
	}
	own := maps.Clone(fields)
	if own == nil {
		own = screen.Fields{}
	}
	maps.Copy(own, screen.NewFields(*table, root, errorf))
	return own
}

// ItemPlace gives the place of the 0-based item i of the front matter with the root table root.
func ItemPlace(root tomlcheck.Place, i int) tomlcheck.Place {
	return root.At("item").Elem(i)
}

// argPos gives the position of an error of the argument name in the table args (screen.CheckArg):
// the key of an invalid name, else the value (screen-render, section 3.1).
func argPos(args tomlcheck.Place, name string) tomlcheck.Pos {
	if !screen.IsIdent(name) {
		return args.At(name).Key()
	}
	return args.At(name).Value()
}

// parseFileKeys parses the keys file, lines, numbers and margin (file-view, section 3).
// It gives nil without file. The loader checks margin against the width limits.
// root is the root table of the front matter.
func parseFileKeys(doc *frontDoc, root tomlcheck.Place, errorf errorFunc) *screen.FileSpec {
	if doc.File == nil {
		if doc.Lines != nil {
			errorf(root.At("lines").Value(), "lines: the screen has no file")
		}
		if doc.Numbers != nil {
			errorf(root.At("numbers").Value(), "numbers: the screen has no file")
		}
		if doc.Margin != nil {
			errorf(root.At("margin").Value(), "margin: the screen has no file")
		}
		return nil
	}
	f := &screen.FileSpec{Contents: map[string]*screen.Content{}}
	v := *doc.File
	if kind, name, token := tokenValue(v); token {
		if kind != "arg" {
			errorf(root.At("file").Value(), "file: invalid token %q, want a content path or {arg.<name>}", v)
		}
		f.Arg = name
	} else if err := content.CheckPath(v); err != nil {
		errorf(root.At("file").Value(), "file: %v", err)
	} else {
		f.Path = v
	}
	if p := doc.Lines; p != nil {
		switch l := *p; {
		case len(l) != 2:
			errorf(root.At("lines").Value(), "lines: must be an array of 2 integers")
		case l[0] < 1 || l[0] > l[1]:
			errorf(root.At("lines").Value(), "lines: [%d, %d] is not in 1 <= first <= last", l[0], l[1])
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
			errorf(root.At("margin").Value(), "margin: must be an array of 2 integers")
		case m[0] < 0 || m[1] < 0:
			errorf(root.At("margin").Value(), "margin: [%d, %d] has a value less than 0", m[0], m[1])
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

// tokenValue reports whether the value v starts with "{": a token. For a value that is exactly
// one token {arg.<name>} or {row.<name>}, kind is "arg" or "row" and name is the <name>.
// For {in}, kind is "in" and name is "". Else both are ""
// (bbt-format, section 8.5).
func tokenValue(v string) (kind, name string, token bool) {
	t, ok := strings.CutPrefix(v, "{")
	if !ok {
		return "", "", false
	}
	t, ok = strings.CutSuffix(t, "}")
	if !ok {
		return "", "", true
	}
	if t == "in" {
		return "in", "", true
	}
	for _, k := range []string{"arg", "row"} {
		if n, ok := strings.CutPrefix(t, k+"."); ok && screen.IsIdent(n) {
			return k, n, true
		}
	}
	return "", "", true
}

// The range of the key rows (list-view, section 3).
const (
	minListRows = 1
	maxListRows = 100
)

// parseListKeys parses the keys data, rows and select (list-view, sections 3, 4 and 8).
// It gives nil without data. The loader reads the rows of the provider.
// root is the root table of the front matter. langs is the languages of the BBS.
func parseListKeys(doc *frontDoc, root tomlcheck.Place, langs []string, errorf errorFunc) *screen.ListSpec {
	if doc.Data == nil {
		if doc.Rows != nil {
			errorf(root.At("rows").Value(), "rows: the screen has no data")
		}
		if doc.Select != nil {
			errorf(root.At("select").Value(), "select: the screen has no data")
		}
		return nil
	}
	data := root.At("data")
	if doc.File != nil {
		errorf(data.Value(), "data: a list screen cannot have file")
	}
	l := &screen.ListSpec{Args: map[string]string{}, ArgRefs: map[string]string{}, Lists: map[string][]screen.ListRow{}}
	var args map[string]string
	if doc.Data.Args != nil {
		args = *doc.Data.Args
	}
	var p *provider.Provider
	if doc.Data.From == nil {
		errorf(data.At("from").Value(), "data: from is missing")
	} else if found, ok := provider.Lookup(*doc.Data.From); !ok {
		errorf(data.At("from").Value(), "data: from: unknown provider %q", *doc.Data.From)
	} else {
		p, l.From = found, found.Name
		for _, err := range p.CheckArgs(slices.Sorted(maps.Keys(args))) {
			errorf(data.At("args").Value(), "data: args: %v", err)
		}
	}
	for _, name := range slices.Sorted(maps.Keys(args)) {
		v := args[name]
		if kind, arg, token := tokenValue(v); token {
			if kind != "arg" {
				errorf(data.At("args", name).Value(), "data: args: %s: invalid token %q, want a value or {arg.<name>}", name, v)
			}
			l.ArgRefs[name] = arg
			continue
		}
		if p != nil && slices.ContainsFunc(p.Args, func(a provider.Arg) bool { return a.Name == name }) {
			if err := p.Check(name, v); err != nil {
				errorf(data.At("args", name).Value(), "data: args: %s: %v", name, err)
			}
		}
		l.Args[name] = v
	}
	switch {
	case doc.Rows == nil:
		errorf(root.At("rows").Value(), "rows is missing")
	case *doc.Rows < minListRows || *doc.Rows > maxListRows:
		errorf(root.At("rows").Value(), "rows: %d, must be from %d to %d", *doc.Rows, minListRows, maxListRows)
	default:
		l.Rows = int(*doc.Rows)
	}
	if doc.Select != nil {
		l.Select = parseSelect(doc.Select.Do, doc.Select.Args, root.At("select"), p, langs, errorf)
	}
	return l
}

// parseSelect parses the key select: its do and its args (list-view, section 8).
// at is the place of select. p is the provider of the screen, nil after an error of data.
func parseSelect(do *string, args *map[string]string, at tomlcheck.Place, p *provider.Provider, langs []string,
	errorf errorFunc) *screen.Select {
	s := &screen.Select{Refs: map[string]string{}}
	if do == nil {
		errorf(at.At("do").Value(), "select: do is missing")
	} else if a, err := parseAction(*do, langs); err != nil {
		errorf(at.At("do").Value(), "select: do: %v", err)
	} else {
		s.Do = a
	}
	if args == nil {
		return s
	}
	if s.Do.Op != 0 && !s.Do.Op.HasTarget() { // Op 0: the do has an error
		errorf(at.At("args").Value(), "select: args: only goto and gosub take arguments")
	}
	s.Do.Args = map[string]string{}
	for _, name := range slices.Sorted(maps.Keys(*args)) {
		v := (*args)[name]
		pos := argPos(at.At("args"), name)
		kind, field, token := tokenValue(v)
		switch {
		case !token:
			if err := screen.CheckArg(name, v); err != nil {
				errorf(pos, "select: args: %v", err)
			}
			s.Do.Args[name] = v
			continue
		case !screen.IsIdent(name):
			errorf(pos, "select: args: invalid argument name %q", name)
			continue
		case kind == "arg":
			if s.Do.ArgRefs == nil {
				s.Do.ArgRefs = map[string]string{}
			}
			s.Do.ArgRefs[name] = field
			continue
		case kind != "row":
			errorf(pos, "select: args: %s: invalid token %q, want a value, {arg.<name>} or {row.<field>}", name, v)
		case field == screen.RowSelected:
			errorf(pos, "select: args: %s: field %q is valid only in a condition", name, "row."+field)
		case p != nil && !p.HasField(field):
			errorf(pos, "select: args: %s: provider %q has no field %q", name, p.Name, field)
		}
		s.Refs[name] = field
	}
	return s
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
