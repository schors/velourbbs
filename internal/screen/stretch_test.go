package screen_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/schors/velourbbs/internal/screen"
)

// pages renders main at cols x rows and gives its pages.
func pages(t *testing.T, main string, cols, rows int) []screen.Page {
	t.Helper()
	s := load(t, main)
	env := &screen.Env{Term: screen.Term{Type: "ansi", Cols: cols, Rows: rows, Colors: 16}}
	var out []screen.Page
	for p := range screen.Render(s, env, theme, "-- More --") {
		out = append(out, p)
	}
	return out
}

// cells gives the text of each row of p with the spaces at the end.
func cells(p screen.Page) []string {
	var out []string
	for _, r := range p.Rows {
		var b strings.Builder
		for _, c := range r {
			b.WriteString(c.Grapheme)
		}
		out = append(out, b.String())
	}
	return out
}

// The stretch tokens of a row share its free cells (section 6.6).
func TestStretch(t *testing.T) {
	w10 := "+++\nwidth = [10, 10]\n+++\n"
	tests := []struct {
		name, body string
		cols       int
		want       []string
	}{
		{"leader", "{~.}", 10, []string{".........."}},
		{"space", "a{~}b", 10, []string{"a        b"}},
		{"between", "a{~.}b", 10, []string{"a........b"}},
		// F / n cells, rounded down; the last F mod n tokens get one more cell.
		{"centre", "{~}x{~}", 10, []string{"    x     "}},
		{"three", "{~1}{~2}{~3}", 10, []string{"1112223333"}},
		{"three of 11", "a{~1}{~2}{~3}", 10, []string{"a111222333"}},
		// The spaces at the start and at the end of the source line count.
		{"spaces", "  a{~.}b  ", 10, []string{"  a....b  "}},
		{"many spaces", "            a{~.}", 10, []string{"          "}},
		// A row of W cells or more: each token has 0 cells, and the host cuts the row at W.
		{"exact", "abcde{~.}fghij", 10, []string{"abcdefghij"}},
		{"over", "abcdefghij{~.}k", 10, []string{"abcdefghij"}},
		// A line with a stretch token does not wrap.
		{"no wrap", "aaa bbb ccc {~.}ddd", 10, []string{"aaa bbb cc"}},
		// The cut through a wide grapheme leaves a space.
		{"wide cut", "a中中中中{~.}中", 10, []string{"a中中中中 "}},
		{"wide", "中{~.}中", 10, []string{"中......中"}},
		// A fixed-width field is its cells.
		{"field", "{term.type   }{~.}{term.cols}", 20, []string{"ansi          ....20"}},
		// A stretch token in a false block has no effect: the line wraps.
		{"false block", "aaaa{#if term.cols > 100}{~.}{/if} bbbb cccc", 10, []string{"aaaa bbbb ", "cccc      "}},
		// A line with a block token and a stretch token gives one row, also of spaces only.
		{"block", "{#if term.cols > 1}{~-}{/if}", 10, []string{"----------"}},
		{"block spaces", "{#if term.cols > 1}{~}{/if}\nx", 10, []string{"          ", "x         "}},
		{"block no stretch", "{#if term.cols > 1} {/if}\nx", 10, []string{"x         "}},
		// A false block joins the source lines: the render line has the stretch token.
		{"joined", "a{#if term.cols > 100}\n{/if}{~.}b", 10, []string{"a........b"}},
		{"joined wrap", "aaaa{#if term.cols > 100}\n{/if} bbbb cccc\n{~.}", 10, []string{"aaaa bbbb ", "cccc      ", ".........."}},
		// term.cols < min: the row has W = min cells and the host cuts it at term.cols.
		{"cut", "+++\nwidth = [12, 12]\n+++\na{~.}b", 10, []string{"a........."}},
	}
	for _, tt := range tests {
		body := tt.body
		if !strings.HasPrefix(body, "+++") {
			body = w10 + body
			if tt.cols != 10 {
				body = "+++\nwidth = [20, 20]\n+++\n" + tt.body
			}
		}
		ps := pages(t, body+"\n", tt.cols, 10)
		if got := cells(ps[0]); len(ps) != 1 || !slices.Equal(got, tt.want) {
			t.Errorf("%s: %q, want %q", tt.name, got, tt.want)
		}
	}
}

