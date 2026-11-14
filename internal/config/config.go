// Package config reads the host settings of bbs.toml: the tables [screen] and [theme]
// of docs/specs/screen-format.md, [menu] of docs/specs/menu-nav.md
// and [telnet] of docs/specs/telnet-term.md.
package config

import (
	"bytes"
	"fmt"
	"maps"
	"slices"

	"newbbs/internal/screen"
	"newbbs/internal/tomlcheck"
)

// Config is the host settings.
type Config struct {
	// Width is the default width limits of a screen (section 6.1).
	Width screen.WidthLimits
	// More is the text of the More row (section 6.4).
	More string
	// Theme maps a style name to a style (section 7.3). It always has normal and more.
	Theme map[string]screen.Style
	// Start is the name of the first screen of a session (menu-nav, section 6).
	Start string
	// Telnet is the table [telnet] (telnet-term, section 9.1).
	Telnet Telnet
}

// The defaults of sections 6.1, 6.4 and 7.3.
var (
	defaultWidth     = screen.WidthLimits{Min: 40, Max: 80}
	defaultMore      = "-- More --"
	defaultNormal    = screen.Style{Fg: screen.IndexColour(7), Bg: screen.IndexColour(0)}
	defaultMoreStyle = screen.StyleSpec{
		Fg: screen.IndexColour(15), Bg: screen.IndexColour(0), HasFg: true, HasBg: true,
	}
)

// doc is the plain decode of bbs.toml. A nil pointer is a key that is not set.
type doc struct {
	Screen struct {
		Width *[]int64 `toml:"width"`
		More  *string  `toml:"more"`
	} `toml:"screen"`
	Theme map[string]string `toml:"theme"`
	Menu  struct {
		Start *string `toml:"start"`
	} `toml:"menu"`
	Telnet telnetDoc `toml:"telnet"`
}

// values is the decode of the values with a check (screen-format, section 9.1).
type values struct {
	Theme  map[string]themeStyle `toml:"theme"`
	Telnet telnetValues          `toml:"telnet"`
}

// themeStyle is a style of [theme] (section 7.2).
type themeStyle struct{ spec screen.StyleSpec }

func (s *themeStyle) UnmarshalText(text []byte) error {
	spec, err := screen.ParseStyle(string(text))
	if err != nil {
		return fmt.Errorf("invalid style %q: %v", text, err)
	}
	s.spec = spec
	return nil
}

// File is the name of the host settings file in the BBS directory and in the diagnostics.
const File = "bbs.toml"

// Parse reads bbs.toml from data.
// An unknown key is an error, as in the front matter of a screen (section 3.2).
// It gives at most screen.MaxFileDiags diagnostics (section 10).
// Parse returns the defaults for each value with an error.
func Parse(data []byte) (Config, screen.Diags) {
	cfg := Config{
		Width: defaultWidth, More: defaultMore, Start: "main", Telnet: defaultTelnet,
	}
	fd := screen.FileDiags{File: File}
	var d doc
	var v values
	errs := tomlcheck.Decode(bytes.TrimPrefix(data, screen.BOM), &d, &v)
	for _, e := range errs {
		fd.Errorf(e.Line, e.Col, "%s", e.Msg)
	}
	if len(errs) > 0 {
		cfg.Theme = theme(nil)
		return cfg, fd.Diags
	}
	// Each other error has the position 1:1 (screen-format, section 9.1).
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
		cfg.Start = *p
	}
	cfg.Telnet = telnet(d.Telnet, v.Telnet, errorf)
	for _, name := range slices.Sorted(maps.Keys(v.Theme)) {
		if !screen.IsIdent(name) {
			errorf("theme: invalid style name %q", name)
			delete(v.Theme, name)
		}
	}
	cfg.Theme = theme(v.Theme)
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

// theme gives the theme of the styles t. A style without a colour takes it from normal;
// normal takes it from its default (section 7.3).
func theme(t map[string]themeStyle) map[string]screen.Style {
	specs := map[string]screen.StyleSpec{"normal": {}, "more": defaultMoreStyle}
	for name, s := range t {
		specs[name] = s.spec
	}
	normal := specs["normal"].Over(defaultNormal)
	base := screen.Style{Fg: normal.Fg, Bg: normal.Bg}
	th := make(map[string]screen.Style, len(specs))
	for name, spec := range specs {
		th[name] = spec.Over(base)
	}
	return th
}
