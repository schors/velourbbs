// Package config reads the host settings of bbs.toml.
// Specification: docs/specs/bbs-toml.md.
package config

import (
	"bytes"
	"maps"
	"slices"

	"github.com/schors/velourbbs/internal/screen"
	"github.com/schors/velourbbs/internal/tomlcheck"
)

// Config is the host settings.
type Config struct {
	// Width is the default width limits of a screen (bbs-toml, section 3.1).
	Width screen.WidthLimits
	// More is the text of the More row (bbs-toml, section 3.1).
	More string
	// Theme maps a style name to a style (bbs-toml, section 3.2). It always has normal and more.
	Theme map[string]screen.Style
	// Fields maps an alias to its field (bbs-toml, section 3.3).
	Fields screen.Fields
	// Start is the name of the first screen of a session (bbs-toml, section 3.4).
	Start string
	// Telnet is the table [telnet] (bbs-toml, section 3.5).
	Telnet Telnet
}

// The defaults of bbs-toml, sections 3.1 and 3.2.
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
	Theme  map[string]string `toml:"theme"`
	Fields map[string]string `toml:"fields"`
	Menu   struct {
		Start *string `toml:"start"`
	} `toml:"menu"`
	Telnet telnetDoc `toml:"telnet"`
}

// File is the name of the host settings file in the BBS directory and in the diagnostics.
const File = "bbs.toml"

// Parse reads bbs.toml from data.
// An unknown key is an error (bbs-toml, section 2).
// It gives at most screen.MaxFileDiags diagnostics (screen-render, section 3.1).
// Parse returns the defaults for each value with an error.
func Parse(data []byte) (Config, screen.Diags) {
	cfg := Config{
		Width: defaultWidth, More: defaultMore, Fields: screen.Fields{}, Start: "main", Telnet: defaultTelnet,
	}
	fd := screen.FileDiags{File: File}
	var d doc
	errs := tomlcheck.Decode(bytes.TrimPrefix(data, screen.BOM), &d)
	for _, e := range errs {
		fd.Errorf(e.Line, e.Col, "%s", e.Msg)
	}
	if len(errs) > 0 {
		cfg.Theme = theme(nil)
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
		cfg.Start = *p
	}
	cfg.Telnet = telnet(d.Telnet, errorf)
	specs := map[string]screen.StyleSpec{}
	for _, name := range slices.Sorted(maps.Keys(d.Theme)) {
		if !screen.IsIdent(name) {
			errorf("theme: invalid style name %q", name)
			continue
		}
		s := d.Theme[name]
		spec, err := screen.ParseStyle(s)
		if err != nil {
			errorf("theme.%s: invalid style %q: %v", name, s, err)
			continue
		}
		specs[name] = spec
	}
	cfg.Theme = theme(specs)
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

// theme gives the theme of the styles t. A style without a colour takes it from normal;
// normal takes it from its default (bbs-toml, section 3.2).
func theme(t map[string]screen.StyleSpec) map[string]screen.Style {
	specs := map[string]screen.StyleSpec{"normal": {}, "more": defaultMoreStyle}
	maps.Copy(specs, t)
	normal := specs["normal"].Over(defaultNormal)
	base := screen.Style{Fg: normal.Fg, Bg: normal.Bg}
	th := make(map[string]screen.Style, len(specs))
	for name, spec := range specs {
		th[name] = spec.Over(base)
	}
	return th
}
