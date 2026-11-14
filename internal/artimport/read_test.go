package artimport

import (
	"encoding/binary"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"newbbs/internal/art"
	"newbbs/internal/screen"
)

const bom = "\xef\xbb\xbf"

// sauceRec gives a character SAUCE record with the width, the height and the TFlags.
func sauceRec(width, height int, flags byte) []byte {
	r := make([]byte, 128)
	copy(r, "SAUCE00")
	r[94], r[95] = 1, 1
	binary.LittleEndian.PutUint16(r[96:], uint16(width))
	binary.LittleEndian.PutUint16(r[98:], uint16(height))
	r[105] = flags
	return r
}

// withSauce gives UTF-8 input: a BOM, data, 0x1A and a character SAUCE record
// with the width, the height and the TFlags.
func withSauce(data string, width, height int, flags byte) []byte {
	return append([]byte(bom+data+"\x1a"), sauceRec(width, height, flags)...)
}

// cp437File gives CP437 input: data, 0x1A and a character SAUCE record.
func cp437File(data string, width, height int, flags byte) []byte {
	return append([]byte(data+"\x1a"), sauceRec(width, height, flags)...)
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
	res, ds := read("a.ans", data)
	if ds.HasErrors() || res == nil {
		t.Fatalf("errors:\n%s", diagText(ds))
	}
	return res.art
}

// diags gives the diagnostics of the input data.
func diags(data []byte) string {
	_, ds := read("a.ans", data)
	return diagText(ds)
}

func cell(a *screen.Art, row, col int) screen.Cell { return a.Rows[row-1][col-1] }

// cells gives the graphemes of n cells from row:col and their style, when they have one style.
func cells(a *screen.Art, row, col, n int) string {
	var b strings.Builder
	for i := range n {
		c := cell(a, row, col+i)
		if c.Style != cell(a, row, col).Style {
			return fmt.Sprintf("style %s at col %d", c.Style, col+i)
		}
		b.WriteString(c.Grapheme)
	}
	return b.String() + " " + cell(a, row, col).Style.String()
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
		{"cursor position and save", "\x1b[2;3Ha\x1b[s\x1b[Hb\x1b[uc", 4, 0, "|b   |\n|  ac|\n"},
		{"ESC 7 and ESC 8", "ab\x1b7\r\n\x1b8c", 4, 0, "|abc |\n"},
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
		{"RGB", "\x1b[38;2;1;2;3;48;2;4;5;6mA", 0, "#010203/#040506"},
		{"RGB of PabloDraw", "\x1b[1;1;2;3t\x1b[0;4;5;6tA", 0, "#010203/#040506"},
		{"bold and RGB of PabloDraw", "\x1b[1m\x1b[1;1;2;3tA", 0, "#010203/0"},
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
	a := parse(t, cp437File("\x1b[44mabcd\x1b[1;2H\x1b[K", 4, 0, 0))
	if got, want := dump(t, a), "|a   |\nstyles\n1: 1-4 7/4\n"; got != want {
		t.Errorf("K 0:\n%s", got)
	}
	a = parse(t, withSauce("\x1b[31;1;44mabc\x1b[1K", 4, 0, 0))
	if got, want := dump(t, a), "|    |\nstyles\n1: 1-4 7/4\n"; got != want {
		t.Errorf("K 1:\n%s", got)
	}
	// CSI 2 J erases the rows up to the height, then the play continues at 1:1.
	// The grid height counts the rows after the last CSI 2 J.
	a = parse(t, withSauce("ab\nc\x1b[41m\x1b[2Jx", 4, 0, 0))
	if got, want := dump(t, a), "|x   |\nstyles\n1: 1-4 7/1\n"; got != want {
		t.Errorf("J:\n%s", got)
	}
	a = parse(t, withSauce("ab\nc\x1b[2J\x1b[3;1H\x1b[K", 4, 2, 0))
	if a.Height != 3 {
		t.Errorf("J and K at row 3: height %d", a.Height)
	}
	a = parse(t, withSauce("ab\nc\x1b[2J", 4, 1, 0))
	if a.Height != 1 {
		t.Errorf("J and TInfo2 1: height %d", a.Height)
	}
}

// TestNoSauce checks the input without a SAUCE record: CP437 of width 80.
func TestNoSauce(t *testing.T) {
	a := parse(t, []byte("\xdb\r\nb"))
	if a.Width != 80 || a.Height != 2 || cell(a, 1, 1).Grapheme != "█" || cell(a, 2, 1).Grapheme != "b" {
		t.Errorf("size %dx%d, 1:1 %q", a.Width, a.Height, cell(a, 1, 1).Grapheme)
	}
	if got, want := diags([]byte(strings.Repeat("x", 81)+"\x1b[?7lyz")), ""; got != want {
		t.Errorf("got %q", got)
	}
}

