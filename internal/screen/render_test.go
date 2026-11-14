package screen_test

import (
	"encoding/binary"
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"testing/fstest"

	"newbbs/internal/bbt"
	"newbbs/internal/screen"
)

var update = flag.Bool("update", false, "rewrite the golden dumps in testdata/render")

func style(text string) screen.Style {
	spec, err := screen.ParseStyle(text)
	if err != nil {
		panic(err)
	}
	return spec.Over(screen.Style{})
}

// theme is the theme of section 7.3 with the default more style.
var theme = map[string]screen.Style{
	"normal": style("7/0"), "more": style("15/0"), "title": style("15/4 bold"), "prompt": style("14/0"),
}

// sauceArt gives an .ans file with a character SAUCE of the width.
func sauceArt(data string, width int) string {
	r := make([]byte, 128)
	copy(r, "SAUCE00")
	r[94], r[95] = 1, 1
	binary.LittleEndian.PutUint16(r[96:], uint16(width))
	return data + "\x1a" + string(r)
}

// dump loads the files of screens/, renders the screen name and gives its cell dump.
func dump(t *testing.T, files map[string]string, name string, cols, rows int, items ...string) string {
	t.Helper()
	fsys := fstest.MapFS{}
	for n, data := range files {
		fsys["screens/"+n] = &fstest.MapFile{Data: []byte(data)}
	}
	screens, ds := bbt.Load(fsys, theme, screen.WidthLimits{Min: 40, Max: 80})
	if ds.HasErrors() {
		t.Fatalf("diags: %v", ds)
	}
	env := &screen.Env{
		Term:  screen.Term{Type: "ansi", Cols: cols, Rows: rows, Charset: "utf-8", Colors: 16},
		Items: map[string]bool{},
	}
	for _, k := range items {
		env.Items[k] = true
	}
	var b strings.Builder
	if err := screen.WriteDump(&b, rows, screen.Render(screens[name], env, theme, "-- More --")); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

// golden compares got with testdata/render/<name>.dump. -update rewrites the file.
func golden(t *testing.T, name, got string) {
	t.Helper()
	file := filepath.Join("testdata", "render", name+".dump")
	if *update {
		if err := os.WriteFile(file, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Errorf("%s: got\n%s\nwant\n%s", file, got, want)
	}
}

func TestRender(t *testing.T) {
	logo := sauceArt("\x1b[1;33m{w}\x1b[0m a \x1b[44m{  w}\x1b[0m\r\nline 2", 12)
	w10 := "+++\nwidth = [10, 10]\n+++\n"
	recipe := map[string]string{
		"main.bbt":   "+++\nwidth = [20, 30]\n+++\n{#if term.cols < 30}\n{>_small}\n{#else}\n{>main.ans}\n{/if}\n",
		"_small.bbt": "{!14}small screen text that wraps\n",
		"main.ans":   sauceArt("\x1b[1;33mart\x1b[0m", 30),
	}
	tests := []struct {
		name       string
		files      map[string]string
		cols, rows int
		items      []string
	}{
		// The example of section 9.3.
		{"example", map[string]string{
			"main.bbt": "+++\n[[item]]\nkey = \"G\"\ndo = \"logoff\"\n+++\n{!title}Welcome!{!}\n{!14}[G]oodbye\n",
		}, 80, 24, nil},
		// The recipe of section 6.1 with art of width 30: the fragment below 30 columns,
		// wrapped at term.cols; the art at 30 columns and more, centred.
		{"recipe-small", recipe, 25, 10, nil},
		{"recipe-art", recipe, 40, 10, nil},
		// term.cols < min: W = min, each row cut at term.cols.
		// The cut through 中 leaves a space in its style.
		{"cut", map[string]string{
			"main.bbt": "+++\nwidth = [22, 30]\n+++\n{!title}abcdefghijklmnopqrs中x\nabcdefghijklmnopqrstuv\n",
		}, 20, 10, nil},
		// term.cols > max: a left margin of (cols - W) / 2, rounded down.
		{"margin", map[string]string{
			"main.bbt": "+++\nwidth = [8, 10]\n+++\n{!title}centred\n",
		}, 21, 10, nil},
		// More rows than term.rows: pages of term.rows - 1 rows and the More row.
		{"pages", map[string]string{
			"main.bbt": "+++\nwidth = [20, 20]\n+++\n" + strings.Repeat("row\n", 12),
		}, 20, 10, nil},
		// The rows fit in term.rows: one page without the More row.
		{"one-page", map[string]string{
			"main.bbt": "+++\nwidth = [20, 20]\n+++\n" + strings.Repeat("row\n", 10),
		}, 20, 10, nil},
		// Wrap at a space: the spaces at a break are removed, the leading spaces stay,
		// the spaces at the end of the line stay, the style continues on the next row.
		{"wrap-space", map[string]string{
			"main.bbt": w10 + "  {!14}one two   three four\nend   \n",
		}, 20, 10, nil},
		// The leading spaces join the first word and stay up to W.
		// A line of only spaces gives one row.
		{"spaces", map[string]string{
			"main.bbt": w10 + "      abcdefgh\n{!/4}" + strings.Repeat(" ", 25) + "\n" + strings.Repeat(" ", 12) + "x\n",
		}, 20, 10, nil},
		// A word wider than W starts on a new row and breaks at a grapheme boundary.
		{"long-word", map[string]string{
			"main.bbt": w10 + "ab abcdefghijklmnopqrstu\n",
		}, 20, 10, nil},
		// A wide grapheme that does not fit at the end of a row goes to the next row.
		{"wide-end", map[string]string{
			"main.bbt": w10 + "abcdefghi中x\nab 中中中中中\n",
		}, 20, 10, nil},
		// When W is 1, a wide grapheme becomes a space.
		{"w1", map[string]string{
			"main.bbt": "+++\nwidth = [1, 1]\n+++\na中\n",
		}, 20, 10, nil},
		// A fixed-width field wider than W is one word of W cells.
		{"fixed-wide", map[string]string{
			"main.bbt": w10 + "{term.type      }\n{      term.type}\nab {term.type }|\n",
		}, 20, 10, nil},
		// Block lines with only spaces give no row; a false block gives nothing.
		{"blocks", map[string]string{
			"main.bbt": "+++\nwidth = [10, 10]\n[[item]]\nkey = \"A\"\ndo = \"logoff\"\n[[item]]\nkey = \"B\"\ndo = \"logoff\"\n+++\n" +
				"{#if item.A}\nyes {term.cols}\n{#else}\nno\n{/if}\n" +
				"a{#if not item.A}b\nc{/if}d\n  {#if item.B}x{/if}  \n\n{!title}\n",
		}, 20, 10, []string{"A"}},
		// An art field takes the value; the art and an include do not change the style
		// of the text after them; the art is cut at W.
		{"art", map[string]string{
			"main.bbt":  "+++\nwidth = [10, 10]\n[fields]\nw = \"term.cols\"\n+++\n{!title}A\n{>logo.ans}\nB\n{>_frag}\nC\n",
			"logo.ans":  logo,
			"_frag.bbt": "{!14}frag\n",
		}, 20, 10, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			golden(t, tt.name, dump(t, tt.files, "main", tt.cols, tt.rows, tt.items...))
		})
	}
}

// The golden dump of the example equals the dump of section 9.3.
func TestRenderExampleDump(t *testing.T) {
	want := `page 1 80x24
|Welcome!                                                                        |
|[G]oodbye                                                                       |
styles
1: 1-8 15/4 bold, 9-80 7/0
2: 1-9 14/0, 10-80 7/0
`
	got, err := os.ReadFile(filepath.Join("testdata", "render", "example.dump"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

func load(t *testing.T, main string) *screen.Screen {
	t.Helper()
	fsys := fstest.MapFS{"screens/main.bbt": {Data: []byte(main)}}
	screens, ds := bbt.Load(fsys, theme, screen.WidthLimits{Min: 40, Max: 80})
	if ds.HasErrors() {
		t.Fatalf("diags: %v", ds)
	}
	return screens["main"]
}

// The render stops at MaxRenderRows rows. At W = 1, each grapheme of a word is one row.
func TestRenderRowLimit(t *testing.T) {
	s := load(t, "+++\nwidth = [1, 1]\n+++\n"+strings.Repeat("x", screen.MaxRenderRows+10)+"\n")
	env := &screen.Env{Term: screen.Term{Cols: 20, Rows: 100}}
	if rows := countRows(s, env); rows != screen.MaxRenderRows {
		t.Errorf("got %d rows, want %d", rows, screen.MaxRenderRows)
	}
}

// countRows renders s and gives the count of its rows without the More rows.
func countRows(s *screen.Screen, env *screen.Env) int {
	rows, pages := 0, 0
	for p := range screen.Render(s, env, theme, "") {
		rows += len(p.Rows)
		pages++
	}
	return rows - (pages - 1)
}

// A session holds one page (section 6.4): the render of MaxRenderRows rows
// at 255 columns allocates less than 3 pages of cells. A row has the cells
// of the screen area only: at W = 20, the render allocates less than 3 pages of 20 cells.
func TestRenderPageMemory(t *testing.T) {
	cell := reflect.TypeFor[screen.Cell]().Size()
	for _, tt := range []struct {
		width string
		w     int
	}{{"[255, 255]", 255}, {"[20, 20]", 20}} {
		s := load(t, "+++\nwidth = "+tt.width+"\n+++\n"+strings.Repeat("x\n", screen.MaxRenderRows))
		env := &screen.Env{Term: screen.Term{Cols: 255, Rows: 25}}
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		rows := countRows(s, env)
		runtime.ReadMemStats(&after)
		if rows != screen.MaxRenderRows {
			t.Errorf("width %s: got %d rows, want %d", tt.width, rows, screen.MaxRenderRows)
		}
		limit := 3 * uint64(env.Term.Rows*tt.w) * uint64(cell)
		if n := after.TotalAlloc - before.TotalAlloc; n > limit {
			t.Errorf("width %s: render allocated %d bytes, want at most %d", tt.width, n, limit)
		}
	}
}

// A long source line that gives no row does not grow the memory of the render:
// the buffers of a line have at most W cells.
func TestRenderLineMemory(t *testing.T) {
	s := load(t, "+++\n[fields]\nt = \"term.type\"\n+++\nx"+strings.Repeat("{t}", 300000)+"\n")
	env := &screen.Env{Term: screen.Term{Type: strings.Repeat(" ", 40), Cols: 80, Rows: 25}}
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	countRows(s, env)
	runtime.ReadMemStats(&after)
	if n := after.TotalAlloc - before.TotalAlloc; n > 1<<20 {
		t.Errorf("render allocated %d bytes, want at most 1 MB", n)
	}
}
