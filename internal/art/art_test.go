package art_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"newbbs/internal/art"
	"newbbs/internal/screen"
)

// withSauce adds a SAUCE record with the width, the height and the TFlags to data.
func withSauce(data string, width, height int, flags byte) []byte {
	r := make([]byte, 128)
	copy(r, "SAUCE00")
	r[94], r[95] = 1, 1
	binary.LittleEndian.PutUint16(r[96:], uint16(width))
	binary.LittleEndian.PutUint16(r[98:], uint16(height))
	r[105] = flags
	return append([]byte(data+"\x1a"), r...)
}

// dump writes the grid of a in the cell dump form of section 9.3.
func dump(t *testing.T, a *screen.Art) string {
	t.Helper()
	var b strings.Builder
	if err := screen.WritePage(&b, a.Height, 1, screen.Page{Rows: a.Rows, Cols: a.Width}); err != nil {
		t.Fatal(err)
	}
	return strings.TrimPrefix(b.String(), fmt.Sprintf("page 1 %dx%d\n", a.Width, a.Height))
}

func diagText(ds screen.Diags) string {
	var s []string
	for _, d := range ds {
		s = append(s, d.String())
	}
	return strings.Join(s, "\n")
}

func parse(t *testing.T, data []byte) *screen.Art {
	t.Helper()
	a, ds := art.Parse("a.ans", data)
	if ds.HasErrors() || a == nil {
		t.Fatalf("errors:\n%s", diagText(ds))
	}
	return a
}

func cell(a *screen.Art, row, col int) screen.Cell { return a.Rows[row-1][col-1] }

func TestMoebius(t *testing.T) {
	data, err := os.ReadFile("testdata/moebius.ans")
	if err != nil {
		t.Fatal(err)
	}
	a, ds := art.Parse("moebius.ans", data)
	if len(ds) > 0 || a == nil {
		t.Fatalf("diagnostics:\n%s", diagText(ds))
	}
	if a.Width != 132 || a.Height != 3 {
		t.Errorf("size %dx%d, want 132x3", a.Width, a.Height)
	}
	if c := cell(a, 1, 1); c.Grapheme != "H" || c.Style.String() != "14/4" {
		t.Errorf("1:1 = %q %s", c.Grapheme, c.Style)
	}
	if c := cell(a, 3, 3); c.Grapheme != "日" || c.Style.String() != "196/0" || cell(a, 3, 4).Grapheme != "" {
		t.Errorf("3:3 = %q %s", c.Grapheme, c.Style)
	}
	want := []screen.ArtField{{Row: 2, Col: 128, Width: 4, Path: "w", Right: true, Style: cell(a, 2, 1).Style}}
	if fmt.Sprint(a.Fields) != fmt.Sprint(want) {
		t.Errorf("fields %v, want %v", a.Fields, want)
	}
}

func TestIcyDraw(t *testing.T) {
	a, ds := art.Load(os.DirFS("testdata"), "icydraw.ans")
	if a == nil {
		t.Fatalf("errors:\n%s", diagText(ds))
	}
	if got, want := diagText(ds), `icydraw.ans:1:1: warning: unknown sequence "ESC[0;1 D"`; got != want {
		t.Errorf("diagnostics:\n%s\nwant:\n%s", got, want)
	}
	if a.Width != 40 || a.Height != 4 {
		t.Errorf("size %dx%d, want 40x4 (SAUCE TInfo1, TInfo2)", a.Width, a.Height)
	}
	if c := cell(a, 1, 1); c.Style.String() != "#ff8800/#000020" {
		t.Errorf("1:1 style %s", c.Style)
	}
	// iCE colours: blink makes the background bright.
	if c := cell(a, 2, 1); c.Grapheme != "X" || c.Style.String() != "7/12" {
		t.Errorf("2:1 = %q %s", c.Grapheme, c.Style)
	}
	if len(a.Fields) != 1 || a.Fields[0].Path != "term.cols" || a.Fields[0].Right || a.Fields[0].Width != 12 {
		t.Errorf("fields %v", a.Fields)
	}
}

