package bbs

import (
	"slices"
	"strings"
	"testing"
)

// listScreen is a list screen of the directory parts with the key select sel, the item of quit
// and a list line.
func listScreen(sel string) string {
	return "+++\ndata = { from = \"files\", args = { dir = \"parts\" } }\nrows = 2\n" + sel +
		"\n[[item]]\nkey = \"Q\"\ndo = \"logoff\"\n+++\n{#each}{#if row.selected}>{/if}{row.title}{/each}\n"
}

// selectFiles gives the files of a BBS with the list screen sel of the select sel.
// The screen main opens it. The screen read is a file screen with the argument file.
func selectFiles(sel string, more map[string]string) map[string]string {
	files := map[string]string{
		"screens/main.bbt":   front("L gosub sel", "Q logoff"),
		"screens/sel.bbt":    listScreen(sel),
		"screens/read.bbt":   "+++\nfile = \"{arg.file}\"\n[[item]]\nkey = \"ESC\"\ndo = \"back\"\n+++\n{>file}\n",
		"content/parts/1.md": "# One\n",
		"content/parts/2.md": "# Two\n",
		"content/parts/3.md": "# Three\n",
	}
	for k, v := range more {
		files[k] = v
	}
	return files
}

// The errors of select in one file (list-view, sections 8 and 10).
func TestSelectFileErrors(t *testing.T) {
	_, msgs := loadLangs(t, nil, map[string]string{
		"screens/main.bbt": front("Q logoff"),
		"screens/a.bbt":    "+++\nselect = { do = \"back\" }\n[[item]]\nkey = \"Q\"\ndo = \"logoff\"\n+++\nText\n",
		"screens/b.bbt":    listScreen("select = { args = { file = \"{row.path}\" } }"),
		"screens/c.bbt":    listScreen("select = { do = \"jump\" }"),
		"screens/d.bbt":    listScreen("select = { do = \"back\", args = { file = \"{row.path}\" } }"),
		"screens/e.bbt": listScreen("select = { do = \"gosub read\", args = { " +
			"a = \"{row.path\", b = \"{row.size}\", c = \"{row.selected}\", d = \"{arg.dir}\", X = \"{row.path}\", e = \"x\ty\" } }"),
		"screens/f.bbt": listScreen("select = { do = \"back\" }\n[[item]]\nkey = \"up\"\ndo = \"back\"\n" +
			"[[item]]\nkey = \"DOWN\"\ndo = \"back\"\n[[item]]\nkey = \"Enter\"\ndo = \"back\""),
		"content/parts/1.md": "# One\n",
	})
	// Each error has the position of its key path (screen-render, section 3.1).
	want := []string{
		`screens/a.bbt:2:10: select: the screen has no data`,
		`screens/b.bbt:4:10: select: do is missing`,
		`screens/c.bbt:4:17: select: do: unknown action "jump"`,
		`screens/d.bbt:4:32: select: args: only goto and gosub take arguments`,
		`screens/e.bbt:4:101: select: args: d: invalid token "{arg.dir}", want a value or {row.<field>}`,
		`screens/e.bbt:4:114: select: args: invalid argument name "X"`,
		`screens/e.bbt:4:136: select: args: e: value has a control character`,
		`screens/e.bbt:4:44: select: args: a: invalid token "{row.path", want a value or {row.<field>}`,
		`screens/e.bbt:4:61: select: args: b: provider "files" has no field "size"`,
		`screens/e.bbt:4:79: select: args: c: field "row.selected" is valid only in a condition`,
		`screens/f.bbt:12:7: item "ENTER": key: "ENTER" is a key of select`,
		`screens/f.bbt:6:7: item "UP": key: "UP" is a key of select`,
		`screens/f.bbt:9:7: item "DOWN": key: "DOWN" is a key of select`,
	}
	slices.Sort(msgs)
	if !slices.Equal(msgs, want) {
		t.Errorf("got\n%s\nwant\n%s", strings.Join(msgs, "\n"), strings.Join(want, "\n"))
	}
}

