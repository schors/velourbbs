package effect_test

import (
	"encoding/binary"
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/schors/velourbbs/internal/bbt"
	"github.com/schors/velourbbs/internal/effect"
	"github.com/schors/velourbbs/internal/screen"
)

var theme = map[string]screen.Style{
	"normal": {Fg: screen.IndexColour(7), Bg: screen.IndexColour(0)}, "more": {Fg: screen.IndexColour(15), Bg: screen.IndexColour(0)},
	"flame": {Fg: screen.IndexColour(11), Bg: screen.IndexColour(1)},
}

// sauceArt gives an .ans file in the flat form: a BOM, data, 0x1A and a character SAUCE of the width.
func sauceArt(data string, width int) string {
	r := make([]byte, 128)
	copy(r, "SAUCE00")
	r[94], r[95] = 1, 1
	binary.LittleEndian.PutUint16(r[96:], uint16(width))
	return "\xef\xbb\xbf" + data + "\x1a" + string(r)
}

// page loads the files of screens/ and gives the last page of the screen s at cols x rows.
func page(t *testing.T, files map[string]string, cols, rows int) (*screen.Screen, screen.Page) {
	t.Helper()
	fsys := fstest.MapFS{}
	for n, data := range files {
		fsys["screens/"+n] = &fstest.MapFile{Data: []byte(data)}
	}
	fields := screen.Fields{}
	clock, err := screen.NewAlias("clock", `bbs.time live "15:04:05"`)
	if err != nil {
		t.Fatal(err)
	}
	fields["clock"] = clock
	sets, ds := bbt.Load(fsys, bbt.Settings{Theme: theme, Layout: screen.Layout{Width: screen.WidthLimits{Min: 10, Max: 40}}, Fields: fields})
	if ds.HasErrors() {
		t.Fatalf("diags: %v", ds)
	}
	s := sets["en"]["s"]
	env := &screen.Env{Term: screen.Term{Type: "ansi", Cols: cols, Rows: rows, Colors: 256},
		Session: screen.SessionValues{Time: "15:04:05", Clock: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)}}
	var last screen.Page
	for p := range screen.Render(s, env, theme, "more") {
		last = p
	}
	return s, last
}

// cells gives the layer of the screen s on the page p at the time t with the seed (1, 2),
// in a short form: row:col grapheme style, from 0.
func cells(s *screen.Screen, p screen.Page, t time.Duration) string {
	l := effect.New(s.Effects, p, theme, rand.New(rand.NewPCG(1, 2)))
	var out []string
	for _, c := range l.Cells(t, p.Grid()) {
		out = append(out, fmt.Sprintf("%d:%d %s %s", c.Row, c.Col, c.Cell.Grapheme, c.Cell.Style))
	}
	return strings.Join(out, "; ")
}

// candle is the front matter of a candle in the place at, with in when in is not "".
func candle(at, in string) string {
	s := "+++\nwidth = [10, 40]\n[[effect]]\ndo = \"candle\"\nchars = \"^\"\ncolors = [\"flame\"]\nat = " + at + "\n"
	if in != "" {
		s += "in = \"" + in + "\"\n"
	}
	return s + "+++\n"
}

// A place with in moves with the art: the text above the art wraps at a smaller width
// (live-screen, section 6.2). The cell keeps its background: the style of the cell
// with the foreground of the effect.
func TestPlaceInArt(t *testing.T) {
	files := map[string]string{
		"s.bbt": candle("[[2, 1], [-1, -1]]", "a.ans") + "one two three\n{>a.ans}",
		"a.ans": sauceArt("\x1b[44m.....\n.....", 5),
	}
	tests := []struct {
		cols int
		want string
	}{
		{40, "1:1 ^ 11/4; 2:4 ^ 11/4"},
		{10, "2:1 ^ 11/4; 3:4 ^ 11/4"},
	}
	for _, tt := range tests {
		s, p := page(t, files, tt.cols, 24)
		if got := cells(s, p, 0); got != tt.want {
			t.Errorf("%d cols: got %q, want %q", tt.cols, got, tt.want)
		}
	}
}

