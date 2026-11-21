package bbs

import (
	"maps"
	"slices"
	"strings"
	"testing"
)

// chainFiles gives a BBS where the values pass through two screens (bbt-format, section 8.5).
// main gives dir and file to hub. hub gives them on to side. side gives file to the file screen read,
// dir to the list screen parts and both back to hub: a cycle. more replaces or adds files.
func chainFiles(more map[string]string) map[string]string {
	files := map[string]string{
		"screens/main.bbt": "+++\n[[item]]\nkey = \"H\"\ndo = \"gosub hub\"\nargs = { dir = \"docs\", file = \"intro.md\" }\n" +
			"[[item]]\nkey = \"Q\"\ndo = \"logoff\"\n+++\n",
		"screens/hub.bbt": "+++\n[[item]]\nkey = \"S\"\ndo = \"gosub side\"\nargs = { dir = \"{arg.dir}\", file = \"{arg.file}\" }\n" +
			"[[item]]\nkey = \"ESC\"\ndo = \"back\"\n+++\n",
		"screens/side.bbt": "+++\n[[item]]\nkey = \"R\"\ndo = \"gosub read\"\nargs = { file = \"{arg.file}\" }\n" +
			"[[item]]\nkey = \"P\"\ndo = \"gosub parts\"\nargs = { dir = \"{arg.dir}\" }\n" +
			"[[item]]\nkey = \"H\"\ndo = \"goto hub\"\nargs = { dir = \"{arg.dir}\", file = \"{arg.file}\" }\n" +
			"[[item]]\nkey = \"ESC\"\ndo = \"back\"\n+++\n",
		"screens/read.bbt": "+++\nfile = \"{arg.file}\"\n[[item]]\nkey = \"ESC\"\ndo = \"back\"\n+++\n{>file}\n",
		"screens/parts.bbt": "+++\ndata = { from = \"files\", args = { dir = \"{arg.dir}\" } }\nrows = 2\n" +
			"select = { do = \"gosub read\", args = { file = \"{row.path}\" } }\n" +
			"[[item]]\nkey = \"ESC\"\ndo = \"back\"\n+++\n{#each}{row.title}{/each}\n",
		"content/intro.md":  "# Intro\n",
		"content/docs/1.md": "# One\n",
		"content/docs/2.md": "# Two\n",
	}
	maps.Copy(files, more)
	return files
}

// The load follows the chain and its cycle: the screens at the end hold the files and the rows of the values.
func TestChainModel(t *testing.T) {
	m, msgs := loadLangs(t, nil, chainFiles(nil))
	if m == nil {
		t.Fatalf("diags:\n%s", strings.Join(msgs, "\n"))
	}
	s := m.Sets["en"]
	if got := s["side"].Args; !slices.Equal(got, []string{"dir", "file"}) {
		t.Errorf("side args: %q", got)
	}
	if rows := s["parts"].List.List(map[string]string{"dir": "docs"}); len(rows) != 2 {
		t.Errorf("parts rows: %v", rows)
	}
	for _, p := range []string{"intro.md", "docs/1.md", "docs/2.md"} {
		if s["read"].FileScreen.Content(map[string]string{"file": p}) == nil {
			t.Errorf("read has no content file %s", p)
		}
	}
}

