// Package config reads the host settings of bbs.toml and the screens settings
// of screens/screens.toml.
// Specification: docs/specs/bbs-toml.md and docs/specs/screens-toml.md.
package config

import (
	"bytes"
	"errors"
	"io/fs"

	"github.com/schors/velourbbs/internal/screen"
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
	// Fields maps an alias to its field (screens-toml, section 3.3).
	Fields screen.Fields
	// Start is the name of the first screen of a session (screens-toml, section 3.4).
	Start string
	// Telnet is the table [telnet] (bbs-toml, section 3.1).
	Telnet Telnet

	// depth maps a term.colors to its session theme, for each depth theme.
	depth map[int]map[string]screen.Style
}

// SessionTheme gives the session theme of colors, the term.colors of a terminal
// (screens-toml, section 3.2). Without a depth theme of colors, it is the base theme.
func (c Config) SessionTheme(colors int) map[string]screen.Style {
	if t, ok := c.depth[colors]; ok {
		return t
	}
	return c.Theme
}

// The names of the settings files in the BBS directory and in the diagnostics.
const (
	File        = "bbs.toml"
	ScreensFile = "screens/screens.toml"
)

// hostDoc is the plain decode of bbs.toml.
type hostDoc struct {
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
	telnet, hds := parseHost(host)
	cfg.Telnet = telnet
	ds = append(ds, sds...)
	ds = append(ds, hds...)
	return cfg, ds
}

// parseHost reads bbs.toml from data.
// An unknown key is an error (bbs-toml, section 2).
// It gives at most screen.MaxFileDiags diagnostics (screen-render, section 3.1).
func parseHost(data []byte) (Telnet, screen.Diags) {
	fd := screen.FileDiags{File: File}
	var d hostDoc
	errs := tomlcheck.Decode(bytes.TrimPrefix(data, screen.BOM), &d)
	for _, e := range errs {
		fd.Errorf(e.Line, e.Col, "%s", e.Msg)
	}
	if len(errs) > 0 {
		return defaultTelnet, fd.Diags
	}
	// Each other error has the position 1:1 (screen-render, section 3.1).
	errorf := func(format string, args ...any) { fd.Errorf(1, 1, format, args...) }
	return telnet(d.Telnet, errorf), fd.Diags
}
