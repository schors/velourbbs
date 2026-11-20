package bbs

import (
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

// front gives a screen with the items "<KEY> <do>" and no body.
// In the file, the do of the n-th item is at (3n+1):6.
func front(items ...string) string {
	var b strings.Builder
	b.WriteString("+++\n")
	for _, it := range items {
		k, do, _ := strings.Cut(it, " ")
		b.WriteString("[[item]]\nkey = \"" + k + "\"\ndo = \"" + do + "\"\n")
	}
	b.WriteString("+++\n")
	return b.String()
}

// loadLangs loads the BBS of the files with the [lang] tables of the codes langs.
func loadLangs(t *testing.T, langs []string, files map[string]string) (*Model, []string) {
	t.Helper()
	fsys := fstest.MapFS{"bbs.toml": {}}
	var toml strings.Builder
	for _, code := range langs {
		toml.WriteString("[lang." + code + "]\n")
	}
	fsys["screens/screens.toml"] = &fstest.MapFile{Data: []byte(toml.String())}
	for name, data := range files {
		fsys[name] = &fstest.MapFile{Data: []byte(data)}
	}
	m, ds := Load(fsys)
	var msgs []string
	for _, d := range ds {
		msgs = append(msgs, d.String())
	}
	return m, msgs
}

// The checks of session-language, section 6, with each error and the language suffix.
func TestCheckSets(t *testing.T) {
	main := front("E gosub en/menu", "R gosub ru/menu", "D gosub de/menu", "G logoff")
	tests := []struct {
		name  string
		langs []string
		files map[string]string
		want  []string
	}{{
		name:  "valid",
		langs: []string{"en", "ru"},
		files: map[string]string{
			"screens/main.bbt":    front("E gosub en/menu", "R gosub ru/menu", "G logoff"),
			"screens/en/menu.bbt": front("L lang ru", "ESC back"),
			"screens/ru/menu.bbt": front("L lang en", "ESC back"),
		},
	}, {
		// Only ru has news; en has another argument of read.
		name:  "sets differ",
		langs: []string{"en", "ru"},
		files: map[string]string{
			"screens/main.bbt":    front("G logoff"),
			"screens/ru/news.bbt": front("G logoff"),
			"screens/en/read.bbt": front("G logoff") + "{arg.a}\n",
			"screens/ru/read.bbt": front("G logoff") + "{arg.b}\n",
		},
		want: []string{
			`screens/en/read.bbt:1:1: arguments differ (language "ru")`,
			`screens/ru/news.bbt:1:1: no screen "news" (language "en")`,
		},
	}, {
		// The common screen about has the item M in each set; only de has no menu.
		name:  "common file",
		langs: []string{"de", "en", "ru"},
		files: map[string]string{
			"screens/main.bbt":    main,
			"screens/about.bbt":   front("M gosub menu"),
			"screens/en/menu.bbt": front("ESC back"),
			"screens/ru/menu.bbt": front("ESC back"),
		},
		want: []string{
			`screens/about.bbt:4:6: item "M": do: unknown screen "menu" (language "de")`,
			`screens/en/menu.bbt:1:1: no screen "menu" (language "de")`,
			`screens/main.bbt:10:6: item "D": do: unknown screen "de/menu"`,
		},
	}, {
		// The reference set of a common screen is the first set that holds it.
		name:  "common reference",
		langs: []string{"en", "ru"},
		files: map[string]string{
			"screens/main.bbt":    front("G logoff"),
			"screens/news.bbt":    front("G logoff"),
			"screens/ru/news.bbt": front("G logoff") + "{arg.a}\n",
		},
		want: []string{
			`screens/news.bbt:1:1: arguments differ (language "ru")`,
		},
	}, {
		// An equal error of a common file in each set appears once, without a suffix.
		name:  "each set",
		langs: []string{"en", "ru"},
		files: map[string]string{
			"screens/main.bbt": front("G gosub nope"),
		},
		want: []string{
			`screens/main.bbt:4:6: item "G": do: unknown screen "nope"`,
		},
	}, {
		// The start screen is in each set.
		name:  "start",
		langs: []string{"en", "ru"},
		files: map[string]string{
			"screens/en/main.bbt": front("G logoff"),
			"screens/ru/menu.bbt": front("G logoff"),
		},
		want: []string{
			`screens/en/main.bbt:1:1: no screen "main" (language "ru")`,
			`screens/ru/menu.bbt:1:1: no screen "menu" (language "en")`,
			`screens/screens.toml:1:1: menu: start: unknown screen "main" (language "ru")`,
		},
	}, {
		// The args of a goto with a code have the checks of the target in the set of the code.
		name:  "args",
		langs: []string{"en", "ru"},
		files: map[string]string{
			"screens/main.bbt": "+++\n[[item]]\nkey = \"R\"\ndo = \"gosub ru/read\"\nargs = { b = \"x\" }\n" +
				"[[item]]\nkey = \"G\"\ndo = \"logoff\"\n+++\n",
			"screens/read.bbt": front("ESC back") + "{arg.a}\n",
		},
		want: []string{
			`screens/main.bbt:5:8: item "R": args: missing argument "a" of screen "ru/read"`,
			`screens/main.bbt:5:8: item "R": args: screen "ru/read" has no argument "b"`,
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, msgs := loadLangs(t, tt.langs, tt.files)
			if !slices.Equal(msgs, tt.want) || (m != nil) != (len(tt.want) == 0) {
				t.Errorf("model %v, got\n%s\nwant\n%s", m != nil, strings.Join(msgs, "\n"), strings.Join(tt.want, "\n"))
			}
		})
	}
}
