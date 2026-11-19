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
	cases := []struct{ screen, size, colors, set, dump string }{
		{screen: "main", size: "80x24", colors: "16"}, {screen: "main", size: "80x24", colors: "256"}, {screen: "main", size: "80x24", colors: "16777216"},
		{screen: "about", size: "80x24", colors: "16"}, {screen: "about", size: "80x24", colors: "256"}, {screen: "about", size: "80x24", colors: "16777216"},
		{screen: "term", size: "80x24", colors: "16"}, {screen: "term", size: "80x24", colors: "256"}, {screen: "term", size: "80x24", colors: "16777216"},
		{screen: "texts", size: "80x24", colors: "16"}, {screen: "texts", size: "80x24", colors: "256"}, {screen: "texts", size: "80x24", colors: "16777216"},
		{screen: "log", size: "80x24", colors: "16"}, {screen: "log", size: "80x24", colors: "256"}, {screen: "log", size: "80x24", colors: "16777216"},
		{screen: "reader", size: "80x24", colors: "16", set: "arg.file=lamp.md"},
		{screen: "reader", size: "80x24", colors: "256", set: "arg.file=lamp.md"},
		{screen: "reader", size: "80x24", colors: "16777216", set: "arg.file=lamp.md"},
		{screen: "reader", size: "80x24", colors: "256", set: "arg.file=rules.txt", dump: "reader-rules-80x24-256"},
		{screen: "goodbye", size: "80x24", colors: "16"}, {screen: "goodbye", size: "80x24", colors: "256"}, {screen: "goodbye", size: "80x24", colors: "16777216"},
		{screen: "main", size: "80x25", colors: "256"}, {screen: "main", size: "132x43", colors: "256"},
	}
	for _, lang := range []string{"en", "ru"} {
		fsys := os.DirFS(filepath.Join("..", "..", "examples", "lamp", lang))
		var out, errOut bytes.Buffer
		if status := run([]string{"check"}, t.TempDir(), fsys, &out, &errOut); out.Len() != 0 || errOut.Len() != 0 || status != 0 {
			t.Errorf("%s: check: stdout %q, stderr %q, status %d; want no output and status 0", lang, out.String(), errOut.String(), status)
		}
		for _, c := range cases {
			out.Reset()
			errOut.Reset()
			args := []string{"preview", c.screen, "--size", c.size, "--set", "term.colors=" + c.colors}
			if c.set != "" {
				args = append(args, "--set", c.set)
			}
			dump := c.dump
			if dump == "" {
				dump = c.screen + "-" + c.size + "-" + c.colors
			}
			status := run(args, t.TempDir(), fsys, &out, &errOut)
			file := filepath.Join("testdata", "lamp", lang, dump+".dump")
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
