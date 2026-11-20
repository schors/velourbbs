package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"testing/fstest"
)

// stationFS is the home BBS with the guide fixture in place of each content/<lang>/guide,
// so an edit of docs/guide/ does not change a test.
type stationFS struct{ base, guide fs.FS }

// newStationFS gives the home BBS with the guide fixture. Each guide has a README.md;
// bbt has 10 parts more, so a list has more than 9 rows.
func newStationFS() stationFS {
	guide := fstest.MapFS{}
	for _, dir := range []string{"setup", "bbt"} {
		guide[dir+"/README.md"] = &fstest.MapFile{Data: []byte("# Fixture guide\n\nThe guide fixture of the station tests.\n")}
	}
	for i := 1; i <= 10; i++ {
		guide[fmt.Sprintf("bbt/%d-part.md", i)] = &fstest.MapFile{Data: []byte(fmt.Sprintf("# Part %d\n\nThe text of part %d.\n", i, i))}
	}
	return stationFS{os.DirFS(filepath.Join("..", "..", "station")), guide}
}

func (s stationFS) Open(name string) (fs.File, error) {
	for _, lang := range []string{"ru", "en"} {
		dir := "content/" + lang + "/guide"
		if name == dir {
			return s.guide.Open(".")
		}
		if rest, ok := strings.CutPrefix(name, dir+"/"); ok {
			return s.guide.Open(rest)
		}
	}
	return s.base.Open(name)
}

// TestStation checks the home BBS of home-bbs, section 5: velourbbs check gives
// no output, and the preview of each screen at 80x24 and 132x43 with 256 colours
// equals its golden dump.
func TestStation(t *testing.T) {
	fsys := newStationFS()
	var out, errOut bytes.Buffer
	if status := run([]string{"check"}, t.TempDir(), fsys, &out, &errOut); out.Len() != 0 || errOut.Len() != 0 || status != 0 {
		t.Errorf("check: stdout %q, stderr %q, status %d; want no output and status 0", out.String(), errOut.String(), status)
	}
	// The station has no layout warning at its sizes (screen-render, section 3.6).
	args := []string{"check", "--size", "80x24", "--size", "132x43"}
	if status := run(args, t.TempDir(), fsys, &out, &errOut); out.Len() != 0 || errOut.Len() != 0 || status != 0 {
		t.Errorf("%q: stdout %q, stderr %q, status %d; want no output and status 0", args, out.String(), errOut.String(), status)
	}
	for _, size := range []string{"80x24", "132x43"} {
		for _, screen := range []string{"main", "main-ambiguous", "goodbye", "shutdown", "ru-menu", "en-menu", "ru-about", "en-about",
			"ru-texts", "en-texts", "ru-setup", "en-setup", "ru-bbt", "en-bbt", "ru-read", "en-read"} {
			out.Reset()
			errOut.Reset()
			args := []string{"preview", screen, "--size", size, "--set", "term.colors=256"}
			// A screen of a language is <lang>-<name> (home-bbs, section 5).
			if screen == "main-ambiguous" {
				args = []string{"preview", "main", "--size", size, "--set", "term.colors=256", "--set", "term.ambiguous=2"}
			} else if lang, name, ok := strings.Cut(screen, "-"); ok {
				args = []string{"preview", name, "--size", size, "--set", "term.colors=256", "--set", "session.lang=" + lang}
			}
			if strings.HasSuffix(screen, "-read") {
				args = append(args, "--set", "arg.file=guide/setup/README.md", "--set", "arg.title=Fixture guide")
			}
			status := run(args, t.TempDir(), fsys, &out, &errOut)
			if errOut.Len() != 0 || status != 0 {
				t.Errorf("%s %s: stderr %q, status %d", screen, size, errOut.String(), status)
			}
			golden(t, filepath.Join("testdata", "station", screen+"-"+size+"-256.dump"), out.String())
		}
	}
}

// A syntax error in screens/screens.toml of the home BBS gives one error:
// the check stops before the screens (screen-render, section 3.1).
func TestStationSettingsError(t *testing.T) {
	base := newStationFS()
	data, err := fs.ReadFile(base, "screens/screens.toml")
	if err != nil {
		t.Fatal(err)
	}
	bad := strings.Replace(string(data), `gold   = "#b8902f"`, `gold   = #b8902f`, 1)
	if bad == string(data) {
		t.Fatal(`screens/screens.toml has no line gold = "#b8902f"`)
	}
	fsys := overlay{base, fstest.MapFS{"screens/screens.toml": {Data: []byte(bad)}}}
	want := "screens/screens.toml:22:10: unexpected character U+0023 '#' at start of value\n"
	var out, errOut bytes.Buffer
	status := run([]string{"check"}, t.TempDir(), fsys, &out, &errOut)
	if out.Len() != 0 || errOut.String() != want || status != exitError {
		t.Errorf("stdout %q, stderr:\n%s\nstatus %d; want stderr:\n%s\nstatus 1", out.String(), errOut.String(), status, want)
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
// have the same items (home-bbs, section 5): both directories have the same files,
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

// TestStationTitles checks the title of each item and select that opens a reading screen
// (home-bbs, section 3): the name of the guide, or Velour for the features text.
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
