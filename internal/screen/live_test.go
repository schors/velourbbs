package screen_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/schors/velourbbs/internal/bbt"
	"github.com/schors/velourbbs/internal/screen"
)

// liveFields is the [fields] of the tests of the live fields.
var liveFields = screen.Fields{"clock": alias("clock", `bbs.time live "15:04:05"`)}

// liveEnv gives the values of a render at cols x rows with the default time of velourbbs preview.
func liveEnv(cols, rows int) *screen.Env {
	clock := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
	layout := "2006-01-02 15:04:05"
	return &screen.Env{
		Term: screen.Term{Type: "ansi", Cols: cols, Rows: rows, Colors: 16},
		Session: screen.SessionValues{
			Online: "00:00:00", Uptime: "00:00:00", Sessions: 1, Time: clock.Format(layout),
			Clock: clock, Layout: layout, MaxConns: 64,
		},
	}
}

// loadLive loads the files of screens/ with the width limits [w, w] and gives the screen name.
func loadLive(t *testing.T, files map[string]string, name string, w int) *screen.Screen {
	t.Helper()
	fsys := fstest.MapFS{}
	for n, data := range files {
		fsys["screens/"+n] = &fstest.MapFile{Data: []byte(data)}
	}
	sets, ds := bbt.Load(fsys, bbt.Settings{Theme: theme, Layout: screen.Layout{Width: screen.WidthLimits{Min: w, Max: w}}, Fields: liveFields})
	if ds.HasErrors() {
		t.Fatalf("diags: %v", ds)
	}
	return sets["en"][name]
}

// showSlots gives the slots of p in a short form: row:col width skip+cells path, and the text of each run.
func showSlots(p screen.Page) string {
	var out []string
	for _, s := range p.Slots {
		var b strings.Builder
		for _, c := range p.Rows[s.Row][s.Col : s.Col+s.Cells] {
			b.WriteString(c.Grapheme)
		}
		out = append(out, fmt.Sprintf("%d:%d %d %d+%d %s %q", s.Row, s.Col, s.Width, s.Skip, s.Cells, s.Ref, b.String()))
	}
	return strings.Join(out, "; ")
}

// The render marks the cells of each slot, and the page lists its slots (live-screen, section 3.1).
func TestSlots(t *testing.T) {
	tests := []struct {
		name, body string
		w, cols    int
		want       string
	}{
		{"text", "Time [{bbs.time live}] up {session.online live}|", 80, 80,
			`0:6 19 0+19 bbs.time "2006-01-02 15:04:05"; 0:30 14 0+14 session.online "00:00:00      "`},
		{"stretch", "{bbs.sessions live}{~.}{bbs.uptime live}", 40, 40,
			`0:0 2 0+2 bbs.sessions "1 "; 0:26 14 0+14 bbs.uptime "00:00:00      "`},
		{"layout and alias", `{bbs.time live "15:04"}x {clock}`, 40, 40,
			`0:0 5 0+5 bbs.time "15:04"; 0:7 8 0+8 bbs.time "15:04:05"`},
		{"fixed width", "{    bbs.sessions live}|{clock  }|", 40, 40,
			`0:0 23 0+23 bbs.sessions "                      1"; 0:24 9 0+9 bbs.time "15:04:05 "`},
		{"no slot", "{bbs.name live} {term.cols live} {bbs.time}", 40, 40, ``},
		// A slot is one word: it does not break at its spaces.
		{"wrap", "abc {bbs.time live}", 20, 40, `1:0 19 0+19 bbs.time "2006-01-02 15:04:05"`},
		// A word wider than W breaks the slot over rows.
		{"long word", "ab{bbs.time live}", 20, 40,
			`0:2 19 0+18 bbs.time "2006-01-02 15:04:0"; 1:0 19 18+1 bbs.time "5"`},
		// A slot wider than W has the width W. A cut at term.cols cuts the slot.
		{"wider than W", "{bbs.time live}", 10, 40, `0:0 10 0+10 bbs.time "2006-01-02"`},
		{"cut", strings.Repeat("x", 20) + "{bbs.time live}", 40, 30, `0:20 19 0+10 bbs.time "2006-01-02"`},
		// The cells that the cut at term.cols hides count for the run on the next row.
		{"cut and long word", strings.Repeat("x", 30) + "{bbs.time live}", 40, 20, `1:0 19 10+9 bbs.time " 15:04:05"`},
		{"block", "{#if term.cols > 1}{bbs.uptime live}{/if}x", 40, 40, `0:0 14 0+14 bbs.uptime "00:00:00      "`},
	}
	for _, tt := range tests {
		s := loadLive(t, map[string]string{"main.bbt": tt.body}, "main", tt.w)
		var got string
		for p := range screen.Render(s, liveEnv(tt.cols, 24), theme, "-- More --") {
			got = showSlots(p)
		}
		if got != tt.want {
			t.Errorf("%s:\n got %s\nwant %s", tt.name, got, tt.want)
		}
	}
}

