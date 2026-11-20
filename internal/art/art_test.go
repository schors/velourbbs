package art_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/schors/velourbbs/internal/art"
	"github.com/schors/velourbbs/internal/screen"
)

// aliases is the [fields] of screens/screens.toml in the tests: w is term.cols.
var aliases = func() screen.Fields {
	r, err := screen.NewAlias("w", "term.cols")
	if err != nil {
		panic(err)
	}
	return screen.Fields{"w": r}
}()

// sauce gives a character SAUCE record with the width, the height and the TFlags.
func sauce(width, height int, flags byte) []byte {
	r := make([]byte, 128)
	copy(r, "SAUCE00")
	r[94], r[95] = 1, 1
	binary.LittleEndian.PutUint16(r[96:], uint16(width))
	binary.LittleEndian.PutUint16(r[98:], uint16(height))
	r[105] = flags
	return r
}

// flat gives a file in the flat form: a BOM, data, 0x1A and a SAUCE record.
func flat(data string, width, height int) []byte {
	return append([]byte("\xef\xbb\xbf"+data+"\x1a"), sauce(width, height, 0)...)
}

// dump writes the grid of a in the cell dump form of screen-render, section 3.3.
func dump(t *testing.T, a *screen.Art) string {
	t.Helper()
	var b strings.Builder
	if err := screen.WritePage(&b, screen.Page{Num: 1, Rows: a.Rows, Cols: a.Width, Height: a.Height}); err != nil {
		t.Fatal(err)
	}
	return strings.TrimPrefix(b.String(), fmt.Sprintf("page 1 %dx%d fill none\n", a.Width, a.Height))
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
	a, ds := art.Parse("a.ans", data, aliases)
	if len(ds) > 0 || a == nil {
		t.Fatalf("diagnostics:\n%s", diagText(ds))
	}
	return a
}

func TestGrid(t *testing.T) {
	tests := []struct {
		name, data    string
		width, height int
		want          string
	}{
		{"rows", "ab\ncd", 4, 0, "|ab  |\n|cd  |\nstyles\n1: 1-4 7/0\n2: 1-4 7/0\n"},
		{"LF at the end", "ab\n", 4, 0, "|ab  |\nstyles\n1: 1-4 7/0\n"},
		{"empty rows", "\n\nab\n\n", 2, 0, "|  |\n|  |\n|ab|\n|  |\nstyles\n1: 1-2 7/0\n2: 1-2 7/0\n3: 1-2 7/0\n4: 1-2 7/0\n"},
		// Bytes after the last LF make a row, also SGR only.
		{"SGR after the last LF", "ab\n\x1b[0m", 2, 0, "|ab|\n|  |\nstyles\n1: 1-2 7/0\n2: 1-2 7/0\n"},
		{"TInfo2 is larger", "ab", 2, 3, "|ab|\n|  |\n|  |\nstyles\n1: 1-2 7/0\n2: 1-2 7/0\n3: 1-2 7/0\n"},
		{"TInfo2 is smaller", "a\nb\nc", 1, 2, "|a|\n|b|\n|c|\nstyles\n1: 1-1 7/0\n2: 1-1 7/0\n3: 1-1 7/0\n"},
		{"no data", "", 3, 0, "styles\n"},
		{"full row", "abcd", 4, 0, "|abcd|\nstyles\n1: 1-4 7/0\n"},
		{"wide grapheme", "日x\n日日", 4, 0, "|日x |\n|日日|\nstyles\n1: 1-4 7/0\n2: 1-4 7/0\n"},
		// The style continues across LF. Unwritten cells are a space in 7/0.
		{"style across LF", "\x1b[31;44ma\nb", 3, 0, "|a  |\n|b  |\nstyles\n1: 1-1 1/4, 2-3 7/0\n2: 1-1 1/4, 2-3 7/0\n"},
		{"space in a style", "\x1b[44m \x1b[0m", 2, 0, "|  |\nstyles\n1: 1-1 7/4, 2-2 7/0\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := parse(t, flat(tt.data, tt.width, tt.height))
			if got := dump(t, a); got != tt.want {
				t.Errorf("got\n%swant\n%s", got, tt.want)
			}
			if a.Width != tt.width || len(a.Rows) != a.Height {
				t.Errorf("width %d, height %d, %d rows", a.Width, a.Height, len(a.Rows))
			}
		})
	}
}

