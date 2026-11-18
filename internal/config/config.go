// Package config reads the host settings of bbs.toml and the screens settings
// of screens/screens.toml.
// Specification: docs/specs/bbs-toml.md and docs/specs/screens-toml.md.
package config

import (
	"bytes"
	"errors"
	"io/fs"

	"github.com/schors/velourbbs/internal/screen"
	"github.com/schors/velourbbs/internal/term"
	"github.com/schors/velourbbs/internal/tomlcheck"
)

// Config is the settings of the BBS directory.
type Config struct {
	// Layout is the default layout of a screen: the keys width, align, valign
	// and fill of [screen] (screens-toml, section 3.1).
	screen.Layout
	// More is the text of the More row (screens-toml, section 3.1).
	More string
	// Theme is the base theme (screens-toml, section 3.2). It always has normal and more.
	Theme map[string]screen.Style
	// ThemeChoice is the theme choice at the start of a session: the key theme of [screen]
	// (screens-toml, section 3.1).
	ThemeChoice ThemeChoice
	// Fields maps an alias to its field (screens-toml, section 3.3).
	Fields screen.Fields
	// Time is the layout of the field bbs.time: the key time of [screen] (screens-toml, section 3.1).
	Time string
	// Start is the name of the first screen of a session (screens-toml, section 3.4).
	Start string
	// Telnet is the table [telnet] (bbs-toml, section 3.1).
	Telnet Telnet
	// Name is the name of the BBS: the key name of [bbs] (bbs-toml, section 3.2).
	Name string

	// dark and light map each term.colors to its dark theme and its light theme.
	dark, light map[int]map[string]screen.Style
}

// ThemeChoice is a theme choice of a session (screens-toml, section 3.2).
type ThemeChoice string

// The theme choices (screens-toml, section 3.2).
const (
	ThemeDark  ThemeChoice = "dark"
	ThemeLight ThemeChoice = "light"
	ThemeAuto  ThemeChoice = "auto"
)

// SessionTheme gives the session theme of the theme choice choice and the terminal t
// (screens-toml, section 3.2): the light theme of t.Colors for light, and for auto
// with t.Background light. Else it gives the dark theme of t.Colors.
func (c Config) SessionTheme(choice ThemeChoice, t screen.Term) map[string]screen.Style {
	if choice == ThemeLight || choice == ThemeAuto && t.Background == term.BackgroundLight {
		return c.light[t.Colors]
	}
	return c.dark[t.Colors]
}

// The names of the settings files in the BBS directory and in the diagnostics.
const (
	File        = "bbs.toml"
	ScreensFile = "screens/screens.toml"
)

// defaultName is the default of [bbs] name (bbs-toml, section 3.2).
const defaultName = "Kind Brothers BBS"

// hostDoc is the plain decode of bbs.toml. A nil pointer is a key that is not set.
type hostDoc struct {
	BBS struct {
		Name *string `toml:"name"`
	} `toml:"bbs"`
	Telnet telnetDoc `toml:"telnet"`
}

// Load reads bbs.toml and screens/screens.toml of the BBS directory fsys.
// A missing bbs.toml is an error. A missing screens/screens.toml gives the defaults
// (screens-toml, section 2). Load gives the defaults for each value with an error.
func Load(fsys fs.FS) (Config, screen.Diags) {
	var ds screen.Diags
	host, err := screen.ReadFile(fsys, File)
	if err != nil {
		ds.Errorf(File, 1, 1, "%v", err)
	}
	data, err := screen.ReadFile(fsys, ScreensFile)
	if errors.Is(err, fs.ErrNotExist) {
		// A dangling symlink exists and is not a regular file (screens-toml, section 2).
		if _, lerr := fs.Lstat(fsys, ScreensFile); lerr == nil {
			ds.Errorf(ScreensFile, 1, 1, "not a regular file")
		}
	} else if err != nil {
		ds.Errorf(ScreensFile, 1, 1, "%v", err)
	}
	cfg, sds := parseScreens(data)
	var hds screen.Diags
	cfg.Telnet, cfg.Name, hds = parseHost(host)
	ds = append(ds, sds...)
	ds = append(ds, hds...)
	return cfg, ds
}

// parseHost reads bbs.toml from data: the table [telnet] and the BBS name.
// An unknown key is an error (bbs-toml, section 2).
// It gives at most screen.MaxFileDiags diagnostics (screen-render, section 3.1).
func parseHost(data []byte) (Telnet, string, screen.Diags) {
	fd := screen.FileDiags{File: File}
	var d hostDoc
	errs := tomlcheck.Decode(bytes.TrimPrefix(data, screen.BOM), &d)
	for _, e := range errs {
		fd.Errorf(e.Line, e.Col, "%s", e.Msg)
	}
	if len(errs) > 0 {
		return defaultTelnet, defaultName, fd.Diags
	}
	name := defaultName
	if d.BBS.Name != nil {
		name = *d.BBS.Name
	}
	// Each other error has the position 1:1 (screen-render, section 3.1).
	errorf := func(format string, args ...any) { fd.Errorf(1, 1, format, args...) }
	return telnet(d.Telnet, errorf), name, fd.Diags
}
