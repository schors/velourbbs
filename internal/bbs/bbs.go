// Package bbs loads the BBS directory into the model and checks the menus between screens.
// Specification: docs/specs/session.md, section 3, docs/specs/menu-nav.md, section 6,
// and docs/specs/session-language.md, section 6.
package bbs

import (
	"io/fs"

	"github.com/schors/velourbbs/internal/bbt"
	"github.com/schors/velourbbs/internal/config"
	"github.com/schors/velourbbs/internal/screen"
)

// Model is the result of a load without errors (session, section 2).
// It does not change after the load: sessions read it at the same time without locks.
type Model struct {
	Config config.Config // the settings of bbs.toml and screens/screens.toml
	// Sets is the screen set of each language of the BBS (session-language, section 4).
	Sets bbt.Sets
}

// Load reads bbs.toml, screens/screens.toml and the screens of the BBS directory fsys
// and does the checks between files (session, section 3). It gives the model
// and the sorted diagnostics. The model is nil when the diagnostics have an error.
func Load(fsys fs.FS) (*Model, screen.Diags) {
	// config.Load also gives the defaults after an error, so the screens load with them.
	cfg, ds := config.Load(fsys)
	sets, sds := bbt.Load(fsys, Settings(cfg))
	ds = append(ds, sds...)
	ds.Sort()
	if ds.HasErrors() {
		return nil, ds
	}
	// The checks between files need all screens (menu-nav, section 6).
	ds = Check(cfg, sets, ds)
	if ds.HasErrors() {
		return nil, ds
	}
	return &Model{Config: cfg, Sets: sets}, ds
}

// Settings gives the settings of the load of the screens from cfg.
func Settings(cfg config.Config) bbt.Settings {
	return bbt.Settings{
		Theme: cfg.Theme, Layout: cfg.Layout, Fields: cfg.Fields, Langs: cfg.Langs, LangTables: len(cfg.LangMore) > 0,
	}
}
