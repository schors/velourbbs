package term

import (
	"slices"
	"strings"
	"testing"

	"github.com/schors/velourbbs/internal/screen"
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
		if got := reduce(tt.c, tt.colors, false); got != tt.want {
			t.Errorf("%s: reduce(%v, %d) = %v, want %v", tt.name, tt.c, tt.colors, got, tt.want)
		}
	}
}

func TestAppendPage(t *testing.T) {
	page := rows(
		screen.Row{cell("a", idx(7), idx(0), 0), cell("b", idx(7), idx(0), 0), cell("c", idx(15), idx(4), screen.Bold)},
		screen.Row{cell(" ", idx(7), idx(0), 0), cell("é", def, def, 0), cell("d", def, def, 0)},
	)
	want := "\x1b[?2026h\x1b[0m\x1b[2J" +
		"\x1b[1;1H\x1b[0;37;40mab\x1b[0;1;97;44mc" +
		"\x1b[2;1H\x1b[0;37;40m \x1b[0mé" + "d" +
		"\x1b[0m\x1b[?2026l"
	if got := string(AppendPage(nil, page, screen.Colors16, false)); got != want {
		t.Errorf("AppendPage =\n%q\nwant\n%q", got, want)
	}
}

// rows gives a page of rows without a margin.
func rows(rs ...screen.Row) screen.Page {
	return screen.Page{Rows: rs, Cols: len(rs[0])}
}

// TestAppendPageMargin: the cells outside the screen area are spaces in Normal.
func TestAppendPageMargin(t *testing.T) {
	page := screen.Page{
		Rows: []screen.Row{{cell("x", idx(15), idx(4), 0)}},
		Cols: 4, Margin: 1, Normal: screen.Style{Fg: idx(7), Bg: idx(0)},
	}
	want := "\x1b[?2026h\x1b[0m\x1b[2J\x1b[1;1H\x1b[0;37;40m \x1b[0;97;44mx\x1b[0;37;40m  \x1b[0m\x1b[?2026l"
	if got := string(AppendPage(nil, page, screen.Colors16, false)); got != want {
		t.Errorf("AppendPage = %q, want %q", got, want)
	}
}

// TestAppendPageDefaultStart: the first cell in the default style needs no SGR.
func TestAppendPageDefaultStart(t *testing.T) {
	page := rows(screen.Row{cell("x", def, def, 0)})
	want := "\x1b[?2026h\x1b[0m\x1b[2J\x1b[1;1Hx\x1b[0m\x1b[?2026l"
	if got := string(AppendPage(nil, page, screen.Colors16, false)); got != want {
		t.Errorf("AppendPage = %q, want %q", got, want)
	}
}

func TestAppendPageWide(t *testing.T) {
	s := screen.Style{Fg: idx(7), Bg: idx(0)}
	page := rows(screen.Row{{Grapheme: "中", Style: s}, {Grapheme: "", Style: s}, {Grapheme: "a", Style: s}})
	got := string(AppendPage(nil, page, screen.Colors16, false))
	if n := strings.Count(got, "中"); n != 1 {
		t.Errorf("AppendPage gives %q %d times, want 1: %q", "中", n, got)
	}
	if !strings.Contains(got, "中a") {
		t.Errorf("AppendPage = %q, want %q", got, "中a")
	}
}

