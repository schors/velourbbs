package artimport

import (
	"strings"
	"testing"

	"github.com/schors/velourbbs/internal/screen"
)

// TestPlayAfterDiagnostics checks that the play goes on after the diagnostic limit (screen-render, section 3.1).
func TestPlayAfterDiagnostics(t *testing.T) {
	d := &screen.FileDiags{File: "a.ans"}
	p := newPlayer(d, 4, false, true)
	if !p.play([]byte(strings.Repeat("\t", 200) + "\x1b[5Ha")) {
		t.Fatal("the play stopped")
	}
	if p.height != 5 || len(d.Diags) != screen.MaxFileDiags+1 {
		t.Errorf("height %d, %d diagnostics", p.height, len(d.Diags))
	}
}
