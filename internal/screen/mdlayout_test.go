package screen_test

import (
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/schors/velourbbs/internal/content"
	"github.com/schors/velourbbs/internal/screen"
)

// mdLetters names the styles of mdTheme with one letter each.
var mdLetters = map[string]byte{
	"normal": 'n', "md_h1": '1', "md_h2": '2', "md_h3": '3', "md_strong": 's', "md_em": 'e',
	"md_code": 'c', "md_link": 'l', "md_mark": 'm', "md_frame": 'f', "md_dim": 'd',
}

// mdTheme has a different style for each name of mdLetters.
var mdTheme = func() map[string]screen.Style {
	t := map[string]screen.Style{}
	i := 0
	for name := range mdLetters {
		t[name] = screen.Style{Fg: screen.IndexColour(uint8(i))}
		i++
	}
	return t
}()

// loadMD loads the Markdown text src as a content file.
func loadMD(t *testing.T, src string) *screen.Content {
	t.Helper()
	fsys := fstest.MapFS{"content/a.md": {Data: []byte(src)}}
	c, diags, err := content.Load(fsys, "a.md")
	if err != nil || len(diags) > 0 {
		t.Fatalf("load: %v %v", err, diags)
	}
	return c
}

// cellText gives the graphemes of cells, and the style letters after "/" when styles is true.
func cellText(cells []screen.Cell, styles bool) string {
	var b strings.Builder
	for _, c := range cells {
		b.WriteString(c.Grapheme)
	}
	if styles {
		b.WriteByte('/')
		for _, c := range cells {
			letter := byte('?')
			for name, st := range mdTheme {
				if st == c.Style {
					letter = mdLetters[name]
				}
			}
			b.WriteByte(letter)
		}
	}
	return b.String()
}

// mdRows lays out src in w cells and gives the rows as text, with the style letters
// when styles is true.
func mdRows(t *testing.T, src string, w int, styles bool) []string {
	t.Helper()
	var out []string
	for row := range screen.MarkdownLayout(loadMD(t, src), w, mdTheme)(0) {
		out = append(out, cellText(row.Cells, styles))
	}
	return out
}

