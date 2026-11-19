package screen_test

import (
	"maps"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/schors/velourbbs/internal/bbt"
	"github.com/schors/velourbbs/internal/screen"
)

// listScreen loads the list screen l with rows = 3, the body body and the directory d
// with the content files names (list-view, section 3).
func listScreen(t *testing.T, valign, body string, names ...string) *screen.Screen {
	t.Helper()
	fsys := fstest.MapFS{
		"screens/l.bbt": {Data: []byte("+++\ndata = { from = \"files\", args = { dir = \"d\" } }\nrows = 3\n" + valign + "+++\n" + body)},
	}
	for _, n := range names {
		fsys["content/d/"+n] = &fstest.MapFile{Data: []byte("x\n")}
	}
	if len(names) == 0 {
		fsys["content/d/.keep"] = &fstest.MapFile{}
	}
	sets, ds := bbt.Load(fsys, bbt.Settings{Theme: theme, Layout: screen.Layout{Width: screen.WidthLimits{Min: 20, Max: 20}}})
	if ds.HasErrors() {
		t.Fatalf("diags: %v", ds)
	}
	return sets["en"]["l"]
}

// listPage renders the list page n of s with the page size p at 30x10.
func listPage(s *screen.Screen, n, p int) screen.Page {
	rows := s.List.List(nil)
	pages := max(1, (len(rows)+p-1)/p)
	env := &screen.Env{
		Term: screen.Term{Type: "ansi", Cols: 30, Rows: 10, Colors: 16},
		List: &screen.ListValues{Page: n, Pages: pages, Empty: len(rows) == 0},
	}
	return screen.RenderList(s, env, theme, rows, n, p)
}

// A list line gives P rows: the rows of the page, then empty rows in normal. A row does not wrap:
// the host cuts it at W, a wide grapheme at W becomes a space. Each row starts with the style
// before the list line, and the style after it is the same (list-view, section 5).
func TestRenderList(t *testing.T) {
	s := listScreen(t, "", "{!title}{list.page}/{list.pages}\n{#each}{row.num} {!prompt}{row.path}  界界界界界界界界{/each}\nEnd\n",
		"a.txt", "b.txt", "c.txt", "d.txt")
	p := listPage(s, 1, 3)
	want := []string{"1/2", "1 d/a.txt  界界界界", "2 d/b.txt  界界界界", "3 d/c.txt  界界界界", "End"}
	if got := texts(p); !slices.Equal(got, want) || p.Num != 1 || p.Top != 0 {
		t.Errorf("page 1: %q, num %d, top %d; want %q", got, p.Num, p.Top, want)
	}
	row := p.Rows[1]
	if row[0].Style != theme["title"] || row[2].Style != theme["prompt"] || row[19].Grapheme != " " {
		t.Errorf("row 1: %+v", row)
	}
	if p.Rows[4][0].Style != theme["title"] {
		t.Errorf("the style after the list line: %v, want title", p.Rows[4][0].Style)
	}
	p = listPage(s, 2, 3)
	want = []string{"2/2", "4 d/d.txt  界界界界", "", "", "End"}
	if got := texts(p); !slices.Equal(got, want) {
		t.Errorf("page 2: %q, want %q", got, want)
	}
	if p.Rows[2][0].Style != theme["normal"] {
		t.Errorf("empty row style %v, want normal", p.Rows[2][0].Style)
	}
}

// The page size of list-view, section 7: page 1 with P = rows gives R rows.
// A page with fewer rows than term.rows has the vertical position of valign.
func TestListRows(t *testing.T) {
	s := listScreen(t, "valign = \"bottom\"\n", "Top\n{#each}{row.num}{/each}\n", "a.txt")
	env := &screen.Env{
		Term: screen.Term{Type: "ansi", Cols: 30, Rows: 10, Colors: 16},
		List: &screen.ListValues{Page: 1, Pages: 1},
	}
	if r := screen.ListRows(s, env, theme, s.List.List(nil), 3); r != 4 {
		t.Errorf("ListRows %d, want 4", r)
	}
	if p := listPage(s, 1, 3); p.Top != 6 || len(p.Rows) != 4 {
		t.Errorf("top %d, rows %d; want 6 and 4", p.Top, len(p.Rows))
	}
}

// row.selected is true on the row of the cursor (list-view, section 8).
func TestRenderListCursor(t *testing.T) {
	s := listScreen(t, "select = { do = \"back\" }\n", "{#each}{#if row.selected}>{#else} {/if}{row.num}{/each}\n",
		"a.txt", "b.txt", "c.txt", "d.txt")
	env := &screen.Env{
		Term: screen.Term{Type: "ansi", Cols: 30, Rows: 10, Colors: 16},
		List: &screen.ListValues{Page: 1, Pages: 2, Cursor: 1},
	}
	rows := s.List.List(nil)
	want := []string{" 1", ">2", " 3"}
	if got := texts(screen.RenderList(s, env, theme, rows, 1, 3)); !slices.Equal(got, want) {
		t.Errorf("page 1: %q, want %q", got, want)
	}
	env.List.Page, env.List.Cursor = 2, 3
	want = []string{">4", "", ""}
	if got := texts(screen.RenderList(s, env, theme, rows, 2, 3)); !slices.Equal(got, want) {
		t.Errorf("page 2: %q, want %q", got, want)
	}
}

