package menu

import (
	"errors"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/schors/velourbbs/internal/config"
	"github.com/schors/velourbbs/internal/line"
	"github.com/schors/velourbbs/internal/screen"
)

// askTheme marks the styles of the ask row: more is bold, select is underline.
var askTheme = map[string]screen.Style{
	"normal": {}, "more": {Attr: screen.Bold}, "select": {Attr: screen.Underline},
}

// askRow gives the text of the row i of p, with [ ] around the cells in select
// and < > around the cells in more.
func askRow(p screen.Page, i int) string {
	var b strings.Builder
	var last screen.Style
	for _, c := range append(p.Line(i, nil), screen.Cell{}) { // the last cell closes the style
		if c.Style != last {
			switch last {
			case askTheme["more"]:
				b.WriteString(">")
			case askTheme["select"]:
				b.WriteString("]")
			}
			switch c.Style {
			case askTheme["more"]:
				b.WriteString("<")
			case askTheme["select"]:
				b.WriteString("[")
			}
			last = c.Style
		}
		b.WriteString(c.Grapheme)
	}
	return strings.TrimRight(b.String(), " ")
}

// rowOf gives a row of the text s in normal.
func rowOf(s string) screen.Row {
	var r screen.Row
	for _, c := range s {
		r = append(r, screen.Cell{Grapheme: string(c)})
	}
	return r
}

// The ask row: M is the free cells, at most ask_max. When term.cols < min, the end of the prompt
// is cut at a grapheme boundary until M is at least 1 (menu-nav, section 5.4).
func TestAskPage(t *testing.T) {
	two := []screen.Row{rowOf("one"), rowOf("two")}
	layout := screen.Layout{Width: screen.WidthLimits{Min: 30, Max: 40}, Gutter: 2}
	tests := []struct {
		name   string
		cols   int
		rows   []screen.Row
		top    int
		margin int
		valign screen.VAlign
		ask    screen.Ask
		want   string // the ask row
		row    int    // the terminal row of the ask row
		field  line.Field
	}{
		// W = 40 at 80 columns with the margin 20: M = 40 - 4 - 1 = 35 free cells.
		{"free cells", 80, two, 0, 20, screen.VAlignTop, screen.Ask{Prompt: "No: "},
			strings.Repeat(" ", 20) + "<No: >[" + strings.Repeat(" ", 36) + "]", 2,
			line.Field{Row: 2, Col: 25, Max: 35}},
		{"ask_max", 80, two, 3, 20, screen.VAlignTop, screen.Ask{Prompt: "No: ", Max: 5},
			strings.Repeat(" ", 20) + "<No: >[      ]", 5,
			line.Field{Row: 5, Col: 25, Max: 5}},
		{"ask_max over the free cells", 80, two, 0, 20, screen.VAlignTop, screen.Ask{Prompt: "No: ", Max: 100},
			strings.Repeat(" ", 20) + "<No: >[" + strings.Repeat(" ", 36) + "]", 2,
			line.Field{Row: 2, Col: 25, Max: 35}},
		// At 20 columns, W = 30 is cut at 20: 18 cells of prompt leave M = 1.
		{"cut at term.cols", 20, two, 0, 0, screen.VAlignTop, screen.Ask{Prompt: strings.Repeat("a", 18)},
			"<" + strings.Repeat("a", 18) + ">[  ]", 2,
			line.Field{Row: 2, Col: 19, Max: 1}},
		// 17 cells and a wide grapheme: the cut removes the wide grapheme, then ask_max applies.
		{"cut a wide grapheme", 20, two, 0, 0, screen.VAlignTop, screen.Ask{Prompt: strings.Repeat("a", 17) + "語", Max: 1},
			"<" + strings.Repeat("a", 17) + ">[  ]", 2,
			line.Field{Row: 2, Col: 18, Max: 1}},
		{"cut a wide grapheme, M 2", 20, two, 0, 0, screen.VAlignTop, screen.Ask{Prompt: strings.Repeat("a", 17) + "語"},
			"<" + strings.Repeat("a", 17) + ">[   ]", 2,
			line.Field{Row: 2, Col: 18, Max: 2}},
		// A page cut at term.rows has the ask row on the row term.rows.
		{"cut at term.rows", 80, slices.Repeat(two, 6), 0, 20, screen.VAlignTop, screen.Ask{Prompt: "No: ", Max: 1},
			strings.Repeat(" ", 20) + "<No: >[  ]", 10,
			line.Field{Row: 10, Col: 25, Max: 1}},
		// A page without rows has the ask row where a page of 1 row starts.
		{"no rows, top", 80, nil, 0, 20, screen.VAlignTop, screen.Ask{Prompt: "No: ", Max: 1},
			strings.Repeat(" ", 20) + "<No: >[  ]", 1,
			line.Field{Row: 1, Col: 25, Max: 1}},
		{"no rows, middle", 80, nil, 5, 20, screen.VAlignMiddle, screen.Ask{Prompt: "No: ", Max: 1},
			strings.Repeat(" ", 20) + "<No: >[  ]", 5,
			line.Field{Row: 5, Col: 25, Max: 1}},
		{"no rows, bottom", 80, nil, 10, 20, screen.VAlignBottom, screen.Ask{Prompt: "No: ", Max: 1},
			strings.Repeat(" ", 20) + "<No: >[  ]", 10,
			line.Field{Row: 10, Col: 25, Max: 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := layout
			l.VAlign = tt.valign
			p := screen.Page{Rows: tt.rows, Cols: tt.cols, Height: 10, Top: tt.top, Margin: tt.margin}
			got, f := askPage(p, l, &tt.ask, askTheme)
			tt.field.Style = askTheme["select"]
			if f != tt.field {
				t.Errorf("field %+v, want %+v", f, tt.field)
			}
			if r := askRow(got, tt.row-got.Top-1); r != tt.want {
				t.Errorf("ask row\n%q, want\n%q", r, tt.want)
			}
			if len(tt.rows) > 0 && (len(got.Rows) != len(tt.rows) || got.Top != tt.top) {
				t.Errorf("rows %d, top %d; want %d, %d", len(got.Rows), got.Top, len(tt.rows), tt.top)
			}
			if len(tt.rows) > 0 && askRow(p, tt.row-p.Top-1) == tt.want {
				t.Error("askPage changed the rows of p")
			}
		})
	}
}

