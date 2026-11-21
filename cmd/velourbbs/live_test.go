package main

import (
	"path/filepath"
	"testing"
)

// The live fields of live-screen, section 3: velourbbs check gives no output, and the preview
// equals its golden dump. Each slot has its largest width: bbs.time in the layout [screen] time
// and in a layout with MST of Europe/Berlin (CEST: 4 cells), session.online and bbs.uptime 14 cells,
// bbs.sessions the 3 digits of max_conns = 250. The dumps show a slot in a text row,
// in a row with stretch tokens, in an art and in a list line. At 40 columns a slot is at most W.
// With --set, bbs.time shows the value in each layout, cut to its slot.
func TestLive(t *testing.T) {
	out, errOut, status := runIn(t, "live", "check")
	if out != "" || errOut != "" || status != 0 {
		t.Errorf("check: stdout %q, stderr %q, status %d", out, errOut, status)
	}
	tests := []struct {
		file string
		args []string
	}{
		{"main-80x24.dump", []string{"main"}},
		{"main-40x24.dump", []string{"main", "--size", "40x24"}},
		{"main-set-80x24.dump", []string{"main", "--set", "bbs.time=Thursday, 1 January 2026, 12:00 noon"}},
		{"list-80x24.dump", []string{"list"}},
	}
	for _, tt := range tests {
		out, errOut, status := runIn(t, "live", append([]string{"preview"}, tt.args...)...)
		if errOut != "" || status != 0 {
			t.Errorf("%s: stderr %q, status %d", tt.file, errOut, status)
		}
		golden(t, filepath.Join("testdata", "live", tt.file), out)
	}
}
