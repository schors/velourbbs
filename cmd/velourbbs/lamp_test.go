package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestLamp checks the example BBS Lamp of example-style, section 8:
// velourbbs check gives no output, and each preview equals its golden dump.
func TestLamp(t *testing.T) {
	cases := []struct{ screen, size, colors string }{
		{"main", "80x24", "16"}, {"main", "80x24", "256"}, {"main", "80x24", "16777216"},
		{"about", "80x24", "16"}, {"about", "80x24", "256"}, {"about", "80x24", "16777216"},
		{"term", "80x24", "16"}, {"term", "80x24", "256"}, {"term", "80x24", "16777216"},
		{"goodbye", "80x24", "16"}, {"goodbye", "80x24", "256"}, {"goodbye", "80x24", "16777216"},
		{"main", "80x25", "256"}, {"main", "132x43", "256"},
	}
	for _, lang := range []string{"en"} {
		fsys := os.DirFS(filepath.Join("..", "..", "examples", "lamp", lang))
		var out, errOut bytes.Buffer
		if status := run([]string{"check"}, fsys, &out, &errOut); out.Len() != 0 || errOut.Len() != 0 || status != 0 {
			t.Errorf("%s: check: stdout %q, stderr %q, status %d; want no output and status 0", lang, out.String(), errOut.String(), status)
		}
		for _, c := range cases {
			out.Reset()
			errOut.Reset()
			status := run([]string{"preview", c.screen, "--size", c.size, "--set", "term.colors=" + c.colors}, fsys, &out, &errOut)
			file := filepath.Join("testdata", "lamp", lang, c.screen+"-"+c.size+"-"+c.colors+".dump")
			want, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			if errOut.Len() != 0 || status != 0 || out.String() != string(want) {
				t.Errorf("%s: stderr %q, status %d, got\n%s\nwant\n%s", file, errOut.String(), status, out.String(), want)
			}
		}
	}
}