func TestMarkdownBlocks(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		w         int
		want      []string
	}{
		{"paragraph wraps", "aaa bbb ccc", 7, []string{"aaa bbb", "ccc"}},
		{"soft break is a space", "a\nb", 10, []string{"a b"}},
		{"hard break", "a  \nb\\\nc", 10, []string{"a", "b", "c"}},
		{"tab in text is a space", "a\tb", 10, []string{"a b"}},
		{"paragraphs", "a\n\nb", 10, []string{"a", "", "b"}},
		{"bullet levels", "- a\n- b\n  - c\n    - d", 30,
			[]string{"  ∙ a", "  ∙ b", "      - c", "          - d"}},
		{"bullet text wraps at the text", "- aaa bbb", 8, []string{"  ∙ aaa", "    bbb"}},
		{"ordered numbers align", "9. a\n10. b", 30, []string{"   9. a", "  10. b"}},
		{"ordered from zero", "0. a\n1. b", 30, []string{"  0. a", "  1. b"}},
		{"ordered nested in ordered", "1. a\n   1. b", 30, []string{"  1. a", "       1. b"}},
		{"loose list", "- a\n\n- b", 30, []string{"  ∙ a", "", "  ∙ b"}},
		{"loose item with two blocks", "- a\n\n  b", 30, []string{"  ∙ a", "    ", "    b"}},
		{"tight list in a loose item", "- a\n  - b\n  - c\n\n  d", 30,
			[]string{"  ∙ a", "    ", "      - b", "      - c", "    ", "    d"}},
		{"tight item with two blocks", "- a\n  ```\n  x\n  ```\n- b", 30,
			[]string{"  ∙ a", "    │ x", "  ∙ b"}},
		{"empty item", "-\n- b", 30, []string{"  ∙ ", "  ∙ b"}},
		{"quote", "> a\n>\n> b", 30, []string{"│ a", "│ ", "│ b"}},
		{"nested quote", "> a\n>\n> > b", 30, []string{"│ a", "│ ", "│ │ b"}},
		{"quote in an item", "- > q\n  > r", 30, []string{"  ∙ │ q r"}},
		{"list in a quote", "> - a\n> - b", 30, []string{"│   ∙ a", "│   ∙ b"}},
		{"empty quote", "a\n\n>\n\nb", 30, []string{"a", "", "b"}},
		{"code block", "```\nx\ty\n\nz\n```", 30, []string{"│ x       y", "│ ", "│ z"}},
		{"indented code", "    a\n    b", 30, []string{"│ a", "│ b"}},
		{"empty code block", "a\n\n```\n```\n\nb", 30, []string{"a", "", "b"}},
		{"code cut", "```\nabcdefghij\n```", 8, []string{"│ abc..."}},
		{"thematic break", "***", 30, []string{"∙   ∙   ∙"}},
		{"thematic break cut", "***", 5, []string{"∙   ∙"}},
		{"html block", "<div>\n a  b\n</div>", 30, []string{"<div>", " a  b", "</div>"}},
		{"html comment", "a\n\n<!-- c -->\n\nb", 30, []string{"a", "", "b"}},
		{"html block wraps", "<div>\naaa bbb\n</div>", 6, []string{"<div>", "aaa", "bbb", "</div>"}},
		{"html comments after a comment", "<!-- a --> b <!-- c --><!---->\n\nd", 30, []string{" b ", "", "d"}},
		{"html comments only", "<!-- a --> <!-- b -->\n\nd", 30, []string{"d"}},
		{"html comment with spaces after", "<!-- c -->  \n\nb", 30, []string{"b"}},
		{"text after an html comment", "<!-- c -->x", 30, []string{"x"}},
		{"link reference definition", "[r]: http://x\n\na", 30, []string{"a"}},
		{"paragraph of a comment", "a\n\n<span><!-- c --></span>\n\nb", 30,
			[]string{"a", "", "<span></span>", "", "b"}},
		{"link with only a note", "[](https://x)\n\nb", 30, []string{"[1]", "", "b", "", "[1] https://x"}},
		{"badge", "[![](https://i)](https://l)", 30, []string{"[1][2]", "", "[1] https://i", "[2] https://l"}},
		{"hard break between no cells", "![](a.png)  \n![](b.png)\n\nc", 30, []string{"c"}},
		{"hard break before no cells", "a\\\n![](b.png)\n\nc", 30, []string{"a", "", "c"}},
		{"soft break after no cells", "![](a.png)\nb", 30, []string{"b"}},
		{"two hard breaks", "a\\\n\\\nb", 30, []string{"a", "", "b"}},
		{"empty heading", "a\n\n#\n\nb", 30, []string{"a", "", "b"}},
		{"paragraph with no cells", "a\n\n&#x200B;\n\nb", 30, []string{"a", "", "b"}},
		{"prefix at most W / 2", "- a\n  - b\n    - c", 10, []string{"  ∙ a", "   - b", "   - c"}},
		{"prefix at W 1", "> - a", 1, []string{"a"}},
		{"character references", "&amp; a&#9;b &#x1b;c", 30, []string{"& a b c"}},
	} {
		if got := mdRows(t, tc.src, tc.w, false); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s:\n got %q\nwant %q", tc.name, got, tc.want)
		}
	}
}

func TestMarkdownStyles(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		want      []string
	}{
		{"headings", "# a\n\n## b\n\n### c\n\n###### d", []string{"a/1", "", "b/2", "", "c/3", "", "d/3"}},
		{"bullet markers", "- a\n  - b", []string{"  ∙ a/nnmnn", "      - b/nnnnnnfnn"}},
		{"ordered markers", "1. a\n   1. b", []string{"  1. a/nnmmnn", "       1. b/nnnnnnnffnn"}},
		{"quote", "> a *b*", []string{"│ a b/mneee"}},
		{"heading in a quote", "> # a", []string{"│ a/mn1"}},
		{"code", "```\nx\n```", []string{"│ x/fnc"}},
		{"code cut mark", "```\nabcdefghijklmnopqrstuvwxyz0123456789\n```",
			[]string{"│ abcdefghijklmnopqrstuvwxy.../fn" + strings.Repeat("c", 25) + "ddd"}},
		{"thematic break", "***", []string{"∙   ∙   ∙/mmmmmmmmm"}},
		{"html block in a quote", "> <div>", []string{"│ <div>/mnnnnnn"}},
		{"strong, em, code", "**a** *b* `c`", []string{"a b c/snenc"}},
		{"innermost", "**a *b* c**", []string{"a b c/ssess"}},
		{"link and note", "[a **b**](https://x)", []string{"a b[1]/llsddd", "", "[1] https://x/" + strings.Repeat("d", 13)}},
		{"image", "![a](https://x)", []string{"a[1]/lddd", "", "[1] https://x/" + strings.Repeat("d", 13)}},
		{"local link has no note", "[a](b.md)", []string{"a/l"}},
		{"autolink", "<https://x>", []string{"https://x/lllllllll"}},
		{"code span in a link", "[`a`](b.md)", []string{"a/c"}},
		{"inline html", "a <b>c</b> <!-- d -->e", []string{"a <b>c</b> e/" + strings.Repeat("n", 12)}},
		{"heading with a link", "# [a](b.md) c", []string{"a c/l11"}},
	} {
		got := mdRows(t, tc.src, 30, true)
		for i := range got {
			got[i] = strings.TrimSuffix(got[i], "/") // an empty row
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s:\n got %q\nwant %q", tc.name, got, tc.want)
		}
	}
}

