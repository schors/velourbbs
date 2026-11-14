package artimport

import (
	"bytes"
	"strings"
	"testing"
)

// pad gives s padded with spaces to n bytes.
func pad(s string, n int) string { return s + strings.Repeat(" ", n-len(s)) }

// editorSauce gives the end of a file as the editors write it: 0x1A, a COMNT block
// when there are comments, and a character SAUCE record with the title, the date,
// the size, the TFlags and the TInfoS "IBM VGA" padded with NUL.
func editorSauce(title, date string, width, height int, flags byte, comments ...string) []byte {
	b := []byte{0x1a}
	if len(comments) > 0 {
		b = append(b, "COMNT"...)
		for _, c := range comments {
			b = append(b, pad(c, 64)...)
		}
	}
	r := sauceRec(width, height, flags)
	copy(r[7:], pad(title, 35)+pad("", 20)+pad("", 20)+date)
	r[104] = byte(len(comments))
	copy(r[106:], "IBM VGA")
	return append(b, r...)
}

// form gives an editor file: the data and the end of editorSauce.
func form(data string, end []byte) []byte { return append([]byte(data), end...) }

// The test art of width 8 and height 3 in each form of section 3:
//
//	row 1: "A" bright red on blue, "B" in 7/0;
//	row 2: 8 full blocks (CP437 0xDB): a full row without CR LF;
//	row 3: "x", 5 spaces, "y", the bullet (CP437 0x07).
//
// The bytes follow the source of each editor: Moebius 5df8344 app/libtextmode/ansi.js,
// icy_tools IcyDraw0.6.2 crates/icy_engine/src/formats/ansi_v2.rs,
// PabloDraw 3.3.14-beta Source/Pablo/Formats/Character/Types/Ansi.cs.
var (
	blocks = strings.Repeat("\xdb", 8)
	forms  = []struct {
		name string
		file []byte
	}{
		{"Moebius Save", form("\x1b[0m\x1b[1;31;44mA\x1b[0mB\r\n"+blocks+"x     y\x07",
			editorSauce("T", "20261114", 8, 3, 0x02))},
		{"icy_draw ANSI.SYS", form("\x1b[1;31;44mA\x1b[0mB\r\n"+blocks+"x     y\x07",
			editorSauce("T", "00000000", 8, 3, 0x12))},
		{"icy_draw VT-100", form("\x1b[1;31;44mA\x1b[0mB\r\n"+blocks+"x\x1b[5Cy\x07",
			editorSauce("T", "00000000", 8, 3, 0x12))},
		{"icy_draw IcyTerm", form("\x1b[1;31;44mA\x1b[0mB\r\n\xdb\x1b[7bx\x1b[5Cy\x07",
			editorSauce("T", "00000000", 8, 3, 0x12))},
		{"icy_draw UTF-8", form(bom+"\x1b[1;31;44mA\x1b[0mB\x1b[0m\r\n█\x1b[7bx\x1b[5Cy•",
			editorSauce("T", "00000000", 8, 3, 0x12))},
		{"PabloDraw Save as ANSI", form("\x1b[0;1;44;31mA\x1b[0mB\r\n"+blocks+"x\x1b[5Cy\x07",
			editorSauce("T", "20261114", 8, 3, 0x12))},
	}
)

func TestForms(t *testing.T) {
	const want = "|AB      |\n|████████|\n|x     y•|\nstyles\n1: 1-1 9/4, 2-8 7/0\n2: 1-8 7/0\n3: 1-8 7/0\n"
	for _, f := range forms {
		t.Run(f.name, func(t *testing.T) {
			res, ds := read("a.ans", f.file)
			if len(ds) > 0 || res == nil {
				t.Fatalf("diagnostics:\n%s", diagText(ds))
			}
			if got := dump(t, res.art); got != want {
				t.Errorf("got\n%swant\n%s", got, want)
			}
			if !bytes.HasPrefix(res.sauce.Title, []byte("T ")) {
				t.Errorf("title %q", res.sauce.Title)
			}
		})
	}
}

