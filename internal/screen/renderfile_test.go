package screen_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/schors/velourbbs/internal/bbt"
	"github.com/schors/velourbbs/internal/screen"
)

// fileScreen loads the file screen f with the body body and the content file a.txt.
func fileScreen(t *testing.T, body string) *screen.Screen {
	t.Helper()
	fsys := fstest.MapFS{
		"screens/f.bbt": {Data: []byte("+++\nfile = \"a.txt\"\n+++\n" + body)},
		"content/a.txt": {Data: []byte("a\n")},
	}
	screens, ds := bbt.Load(fsys, theme, screen.Layout{Width: screen.WidthLimits{Min: 20, Max: 20}}, nil)
	if ds.HasErrors() {
		t.Fatalf("diags: %v", ds)
	}
	return screens["f"]
}

// texts gives the text of each row of p without the spaces at the end.
func texts(p screen.Page) []string {
	var out []string
	for _, r := range p.Rows {
		var b strings.Builder
		for _, c := range r {
			b.WriteString(c.Grapheme)
		}
		out = append(out, strings.TrimRight(b.String(), " "))
	}
	return out
}

// fileRows gives n rows "f1" to "fn"; a nil row is an empty row.
func fileRows(n int) []screen.Row {
	rows := make([]screen.Row, n)
	for i := range rows {
		for _, c := range fmt.Sprintf("f%d", i+1) {
			rows[i] = append(rows[i], screen.Cell{Grapheme: string(c), Style: theme["normal"]})
		}
	}
	return rows
}

// The rows above and below the file area and the cut at term.rows (file-view, section 3.2).
func TestRenderFile(t *testing.T) {
	env := &screen.Env{Term: screen.Term{Type: "ansi", Cols: 30, Rows: 10, Charset: "utf-8", Colors: 16}}
	s := fileScreen(t, "Top\n{>file}\nEnd 1\nEnd 2\n")
	if n := screen.FrameRows(s, env, theme); n != 3 {
		t.Errorf("FrameRows %d, want 3", n)
	}
	p := screen.RenderFile(s, env, theme, append(fileRows(2), nil, nil), 3)
	if want := []string{"Top", "f1", "f2", "", "", "End 1", "End 2"}; !slices.Equal(texts(p), want) {
		t.Errorf("rows %q, want %q", texts(p), want)
	}
	if p.Num != 3 || p.Top != 0 || p.Height != 10 || p.Margin != 5 || len(p.Rows[0]) != 20 {
		t.Errorf("page %d top %d height %d margin %d width %d", p.Num, p.Top, p.Height, p.Margin, len(p.Rows[0]))
	}

	// A page with more rows than term.rows is cut.
	p = screen.RenderFile(s, env, theme, fileRows(8), 1)
	if want := []string{"Top", "f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "End 1"}; !slices.Equal(texts(p), want) {
		t.Errorf("cut: rows %q, want %q", texts(p), want)
	}

	// Rows above the file area over term.rows: the page has only rows above.
	var above strings.Builder
	for i := 1; i <= 12; i++ {
		fmt.Fprintf(&above, "A%d\n", i)
	}
	s = fileScreen(t, above.String()+"{>file}\nEnd\n")
	if n := screen.FrameRows(s, env, theme); n != 13 {
		t.Errorf("FrameRows %d, want 13", n)
	}
	p = screen.RenderFile(s, env, theme, fileRows(1), 1)
	if want := []string{"A1", "A2", "A3", "A4", "A5", "A6", "A7", "A8", "A9", "A10"}; !slices.Equal(texts(p), want) {
		t.Errorf("above: rows %q, want %q", texts(p), want)
	}

	// The rows above reach MaxRenderRows: the render stops before the file area,
	// and the file rows come after the kept rows.
	// Each line wraps into 2 rows of W = 20.
	s = fileScreen(t, strings.Repeat("aaaaaaaaaaaaaa bbbbbbbbbbbbb\n", screen.MaxRenderRows/2)+"{>file}\nEnd\n")
	if n := screen.FrameRows(s, env, theme); n != screen.MaxRenderRows {
		t.Errorf("FrameRows %d, want %d", n, screen.MaxRenderRows)
	}
	p = screen.RenderFile(s, env, theme, fileRows(1), 1)
	if got := texts(p); len(got) != 10 || slices.Contains(got, "f1") {
		t.Errorf("over the row limit: rows %q", got)
	}

	// The file area as the first and the last line, and a field of the page.
	s = fileScreen(t, "{>file}\n")
	p = screen.RenderFile(s, env, theme, fileRows(2), 1)
	if want := []string{"f1", "f2"}; !slices.Equal(texts(p), want) {
		t.Errorf("only the area: rows %q, want %q", texts(p), want)
	}
	s = fileScreen(t, "{>file}\n{file.page}/{file.pages} {file.title}\n")
	env.File = &screen.FileValues{Content: s.FileScreen.Contents["a.txt"], Page: 2, Pages: 5}
	p = screen.RenderFile(s, env, theme, fileRows(1), 2)
	if want := []string{"f1", "2/5 a.txt"}; !slices.Equal(texts(p), want) {
		t.Errorf("fields: rows %q, want %q", texts(p), want)
	}
}

// The key margin: the layout at the area width A and the margin cells in normal (file-view, section 3.1).
func TestRenderFileMargin(t *testing.T) {
	fsys := fstest.MapFS{
		"screens/f.bbt": {Data: []byte("+++\nfile = \"a.txt\"\nmargin = [3, 2]\n+++\nTop\n{>file}\n")},
		"content/a.txt": {Data: []byte("aaaa bbbb cccc dddd\n")},
	}
	screens, ds := bbt.Load(fsys, theme, screen.Layout{Width: screen.WidthLimits{Min: 20, Max: 20}}, nil)
	if ds.HasErrors() {
		t.Fatalf("diags: %v", ds)
	}
	s := screens["f"]
	env := &screen.Env{Term: screen.Term{Type: "ansi", Cols: 30, Rows: 10, Charset: "utf-8", Colors: 16}}

	// W = 20, A = 20 - 3 - 2 = 15: the line of 19 cells wraps.
	l := screen.ContentLayout(s.FileScreen, s.FileScreen.Contents["a.txt"], 20, theme)
	rows := screen.PageRows(l, screen.FilePage{N: 2}, 2)
	p := screen.RenderFile(s, env, theme, rows, 1)
	if want := []string{"Top", "   aaaa bbbb cccc", "   dddd"}; !slices.Equal(texts(p), want) {
		t.Errorf("rows %q, want %q", texts(p), want)
	}

	// The margin cells are spaces in normal; the file cells keep their style.
	title := theme["title"]
	file := []screen.Row{slices.Repeat([]screen.Cell{{Grapheme: "x", Style: title}}, 15), nil}
	p = screen.RenderFile(s, env, theme, file, 1)
	for i, row := range p.Rows[1:] {
		if len(row) != 20 {
			t.Fatalf("row %d: %d cells, want 20", i+1, len(row))
		}
		for j, c := range row {
			want := screen.Cell{Grapheme: " ", Style: theme["normal"]}
			if i == 0 && j >= 3 && j < 18 {
				want = screen.Cell{Grapheme: "x", Style: title}
			}
			if c != want {
				t.Errorf("row %d cell %d: %v, want %v", i+1, j+1, c, want)
			}
		}
	}
}
