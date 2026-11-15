package config

import (
	"bytes"
	"maps"
	"slices"

	"github.com/schors/velourbbs/internal/screen"
	"github.com/schors/velourbbs/internal/tomlcheck"
)

// The defaults of screens-toml, sections 3.1 and 3.2.
var (
	defaultWidth     = screen.WidthLimits{Min: 40, Max: 80}
	defaultMore      = "-- More --"
	defaultNormal    = screen.Style{Fg: screen.IndexColour(7), Bg: screen.IndexColour(0)}
	defaultMoreStyle = screen.StyleSpec{
		Fg: screen.IndexColour(15), Bg: screen.IndexColour(0), HasFg: true, HasBg: true,
	}
)

// screensDoc is the plain decode of screens/screens.toml. A nil pointer is a key that is not set.
type screensDoc struct {
	Screen struct {
		Width *[]int64 `toml:"width"`
		More  *string  `toml:"more"`
	} `toml:"screen"`
	Theme struct {
		Default map[string]string `toml:"default"`
		C16     map[string]string `toml:"16"`
		C256    map[string]string `toml:"256"`
	} `toml:"theme"`
	Fields map[string]string `toml:"fields"`
	Menu   struct {
		Start *string `toml:"start"`
	} `toml:"menu"`
}

// parseScreens reads screens/screens.toml from data. The Telnet of the result is the default.
// An unknown key is an error (screens-toml, section 2).
// It gives at most screen.MaxFileDiags diagnostics (screen-render, section 3.1).
func parseScreens(data []byte) (Config, screen.Diags) {
	cfg := Config{
		Width: defaultWidth, More: defaultMore, Fields: screen.Fields{}, Start: "main", Telnet: defaultTelnet,
	}
	fd := screen.FileDiags{File: ScreensFile}
	var d screensDoc
	errs := tomlcheck.Decode(bytes.TrimPrefix(data, screen.BOM), &d)
	for _, e := range errs {
		fd.Errorf(e.Line, e.Col, "%s", e.Msg)
	}
	if len(errs) > 0 {
		cfg.Theme = resolve(baseSpecs())
		return cfg, fd.Diags
	}
	// Each other error has the position 1:1 (screen-render, section 3.1).
	errorf := func(format string, args ...any) { fd.Errorf(1, 1, format, args...) }
	if p := d.Screen.Width; p != nil {
		w, err := screen.NewWidth(*p)
		if err != nil {
			errorf("screen.width: %v", err)
		} else {
			cfg.Width = w
		}
	}
	if p := d.Screen.More; p != nil {
		if err := checkText(*p); err != nil {
			errorf("screen.more: %v", err)
		} else {
			cfg.More = *p
		}
	}
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

func checkText(s string) error {
	for g, w := range screen.Graphemes(s) {
		if err := screen.CheckGrapheme(g, w); err != nil {
			return err
		}
	}
	return nil
}

// theme sets the base theme and the session theme of each depth theme of d
// (screens-toml, section 3.2).
func theme(cfg *Config, d screensDoc, errorf func(format string, args ...any)) {
	base := baseSpecs()
	// names is the names of the theme. A name with an invalid style stays a name:
	// one key gives one error.
	names := map[string]bool{}
	for name := range base {
		names[name] = true
	}
	styles(d.Theme.Default, "default", names, false, base, errorf)
	cfg.Theme = resolve(base)
	depths := []struct {
		colors int
		key    string
		styles map[string]string
	}{
		{16, "16", d.Theme.C16},
		{256, "256", d.Theme.C256},
	}
	for _, dt := range depths {
		if dt.styles == nil {
			continue
		}
		specs := maps.Clone(base)
		styles(dt.styles, dt.key, names, true, specs, errorf)
		if cfg.depth == nil {
			cfg.depth = map[int]map[string]screen.Style{}
		}
		cfg.depth[dt.colors] = resolve(specs)
	}
}

// styles parses the styles of the table theme.<key> into specs.
// With depth, each name must be in names. Else styles adds each valid name to names.
func styles(t map[string]string, key string, names map[string]bool, depth bool,
	specs map[string]screen.StyleSpec, errorf func(format string, args ...any)) {
	for _, n := range slices.Sorted(maps.Keys(t)) {
		switch {
		case !screen.IsIdent(n):
			errorf("theme.%s: invalid style name %q", key, n)
			continue
		case !depth:
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

// baseSpecs gives the defaults of the base theme: normal and more.
func baseSpecs() map[string]screen.StyleSpec {
	return map[string]screen.StyleSpec{"normal": {}, "more": defaultMoreStyle}
}

// resolve gives the theme of the styles specs. A style without a colour takes it from normal;
// normal takes it from 7/0 (screens-toml, section 3.2).
func resolve(specs map[string]screen.StyleSpec) map[string]screen.Style {
	normal := specs["normal"].Over(defaultNormal)
	base := screen.Style{Fg: normal.Fg, Bg: normal.Bg}
	th := make(map[string]screen.Style, len(specs))
	for name, spec := range specs {
		th[name] = spec.Over(base)
	}
	return th
}