func TestMarkdownTable(t *testing.T) {
	src := "| a | bb | c |\n|:-:|--:|---|\n| ccc | d |\n| `x` | [y](https://y) | z |"
	want := []string{
		" a     bb  c/" + strings.Repeat("s", 12),
		strings.Repeat("─", 12) + "/" + strings.Repeat("f", 12),
		"ccc     d   /" + strings.Repeat("n", 12),
		" x   y[1]  z/ncnnnldddnnn",
		"/",
		"[1] https://y/" + strings.Repeat("d", 13),
	}
	got := mdRows(t, src, 30, true)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("table:\n got %q\nwant %q", got, want)
	}
	// A row wider than the space ends with the cut mark.
	got = mdRows(t, "| abcdef | g |\n|---|---|\n| h | i |", 8, false)
	want = []string{"abcde...", "─────...", "h    ..."}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("cut table:\n got %q\nwant %q", got, want)
	}
	// The table in a quote: the header style is innermost, the body is in md_em.
	got = mdRows(t, "> | a |\n> |---|\n> | b |", 30, true)
	want = []string{"│ a/mns", "│ ─/mnf", "│ b/mne"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("table in a quote:\n got %q\nwant %q", got, want)
	}
}

func TestMarkdownNotes(t *testing.T) {
	src := "[a](https://x) [b](mailto:q) [c](https://x) [d](o.md) [e](c:x)"
	want := []string{"a[1] b[2] c[1] d e", "", "[1] https://x", "[2] mailto:q"}
	if got := mdRows(t, src, 30, false); !reflect.DeepEqual(got, want) {
		t.Errorf("notes:\n got %q\nwant %q", got, want)
	}
	// The address wraps after "[n] ".
	got := mdRows(t, "[a](https://example.com)", 10, false)
	want = []string{"a[1]", "", "[1] https:", "    //exam", "    ple.co", "    m"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("note wrap:\n got %q\nwant %q", got, want)
	}
	// "[n] " wider than W / 2 is text: the address wraps from the start of the row.
	got = mdRows(t, "[a](https://x.y)", 6, false)
	want = []string{"a[1]", "", "[1]", "https:", "//x.y"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("note wrap at small W:\n got %q\nwant %q", got, want)
	}
}

// flagText gives the flags of a row: "B" Break, "K" Keep, "S" Sep, "#<section>" NewSection.
func flagText(r screen.FileRow) string {
	var b strings.Builder
	if r.Break {
		b.WriteByte('B')
	}
	if r.Keep {
		b.WriteByte('K')
	}
	if r.Sep {
		b.WriteByte('S')
	}
	if r.NewSection {
		b.WriteString("#" + r.Section)
	}
	return b.String()
}

func TestMarkdownFlags(t *testing.T) {
	src := "a\n\n# H **1**\n\nb\n\n## H2 ccc\n\n### H3\n\n- # x"
	var got []string
	for row := range screen.MarkdownLayout(loadMD(t, src), 5, mdTheme)(0) {
		got = append(got, cellText(row.Cells, false)+":"+flagText(row))
	}
	want := []string{
		"a:", ":S", "H 1:BK#H 1", ":S", "b:", ":S",
		"H2:K#H2 ccc", "ccc:K", ":S", "H3:K", ":S", "∙ x:BK#x",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("flags:\n got %q\nwant %q", got, want)
	}
}

