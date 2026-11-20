package screen_test

import (
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/schors/velourbbs/internal/bbt"
	"github.com/schors/velourbbs/internal/screen"
)

// A fixed-width field wider than W (screen-render, section 3.6): the row of Width has the field
// before the cut. The log does not change the render.
func TestLayoutField(t *testing.T) {
	right := "{" + strings.Repeat(" ", 43) + "t}" // 46 cells, right-aligned: "ansi" at the end
	left := "{t" + strings.Repeat(" ", 43) + "}"  // 46 cells: "ansi" at the start
	body := "+++\nwidth = [40, 40]\n+++\n" +
		right + "\n" + // 46 cells
		"x{~}" + right + "\n" + // a stretch line: 1 + 46 cells
		left + "\n" // the cut drops only spaces
	fsys := fstest.MapFS{"screens/main.bbt": {Data: []byte(body)}}
	sets, ds := bbt.Load(fsys, bbt.Settings{Theme: theme, Layout: screen.Layout{Width: screen.WidthLimits{Min: 40, Max: 80}}, Fields: aliases})
	if ds.HasErrors() {
		t.Fatalf("diags: %v", ds)
	}
	s := sets["en"]["main"]
	render := func(log *screen.LayoutLog) []screen.Page {
		env := &screen.Env{Term: screen.Term{Type: "ansi", Cols: 80, Rows: 24, Colors: 16}, Items: map[string]bool{}, Layout: log}
		var pages []screen.Page
		for p := range screen.Render(s, env, theme, "-- More --") {
			pages = append(pages, p)
		}
		return pages
	}
	log := &screen.LayoutLog{}
	if got, want := render(log), render(nil); !reflect.DeepEqual(got, want) {
		t.Errorf("pages with the log %v, want %v", got, want)
	}
	pos := func(line int) screen.Pos { return screen.Pos{File: "screens/main.bbt", Line: line, Col: 1} }
	want := []screen.LayoutNote{
		{Kind: screen.LayoutWidth, Pos: pos(4), N: 46, M: 40},
		{Kind: screen.LayoutWidth, Pos: pos(5), N: 47, M: 40},
	}
	if !reflect.DeepEqual(log.Notes, want) {
		t.Errorf("notes %+v, want %+v", log.Notes, want)
	}
}

// A line that wraps over a More row gives the widest row of the first page where the cut drops text
// (screen-render, section 3.6).
func TestLayoutWidthPage(t *testing.T) {
	long := strings.Repeat("X", 30) + " " + strings.Repeat("Y", 40) + " " + strings.Repeat("Z", 40)
	body := "+++\nwidth = [40, 40]\n+++\n" + strings.Repeat("a\n", 8) + long + "\n"
	fsys := fstest.MapFS{"screens/main.bbt": {Data: []byte(body)}}
	sets, ds := bbt.Load(fsys, bbt.Settings{Theme: theme, Layout: screen.Layout{Width: screen.WidthLimits{Min: 40, Max: 80}}, Fields: aliases})
	if ds.HasErrors() {
		t.Fatalf("diags: %v", ds)
	}
	log := &screen.LayoutLog{}
	env := &screen.Env{Term: screen.Term{Type: "ansi", Cols: 20, Rows: 10, Colors: 16}, Items: map[string]bool{}, Layout: log}
	for range screen.Render(sets["en"]["main"], env, theme, "-- More --") {
	}
	// The rows of the line are 9 to 11; page 1 has the rows 1 to 9.
	for _, n := range log.Notes {
		if n.Kind == screen.LayoutWidth {
			if n.N != 30 {
				t.Errorf("Width %d, want 30", n.N)
			}
			return
		}
	}
	t.Errorf("notes %+v, want Width", log.Notes)
}
