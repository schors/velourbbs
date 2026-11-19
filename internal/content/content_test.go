package content

import (
	"fmt"
	"io/fs"
	"maps"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/schors/velourbbs/internal/screen"
)

// load loads the file content/<p> with data.
func load(t *testing.T, p, data string) (*screen.Content, []string) {
	t.Helper()
	c, ds, err := Load(fstest.MapFS{"content/" + p: {Data: []byte(data)}}, "en", p)
	if err != nil {
		t.Fatalf("%q: %v", data, err)
	}
	var got []string
	for _, d := range ds {
		got = append(got, d.String())
	}
	if (c == nil) != (len(got) > 0) {
		t.Fatalf("%q: content %v with diagnostics %q", data, c != nil, got)
	}
	return c, got
}

func TestCheckPath(t *testing.T) {
	for _, p := range []string{"a", "gold.md", "xfer/src/crc16.c", "_x/A-b.c_d", "0"} {
		if err := CheckPath(p); err != nil {
			t.Errorf("%q: %v", p, err)
		}
	}
	for _, p := range []string{"", "/a", "a/", "a//b", ".a", "a/../b", "..", "-a", "a b", "a\\b", "я.md"} {
		if err := CheckPath(p); err == nil || err.Error() != fmt.Sprintf("invalid content path %q", p) {
			t.Errorf("%q: %v", p, err)
		}
	}
}

func TestKindOf(t *testing.T) {
	for p, want := range map[string]screen.ContentKind{
		"a.md": screen.Markdown, "a.MARKDOWN": screen.Markdown, "a.Md": screen.Markdown,
		"a.txt": screen.PlainText, "a.TXT": screen.PlainText, "README": screen.PlainText,
		"d.md/a": screen.PlainText, "a.c": screen.Source, "a.md.c": screen.Source, "a.go": screen.Source,
	} {
		if got := KindOf(p); got != want {
			t.Errorf("%q: %v, want %v", p, got, want)
		}
	}
}

func TestLoadFileErrors(t *testing.T) {
	fsys := fstest.MapFS{
		"content/dir/a.txt": {Data: []byte("a")},
		"content/fifo.txt":  {Mode: fs.ModeNamedPipe},
		"content/max.txt":   {Data: []byte(strings.Repeat("a", screen.MaxFileBytes))},
		"content/over.txt":  {Data: []byte(strings.Repeat("a", screen.MaxFileBytes+1))},
	}
	for p, want := range map[string]string{
		"none.md":  `content file "none.md" does not exist`,
		"dir":      `content file "dir": not a regular file`,
		"fifo.txt": `content file "fifo.txt": not a regular file`,
		"over.txt": `content file "over.txt": file is over 1048576 bytes`,
		"../a.txt": `invalid content path "../a.txt"`,
	} {
		if _, _, err := Load(fsys, "en", p); err == nil || err.Error() != want {
			t.Errorf("%s: %v, want %q", p, err, want)
		}
	}
	if c, ds, err := Load(fsys, "en", "max.txt"); err != nil || len(ds) > 0 || len(c.Lines) != 1 {
		t.Errorf("max: %v %v", ds, err)
	}
}

func TestLoadLines(t *testing.T) {
	for data, want := range map[string][]string{
		"":              nil,
		"\n":            {""},
		"a":             {"a"},
		"a\n":           {"a"},
		"a\n\n":         {"a", ""},
		"a\r\nb\nc":     {"a", "b", "c"},
		"\ufeffa\r\n":   {"a"},
		"\ta\t":         {"\ta\t"},
		"a\n\n\nb\n":    {"a", "", "", "b"},
		"\ufeff":        nil,
		"я\r\n\r\n":     {"я", ""},
		"x\u00e9\u0301": {"x\u00e9\u0301"},
	} {
		c, ds := load(t, "a.txt", data)
		if ds != nil {
			t.Errorf("%q: %q", data, ds)
			continue
		}
		if !slices.Equal(c.Lines, want) || len(c.Lines) != len(want) {
			t.Errorf("%q: %q, want %q", data, c.Lines, want)
		}
	}
}

