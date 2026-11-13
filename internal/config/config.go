// Package config reads the host settings of bbs.toml: the tables [screen] and [theme]
// of docs/specs/screen-format.md.
package config

import (
	"errors"

	"newbbs/internal/screen"
	"newbbs/internal/tomlpos"
)

// Config is the host settings.
type Config struct {
	// Width is the default width limits of a screen (section 6.1).
	Width screen.WidthLimits
	// More is the text of the More row (section 6.4).
	More string
	// Theme maps a style name to a style (section 7.3). It always has normal and more.
	Theme map[string]screen.Style
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

// Parse reads bbs.toml from data. file is the name in the diagnostics.
// An unknown key is an error, as in the front matter of a screen (section 3.2).
// Parse returns the defaults for each value with an error.
func Parse(file string, data []byte) (Config, screen.Diags) {
	p := parser{file: file, cfg: Config{Width: defaultWidth, More: defaultMore}}
	theme := map[string]any{}
	doc, err := tomlpos.Parse(data)
	if err != nil {
		var e *tomlpos.Error
		errors.As(err, &e)
		p.diags.Errorf(file, e.Pos.Line, e.Pos.Col, "%s", e.Msg)
		doc = &tomlpos.Doc{}
	}
	p.doc = doc
	for key, v := range doc.Values {
		switch key {
		case "screen":
			p.screen(v)
		case "theme":
			if t, ok := p.table(v, key); ok {
				theme = t
			}
		default:
			p.errorf([]string{key}, "unknown key %q", key)
		}
	}
	p.theme(theme)
	p.diags.Sort()
	return p.cfg, p.diags
}

type parser struct {
	file  string
	doc   *tomlpos.Doc
	cfg   Config
	diags screen.Diags
}

func (p *parser) errorf(path []string, format string, args ...any) {
	pos := p.doc.Pos(path...)
	p.diags.Errorf(p.file, pos.Line, pos.Col, format, args...)
}

func (p *parser) table(v any, name string) (map[string]any, bool) {
	t, ok := v.(map[string]any)
	if !ok {
		p.errorf([]string{name}, "%s: must be a table", name)
	}
	return t, ok
}

func (p *parser) screen(v any) {
	t, ok := p.table(v, "screen")
	if !ok {
		return
	}
	for key, v := range t {
		path := []string{"screen", key}
		switch key {
		case "width":
			w, err := screen.ParseWidth(v)
			if err != nil {
				p.errorf(path, "screen.width: %v", err)
				continue
			}
			p.cfg.Width = w
		case "more":
			s, ok := v.(string)
			if !ok {
				p.errorf(path, "screen.more: must be a string")
				continue
			}
			if err := checkText(s); err != nil {
				p.errorf(path, "screen.more: %v", err)
				continue
			}
			p.cfg.More = s
		default:
			p.errorf(path, "unknown key \"screen.%s\"", key)
		}
	}
}

func checkText(s string) error {
	for g, w := range screen.Graphemes(s) {
		if err := screen.CheckGrapheme(g, w); err != nil {
			return err
		}
	}
	return nil
}

// theme parses the styles. A style without a colour takes it from normal;
// normal takes it from its default (section 7.3).
func (p *parser) theme(t map[string]any) {
	specs := map[string]screen.StyleSpec{"normal": {}, "more": defaultMoreStyle}
	for name, v := range t {
		path := []string{"theme", name}
		if !screen.IsIdent(name) {
			p.errorf(path, "theme: invalid style name %q", name)
			continue
		}
		text, ok := v.(string)
		if !ok {
			p.errorf(path, "theme.%s: must be a string", name)
			continue
		}
		spec, err := screen.ParseStyle(text)
		if err != nil {
			p.errorf(path, "theme.%s: invalid style %q: %v", name, text, err)
			continue
		}
		specs[name] = spec
	}
	normal := specs["normal"].Over(defaultNormal)
	base := screen.Style{Fg: normal.Fg, Bg: normal.Bg}
	p.cfg.Theme = make(map[string]screen.Style, len(specs))
	for name, spec := range specs {
		p.cfg.Theme[name] = spec.Over(base)
	}
}