// TestFormDetails checks the sequences and SAUCE fields of each editor
// that the test art of TestForms does not use.
func TestFormDetails(t *testing.T) {
	tests := []struct {
		name     string
		file     []byte
		row, col int
		want     string // the grapheme and the style of the cell
		height   int
	}{
		{"icy_draw font page and RGB", form("\x1b[0;1 DA\x1b[1;255;136;0tB", editorSauce("", "00000000", 4, 1, 0x12)),
			1, 2, "B #ff8800/0", 1},
		{"icy_draw UTF-8 RGB", form(bom+"\x1b[38;2;1;2;3;48;2;4;5;6mA", editorSauce("", "00000000", 4, 1, 0x12)),
			1, 1, "A #010203/#040506", 1},
		{"icy_draw iCE", form("\x1b[?33h\x1b[5;44mX\x1b[?33l", editorSauce("", "00000000", 4, 1, 0x13)),
			1, 1, "X 7/12", 1},
		{"icy_draw bg before a row break", form("\x1b[44mA\x1b[40m\r\nB", editorSauce("", "00000000", 4, 2, 0x12)),
			2, 1, "B 7/0", 2},
		{"icy_draw full buffer height", form("\x1b[1;31mA", editorSauce("", "00000000", 4, 25, 0x12)),
			1, 1, "A 9/0", 25},
		{"icy_draw raw control glyphs", form("\x00\x07\x08\x09\x7f", editorSauce("", "00000000", 5, 1, 0x12)),
			1, 2, "•◘○⌂ 7/0", 1},
		{"Moebius style after a full row", form("\x1b[0m\x1b[44mABCDx", editorSauce("", "20261114", 4, 2, 0x02)),
			2, 1, "x 7/4", 2},
		{"Moebius iCE", form("\x1b[0m\x1b[5;44mX", editorSauce("", "20261114", 4, 1, 0x03)),
			1, 1, "X 7/12", 1},
		{"Moebius blink", form("\x1b[0m\x1b[5;44mX", editorSauce("", "20261114", 4, 1, 0x02)),
			1, 1, "X 7/4 blink", 1},
		{"Moebius control glyph substitutes", form("\x1b[0m\x10\x11", editorSauce("", "20261114", 4, 1, 0x02)),
			1, 1, "►◄ 7/0", 1},
		{"icy_draw UTF-8 style after a full row", form(bom+"\x1b[44mABCDx", editorSauce("", "00000000", 4, 2, 0x12)),
			2, 1, "x 7/0", 2},
		{"PabloDraw RGB", form("\x1b[0;40;37mA\x1b[0;0;0;32tB", editorSauce("", "20261114", 4, 1, 0x12)),
			1, 2, "B 7/#000020", 1},
		{"PabloDraw CR LF after the last row", form("\x1b[0;40;37mA\r\n", editorSauce("", "20261114", 4, 1, 0x12)),
			1, 1, "A 7/0", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := parse(t, tt.file)
			n := len([]rune(strings.Fields(tt.want)[0]))
			if got := cells(a, tt.row, tt.col, n); got != tt.want {
				t.Errorf("%d:%d = %q, want %q", tt.row, tt.col, got, tt.want)
			}
			if a.Height != tt.height {
				t.Errorf("height %d, want %d", a.Height, tt.height)
			}
		})
	}
}

// TestSauceResult checks that the result has the SAUCE record of the input.
func TestSauceResult(t *testing.T) {
	res, ds := read("a.ans", form("ab", editorSauce("Title", "20261114", 4, 1, 0x12, "made by hand", "line 2")))
	if len(ds) > 0 {
		t.Fatalf("diagnostics:\n%s", diagText(ds))
	}
	s := res.sauce
	if string(s.Title) != pad("Title", 35) || string(s.Date) != "20261114" || s.TFlags != 0x12 ||
		string(s.TInfoS) != "IBM VGA"+strings.Repeat("\x00", 15) {
		t.Errorf("record %+v", s)
	}
	if len(s.Lines) != 2 || string(s.Lines[0]) != pad("made by hand", 64) || string(s.Lines[1]) != pad("line 2", 64) {
		t.Errorf("comments %q", s.Lines)
	}
	if got := dump(t, res.art); got != "|ab  |\nstyles\n1: 1-4 7/0\n" {
		t.Errorf("art:\n%s", got)
	}
	res, _ = read("a.ans", []byte("ab"))
	if res.sauce != nil {
		t.Errorf("no SAUCE: record %+v", res.sauce)
	}
}

// TestMoebiusComments checks the COMNT block of Moebius after the first 0x1A (section 4, step 1):
// lines of 64 characters in UTF-8, and a block without lines.
func TestMoebiusComments(t *testing.T) {
	line := "Привет" + strings.Repeat(" ", 58)
	emoji := "😀" + strings.Repeat(" ", 62) // 2 UTF-16 code units
	rec := sauceRec(4, 1, 0x02)
	rec[104] = 3
	file := append([]byte("ab\x1aCOMNT"+line+emoji+pad("c", 64)), rec...)
	res, ds := read("a.ans", file)
	if len(ds) > 0 {
		t.Fatalf("diagnostics:\n%s", diagText(ds))
	}
	if l := res.sauce.Lines; len(l) != 3 || string(l[0]) != line || string(l[1]) != emoji || string(l[2]) != pad("c", 64) {
		t.Errorf("lines %q", l)
	}
	// Moebius writes a bare COMNT for a comment of only spaces, with Comments 0.
	res, ds = read("a.ans", append([]byte("ab\x1aCOMNT"), sauceRec(4, 1, 0x02)...))
	if len(ds) > 0 || res.sauce.Lines != nil {
		t.Errorf("bare COMNT: %q, lines %q", diagText(ds), res.sauce.Lines)
	}
	// Without a SAUCE record, COMNT is data after 0x1A.
	if got := diags([]byte("ab\x1aCOMNT")); got != "a.ans:1:3: data after 0x1A" {
		t.Errorf("no SAUCE: %q", got)
	}
}
