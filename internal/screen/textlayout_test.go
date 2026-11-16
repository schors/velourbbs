package screen

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// fileTheme has the styles of a text layout. The letters of styleText name them.
var fileTheme = map[string]Style{
	"normal": {Fg: IndexColour(7)}, "md_code": {Fg: IndexColour(10)}, "md_dim": {Fg: IndexColour(8)},
}

// styleText gives a letter for each cell of r: "n" normal, "c" md_code, "d" md_dim.
func styleText(r []Cell) string {
	var b strings.Builder
	for _, c := range r {
		switch c.Style {
		case fileTheme["normal"]:
			b.WriteByte('n')
		case fileTheme["md_code"]:
			b.WriteByte('c')
		case fileTheme["md_dim"]:
			b.WriteByte('d')
		default:
			b.WriteByte('?')
		}
	}
	return b.String()
}

// rowText gives the graphemes of r.
func rowText(r []Cell) string {
	var b strings.Builder
	for _, c := range r {
		b.WriteString(c.Grapheme)
	}
	return b.String()
}

// textRows lays out the lines of kind and gives the rows as text, with the style
// letters after "/" when styles is true.
func textRows(kind ContentKind, lines []string, first, last int, numbers bool, w int, styles bool) []string {
	c := &Content{Kind: kind, Lines: lines}
	var out []string
	for row := range TextLayout(c, first, last, numbers, w, fileTheme)(0) {
		s := rowText(row.Cells)
		if styles {
			s += "/" + styleText(row.Cells)
		}
		out = append(out, s)
	}
	return out
}

func TestTextLayoutPlain(t *testing.T) {
	for _, tc := range []struct {
		name string
		line string
		w    int
		want []string
	}{
		{"wrap at a space", "aaa bbb ccc", 10, []string{"aaa bbb", "ccc"}},
		{"long word", "abcdefghijkl", 10, []string{"abcdefghij", "kl"}},
		{"wide grapheme at the end", "abcdefghi世", 10, []string{"abcdefghi", "世"}},
		{"spaces at the start", "  ab cd", 5, []string{"  ab", "cd"}},
		{"spaces at the end", "ab   ", 10, []string{"ab   "}},
		{"empty line", "", 10, []string{""}},
		{"line of spaces", "   ", 10, []string{"   "}},
		{"tab at column 0", "\tx", 20, []string{"        x"}},
		{"tab at column 3", "abc\tx", 20, []string{"abc     x"}},
		{"tab at column 7", "abcdefg\tx", 20, []string{"abcdefg x"}},
		{"tab at column 8", "abcdefgh\tx", 20, []string{"abcdefgh        x"}},
		{"tab before the wrap", "abcd\tefgh", 10, []string{"abcd", "efgh"}},
		{"tab after a wide grapheme", "世\tx", 20, []string{"世      x"}},
	} {
		got := textRows(PlainText, []string{tc.line}, 0, 0, false, tc.w, false)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: %q, want %q", tc.name, got, tc.want)
		}
	}
	if got := textRows(PlainText, []string{"ab c"}, 0, 0, false, 10, true); got[0] != "ab c/nnnn" {
		t.Errorf("style: %q", got)
	}
}

