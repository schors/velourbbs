package bbt_test

import (
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/schors/velourbbs/internal/bbt"
	"github.com/schors/velourbbs/internal/screen"
)

// loadFS loads the files, given by their paths in the BBS directory.
// A path that ends with "/" is a directory.
func loadFS(files map[string]string) (map[string]*screen.Screen, []string) {
	fsys := fstest.MapFS{}
	for name, data := range files {
		if dir, ok := strings.CutSuffix(name, "/"); ok {
			fsys[dir+"/x"] = &fstest.MapFile{}
			continue
		}
		fsys[name] = &fstest.MapFile{Data: []byte(data)}
	}
	screens, ds := bbt.Load(fsys, theme, screen.Layout{Width: defaultWidth}, aliases)
	var msgs []string
	for _, d := range ds {
		msgs = append(msgs, d.String())
	}
	return screens, msgs
}

// reader is a file screen with the front matter keys keys, the items of quit and the body.
func reader(keys, body string) string {
	return "+++\n" + keys + "\n[[item]]\nkey = \"Q\"\ndo = \"logoff\"\n+++\n" + body
}

// gotoReader is a screen with the item 1: a goto to reader with the argument file.
func gotoReader(file string) string {
	return "+++\n[[item]]\nkey = \"1\"\ndo = \"goto reader\"\nargs = { file = \"" + file + "\" }\n+++\nMenu\n"
}