func TestPlay(t *testing.T) {
	tests := []struct {
		name, data string
		width      int
		flags      byte
		want       string
	}{
		{"pending wrap and LF", "abcd\nef", 4, 0, "|abcd|\n|ef  |\n"},
		{"pending wrap and grapheme", "abcde", 4, 0, "|abcd|\n|e   |\n"},
		{"pending wrap and CR", "abcd\re", 4, 0, "|abcd|\n|e   |\n"},
		{"pending wrap and cursor", "abcd\x1b[Ce", 4, 0, "|abcd|\n| e  |\n"},
		{"SGR keeps the wrap pending", "abcd\x1b[m\x1b[De", 4, 0, "|abcd|\n|e   |\n"},
		{"wrap off", "\x1b[?7labcdef", 4, 0, "|abcf|\n"},
		{"wide grapheme at the edge", "abc日", 4, 0, "|abc |\n|日  |\n"},
		{"wide grapheme at the edge, wrap off", "\x1b[?7labc日", 4, 0, "|ab日|\n"},
		{"repeat", "a\x1b[5b", 4, 0, "|aaaa|\n|aa  |\n"},
		{"repeat, wrap off", "\x1b[?7la\x1b[9999bc", 4, 0, "|aaac|\n"},
		{"cut wide grapheme", "日\rx", 4, 0, "|x   |\n"},
		{"ESC and another byte", "\x1bXa", 4, 0, "|Xa  |\n"},
		{"cursor position and save", "\x1b[2;3Ha\x1b[s\x1b[Hb\x1b[uc", 4, 0, "|b   |\n|  ac|\n"},
		{"cursor stays inside the grid", "\x1b[99Ca\x1b[9A\x1b[99Db", 4, 0, "|b  a|\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dump(t, parse(t, withSauce(tt.data, tt.width, 0, tt.flags)))
			got = got[:strings.Index(got, "styles\n")]
			if got != tt.want {
				t.Errorf("got\n%swant\n%s", got, tt.want)
			}
		})
	}
}

func TestStyles(t *testing.T) {
	tests := []struct {
		name, data string
		flags      byte
		want       string
	}{
		{"start style", "a", 0, "7/0"},
		{"bold adds 8", "\x1b[1;31mA", 0, "9/0"},
		{"bold and RGB", "\x1b[1;38;2;1;2;3mA", 0, "#010203/0"},
		{"blink without iCE", "\x1b[5;41mA", 0, "7/1 blink"},
		{"blink with iCE from SAUCE", "\x1b[5;41mA", 1, "7/9"},
		{"blink with iCE from CSI", "\x1b[?33h\x1b[5;41mA", 0, "7/9"},
		{"iCE off", "\x1b[?33l\x1b[5;41mA", 1, "7/1 blink"},
		{"order: bold, blink, reverse", "\x1b[1;5;31;44;7mA", 1, "12/9"},
		{"underline", "\x1b[4mA", 0, "7/0 underline"},
		{"default colours", "\x1b[31;41;39;49mA", 0, "7/0"},
		{"bright", "\x1b[91;101mA", 0, "9/9"},
		{"index", "\x1b[38;5;200;48;5;17mA", 0, "200/17"},
		{"reset", "\x1b[1;5;4;7;31;41;0mA", 0, "7/0"},
		{"off", "\x1b[1;5;4;7;31m\x1b[22;25;24;27mA", 0, "1/0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := parse(t, withSauce(tt.data, 4, 0, tt.flags))
			if got := cell(a, 1, 1).Style.String(); got != tt.want {
				t.Errorf("style %s, want %s", got, tt.want)
			}
		})
	}
}

func TestErase(t *testing.T) {
	a := parse(t, withSauce("\x1b[44mabcd\x1b[1;2H\x1b[K", 4, 0, 0))
	if got, want := dump(t, a), "|a   |\nstyles\n1: 1-4 7/4\n"; got != want {
		t.Errorf("K 0:\n%s", got)
	}
	a = parse(t, withSauce("\x1b[31;1;44mabc\x1b[1K", 4, 0, 0))
	if got, want := dump(t, a), "|    |\nstyles\n1: 1-4 7/4\n"; got != want {
		t.Errorf("K 1:\n%s", got)
	}
	// CSI 2 J erases the rows up to the height, then the play continues at 1:1.
	a = parse(t, withSauce("ab\nc\x1b[41m\x1b[2Jx", 4, 0, 0))
	if got, want := dump(t, a), "|x   |\n|    |\nstyles\n1: 1-4 7/1\n2: 1-4 7/1\n"; got != want {
		t.Errorf("J:\n%s", got)
	}
}

