package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestStation checks the home BBS of home-bbs, section 8: velourbbs check gives
// no output, and each preview equals its golden dump.
func TestStation(t *testing.T) {
	type tcase struct{ screen, size, colors string }
	cases := []tcase{
		{"main", "80x24", "16"}, {"main", "80x24", "16777216"},
		// The 80 form on a wide terminal with fewer than 43 rows.
		{"main", "132x42", "256"},
	}
	for _, screen := range []string{"main", "ru-menu", "en-menu", "ru-about", "en-about", "goodbye"} {
		cases = append(cases, tcase{screen, "80x24", "256"}, tcase{screen, "132x43", "256"})
	}
	fsys := os.DirFS(filepath.Join("..", "..", "station"))
	var out, errOut bytes.Buffer
	if status := run([]string{"check"}, fsys, &out, &errOut); out.Len() != 0 || errOut.Len() != 0 || status != 0 {
		t.Errorf("check: stdout %q, stderr %q, status %d; want no output and status 0", out.String(), errOut.String(), status)
	}
	for _, c := range cases {
		out.Reset()
		errOut.Reset()
		status := run([]string{"preview", c.screen, "--size", c.size, "--set", "term.colors=" + c.colors}, fsys, &out, &errOut)
		file := filepath.Join("testdata", "station", c.screen+"-"+c.size+"-"+c.colors+".dump")
		want, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if errOut.Len() != 0 || status != 0 || out.String() != string(want) {
			t.Errorf("%s: stderr %q, status %d, got\n%s\nwant\n%s", file, errOut.String(), status, out.String(), want)
		}
	}
}

// TestStationItems checks that the ru- and the en- screens of the home BBS
// have the same items: the front matter of each file equals the front matter
// of the other file with the language prefix replaced. Both directions catch
// a goto from a screen of one language to a screen of the other.
func TestStationItems(t *testing.T) {
	dir := filepath.Join("..", "..", "station", "screens")
	ru, err := filepath.Glob(filepath.Join(dir, "ru-*.bbt"))
	if err != nil {
		t.Fatal(err)
	}
	en, err := filepath.Glob(filepath.Join(dir, "en-*.bbt"))
	if err != nil {
		t.Fatal(err)
	}
	if len(ru) == 0 || len(ru) != len(en) {
		t.Fatalf("%d ru- screens and %d en- screens; want the same number, not 0", len(ru), len(en))
	}
	for _, ruFile := range ru {
		enFile := filepath.Join(dir, "en-"+strings.TrimPrefix(filepath.Base(ruFile), "ru-"))
		ruFront, enFront := frontMatter(t, ruFile), frontMatter(t, enFile)
		if want := strings.ReplaceAll(ruFront, "ru-", "en-"); enFront != want {
			t.Errorf("%s: front matter\n%s\nwant\n%s", enFile, enFront, want)
		}
		if want := strings.ReplaceAll(enFront, "en-", "ru-"); ruFront != want {
			t.Errorf("%s: front matter\n%s\nwant\n%s", ruFile, ruFront, want)
		}
	}
}

// frontMatter gives the front matter of the BBT file, with its +++ lines.
func frontMatter(t *testing.T, file string) string {
	t.Helper()
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	end := strings.Index(s, "\n+++\n")
	if !strings.HasPrefix(s, "+++\n") || end < 0 {
		t.Fatalf("%s: no front matter", file)
	}
	return s[:end+len("\n+++\n")]
}