func TestStyles(t *testing.T) {
	tests := []struct{ data, want string }{
		{"a", "7/0"},
		{"\x1b[mA", "7/0"},
		{"\x1b[31;41;0mA", "7/0"},
		{"\x1b[4mA", "7/0 underline"},
		{"\x1b[4;24mA", "7/0"},
		{"\x1b[5mA", "7/0 blink"},
		{"\x1b[4;5;25mA", "7/0 underline"},
		{"\x1b[30;47mA", "0/7"},
		{"\x1b[37;40mA", "7/0"},
		{"\x1b[90;100mA", "8/8"},
		{"\x1b[97;107mA", "15/15"},
		{"\x1b[31;41;39;49mA", "7/0"},
		{"\x1b[38;5;200;48;5;17mA", "200/17"},
		{"\x1b[38;2;255;136;0;48;2;0;0;32mA", "#ff8800/#000020"},
		{"\x1b[31m\x1b[44mA", "1/4"},
		{"\x1b[;31mA", "1/0"},
		{"\x1b[00031mA", "1/0"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", tt.data), func(t *testing.T) {
			a := parse(t, flat(tt.data, 4, 0))
			if got := a.Rows[0][0].Style.String(); got != tt.want {
				t.Errorf("style %s, want %s", got, tt.want)
			}
		})
	}
}

// TestVGA: each index 0–15 of art is a VGA colour, also from 38;5;n and 48;5;n,
// and also in a cell that the data does not write (section 3).
func TestVGA(t *testing.T) {
	vga, idx := screen.VGAColour, screen.IndexColour
	a := parse(t, flat("\x1b[31;104mA\x1b[0;38;5;15;48;5;16mB\x1b[0;39;49mC", 4, 0))
	want := []screen.Style{
		{Fg: vga(1), Bg: vga(12)},
		{Fg: vga(15), Bg: idx(16)},
		{Fg: vga(7), Bg: vga(0)},
		{Fg: vga(7), Bg: vga(0)},
	}
	for c, w := range want {
		if got := a.Rows[0][c].Style; got != w {
			t.Errorf("cell %d: style %+v, want %+v", c+1, got, w)
		}
	}
}

// withField sets the byte at the offset of the SAUCE record of file to v.
func withField(file []byte, offset int, v byte) []byte {
	file[len(file)-128+offset] = v
	return file
}

// TestNotFlat checks the steps 1, 2 and 4 of section 2.
func TestNotFlat(t *testing.T) {
	const notFlat = "a.ans:1:1: not in the flat form, convert it with velourbbs art import"
	tests := []struct {
		name string
		data []byte
	}{
		{"no SAUCE", []byte("\xef\xbb\xbfab")},
		{"empty file", nil},
		{"SAUCE without the ID", withField(flat("ab", 4, 0), 0, 'X')},
		{"DataType 0", withField(flat("ab", 4, 0), 94, 0)},
		{"FileType 2", withField(flat("ab", 4, 0), 95, 2)},
		{"TInfo1 0", flat("ab", 0, 0)},
		{"TInfo1 256", flat("ab", 256, 0)},
		{"TInfo2 1001", flat("ab", 4, 1001)},
		{"iCE colours", withField(flat("ab", 4, 0), 105, 1)},
		{"no BOM", append([]byte("ab\x1a"), sauce(4, 0, 0)...)},
		{"BOM after 0x1A", append([]byte("ab\x1a\xef\xbb\xbf"), sauce(4, 0, 0)...)},
		{"CP437", append([]byte("\xb0\xb1\x1a"), sauce(4, 0, 0)...)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, ds := art.Parse("a.ans", tt.data, aliases)
			if got := diagText(ds); got != notFlat || a != nil {
				t.Errorf("got %q", got)
			}
		})
	}
}