// The errors of file-view, section 7, with their positions.
func TestFileErrors(t *testing.T) {
	gold := "# Gold\n\nText.\n"
	tests := []struct {
		name  string
		files map[string]string
		want  []string
	}{{
		name:  "literal file does not exist",
		files: map[string]string{"screens/reader.bbt": reader(`file = "gold.md"`, "{>file}\n")},
		want:  []string{`screens/reader.bbt:2:1: file: content file "gold.md" does not exist`},
	}, {
		name: "goto file does not exist",
		files: map[string]string{
			"screens/reader.bbt": reader(`file = "{arg.file}"`, "{>file}\n"),
			"screens/texts.bbt":  gotoReader("gild.md"),
		},
		want: []string{`screens/texts.bbt:2:1: item "1": args: file: content file "gild.md" does not exist`},
	}, {
		name: "goto invalid path",
		files: map[string]string{
			"screens/reader.bbt": reader(`file = "{arg.file}"`, "{>file}\n"),
			"screens/texts.bbt":  gotoReader("../gold.md"),
		},
		want: []string{`screens/texts.bbt:2:1: item "1": args: file: invalid content path "../gold.md"`},
	}, {
		name: "not a regular file",
		files: map[string]string{
			"screens/reader.bbt": reader(`file = "dir"`, "{>file}\n"),
			"content/dir/":       "",
		},
		want: []string{`screens/reader.bbt:2:1: file: content file "dir": not a regular file`},
	}, {
		name: "over the size limit",
		files: map[string]string{
			"screens/reader.bbt": reader(`file = "big.txt"`, "{>file}\n"),
			"content/big.txt":    strings.Repeat("a", screen.MaxFileBytes+1),
		},
		want: []string{`screens/reader.bbt:2:1: file: content file "big.txt": file is over 1048576 bytes`},
	}, {
		// An error in a content file appears once for the two screens that use it.
		name: "content error",
		files: map[string]string{
			"screens/reader.bbt": reader(`file = "{arg.file}"`, "{>file}\n"),
			"screens/texts.bbt":  gotoReader("a.txt"),
			"screens/plain.bbt":  reader(`file = "a.txt"`, "{>file}\n"),
			"content/a.txt":      "ok\nab\xff\n",
		},
		want: []string{`content/a.txt:2:3: invalid UTF-8`},
	}, {
		name: "front matter error",
		files: map[string]string{
			"screens/reader.bbt": reader(`file = "gold.md"`, "{>file}\n"),
			"content/gold.md":    "+++\ntitle = \"Gold\"\n",
		},
		want: []string{`content/gold.md:1:1: front matter: no end line "+++"`},
	}, {
		name: "lines and numbers of a Markdown file",
		files: map[string]string{
			"screens/code.bbt": reader("file = \"gold.md\"\nlines = [1, 2]", "{>file}\n"),
			"screens/num.bbt":  reader("file = \"gold.md\"\nnumbers = false", "{>file}\n"),
			"screens/both.bbt": reader("file = \"gold.md\"\nlines = [1, 2]\nnumbers = true", "{>file}\n"),
			"content/gold.md":  gold,
		},
		want: []string{
			`screens/both.bbt:2:1: lines: "gold.md" is a Markdown file`,
			`screens/both.bbt:2:1: numbers: "gold.md" is a Markdown file`,
			`screens/code.bbt:2:1: lines: "gold.md" is a Markdown file`,
			`screens/num.bbt:2:1: numbers: "gold.md" is a Markdown file`,
		},
	}, {
		name: "last after the end",
		files: map[string]string{
			"screens/code.bbt": reader("file = \"a.c\"\nlines = [2, 4]", "{>file}\n"),
			"content/a.c":      "1\n2\n3\n",
		},
		want: []string{`screens/code.bbt:2:1: lines: last line 4 is after the end of "a.c" (3 lines)`},
	}, {
		// The lines of the file screen apply to each file of a goto.
		name: "goto lines of a Markdown file",
		files: map[string]string{
			"screens/reader.bbt": reader("file = \"{arg.file}\"\nlines = [1, 1]\nnumbers = true", "{>file}\n"),
			"screens/code.bbt":   "+++\n[[item]]\nkey = \"2\"\ndo = \"goto reader\"\nargs = { file = \"gold.md\" }\n+++\nCode\n",
			"content/gold.md":    gold,
		},
		want: []string{
			`screens/code.bbt:2:1: item "2": args: file: lines: "gold.md" is a Markdown file`,
			`screens/code.bbt:2:1: item "2": args: file: numbers: "gold.md" is a Markdown file`,
		},
	}, {
		name: "goodbye as a file screen",
		files: map[string]string{
			"screens/goodbye.bbt": "+++\nfile = \"a.txt\"\n+++\n{>file}\n",
			"content/a.txt":       "a\n",
		},
		want: []string{`screens/goodbye.bbt:2:1: file: the goodbye screen cannot be a file screen`},
	}, {
		name: "file keys",
		files: map[string]string{
			"screens/a.bbt": reader(`file = "../gold.md"`, "{>file}\n"),
			"screens/b.bbt": reader(`file = "{arg.File}"`, "{>file}\n"),
			"screens/c.bbt": reader(`file = "{arg.file} "`, "{>file}\n"),
			"screens/d.bbt": reader(`file = "{term.type}"`, "{>file}\n"),
			"screens/e.bbt": reader("file = \"a.txt\"\nlines = [3, 2]", "{>file}\n"),
			"screens/f.bbt": reader("file = \"a.txt\"\nlines = [1]", "{>file}\n"),
			"screens/g.bbt": reader("file = \"a.txt\"\nlines = [0, 1]", "{>file}\n"),
			"screens/h.bbt": reader("lines = [1, 1]\nnumbers = false", "Text\n"),
			"screens/i.bbt": reader(`file = 1`, "{>file}\n"),
			"content/a.txt": "a\n",
		},
		want: []string{
			`screens/a.bbt:2:1: file: invalid content path "../gold.md"`,
			`screens/b.bbt:2:1: file: invalid token "{arg.File}", want a content path or {arg.<name>}`,
			`screens/c.bbt:2:1: file: invalid token "{arg.file} ", want a content path or {arg.<name>}`,
			`screens/d.bbt:2:1: file: invalid token "{term.type}", want a content path or {arg.<name>}`,
			`screens/e.bbt:2:1: lines: [3, 2] is not in 1 <= first <= last`,
			`screens/f.bbt:2:1: lines: must be an array of 2 integers`,
			`screens/g.bbt:2:1: lines: [0, 1] is not in 1 <= first <= last`,
			`screens/h.bbt:2:1: lines: the screen has no file`,
			`screens/h.bbt:2:1: numbers: the screen has no file`,
			`screens/i.bbt:2:8: file: must be a string`,
		},
	}, {
		name: "file area",
		files: map[string]string{
			"screens/a.bbt":     reader(`file = "a.txt"`, "Text\n"),
			"screens/b.bbt":     reader(`file = "a.txt"`, "{>file}\nText\n{>file}\n"),
			"screens/c.bbt":     reader(`file = "a.txt"`, "{#if term.cols > 40}\n{>file}\n{/if}\n{>file}\n"),
			"screens/d.bbt":     reader(`file = "a.txt"`, "{>_frag}\n{>file}\n"),
			"screens/_frag.bbt": "Head\n{>file}\n",
			"screens/e.bbt":     reader(``, "Text\n{>file}\n"),
			"screens/f.bbt":     reader(`file = "a.txt"`, "Text {>file}\n{>file}\n"),
			"content/a.txt":     "a\n",
		},
		want: []string{
			`screens/_frag.bbt:2:1: {>file} in a fragment`,
			`screens/a.bbt:2:1: file: the body has no {>file}`,
			`screens/b.bbt:9:1: second {>file}`,
			`screens/c.bbt:8:1: {>file} in a block`,
			`screens/e.bbt:8:1: {>file} in a screen without file`,
			`screens/f.bbt:7:6: an include or an art insert must be the only content of its line`,
		},
	}, {
		name: "page actions and fields outside a file screen",
		files: map[string]string{
			"screens/a.bbt": "+++\n[[item]]\nkey = \"B\"\ndo = \"page next\"\n" +
				"[[item]]\nkey = \"C\"\ndo = \"logoff\"\nwhen = \"file.page > 1\"\n" +
				"[[item]]\nkey = \"D\"\ndo = \"page up\"\n+++\nText\n",
			"screens/b.bbt":     "+++\n[[item]]\nkey = \"Q\"\ndo = \"logoff\"\n+++\n{file.title}\n{#if file.last}x{/if}\n{>_frag}\n{>art.ans}\n",
			"screens/_frag.bbt": "{meta.author}\n",
			"screens/art.ans":   sauceArt("{file.page}", 11),
		},
		want: []string{
			`screens/_frag.bbt:1:1: screen "b": field "meta.author" is valid only in a file screen`,
			`screens/a.bbt:2:1: item "B": do: "page next" is valid only in a file screen`,
			`screens/a.bbt:2:1: item "C": when: field "file.page" is valid only in a file screen`,
			`screens/a.bbt:2:1: item "D": do: invalid page action "page up"`,
			`screens/art.ans:1:1: screen "b": field "file.page" is valid only in a file screen`,
			`screens/b.bbt:6:1: field "file.title" is valid only in a file screen`,
			`screens/b.bbt:7:1: field "file.last" is valid only in a file screen`,
		},
	}, {
		name: "boolean field token",
		files: map[string]string{
			"screens/reader.bbt": reader(`file = "a.txt"`, "{file.last}\n{>file}\n"),
			"content/a.txt":      "a\n",
		},
		want: []string{`screens/reader.bbt:7:1: field "file.last" is boolean, valid only in a condition`},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			screens, msgs := loadFS(tt.files)
			if screens != nil || !slices.Equal(msgs, tt.want) {
				t.Errorf("got\n%s\nwant\n%s", strings.Join(msgs, "\n"), strings.Join(tt.want, "\n"))
			}
		})
	}
}

