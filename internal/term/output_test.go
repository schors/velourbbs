package term

import (
	"strings"
	"testing"

	"newbbs/internal/screen"
)

var (
	idx = screen.IndexColour
	rgb = screen.RGBColour
	def = screen.Colour{}
)

func cell(g string, fg, bg screen.Colour, attr screen.Attr) screen.Cell {
	return screen.Cell{Grapheme: g, Style: screen.Style{Fg: fg, Bg: bg, Attr: attr}}
}

// TestSGR covers each row of the table of section 7.2.
func TestSGR(t *testing.T) {
	tests := []struct {
		name  string
		style screen.Style
		want  string
	}{
		{"no parameters", screen.Style{}, "\x1b[0m"},
		{"bold", screen.Style{Attr: screen.Bold}, "\x1b[0;1m"},
		{"underline", screen.Style{Attr: screen.Underline}, "\x1b[0;4m"},
		{"blink", screen.Style{Attr: screen.Blink}, "\x1b[0;5m"},
		{"index 0-7", screen.Style{Fg: idx(0), Bg: idx(7)}, "\x1b[0;30;47m"},
		{"index 8-15", screen.Style{Fg: idx(8), Bg: idx(15)}, "\x1b[0;90;107m"},
		{"index 16-255", screen.Style{Fg: idx(16), Bg: idx(255)}, "\x1b[0;38;5;16;48;5;255m"},
		{"rgb", screen.Style{Fg: rgb(1, 2, 3), Bg: rgb(255, 128, 0)}, "\x1b[0;38;2;1;2;3;48;2;255;128;0m"},
		{"default foreground", screen.Style{Bg: idx(4)}, "\x1b[0;44m"},
		{"default background", screen.Style{Fg: idx(14)}, "\x1b[0;96m"},
		{"order", screen.Style{Fg: idx(3), Bg: idx(12), Attr: screen.Bold | screen.Underline | screen.Blink},
			"\x1b[0;1;4;5;33;104m"},
	}
	for _, tt := range tests {
		if got := string(appendSGR(nil, tt.style)); got != tt.want {
			t.Errorf("%s: appendSGR(%v) = %q, want %q", tt.name, tt.style, got, tt.want)
		}
	}
}

// TestReduce covers each case of section 7.3.
func TestReduce(t *testing.T) {
	tests := []struct {
		name   string
		c      screen.Colour
		colors int
		want   screen.Colour
	}{
		{"truecolor rgb", rgb(1, 2, 3), screen.ColorsTrue, rgb(1, 2, 3)},
		{"truecolor index", idx(200), screen.ColorsTrue, idx(200)},
		{"truecolor default", def, screen.ColorsTrue, def},
		{"256 rgb in the cube", rgb(255, 135, 0), screen.Colors256, idx(208)},
		{"256 rgb near a grey", rgb(128, 129, 127), screen.Colors256, idx(244)},
		{"256 rgb black", rgb(0, 0, 0), screen.Colors256, idx(16)},
		{"256 tie", rgb(115, 0, 0), screen.Colors256, idx(52)}, // 52 and 88 have the distance 400
		{"256 index 0-15", idx(9), screen.Colors256, idx(9)},
		{"256 index 16-255", idx(200), screen.Colors256, idx(200)},
		{"256 default", def, screen.Colors256, def},
		{"16 index 16-255", idx(196), screen.Colors16, idx(1)},
		{"16 grey", idx(240), screen.Colors16, idx(8)},
		{"16 rgb", rgb(250, 250, 250), screen.Colors16, idx(15)},
		{"16 rgb dark", rgb(0, 0, 0x99), screen.Colors16, idx(4)},
		{"16 index 0-15", idx(12), screen.Colors16, idx(12)},
		{"16 default", def, screen.Colors16, def},
	}
	for _, tt := range tests {
		if got := reduce(tt.c, tt.colors); got != tt.want {
			t.Errorf("%s: reduce(%v, %d) = %v, want %v", tt.name, tt.c, tt.colors, got, tt.want)
		}
	}
}

