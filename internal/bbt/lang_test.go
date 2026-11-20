package bbt_test

import (
	"io/fs"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/schors/velourbbs/internal/bbt"
	"github.com/schors/velourbbs/internal/screen"
)

// item is a front matter with one item.
func item(key, do string) string {
	return "+++\n[[item]]\nkey = \"" + key + "\"\ndo = \"" + do + "\"\n+++\n"
}

// loadLangs loads the files, given by their paths in the BBS directory, with the languages langs.
// A file with the mode in modes has no data. A symlink in modes links to "missing".
func loadLangs(files map[string]string, modes map[string]fs.FileMode, langs []string, tables bool) (bbt.Sets, []string) {
	fsys := fstest.MapFS{}
	for name, data := range files {
		fsys[name] = &fstest.MapFile{Data: []byte(data)}
	}
	for name, mode := range modes {
		fsys[name] = &fstest.MapFile{Mode: mode}
		if mode == fs.ModeSymlink {
			fsys[name].Data = []byte("missing")
		}
	}
	sets, ds := bbt.Load(fsys, bbt.Settings{
		Theme: theme, Layout: screen.Layout{Width: defaultWidth}, Fields: aliases, Langs: langs, LangTables: tables,
	})
	var msgs []string
	for _, d := range ds {
		msgs = append(msgs, d.String())
	}
	return sets, msgs
}

// The search of session-language, section 4: a language file, else the common file,
// for a screen, an include, an art insert and a content file, also in a common screen.
func TestSearch(t *testing.T) {
	files := map[string]string{
		"screens/main.bbt":     item("G", "logoff") + "{>_logo}\n{>pic.ans}\n",
		"screens/_logo.bbt":    "Logo",
		"screens/ru/_logo.bbt": "Лого",
		"screens/pic.ans":      sauceArt("x", 10),
		"screens/ru/pic.ans":   sauceArt("y", 12),
		"screens/menu.bbt":     item("G", "logoff") + "Common menu\n",
		"screens/en/menu.bbt":  item("G", "logoff") + "Menu\n",
		"screens/ru/menu.bbt":  item("G", "logoff") + "Меню\n",
		"screens/text.bbt":     reader(`file = "a.txt"`, "{>file}\n"),
		"content/a.txt":        "common\n",
		"content/ru/a.txt":     "русский\n",
		"screens/de/x.bbt":     "not a language\n",
	}
	sets, msgs := loadLangs(files, nil, []string{"en", "ru"}, true)
	want := []string{
		`screens/de:1:1: warning: "de" is not a language of [lang]`,
		`screens/menu.bbt:1:1: warning: no screen uses this file`,
	}
	if !slices.Equal(msgs, want) {
		t.Fatalf("got\n%s\nwant\n%s", strings.Join(msgs, "\n"), strings.Join(want, "\n"))
	}
	for _, tt := range []struct{ lang, main, menu, text string }{
		{"en", `screens/_logo.bbt("Logo" /) art10[]`, "screens/en/menu.bbt", "common"},
		{"ru", `screens/ru/_logo.bbt("Лого" /) art12[]`, "screens/ru/menu.bbt", "русский"},
	} {
		set := sets[tt.lang]
		if got := show(set["main"].Body); !strings.HasPrefix(got, tt.main) {
			t.Errorf("%s: main %s, want the prefix %s", tt.lang, got, tt.main)
		}
		if got := set["menu"].File; got != tt.menu {
			t.Errorf("%s: menu file %s, want %s", tt.lang, got, tt.menu)
		}
		if c := set["text"].FileScreen.Contents["a.txt"]; c == nil || !slices.Equal(c.Lines, []string{tt.text}) {
			t.Errorf("%s: text content %v, want %q", tt.lang, c, tt.text)
		}
	}

	// Without [lang] tables, a directory of screens/ gives no warning.
	_, msgs = loadLangs(map[string]string{"screens/main.bbt": item("G", "logoff"), "screens/art/x.ans": "x"}, nil, nil, false)
	if len(msgs) != 0 {
		t.Errorf("no tables: %q", msgs)
	}
}

// A directory entry of the language is not a step to the common file: a broken symlink
// and a file that is not a regular file are errors (session-language, section 4).
func TestSearchNotRegular(t *testing.T) {
	files := map[string]string{
		"screens/main.bbt":  item("G", "logoff") + "{>_logo}\n",
		"screens/_logo.bbt": "Logo",
		"screens/menu.bbt":  item("G", "logoff"),
		"content/a.txt":     "a\n",
		"screens/text.bbt":  reader(`file = "a.txt"`, "{>file}\n"),
	}
	modes := map[string]fs.FileMode{
		"screens/ru/_logo.bbt": fs.ModeSymlink,
		"screens/ru/menu.bbt":  fs.ModeDir,
		"screens/ru/news.bbt":  fs.ModeSymlink,
		"content/ru/a.txt":     fs.ModeNamedPipe,
	}
	_, msgs := loadLangs(files, modes, []string{"en", "ru"}, true)
	want := []string{
		`screens/main.bbt:6:1: fragment "screens/ru/_logo.bbt" does not exist (language "ru")`,
		`screens/ru/menu.bbt:1:1: not a regular file`,
		`screens/ru/news.bbt:1:1: not a regular file`,
		`screens/text.bbt:2:8: file: content file "a.txt": not a regular file (language "ru")`,
	}
	if !slices.Equal(msgs, want) {
		t.Errorf("got\n%s\nwant\n%s", strings.Join(msgs, "\n"), strings.Join(want, "\n"))
	}

	// A directory with an invalid screen name is not a regular file, as without languages.
	_, msgs = loadLangs(map[string]string{"screens/main.bbt": item("G", "logoff")},
		map[string]fs.FileMode{"screens/BAD.bbt": fs.ModeDir}, nil, false)
	if want := []string{`screens/BAD.bbt:1:1: not a regular file`}; !slices.Equal(msgs, want) {
		t.Errorf("directory: %q, want %q", msgs, want)
	}
}

