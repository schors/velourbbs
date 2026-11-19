package main

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
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
		tag                  string   // The name suffix of a case with set or keys.
		keys                 string   // The --keys value.
	}
	cases := []tcase{
		{screen: "main", size: "80x24", colors: "16"}, {screen: "main", size: "80x24", colors: "16777216"},
		{screen: "ru-texts", size: "80x24", colors: "16"}, {screen: "ru-texts", size: "80x24", colors: "16777216"},
		// The 80 form on a wide terminal with fewer than 43 rows.
		{screen: "main", size: "132x42", colors: "256"},
		// The hint of home-bbs, section 3.2, in both forms.
		{screen: "main", size: "80x24", colors: "256", ambiguous: "2"},
		{screen: "main", size: "132x43", colors: "256", ambiguous: "2"},
		// The entry, the Russian menu, the entry and the English menu.
		{screen: "main", size: "80x24", colors: "256", keys: "R,ESC,E", tag: "-keys"},
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
	// The file screen with the first file of each catalog: 80 form only.
	for _, c := range []struct{ lang, setup, bbt string }{
		{"ru", "Руководство по установке", "Руководство по BBT"}, {"en", "Setup guide", "BBT guide"},
	} {
		read := func(file, title, tag string) tcase {
			return tcase{screen: c.lang + "-read", size: "80x24", colors: "256", file: file,
				set: []string{"arg.title=" + title}, tag: "-" + tag}
		}
		cases = append(cases, read("guide/setup/README.md", c.setup, "setup"),
			read("guide/bbt/README.md", c.bbt, "bbt"), read("features.md", "Velour", "features"))
	}
	fsys := os.DirFS(filepath.Join("..", "..", "station"))
	var out, errOut bytes.Buffer
	if status := run([]string{"check"}, t.TempDir(), fsys, &out, &errOut); out.Len() != 0 || errOut.Len() != 0 || status != 0 {
		t.Errorf("check: stdout %q, stderr %q, status %d; want no output and status 0", out.String(), errOut.String(), status)
	}
	for _, c := range cases {
		out.Reset()
		errOut.Reset()
		// A screen of a language is <lang>-<name> (home-bbs, section 8).
		screen, lang := c.screen, ""
		if l, name, ok := strings.Cut(c.screen, "-"); ok {
			screen, lang = name, l
		}
		args := []string{"preview", screen, "--size", c.size, "--set", "term.colors=" + c.colors}
		if lang != "" {
			args = append(args, "--set", "session.lang="+lang)
		}
		if c.keys != "" {
			args = append(args, "--keys", c.keys)
		}
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
		status := run(args, t.TempDir(), fsys, &out, &errOut)
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

// titleArg is the title arg of an item. Its group 1 is the value.
var titleArg = regexp.MustCompile(`title = "([^"]*)"`)

// langLink is an action that sets the language, or a content path with
// a language directory. A screen of a language stays in its language.
var langLink = regexp.MustCompile(`do = "(lang |(goto|gosub) (ru|en)/)|file = "(ru|en)/`)

// TestStationItems checks that the screens of screens/ru/ and of screens/en/
// have the same items (home-bbs, section 8): both directories have the same files,
// and their front matter is equal. The title args are text of the language,
// so the test ignores their values. TestStationTitles checks the title values.
// No screen of a language names a language: the same items in both directories
// can still lead to the other language, as gosub en/about in both menus.
func TestStationItems(t *testing.T) {
	dir := filepath.Join("..", "..", "station", "screens")
	ru, err := filepath.Glob(filepath.Join(dir, "ru", "*.bbt"))
	if err != nil {
		t.Fatal(err)
	}
	en, err := filepath.Glob(filepath.Join(dir, "en", "*.bbt"))
	if err != nil {
		t.Fatal(err)
	}
	if len(ru) == 0 || len(ru) != len(en) {
		t.Fatalf("%d ru screens and %d en screens; want the same number, not 0", len(ru), len(en))
	}
	for _, ruFile := range ru {
		enFile := filepath.Join(dir, "en", filepath.Base(ruFile))
		ruFront := titleArg.ReplaceAllString(frontMatter(t, ruFile), `title = ""`)
		enFront := titleArg.ReplaceAllString(frontMatter(t, enFile), `title = ""`)
		if m := langLink.FindString(ruFront); m != "" {
			t.Errorf("%s: %s names a language", ruFile, m)
		}
		if enFront != ruFront {
			t.Errorf("%s: front matter\n%s\nwant the front matter of %s\n%s", enFile, enFront, ruFile, ruFront)
		}
	}
}

// TestStationTitles checks the title of each item that opens a reading screen
// (home-bbs, section 3.4): the name of the guide, or Velour for the features text.
func TestStationTitles(t *testing.T) {
	want := map[string]string{
		"en/texts.bbt": "Velour", "en/setup.bbt": "Setup guide", "en/bbt.bbt": "BBT guide",
		"ru/texts.bbt": "Velour", "ru/setup.bbt": "Руководство по установке", "ru/bbt.bbt": "Руководство по BBT",
	}
	// The fragments have no front matter, so the test reads only the screens of a language.
	files, err := filepath.Glob(filepath.Join("..", "..", "station", "screens", "[er][nu]", "*.bbt"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no screens of a language")
	}
	for _, file := range files {
		front := frontMatter(t, file)
		var titles []string
		for _, m := range titleArg.FindAllStringSubmatch(front, -1) {
			titles = append(titles, m[1])
		}
		reads := strings.Count(front, `"gosub read"`)
		w := want[filepath.Base(filepath.Dir(file))+"/"+filepath.Base(file)]
		if len(titles) != reads || reads > 0 && w == "" {
			t.Errorf("%s: %d reading items, %d titles", file, reads, len(titles))
			continue
		}
		for _, title := range titles {
			if title != w {
				t.Errorf("%s: title %q, want %q", file, title, w)
			}
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
