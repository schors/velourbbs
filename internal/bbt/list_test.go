package bbt_test

import (
	"slices"
	"strings"
	"testing"
)

// list is a list screen with the front matter keys keys, the item of quit and the body.
func list(keys, body string) string {
	return "+++\n" + keys + "\n[[item]]\nkey = \"Q\"\ndo = \"logoff\"\n+++\n" + body
}

// parts is the data and rows of a list screen of the directory parts.
const parts = "data = { from = \"files\", args = { dir = \"parts\" } }\nrows = 3"

// eachLine is a valid list line.
const eachLine = "{#each}{row.num} {row.title}{/each}\n"

// The errors of list-view, section 10, without select and art, with their positions.
func TestListErrors(t *testing.T) {
	part := map[string]string{"content/parts/a.md": "# A\n"}
	with := func(files map[string]string) map[string]string {
		for k, v := range part {
			files[k] = v
		}
		return files
	}
	tests := []struct {
		name  string
		files map[string]string
		want  []string
	}{{
		name: "keys",
		files: with(map[string]string{
			"screens/a.bbt": list("data = { args = { dir = \"parts\" } }\nrows = 3", eachLine),
			"screens/b.bbt": list("data = { from = \"file\", args = { dir = \"parts\" } }\nrows = 3", eachLine),
			"screens/c.bbt": list("data = { from = \"files\", args = { x = \"parts\" } }\nrows = 3", eachLine),
			"screens/d.bbt": list("data = { from = \"files\", args = { dir = \"../parts\" } }\nrows = 0", eachLine),
			"screens/e.bbt": list("data = { from = \"files\", args = { dir = \"{arg.dir\" } }\nrows = 101", eachLine),
			"screens/f.bbt": list("data = { from = \"files\", args = { dir = \"parts\" } }", eachLine),
			"screens/g.bbt": list("rows = 3", "Text\n"),
			"screens/h.bbt": list(parts+"\nfile = \"parts/a.md\"", eachLine+"{>file}\n"),
			"screens/i.bbt": list("data = { from = \"files\", args = { dir = 1 } }\nrows = 3", eachLine),
		}),
		want: []string{
			`screens/a.bbt:2:1: data: from is missing`,
			`screens/b.bbt:2:1: data: from: unknown provider "file"`,
			`screens/c.bbt:2:1: data: args: missing argument "dir"`,
			`screens/c.bbt:2:1: data: args: unknown argument "x"`,
			`screens/d.bbt:2:1: data: args: dir: invalid content path "../parts"`,
			`screens/d.bbt:2:1: rows: 0, must be from 1 to 100`,
			`screens/e.bbt:2:1: data: args: dir: invalid token "{arg.dir", want a value or {arg.<name>}`,
			`screens/e.bbt:2:1: rows: 101, must be from 1 to 100`,
			`screens/f.bbt:2:1: rows is missing`,
			`screens/g.bbt:2:1: rows: the screen has no data`,
			`screens/h.bbt:2:1: data: a list screen cannot have file`,
			`screens/i.bbt:2:41: data: invalid type`,
		},
	}, {
		name: "no list line and the end screens",
		files: with(map[string]string{
			"screens/a.bbt":        list(parts, "Text\n"),
			"screens/goodbye.bbt":  "+++\n" + parts + "\n+++\n" + eachLine,
			"screens/shutdown.bbt": "+++\n" + parts + "\n+++\n" + eachLine,
		}),
		want: []string{
			`screens/a.bbt:2:1: data: the screen has no list line`,
			`screens/goodbye.bbt:2:1: data: the goodbye screen cannot be a list screen`,
			`screens/shutdown.bbt:2:1: data: the shutdown screen cannot be a list screen`,
		},
	}, {
		name: "rows",
		files: map[string]string{
			"screens/a.bbt": list(parts, eachLine),
			"screens/b.bbt": list("data = { from = \"files\", args = { dir = \"{arg.dir}\" } }\nrows = 3", eachLine),
			"screens/c.bbt": "+++\n[[item]]\nkey = \"1\"\ndo = \"goto b\"\nargs = { dir = \"guide/bbtt\" }\n" +
				"[[item]]\nkey = \"2\"\ndo = \"goto b\"\nargs = { dir = \"x y\" }\n" +
				"[[item]]\nkey = \"3\"\ndo = \"goto b\"\nargs = { dir = \"bad\" }\n" +
				"[[item]]\nkey = \"4\"\ndo = \"goto b\"\nargs = { dir = \"big\" }\n+++\nMenu\n",
			"screens/d.bbt":    list("data = { from = \"files\", args = { dir = \"big\" } }\nrows = 3", eachLine),
			"content/bad/a.md": "+++\ntitle = \"A\"\n",
			// A whole row file with an error: an error of the key data or of the item.
			"content/big/a.md": strings.Repeat("x", 1<<20+1),
		},
		want: []string{
			`content/bad/a.md:1:1: front matter: no end line "+++"`,
			`screens/a.bbt:2:1: data: args: dir: directory "parts" does not exist`,
			`screens/c.bbt:2:1: item "1": args: dir: directory "guide/bbtt" does not exist`,
			`screens/c.bbt:2:1: item "2": args: dir: invalid content path "x y"`,
			`screens/c.bbt:2:1: item "4": args: dir: content file "big/a.md": file is over 1048576 bytes`,
			`screens/d.bbt:2:1: data: args: dir: content file "big/a.md": file is over 1048576 bytes`,
		},
	}, {
		name: "list lines",
		files: with(map[string]string{
			"screens/a.bbt": list(parts, "{#each}{row.num}\n"+
				"x {#each}{row.num}{/each}\n"+
				"{#each}{row.num}{/each} x\n"+
				"{#each}{#each}{/each}\n"+
				"{#each}{#if row.num > 1}x{/each}\n"+
				"{#if term.cols > 40}\n{#each}{/if}{/each}\n{/if}\n"+
				"{#if term.cols > 40}\n{#each}{#else}{/each}\n{/if}\n"+
				eachLine),
			"screens/b.bbt": list("", eachLine),
			"screens/c.bbt": list(parts, "{#each}{>_frag}{/each}\n"+eachLine),
		}),
		want: []string{
			`screens/a.bbt:8:1: {#each} has no {/each} at the end of its line`,
			`screens/a.bbt:9:3: {#each} is valid only at the start of a list line`,
			`screens/a.bbt:9:10: field "row.num" is valid only in a list line`,
			`screens/a.bbt:9:19: {/each} is valid only at the end of a list line`,
			`screens/a.bbt:10:1: {#each} has no {/each} at the end of its line`,
			`screens/a.bbt:11:8: {#each} is valid only at the start of a list line`,
			`screens/a.bbt:12:8: {#if} has no {/if} in its list line`,
			`screens/a.bbt:14:8: {/if} without {#if}`,
			`screens/a.bbt:17:8: {#else} without {#if}`,
			`screens/b.bbt:7:1: list line in a screen without data`,
			`screens/c.bbt:8:8: an include or an art insert must be the only content of its line`,
		},
	}, {
		name: "fields",
		files: with(map[string]string{
			"screens/a.bbt": list(parts, "{row.title}\n{#if row.num > 1}x{/if}\n{list.page}\n"+
				"{#each}{row.size}{/each}\n{#each}{#if row.selected}x{/if}{/each}\n{#each}{row.selected}{/each}\n"+
				"{#each}{#if row.size == 1}x{/if}{/each}\n{>_frag}\n{>art.ans}\n{list.last}\n{list.empty}\n"),
			"screens/_frag.bbt": "{row.path}\n{#each}{row.size}{/each}\n",
			"screens/art.ans":   sauceArt("{row.num} {list.pages}", 22),
			"screens/b.bbt": "+++\n[[item]]\nkey = \"A\"\ndo = \"logoff\"\nwhen = \"list.last\"\n" +
				"[[item]]\nkey = \"B\"\ndo = \"logoff\"\nwhen = \"row.num > 1\"\n" +
				"[[item]]\nkey = \"C\"\ndo = \"logoff\"\n+++\nText\n",
			// The fields of a list screen in the text of a screen without data.
			"screens/c.bbt": "+++\n[[item]]\nkey = \"Q\"\ndo = \"logoff\"\n+++\n{list.page}\n{#if list.empty}x{/if}\n",
		}),
		want: []string{
			`screens/_frag.bbt:1:1: field "row.path" is valid only in a list line`,
			`screens/_frag.bbt:2:8: screen "a": field "row.size": provider "files" has no field "size"`,
			`screens/a.bbt:8:1: field "row.title" is valid only in a list line`,
			`screens/a.bbt:9:1: field "row.num" is valid only in a list line`,
			`screens/a.bbt:11:8: field "row.size": provider "files" has no field "size"`,
			`screens/a.bbt:12:8: field "row.selected" is valid only in a screen with select`,
			`screens/a.bbt:13:8: field "row.selected" is boolean, valid only in a condition`,
			`screens/a.bbt:14:8: field "row.size": provider "files" has no field "size"`,
			`screens/a.bbt:17:1: field "list.last" is boolean, valid only in a condition`,
			`screens/a.bbt:18:1: field "list.empty" is boolean, valid only in a condition`,
			`screens/art.ans:1:1: screen "a": field "row.num" is valid only in a list line`,
			`screens/b.bbt:2:1: item "A": when: field "list.last" is valid only in a list screen`,
			`screens/b.bbt:2:1: item "B": when: field "row.num" is valid only in a list line`,
			`screens/c.bbt:6:1: field "list.page" is valid only in a list screen`,
			`screens/c.bbt:7:1: field "list.empty" is valid only in a list screen`,
		},
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

// A list screen holds the rows of its literal value and of each value of a goto.
// It has the page actions and the list.* fields. A list line in a block and in a fragment.
func TestListModel(t *testing.T) {
	screens, msgs := loadFS(map[string]string{
		"screens/a.bbt": list(parts+"\n[[item]]\nkey = \"N\"\ndo = \"page next\"\nwhen = \"not list.last\"",
			"{list.page}/{list.pages}\n{#if term.cols > 40}\n"+eachLine+"{#else}\n{>_frag}\n{/if}\n"),
		"screens/_frag.bbt": "{#each}{#if row.num == 1}{!title}{/if}{row.minutes   }{/each}\n",
		"screens/b.bbt":     list("data = { from = \"files\", args = { dir = \"{arg.dir}\" } }\nrows = 2", eachLine),
		"screens/c.bbt": "+++\n[[item]]\nkey = \"1\"\ndo = \"goto b\"\nargs = { dir = \"parts\" }\n" +
			"[[item]]\nkey = \"2\"\ndo = \"goto b\"\nargs = { dir = \"more\" }\n+++\nMenu\n",
		"content/parts/1.md": "# One\n",
		"content/parts/2.md": "# Two\n",
		"content/more/x.txt": "x\n",
	})
	if screens == nil {
		t.Fatalf("diags:\n%s", strings.Join(msgs, "\n"))
	}
	a := screens["a"]
	if a.List == nil || a.List.From != "files" || a.List.Rows != 3 || len(a.List.ArgRefs) != 0 {
		t.Fatalf("a: %+v", a.List)
	}
	rows := a.List.List(nil)
	if len(rows) != 2 || rows[0]["title"] != "One" || rows[1]["path"] != "parts/2.md" {
		t.Errorf("a rows: %v", rows)
	}
	want := `{list.page} "/" {list.pages} / if(/b list({row.num} " " {row.title}))else(/b screens/_frag.bbt(list(if(!title)else() {row.minutes:16}))) /b`
	if got := show(a.Body); got != want {
		t.Errorf("a body\n%s\nwant\n%s", got, want)
	}
	b := screens["b"]
	if !slices.Equal(b.Args, []string{"dir"}) || len(b.List.Lists) != 2 {
		t.Fatalf("b: args %q, lists %v", b.Args, b.List.Lists)
	}
	if rows := b.List.List(map[string]string{"dir": "more"}); len(rows) != 1 || rows[0]["title"] != "x.txt" {
		t.Errorf("b rows of more: %v", rows)
	}
	if b.List.List(map[string]string{"dir": "none"}) != nil {
		t.Error("b has rows of a value that no goto gives")
	}
	if screens["c"].List != nil {
		t.Error("c is a list screen")
	}
}
