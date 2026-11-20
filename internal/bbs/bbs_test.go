package bbs

import (
	"testing"
	"testing/fstest"
)

// Load gives the model after a load without errors, and no model after an error
// of a file or of the checks between files (session, section 3).
func TestLoad(t *testing.T) {
	main := "+++\n[[item]]\nkey = \"G\"\ndo = \"logoff\"\n+++\nMain\n"
	tests := []struct {
		name  string
		files map[string]string
		model bool
	}{
		{"valid", map[string]string{"bbs.toml": "", "screens/main.bbt": main}, true},
		{"file error", map[string]string{"bbs.toml": "name = 1\n", "screens/main.bbt": main}, false},
		{"check error", map[string]string{"bbs.toml": "", "screens/main.bbt": "Main\n"}, false},
	}
	for _, tt := range tests {
		fsys := fstest.MapFS{}
		for name, data := range tt.files {
			fsys[name] = &fstest.MapFile{Data: []byte(data)}
		}
		m, ds := Load(fsys)
		if got := m != nil; got != tt.model || got == ds.HasErrors() {
			t.Errorf("%s: model %v, diagnostics %v; want a model %v", tt.name, m, ds, tt.model)
			continue
		}
		if m != nil && (m.Sets["en"]["main"] == nil || m.Config.Start != "main") {
			t.Errorf("%s: model %+v, want the screen main and the start main", tt.name, m)
		}
	}
}

// An error of bbs.toml or screens/screens.toml stops the load before the screens
// (screen-render, section 3.1): the screen errors and the checks between files do not run.
func TestLoadSettingsError(t *testing.T) {
	fsys := fstest.MapFS{
		"bbs.toml":             {Data: []byte("[bbs\n")},
		"screens/screens.toml": {Data: []byte("[theme.default]\ngold = #b8902f\n")},
		"screens/main.bbt":     {Data: []byte("{!gold}Main{x\n")},
	}
	m, ds := Load(fsys)
	if m != nil || len(ds) != 2 || ds[0].File != "bbs.toml" || ds[1].File != "screens/screens.toml" {
		t.Errorf("model %v, diagnostics %v; want no model and one error of each settings file", m, ds)
	}
}