// A place with in of a fragment moves with the fragment: the text above it wraps at a smaller width.
// Column 1 is the left edge of the screen area, and -1 is its last column and the last fragment row.
// The host cuts the place to the fragment rows: a point below or above them has no cell (live-screen, section 6.2).
func TestPlaceInFragment(t *testing.T) {
	tests := []struct {
		at   string
		cols int
		want string
	}{
		{"[[2, 1], [-1, -1]]", 40, "1:1 ^ 11/0; 2:39 ^ 11/0"},
		{"[[2, 1], [-1, -1]]", 10, "2:1 ^ 11/0; 3:9 ^ 11/0"},
		{"[[1, 3], [1, -3]]", 40, ""},
	}
	for _, tt := range tests {
		files := map[string]string{"s.bbt": candle(tt.at, "_f") + "one two three\n{>_f}\nend", "_f.bbt": "abc\ndef"}
		s, p := page(t, files, tt.cols, 24)
		if got := cells(s, p, 0); got != tt.want {
			t.Errorf("%s at %d cols: got %q, want %q", tt.at, tt.cols, got, tt.want)
		}
	}
	// A fragment row has W cells also when term.cols is smaller: -1 is the column W, cut at term.cols.
	files := map[string]string{"s.bbt": strings.Replace(candle("[[-1, 1], [20, 1]]", "_f"), "[10, 40]", "[40, 40]", 1) + "{>_f}", "_f.bbt": "abc"}
	s, p := page(t, files, 20, 24)
	if got, want := cells(s, p, 0), "0:19 ^ 11/0"; got != want {
		t.Errorf("cut at term.cols: got %q, want %q", got, want)
	}
}

// A negative coordinate counts from the end of the screen area and of the page, or of the art grid.
// The screen area has a margin, and the page has fill rows above it (live-screen, section 6.2).
func TestPlaceNegative(t *testing.T) {
	tests := []struct {
		front, body, want string
	}{
		// W = 10 in the middle of 20 columns: the margin is 5. The page of 2 rows is in the middle of 10 rows: Top is 4.
		{"[-1, -1]", "a\nb", "5:14 ^ 11/0"},
		{"[-10, -2]", "a\nb", "4:5 ^ 11/0"},
		// A rectangle from the last row reaches the margin and the fill rows: the host cuts it to the grid.
		{"[-1, -1, 10, 10]", "a\nb", "5:14 ^ 11/0; 5:15 ^ 11/0; 5:16 ^ 11/0; 5:17 ^ 11/0; 5:18 ^ 11/0; 5:19 ^ 11/0; " +
			"6:14 ^ 11/-; 6:15 ^ 11/-; 6:16 ^ 11/-; 6:17 ^ 11/-; 6:18 ^ 11/-; 6:19 ^ 11/-; " +
			"7:14 ^ 11/-; 7:15 ^ 11/-; 7:16 ^ 11/-; 7:17 ^ 11/-; 7:18 ^ 11/-; 7:19 ^ 11/-; " +
			"8:14 ^ 11/-; 8:15 ^ 11/-; 8:16 ^ 11/-; 8:17 ^ 11/-; 8:18 ^ 11/-; 8:19 ^ 11/-; " +
			"9:14 ^ 11/-; 9:15 ^ 11/-; 9:16 ^ 11/-; 9:17 ^ 11/-; 9:18 ^ 11/-; 9:19 ^ 11/-"},
	}
	for _, tt := range tests {
		files := map[string]string{"s.bbt": strings.Replace(candle(tt.front, ""), "width = [10, 40]", "width = [10, 10]\nvalign = \"middle\"", 1) + tt.body}
		s, p := page(t, files, 20, 10)
		if got := cells(s, p, 0); got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.front, got, tt.want)
		}
	}
	// In the art: the last cell of a 3x2 art below one text row.
	files := map[string]string{"s.bbt": candle("[-1, -1]", "a.ans") + "a\n{>a.ans}", "a.ans": sauceArt("abc\ndef", 3)}
	s, p := page(t, files, 40, 24)
	if got, want := cells(s, p, 0), "2:2 ^ 11/0"; got != want {
		t.Errorf("art: got %q, want %q", got, want)
	}
}