func TestLoadContentErrors(t *testing.T) {
	g32 := "\u00e9" + strings.Repeat("\u0301", 15) // 32 bytes
	g33 := "a" + strings.Repeat("\u0301", 16)      // 33 bytes
	if _, ds := load(t, "a.txt", "яя"+g32); ds != nil {
		t.Errorf("32 bytes: %q", ds)
	}
	for data, want := range map[string][]string{
		"ab\nc\xffd":       {"content/a.c:2:2: invalid UTF-8"},
		"яя\nяя\xff":       {"content/a.c:2:3: invalid UTF-8"},
		"a\x07b":           {`content/a.c:1:2: control character U+0007`},
		"яяя\x07":          {`content/a.c:1:4: control character U+0007`},
		"a\rb\n":           {`content/a.c:1:2: control character U+000D`},
		"a\n\rb":           {`content/a.c:2:1: control character U+000D`},
		"a\r":              {`content/a.c:1:2: control character U+000D`},
		"a\r\r\n":          {`content/a.c:1:2: control character U+000D`},
		"\u0085 \u202e":    {`content/a.c:1:1: control character U+0085`, `content/a.c:1:3: control character U+202E`},
		"a\u200bb":         {`content/a.c:1:2: zero-width grapheme U+200B`},
		"\u0301a":          {`content/a.c:1:1: zero-width grapheme U+0301`},
		"яя" + g33:         {`content/a.c:1:3: grapheme over 32 bytes: 33 bytes`},
		"a\n\tb\x00\n\x1b": {`content/a.c:2:3: control character U+0000`, `content/a.c:3:1: control character U+001B`},
		"\ufeff\x7f":       {`content/a.c:1:1: control character U+007F`},
	} {
		if _, ds := load(t, "a.c", data); !slices.Equal(ds, want) {
			t.Errorf("%q: %q, want %q", data, ds, want)
		}
	}
	// Markdown: the content check comes before the front matter.
	if _, ds := load(t, "a.md", "+++\nx = \"\x07\"\n+++\n"); !slices.Equal(ds, []string{`content/a.md:2:6: control character U+0007`}) {
		t.Errorf("markdown: %q", ds)
	}
}

func TestFrontMatter(t *testing.T) {
	for _, tc := range []struct {
		data string
		meta map[string]string
		body string // the first body line
	}{
		{"# A\n", nil, "# A"},
		{"", nil, ""},
		{"+++\n+++\n", map[string]string{}, ""},
		{"+++\nx = \"1\"\n+++\nbody\n", map[string]string{"x": "1"}, "body"},
		{"+++  \nx = \"1\"\n+++ \nbody\n", map[string]string{"x": "1"}, "body"},
		{" +++\nx = \"1\"\n+++\n", nil, " +++"},
		// A "---" line that does not start a YAML front matter is Markdown.
		{"---\n\nx: 1\n---\n", nil, "---"},
		{"---\nx: 1\n", nil, "---"},
		{"----\nx: 1\n---\n", nil, "----"},
		{"\ufeff+++\r\nx = 1\r\n+++\r\nbody\r\n", map[string]string{"x": "1"}, "body"},
	} {
		c, ds := load(t, "a.md", tc.data)
		if ds != nil {
			t.Errorf("%q: %q", tc.data, ds)
			continue
		}
		if !maps.Equal(c.Meta, tc.meta) || (c.Meta == nil) != (tc.meta == nil) {
			t.Errorf("%q: meta %q, want %q", tc.data, c.Meta, tc.meta)
		}
		if first, _, _ := strings.Cut(string(c.Body), "\n"); first != tc.body {
			t.Errorf("%q: body %q, want first line %q", tc.data, c.Body, tc.body)
		}
	}
	// A plain text or source file has no front matter.
	for _, p := range []string{"a.txt", "a.c"} {
		if c, ds := load(t, p, "+++\nx = 1\n"); ds != nil || len(c.Lines) != 2 || c.Meta != nil {
			t.Errorf("%s: %q %q", p, c.Lines, ds)
		}
	}
}