// The parse of goto <code>/<name>, gosub <code>/<name> and lang <code> (bbt-format, section 8.3).
func TestLangActions(t *testing.T) {
	files := map[string]string{
		"screens/main.bbt": "+++\n" +
			"[[item]]\nkey = \"A\"\ndo = \"goto ru/main\"\n" +
			"[[item]]\nkey = \"B\"\ndo = \"gosub en/main\"\n" +
			"[[item]]\nkey = \"C\"\ndo = \"lang ru\"\n+++\n",
	}
	sets, msgs := loadLangs(files, nil, []string{"en", "ru"}, true)
	if len(msgs) != 0 {
		t.Fatalf("diags %q", msgs)
	}
	want := []screen.Action{
		{Op: screen.Goto, Screen: "main", Lang: "ru"},
		{Op: screen.Gosub, Screen: "main", Lang: "en"},
		{Op: screen.Lang, Lang: "ru"},
	}
	for i, it := range sets["ru"]["main"].Items {
		if it.Do.Op != want[i].Op || it.Do.Screen != want[i].Screen || it.Do.Lang != want[i].Lang {
			t.Errorf("item %s: %+v, want %+v", it.Key, it.Do, want[i])
		}
	}

	tests := []struct{ do, want string }{
		{"lang de", `"de" is not a language of the BBS`},
		{"goto de/main", `"de" is not a language of the BBS`},
		{"gosub de/main", `"de" is not a language of the BBS`},
		{"lang", `invalid lang action "lang"`},
		{"lang Ru", `invalid lang action "lang Ru"`},
		{"lang ru ", `invalid lang action "lang ru "`},
		{"goto RU/main", `invalid screen name "RU/main"`},
		{"goto ru/", `invalid screen name "ru/"`},
		{"goto ru/main/x", `invalid screen name "ru/main/x"`},
		{"gosub /main", `invalid screen name "/main"`},
	}
	for _, tt := range tests {
		_, msgs := loadLangs(map[string]string{"screens/main.bbt": item("A", tt.do)}, nil, []string{"en", "ru"}, true)
		want := `screens/main.bbt:4:6: item "A": do: ` + tt.want
		if !slices.Equal(msgs, []string{want}) {
			t.Errorf("%q: %q, want %q", tt.do, msgs, want)
		}
	}
}

// The diagnostics of the screen sets (session-language, section 6): an equal diagnostic once,
// the suffix for an error that only some sets find, and the values of an argument in each set.
func TestSetDiags(t *testing.T) {
	files := map[string]string{
		// The unknown style is an error in each set: once, without a suffix.
		"screens/main.bbt": item("G", "logoff") + "{!nostyle}\n{>_x}\n",
		// Only the set ru has the fragment.
		"screens/ru/_x.bbt": "x",
		// The set en finds the value x.md of the item of ru/texts in content/: it does not exist there.
		"screens/ru/texts.bbt": "+++\n[[item]]\nkey = \"3\"\ndo = \"goto read\"\nargs = { file = \"x.md\" }\n+++\n",
		"screens/en/texts.bbt": "+++\n[[item]]\nkey = \"3\"\ndo = \"goto read\"\nargs = { file = \"y.md\" }\n+++\n",
		"screens/read.bbt":     reader(`file = "{arg.file}"`, "{>file}\n"),
		"content/ru/x.md":      "x\n",
		"content/y.md":         "y\n",
	}
	_, msgs := loadLangs(files, nil, []string{"en", "ru"}, true)
	want := []string{
		`screens/main.bbt:6:1: unknown style "nostyle"`,
		`screens/main.bbt:7:1: fragment "screens/_x.bbt" does not exist (language "en")`,
		`screens/ru/texts.bbt:5:17: item "3": args: file: content file "x.md" does not exist (language "en")`,
	}
	if !slices.Equal(msgs, want) {
		t.Errorf("got\n%s\nwant\n%s", strings.Join(msgs, "\n"), strings.Join(want, "\n"))
	}

	// The sets en and ru read content/ru/bad.txt: its error appears once.
	_, msgs = loadLangs(map[string]string{
		"screens/main.bbt":   reader(`file = "ru/bad.txt"`, "{>file}\n"),
		"content/ru/bad.txt": "\x01\n",
	}, nil, []string{"en", "ru"}, true)
	if want := []string{`content/ru/bad.txt:1:1: control character U+0001`}; !slices.Equal(msgs, want) {
		t.Errorf("content/ru/bad.txt: got %q, want %q", msgs, want)
	}

	// With one language, there is no suffix.
	delete(files, "screens/ru/_x.bbt")
	_, msgs = loadLangs(map[string]string{"screens/main.bbt": files["screens/main.bbt"]}, nil, []string{"ru"}, false)
	want = []string{
		`screens/main.bbt:6:1: unknown style "nostyle"`,
		`screens/main.bbt:7:1: fragment "screens/_x.bbt" does not exist`,
	}
	if !slices.Equal(msgs, want) {
		t.Errorf("one language: got\n%s\nwant\n%s", strings.Join(msgs, "\n"), strings.Join(want, "\n"))
	}
}