// The errors of select between files: the target, its arguments, a value for an argument of data
// and a value that the file screen cannot show (list-view, sections 8 and 10).
func TestSelectTargetErrors(t *testing.T) {
	sub := "+++\ndata = { from = \"files\", args = { dir = \"{arg.dir}\" } }\nrows = 2\n" +
		"[[item]]\nkey = \"Q\"\ndo = \"logoff\"\n+++\n{#each}{row.title}{/each}\n"
	tests := []struct {
		name, sel string
		more      map[string]string
		want      []string
	}{
		{"target", `select = { do = "gosub nope" }`, nil,
			[]string{`screens/sel.bbt:4:17: select: do: unknown screen "nope"`}},
		{"args", `select = { do = "gosub read", args = { x = "{row.path}" } }`, nil, []string{
			`screens/sel.bbt:4:38: select: args: missing argument "file" of screen "read"`,
			`screens/sel.bbt:4:38: select: args: screen "read" has no argument "x"`,
		}},
		{"data", `select = { do = "gosub sub", args = { dir = "{row.path}" } }`, map[string]string{"screens/sub.bbt": sub},
			[]string{`screens/sel.bbt:4:45: select: args: dir: screen "sub" uses the argument in data`}},
		{"file", `select = { do = "gosub read", args = { file = "{row.title}" } }`, nil, []string{
			`screens/sel.bbt:4:47: select: args: file: content file "One" does not exist`,
			`screens/sel.bbt:4:47: select: args: file: content file "Three" does not exist`,
			`screens/sel.bbt:4:47: select: args: file: content file "Two" does not exist`,
		}},
		// A literal value is checked also when the list has no rows.
		{"literal", `select = { do = "gosub read", args = { file = "missing.md" } }`, map[string]string{
			"screens/sel.bbt": strings.Replace(listScreen(`select = { do = "gosub read", args = { file = "missing.md" } }`),
				`dir = "parts"`, `dir = "empty"`, 1),
			"content/empty/.keep": "",
		}, []string{`screens/sel.bbt:4:47: select: args: file: content file "missing.md" does not exist`}},
		{"numbers", `select = { do = "gosub num", args = { file = "{row.path}" } }`, map[string]string{
			"screens/num.bbt": "+++\nfile = \"{arg.file}\"\nnumbers = true\n[[item]]\nkey = \"ESC\"\ndo = \"back\"\n+++\n{>file}\n",
		}, []string{
			`screens/sel.bbt:4:46: select: args: file: numbers: "parts/1.md" is a Markdown file`,
			`screens/sel.bbt:4:46: select: args: file: numbers: "parts/2.md" is a Markdown file`,
			`screens/sel.bbt:4:46: select: args: file: numbers: "parts/3.md" is a Markdown file`,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, msgs := loadLangs(t, nil, selectFiles(tt.sel, tt.more))
			slices.Sort(msgs)
			if !slices.Equal(msgs, tt.want) {
				t.Errorf("got\n%s\nwant\n%s", strings.Join(msgs, "\n"), strings.Join(tt.want, "\n"))
			}
		})
	}
}

// The file screen holds the content file of each row of the list of select.
// A literal value of select gives the rows of a list screen.
func TestSelectModel(t *testing.T) {
	sub := "+++\ndata = { from = \"files\", args = { dir = \"{arg.dir}\" } }\nrows = 2\n" +
		"select = { do = \"gosub read\", args = { file = \"{row.path}\" } }\n" +
		"[[item]]\nkey = \"Q\"\ndo = \"logoff\"\n+++\n{#each}{row.title}{/each}\n"
	m, msgs := loadLangs(t, nil, selectFiles(`select = { do = "gosub sub", args = { dir = "more" } }`, map[string]string{
		"screens/sub.bbt":    sub,
		"content/more/a.txt": "a\n",
	}))
	if m == nil {
		t.Fatalf("diags:\n%s", strings.Join(msgs, "\n"))
	}
	s := m.Sets["en"]
	if rows := s["sub"].List.List(map[string]string{"dir": "more"}); len(rows) != 1 || rows[0]["path"] != "more/a.txt" {
		t.Errorf("sub rows: %v", rows)
	}
	if got := s["read"].FileScreen.Content(map[string]string{"file": "more/a.txt"}); got == nil {
		t.Error("read has no content file more/a.txt")
	}
	sel := s["sel"].List.Select
	rows := s["sel"].List.List(nil)
	if a := sel.Action(rows[1], 2); a.Screen != "sub" || a.Args["dir"] != "more" {
		t.Errorf("sel action: %+v", a)
	}
}
