package main

import (
	"path/filepath"
	"testing"
)

// The test BBS testdata/ask asks for a line on a plain screen, a file screen, a list screen
// and a screen without rows (menu-nav, sections 5.4 and 7).
func TestAsk(t *testing.T) {
	out, errOut, status := runIn(t, "ask", "check")
	if out != "" || errOut != "" || status != 0 {
		t.Errorf("check: stdout %q, stderr %q, status %d; want no output and status 0", out, errOut, status)
	}
	// main: an ignored key, ESC, an empty ENTER, and a line with lower-case letters, an ignored TAB,
	// BS, " and \. The ESC after the line goes to hello. Then the ask of a file screen, of a list screen
	// with the cursor on its second row, and of a screen without rows.
	out, errOut, status = runIn(t, "ask", "preview", "main", "--size", "40x10", "--keys",
		`X,J,ESC,J,ENTER,J,j,o,TAB,e,BS,y,",\,ENTER,ESC,F,A,q,ENTER,ESC,ESC,L,DOWN,A,ESC,ENTER,ESC,ESC,E,A,x,ENTER,ESC,ESC,G`)
	if errOut != "" || status != 0 {
		t.Errorf("preview --keys: stderr %q, status %d", errOut, status)
	}
	golden(t, filepath.Join("testdata", "ask", "keys-40x10.dump"), out)
}