// TestSauce checks the SAUCE fields that the loader accepts and the end of the data.
func TestSauce(t *testing.T) {
	// TInfo1 255, TInfo2 1000 and the other TFlags bits.
	a := parse(t, withField(flat("ab", 255, 1000), 105, 0x1e))
	if a.Width != 255 || a.Height != 1000 {
		t.Errorf("size %dx%d", a.Width, a.Height)
	}
	// No 0x1A before the SAUCE.
	a = parse(t, append([]byte("\xef\xbb\xbfab"), sauce(2, 0, 0)...))
	if got := dump(t, a); got != "|ab|\nstyles\n1: 1-2 7/0\n" {
		t.Errorf("no 0x1A:\n%s", got)
	}
	// A COMNT block of two lines comes before the SAUCE.
	rec := sauce(2, 0, 0)
	rec[104] = 2
	file := append([]byte("\xef\xbb\xbfab\x1aCOMNT"), bytes.Repeat([]byte("c"), 128)...)
	a = parse(t, append(file, rec...))
	if got := dump(t, a); got != "|ab|\nstyles\n1: 1-2 7/0\n" {
		t.Errorf("COMNT:\n%s", got)
	}
	// The loader removes only one 0x1A.
	_, ds := art.Parse("a.ans", flat("ab\x1a", 4, 0), aliases)
	if got, want := diagText(ds), "a.ans:1:3: control character U+001A"; got != want {
		t.Errorf("two 0x1A: got %q, want %q", got, want)
	}
}

func TestErrors(t *testing.T) {
	tests := []struct {
		name, data string
		width      int
		diags      string
	}{
		{"CR", "ab\r\ncd", 4, "a.ans:1:3: control character U+000D"},
		{"tab", "a\tb", 4, "a.ans:1:2: control character U+0009"},
		{"DEL", "a\x7fb", 4, "a.ans:1:2: control character U+007F"},
		{"C1 control", "a\u0085b", 4, "a.ans:1:2: control character U+0085"},
		{"zero width", "\n\u200bb", 4, "a.ans:2:1: zero-width grapheme U+200B"},
		{"long grapheme", "a" + strings.Repeat("\u0301", 20), 4, "a.ans:1:1: grapheme over 32 bytes: 41 bytes"},
		{"invalid UTF-8", "ab\nc\xb0d\xb1", 4, "a.ans:2:2: invalid UTF-8"},
		// Invalid UTF-8 is the only error of the file.
		{"invalid UTF-8 after errors", "\t\x1b[1m\nb\t\xff\t", 4, "a.ans:2:2: invalid UTF-8"},
		{"invalid UTF-8 after the diagnostics limit", strings.Repeat("\t", 200) + "\xff", 4, "a.ans:1:1: invalid UTF-8"},
		{"invalid UTF-8 over the width", "abcde\xff", 4, "a.ans:1:6: invalid UTF-8"},
		{"row over the width", "abcde\nabcdef", 4, "a.ans:1:5: row is over 4 columns\na.ans:2:5: row is over 4 columns"},
		// After the row is over the width, the columns count on.
		{"error after the row is over the width", "abcde日\tx\x1b[1m", 4,
			"a.ans:1:5: row is over 4 columns\na.ans:1:8: control character U+0009\n" + `a.ans:1:9: unknown SGR parameter 1 in "ESC[1m"`},
		{"wide grapheme over the width", "abc日", 4, "a.ans:1:4: row is over 4 columns"},
		{"wide grapheme in width 1", "日", 1, "a.ans:1:1: row is over 1 columns"},
		{"row 1001", strings.Repeat("\n", 1000) + "a\x1b[5n", 4, "a.ans:1001:1: art is over 1000 rows"},
		{"SGR in row 1001", strings.Repeat("\n", 1000) + "\x1b[0m", 4, "a.ans:1001:1: art is over 1000 rows"},
		{"bold", "a\x1b[1mb", 4, `a.ans:1:2: unknown SGR parameter 1 in "ESC[1m"`},
		{"reverse", "\x1b[7m", 4, `a.ans:1:1: unknown SGR parameter 7 in "ESC[7m"`},
		{"bold off", "\x1b[22m", 4, `a.ans:1:1: unknown SGR parameter 22 in "ESC[22m"`},
		{"large parameter", "\x1b[123456m", 4, `a.ans:1:1: unknown SGR parameter 123456 in "ESC[123456m"`},
		{"leading zeros", "\x1b[0;0001m", 4, `a.ans:1:1: unknown SGR parameter 0001 in "ESC[0;0001m"`},
		{"cursor position", "a\x1b[2;3Hb", 4, `a.ans:1:2: unknown sequence "ESC[2;3H"`},
		{"cursor right", "\x1b[C", 4, `a.ans:1:1: unknown sequence "ESC[C"`},
		{"iCE mode", "\x1b[?33h", 4, `a.ans:1:1: unknown sequence "ESC[?33h"`},
		{"font page", "\x1b[0;1 D", 4, `a.ans:1:1: unknown sequence "ESC[0;1 D"`},
		{"intermediate before m", "\x1b[0 m", 4, `a.ans:1:1: unknown sequence "ESC[0 m"`},
		{"colon", "\x1b[1:2m", 4, `a.ans:1:1: unknown sequence "ESC[1:2m"`},
		{"invalid colour", "\x1b[38;5m\x1b[48;5;256m\x1b[38;2;1;2m\x1b[38;3;1m", 4,
			`a.ans:1:1: invalid colour in "ESC[38;5m"` + "\n" + `a.ans:1:1: invalid colour in "ESC[48;5;256m"` + "\n" +
				`a.ans:1:1: invalid colour in "ESC[38;2;1;2m"` + "\n" + `a.ans:1:1: invalid colour in "ESC[38;3;1m"`},
		{"one error for a sequence", "\x1b[1;7;99m", 4, `a.ans:1:1: unknown SGR parameter 1 in "ESC[1;7;99m"`},
		{"incomplete at the end", "a\x1b[31", 4, `a.ans:1:2: incomplete sequence "ESC[31"`},
		// The byte that ends an incomplete sequence reads.
		{"incomplete before LF", "\x1b[31\nb\x1b[1m", 4, `a.ans:1:1: incomplete sequence "ESC[31"` + "\n" + `a.ans:2:2: unknown SGR parameter 1 in "ESC[1m"`},
		{"ESC at the end", "ab\x1b", 4, `a.ans:1:3: ESC without "["`},
		{"ESC and another byte", "\x1b7a\x1b[1m", 4, `a.ans:1:1: ESC without "["` + "\n" + `a.ans:1:3: unknown SGR parameter 1 in "ESC[1m"`},
		{"long sequence", "\x1b[" + strings.Repeat("1;", 100) + "z", 4, `a.ans:1:1: unknown sequence "ESC[1;1;1;1;1;1;1;1;1;1;1;1;..."`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, ds := art.Parse("a.ans", flat(tt.data, tt.width, 0), aliases)
			if got := diagText(ds); got != tt.diags || a != nil {
				t.Errorf("got\n%s\nwant\n%s", got, tt.diags)
			}
		})
	}
}