// A slot in art has the width of its token and the style of the field (live-screen, section 3.1).
// A slot in an include has the rules of a slot in the body.
func TestArtSlots(t *testing.T) {
	art := sauceArt("\x1b[36mab{clock  }\x1b[33m{ bbs.sessions live}\n{bbs.time live}", 40)
	s := loadLive(t, map[string]string{
		"main.bbt": "{>_f}\n{>a.ans}\n", "_f.bbt": "{session.online live}\n", "a.ans": art,
	}, "main", 40)
	var p screen.Page
	for p = range screen.Render(s, liveEnv(40, 24), theme, "-- More --") {
	}
	want := `0:0 14 0+14 session.online "00:00:00      "; 1:2 9 0+9 bbs.time "15:04:05 "; 1:11 20 0+20 bbs.sessions "                   1"; 2:0 15 0+15 bbs.time "2006-01-02 15:0"`
	if got := showSlots(p); got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
	if c := p.Rows[1][11]; c.Style.String() != "3/0" || c.Slot == 0 {
		t.Errorf("slot cell %+v", c)
	}
}

// A list line gives a slot in each row of the list page (live-screen, section 3.1).
func TestListSlots(t *testing.T) {
	s := listScreen(t, "", "{#each}{row.num} {bbs.uptime live}{/each}\n", "a.txt", "b.txt")
	env := liveEnv(30, 10)
	env.List = &screen.ListValues{Page: 1, Pages: 1}
	p := screen.RenderList(s, env, theme, s.List.List(nil), 1, 3)
	want := `0:2 14 0+14 bbs.uptime "00:00:00      "; 1:2 14 0+14 bbs.uptime "00:00:00      "`
	if got := showSlots(p); got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}

// A bbs.time with a layout shows Session.Time when Clock has a zero instant: a value of --set.
func TestSlotSetValue(t *testing.T) {
	s := loadLive(t, map[string]string{"main.bbt": `{bbs.time live "15:04"}|{clock}|`}, "main", 40)
	env := liveEnv(40, 24)
	env.Session.Time, env.Session.Clock = "a long value", time.Time{}
	for p := range screen.Render(s, env, theme, "-- More --") {
		if got := texts(p)[0]; got != "a lon|a long v|" {
			t.Errorf("got %q", got)
		}
	}
}

// A frame row of a file screen gives a slot: the rows above and below the file area (live-screen, section 3.1).
func TestFileSlots(t *testing.T) {
	s := fileScreen(t, "{bbs.uptime live}\n{>file}\n{bbs.sessions live}\n")
	env := liveEnv(30, 10)
	env.File = &screen.FileValues{Content: s.FileScreen.Content(nil), Page: 1, Pages: 1}
	p := screen.RenderFile(s, env, theme, fileRows(2), 1)
	want := `0:0 14 0+14 bbs.uptime "00:00:00      "; 3:0 2 0+2 bbs.sessions "1 "`
	if got := showSlots(p); got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}