func TestMoebiusForm(t *testing.T) {
	a := parse(t, []byte("\n\nab\ncd"))
	if a.Width != 80 || a.Height != 2 || cell(a, 1, 1).Grapheme != "a" {
		t.Errorf("size %dx%d, 1:1 %q", a.Width, a.Height, cell(a, 1, 1).Grapheme)
	}
	_, ds := art.Parse("a.ans", []byte(strings.Repeat("x", 256)))
	if got, want := diagText(ds), "a.ans:1:255: row is over 255 columns"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFields(t *testing.T) {
	tests := []struct {
		name, data string
		fields     string
		diags      string
	}{
		{"left", "{w  }", "[{1 1 5 w false 7/0}]", ""},
		{"right", "{  w}", "[{1 1 5 w true 7/0}]", ""},
		{"spaces on both sides", "{ term.cols }", "[{1 1 13 term.cols true 7/0}]", ""},
		{"plain", "{ BBS } {a {} {!x} {{w}", "[{1 21 3 w false 7/0}]", ""},
		{"style of the first cell", "\x1b[33;44m{w}", "[{1 1 3 w false 3/4}]", ""},
		{"space with another foreground", "\x1b[31m{\x1b[32m \x1b[31mw}", "[{1 1 4 w true 1/0}]", ""},
		{"two foregrounds", "\x1b[3;15H\x1b[93m{\x1b[97mw}", "[]",
			`a.ans:3:15: token "{w}" changes the foreground: 11 at col 15, 15 at col 16`},
		{"two backgrounds", "{\x1b[41m w}", "[]",
			`a.ans:1:1: token "{ w}" changes the background: 0 at col 1, 1 at col 2`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, ds := art.Parse("a.ans", withSauce(tt.data, 30, 0, 0))
			if got := diagText(ds); got != tt.diags {
				t.Errorf("diagnostics %q, want %q", got, tt.diags)
			}
			var fields []string
			if a != nil {
				for _, f := range a.Fields {
					fields = append(fields, fmt.Sprintf("{%d %d %d %s %t %s}", f.Row, f.Col, f.Width, f.Path, f.Right, f.Style))
				}
			}
			if got := "[" + strings.Join(fields, " ") + "]"; got != tt.fields {
				t.Errorf("fields %s, want %s", got, tt.fields)
			}
		})
	}
}

func TestErrors(t *testing.T) {
	tests := []struct {
		name  string
		data  []byte
		diags string
	}{
		{"file over 1 MB", make([]byte, screen.MaxFileBytes+1), "a.ans:1:1: file is over 1048576 bytes"},
		{"row 1001", withSauce("\x1b[1000Ha\nb", 4, 0, 0), "a.ans:1001:1: art is over 1000 rows"},
		{"erase at row 1001", withSauce("\x1b[1000H\n\x1b[K", 4, 0, 0), "a.ans:1001:1: art is over 1000 rows"},
		{"CP437", withSauce("ab\xb0\xb1", 4, 0, 0),
			"a.ans:1:3: invalid UTF-8: art must be UTF-8, CP437 art needs a conversion"},
		{"SAUCE width", withSauce("", 300, 0, 0), "a.ans:1:1: SAUCE width 300 is over 255"},
		{"SAUCE height", withSauce("", 80, 1001, 0), "a.ans:1:1: SAUCE height 1001 is over 1000 rows"},
		{"control character", withSauce("a\tb", 4, 0, 0), "a.ans:1:2: control character U+0009"},
		{"C1 control", withSauce("a\u0085b", 4, 0, 0), "a.ans:1:2: control character U+0085"},
		{"zero width", withSauce("\u200bb", 4, 0, 0), "a.ans:1:1: zero-width grapheme U+200B"},
		{"wide in width 1", withSauce("日", 1, 0, 0), "a.ans:1:1: wide grapheme in art of width 1"},
		{"warnings", withSauce("\x1b[5n\x1b[99m\x1b[38;5m\x1b[?1h\x1b[1J\x1b[3K\x1b[1:2m\x1b[", 4, 0, 0),
			`a.ans:1:1: warning: unknown sequence "ESC[5n"` + "\n" +
				`a.ans:1:1: warning: unknown SGR parameter 99 in "ESC[99m"` + "\n" +
				`a.ans:1:1: warning: invalid colour in "ESC[38;5m"` + "\n" +
				`a.ans:1:1: warning: unknown mode 1 in "ESC[?1h"` + "\n" +
				`a.ans:1:1: warning: unknown sequence "ESC[1J"` + "\n" +
				`a.ans:1:1: warning: unknown sequence "ESC[3K"` + "\n" +
				`a.ans:1:1: warning: unknown sequence "ESC[1:2m"` + "\n" +
				`a.ans:1:1: warning: incomplete sequence "ESC["`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ds := art.Parse("a.ans", tt.data)
			if got := diagText(ds); got != tt.diags {
				t.Errorf("got\n%s\nwant\n%s", got, tt.diags)
			}
		})
	}
}