// TestRowLimit checks that 1000 rows load.
func TestRowLimit(t *testing.T) {
	a := parse(t, flat(strings.Repeat("a\n", 1000), 1, 0))
	if a.Height != 1000 {
		t.Errorf("height %d", a.Height)
	}
}

func TestFields(t *testing.T) {
	tests := []struct {
		name, data string
		fields     string
		diags      string
	}{
		{"left", "{w  }", "[{1 1 5 term.cols false 7/0}]", ""},
		{"right", "{  w}", "[{1 1 5 term.cols true 7/0}]", ""},
		{"spaces on both sides", "{ term.cols }", "[{1 1 13 term.cols true 7/0}]", ""},
		{"plain", "{ BBS } {a {} {{w} {@x}", "[{1 16 3 term.cols false 7/0}]", ""},
		// Art has no stretch token (bbt-format, section 6.6).
		{"stretch is plain", "{~} {~.}", "[]", ""},
		{"style of the first cell", "\x1b[33;44m{w}", "[{1 1 3 term.cols false 3/4}]", ""},
		{"space with another foreground", "\x1b[31m{\x1b[32m \x1b[31mw}", "[{1 1 4 term.cols true 1/0}]", ""},
		{"two foregrounds", "\n\n              \x1b[93m{\x1b[97mw}", "[]",
			`a.ans:3:15: token "{w}" changes the foreground: 11 at col 15, 15 at col 16`},
		{"two backgrounds", "{\x1b[41m w}", "[]",
			`a.ans:1:1: token "{ w}" changes the background: 0 at col 1, 1 at col 2`},
		// The aliases of screens/screens.toml resolve the path once, at the load.
		{"unknown field", "{x} {item.A} {term.rows}", "[]",
			"a.ans:1:1: unknown field \"x\"\na.ans:1:5: field \"item.A\" is valid only in {#if}"},
		{"unknown field and two backgrounds", "{\x1b[41m x}", "[]",
			"a.ans:1:1: token \"{ x}\" changes the background: 0 at col 1, 1 at col 2\na.ans:1:1: unknown field \"x\""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, ds := art.Parse("a.ans", flat(tt.data, 30, 0), aliases)
			if got := diagText(ds); got != tt.diags {
				t.Errorf("diagnostics %q, want %q", got, tt.diags)
			}
			var fields []string
			if a != nil {
				for _, f := range a.Fields {
					fields = append(fields, fmt.Sprintf("{%d %d %d %s %t %s}", f.Row, f.Col, f.Width, f.Ref, f.Right, f.Style))
				}
			}
			if got := "[" + strings.Join(fields, " ") + "]"; got != tt.fields {
				t.Errorf("fields %s, want %s", got, tt.fields)
			}
		})
	}
}