// A token fills its cells with its character in the style at the token (section 6.6).
func TestStretchStyle(t *testing.T) {
	ps := pages(t, "+++\nwidth = [10, 10]\n+++\n{!title}a{!prompt}{~.}{!}b\n", 10, 10)
	row := ps[0].Rows[0]
	if row[0].Style != theme["title"] || row[1].Style != theme["prompt"] || row[8].Style != theme["prompt"] || row[9].Style != theme["normal"] {
		t.Errorf("row %+v", row)
	}
}

// The render width and the margin of section 6.1.
func TestGutter(t *testing.T) {
	tests := []struct {
		min, max, gutter int
		align            screen.Align
		cols             int
		w, margin        int
	}{
		// The example of section 6.1.
		{68, 108, 6, screen.AlignCenter, 80, 68, 6},
		{68, 108, 6, screen.AlignCenter, 132, 108, 12},
		// W = term.cols - 2 * gutter: the margin is the gutter.
		{40, 200, 6, screen.AlignLeft, 80, 68, 6},
		{40, 200, 6, screen.AlignCenter, 80, 68, 6},
		{40, 200, 6, screen.AlignRight, 80, 68, 6},
		// W at max: F = 12, the gutter is 4.
		{40, 68, 4, screen.AlignLeft, 80, 68, 4},
		{40, 68, 4, screen.AlignRight, 80, 68, 8},
		{40, 68, 4, screen.AlignCenter, 80, 68, 6},
		// W at min: F = 2 is less than the gutter.
		{78, 78, 6, screen.AlignLeft, 80, 78, 2},
		{78, 78, 6, screen.AlignRight, 80, 78, 0},
		{78, 78, 6, screen.AlignCenter, 80, 78, 1},
		// term.cols <= W: no margin.
		{80, 80, 6, screen.AlignLeft, 80, 80, 0},
		{90, 90, 6, screen.AlignRight, 80, 90, 0},
		// gutter = 0: the margin of the align.
		{40, 60, 0, screen.AlignLeft, 80, 60, 0},
		{40, 60, 0, screen.AlignRight, 80, 60, 20},
		{40, 60, 0, screen.AlignCenter, 81, 60, 10},
		// A large gutter: W at min.
		{20, 80, 100, screen.AlignLeft, 80, 20, 60},
	}
	for _, tt := range tests {
		l := screen.Layout{Width: screen.WidthLimits{Min: tt.min, Max: tt.max}, Gutter: tt.gutter, Align: tt.align}
		if w, m := l.RenderWidth(tt.cols), l.Margin(tt.cols); w != tt.w || m != tt.margin {
			t.Errorf("%+v at %d: W %d, margin %d; want %d and %d", l, tt.cols, w, m, tt.w, tt.margin)
		}
	}
	// The render uses W and the margin: a stretch row fills W.
	ps := pages(t, "+++\nwidth = [68, 108]\ngutter = 6\n+++\n{~.}\n", 80, 10)
	if p := ps[0]; p.Margin != 6 || len(p.Rows[0]) != 68 || p.Rows[0][67].Grapheme != "." {
		t.Errorf("margin %d, row %d cells", p.Margin, len(p.Rows[0]))
	}
}