// TestAppendPageResync: a grapheme of more than one code point gets the re-sync
// when a cell of the row follows it (section 7.4).
func TestAppendPageResync(t *testing.T) {
	s := screen.Style{Fg: idx(7), Bg: idx(0)}
	wide := func(g string) screen.Row { return screen.Row{{Grapheme: g, Style: s}, {Grapheme: "", Style: s}} }
	a := screen.Row{{Grapheme: "a", Style: s}}
	cat := func(rs ...screen.Row) screen.Row {
		var r screen.Row
		for _, x := range rs {
			r = append(r, x...)
		}
		return r
	}
	const start, end = "\x1b[?2026h\x1b[0m\x1b[2J\x1b[1;1H\x1b[0;37;40m", "\x1b[0m\x1b[?2026l"
	cases := []struct {
		name string
		row  screen.Row
		want string
	}{
		{"VS16 emoji", cat(a, wide("❤️"), a), "a❤️ \x1b[1;4Ha"},
		{"ZWJ sequence", cat(wide("👨‍👩‍👧"), a), "👨‍👩‍👧 \x1b[1;3Ha"},
		{"flag", cat(wide("🇺🇦"), a, a), "🇺🇦 \x1b[1;3Haa"},
		{"one code point", cat(wide("中"), a), "中a"},
		{"last cells", cat(a, wide("❤️")), "a❤️"},
	}
	for _, c := range cases {
		page := rows(c.row)
		if got := string(AppendPage(nil, page, screen.Colors16, false)); got != start+c.want+end {
			t.Errorf("%s: AppendPage =\n%q\nwant\n%q", c.name, got, start+c.want+end)
		}
	}
	// The re-sync uses the terminal row and the style of the grapheme.
	page := screen.Page{
		Rows: []screen.Row{cat(screen.Row{cell("❤️", idx(9), idx(4), 0), {Grapheme: "", Style: screen.Style{Fg: idx(9), Bg: idx(4)}}}, a)},
		Cols: 3, Height: 3, Top: 1,
	}
	want := "\x1b[2;1H\x1b[0;91;44m❤️ \x1b[2;3H\x1b[0;37;40ma"
	if got := string(AppendPage(nil, page, screen.Colors16, false)); !strings.Contains(got, want) {
		t.Errorf("AppendPage =\n%q\nwant it to contain\n%q", got, want)
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
		want := "\x1b[?2026h\x1b[0m\x1b[2J\x1b[1;1H" + tt.want + "\x1b[0m\x1b[?2026l"
		if got := string(AppendPage(nil, page, tt.colors, false)); got != want {
			t.Errorf("colors %d: AppendPage = %q, want %q", tt.colors, got, want)
		}
	}
}

// TestAppendPageFill: the fill rows of each fill (section 7.1). A page of 1 row on 3 terminal rows,
// with valign = "bottom" (Top 2) and with valign = "top" (Top 0).
func TestAppendPageFill(t *testing.T) {
	normal := screen.Style{Fg: idx(7), Bg: idx(0)}
	row := screen.Row{cell("x", idx(15), idx(4), 0)}
	tests := []struct {
		fill screen.Fill
		top  int
		want string
	}{
		{screen.FillNone, 2, "\x1b[?2026h\x1b[0m\x1b[2J\x1b[3;1H\x1b[0;97;44mx\x1b[0;37;40m \x1b[0m\x1b[?2026l"},
		{screen.FillSpaces, 2, "\x1b[?2026h\x1b[0m\x1b[2J\x1b[1;1H\x1b[0;37;40m  \x1b[2;1H  " +
			"\x1b[3;1H\x1b[0;97;44mx\x1b[0;37;40m \x1b[0m\x1b[?2026l"},
		{screen.FillSpaces, 0, "\x1b[?2026h\x1b[0m\x1b[2J\x1b[1;1H\x1b[0;97;44mx\x1b[0;37;40m " +
			"\x1b[2;1H  \x1b[3;1H  \x1b[0m\x1b[?2026l"},
		{screen.FillErase, 2, "\x1b[?2026h\x1b[0;37;40m\x1b[2J\x1b[3;1H\x1b[0;97;44mx\x1b[0;37;40m \x1b[0m\x1b[?2026l"},
	}
	for _, tt := range tests {
		page := screen.Page{Rows: []screen.Row{row}, Cols: 2, Height: 3, Top: tt.top, Normal: normal, Fill: tt.fill}
		if got := string(AppendPage(nil, page, screen.Colors16, false)); got != tt.want {
			t.Errorf("fill %s, top %d: AppendPage =\n%q\nwant\n%q", tt.fill, tt.top, got, tt.want)
		}
	}
}

// TestAppendPageErase: with fill = "erase", step 1 sends the reduced SGR of normal,
// and a cell in normal after it needs no SGR (section 7.2).
func TestAppendPageErase(t *testing.T) {
	normal := screen.Style{Fg: rgb(255, 255, 255), Bg: rgb(0, 0, 0)}
	page := screen.Page{
		Rows: []screen.Row{{{Grapheme: "a", Style: normal}}}, Cols: 1, Height: 1, Normal: normal, Fill: screen.FillErase,
	}
	want := "\x1b[?2026h\x1b[0;97;40m\x1b[2J\x1b[1;1Ha\x1b[0m\x1b[?2026l"
	if got := string(AppendPage(nil, page, screen.Colors16, false)); got != want {
		t.Errorf("AppendPage = %q, want %q", got, want)
	}
}