// The list row is the art row with row.* fields (list-view, section 9).
// The rectangle goes from the first row.* token to the last. A row.* field on another row is an error.
func TestListRow(t *testing.T) {
	tests := []struct {
		name, data string
		list       string
		diags      string
	}{
		{"no list row", "{w}", "<nil>", ""},
		{"list row", "{w}\n {row.num} {list.page} {  row.title}", "&{2 2 36}", ""},
		{"two list rows", "{row.num}\n{row.title} {row.path}\n{row.num}", "",
			"a.ans:2:1: field \"row.title\": art has at most one list row, the list row is row 1\n" +
				"a.ans:3:1: field \"row.num\": art has at most one list row, the list row is row 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, ds := art.Parse("a.ans", flat(tt.data, 40, 0), aliases)
			if got := diagText(ds); got != tt.diags {
				t.Errorf("diagnostics %q, want %q", got, tt.diags)
			}
			if a != nil {
				if got := fmt.Sprint(a.List); got != tt.list {
					t.Errorf("list %s, want %s", got, tt.list)
				}
			}
		})
	}
}

// The boxes of section 5 and their errors. The position of an error is the start of its token.
func TestBoxes(t *testing.T) {
	tests := []struct {
		name, data string
		boxes      string
		diags      string
	}{
		{"box", "ab\x1b[31m{#if term.cols > 40}[{w}]{/if}x", "[{1 3 32 23 5 1/0} {term.cols at 24}]", ""},
		{"empty content", "{#if term.cols>1}{/if}", "[{1 1 22 18 0 7/0}]", ""},
		{"two boxes", "{#if term.cols>1}a{/if}{#if term.rows>1}b{/if}", "[{1 1 23 18 1 7/0} {1 24 46 41 1 7/0}]", ""},
		{"string", `{#if term.type == "{w"}x{/if}`, "[{1 1 29 24 1 7/0}]", ""},
		{"item", "{#if item.S}[S]ysop{/if}", "[{1 1 24 13 7 7/0}]", ""},
		{"no end", "x {#if term.cols > 1}abc", "",
			"a.ans:1:3: {#if} has no {/if} on its row"},
		{"end on the next row", "{#if term.cols > 1}a\n{/if}", "",
			"a.ans:1:1: {#if} has no {/if} on its row\na.ans:2:1: {/if} without {#if}"},
		{"nest", "{#if term.cols > 1}{#if term.rows > 1}a{/if}{/if}", "",
			"a.ans:1:20: {#if} in a box: boxes do not nest\na.ans:1:45: {/if} without {#if}"},
		{"other tokens", "{#else} {!title} {>file} {#each} {/each} {!}", "",
			"a.ans:1:1: token \"{#else}\" is not valid in art\n" +
				"a.ans:1:9: token \"{!title}\" is not valid in art\n" +
				"a.ans:1:18: token \"{>file}\" is not valid in art\n" +
				"a.ans:1:26: token \"{#each}\" is not valid in art\n" +
				"a.ans:1:34: token \"{/each}\" is not valid in art\n" +
				"a.ans:1:42: token \"{!}\" is not valid in art"},
		{"no closing brace", "{w} {!title", "", `a.ans:1:5: token has no "}"`},
		{"condition", "{#if term.colss > 1}a{/if}{#if}b{/if}{#if term.cols}c{/if}", "",
			"a.ans:1:1: unknown field \"term.colss\"\n" +
				"a.ans:1:27: expected a field, found end of condition\n" +
				"a.ans:1:38: integer field \"term.cols\" needs an operator"},
		{"marker colours", "{\x1b[31m#if term.cols > 1}a{\x1b[41m/if}", "",
			"a.ans:1:1: token \"{#if term.cols > 1}\" changes the foreground: 7 at col 1, 1 at col 2\n" +
				"a.ans:1:21: token \"{/if}\" changes the background: 0 at col 21, 1 at col 22"},
		{"row field", "{#if row.num > 1}a{/if}", "", `a.ans:1:1: field "row.num" is not valid in a box condition`},
		{"list row", "{row.num} {#if term.cols > 1}a{/if}", "", "a.ans:1:11: box in the list row"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, ds := art.Parse("a.ans", flat(tt.data, 60, 0), aliases)
			if got := diagText(ds); got != tt.diags {
				t.Errorf("diagnostics %q, want %q", got, tt.diags)
			}
			if a == nil {
				return
			}
			var got []string
			for _, b := range a.Boxes {
				got = append(got, fmt.Sprintf("{%d %d %d %d %d %s}", b.Row, b.Col, b.End, b.Content, b.Len, b.Style))
			}
			for _, f := range a.Fields {
				got = append(got, fmt.Sprintf("{%s at %d}", f.Ref, f.Col))
			}
			if s := "[" + strings.Join(got, " ") + "]"; s != tt.boxes {
				t.Errorf("boxes %s, want %s", s, tt.boxes)
			}
		})
	}
}