func TestFrontMatterErrors(t *testing.T) {
	const yamlErr = `content/a.md:1:1: front matter: YAML is not supported, use TOML between "+++" lines`
	pad := func(n int) string { return "x = \"" + strings.Repeat("a", n-6) + "\"" } // n bytes
	if _, ds := load(t, "a.md", "+++\n"+pad(MaxFrontBytes)+"\n+++\n"); ds != nil {
		t.Errorf("64 KB: %q", ds)
	}
	for data, want := range map[string][]string{
		"+++\nx = 1\n": {`content/a.md:1:1: front matter: no end line "+++"`},
		"+++\n":        {`content/a.md:1:1: front matter: no end line "+++"`},
		"+++\n" + pad(MaxFrontBytes+1) + "\n+++\n": {`content/a.md:1:1: front matter: over 65536 bytes`},
		// TOML: the row in the file and the column in code points.
		"+++\ny = 1\nx = \"я\" z\n+++\n": {`content/a.md:3:9: front matter: expected newline but got U+007A 'z'`},
		"+++\nx = 1\nx = 2\n+++\n":       {`content/a.md:3:1: front matter: key x is already defined`},
		// YAML.
		"---\nx: 1\n---\nbody\n":       {yamlErr},
		"---  \nx: 1\n...  \nbody\n":   {yamlErr},
		"---\n \nx: 1\n---\n":          {yamlErr},
		"---\n---\nbody":               {yamlErr},
		"---\n# only a comment\n---\n": {yamlErr},
	} {
		if _, ds := load(t, "a.md", data); !slices.Equal(ds, want) {
			t.Errorf("%q:\n got %q\nwant %q", data, ds, want)
		}
	}
}

func TestMetaValues(t *testing.T) {
	toml := `+++
s = "str"
lit = 'x\y'
i = 0x1F
u = 1_000
neg = -7
f = +1.5e3
inf = -inf
b = true
d = 2024-01-02
dt = 1979-05-27T07:32:00.120
dto = 1979-05-27 07:32:00.50z
dto2 = 1979-05-27T07:32:00-07:00
t = 07:32:00.999
t2 = 07:32
t3 = 07:32:01
arr = [1]
it = {a = 1}
Upper = "x"
"q k" = "x"
"title" = "quoted"
a.b = 1
[tab]
x = 1
+++
`
	for data, want := range map[string]map[string]string{
		toml: {
			"s": "str", "lit": `x\y`, "i": "0x1F", "u": "1_000", "neg": "-7", "f": "+1.5e3", "inf": "-inf",
			"b": "true", "d": "2024-01-02", "dt": "1979-05-27T07:32:00.120",
			"dto": "1979-05-27T07:32:00.5Z", "dto2": "1979-05-27T07:32:00-07:00",
			"t": "07:32:00.999", "t2": "07:32:00", "t3": "07:32:01", "title": "quoted",
		},
	} {
		c, ds := load(t, "a.md", data)
		if ds != nil {
			t.Errorf("%q", ds)
			continue
		}
		if !maps.Equal(c.Meta, want) {
			t.Errorf("meta:\n got %q\nwant %q", c.Meta, want)
		}
	}
}

