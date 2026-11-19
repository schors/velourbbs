package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestArtExample checks the art example of bbt-guide, sections 3 and 5:
// velourbbs check gives no output, and the preview equals its golden dump.
func TestArtExample(t *testing.T) {
	fsys := os.DirFS(filepath.Join("..", "..", "examples", "art"))
	var out, errOut bytes.Buffer
	if status := run([]string{"check"}, t.TempDir(), fsys, &out, &errOut); out.Len() != 0 || errOut.Len() != 0 || status != 0 {
		t.Errorf("check: stdout %q, stderr %q, status %d; want no output and status 0", out.String(), errOut.String(), status)
	}
	out.Reset()
	errOut.Reset()
	status := run([]string{"preview", "main", "--size", "80x24", "--set", "term.colors=256"}, t.TempDir(), fsys, &out, &errOut)
	if errOut.Len() != 0 || status != 0 {
		t.Errorf("preview: stderr %q, status %d", errOut.String(), status)
	}
	golden(t, filepath.Join("testdata", "art", "main-80x24-256.dump"), out.String())
}