func TestTooManyDiagnostics(t *testing.T) {
	_, ds := art.Parse("a.ans", flat(strings.Repeat("\t", 200)+"\n\x1b[1m", 4, 0), aliases)
	if len(ds) != screen.MaxFileDiags+1 || ds[len(ds)-1].Msg != "too many diagnostics" {
		t.Errorf("%d diagnostics, last %q", len(ds), ds[len(ds)-1].Msg)
	}
}

func TestLoad(t *testing.T) {
	fsys := fstest.MapFS{
		"big.ans": {Data: bytes.Repeat([]byte("a"), screen.MaxFileBytes+1)},
		"ok.ans":  {Data: flat("ab", 2, 0)},
	}
	if _, ds := art.Load(fsys, "big.ans", aliases); diagText(ds) != "big.ans:1:1: file is over 1048576 bytes" {
		t.Errorf("big: %s", diagText(ds))
	}
	if _, ds := art.Load(fsys, "none.ans", aliases); !ds.HasErrors() {
		t.Error("missing file: no error")
	}
	if a, ds := art.Load(fsys, "ok.ans", aliases); a == nil || len(ds) > 0 {
		t.Errorf("ok: %s", diagText(ds))
	}
}

func FuzzParse(f *testing.F) {
	f.Add(flat("\x1b[38;5;14;48;5;4mHello\n\x1b[0m{  w}=日\x1b[4;5m\n\n", 12, 5))
	f.Add(flat("\x1b[38;2;255;136;0m{term.cols }\x1b[0;1 D\r\n\x1b[1;5;44mX", 40, 4))
	f.Fuzz(func(t *testing.T, data []byte) {
		a, ds := art.Parse("f.ans", data, aliases)
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
			for c, cell := range r {
				if cell.Style.Attr&screen.Bold != 0 || cell.Grapheme == "" && (c == 0 || r[c-1].Grapheme == "") {
					t.Fatalf("cell %d:%d = %q %s", i+1, c+1, cell.Grapheme, cell.Style)
				}
			}
		}
		for _, fl := range a.Fields {
			if fl.Row > a.Height || fl.Col+fl.Width-1 > a.Width {
				t.Fatalf("field %v outside the grid", fl)
			}
		}
	})
}

// TestReadGrid checks that ReadGrid reads the grid by sections 2 and 3 without the fields:
// an unknown alias and a token in two colours are no errors.
func TestReadGrid(t *testing.T) {
	a, ds := art.ReadGrid("a.ans", flat("{x} {\x1b[31mw}", 8, 1))
	if len(ds) > 0 || a == nil {
		t.Fatalf("diagnostics:\n%s", diagText(ds))
	}
	if got := dump(t, a); got != "|{x} {w} |\nstyles\n1: 1-5 7/0, 6-7 1/0, 8-8 7/0\n" || a.Fields != nil {
		t.Errorf("got\n%sfields %v", got, a.Fields)
	}
	if _, ds := art.ReadGrid("a.ans", flat("\x1b[1m", 8, 1)); diagText(ds) != `a.ans:1:1: unknown SGR parameter 1 in "ESC[1m"` {
		t.Errorf("error: %q", diagText(ds))
	}
}
