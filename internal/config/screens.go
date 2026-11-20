package config

import (
	"bytes"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/schors/velourbbs/internal/screen"
	"github.com/schors/velourbbs/internal/tomlcheck"
)

// defaultLang is the default of [screen] lang (screens-toml, section 3.1).
const defaultLang = "en"

// The defaults of screens-toml, sections 3.1 and 3.2.
var (
	defaultLayout = screen.Layout{Width: screen.WidthLimits{Min: 40, Max: 80}}
	defaultTime   = "2006-01-02 15:04:05"
	defaultNormal = screen.Style{Fg: screen.IndexColour(7), Bg: screen.IndexColour(0)}
	lightNormal   = screen.Style{Fg: screen.IndexColour(0), Bg: screen.IndexColour(15)}
	// defaultNormalStyle is normal of the base theme: it has both colours, so the light
	// theme without a light normal keeps 7/0.
	defaultNormalStyle = screen.StyleSpec{
		Fg: screen.IndexColour(7), Bg: screen.IndexColour(0), HasFg: true, HasBg: true,
	}
	defaultMoreStyle = screen.StyleSpec{
		Fg: screen.IndexColour(15), Bg: screen.IndexColour(0), HasFg: true, HasBg: true,
	}
	// defaultSelectStyle is the style of the selected row in art (list-view, section 9).
	defaultSelectStyle = screen.StyleSpec{
		Fg: screen.IndexColour(0), Bg: screen.IndexColour(7), HasFg: true, HasBg: true,
	}
)

// screensDoc is the plain decode of screens/screens.toml. A nil pointer is a key that is not set.
type screensDoc struct {
	Screen struct {
		Width  *[]int64 `toml:"width"`
		Align  *string  `toml:"align"`
		VAlign *string  `toml:"valign"`
		Fill   *string  `toml:"fill"`
		Gutter *int64   `toml:"gutter"`
		More   *string  `toml:"more"`
		Time   *string  `toml:"time"`
		Zone   *string  `toml:"zone"`
		Theme  *string  `toml:"theme"`
		Lang   *string  `toml:"lang"`
	} `toml:"screen"`
	Lang map[string]struct {
		More *string `toml:"more"`
	} `toml:"lang"`
	Theme struct {
		Default map[string]string `toml:"default"`
		C16     map[string]string `toml:"16"`
		C256    map[string]string `toml:"256"`
		Light   struct {
			Default map[string]string `toml:"default"`
			C16     map[string]string `toml:"16"`
			C256    map[string]string `toml:"256"`
		} `toml:"light"`
	} `toml:"theme"`
	Fields map[string]string `toml:"fields"`
	Menu   struct {
		Start *string `toml:"start"`
	} `toml:"menu"`
}

// parseScreens reads screens/screens.toml from data. The values of bbs.toml in the result are the defaults.
// An unknown key is an error (screens-toml, section 2).
// It gives at most screen.MaxFileDiags diagnostics (screen-render, section 3.1).
func parseScreens(data []byte) (Config, screen.Diags) {
	cfg := Config{
		Layout: defaultLayout, Lang: defaultLang, Time: defaultTime, Zone: time.Local, Fields: screen.Fields{}, Start: "main",
		StartPos: tomlcheck.Pos{Line: 1, Col: 1}, ThemeChoice: ThemeDark, Telnet: defaultTelnet, Limits: defaultLimits,
		Name: defaultName,
	}
	fd := screen.FileDiags{File: ScreensFile}
	var d screensDoc
	root, errs := tomlcheck.Decode(bytes.TrimPrefix(data, screen.BOM), &d)
	for _, e := range errs {
		fd.Errorf(e.Line, e.Col, "%s", e.Msg)
	}
	if len(errs) > 0 {
		theme(&cfg, screensDoc{}, root, nil)
		langs(&cfg, screensDoc{}, root, nil)
		return cfg, fd.Diags
	}
	errorf := func(at tomlcheck.Pos, format string, args ...any) { fd.Errorf(at.Line, at.Col, format, args...) }
	scr := root.At("screen")
	cfg.Layout = screen.ParseLayoutKeys(d.Screen.Width, d.Screen.Align, d.Screen.VAlign, d.Screen.Fill, d.Screen.Gutter,
		func(key string, err error) { errorf(scr.At(key).Value(), "screen.%s: %v", key, err) }).Over(defaultLayout)
	cfg.More = more(d.Screen.More, scr.At("more"), "screen.more", errorf)
	if p := d.Screen.Time; p != nil {
		cfg.Time = *p
	}
	if p := d.Screen.Zone; p != nil {
		// LoadLocation gives UTC for an empty name: the key wants a name.
		if loc, err := time.LoadLocation(*p); err != nil || *p == "" {
			errorf(scr.At("zone").Value(), "screen.zone: unknown time zone %q", *p)
		} else {
			cfg.Zone = loc
		}
	}
	if p := d.Screen.Theme; p != nil {
		switch c := ThemeChoice(*p); c {
		case ThemeDark, ThemeLight, ThemeAuto:
			cfg.ThemeChoice = c
		default:
			errorf(scr.At("theme").Value(), "screen.theme: invalid value %q, want \"dark\", \"light\" or \"auto\"", *p)
		}
	}
	if p := d.Screen.Lang; p != nil {
		if !screen.IsLang(*p) {
			errorf(scr.At("lang").Value(), "screen.lang: invalid language code %q", *p)
		} else {
			cfg.Lang = *p
		}
	}
	langs(&cfg, d, root, errorf)
	cfg.StartPos = root.At("menu", "start").Value()
	if p := d.Menu.Start; p != nil {
		if !screen.IsName(*p) {
			errorf(cfg.StartPos, "menu.start: invalid screen name %q", *p)
		} else {
			cfg.Start = *p
		}
	}
	theme(&cfg, d, root, errorf)
	for _, name := range slices.Sorted(maps.Keys(d.Fields)) {
		at := root.At("fields", name)
		if err := screen.CheckAlias(name); err != nil {
			errorf(at.Key(), "fields: %v", err)
			continue
		}
		ref, err := screen.NewAlias(name, d.Fields[name])
		if err != nil {
			errorf(at.Value(), "fields: %v", err)
			continue
		}
		cfg.Fields[name] = ref
	}
	return cfg, fd.Diags
}