// The spacer lines share the free rows of the page (section 6.6).
func TestSpacer(t *testing.T) {
	w10 := "+++\nwidth = [10, 10]\nvalign = \"bottom\"\n+++\n"
	tests := []struct {
		name, body string
		rows       int
		want       []string // the rows without the spaces at the end
		top        int
	}{
		{"one", "a\n{~}\nb", 6, []string{"a", "", "", "", "", "b"}, 0},
		// 7 free rows: 3 and 4.
		{"two", "a\n{~}\nb\n{~}\nc", 10, []string{"a", "", "", "", "b", "", "", "", "", "c"}, 0},
		{"none free", "a\n{~}\nb", 2, []string{"a", "b"}, 0},
		// A spacer line in a false block gives nothing: valign has its effect.
		{"false block", "a\n{#if term.cols > 100}\n{~}\n{/if}\nb", 5, []string{"a", "b"}, 3},
		{"true block", "a\n{#if term.cols > 1}\n{~}\n{/if}\nb", 5, []string{"a", "", "", "", "b"}, 0},
		// The free rows count the rows of the wrap, and a stretch line as one row.
		{"wrap", "aaaa bbbb cccc\n{~}\naaaa bbbb cccc{~.}", 6, []string{"aaaa bbbb", "cccc", "", "", "", "aaaa bbbb"}, 0},
	}
	for _, tt := range tests {
		ps := pages(t, w10+tt.body+"\n", 10, tt.rows)
		if got := texts(ps[0]); len(ps) != 1 || !slices.Equal(got, tt.want) || ps[0].Top != tt.top {
			t.Errorf("%s: %d pages, %q, top %d; want %q, top %d", tt.name, len(ps), got, ps[0].Top, tt.want, tt.top)
		}
	}

	// More rows than term.rows: each spacer line gives 0 rows, and the pages have More rows.
	ps := pages(t, w10+"a\n{~}\nb\nc\nd\n", 10, 3)
	if len(ps) != 2 || !slices.Equal(texts(ps[0]), []string{"a", "b", "-- More --"}) || !slices.Equal(texts(ps[1]), []string{"c", "d"}) {
		t.Errorf("pages %d: %q", len(ps), texts(ps[0]))
	}
}

// A list line has the rules of section 6.6, and the spacer lines take the free rows after
// the page size (list-view, sections 5 and 7).
func TestStretchList(t *testing.T) {
	s := listScreen(t, "", "{~}{list.page}{~}\n{#each}{row.num}{~.}{row.path}{/each}\n{~}\nEnd\n", "a.txt", "b.txt")
	p := listPage(s, 1, 3)
	dots := strings.Repeat(".", 12)
	// The header, the list rows a and b, an empty list row, the 5 free rows and End.
	want := []string{"         1          ", "1" + dots + "d/a.txt", "2" + dots + "d/b.txt", "", "", "", "", "", "", "End"}
	got := cells(p)
	for i := 3; i < len(got); i++ {
		got[i] = strings.TrimRight(got[i], " ")
	}
	if !slices.Equal(got, want) || p.Top != 0 {
		t.Errorf("page: %q, top %d; want %q", got, p.Top, want)
	}
	env := &screen.Env{
		Term: screen.Term{Type: "ansi", Cols: 30, Rows: 10, Colors: 16},
		List: &screen.ListValues{Page: 1, Pages: 1},
	}
	if r := screen.ListRows(s, env, theme, s.List.List(nil), 3); r != 5 {
		t.Errorf("ListRows %d, want 5: the spacer lines count 0 rows", r)
	}
}

// In a file screen, a spacer line gives 0 rows (section 6.6).
func TestStretchFile(t *testing.T) {
	env := &screen.Env{Term: screen.Term{Type: "ansi", Cols: 30, Rows: 10, Colors: 16}}
	s := fileScreen(t, "{~}{term.cols}{~}\n{~}\n{>file}\n{~}\nEnd\n")
	if n := screen.FrameRows(s, env, theme); n != 2 {
		t.Errorf("FrameRows %d, want 2", n)
	}
	p := screen.RenderFile(s, env, theme, fileRows(2), 1)
	if got, want := texts(p), []string{"         30", "f1", "f2", "End"}; !slices.Equal(got, want) {
		t.Errorf("page: %q, want %q", got, want)
	}
}