// An error of a value is an error at its origin, also through other screens (bbt-format, section 8.5).
func TestChainErrors(t *testing.T) {
	// mid gives each argument on to the screen sub.
	mid := func(arg, sub string) string {
		return "+++\n[[item]]\nkey = \"N\"\ndo = \"gosub " + sub + "\"\nargs = { " + arg + " = \"{arg." + arg + "}\" }\n" +
			"[[item]]\nkey = \"ESC\"\ndo = \"back\"\n+++\n"
	}
	parts := func(sel string) string {
		return "+++\ndata = { from = \"files\", args = { dir = \"{arg.dir}\" } }\nrows = 2\n" + sel +
			"\n[[item]]\nkey = \"ESC\"\ndo = \"back\"\n+++\n{#each}{row.title}{/each}\n"
	}
	tests := []struct {
		name  string
		langs []string
		more  map[string]string
		want  []string
	}{
		{"file", nil, map[string]string{
			"screens/main.bbt": "+++\n[[item]]\nkey = \"H\"\ndo = \"gosub hub\"\nargs = { dir = \"docs\", file = \"nope.md\" }\n" +
				"[[item]]\nkey = \"Q\"\ndo = \"logoff\"\n+++\n",
		}, []string{`screens/main.bbt:5:31: item "H": args: file: content file "nope.md" does not exist`}},
		{"rows", nil, map[string]string{
			"screens/main.bbt": "+++\n[[item]]\nkey = \"H\"\ndo = \"gosub hub\"\nargs = { dir = \"nope\", file = \"intro.md\" }\n" +
				"[[item]]\nkey = \"Q\"\ndo = \"logoff\"\n+++\n",
		}, []string{`screens/main.bbt:5:16: item "H": args: dir: directory "nope" does not exist`}},
		// A {row.<field>} value of a file screen through the screen mid.
		{"select file", nil, map[string]string{
			"screens/parts.bbt": parts(`select = { do = "gosub mid", args = { file = "{row.path}" } }`),
			"screens/mid.bbt":   mid("file", "num"),
			"screens/num.bbt":   "+++\nfile = \"{arg.file}\"\nnumbers = true\n[[item]]\nkey = \"ESC\"\ndo = \"back\"\n+++\n{>file}\n",
		}, []string{
			`screens/parts.bbt:4:46: select: args: file: numbers: "docs/1.md" is a Markdown file`,
			`screens/parts.bbt:4:46: select: args: file: numbers: "docs/2.md" is a Markdown file`,
		}},
		// A {row.<field>} value cannot give rows, also through the screen mid.
		{"select data", nil, map[string]string{
			"screens/parts.bbt": parts(`select = { do = "gosub mid", args = { dir = "{row.path}" } }`),
			"screens/mid.bbt":   mid("dir", "sub"),
			"screens/sub.bbt":   strings.Replace(parts(""), "rows = 2\n", "rows = 1\n", 1),
		}, []string{`screens/parts.bbt:4:45: select: args: dir: the value reaches the data of screen "sub"`}},
		// An open value {in} cannot reach file or data, also through the screen mid.
		{"in file", nil, map[string]string{
			"screens/main.bbt": "+++\n[[item]]\nkey = \"N\"\nask = \"File: \"\ndo = \"gosub mid\"\nargs = { file = \"{in}\" }\n" +
				"[[item]]\nkey = \"H\"\ndo = \"gosub hub\"\nargs = { dir = \"docs\", file = \"intro.md\" }\n+++\n",
			"screens/mid.bbt": mid("file", "read"),
		}, []string{`screens/main.bbt:6:17: item "N": args: file: the value reaches the file of screen "read"`}},
		{"in data", nil, map[string]string{
			"screens/main.bbt": "+++\n[[item]]\nkey = \"N\"\nask = \"Dir: \"\ndo = \"gosub mid\"\nargs = { dir = \"{in}\" }\n" +
				"[[item]]\nkey = \"H\"\ndo = \"gosub hub\"\nargs = { dir = \"docs\", file = \"intro.md\" }\n+++\n",
			"screens/mid.bbt": mid("dir", "parts"),
		}, []string{`screens/main.bbt:6:16: item "N": args: dir: the value reaches the data of screen "parts"`}},
		// Each set finds the error of a common select once (session-language, section 6).
		{"select data in each set", []string{"en", "ru"}, map[string]string{
			"screens/parts.bbt": parts(`select = { do = "gosub mid", args = { dir = "{row.path}" } }`),
			"screens/mid.bbt":   mid("dir", "sub"),
			"screens/sub.bbt":   strings.Replace(parts(""), "rows = 2\n", "rows = 1\n", 1),
		}, []string{`screens/parts.bbt:4:45: select: args: dir: the value reaches the data of screen "sub"`}},
		// Only the sub of ru has data: only the set ru finds the error.
		{"select data in one set", []string{"en", "ru"}, map[string]string{
			"screens/parts.bbt":  parts(`select = { do = "gosub mid", args = { dir = "{row.path}" } }`),
			"screens/mid.bbt":    mid("dir", "sub"),
			"screens/en/sub.bbt": "+++\n[[item]]\nkey = \"ESC\"\ndo = \"back\"\n+++\n{arg.dir}\n",
			"screens/ru/sub.bbt": strings.Replace(parts(""), "rows = 2\n", "rows = 1\n", 1),
		}, []string{`screens/parts.bbt:4:45: select: args: dir: the value reaches the data of screen "sub" (language "ru")`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, msgs := loadLangs(t, tt.langs, chainFiles(tt.more))
			slices.Sort(msgs)
			if !slices.Equal(msgs, tt.want) {
				t.Errorf("got\n%s\nwant\n%s", strings.Join(msgs, "\n"), strings.Join(tt.want, "\n"))
			}
		})
	}
}