// An item with ask reads a line on the last row of the shown page. ENTER with a line runs the action
// with the line as {in}. ESC and ENTER with an empty line show the page again (menu-nav, section 5.4).
func TestRunAsk(t *testing.T) {
	ss := load(t, map[string]string{
		"main": `+++
[[item]]
key = "J"
ask = "N: "
ask_max = 3
do = "gosub msg"
args = { n = "{in}", m = "{in}", a = "x" }
[[item]]
key = "G"
do = "logoff"
+++
Main
`,
		"msg":  "+++\n[[item]]\nkey = \"B\"\ndo = \"back\"\n+++\n{arg.a}{arg.n}{arg.m}\n",
		"info": "",
		"long": "",
	})
	ft := &fakeTerm{cols: []int{80}, keys: []string{"J", "ESC", "X", "J", "ENTER", "j", "a", "BS", "b", "ы", "c", "d", "ENTER", "B"}}
	_, err := runSession(eventsTerm{ft}, config.Config{}, ss, "main", noValues)
	want := []string{
		"screen main", "page: Main", "key J", "ask", "page: N:", "line 1:4 3", "line cancelled", "page: Main",
		"key X ignored", "key J", "ask", "page: N:", "line 1:4 3", "line cancelled", "page: Main",
		"key J", "ask", "page: N:", "line 1:4 3", "line bыc", "screen msg", "page: xbыcbыc", "key B",
		"screen main", "page: Main",
	}
	if !errors.Is(err, io.EOF) || !slices.Equal(ft.log, want) {
		t.Errorf("err %v, log:\n%s\nwant:\n%s", err, strings.Join(ft.log, "\n"), strings.Join(want, "\n"))
	}
}
