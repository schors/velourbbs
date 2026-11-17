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
	type tcase struct {
		screen, size, colors string
		file                 string // The argument file of a file screen.
	}
	cases := []tcase{
		{"main", "80x24", "16", ""}, {"main", "80x24", "16777216", ""},
		{"ru-texts", "80x24", "16", ""}, {"ru-texts", "80x24", "16777216", ""},
		// The 80 form on a wide terminal with fewer than 43 rows.
		{"main", "132x42", "256", ""},
	}
	for _, screen := range []string{"main", "ru-menu", "en-menu", "ru-about", "en-about", "goodbye", "shutdown",
		"ru-texts", "en-texts", "ru-setup", "en-setup", "ru-bbt", "en-bbt"} {
		cases = append(cases, tcase{screen, "80x24", "256", ""}, tcase{screen, "132x43", "256", ""})
	}
	// Each file screen with its first file: 80 form only.
	for _, lang := range []string{"ru", "en"} {
		cases = append(cases,
			tcase{lang + "-setup-read", "80x24", "256", lang + "/guide/setup/README.md"},
			tcase{lang + "-bbt-read", "80x24", "256", lang + "/guide/bbt/README.md"},
			tcase{lang + "-text-read", "80x24", "256", lang + "/features.md"})
	}
	fsys := os.DirFS(filepath.Join("..", "..", "station"))
	var out, errOut bytes.Buffer
	if status := run([]string{"check"}, fsys, &out, &errOut); out.Len() != 0 || errOut.Len() != 0 || status != 0 {
		t.Errorf("check: stdout %q, stderr %q, status %d; want no output and status 0", out.String(), errOut.String(), status)
	}
	for _, c := range cases {
		out.Reset()
		errOut.Reset()
		args := []string{"preview", c.screen, "--size", c.size, "--set", "term.colors=" + c.colors}
		if c.file != "" {
			args = append(args, "--set", "arg.file="+c.file)
		}
		status := run(args, fsys, &out, &errOut)
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

// TestStationGuideLinks checks that content/<lang>/guide is a symlink into docs/guide/
// (home-bbs, section 2).
func TestStationGuideLinks(t *testing.T) {
	for _, lang := range []string{"ru", "en"} {
		link := filepath.Join("..", "..", "station", "content", lang, "guide")
		target, err := os.Readlink(link)
		if err != nil {
			t.Fatal(err)
		}
		if want := "../../../docs/guide/" + lang; target != want {
			t.Errorf("%s: target %q, want %q", link, target, want)
		}
	}
}

// TestStationItems checks that the ru- and the en- screens of the home BBS
// have the same items: the front matter of each file equals the front matter
// of the other file with the language replaced in the screen names and in the
// content paths. Both directions catch a goto from a screen of one language
// to a screen or a file of the other.
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
	toEn := strings.NewReplacer("ru-", "en-", `"ru/`, `"en/`)
	toRu := strings.NewReplacer("en-", "ru-", `"en/`, `"ru/`)
	for _, ruFile := range ru {
		enFile := filepath.Join(dir, "en-"+strings.TrimPrefix(filepath.Base(ruFile), "ru-"))
		ruFront, enFront := frontMatter(t, ruFile), frontMatter(t, enFile)
		if want := toEn.Replace(ruFront); enFront != want {
			t.Errorf("%s: front matter\n%s\nwant\n%s", enFile, enFront, want)
		}
		if want := toRu.Replace(enFront); ruFront != want {
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