// A file screen holds each content file that it can show: the literal file,
// and the file of each goto. Two screens share one load.
func TestFileContents(t *testing.T) {
	screens, msgs := loadFS(map[string]string{
		"screens/reader.bbt": reader(`file = "{arg.file}"`, "{>file}\n"),
		"screens/texts.bbt": "+++\n[[item]]\nkey = \"1\"\ndo = \"goto reader\"\nargs = { file = \"a.txt\" }\n" +
			"[[item]]\nkey = \"2\"\ndo = \"goto reader\"\nargs = { file = \"b/c.md\" }\n+++\nMenu\n",
		"screens/code.bbt": reader("file = \"a.txt\"\nlines = [2, 3]\nnumbers = true", "Top\n{>file}\nBottom\n"),
		"content/a.txt":    "1\n2\n3\n",
		"content/b/c.md":   "# C\n",
	})
	if screens == nil {
		t.Fatalf("diags:\n%s", strings.Join(msgs, "\n"))
	}
	r := screens["reader"].FileScreen
	if r == nil || r.Arg != "file" || len(r.Contents) != 2 || r.Contents["b/c.md"].Title != "C" {
		t.Errorf("reader: %+v", r)
	}
	if !slices.Equal(screens["reader"].Args, []string{"file"}) {
		t.Errorf("reader args %q", screens["reader"].Args)
	}
	c := screens["code"].FileScreen
	if c == nil || c.Path != "a.txt" || c.First != 2 || c.Last != 3 || !c.Numbers || c.Contents["a.txt"] != r.Contents["a.txt"] {
		t.Errorf("code: %+v", c)
	}
	if got := show(screens["code"].Body); got != `"Top" / {>file} "Bottom" /` {
		t.Errorf("code body %q", got)
	}
	if screens["texts"].FileScreen != nil {
		t.Error("texts is a file screen")
	}
}
