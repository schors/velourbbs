package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestLamp checks the example BBS Lamp of example-style, section 4:
// velourbbs check gives no output, and the preview of each screen at 80x24 with 256 colours
// equals its golden dump.
func TestLamp(t *testing.T) {
	for _, lang := range []string{"en", "ru", "lt"} {
		fsys := os.DirFS(filepath.Join("..", "..", "examples", "lamp", lang))
		var out, errOut bytes.Buffer
		if status := run([]string{"check"}, t.TempDir(), fsys, &out, &errOut); out.Len() != 0 || errOut.Len() != 0 || status != 0 {
			t.Errorf("%s: check: stdout %q, stderr %q, status %d; want no output and status 0", lang, out.String(), errOut.String(), status)
		}
		for _, screen := range []string{"main", "about", "term", "texts", "log", "reader", "goodbye"} {
			out.Reset()
			errOut.Reset()
			args := []string{"preview", screen, "--size", "80x24", "--set", "term.colors=256"}
			if screen == "reader" {
				args = append(args, "--set", "arg.file=lamp.md")
			}
			status := run(args, t.TempDir(), fsys, &out, &errOut)
			if errOut.Len() != 0 || status != 0 {
				t.Errorf("%s: %s: stderr %q, status %d", lang, screen, errOut.String(), status)
			}
			golden(t, filepath.Join("testdata", "lamp", lang, screen+"-80x24-256.dump"), out.String())
		}
	}
}
