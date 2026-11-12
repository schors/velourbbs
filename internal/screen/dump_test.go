package screen

import (
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

// The example of section 9.3.
func TestWriteDump(t *testing.T) {
	normal := Style{Fg: IndexColour(7), Bg: IndexColour(0)}
	title := Style{Fg: IndexColour(15), Bg: IndexColour(4), Attr: Bold}
	cyan := Style{Fg: IndexColour(14), Bg: IndexColour(0)}
	page := Page{row(80, "Welcome!", title, normal), row(80, "[G]oodbye", cyan, normal)}
	var b strings.Builder
	if err := WriteDump(&b, 80, 25, []Page{page}); err != nil {
		t.Fatal(err)
	}
	want := `page 1/1 80x25
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

func TestWriteDumpPages(t *testing.T) {
	normal := Style{Fg: IndexColour(7), Bg: IndexColour(0)}
	wide := Style{Fg: IndexColour(11), Bg: IndexColour(0)}
	pages := []Page{
		{row(5, "中a", wide, normal), row(5, "", normal, normal)},
		{row(5, "b", rgbStyle(), normal)},
	}
	var b strings.Builder
	if err := WriteDump(&b, 5, 3, pages); err != nil {
		t.Fatal(err)
	}
	want := `page 1/2 5x3
|中a  |
|     |
styles
1: 1-3 11/0, 4-5 7/0
2: 1-5 7/0
page 2/2 5x3
|b    |
styles
1: 1-1 #102030/-, 2-5 7/0
`
	if b.String() != want {
		t.Errorf("got\n%s\nwant\n%s", b.String(), want)
	}
}

func rgbStyle() Style { return Style{Fg: RGBColour(0x10, 0x20, 0x30)} }