func TestPalette(t *testing.T) {
	for n, want := range map[uint32]uint32{
		0: 0x000000, 3: 0xaa5500, 15: 0xffffff,
		16: 0x000000, 17: 0x00005f, 52: 0x5f0000, 208: 0xff8700, 231: 0xffffff,
		232: 0x080808, 244: 0x808080, 255: 0xeeeeee,
	} {
		if got := palette(n); got != want {
			t.Errorf("palette(%d) = %06x, want %06x", n, got, want)
		}
	}
}

func TestAppendPage(t *testing.T) {
	page := rows(
		screen.Row{cell("a", idx(7), idx(0), 0), cell("b", idx(7), idx(0), 0), cell("c", idx(15), idx(4), screen.Bold)},
		screen.Row{cell(" ", idx(7), idx(0), 0), cell("é", def, def, 0), cell("d", def, def, 0)},
	)
	want := "\x1b[0m\x1b[2J" +
		"\x1b[1;1H\x1b[0;37;40mab\x1b[0;1;97;44mc" +
		"\x1b[2;1H\x1b[0;37;40m \x1b[0mé" + "d" +
		"\x1b[0m"
	if got := string(AppendPage(nil, page, screen.Colors16)); got != want {
		t.Errorf("AppendPage =\n%q\nwant\n%q", got, want)
	}
}

// rows gives a page of rows without a margin.
func rows(rs ...screen.Row) screen.Page {
	return screen.Page{Rows: rs, Cols: len(rs[0])}
}

// TestAppendPageMargin: the cells outside the screen area are spaces in Fill.
func TestAppendPageMargin(t *testing.T) {
	page := screen.Page{
		Rows: []screen.Row{{cell("x", idx(15), idx(4), 0)}},
		Cols: 4, Margin: 1, Fill: screen.Style{Fg: idx(7), Bg: idx(0)},
	}
	want := "\x1b[0m\x1b[2J\x1b[1;1H\x1b[0;37;40m \x1b[0;97;44mx\x1b[0;37;40m  \x1b[0m"
	if got := string(AppendPage(nil, page, screen.Colors16)); got != want {
		t.Errorf("AppendPage = %q, want %q", got, want)
	}
}

// TestAppendPageDefaultStart: the first cell in the default style needs no SGR.
func TestAppendPageDefaultStart(t *testing.T) {
	page := rows(screen.Row{cell("x", def, def, 0)})
	want := "\x1b[0m\x1b[2J\x1b[1;1Hx\x1b[0m"
	if got := string(AppendPage(nil, page, screen.Colors16)); got != want {
		t.Errorf("AppendPage = %q, want %q", got, want)
	}
}

func TestAppendPageWide(t *testing.T) {
	s := screen.Style{Fg: idx(7), Bg: idx(0)}
	page := rows(screen.Row{{Grapheme: "中", Style: s}, {Grapheme: "", Style: s}, {Grapheme: "a", Style: s}})
	got := string(AppendPage(nil, page, screen.Colors16))
	if n := strings.Count(got, "中"); n != 1 {
		t.Errorf("AppendPage gives %q %d times, want 1: %q", "中", n, got)
	}
	if !strings.Contains(got, "中a") {
		t.Errorf("AppendPage = %q, want %q", got, "中a")
	}
}

// TestAppendPageReducedEqual: two model styles with one reduced style give one SGR.
func TestAppendPageReducedEqual(t *testing.T) {
	page := rows(screen.Row{cell("a", rgb(250, 250, 250), def, 0), cell("b", rgb(255, 255, 255), def, 0)})
	tests := []struct {
		colors int
		want   string
	}{
		{screen.Colors16, "\x1b[0;97mab"},
		{screen.ColorsTrue, "\x1b[0;38;2;250;250;250ma\x1b[0;38;2;255;255;255mb"},
	}
	for _, tt := range tests {
		want := "\x1b[0m\x1b[2J\x1b[1;1H" + tt.want + "\x1b[0m"
		if got := string(AppendPage(nil, page, tt.colors)); got != want {
			t.Errorf("colors %d: AppendPage = %q, want %q", tt.colors, got, want)
		}
	}
}

func TestAppendPageAppends(t *testing.T) {
	got := string(AppendPage([]byte("x"), screen.Page{}, screen.Colors16))
	if want := "x\x1b[0m\x1b[2J\x1b[0m"; got != want {
		t.Errorf("AppendPage = %q, want %q", got, want)
	}
}
