package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// The test BBS testdata/chain passes the arguments dir and file of main through the screens hub
// and side to the file screen read and to the list screen parts (bbt-format, section 8.5).
// side also gives them back to hub: a cycle.
func TestChain(t *testing.T) {
	out, errOut, status := runIn(t, "chain", "check")
	if out != "" || errOut != "" || status != 0 {
		t.Errorf("check: stdout %q, stderr %q, status %d; want no output and status 0", out, errOut, status)
	}
	// The values on hub and side, the file of the value on read, the rows of the value on parts,
	// the file of a row on read, and the goto of the cycle.
	out, errOut, status = runIn(t, "chain", "preview", "main", "--size", "40x10", "--keys",
		"1,S,R,ESC,P,DOWN,ENTER,ESC,ESC,H,ESC,ESC,G")
	if errOut != "" || status != 0 {
		t.Errorf("preview --keys: stderr %q, status %d", errOut, status)
	}
	golden(t, filepath.Join("testdata", "chain", "keys-40x10.dump"), out)

	// The first action in name order is the {arg.<name>} of hub: its sample value is the value of main.
	out, errOut, status = runIn(t, "chain", "fields")
	if errOut != "" || status != 0 {
		t.Errorf("fields: stderr %q, status %d", errOut, status)
	}
	golden(t, filepath.Join("testdata", "chain", "fields.txt"), out)

	// Only the chain gives the arguments of side: the layout renders it with the sample values.
	want := "screens/side.bbt:19:1: warning: at 30x10: the row is 40 cells wide, the screen shows 30\n"
	if out, errOut, status := runIn(t, "chain", "check", "--size", "30x10"); out != "" || errOut != want || status != 0 {
		t.Errorf("check --size: stdout %q, stderr %q, status %d; want stderr %q", out, errOut, status, want)
	}
}

// A --set value reaches the file screen read and the list screen parts through the screen side
// (screen-render, section 3.2).
func TestChainPreviewSet(t *testing.T) {
	out, errOut, status := runIn(t, "chain", "preview", "hub", "--set", "arg.dir=docs", "--set", "arg.file=other.md",
		"--size", "40x10", "--keys", "S,R")
	if want := "|A file that no goto gives.              |"; !strings.Contains(out, want) || errOut != "" || status != 0 {
		t.Errorf("other.md: stderr %q, status %d, stdout:\n%s", errOut, status, out)
	}
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"arg.dir=docs", "arg.file=nope.md"}, "velourbbs preview: arg.file: content file \"nope.md\" does not exist\n"},
		{[]string{"arg.dir=nope", "arg.file=intro.md"}, "velourbbs preview: arg.dir: directory \"nope\" does not exist\n"},
	}
	for _, tt := range tests {
		out, errOut, status := runIn(t, "chain", "preview", "hub", "--set", tt.args[0], "--set", tt.args[1])
		if out != "" || errOut != tt.want || status != exitError {
			t.Errorf("%q: stdout %q, stderr %q, status %d; want stderr %q, status 1", tt.args, out, errOut, status, tt.want)
		}
	}
}