// mdPages gives the pages of src at w and h: the rows of each page joined with "|",
// the section after "@".
func mdPages(t *testing.T, src string, w, h int) []string {
	t.Helper()
	l := screen.MarkdownLayout(loadMD(t, src), w, mdTheme)
	var out []string
	for _, p := range screen.SplitPages(l, h) {
		var rows []string
		for _, r := range screen.PageRows(l, p, h) {
			rows = append(rows, cellText(r, false))
		}
		out = append(out, strings.Join(rows, "|")+"@"+p.Section)
	}
	return out
}

func TestMarkdownPages(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		h         int
		want      []string
	}{
		{"level-1 heading starts a page", "a\n\n# B\n\nc", 5, []string{"a||||@", "B||c||@B"}},
		{"level-1 heading at the first row", "# A\n\nb", 5, []string{"A||b||@A"}},
		{"heading moves with the next row", "a\n\nb\n\n## C\n\nd", 4,
			[]string{"a||b|@", "C||d|@C"}},
		{"heading longer than the rest stays", "a\n\n## C\n\nd", 2, []string{"a|@", "C|@C", "d|@C"}},
		{"no page starts with a separator", "a\n\nb\n\nc", 2, []string{"a|@", "b|@", "c|@"}},
		{"section of a level-3 heading", "## A\n\nb\n\n### C\n\nd", 3, []string{"A||b@A", "C||d@A"}},
		{"section of an empty heading", "## A\n\nb\n\n##\n\nc", 2, []string{"A|@A", "b|@A", "c|@A"}},
		{"section before a heading", "a\n\n## B", 1, []string{"a@", "B@B"}},
	} {
		if got := mdPages(t, tc.src, 10, tc.h); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s:\n got %q\nwant %q", tc.name, got, tc.want)
		}
	}
}

// allElements is a Markdown file with each element of file-view, section 5.
const allElements = `+++
title = "T"
+++
# One **two** three four

Text with a [link](https://example.com/a) and ![image](https://example.com/b.png),
` + "`code`" + `, *em*, <https://auto.link>, <b>html</b> <!-- gone -->.
A hard break\
here.

## Lists

- tight one
- tight two
  1. nested
  2. nested two

- loose

  two blocks

> quote line one
> quote line two
>
> - list in a quote

` + "```go\nfunc main() {\n\tprintln(\"a very long line of code that is cut\")\n}\n```" + `

***

| left | center | right |
|:-----|:------:|------:|
| a | b | c |
| longer cell | x | [y](https://example.com/a) |

<div>
html block
</div>

# Second part

[ref]: https://example.com/r

Last [ref] paragraph.
`

// TestMarkdownRestart checks that each page laid out from its restart point gives
// the rows of one run from the start.
func TestMarkdownRestart(t *testing.T) {
	c := loadMD(t, allElements)
	for _, w := range []int{1, 2, 5, 13, 40, 80} {
		l := screen.MarkdownLayout(c, w, mdTheme)
		type pos struct{ from, skip int }
		var all []string
		at := map[pos]int{}
		last := pos{from: -1}
		for row := range l(0) {
			p := pos{from: row.From}
			if row.From == last.from {
				p.skip = last.skip + 1
			}
			at[p], last = len(all), p
			all = append(all, cellText(row.Cells, true))
		}
		for h := 1; h <= 9; h++ {
			for _, p := range screen.SplitPages(l, h) {
				start, ok := at[pos{p.From, p.Skip}]
				if !ok {
					t.Fatalf("w %d h %d: page %+v not in the rows", w, h, p)
				}
				rows := screen.PageRows(l, p, h)
				for i := range p.N {
					if got := cellText(rows[i], true); got != all[start+i] {
						t.Errorf("w %d h %d page %+v row %d: %q, want %q", w, h, p, i, got, all[start+i])
					}
				}
			}
		}
	}
}

// TestMarkdownWidth checks that no row is wider than W.
func TestMarkdownWidth(t *testing.T) {
	c := loadMD(t, allElements)
	for w := 1; w <= 40; w++ {
		for row := range screen.MarkdownLayout(c, w, mdTheme)(0) {
			if len(row.Cells) > w {
				t.Errorf("w %d: row %q has %d cells", w, cellText(row.Cells, false), len(row.Cells))
			}
		}
	}
}
