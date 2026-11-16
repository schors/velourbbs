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
