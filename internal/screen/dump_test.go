package screen

import (
	"slices"
	"strings"
	"testing"
)

// row builds a row of width cells from text in the style s, and pads it with pad.
func row(width int, text string, s, pad Style) Row {
	var r Row
	for g, w := range Graphemes(text) {
		r = append(r, Cell{g, s})
		if w == 2 {
			r = append(r, Cell{"", s})
		}
	}
	for len(r) < width {
		r = append(r, Cell{" ", pad})
	}
	return r
}

// The example of screen-render, section 3.3.
func TestWriteDump(t *testing.T) {
	normal := Style{Fg: IndexColour(7), Bg: IndexColour(0)}
	title := Style{Fg: IndexColour(15), Bg: IndexColour(4), Attr: Bold}
	cyan := Style{Fg: IndexColour(14), Bg: IndexColour(0)}
	page := Page{Num: 1, Rows: []Row{row(80, "Welcome!", title, normal), row(80, "[G]oodbye", cyan, normal)}, Cols: 80, Height: 24}
	var b strings.Builder
	if err := WriteDump(&b, slices.Values([]Page{page})); err != nil {
		t.Fatal(err)
	}
	want := `page 1 80x24 fill none
|Welcome!                                                                        |
|[G]oodbye                                                                       |
styles
1: 1-8 15/4 bold, 9-80 7/0
2: 1-9 14/0, 10-80 7/0
`
	if b.String() != want {
		t.Errorf("got\n%s\nwant\n%s", b.String(), want)
	}
}

// The second page has a screen area of 2 cells with a margin of 1:
// the cells outside it are spaces in Normal. It starts at the terminal row 2 and fills with spaces.
func TestWriteDumpPages(t *testing.T) {
	normal := Style{Fg: IndexColour(7), Bg: IndexColour(0)}
	wide := Style{Fg: IndexColour(11), Bg: IndexColour(0)}
	pages := []Page{
		{Num: 1, Rows: []Row{row(5, "中a", wide, normal), row(5, "", normal, normal)}, Cols: 5, Height: 3},
		{Num: 2, Rows: []Row{row(2, "b", rgbStyle(), normal)}, Cols: 5, Height: 3, Top: 1, Margin: 1, Normal: normal, Fill: FillSpaces},
	}
	var b strings.Builder
	if err := WriteDump(&b, slices.Values(pages)); err != nil {
		t.Fatal(err)
	}
	want := `page 1 5x3 fill none
|中a  |
|     |
styles
1: 1-3 11/0, 4-5 7/0
2: 1-5 7/0
page 2 5x3 fill spaces
| b   |
styles
2: 1-1 7/0, 2-2 #102030/-, 3-5 7/0
`
	if b.String() != want {
		t.Errorf("got\n%s\nwant\n%s", b.String(), want)
	}
}

func rgbStyle() Style { return Style{Fg: RGBColour(0x10, 0x20, 0x30)} }

// TestWriteDumpVGA: a VGA colour appears as its index, and a VGA colour
// and the same index are one run (section 3.3).
func TestWriteDumpVGA(t *testing.T) {
	art := Style{Fg: VGAColour(7), Bg: VGAColour(0)}
	normal := Style{Fg: IndexColour(7), Bg: IndexColour(0)}
	page := Page{Num: 1, Rows: []Row{row(4, "ab", art, normal)}, Cols: 4, Height: 1}
	var b strings.Builder
	if err := WriteDump(&b, slices.Values([]Page{page})); err != nil {
		t.Fatal(err)
	}
	want := "page 1 4x1 fill none\n|ab  |\nstyles\n1: 1-4 7/0\n"
	if b.String() != want {
		t.Errorf("got\n%s\nwant\n%s", b.String(), want)
	}
}