// A slot cell and the cells of a grapheme of width 2 are not in a place (live-screen, section 6.2).
func TestPlaceFreeCells(t *testing.T) {
	files := map[string]string{"s.bbt": candle("[1, 1, 13, 1]", "") + "{clock}語ab"}
	s, p := page(t, files, 40, 24)
	if got, want := cells(s, p, 0), "0:10 ^ 11/0; 0:11 ^ 11/0; 0:12 ^ 11/0"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// A cell that the page does not show is not in a place: an art or a fragment cut at W, the rows
// of an art or a fragment on a page before a More row, and an art or a fragment in a false block
// (live-screen, section 6.2).
func TestPlaceShown(t *testing.T) {
	tests := []struct {
		in, at, body string
		rows         int
		want         string
	}{
		// W = 10 in the middle of 40 columns: the margin is 15.
		{"a.ans", "[[12, 1], [10, 1]]", "{>a.ans}", 24, "0:24 ^ 11/0"},
		{"_f", "[[12, 1], [10, 1]]", "{>_f}", 24, "0:24 ^ 11/0"},
		// 10 rows: page 1 has 9 rows and the More row. The last page starts at row 10 of the art or the fragment.
		{"a.ans", "[[1, 8], [1, 9], [1, 12]]", "{>a.ans}", 10, "2:15 ^ 11/0"},
		{"_f", "[[1, 8], [1, 9], [1, 12]]", "{>_f}", 10, "2:15 ^ 11/0"},
		{"a.ans", "[1, 1]", "{#if term.cols > 100}\n{>a.ans}\n{/if}\nabc", 24, ""},
		{"_f", "[1, 1]", "{#if term.cols > 100}\n{>_f}\n{/if}\nabc", 24, ""},
	}
	art := sauceArt(strings.Repeat("abcdefghijkl\n", 12), 12)
	for _, tt := range tests {
		files := map[string]string{"s.bbt": strings.Replace(candle(tt.at, tt.in), "[10, 40]", "[10, 10]", 1) + tt.body,
			"a.ans": art, "_f.bbt": strings.Repeat("abcdefghij\n", 12)}
		s, p := page(t, files, 40, tt.rows)
		if got := cells(s, p, 0); got != tt.want {
			t.Errorf("%s %s: got %q, want %q", tt.in, tt.at, got, tt.want)
		}
	}
}

// The first event starts at the show time plus delay. Each next event starts d after the event before:
// every plus a random value from -spread to spread (live-screen, section 6.3).
func TestEvents(t *testing.T) {
	files := map[string]string{"s.bbt": "+++\n[[effect]]\ndo = \"candle\"\nchars = \"^'\"\ncolors = [\"flame\", \"9\"]\nat = [1, 1, 4, 1]\n" +
		"every = 0.5\nspread = 0\ndelay = 1\n" +
		"[[effect]]\ndo = \"twinkle\"\nchars = \"ab\"\nat = [5, 1]\nevery = 0.2\n+++\nxxxxa"}
	s, p := page(t, files, 40, 24)
	l := effect.New(s.Effects, p, theme, rand.New(rand.NewPCG(1, 2)))
	grid := p.Grid()
	var times []time.Duration
	for {
		d, ok := l.Next()
		if !ok {
			t.Fatal("no next event")
		}
		if d > 2*time.Second {
			break
		}
		times = append(times, d)
		l.Cells(d, grid)
	}
	// The twinkle has no delay: its first event is at 0 s, the second at 0.1–0.3 s.
	// The candle has delay 1 and no spread: 1 s, 1.5 s and 2 s.
	if times[0] != 0 || times[1] < 100*time.Millisecond || times[1] > 300*time.Millisecond ||
		!slices.Contains(times, time.Second) || !slices.Contains(times, 1500*time.Millisecond) ||
		!slices.Contains(times, 2*time.Second) {
		t.Errorf("event times %v", times)
	}
	// The same seed gives the same layer.
	if a, b := cells(s, p, 3*time.Second), cells(s, p, 3*time.Second); a != b {
		t.Errorf("layers differ:\n%s\n%s", a, b)
	}
}

// eventTimes gives the start times of the events of the screen s on the page p up to the time end.
func eventTimes(t *testing.T, s *screen.Screen, p screen.Page, end time.Duration) []time.Duration {
	t.Helper()
	l := effect.New(s.Effects, p, theme, rand.New(rand.NewPCG(1, 2)))
	grid := p.Grid()
	var out []time.Duration
	for {
		d, ok := l.Next()
		if !ok || d > end {
			return out
		}
		out = append(out, d)
		l.Cells(d, grid)
	}
}

// A place has times events, and no event after them (live-screen, section 6.3).
func TestTimes(t *testing.T) {
	files := map[string]string{"s.bbt": "+++\n[[effect]]\ndo = \"candle\"\nchars = \"^'\"\ncolors = [\"flame\"]\nat = [1, 1]\n" +
		"every = 0.5\nspread = 0\ntimes = 2\n+++\nx"}
	s, p := page(t, files, 40, 24)
	if got, want := eventTimes(t, s, p, time.Hour), []time.Duration{0, 500 * time.Millisecond}; !slices.Equal(got, want) {
		t.Errorf("event times %v, want %v", got, want)
	}
}

// The first event starts at the delay plus a random value from -delay_spread to delay_spread,
// and each place takes its own value. A sweep starts its runs as the other effects
// (live-screen, sections 6.3 and 6.8).
func TestFirstEvent(t *testing.T) {
	files := map[string]string{"s.bbt": "+++\n[[effect]]\ndo = \"candle\"\nchars = \"^\"\ncolors = [\"flame\"]\nat = [[1, 1], [2, 1]]\n" +
		"every = 10\nspread = 0\ndelay = 2\ndelay_spread = 1\ntimes = 1\n+++\nxx"}
	s, p := page(t, files, 40, 24)
	got := eventTimes(t, s, p, time.Hour)
	if len(got) != 2 || got[0] == got[1] || got[0] < time.Second || got[1] > 3*time.Second {
		t.Errorf("event times %v, want two different times in 1–3 s", got)
	}
	files = map[string]string{"s.bbt": "+++\n[[effect]]\ndo = \"sweep\"\ndelay = 0.3\nevery = 2\ntimes = 2\nspeed = 1000\n+++\nx"}
	s, p = page(t, files, 40, 24)
	if got := eventTimes(t, s, p, time.Hour); len(got) == 0 || got[0] != 300*time.Millisecond ||
		!slices.Contains(got, 2300*time.Millisecond) || got[len(got)-1] > 3*time.Second {
		t.Errorf("sweep event times %v, want runs at 0.3 s and 2.3 s", got)
	}
}

// The layer draws over the current base: after a new base (a cursor move), a drawn cell takes
// the background of the new cell, and a cell that is now a slot or a part of a wide grapheme
// is not drawn (live-screen, sections 4, 6.2 and 6.3).
func TestNewBase(t *testing.T) {
	s, p := page(t, map[string]string{"s.bbt": candle("[1, 1, 3, 1]", "") + "abc"}, 40, 24)
	l := effect.New(s.Effects, p, theme, rand.New(rand.NewPCG(1, 2)))
	grid := p.Grid()
	grid[0][0].Style.Bg = screen.IndexColour(4)
	grid[0][1].Slot = 1
	grid[0][2].Grapheme, grid[0][3].Grapheme = "語", ""
	var got []string
	for _, c := range l.Cells(0, grid) {
		got = append(got, fmt.Sprintf("%d:%d %s %s", c.Row, c.Col, c.Cell.Grapheme, c.Cell.Style))
	}
	if want := "0:0 ^ 11/4"; strings.Join(got, "; ") != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// A twinkle point with a grapheme of chars keeps it over the effects before it until its first change
// (live-screen, section 6.4). A twinkle rectangle without stars keeps its events: they find no cell.
func TestTwinkleStars(t *testing.T) {
	files := map[string]string{"s.bbt": "+++\n[[effect]]\ndo = \"candle\"\nchars = \"b\"\ncolors = [\"1\", \"2\"]\nat = [1, 1]\n" +
		"[[effect]]\ndo = \"twinkle\"\nchars = \"bc\"\nat = [1, 1]\ndelay = 100\n" +
		"[[effect]]\ndo = \"twinkle\"\nchars = \"xy\"\nat = [2, 1, 3, 1]\n+++\nabcd"}
	s, p := page(t, files, 40, 24)
	if got, want := cells(s, p, 6*time.Second), "0:0 b 2/0; 0:0 b 1/0"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	files["s.bbt"] = "+++\n[[effect]]\ndo = \"twinkle\"\nchars = \"xy\"\nat = [1, 1, 3, 1]\n+++\nabc"
	s, p = page(t, files, 40, 24)
	if _, ok := effect.New(s.Effects, p, theme, rand.New(rand.NewPCG(1, 2))).Next(); !ok {
		t.Error("a rectangle without stars has no events")
	}
}

// A glint takes a random cell that is not a space. At the time t of the event, the cell is the mix
// with peak at the amount 1 - t / time (live-screen, section 6.7). While the event runs,
// the layer changes each 50 ms; it changes at the end of the event too.
func TestGlint(t *testing.T) {
	files := map[string]string{"s.bbt": "+++\n[[effect]]\ndo = \"glint\"\nat = [1, 1, 3, 1]\ndelay = 1\nevery = 1\nspread = 0\ntime = 0.5\n+++\na b"}
	s, p := page(t, files, 40, 24)
	for _, tt := range []struct {
		t    time.Duration
		want string
	}{
		{900 * time.Millisecond, ""},
		{time.Second, "0:0 a #ffffff/0"},
		{1250 * time.Millisecond, "0:0 a #d5d5d5/0"},
		{1500 * time.Millisecond, ""},
		{7 * time.Second, "0:2 b #ffffff/0"},
	} {
		if got := cells(s, p, tt.t); got != tt.want {
			t.Errorf("%v: got %q, want %q", tt.t, got, tt.want)
		}
	}
	l := effect.New(s.Effects, p, theme, rand.New(rand.NewPCG(1, 2)))
	grid := p.Grid()
	for _, tt := range []struct{ t, next time.Duration }{
		{0, time.Second},
		{time.Second, 1050 * time.Millisecond},
		{1480 * time.Millisecond, 1500 * time.Millisecond},
		{1500 * time.Millisecond, 2 * time.Second},
	} {
		l.Cells(tt.t, grid)
		if next, ok := l.Next(); !ok || next != tt.next {
			t.Errorf("after %v: next %v %v, want %v", tt.t, next, ok, tt.next)
		}
	}
}

// A glint takes the cell after the effects before it. On a cell of two events, the event
// that started later wins (live-screen, section 6.3). A point on a space has no glint.
func TestGlintOver(t *testing.T) {
	files := map[string]string{"s.bbt": "+++\n[[effect]]\ndo = \"candle\"\nchars = \"^\"\ncolors = [\"11\"]\nat = [1, 1]\nevery = 100\n" +
		"[[effect]]\ndo = \"glint\"\nat = [[1, 1], [2, 1]]\nevery = 0.2\nspread = 0\ntime = 0.5\n+++\nx "}
	s, p := page(t, files, 40, 24)
	// At 0.5 s, the event of 0.2 s has the amount 0.4 and the event of 0.4 s the amount 0.8.
	if got, want := cells(s, p, 500*time.Millisecond), "0:0 ^ 11/0; 0:0 ^ #ffffdd/0"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// The band of a sweep runs over the lit cells of all rows at the same time. The last lit cell
// of each row starts a glint when the centre is at its column (live-screen, section 6.8).
// Without every, the sweep has one event.
func TestSweep(t *testing.T) {
	// L = 0 and R = 2: the centre runs from -2 to 4 at 10 columns each second, in 0.6 s.
	// The glints start at 0.2 s (row 2) and 0.4 s (row 1) and end at 0.4 s and 0.6 s.
	files := map[string]string{"s.bbt": "+++\n[[effect]]\ndo = \"sweep\"\ndelay = 0\nwidth = 2\nspeed = 10\ntime = 0.2\n+++\nabc\na"}
	s, p := page(t, files, 40, 24)
	for _, tt := range []struct {
		t    time.Duration
		want string
	}{
		{100 * time.Millisecond, "0:0 a #d5d5d5/0; 1:0 a #d5d5d5/0"},
		// The centre at 1: the band gives 0.5, 1 and 0.5; the glint of row 2 has 0.5.
		{300 * time.Millisecond, "0:0 a #d5d5d5/0; 0:1 b #ffffff/0; 0:2 c #d5d5d5/0; 1:0 a #d5d5d5/0"},
		// The centre at 2.5: the band gives 0.25 and 0.75; the glint of row 1 has 0.75.
		{450 * time.Millisecond, "0:1 b #bfbfbf/0; 0:2 c #eaeaea/0"},
		{600 * time.Millisecond, ""},
	} {
		if got := cells(s, p, tt.t); got != tt.want {
			t.Errorf("%v: got %q, want %q", tt.t, got, tt.want)
		}
	}
	l := effect.New(s.Effects, p, theme, rand.New(rand.NewPCG(1, 2)))
	l.Cells(600*time.Millisecond, p.Grid())
	if next, ok := l.Next(); ok {
		t.Errorf("a sweep without every has a next event at %v", next)
	}
}

// An event takes its cells from the current base: after a new base, a cell that is now a part
// of a wide grapheme is not a lit cell of a sweep (live-screen, sections 6.2 and 6.3).
func TestSweepNewBase(t *testing.T) {
	files := map[string]string{"s.bbt": "+++\n[[effect]]\ndo = \"sweep\"\nat = [1, 1, 3, 1]\ndelay = 0\nwidth = 2\nspeed = 10\n+++\nabc"}
	s, p := page(t, files, 40, 24)
	l := effect.New(s.Effects, p, theme, rand.New(rand.NewPCG(1, 2)))
	grid := p.Grid()
	grid[0][2].Grapheme, grid[0][3].Grapheme = "語", ""
	// The last lit cell is b: its glint starts at 0.3 s. At 0.45 s, the band gives b 0.25 and the glint 0.75.
	var got []string
	for _, c := range l.Cells(450*time.Millisecond, grid) {
		got = append(got, fmt.Sprintf("%d:%d %s %s", c.Row, c.Col, c.Cell.Grapheme, c.Cell.Style))
	}
	if want := "0:1 b #eaeaea/0"; strings.Join(got, "; ") != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// With over = false, only an empty cell, a space, can take a star; with over = true, each cell
// of the place can. A star shows for time. A cell with a star of the effect cannot take another
// (live-screen, section 6.5).
func TestSparkle(t *testing.T) {
	front := "+++\n[[effect]]\ndo = \"sparkle\"\nchars = \"*\"\nat = [1, 1, 4, 1]\ndelay = 1\nevery = 1\nspread = 0\ntime = 0.5\ncount = 4\n"
	for _, tt := range []struct {
		over string
		t    time.Duration
		want string
	}{
		{"false", 900 * time.Millisecond, ""},
		{"false", time.Second, "0:1 * 7/0; 0:3 * 7/0"},
		{"false", 1500 * time.Millisecond, ""},
		{"true", time.Second, "0:0 * 7/0; 0:1 * 7/0; 0:2 * 7/0; 0:3 * 7/0"},
	} {
		s, p := page(t, map[string]string{"s.bbt": front + "over = " + tt.over + "\n+++\na b"}, 40, 24)
		if got := sorted(cells(s, p, tt.t)); got != tt.want {
			t.Errorf("over = %s, %v: got %q, want %q", tt.over, tt.t, got, tt.want)
		}
	}
	// Each second, one star of 2.5 s: the third event finds the two empty cells with a star.
	// The fourth event at 4 s finds the cell of the first star free again.
	files := map[string]string{"s.bbt": "+++\n[[effect]]\ndo = \"sparkle\"\nchars = \"*\"\nat = [1, 1, 3, 1]\ndelay = 1\nevery = 1\nspread = 0\ntime = 2.5\n+++\na"}
	s, p := page(t, files, 40, 24)
	for _, tt := range []struct {
		t time.Duration
		n int
	}{{time.Second, 1}, {2 * time.Second, 2}, {3 * time.Second, 2}, {3500 * time.Millisecond, 1}, {4 * time.Second, 2}} {
		if got := cells(s, p, tt.t); strings.Count(got, "*") != tt.n || strings.Contains(got, "0:0") {
			t.Errorf("%v: got %q, want %d stars on empty cells", tt.t, got, tt.n)
		}
	}
}

// With over = false, the bolt draws only on empty cells; with over = true, on each cell of the place.
// A drawn cell has the foreground of peak, the background of the cell and no attributes. In a rectangle
// of one column, each step is a step 0 and no branch has a cell (live-screen, section 6.9).
func TestLightning(t *testing.T) {
	// Column 2 has x on the rows 1 and 3 and spaces on the rows 2 and 4.
	front := "+++\n[[effect]]\ndo = \"lightning\"\nshape = \"slant\"\nchars = \"\\\\|/\"\npeak = \"15\"\n" +
		"at = [2, 1, 1, 4]\nevery = 1\nspread = 0\ntime = 0.5\n"
	layer := func(over string, at time.Duration) []string {
		s, p := page(t, map[string]string{"s.bbt": front + "over = " + over + "\n+++\nax\na\nax\na"}, 40, 24)
		l := effect.New(s.Effects, p, theme, rand.New(rand.NewPCG(1, 2)))
		grid := p.Grid()
		for _, row := range grid {
			row[1].Style.Attr = screen.Bold
		}
		var out []string
		for _, c := range l.Cells(at, grid) {
			out = append(out, fmt.Sprintf("%d:%d %s %s", c.Row, c.Col, c.Cell.Grapheme, c.Cell.Style))
		}
		return out
	}
	// The main part runs down from the first row of the area, the same rows with each value of over.
	all := layer("true", time.Second)
	if len(all) == 0 {
		t.Fatal("over = true: no cell")
	}
	var empty []string
	for k, c := range all {
		var row int
		if _, err := fmt.Sscanf(c, "%d:", &row); err != nil {
			t.Fatal(err)
		}
		if want := fmt.Sprintf("%d:1 | 15/0", row); c != want || k > 0 && !strings.HasPrefix(all[k-1], fmt.Sprintf("%d:", row-1)) {
			t.Errorf("over = true: cell %d is %q, want %q below the cell before", k, c, want)
		}
		if row%2 == 1 {
			empty = append(empty, c)
		}
	}
	if got := layer("false", time.Second); !slices.Equal(got, empty) {
		t.Errorf("over = false: got %q, want %q", got, empty)
	}
	// The last flash ends at 1 + 0.4 + 0.5 s at the latest.
	if got := layer("true", 1950*time.Millisecond); got != nil {
		t.Errorf("after time: got %q", got)
	}
}

// An event flashes 1–3 times: each flash before the last shows for 0.1 s, then nothing for 0.1 s;
// the last flash shows for time. Each later flash has the cells of the flash before it and new cells
// of its branches. The main part goes from the top row to the bottom row. The layer changes next
// at each start and each end of a flash (live-screen, sections 5 and 6.9).
func TestLightningFlashes(t *testing.T) {
	ms := time.Millisecond
	for _, shape := range []string{"slant", "square"} {
		s, p := page(t, map[string]string{"s.bbt": "+++\n[[effect]]\ndo = \"lightning\"\nshape = \"" + shape +
			"\"\nover = true\nevery = 1\nspread = 0\ntime = 0.5\n+++\n"}, 40, 24)
		seen, grow, later := map[int]bool{}, 0, 0
		for seed := uint64(1); seed <= 30; seed++ {
			l := effect.New(s.Effects, p, theme, rand.New(rand.NewPCG(seed, 2)))
			grid := p.Grid()
			// The cells of a flash by their row and column, and its rows.
			flash := func(at time.Duration) (map[[2]int]bool, map[int]bool) {
				cells, rows := map[[2]int]bool{}, map[int]bool{}
				for _, c := range l.Cells(at, grid) {
					cells[[2]int{c.Row, c.Col}] = true
					rows[c.Row] = true
				}
				return cells, rows
			}
			last, rows := flash(time.Second)
			if len(rows) != 24 {
				t.Fatalf("%s seed %d: the first flash has %d rows, want 24", shape, seed, len(rows))
			}
			// The times of the changes up to the next event at 2 s, and the state after each change.
			got := []string{"1s on"}
			for {
				next, ok := l.Next()
				if !ok || next >= 2*time.Second {
					break
				}
				state := "off"
				if cells, _ := flash(next); len(cells) > 0 {
					state = "on"
					for c := range last {
						if !cells[c] {
							t.Fatalf("%s seed %d: at %v the cell %v of the flash before is not drawn", shape, seed, next, c)
						}
					}
					if len(cells) > len(last) {
						grow++
					}
					last = cells
				}
				got = append(got, fmt.Sprintf("%v %s", next, state))
			}
			k := len(got) / 2 // each flash is on, then off
			var want []string
			for j := range k {
				on := time.Second + time.Duration(j)*200*ms
				want = append(want, fmt.Sprintf("%v on", on))
				if j < k-1 {
					want = append(want, fmt.Sprintf("%v off", on+100*ms))
				} else {
					want = append(want, fmt.Sprintf("%v off", on+500*ms))
				}
			}
			if !slices.Equal(got, want) {
				t.Errorf("%s seed %d: got %q, want %q", shape, seed, got, want)
			}
			seen[k] = true
			later += k - 1
		}
		if len(seen) != 3 || !seen[1] || !seen[2] || !seen[3] {
			t.Errorf("%s flash counts: got %v, want 1, 2 and 3", shape, seen)
		}
		if grow != later || later == 0 {
			t.Errorf("%s: %d of %d later flashes have new cells, want all", shape, grow, later)
		}
	}
}

// The connections of each grapheme of a square bolt: up, down, left, right. │ and ─ can also have
// one connection only (live-screen, section 6.9).
var boxConn = map[string][4]bool{
	"│": {true, true, false, false}, "─": {false, false, true, true},
	"┌": {false, true, false, true}, "┐": {false, true, true, false}, "└": {true, false, false, true}, "┘": {true, false, true, false},
	"├": {true, true, false, true}, "┤": {true, true, true, false}, "┬": {false, true, true, true}, "┴": {true, false, true, true},
	"┼": {true, true, true, true},
}

// A square bolt has the graphemes of its connections: each connection of a corner or a junction goes
// to a cell of the bolt with the opposite connection. All cells are in the rectangle. The areas of
// the events differ in size. shape = "both" gives both shapes (live-screen, section 6.9).
func TestLightningSquare(t *testing.T) {
	load := func(shape string) (*screen.Screen, screen.Page) {
		return page(t, map[string]string{"s.bbt": "+++\n[[effect]]\ndo = \"lightning\"\nshape = \"" + shape +
			"\"\nat = [3, 2, 30, 16]\nover = true\nevery = 1\nspread = 0\n+++\n"}, 40, 24)
	}
	s, p := load("square")
	sizes := map[int]bool{}
	for seed := uint64(1); seed <= 20; seed++ {
		l := effect.New(s.Effects, p, theme, rand.New(rand.NewPCG(seed, 2)))
		got := map[[2]int]string{}
		for _, c := range l.Cells(time.Second, p.Grid()) {
			if c.Row < 1 || c.Row >= 17 || c.Col < 2 || c.Col >= 32 {
				t.Errorf("seed %d: cell %d:%d is outside the rectangle", seed, c.Row, c.Col)
			}
			got[[2]int{c.Row, c.Col}] = c.Cell.Grapheme
		}
		sizes[len(got)] = true
		other := func(g string, k int) bool { // g has the connection k, or is a line that can end there
			cn, ok := boxConn[g]
			return ok && (cn[k] || g == "│" && k < 2 || g == "─" && k >= 2)
		}
		top, sky := 99, 0 // the first row of the bolt, and the cells with an up connection to no cell
		for at, g := range got {
			cn, ok := boxConn[g]
			if !ok {
				t.Fatalf("seed %d: cell %v has %q", seed, at, g)
			}
			top = min(top, at[0])
			if g == "│" || g == "─" {
				continue
			}
			for k, d := range [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
				nb, in := got[[2]int{at[0] + d[0], at[1] + d[1]}]
				switch {
				case !cn[k] || in && other(nb, k^1):
				case k == 0 && !in: // the first cell of the main part connects up
					sky++
				default:
					t.Errorf("seed %d: cell %v %q has no neighbour for connection %d", seed, at, g, k)
				}
			}
		}
		for at, g := range got {
			if at[0] > top && boxConn[g][0] && g != "│" && got[[2]int{at[0] - 1, at[1]}] == "" {
				t.Errorf("seed %d: cell %v %q below the first row connects up to no cell", seed, at, g)
			}
		}
		if sky > 1 {
			t.Errorf("seed %d: %d cells connect up to no cell", seed, sky)
		}
	}
	if len(sizes) < 5 {
		t.Errorf("bolt sizes: got %v, want 5 different sizes or more", sizes)
	}
	s, p = load("both")
	shapes := map[bool]bool{}
	for seed := uint64(1); seed <= 20; seed++ {
		l := effect.New(s.Effects, p, theme, rand.New(rand.NewPCG(seed, 2)))
		square := false
		for _, c := range l.Cells(time.Second, p.Grid()) {
			_, box := boxConn[c.Cell.Grapheme]
			square = square || box && c.Cell.Grapheme != "│"
		}
		shapes[square] = true
	}
	if len(shapes) != 2 {
		t.Errorf("shape both: got only square = %v", shapes)
	}
}

// sorted gives the cells of a layer in the short form of cells, sorted.
func sorted(s string) string {
	if s == "" {
		return s
	}
	parts := strings.Split(s, "; ")
	slices.Sort(parts)
	return strings.Join(parts, "; ")
}

// The rectangle of a place with in of a fragment is cut to the fragment rows (live-screen, section 6.2):
// a lightning in a rectangle below the fragment rows has the bolt of the rectangle of the fragment rows.
func TestLightningFragmentCut(t *testing.T) {
	layer := func(at string) string {
		front := "+++\n[[effect]]\ndo = \"lightning\"\nin = \"_f\"\nover = true\nat = " + at + "\nevery = 1\nspread = 0\n+++\n"
		s, p := page(t, map[string]string{"s.bbt": front + "top\n{>_f}\n" + strings.Repeat("below\n", 8), "_f.bbt": "abc\nabc\nabc"}, 40, 24)
		return cells(s, p, time.Second)
	}
	want := layer("[1, 1, 3, 3]")
	if want == "" {
		t.Fatal("no bolt")
	}
	if got := layer("[1, 1, 3, 9]"); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
