package main

import (
	"bytes"
	"testing"
	"testing/fstest"
)

// The layout warnings of screen-render, section 3.6, in testdata/layout: each kind, each row of Width,
// a fragment, the suffix of a line that only the set ru gives, and a repeated size.
func TestCheckLayout(t *testing.T) {
	want := `screens/_head.bbt:1:1: warning: screen "main": at 40x10: the row is 43 cells wide, the screen shows 40
screens/_head.bbt:1:1: warning: screen "main": at 30x10: the row is 43 cells wide, the screen shows 30
screens/main.bbt:13:1: warning: at 40x10: the line wraps to 2 rows
screens/main.bbt:13:1: warning: at 30x10: the line wraps to 2 rows
screens/main.bbt:13:1: warning: at 30x10: the row is 36 cells wide, the screen shows 30
screens/main.bbt:14:1: warning: at 40x10: the line wraps to 2 rows (language "ru")
screens/main.bbt:14:1: warning: at 30x10: the line wraps to 2 rows (language "ru")
screens/main.bbt:14:1: warning: at 30x10: the row is 38 cells wide, the screen shows 30 (language "ru")
screens/main.bbt:15:1: warning: at 40x10: art "logo.ans" is 50 cells wide, the screen shows 40
screens/main.bbt:15:1: warning: at 30x10: art "logo.ans" is 50 cells wide, the screen shows 30
screens/main.bbt:16:1: warning: at 40x10: the row is 54 cells wide, the screen shows 40
screens/main.bbt:16:1: warning: at 30x10: the row is 54 cells wide, the screen shows 30
screens/parts.bbt:1:1: warning: at 40x10: 12 rows, the rows after 10 are cut
screens/parts.bbt:1:1: warning: at 30x10: 12 rows, the rows after 10 are cut
screens/parts.bbt:10:1: warning: at 40x10: the row is 54 cells wide, the screen shows 40
screens/parts.bbt:10:1: warning: at 30x10: the row is 54 cells wide, the screen shows 30
screens/read.bbt:8:1: warning: at 30x10: the row is 39 cells wide, the screen shows 30
screens/tall.bbt:1:1: warning: at 40x10: 12 rows, the screen gets a More row
screens/tall.bbt:1:1: warning: at 30x10: 12 rows, the screen gets a More row
`
	out, errOut, status := runIn(t, "layout", "check", "--size", "40x10", "--size", "30x10", "--size", "40x10")
	if out != "" || errOut != want || status != 0 {
		t.Errorf("stdout %q, stderr:\n%s\nstatus %d; want stderr:\n%s\nstatus 0", out, errOut, status, want)
	}
	// Without --size, check gives no layout warning.
	if out, errOut, status := runIn(t, "layout", "check"); out != "" || errOut != "" || status != 0 {
		t.Errorf("stdout %q, stderr %q, status %d; want no output and status 0", out, errOut, status)
	}
}

// After an error, check does not render (screen-render, section 3.6).
func TestCheckLayoutError(t *testing.T) {
	fsys := fstest.MapFS{
		"bbs.toml": {},
		"screens/main.bbt": {Data: []byte("+++\n[[item]]\nkey = \"G\"\ndo = \"logoff\"\n+++\n" +
			"A text line that wraps at forty columns of the screen, {!nostyle}twice.\n")},
	}
	want := "screens/main.bbt:6:56: unknown style \"nostyle\"\n"
	var out, errOut bytes.Buffer
	status := run([]string{"check", "--size", "40x10"}, t.TempDir(), fsys, &out, &errOut)
	if out.String() != "" || errOut.String() != want || status != exitError {
		t.Errorf("stdout %q, stderr:\n%s\nstatus %d; want stderr:\n%s\nstatus 1", out.String(), errOut.String(), status, want)
	}
}