func TestTextLayoutSource(t *testing.T) {
	for _, tc := range []struct {
		name string
		line string
		w    int
		want string
	}{
		{"fits", "abcdefghij", 10, "abcdefghij/cccccccccc"},
		{"no wrap", "aaa bbb ccc", 10, "aaa bbb.../cccccccddd"},
		{"cut", "abcdefghijk", 10, "abcdefg.../cccccccddd"},
		{"wide grapheme at the cut", "abcdef世xyz", 10, "abcdef .../cccccccddd"},
		{"wide grapheme fits", "abcdefgh世", 10, "abcdefgh世/cccccccccc"},
		{"space 3", "abcd", 3, ".../ddd"},
		{"space 2", "abc", 2, "../dd"},
		{"space 1", "ab", 1, "./d"},
		{"wide grapheme in space 1", "世", 1, "./d"},
		{"tab", "a\tb", 20, "a       b/ccccccccc"},
		{"tab over the space", "\tabc", 10, "       .../cccccccddd"},
		{"empty line", "", 10, "/"},
	} {
		got := textRows(Source, []string{tc.line}, 0, 0, false, tc.w, true)
		if len(got) != 1 || got[0] != tc.want {
			t.Errorf("%s: %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestTextLayoutLinesNumbers(t *testing.T) {
	lines := make([]string, 12)
	for i := range lines {
		lines[i] = fmt.Sprintf("l%d", i+1)
	}
	for _, tc := range []struct {
		name        string
		kind        ContentKind
		lines       []string
		first, last int
		numbers     bool
		w           int
		want        []string
	}{
		{"lines", PlainText, lines, 10, 12, false, 20, []string{"l10/nnn", "l11/nnn", "l12/nnn"}},
		{"numbers of the line count", PlainText, lines[:10], 0, 0, true, 20,
			[]string{" 1 l1/dddnn", " 2 l2/dddnn", " 3 l3/dddnn", " 4 l4/dddnn", " 5 l5/dddnn",
				" 6 l6/dddnn", " 7 l7/dddnn", " 8 l8/dddnn", " 9 l9/dddnn", "10 l10/dddnnn"}},
		{"numbers of last", PlainText, lines, 9, 10, true, 20, []string{" 9 l9/dddnn", "10 l10/dddnnn"}},
		{"numbers of a short last", PlainText, lines, 2, 3, true, 20, []string{"2 l2/ddnn", "3 l3/ddnn"}},
		{"numbers on wrapped rows", PlainText, []string{"aaa bbb", ""}, 0, 0, true, 6,
			[]string{"1 aaa/ddnnn", "  bbb/ddnnn", "2 /dd"}},
		{"numbers in source", Source, []string{"abcdefgh"}, 0, 0, true, 8, []string{"1 abc.../ddcccddd"}},
		{"tab without the line number", PlainText, []string{"ab\tc"}, 0, 0, true, 14, []string{"1 ab      c/ddnnnnnnnnn"}},
		{"tab in source with numbers", Source, []string{"ab\tc"}, 0, 0, true, 20, []string{"1 ab      c/ddccccccccc"}},
		{"number column at W / 2", PlainText, lines, 0, 0, true, 6, []string{" 1 l1/dddnn"}},
		{"number column over W / 2", PlainText, lines, 0, 0, true, 5, []string{"l1/nn"}},
	} {
		got := textRows(tc.kind, tc.lines, tc.first, tc.last, tc.numbers, tc.w, true)
		if len(got) > len(tc.want) {
			got = got[:len(tc.want)]
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestTextLayoutPages(t *testing.T) {
	text := func(lines []string, w, h int) string {
		return pageText(TextLayout(&Content{Kind: PlainText, Lines: lines}, 0, 0, false, w, fileTheme), h)
	}
	for _, tc := range []struct {
		name  string
		lines []string
		w, h  int
		want  string
	}{
		{"pages of h rows", []string{"a", "b", "c", "d", "e"}, 10, 2, "a b|c d|e -"},
		{"empty file", nil, 10, 3, "- - -"},
		{"page in a wrapped line", []string{"aaa bbb ccc", "d"}, 3, 2, "aaa bbb|ccc d"},
		{"empty row at the start of a page", []string{"a", "", "b"}, 10, 1, "a|-|b"},
		{"page from a restart point", []string{"a", "bb cc dd", "e"}, 2, 2, "a bb|cc dd|e -"},
	} {
		if got := text(tc.lines, tc.w, tc.h); got != tc.want {
			t.Errorf("%s: %q, want %q", tc.name, got, tc.want)
		}
	}
	lines := make([]string, MaxFileRows+1)
	pages := SplitPages(TextLayout(&Content{Kind: Source, Lines: lines}, 0, 0, false, 10, fileTheme), MaxFileRows)
	if len(pages) != 1 || pages[0].N != MaxFileRows {
		t.Errorf("row limit: %+v", pages)
	}
}
