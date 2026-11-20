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