// plain gives a row of the text s: one cell for each code point, "_" for the second cell
// of a wide grapheme, in the style default.
func plain(s string) screen.Row {
	var r screen.Row
	for _, c := range s {
		g := string(c)
		if g == "_" {
			g = ""
		}
		r = append(r, screen.Cell{Grapheme: g})
	}
	return r
}

// TestRuns covers the runs of section 7.5: a gap of at most 3 equal cells joins two runs,
// and a run grows over a wide grapheme at its edges, over a grapheme of more than one code point
// before it and over the cell after such a grapheme at its end.
func TestRuns(t *testing.T) {
	const flag = "\U0001F1FA\U0001F1F8" // a wide grapheme of two code points
	tests := []struct {
		name     string
		old, new screen.Row
		want     []run
	}{
		{"none", plain("abcdef"), plain("abcdef"), nil},
		{"one", plain("abcdef"), plain("abXdef"), []run{{2, 3}}},
		{"gap 3", plain("abcdefgh"), plain("XbcdXfgh"), []run{{0, 5}}},
		{"gap 4", plain("abcdefgh"), plain("XbcdeXgh"), []run{{0, 1}, {5, 6}}},
		{"style", plain("ab"), screen.Row{{Grapheme: "a", Style: screen.Style{Attr: screen.Bold}}, {Grapheme: "b"}},
			[]run{{0, 1}}},
		{"second cell of a wide grapheme", plain("a世_b"), plain("a世_c"), []run{{3, 4}}},
		{"wide grapheme in the run", plain("abcd"), plain("a世_d"), []run{{1, 3}}},
		{"wide grapheme at the end", plain("ab世_"), plain("aX世_"), []run{{1, 2}}},
		{"change of the second cell", plain("a世_b"), append(plain("a世"), screen.Cell{Grapheme: "", Style: screen.Style{Attr: screen.Bold}}, screen.Cell{Grapheme: "b"}),
			[]run{{1, 3}}},
		{"long grapheme before", append(plain("a"), screen.Cell{Grapheme: flag}, screen.Cell{}, screen.Cell{Grapheme: "b"}),
			append(plain("a"), screen.Cell{Grapheme: flag}, screen.Cell{}, screen.Cell{Grapheme: "c"}), []run{{1, 4}}},
		{"long grapheme at the end", plain("abcd"),
			append(plain("a"), screen.Cell{Grapheme: "e\u0301"}, screen.Cell{Grapheme: "c"}, screen.Cell{Grapheme: "d"}),
			[]run{{1, 3}}},
		{"long grapheme at the row end", plain("ab"), append(plain("a"), screen.Cell{Grapheme: "e\u0301"}), []run{{1, 2}}},
		{"two runs grow into one", plain("abcdefgh"),
			append(plain("aX"), screen.Cell{Grapheme: "e\u0301"}, screen.Cell{Grapheme: "d"}, screen.Cell{Grapheme: "e"},
				screen.Cell{Grapheme: "f"}, screen.Cell{Grapheme: "X"}, screen.Cell{Grapheme: "h"}),
			[]run{{1, 7}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := runs(tc.new, tc.old); !slices.Equal(got, tc.want) {
				t.Errorf("runs = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestAppendUpdate covers the bytes of section 7.5: the synchronized output and the hidden cursor,
// the runs with their positions, the styles from default colours, the re-sync, a fill row
// with fill = "none" and the move to the cursor place.
func TestAppendUpdate(t *testing.T) {
	normal := screen.Style{Fg: idx(7), Bg: idx(0)}
	hi := screen.Style{Fg: idx(15), Bg: idx(4)}
	row := func(s string, style screen.Style) screen.Row {
		r := plain(s)
		for i := range r {
			r[i].Style = style
		}
		return r
	}
	prev := []screen.Row{row("    ", screen.Style{}), row("abcd", normal), row("efgh", normal)}
	g := []screen.Row{row("  x ", screen.Style{}), row("abcd", normal),
		{{Grapheme: "e\u0301", Style: hi}, {Grapheme: "f", Style: normal}, {Grapheme: "g", Style: normal}, {Grapheme: "Z", Style: hi}}}
	want := "\x1b[?2026h\x1b[?25l" +
		"\x1b[1;3Hx" +
		"\x1b[3;1H\x1b[0;97;44me\u0301 \x1b[3;2H\x1b[0;37;40mfg\x1b[0;97;44mZ" +
		"\x1b[0m\x1b[3;4H\x1b[?25h\x1b[?2026l"
	if got := string(AppendUpdate(nil, g, prev, 3, 4, screen.Colors16, false)); got != want {
		t.Errorf("AppendUpdate =\n%q\nwant\n%q", got, want)
	}
	if got := AppendUpdate([]byte("x"), prev, prev, 3, 4, screen.Colors16, false); string(got) != "x" {
		t.Errorf("AppendUpdate of the same grid = %q, want only the bytes before", got)
	}
}

// TestCursorPlace: the last terminal row that the page writes, at the column term.cols (section 7.5).
func TestCursorPlace(t *testing.T) {
	r := screen.Row{{Grapheme: "a"}}
	tests := []struct {
		name     string
		p        screen.Page
		row, col int
	}{
		{"rows", screen.Page{Rows: []screen.Row{r, r}, Cols: 80, Height: 24, Top: 3}, 5, 80},
		{"fill spaces", screen.Page{Rows: []screen.Row{r}, Cols: 80, Height: 24, Fill: screen.FillSpaces}, 24, 80},
		{"fill erase", screen.Page{Rows: []screen.Row{r}, Cols: 80, Height: 24, Fill: screen.FillErase}, 1, 80},
		{"no rows", screen.Page{Cols: 80, Height: 24}, 1, 1},
	}
	for _, tc := range tests {
		if row, col := CursorPlace(tc.p); row != tc.row || col != tc.col {
			t.Errorf("%s: CursorPlace = %d, %d, want %d, %d", tc.name, row, col, tc.row, tc.col)
		}
	}
}

func TestAppendPageAppends(t *testing.T) {
	got := string(AppendPage([]byte("x"), screen.Page{}, screen.Colors16, false))
	if want := "x\x1b[?2026h\x1b[0m\x1b[2J\x1b[0m\x1b[?2026l"; got != want {
		t.Errorf("AppendPage = %q, want %q", got, want)
	}
}

// TestAppendPageVGA: a VGA colour of art goes out as its index with a VGA palette,
// and as its RGB value reduced to term.colors with a palette that is not VGA.
// A theme index does not change (section 7.3).
func TestAppendPageVGA(t *testing.T) {
	vga := screen.VGAColour
	page := rows(screen.Row{cell("a", vga(1), vga(0), 0), cell("b", idx(1), idx(0), 0)})
	tests := []struct {
		colors int
		vgaRGB bool
		want   string
	}{
		{screen.Colors16, false, "\x1b[0;31;40mab"},
		{screen.Colors256, false, "\x1b[0;31;40mab"},
		{screen.ColorsTrue, false, "\x1b[0;31;40mab"},
		{screen.Colors16, true, "\x1b[0;31;40mab"},
		{screen.Colors256, true, "\x1b[0;38;5;124;48;5;16ma\x1b[0;31;40mb"},
		{screen.ColorsTrue, true, "\x1b[0;38;2;170;0;0;48;2;0;0;0ma\x1b[0;31;40mb"},
	}
	for _, tt := range tests {
		want := "\x1b[?2026h\x1b[0m\x1b[2J\x1b[1;1H" + tt.want + "\x1b[0m\x1b[?2026l"
		if got := string(AppendPage(nil, page, tt.colors, tt.vgaRGB)); got != want {
			t.Errorf("colors %d, vgaRGB %v: AppendPage = %q, want %q", tt.colors, tt.vgaRGB, got, want)
		}
	}
}

// TestReduceVGA16: with 16 colours, the RGB value of each VGA colour becomes
// the same index again (section 7.3).
func TestReduceVGA16(t *testing.T) {
	for n := range uint8(16) {
		if got := reduce(screen.VGAColour(n), screen.Colors16, true); got != idx(n) {
			t.Errorf("reduce(VGA %d, 16, true) = %v, want %d", n, got, n)
		}
	}
}

// TestAppendLinesExample is the example of screen-render, section 3.7.
func TestAppendLinesExample(t *testing.T) {
	normal := screen.Style{Fg: idx(7), Bg: idx(0)}
	page := screen.Page{
		Num: 1, Cols: 80, Height: 10, Normal: normal,
		Rows: []screen.Row{
			append(text("Welcome!", screen.Style{Fg: idx(15), Bg: idx(4), Attr: screen.Bold}), text(strings.Repeat(" ", 72), normal)...),
			append(text("[G]oodbye", screen.Style{Fg: idx(14), Bg: idx(0)}), text(strings.Repeat(" ", 71), normal)...),
		},
	}
	want := "page 1 80x10 fill none\n" +
		"\x1b[0;1;97;44mWelcome!\x1b[0;37;40m" + strings.Repeat(" ", 72) + "\x1b[0m\n" +
		"\x1b[0;96;40m[G]oodbye\x1b[0;37;40m" + strings.Repeat(" ", 71) + "\x1b[0m\n" +
		strings.Repeat("\x1b[0m\n", 8)
	if got := string(AppendLines(nil, page, screen.Colors256)); got != want {
		t.Errorf("AppendLines =\n%q\nwant\n%q", got, want)
	}
}

// text gives a cell for each rune of s in style st.
func text(s string, st screen.Style) screen.Row {
	var r screen.Row
	for _, c := range s {
		r = append(r, screen.Cell{Grapheme: string(c), Style: st})
	}
	return r
}

// TestAppendLinesFill: the fill rows above and below the rows for each fill.
// Each row starts with default colours without attributes, so a row in normal
// after a fill row in normal gets the SGR again.
func TestAppendLinesFill(t *testing.T) {
	normal := screen.Style{Fg: idx(7), Bg: idx(0)}
	tests := []struct {
		fill screen.Fill
		row  string // a fill row
	}{
		{screen.FillNone, "\x1b[0m\n"},
		{screen.FillSpaces, "\x1b[0;37;40m  \x1b[0m\n"},
		{screen.FillErase, "\x1b[0;37;40m  \x1b[0m\n"},
	}
	for _, tt := range tests {
		page := screen.Page{
			Num: 2, Cols: 2, Height: 4, Top: 1, Normal: normal, Fill: tt.fill,
			Rows: []screen.Row{text("ab", normal)},
		}
		want := "page 2 2x4 fill " + tt.fill.String() + "\n" +
			tt.row + "\x1b[0;37;40mab\x1b[0m\n" + tt.row + tt.row
		if got := string(AppendLines(nil, page, screen.Colors16)); got != want {
			t.Errorf("fill %s: AppendLines =\n%q\nwant\n%q", tt.fill, got, want)
		}
	}
}

// TestAppendLinesResync: the re-sync moves to the column with ESC[<col>G (screen-render, section 3.7).
func TestAppendLinesResync(t *testing.T) {
	s := screen.Style{Fg: idx(9), Bg: idx(4)}
	page := screen.Page{
		Num: 1, Cols: 4, Height: 1,
		Rows: []screen.Row{{{Grapheme: "a", Style: s}, {Grapheme: "❤️", Style: s}, {Grapheme: "", Style: s}, {Grapheme: "b", Style: s}}},
	}
	want := "page 1 4x1 fill none\n\x1b[0;91;44ma❤️ \x1b[4Gb\x1b[0m\n"
	if got := string(AppendLines(nil, page, screen.Colors16)); got != want {
		t.Errorf("AppendLines =\n%q\nwant\n%q", got, want)
	}
}

// TestAppendLinesReduce: the colours have the reduction to term.colors,
// and a VGA colour goes out as its RGB value (screen-render, section 3.7).
func TestAppendLinesReduce(t *testing.T) {
	page := screen.Page{
		Num: 1, Cols: 3, Height: 1,
		Rows: []screen.Row{{
			cell("a", screen.VGAColour(3), def, 0),
			cell("b", rgb(250, 250, 250), def, 0),
			cell("c", idx(196), idx(4), 0),
		}},
	}
	tests := []struct {
		colors int
		want   string
	}{
		{screen.ColorsTrue, "\x1b[0;38;2;170;85;0ma\x1b[0;38;2;250;250;250mb\x1b[0;38;5;196;44mc"},
		{screen.Colors256, "\x1b[0;38;5;130ma\x1b[0;38;5;231mb\x1b[0;38;5;196;44mc"},
		{screen.Colors16, "\x1b[0;33ma\x1b[0;97mb\x1b[0;31;44mc"},
	}
	for _, tt := range tests {
		want := "page 1 3x1 fill none\n" + tt.want + "\x1b[0m\n"
		if got := string(AppendLines(nil, page, tt.colors)); got != want {
			t.Errorf("colors %d: AppendLines =\n%q\nwant\n%q", tt.colors, got, want)
		}
	}
}