// errorFunc reports an error at a position of the document.
type errorFunc func(at tomlcheck.Pos, format string, args ...any)

// more checks the key more at the TOML path key: its text has the rules of a text (screens-toml, section 3.1).
// at is the place of the key. It gives the text, else nil without the key or after an error.
func more(p *string, at tomlcheck.Place, key string, errorf errorFunc) *string {
	if p == nil {
		return nil
	}
	if err := screen.CheckText(*p); err != nil {
		errorf(at.Value(), "%s: %v", key, err)
		return nil
	}
	return p
}

// langs sets the tables [lang.<code>] and the languages of the BBS: [screen] lang
// and the codes of the tables, sorted (screens-toml, section 3.5). A doc without tables
// gives no errors, so errorf can be nil. root is the root of the document.
func langs(cfg *Config, d screensDoc, root tomlcheck.Place, errorf errorFunc) {
	cfg.LangMore = map[string]*string{}
	codes := map[string]bool{cfg.Lang: true}
	for _, code := range slices.Sorted(maps.Keys(d.Lang)) {
		at := root.At("lang", code)
		if !screen.IsLang(code) {
			errorf(at.Key(), "lang: invalid language code %q", code)
			continue
		}
		cfg.LangMore[code] = more(d.Lang[code].More, at.At("more"), "lang."+code+".more", errorf)
		codes[code] = true
	}
	cfg.Langs = slices.Sorted(maps.Keys(codes))
}

// theme sets the base theme, and the dark theme and the light theme of each term.colors of d
// (screens-toml, section 3.2). A doc without themes gives no errors, so errorf can be nil.
// root is the root of the document.
func theme(cfg *Config, d screensDoc, root tomlcheck.Place, errorf errorFunc) {
	base := baseValues()
	// names is the names of the theme. A name with an invalid style stays a name:
	// one key gives one error.
	names := map[string]bool{}
	for name := range base {
		names[name] = true
	}
	th := root.At("theme")
	styles(d.Theme.Default, th, "default", names, false, base, errorf)
	// table parses a depth theme or a light table once: one key gives one error.
	table := func(t map[string]string, key string) map[string]value {
		vs := map[string]value{}
		styles(t, th, key, names, true, vs, errorf)
		return vs
	}
	depth := map[int]map[string]value{
		screen.Colors16: table(d.Theme.C16, "16"), screen.Colors256: table(d.Theme.C256, "256"),
	}
	lightDefault := table(d.Theme.Light.Default, "light.default")
	light := map[int]map[string]value{
		screen.Colors16: table(d.Theme.Light.C16, "light.16"), screen.Colors256: table(d.Theme.Light.C256, "light.256"),
	}
	// cycle reports each key of a cycle once, with the path of the first theme
	// in the order of the loops below.
	reported := map[string]bool{}
	cycle := func(v value, path []string) {
		if !reported[v.key] {
			reported[v.key] = true
			errorf(v.at, "%s: cycle: %s", v.key, strings.Join(path, " -> "))
		}
	}
	all := []int{screen.Colors16, screen.Colors256, screen.ColorsTrue}
	cfg.dark, cfg.light = map[int]map[string]screen.Style{}, map[int]map[string]screen.Style{}
	darkValues := map[int]map[string]value{}
	for _, colors := range all {
		vs := maps.Clone(base)
		maps.Copy(vs, depth[colors])
		darkValues[colors] = vs
		cfg.dark[colors] = resolve(refs(vs, cycle), defaultNormal)
	}
	for _, colors := range all {
		vs := maps.Clone(darkValues[colors])
		maps.Copy(vs, lightDefault)
		maps.Copy(vs, light[colors])
		cfg.light[colors] = resolve(refs(vs, cycle), lightNormal)
	}
	// There is no depth theme of 16777216 colours, so its dark theme is the base theme.
	cfg.Theme = cfg.dark[screen.ColorsTrue]
}

