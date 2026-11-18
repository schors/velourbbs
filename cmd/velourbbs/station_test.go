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
		file                 string   // The argument file of a file screen.
		ambiguous            string   // term.ambiguous; "" is the default 1.
		set                  []string // More --set values.
		tag                  string   // The name suffix of a case with set.
	}
	cases := []tcase{
		{screen: "main", size: "80x24", colors: "16"}, {screen: "main", size: "80x24", colors: "16777216"},
		{screen: "ru-texts", size: "80x24", colors: "16"}, {screen: "ru-texts", size: "80x24", colors: "16777216"},
		// The 80 form on a wide terminal with fewer than 43 rows.
		{screen: "main", size: "132x42", colors: "256"},
		// The hint of home-bbs, section 3.2, in both forms.
		{screen: "main", size: "80x24", colors: "256", ambiguous: "2"},
		{screen: "main", size: "132x43", colors: "256", ambiguous: "2"},
	}
	// The terminal rows of home-bbs, section 3.5, with each detected value.
	for _, c := range []struct{ screen, bg string }{{"ru-about", "#1e1e1e"}, {"en-about", "#ffffff"}} {
		for _, size := range []string{"80x24", "132x43"} {
			cases = append(cases, tcase{screen: c.screen, size: size, colors: "16777216", ambiguous: "2",
				set: []string{"term.name=WezTerm 20240203", "term.images=sixel", "term.bg_color=" + c.bg}, tag: "-detected"})
		}
	}
	// The longest term.type and term.name of wide graphemes keep the 80 form in 23 rows.
	wide := strings.Repeat("界", 40)
	cases = append(cases, tcase{screen: "ru-about", size: "80x24", colors: "256",
		set: []string{"term.type=" + wide, "term.name=" + wide}, tag: "-wide"})
	for _, screen := range []string{"main", "ru-menu", "en-menu", "ru-about", "en-about", "goodbye", "shutdown",
		"ru-texts", "en-texts", "ru-setup", "en-setup", "ru-bbt", "en-bbt"} {
		cases = append(cases, tcase{screen: screen, size: "80x24", colors: "256"},
			tcase{screen: screen, size: "132x43", colors: "256"})
	}
	// Each file screen with its first file: 80 form only.
	for _, lang := range []string{"ru", "en"} {
		cases = append(cases,
			tcase{screen: lang + "-setup-read", size: "80x24", colors: "256", file: lang + "/guide/setup/README.md"},
			tcase{screen: lang + "-bbt-read", size: "80x24", colors: "256", file: lang + "/guide/bbt/README.md"},
			tcase{screen: lang + "-text-read", size: "80x24", colors: "256", file: lang + "/features.md"})
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
		name := c.screen + "-" + c.size + "-" + c.colors
		if c.ambiguous != "" {
			args = append(args, "--set", "term.ambiguous="+c.ambiguous)
			name += "-ambiguous" + c.ambiguous
		}
		for _, v := range c.set {
			args = append(args, "--set", v)
		}
		name += c.tag
		status := run(args, fsys, &out, &errOut)
		file := filepath.Join("testdata", "station", name+".dump")
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