func TestLimits(t *testing.T) {
	// CSI n b stops at the row limit without an error.
	a := parse(t, withSauce("\x1b[999Ha\x1b[9999b", 4, 0, 0))
	if a.Height != 1000 {
		t.Errorf("height %d, want 1000", a.Height)
	}
	a = parse(t, withSauce("a\x1b[9999b", 80, 0, 0))
	if a.Height != 125 || cell(a, 125, 80).Grapheme != "a" {
		t.Errorf("height %d, want 125", a.Height)
	}
	// A long sequence is cut in the message.
	_, ds := art.Parse("a.ans", withSauce("\x1b["+strings.Repeat("1;", 1000)+"z", 4, 0, 0))
	if got := diagText(ds); len(got) > 80 || !strings.Contains(got, `..."`) {
		t.Errorf("message %q", got)
	}
}

func TestTooManyDiagnostics(t *testing.T) {
	_, ds := art.Parse("a.ans", withSauce(strings.Repeat("\t", 200), 4, 0, 0))
	if len(ds) != screen.MaxFileDiags+1 || ds[len(ds)-1].Msg != "too many diagnostics" {
		t.Errorf("%d diagnostics, last %q", len(ds), ds[len(ds)-1].Msg)
	}
}

func TestLoad(t *testing.T) {
	fsys := fstest.MapFS{
		"big.ans": {Data: bytes.Repeat([]byte("a"), screen.MaxFileBytes+1)},
		"ok.ans":  {Data: []byte("ab")},
	}
	if _, ds := art.Load(fsys, "big.ans"); diagText(ds) != "big.ans:1:1: file is over 1048576 bytes" {
		t.Errorf("big: %s", diagText(ds))
	}
	if _, ds := art.Load(fsys, "none.ans"); !ds.HasErrors() {
		t.Error("missing file: no error")
	}
	if a, ds := art.Load(fsys, "ok.ans"); a == nil || len(ds) > 0 {
		t.Errorf("ok: %s", diagText(ds))
	}
}

func FuzzParse(f *testing.F) {
	for _, name := range []string{"testdata/moebius.ans", "testdata/icydraw.ans"} {
		data, err := os.ReadFile(name)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(data)
	}
	f.Add(withSauce("日\x1b[?7l\x1b[5b\x1b[2J{w}\x1b[1;5;7m\x1b[K", 3, 2, 1))
	f.Fuzz(func(t *testing.T, data []byte) {
		a, ds := art.Parse("f.ans", data)
		if a == nil {
			if !ds.HasErrors() {
				t.Fatal("no art and no error")
			}
			return
		}
		if a.Height > art.MaxRows || a.Width > screen.MaxWidth || len(a.Rows) != a.Height {
			t.Fatalf("size %dx%d, %d rows", a.Width, a.Height, len(a.Rows))
		}
		for i, r := range a.Rows {
			if len(r) != a.Width {
				t.Fatalf("row %d has %d cells", i+1, len(r))
			}
		}
		for _, fl := range a.Fields {
			if fl.Row > a.Height || fl.Col+fl.Width-1 > a.Width {
				t.Fatalf("field %v outside the grid", fl)
			}
		}
	})
}

// TestRepeat checks that CSI n b gives the grid of n more copies of the last grapheme.
func TestRepeat(t *testing.T) {
	tests := []struct{ rep, plain string }{
		{"a\x1b[9b", strings.Repeat("a", 10)},
		{"日\x1b[6b", strings.Repeat("日", 7)},
		{"xx日\x1b[3b", "xx" + strings.Repeat("日", 4)},
		{"\x1b[?7lab\x1b[9999b", "\x1b[?7lab" + strings.Repeat("b", 10)},
		{"\x1b[?7lx日\x1b[9999b", "\x1b[?7lx" + strings.Repeat("日", 10)},
		{"日日日\rx\x1b[2b", "日日日\rxxx"},
		{"\x1b[1;31ma\x1b[4b", "\x1b[1;31maaaaa"},
	}
	for _, tt := range tests {
		for _, width := range []int{1, 2, 3, 5} {
			got, gds := art.Parse("a.ans", withSauce(tt.rep, width, 0, 0))
			want, wds := art.Parse("a.ans", withSauce(tt.plain, width, 0, 0))
			if gds.HasErrors() != wds.HasErrors() || got != nil && dump(t, got) != dump(t, want) {
				t.Errorf("%q, width %d: differs from %q", tt.rep, width, tt.plain)
			}
		}
	}
}