// value is a value of a theme table: a style or a reference (screens-toml, section 3.2).
type value struct {
	// spec is the style. For a reference, it has only the attributes of the reference.
	spec screen.StyleSpec
	// ref is the name of a reference, else "".
	ref string
	// key is the TOML path of the value, for example "theme.16.md_h1", and at is the position
	// of the value. A default of the base theme has no key.
	key string
	at  tomlcheck.Pos
}

// parseValue parses a style or a reference. A value whose first word matches the name syntax
// is a reference (screens-toml, section 3.2).
func parseValue(text string) (value, error) {
	name, attrs, hasAttrs := strings.Cut(text, " ")
	if !screen.IsIdent(name) {
		spec, err := screen.ParseStyle(text)
		return value{spec: spec}, err
	}
	v := value{ref: name}
	if hasAttrs {
		var err error
		if v.spec.Attr, err = screen.ParseAttrs(attrs); err != nil {
			return value{}, err
		}
	}
	return v, nil
}

// styles parses the values of the table theme.<key> into vs. th is the place of the table theme.
// With known, each name must be in names. Else styles adds each valid name to names.
// After the names of the table, the name of each reference must be in names.
func styles(t map[string]string, th tomlcheck.Place, key string, names map[string]bool, known bool,
	vs map[string]value, errorf errorFunc) {
	table := th.At(strings.Split(key, ".")...)
	withRef := map[string]value{}
	for _, n := range slices.Sorted(maps.Keys(t)) {
		at := table.At(n)
		switch {
		case !screen.IsIdent(n):
			errorf(at.Key(), "theme.%s: invalid style name %q", key, n)
			continue
		case !known:
			names[n] = true
		case !names[n]:
			errorf(at.Key(), "theme.%s: unknown style %q", key, n)
			continue
		}
		s := t[n]
		v, err := parseValue(s)
		if err != nil {
			errorf(at.Value(), "theme.%s.%s: invalid style %q: %v", key, n, s, err)
			continue
		}
		v.key, v.at = "theme."+key+"."+n, at.Value()
		if v.ref != "" {
			withRef[n] = v
			continue
		}
		vs[n] = v
	}
	// A reference with an unknown name keeps the value before it, as an invalid style does.
	for _, n := range slices.Sorted(maps.Keys(withRef)) {
		v := withRef[n]
		if !names[v.ref] {
			errorf(v.at, "%s: unknown style %q", v.key, v.ref)
			continue
		}
		vs[n] = v
	}
}

// refs gives the styles of the values vs of one theme with each reference resolved
// (screens-toml, section 3.2): a reference gives the style of its name with the attributes
// of the reference added. cycle gets each reference that leads back to its own style,
// and the path from its name to its name.
// A style in a cycle, a reference that leads into a cycle and a reference to a name
// with an invalid style get only the attributes on the path: the colours of normal.
func refs(vs map[string]value, cycle func(v value, path []string)) map[string]screen.StyleSpec {
	specs := make(map[string]screen.StyleSpec, len(vs))
	for _, name := range slices.Sorted(maps.Keys(vs)) {
		v := vs[name]
		path := []string{name}
		var attr screen.Attr
		for v.ref != "" {
			attr |= v.spec.Attr
			if i := slices.Index(path, v.ref); i >= 0 {
				if i == 0 {
					cycle(vs[name], append(path, name))
				}
				v = value{}
				break
			}
			path = append(path, v.ref)
			v = vs[v.ref]
		}
		spec := v.spec
		spec.Attr |= attr
		specs[name] = spec
	}
	return specs
}

// mdStyles is the md_* styles of the base theme (file-view, section 4).
var mdStyles = map[string]string{
	"md_h1": "15 bold", "md_h2": "15 bold", "md_h3": "15", "md_strong": "15 bold", "md_em": "14",
	"md_code": "10", "md_link": "14 underline", "md_mark": "11", "md_frame": "8", "md_dim": "8",
}

// baseValues gives the defaults of the base theme: normal, more, select and the md_* styles.
func baseValues() map[string]value {
	vs := map[string]value{
		"normal": {spec: defaultNormalStyle}, "more": {spec: defaultMoreStyle}, "select": {spec: defaultSelectStyle},
	}
	for name, text := range mdStyles {
		spec, err := screen.ParseStyle(text)
		if err != nil {
			panic("config: md style " + name + ": " + err.Error())
		}
		vs[name] = value{spec: spec}
	}
	return vs
}

// resolve gives the theme of the styles specs. A style without a colour takes it from normal;
// normal takes it from def: 7/0 in the dark theme, 0/15 in the light theme (screens-toml, section 3.2).
func resolve(specs map[string]screen.StyleSpec, def screen.Style) map[string]screen.Style {
	normal := specs["normal"].Over(def)
	base := screen.Style{Fg: normal.Fg, Bg: normal.Bg}
	th := make(map[string]screen.Style, len(specs))
	for name, spec := range specs {
		th[name] = spec.Over(base)
	}
	return th
}