// The action of select has the literal values and the values of the row: an integer
// in its decimal form (list-view, section 8). The action of the model does not change.
func TestSelectAction(t *testing.T) {
	sel := &screen.Select{
		Do:   screen.Action{Op: screen.Gosub, Screen: "r", Args: map[string]string{"a": "lit"}},
		Refs: map[string]string{"n": "num", "m": "minutes", "p": "path"},
	}
	a := sel.Action(screen.ListRow{"minutes": 3, "path": "x.md"}, 7)
	want := map[string]string{"a": "lit", "n": "7", "m": "3", "p": "x.md"}
	if a.Op != screen.Gosub || a.Screen != "r" || !maps.Equal(a.Args, want) {
		t.Errorf("action %+v, want args %v", a, want)
	}
	if len(sel.Do.Args) != 1 {
		t.Errorf("the action of the model changed: %v", sel.Do.Args)
	}
}

// The list row in art (list-view, section 9): rows rectangle rows show the list page in the cells
// of the list row; the cells outside the rectangle stay. A rectangle row without a row is spaces
// in the styles of the list row cells. The selected row takes select only in a screen with select.
func TestRenderListArt(t *testing.T) {
	row := "|\x1b[33m{ row.num}\x1b[0m {row.path  }|"
	blank := "|" + strings.Repeat(" ", 23) + "|"
	art := "+-----------------------+\n" + row + "\n" + blank + "\n" + blank + "\n+--{list.page}---------+"
	fsys := fstest.MapFS{
		"screens/l.bbt": {Data: []byte("+++\ndata = { from = \"files\", args = { dir = \"d\" } }\nrows = 3\n" +
			"select = { do = \"back\" }\n+++\n{>l.ans}\n")},
		"screens/m.bbt": {Data: []byte("+++\ndata = { from = \"files\", args = { dir = \"d\" } }\nrows = 3\n+++\n{>l.ans}\n")},
		"screens/l.ans": {Data: []byte(sauceArt(art, 25))},
	}
	for _, n := range []string{"a.txt", "b.txt", "c.txt", "d.txt"} {
		fsys["content/d/"+n] = &fstest.MapFile{Data: []byte("x\n")}
	}
	sets, ds := bbt.Load(fsys, bbt.Settings{Theme: theme, Layout: screen.Layout{Width: screen.WidthLimits{Min: 25, Max: 25}}})
	if ds.HasErrors() {
		t.Fatalf("diags: %v", ds)
	}
	render := func(s *screen.Screen, n, p, cursor int) screen.Page {
		rows := s.List.List(nil)
		env := &screen.Env{
			Term: screen.Term{Type: "ansi", Cols: 30, Rows: 10, Colors: 16},
			List: &screen.ListValues{Page: n, Pages: (len(rows) + p - 1) / p, Cursor: cursor},
		}
		return screen.RenderList(s, env, theme, rows, n, p)
	}
	l := sets["en"]["l"]
	a := l.Body[0].(screen.ArtInsert).Art
	grey, yellow := a.Rows[1][0].Style, a.Rows[1][1].Style
	if grey == yellow {
		t.Fatalf("the art styles: %v %v", grey, yellow)
	}

	p := render(l, 1, 3, 1)
	want := []string{"+-----------------------+", "|         1 d/a.txt     |", "|         2 d/b.txt     |",
		"|         3 d/c.txt     |", "+--1          ---------+"}
	if got := texts(p); !slices.Equal(got, want) {
		t.Errorf("page 1: %q, want %q", got, want)
	}
	if c := p.Rows[1]; c[0].Style != grey || c[1].Style != yellow || c[11].Style != grey {
		t.Errorf("row 1 styles: %v %v %v", c[0].Style, c[1].Style, c[11].Style)
	}
	if c := p.Rows[2]; c[0].Style != grey || c[1].Style != theme["select"] || c[23].Style != theme["select"] ||
		c[24].Style != grey {
		t.Errorf("selected row styles: %v %v %v %v", c[0].Style, c[1].Style, c[23].Style, c[24].Style)
	}

	// Page 2 has one row. P = 2 leaves the third rectangle row of page 1 without a row.
	empty := "|" + strings.Repeat(" ", 23) + "|"
	for _, tt := range []struct {
		p    screen.Page
		want []string
	}{
		{render(l, 2, 3, 3), []string{"|         4 d/d.txt     |", empty, empty}},
		{render(l, 1, 2, 0), []string{"|         1 d/a.txt     |", "|         2 d/b.txt     |", empty}},
	} {
		if got := texts(tt.p)[1:4]; !slices.Equal(got, tt.want) {
			t.Errorf("page %d: %q, want %q", tt.p.Num, got, tt.want)
		}
		if c := tt.p.Rows[3]; c[1].Style != yellow || c[11].Style != grey || c[24] != a.Rows[3][24] {
			t.Errorf("page %d empty row: %v %v %v", tt.p.Num, c[1].Style, c[11].Style, c[24])
		}
	}

	// Without select, the cursor row has the styles of the art.
	if c := render(sets["en"]["m"], 1, 3, 1).Rows[2]; c[1].Style != yellow {
		t.Errorf("row 2 without select: %v, want %v", c[1].Style, yellow)
	}
	// The render does not change the model.
	if a.Rows[1][1].Style != yellow || a.Rows[2][1] != (screen.Cell{Grapheme: " ", Style: grey}) {
		t.Errorf("the art model changed: %v %v", a.Rows[1][1], a.Rows[2][1])
	}
}