func TestEraseTwice(t *testing.T) {
	// The second clear refills the row in its own style.
	a := parse(t, withSauce("ab\x1b[41m\x1b[2J\x1b[42m\x1b[2Jx", 4, 0, 0))
	if got, want := dump(t, a), "|x   |\nstyles\n1: 1-4 7/2\n"; got != want {
		t.Errorf("two clears:\n%s", got)
	}
}

func TestOtherSauce(t *testing.T) {
	// TInfo1, TInfo2 and TFlags count only for DataType 1 and FileType 1.
	data := withSauce("\x1b[5;41ma", 4, 5, 1)
	data[len(data)-128+94] = 0
	a := parse(t, data)
	if a.Width != 80 || a.Height != 1 || cell(a, 1, 1).Style.String() != "7/1 blink" {
		t.Errorf("size %dx%d, style %s", a.Width, a.Height, cell(a, 1, 1).Style)
	}
}

// filler gives a play of n erased cells at the play width 255: CSI 2 K for each 255 cells
// and one CSI K from the column 256-r for the rest r.
func filler(n int) string {
	s := strings.Repeat("\x1b[2K", n/screen.MaxWidth)
	if r := n % screen.MaxWidth; r > 0 {
		s += fmt.Sprintf("\x1b[1;%dH\x1b[K", screen.MaxWidth-r+1)
	}
	return s
}

// TestBudget checks each count rule of the art play budget (section 10):
// the play of the budget gives no diagnostic, and one more cell gives the error at the cursor.
func TestBudget(t *testing.T) {
	const w = screen.MaxWidth
	tests := []struct {
		name       string
		sauce      bool // else the Moebius form: the play width is 255, the art width 80
		prefix, op string
		cost       int    // the cells of prefix and op
		at         string // the position of the error
	}{
		{"grapheme", true, "", "\x1b[5;2Ha", 1, "5:2"},
		{"wide grapheme", true, "", "\x1b[5;2H日", 2, "5:2"},
		{"erase in the Moebius form", false, "", "\x1b[5;2H\x1b[K", w - 1, "5:2"},
		{"clear", true, "\x1b[7H\x1b[2K", "\x1b[3;3H\x1b[2J", w + 7*w, "3:3"},
		{"repeat with wrap", true, "", "\x1b[5;250Ha\x1b[9b", 10, "6:4"},
		{"repeat without wrap", true, "", "\x1b[?7l\x1b[5;250Ha\x1b[9b", 10, "5:255"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			load := func(extra int) screen.Diags {
				data := []byte(tt.prefix + filler(art.MaxPlayCells-tt.cost+extra) + tt.op)
				if tt.sauce {
					data = withSauce(string(data), w, 0, 0)
				}
				_, ds := art.Parse("a.ans", data)
				return ds
			}
			if ds := load(0); len(ds) > 0 {
				t.Errorf("at the budget:\n%s", diagText(ds))
			}
			want := "a.ans:" + tt.at + ": art play is over 4080000 written and erased cells"
			if got := diagText(load(1)); got != want {
				t.Errorf("over the budget: got %q, want %q", got, want)
			}
		})
	}
}

// TestBudgetFiles checks that crafted files of 1 MB stop at the budget in less than 1 s.
func TestBudgetFiles(t *testing.T) {
	tests := []struct {
		name, start, unit string
		width             int
	}{
		{"repeat with wrap", "A", "\x1b[H\x1b[9999b", 80},
		{"repeat without wrap", "\x1b[?7lA", "\x1b[H\x1b[9999b", 80},
		{"clear", "", "\x1b[1000;255HA\x1b[2J", 255},
		{"erase", "", "\x1b[2K\x1b[B", 255},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := tt.start + strings.Repeat(tt.unit, (screen.MaxFileBytes-200)/len(tt.unit))
			data := withSauce(body, tt.width, 0, 0)
			start := time.Now()
			_, ds := art.Parse("a.ans", data)
			if d := time.Since(start); d >= time.Second {
				t.Errorf("load took %v", d)
			}
			if !strings.HasSuffix(diagText(ds), ": art play is over 4080000 written and erased cells") {
				t.Errorf("diagnostics:\n%s", diagText(ds))
			}
		})
	}
}
