package main

import (
	"path/filepath"
	"testing"
)

// TestPreviewView: the golden line form of velourbbs preview --view (screen-render, section 3.7)
// on the test BBS testdata/file. gold.md has more than one page at 40x10.
func TestPreviewView(t *testing.T) {
	goldens := []struct {
		name string
		args []string
	}{
		{"gold-40x10.view", []string{"reader", "--set", "arg.file=gold.md", "--size", "40x10", "--view"}},
		{"keys-40x10.view", []string{"main", "--size", "40x10", "--keys", "1,SPACE,ESC", "--view"}},
	}
	for _, g := range goldens {
		t.Run(g.name, func(t *testing.T) {
			out, errOut, status := runIn(t, "file", append([]string{"preview"}, g.args...)...)
			if errOut != "" || status != 0 {
				t.Errorf("stderr %q, status %d", errOut, status)
			}
			golden(t, filepath.Join("testdata", "file", g.name), out)
		})
	}
}