func TestErrors(t *testing.T) {
	noComnt := withSauce("ab", 4, 0, 0)
	noComnt[len(noComnt)-128+104] = 2
	tests := []struct {
		name  string
		data  []byte
		diags string
	}{
		{"file over 1 MB", make([]byte, screen.MaxFileBytes+1), "a.ans:1:1: file is over 1048576 bytes"},
		{"row 1001", withSauce("\x1b[1000Ha\nb", 4, 0, 0), "a.ans:1001:1: art is over 1000 rows"},
		{"erase at row 1001", withSauce("\x1b[1000H\n\x1b[K", 4, 0, 0), "a.ans:1001:1: art is over 1000 rows"},
		{"SAUCE width", withSauce("", 300, 0, 0), "a.ans:1:1: SAUCE width 300 is over 255"},
		{"SAUCE height", withSauce("", 80, 1001, 0), "a.ans:1:1: SAUCE height 1001 is over 1000 rows"},
		{"no COMNT block", noComnt, "a.ans:1:1: SAUCE has 2 comment lines, but no COMNT block"},
		{"data after 0x1A", cp437File("ab\x1acd", 4, 0, 0), "a.ans:1:3: data after 0x1A"},
		{"data after 0x1A and a full row", cp437File("abcd\x1a\x1a", 4, 0, 0), "a.ans:2:1: data after 0x1A"},
		{"invalid UTF-8", withSauce("\tab\xb0\xb1", 4, 0, 0), "a.ans:1:3: invalid UTF-8"},
		{"control character", withSauce("a\tb", 4, 0, 0), "a.ans:1:2: control character U+0009"},
		{"glyph controls of icy_draw UTF-8", withSauce("a\x08\x7f", 4, 0, 0),
			"a.ans:1:2: control character U+0008\na.ans:1:2: control character U+007F"},
		{"C1 control", withSauce("a\u0085b", 4, 0, 0), "a.ans:1:2: control character U+0085"},
		{"zero width", withSauce("\u200bb", 4, 0, 0), "a.ans:1:1: zero-width grapheme U+200B"},
		{"wide in width 1", withSauce("日", 1, 0, 0), "a.ans:1:1: wide grapheme in art of width 1"},
		{"grapheme over 32 bytes", withSauce("a"+"e"+strings.Repeat("\u0301", 20), 4, 0, 0),
			"a.ans:1:2: grapheme over 32 bytes: 41 bytes"},
		{"cursor right past row 1000", withSauce(strings.Repeat("\n", 1000)+"\x1b[1Cx", 4, 0, 0),
			"a.ans:1001:2: art is over 1000 rows"},
		{"restore past row 1000", withSauce(strings.Repeat("\n", 1001)+"\x1b[s\x1b[H\x1b[ux", 4, 0, 0),
			"a.ans:1002:1: art is over 1000 rows"},
		{"ESC 8 past row 1000", withSauce(strings.Repeat("\n", 1001)+"\x1b7\x1b[H\x1b8x", 4, 0, 0),
			"a.ans:1002:1: art is over 1000 rows"},
		{"sequences", withSauce("ab\x1b[5n\x1b[1J\x1b[3K\x1b[1:2m\x1b[=5h\x1b[1 A", 4, 0, 0),
			`a.ans:1:3: unknown sequence "ESC[5n"` + "\n" +
				`a.ans:1:3: unknown sequence "ESC[1J"` + "\n" +
				`a.ans:1:3: unknown sequence "ESC[3K"` + "\n" +
				`a.ans:1:3: unknown sequence "ESC[1:2m"` + "\n" +
				`a.ans:1:3: unknown sequence "ESC[=5h"` + "\n" +
				`a.ans:1:3: unknown sequence "ESC[1 A"`},
		{"parameter counts", withSauce("\x1b[1;2A\x1b[1;2b\x1b[1;2;3H\x1b[1s\x1b[1u\x1b[2;2J\x1b[0;0K", 4, 0, 0),
			`a.ans:1:1: unknown sequence "ESC[1;2A"` + "\n" +
				`a.ans:1:1: unknown sequence "ESC[1;2b"` + "\n" +
				`a.ans:1:1: unknown sequence "ESC[1;2;3H"` + "\n" +
				`a.ans:1:1: unknown sequence "ESC[1s"` + "\n" +
				`a.ans:1:1: unknown sequence "ESC[1u"` + "\n" +
				`a.ans:1:1: unknown sequence "ESC[2;2J"` + "\n" +
				`a.ans:1:1: unknown sequence "ESC[0;0K"`},
		{"modes", withSauce("\x1b[?1h\x1b[?7;33h\x1b[?h\x1b[33h", 4, 0, 0),
			`a.ans:1:1: unknown sequence "ESC[?1h"` + "\n" +
				`a.ans:1:1: unknown sequence "ESC[?7;33h"` + "\n" +
				`a.ans:1:1: unknown sequence "ESC[?h"` + "\n" +
				`a.ans:1:1: unknown sequence "ESC[33h"`},
		{"font page", withSauce("\x1b[0;5 D\x1b[1;5 D\x1b[0 D\x1b[0;5  D", 4, 0, 0),
			`a.ans:1:1: unknown sequence "ESC[1;5 D"` + "\n" +
				`a.ans:1:1: unknown sequence "ESC[0 D"` + "\n" +
				`a.ans:1:1: unknown sequence "ESC[0;5  D"`},
		{"SGR", withSauce("a\x1b[99;099;31mb\x1b[38;5m\x1b[48;2;1;2;256;31mc", 4, 0, 0),
			`a.ans:1:2: unknown SGR parameter 99 in "ESC[99;099;31m"` + "\n" +
				`a.ans:1:2: unknown SGR parameter 099 in "ESC[99;099;31m"` + "\n" +
				`a.ans:1:3: invalid colour in "ESC[38;5m"` + "\n" +
				`a.ans:1:3: invalid colour in "ESC[48;2;1;2;256;31m"`},
		{"RGB of PabloDraw", withSauce("\x1b[2;1;2;3t\x1b[1;1;2t\x1b[0;1;2;3;4t\x1b[1;256;0;0t", 4, 0, 0),
			`a.ans:1:1: unknown sequence "ESC[2;1;2;3t"` + "\n" +
				`a.ans:1:1: invalid colour in "ESC[1;1;2t"` + "\n" +
				`a.ans:1:1: invalid colour in "ESC[0;1;2;3;4t"` + "\n" +
				`a.ans:1:1: invalid colour in "ESC[1;256;0;0t"`},
		{"ESC and another byte", withSauce("abcd\x1bXa\x1b\n\x1b░", 4, 0, 0),
			`a.ans:2:1: unknown sequence "ESCX"` + "\n" + `a.ans:2:3: unknown sequence "ESC\n"` + "\n" +
				`a.ans:3:1: unknown sequence "ESC░"`},
		{"ESC and CP437", cp437File("\x1b\xb0", 4, 0, 0), `a.ans:1:1: unknown sequence "ESC░"`},
		{"incomplete sequences", withSauce("a\x1b[1;\x01b\x1b[", 4, 0, 0),
			`a.ans:1:2: incomplete sequence "ESC[1;"` + "\n" +
				`a.ans:1:2: control character U+0001` + "\n" +
				`a.ans:1:3: incomplete sequence "ESC["`},
		{"ESC at the end", withSauce("a\x1b", 4, 0, 0), `a.ans:1:2: incomplete sequence "ESC"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := diags(tt.data); got != tt.diags {
				t.Errorf("got\n%s\nwant\n%s", got, tt.diags)
			}
		})
	}
}

// TestPlayAfterErrors checks that the play goes on after an error of section 5:
// the byte that ends an incomplete sequence plays, and the rest of an SGR plays.
func TestPlayAfterErrors(t *testing.T) {
	d := &screen.FileDiags{File: "a.ans"}
	p := newPlayer(d, 4, false, true)
	p.play([]byte("\x1b[1;\x1b[31;99;44mA\x1bXB\x1b[5nC"))
	if got := cells(&screen.Art{Rows: p.grid(1), Width: 4}, 1, 1, 1); got != "A 1/4" {
		t.Errorf("1:1 = %q", got)
	}
	if got := p.grid(1)[0]; got[1].Grapheme != "X" || got[2].Grapheme != "B" || got[3].Grapheme != "C" {
		t.Errorf("row %v", got)
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
	// A long SGR of unknown parameters stops at the diagnostic limit in less than 1 s.
	start := time.Now()
	sgr := withSauce("\x1b["+strings.Repeat("9;", (screen.MaxFileBytes-200)/2)+"9m", 4, 0, 0)
	if _, ds := read("a.ans", sgr); len(ds) != screen.MaxFileDiags+1 || time.Since(start) >= time.Second {
		t.Errorf("%d diagnostics in %v", len(ds), time.Since(start))
	}
	// A long sequence is cut in the message.
	if got := diags(withSauce("\x1b["+strings.Repeat("1;", 1000)+"z", 4, 0, 0)); len(got) > 80 || !strings.Contains(got, `..."`) {
		t.Errorf("message %q", got)
	}
}

func TestTooManyDiagnostics(t *testing.T) {
	_, ds := read("a.ans", withSauce(strings.Repeat("\t", 200), 4, 0, 0))
	if len(ds) != screen.MaxFileDiags+1 || ds[len(ds)-1].Msg != "too many diagnostics" {
		t.Errorf("%d diagnostics, last %q", len(ds), ds[len(ds)-1].Msg)
	}
}

func FuzzParse(f *testing.F) {
	for _, form := range forms {
		f.Add(form.file)
	}
	f.Add(withSauce("日\x1b[?7l\x1b[5b\x1b[2J{w}\x1b[1;5;7m\x1b[K", 3, 2, 1))
	f.Fuzz(func(t *testing.T, data []byte) {
		res, ds := read("f.ans", data)
		if res == nil {
			if !ds.HasErrors() {
				t.Fatal("no art and no error")
			}
			return
		}
		a := res.art
		if a.Height > art.MaxRows || a.Width > screen.MaxWidth || len(a.Rows) != a.Height {
			t.Fatalf("size %dx%d, %d rows", a.Width, a.Height, len(a.Rows))
		}
		for i, r := range a.Rows {
			if len(r) != a.Width {
				t.Fatalf("row %d has %d cells", i+1, len(r))
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
			got, gds := read("a.ans", withSauce(tt.rep, width, 0, 0))
			want, wds := read("a.ans", withSauce(tt.plain, width, 0, 0))
			if gds.HasErrors() != wds.HasErrors() || got != nil && dump(t, got.art) != dump(t, want.art) {
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
		prefix, op string
		cost       int    // the cells of prefix and op
		at         string // the position of the error
	}{
		{"grapheme", "", "\x1b[5;2Ha", 1, "5:2"},
		{"wide grapheme", "", "\x1b[5;2H日", 2, "5:2"},
		{"erase", "", "\x1b[5;2H\x1b[K", w - 1, "5:2"},
		{"clear", "\x1b[7H\x1b[2K", "\x1b[3;3H\x1b[2J", w + 7*w, "3:3"},
		{"repeat with wrap", "", "\x1b[5;250Ha\x1b[9b", 10, "6:4"},
		{"repeat without wrap", "", "\x1b[?7l\x1b[5;250Ha\x1b[9b", 10, "5:255"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// load gives the errors; the repeat without wrap also overwrites a cell (section 6).
			load := func(extra int) screen.Diags {
				data := withSauce(tt.prefix+filler(MaxPlayCells-tt.cost+extra)+tt.op, w, 0, 0)
				_, ds := read("a.ans", data)
				return slices.DeleteFunc(ds, func(d screen.Diag) bool { return d.Warning })
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
			_, ds := read("a.ans", data)
			if d := time.Since(start); d >= time.Second {
				t.Errorf("load took %v", d)
			}
			if !strings.HasSuffix(diagText(ds), ": art play is over 4080000 written and erased cells") {
				t.Errorf("diagnostics:\n%s", diagText(ds))
			}
		})
	}
}

// TestOverwrite checks the animation warning of section 6: only the first overwrite
// of a cell that a grapheme wrote gives it.
func TestOverwrite(t *testing.T) {
	const msg = ": warning: art overwrites a cell, the import keeps the last frame"
	tests := []struct{ name, data, at string }{
		{"grapheme", "ab\x1b[Hc", "1:1"},
		{"erase in the row", "ab\x1b[1;2H\x1b[K", "1:2"},
		{"clear", "ab\x1b[2;2H\x1b[2J", "1:1"},
		{"erase with a pending wrap", "abcd\x1b[K", "1:4"},
		{"half of a wide grapheme", "日\x1b[1;2Hx", "1:2"},
		{"wrap off at the right edge", "\x1b[?7labcde", "1:4"},
		{"repeat", "ab\x1b[H\x1b[3b", "1:1"},
		{"only the first", "ab\x1b[Hcd\x1b[He", "1:1"},
		{"erased cell", "\x1b[44m\x1b[K\x1b[2J\x1b[Ha", ""},
		{"clear at the start", "\x1b[2Jab", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := ""
			if tt.at != "" {
				want = "a.ans:" + tt.at + msg
			}
			if got := diags(withSauce(tt.data, 4, 0, 0)); got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}
	// The cell keeps the last write.
	if got := cells(parse(t, withSauce("ab\x1b[Hc", 4, 0, 0)), 1, 1, 2); got != "cb 7/0" {
		t.Errorf("1:1 = %q", got)
	}
}