// A slot that a long word breaks over two pages keeps its offset on the second page.
func TestSlotPages(t *testing.T) {
	s := loadLive(t, map[string]string{"main.bbt": strings.Repeat("l\n", 8) + "ab{bbs.time live}\nz\n"}, "main", 20)
	var got []string
	for p := range screen.Render(s, liveEnv(20, 10), theme, "-- More --") {
		got = append(got, showSlots(p))
	}
	want := []string{`8:2 19 0+18 bbs.time "2006-01-02 15:04:0"`, `0:0 19 18+1 bbs.time "5"`}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

// A slot in art has the width of its token also when the token is wider than W: the cut at W cuts it.
func TestArtSlotCut(t *testing.T) {
	s := loadLive(t, map[string]string{"main.bbt": "{>a.ans}\n", "a.ans": sauceArt("{    bbs.sessions live}", 30)}, "main", 20)
	for p := range screen.Render(s, liveEnv(40, 24), theme, "-- More --") {
		want := `0:0 23 0+20 bbs.sessions "                    "`
		if got := showSlots(p); got != want {
			t.Errorf("got  %s\nwant %s", got, want)
		}
	}
}

// A tick writes the values into the slots: the page equals the render with these values
// (live-screen, section 4). The cases cover a slot in text, in art, a fixed width aligned right,
// a slot that a long word breaks over rows, a cut at term.cols and a value of --set.
func TestTick(t *testing.T) {
	art := sauceArt("ab{clock  }{ bbs.sessions live}", 40)
	tests := []struct {
		name, body string
		w, cols    int
	}{
		{"text", "Time [{bbs.time live}] up {session.online live} {bbs.uptime live}|", 80, 80},
		{"art", "{>a.ans}\n", 40, 40},
		{"long word", "ab{bbs.time live}", 20, 40},
		{"cut", strings.Repeat("x", 20) + "{bbs.time live}", 40, 30},
	}
	later := func(env *screen.Env) *screen.Env {
		e := *env
		clock := env.Session.Clock.Add(26*time.Hour + 61*time.Second)
		e.Session.Clock, e.Session.Time = clock, clock.Format(env.Session.Layout)
		e.Session.Online, e.Session.Uptime, e.Session.Sessions = "1d 02:01:01", "12d 00:00:00", 17
		return &e
	}
	for _, tt := range tests {
		s := loadLive(t, map[string]string{"main.bbt": tt.body, "a.ans": art}, "main", tt.w)
		env := liveEnv(tt.cols, 24)
		var p, want screen.Page
		for p = range screen.Render(s, env, theme, "-- More --") {
		}
		for want = range screen.Render(s, later(env), theme, "-- More --") {
		}
		got := p.Tick(later(env).Session)
		if !slices.EqualFunc(got.Rows, want.Rows, slices.Equal) {
			t.Errorf("%s:\n got %q\nwant %q", tt.name, texts(got), texts(want))
		}
		if strings.Join(texts(p), "|") == strings.Join(texts(got), "|") {
			t.Errorf("%s: the tick changed the rows of the page", tt.name)
		}
	}
}

// A cut through a wide grapheme at the edge of a run leaves a space.
func TestTickWide(t *testing.T) {
	ref, _, err := screen.LookupField("bbs.time", nil)
	if err != nil {
		t.Fatal(err)
	}
	p := screen.Page{
		Rows:  []screen.Row{{{Grapheme: "a"}, {Grapheme: "b"}, {Grapheme: "c"}, {Grapheme: "d"}}},
		Slots: []screen.Slot{{Width: 6, Skip: 1, Cells: 4, Ref: ref}},
	}
	got := p.Tick(screen.SessionValues{Time: "世界x"})
	var b strings.Builder
	for _, c := range got.Rows[0] {
		b.WriteString(c.Grapheme + "|")
	}
	if b.String() != " |界||x|" {
		t.Errorf("got %q, want \" |界||x|\"", b.String())
	}
	p.Slots[0].Cells = 1
	if got := texts(p.Tick(screen.SessionValues{Time: "a世界"})); got[0] != " bcd" {
		t.Errorf("got %q, want a space for the cut wide grapheme", got[0])
	}
}

// The grid has all terminal rows of term.cols cells. A fill row is spaces in default
// with fill none, else in normal (live-screen, section 2).
func TestGrid(t *testing.T) {
	normal := screen.Style{Fg: screen.Colour{Kind: screen.Index, Value: 7}}
	for _, fill := range []screen.Fill{screen.FillNone, screen.FillSpaces} {
		p := screen.Page{Rows: []screen.Row{{{Grapheme: "a", Style: normal}}}, Cols: 3, Height: 3, Top: 1, Margin: 1,
			Normal: normal, Fill: fill}
		var b strings.Builder
		if err := screen.WriteGrid(&b, p); err != nil {
			t.Fatal(err)
		}
		fillStyle := "-/-"
		if fill != screen.FillNone {
			fillStyle = "7/-"
		}
		want := "grid 3x3\n|   |\n| a |\n|   |\nstyles\n1: 1-3 " + fillStyle + "\n2: 1-3 7/-\n3: 1-3 " + fillStyle + "\n"
		if b.String() != want {
			t.Errorf("fill %s:\n%s\nwant\n%s", fill, b.String(), want)
		}
	}
}
