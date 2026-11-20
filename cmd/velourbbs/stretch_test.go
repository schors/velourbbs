package main

import (
	"bytes"
	"path/filepath"
	"testing"
	"testing/fstest"
)

// The stretch tokens, the spacer line and the gutter of bbt-format, sections 6.1 and 6.6:
// velourbbs check gives no output, and the preview at 80x24 and 132x43 equals its golden dump.
func TestStretch(t *testing.T) {
	out, errOut, status := runIn(t, "stretch", "check")
	if out != "" || errOut != "" || status != 0 {
		t.Errorf("check: stdout %q, stderr %q, status %d", out, errOut, status)
	}
	for _, size := range []string{"80x24", "132x43"} {
		out, errOut, status := runIn(t, "stretch", "preview", "main", "--size", size)
		if errOut != "" || status != 0 {
			t.Errorf("%s: stderr %q, status %d", size, errOut, status)
		}
		golden(t, filepath.Join("testdata", "stretch", "main-"+size+".dump"), out)
	}
}

// velourbbs check gives the errors of a wrong stretch token and of a gutter out of 0 to 100.
// An error of screens/screens.toml stops the check before the screens (screen-render, section 3.1),
// so each case is one run.
func TestCheckStretch(t *testing.T) {
	tests := []struct{ config, want string }{
		{"[screen]\ngutter = 101\n", `screens/screens.toml:2:10: screen.gutter: 101, must be from 0 to 100
`},
		{"", `screens/about.bbt:1:3: invalid stretch token "{~..}", want {~} or {~<c>} with one character of width 1
screens/main.bbt:2:1: gutter: -1, must be from 0 to 100
`},
	}
	for _, tt := range tests {
		fsys := fstest.MapFS{
			"bbs.toml":             {},
			"screens/screens.toml": {Data: []byte(tt.config)},
			"screens/main.bbt":     {Data: []byte("+++\ngutter = -1\n+++\n")},
			"screens/about.bbt":    {Data: []byte("a {~..} b\n")},
		}
		var out, errOut bytes.Buffer
		status := run([]string{"check"}, t.TempDir(), fsys, &out, &errOut)
		if out.Len() != 0 || errOut.String() != tt.want || status != exitError {
			t.Errorf("%q: stdout %q, stderr\n%s\nstatus %d; want stderr\n%s", tt.config, out.String(), errOut.String(), status, tt.want)
		}
	}
}