func TestTitle(t *testing.T) {
	for data, want := range map[string]string{
		"+++\ntitle = \"Meta\"\n+++\n# Head\n": "Meta",
		"+++\ntitle = 2024\n+++\n# Head\n":     "2024",
		"+++\ntitle = [\"a\"]\n+++\n# Head\n":  "Head",
		"+++\ntitle = \"\"\n+++\n# Head\n":     "",
		"# The *gold* `bug` &amp; [x](y)\n":    "The gold bug & x",
		"Gold\n====\n\n# Second\n":             "Gold",
		"## Sub\n":                             "a.md",
		"para\n\n## Sub\n\n# Late\n":           "Late",
		"#\n\n# B\n":                           "",
		"> # Quoted\n":                         "Quoted",
		"":                                     "a.md",
	} {
		c, ds := load(t, "dir/a.md", data)
		if ds != nil || c.Title != want {
			t.Errorf("%q: %q %q, want %q", data, c.Title, ds, want)
		}
	}
	for _, p := range []string{"dir/notes.txt", "x/y/crc16.c"} {
		if c, _ := load(t, p, "# Head\n"); c.Title != p[strings.LastIndexByte(p, '/')+1:] {
			t.Errorf("%s: %q", p, c.Title)
		}
	}
}

func TestWords(t *testing.T) {
	md := "+++\nauthor = \"Some One Else\"\n+++\n" +
		"# Title words\n\n" +
		"One [two](https://x.org \"a long title\") <https://y.org> **thr**ee ![img alt](i.png)\n\n" +
		"<!-- hidden words -->after\n\n" +
		"<div>in html</div>\n\n" +
		"```go info\ncode here\n```\n\n" +
		"| a | b |\n|---|---|\n| c d | e |\n\n" +
		"- item &amp; x\n- y\n\n" +
		"[ref]: https://z.org \"ref title\"\n"
	c, _ := load(t, "a.md", md)
	// 2 title, 6 paragraph, 1 after, 2 html, 2 code, 5 table, 4 list.
	if c.Words != 22 {
		t.Errorf("words: %d, want 22", c.Words)
	}
	for words, want := range map[int]int{0: 1, 1: 1, 200: 1, 201: 2, 400: 2, 401: 3} {
		c, _ := load(t, "a.md", strings.Repeat("w ", words))
		if c.Words != words || c.Minutes(0, 0) != want {
			t.Errorf("%d words: %d words, %d min, want %d", words, c.Words, c.Minutes(0, 0), want)
		}
	}
	// Plain text and source: the lines of the key lines.
	text := strings.Repeat("a b\tc  d\n", 50) + "e\n" + strings.Repeat("f\n", 199)
	for _, p := range []string{"a.txt", "a.c"} {
		c, _ := load(t, p, text)
		for lines, want := range map[[2]int]int{{0, 0}: 2, {1, 50}: 1, {1, 51}: 2, {51, 250}: 1, {51, 51}: 1} {
			if got := c.Minutes(lines[0], lines[1]); got != want {
				t.Errorf("%s %v: %d min, want %d", p, lines, got, want)
			}
		}
	}
}

// Load finds content/<lang>/<p>, else content/<p> (session-language, section 4).
func TestLoadLang(t *testing.T) {
	fsys := fstest.MapFS{
		"content/a.txt":       {Data: []byte("common\n")},
		"content/ru/a.txt":    {Data: []byte("русский\n")},
		"content/ru/bad.txt":  {Data: []byte("\xff\n")},
		"content/bad.txt":     {Data: []byte("ok\n")},
		"content/ru/link.txt": {Mode: fs.ModeSymlink, Data: []byte("missing")},
		"content/link.txt":    {Data: []byte("common\n")},
	}
	for lang, want := range map[string]string{"en": "common", "ru": "русский"} {
		if c, ds, err := Load(fsys, lang, "a.txt"); err != nil || len(ds) > 0 || c.Lines[0] != want {
			t.Errorf("%s: %v %v %v, want %q", lang, c, ds, err, want)
		}
	}
	if _, ds, _ := Load(fsys, "ru", "bad.txt"); len(ds) != 1 || ds[0].File != "content/ru/bad.txt" {
		t.Errorf("bad.txt: %v", ds)
	}
	// A broken symlink of the language is not a step to the common file.
	if _, _, err := Load(fsys, "ru", "link.txt"); err == nil || err.Error() != `content file "link.txt" does not exist` {
		t.Errorf("link.txt: %v", err)
	}
}
