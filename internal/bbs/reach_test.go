package bbs

import (
	"slices"
	"strings"
	"testing"
)

// The warning of a screen that a user cannot reach (menu-nav, section 6.1).
func TestUnreachable(t *testing.T) {
	const warn = ":1:1: warning: a user cannot reach this screen"
	tests := []struct {
		name  string
		langs []string
		files map[string]string
		want  []string
	}{{
		// old and old2 only open each other. goodbye and shutdown get no warning.
		name: "goto, gosub and a cycle",
		files: map[string]string{
			"screens/main.bbt":     front("A goto a", "B gosub b", "G logoff"),
			"screens/a.bbt":        front("ESC goto main"),
			"screens/b.bbt":        front("ESC back"),
			"screens/old.bbt":      front("O goto old2"),
			"screens/old2.bbt":     front("O gosub old"),
			"screens/goodbye.bbt":  "Bye\n",
			"screens/shutdown.bbt": "Stop\n",
		},
		want: []string{"screens/old.bbt" + warn, "screens/old2.bbt" + warn},
	}, {
		// The walk ignores when: the item is a path.
		name: "when",
		files: map[string]string{
			"screens/main.bbt": "+++\n[[item]]\nkey = \"A\"\ndo = \"gosub a\"\nwhen = 'term.cols > 1000'\n" +
				"[[item]]\nkey = \"G\"\ndo = \"logoff\"\n+++\n",
			"screens/a.bbt": front("ESC back"),
		},
	}, {
		name: "select",
		files: map[string]string{
			"screens/main.bbt": "+++\ndata = { from = \"files\", args = { dir = \"log\" } }\nrows = 3\n" +
				"select = { do = \"gosub read\", args = { file = \"{row.path}\" } }\n" +
				"[[item]]\nkey = \"G\"\ndo = \"logoff\"\n+++\n{#each}{row.title}{/each}\n",
			"screens/read.bbt": front("ESC back") + "{arg.file}\n",
			"content/log/a.md": "# A\n",
		},
	}, {
		// gosub ru/menu reaches menu in ru, not in en. back from it shows main in ru.
		name:  "code",
		langs: []string{"en", "ru"},
		files: map[string]string{
			"screens/main.bbt":    front("R gosub ru/menu", "G logoff"),
			"screens/en/menu.bbt": front("ESC back"),
			"screens/ru/menu.bbt": front("ESC back"),
		},
		want: []string{"screens/en/menu.bbt" + warn},
	}, {
		name:  "lang",
		langs: []string{"en", "ru"},
		files: map[string]string{
			"screens/main.bbt": front("L lang ru", "G logoff"),
		},
	}, {
		// No path goes to ru: the common file gets the suffix.
		name:  "no language path",
		langs: []string{"en", "ru"},
		files: map[string]string{
			"screens/main.bbt": front("G logoff"),
		},
		want: []string{"screens/main.bbt" + warn + ` (language "ru")`},
	}, {
		// The example of section 6.1: main in ru has a path only through back from ru/menu.
		name:  "stack",
		langs: []string{"en", "ru"},
		files: map[string]string{
			"screens/main.bbt":    front("E gosub en/menu", "R gosub ru/menu", "G logoff"),
			"screens/en/menu.bbt": front("N gosub news", "ESC back"),
			"screens/ru/menu.bbt": front("ESC back"),
			"screens/news.bbt":    front("ESC back"),
		},
		want: []string{"screens/news.bbt" + warn + ` (language "ru")`},
	}, {
		// en/menu goes to ru/x, so back shows menu in ru: ru/menu goes to news.
		// back from menu then shows main in ru.
		name:  "stack in stack",
		langs: []string{"en", "ru"},
		files: map[string]string{
			"screens/main.bbt":    front("M gosub menu", "G logoff"),
			"screens/en/menu.bbt": front("R gosub ru/x", "ESC back"),
			"screens/ru/menu.bbt": front("N gosub news", "ESC back"),
			"screens/x.bbt":       front("ESC back"),
			"screens/news.bbt":    front("ESC back"),
			"screens/ru/old.bbt":  front("M goto main"),
			"screens/en/old.bbt":  front("M goto main"),
		},
		want: []string{
			"screens/en/old.bbt" + warn,
			"screens/news.bbt" + warn + ` (language "en")`,
			"screens/ru/old.bbt" + warn,
			"screens/x.bbt" + warn + ` (language "en")`,
		},
	}, {
		// The walk from en/menu reaches de only through back from ru/x to ru/menu.
		// So back from menu shows main in de. ru/main has no other path to de.
		name:  "language through back",
		langs: []string{"de", "en", "ru"},
		files: map[string]string{
			"screens/en/main.bbt": front("M gosub menu", "G logoff"),
			"screens/ru/main.bbt": front("G logoff"),
			"screens/de/main.bbt": front("G logoff"),
			"screens/en/menu.bbt": front("R gosub ru/x", "ESC back"),
			"screens/ru/menu.bbt": front("D gosub de/x", "ESC back"),
			"screens/de/menu.bbt": front("ESC back"),
			"screens/x.bbt":       front("ESC back"),
		},
		want: []string{"screens/x.bbt" + warn + ` (language "en")`},
	}, {
		// The walk runs only without errors.
		name: "error",
		files: map[string]string{
			"screens/main.bbt": front("A gosub nope", "G logoff"),
			"screens/old.bbt":  front("ESC back"),
		},
		want: []string{`screens/main.bbt:2:1: item "A": do: unknown screen "nope"`},
	}, {
		name:  "error of the sets",
		langs: []string{"en", "ru"},
		files: map[string]string{
			"screens/main.bbt":   front("G logoff"),
			"screens/ru/old.bbt": front("ESC back"),
		},
		want: []string{`screens/ru/old.bbt:1:1: no screen "old" (language "en")`},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, msgs := loadLangs(t, tt.langs, tt.files)
			// The warning does not stop the load.
			errs := slices.ContainsFunc(tt.want, func(w string) bool { return !strings.Contains(w, ": warning: ") })
			if !slices.Equal(msgs, tt.want) || (m == nil) != errs {
				t.Errorf("model %v, got\n%s\nwant\n%s", m != nil, strings.Join(msgs, "\n"), strings.Join(tt.want, "\n"))
			}
		})
	}
}
