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
		if m != nil && (m.Screens["main"] == nil || m.Config.Start != "main") {
			t.Errorf("%s: model %+v, want the screen main and the start main", tt.name, m)
		}
	}
}
