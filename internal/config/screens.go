package config

import (
	"bytes"
	"maps"
	"slices"
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
		ThemeChoice: ThemeDark, Telnet: defaultTelnet, Limits: defaultLimits, Name: defaultName,
	}
	fd := screen.FileDiags{File: ScreensFile}
	var d screensDoc
	errs := tomlcheck.Decode(bytes.TrimPrefix(data, screen.BOM), &d)
	for _, e := range errs {
		fd.Errorf(e.Line, e.Col, "%s", e.Msg)
	}
	if len(errs) > 0 {
		theme(&cfg, screensDoc{}, nil)
		langs(&cfg, screensDoc{}, nil)
		return cfg, fd.Diags
	}
	// Each other error has the position 1:1 (screen-render, section 3.1).
	errorf := func(format string, args ...any) { fd.Errorf(1, 1, format, args...) }
	cfg.Layout = screen.ParseLayoutKeys(d.Screen.Width, d.Screen.Align, d.Screen.VAlign, d.Screen.Fill,
		func(key string, err error) { errorf("screen.%s: %v", key, err) }).Over(defaultLayout)
	cfg.More = more(d.Screen.More, "screen.more", errorf)
	if p := d.Screen.Time; p != nil {
		cfg.Time = *p
	}
	if p := d.Screen.Zone; p != nil {
		// LoadLocation gives UTC for an empty name: the key wants a name.
		if loc, err := time.LoadLocation(*p); err != nil || *p == "" {
			errorf("screen.zone: unknown time zone %q", *p)
		} else {
			cfg.Zone = loc
		}
	}
	if p := d.Screen.Theme; p != nil {
		switch c := ThemeChoice(*p); c {
		case ThemeDark, ThemeLight, ThemeAuto:
			cfg.ThemeChoice = c
		default:
			errorf("screen.theme: invalid value %q, want \"dark\", \"light\" or \"auto\"", *p)
		}
	}
	if p := d.Screen.Lang; p != nil {
		if !screen.IsLang(*p) {
			errorf("screen.lang: invalid language code %q", *p)
		} else {
			cfg.Lang = *p
		}
	}
	langs(&cfg, d, errorf)
	if p := d.Menu.Start; p != nil {
		if !screen.IsName(*p) {
			errorf("menu.start: invalid screen name %q", *p)
		} else {
			cfg.Start = *p
		}
	}
	theme(&cfg, d, errorf)
	for _, name := range slices.Sorted(maps.Keys(d.Fields)) {
		ref, err := screen.NewAlias(name, d.Fields[name])
		if err != nil {
			errorf("fields: %v", err)
			continue
		}
		cfg.Fields[name] = ref
	}
	return cfg, fd.Diags
}

// more checks the key more at the TOML path key: its text has the rules of a text (screens-toml, section 3.1).
// It gives the text, else nil without the key or after an error.
func more(p *string, key string, errorf func(format string, args ...any)) *string {
	if p == nil {
		return nil
	}
	if err := screen.CheckText(*p); err != nil {
		errorf("%s: %v", key, err)
		return nil
	}
	return p
}

// langs sets the tables [lang.<code>] and the languages of the BBS: [screen] lang
// and the codes of the tables, sorted (screens-toml, section 3.5). A doc without tables
// gives no errors, so errorf can be nil.
func langs(cfg *Config, d screensDoc, errorf func(format string, args ...any)) {
	cfg.LangMore = map[string]*string{}
	codes := map[string]bool{cfg.Lang: true}
	for _, code := range slices.Sorted(maps.Keys(d.Lang)) {
		if !screen.IsLang(code) {
			errorf("lang: invalid language code %q", code)
			continue
		}
		cfg.LangMore[code] = more(d.Lang[code].More, "lang."+code+".more", errorf)
		codes[code] = true
	}
	cfg.Langs = slices.Sorted(maps.Keys(codes))
}

// theme sets the base theme, and the dark theme and the light theme of each term.colors of d
// (screens-toml, section 3.2). A doc without themes gives no errors, so errorf can be nil.
func theme(cfg *Config, d screensDoc, errorf func(format string, args ...any)) {
	base := baseSpecs()
	// names is the names of the theme. A name with an invalid style stays a name:
	// one key gives one error.
	names := map[string]bool{}
	for name := range base {
		names[name] = true
	}
	styles(d.Theme.Default, "default", names, false, base, errorf)
	cfg.Theme = resolve(base, defaultNormal)
	// table parses a depth theme or a light table once: one key gives one error.
	table := func(t map[string]string, key string) map[string]screen.StyleSpec {
		specs := map[string]screen.StyleSpec{}
		styles(t, key, names, true, specs, errorf)
		return specs
	}
	depth := map[int]map[string]screen.StyleSpec{
		screen.Colors16: table(d.Theme.C16, "16"), screen.Colors256: table(d.Theme.C256, "256"),
	}
	lightDefault := table(d.Theme.Light.Default, "light.default")
	light := map[int]map[string]screen.StyleSpec{
		screen.Colors16: table(d.Theme.Light.C16, "light.16"), screen.Colors256: table(d.Theme.Light.C256, "light.256"),
	}
	cfg.dark, cfg.light = map[int]map[string]screen.Style{}, map[int]map[string]screen.Style{}
	for _, colors := range []int{screen.Colors16, screen.Colors256, screen.ColorsTrue} {
		specs := maps.Clone(base)
		maps.Copy(specs, depth[colors])
		cfg.dark[colors] = resolve(specs, defaultNormal)
		maps.Copy(specs, lightDefault)
		maps.Copy(specs, light[colors])
		cfg.light[colors] = resolve(specs, lightNormal)
	}
}

// styles parses the styles of the table theme.<key> into specs.
// With known, each name must be in names. Else styles adds each valid name to names.
func styles(t map[string]string, key string, names map[string]bool, known bool,
	specs map[string]screen.StyleSpec, errorf func(format string, args ...any)) {
	for _, n := range slices.Sorted(maps.Keys(t)) {
		switch {
		case !screen.IsIdent(n):
			errorf("theme.%s: invalid style name %q", key, n)
			continue
		case !known:
			names[n] = true
		case !names[n]:
			errorf("theme.%s: unknown style %q", key, n)
			continue
		}
		s := t[n]
		spec, err := screen.ParseStyle(s)
		if err != nil {
			errorf("theme.%s.%s: invalid style %q: %v", key, n, s, err)
			continue
		}
		specs[n] = spec
	}
}

// mdStyles is the md_* styles of the base theme (file-view, section 4).
var mdStyles = map[string]string{
	"md_h1": "15 bold", "md_h2": "15 bold", "md_h3": "15", "md_strong": "15 bold", "md_em": "14",
	"md_code": "10", "md_link": "14 underline", "md_mark": "11", "md_frame": "8", "md_dim": "8",
}

// baseSpecs gives the defaults of the base theme: normal, more, select and the md_* styles.
func baseSpecs() map[string]screen.StyleSpec {
	specs := map[string]screen.StyleSpec{
		"normal": defaultNormalStyle, "more": defaultMoreStyle, "select": defaultSelectStyle,
	}
	for name, text := range mdStyles {
		spec, err := screen.ParseStyle(text)
		if err != nil {
			panic("config: md style " + name + ": " + err.Error())
		}
		specs[name] = spec
	}
	return specs
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
