package artimport

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"

	"github.com/schors/velourbbs/internal/art"
	"github.com/schors/velourbbs/internal/screen"
)

// importOK imports file and fails the test on a diagnostic.
func importOK(t *testing.T, file []byte) []byte {
	t.Helper()
	out, ds := Import("a.ans", file)
	if len(ds) > 0 || out == nil {
		t.Fatalf("diagnostics:\n%s", diagText(ds))
	}
	return out
}

// outRecord gives the SAUCE record of an output.
func outRecord(t *testing.T, out []byte) []byte {
	t.Helper()
	if len(out) < 128 || !bytes.HasPrefix(out[len(out)-128:], []byte("SAUCE00")) {
		t.Fatalf("no SAUCE record: %q", out)
	}
	return out[len(out)-128:]
}

// TestImportForms imports each form of section 3 and reads the output by art-format,
// sections 2 and 3: the grid is the grid of the play.
func TestImportForms(t *testing.T) {
	for _, f := range forms {
		t.Run(f.name, func(t *testing.T) {
			res, _ := read("a.ans", f.file)
			out := importOK(t, f.file)
			back, ds := art.ReadGrid("a.ans", out)
			if back == nil {
				t.Fatalf("output diagnostics:\n%s", diagText(ds))
			}
			if got, want := dump(t, back), dump(t, res.art); got != want {
				t.Errorf("got\n%swant\n%s", got, want)
			}
			r := outRecord(t, out)
			if string(r[7:42]) != pad("T", 35) || string(r[106:128]) != "IBM VGA"+strings.Repeat("\x00", 15) {
				t.Errorf("title %q, font %q", r[7:42], r[106:128])
			}
			if r[105] != f.file[len(f.file)-128+105]&0x1e {
				t.Errorf("TFlags %#x", r[105])
			}
		})
	}
}

// TestImportBytes checks the bytes of one output (section 9).
func TestImportBytes(t *testing.T) {
	in := form("\x1b[1;31;44mA\x1b[0mB  \r\n\r\n x", editorSauce("Title", "20261114", 6, 3, 0x13, "note"))
	data := bom + "\x1b[0;91;44mA\x1b[0mB\n\n x\n"
	rec := sauceRec(6, 3, 0x12)
	copy(rec[7:], pad("Title", 35)+pad("", 40)+"20261114")
	binary.LittleEndian.PutUint32(rec[90:], uint32(len(data)))
	rec[104] = 1
	copy(rec[106:], "IBM VGA")
	want := append([]byte(data+"\x1aCOMNT"+pad("note", 64)), rec...)
	if got := importOK(t, in); !bytes.Equal(got, want) {
		t.Errorf("got\n%q\nwant\n%q", got, want)
	}
}

func TestSGR(t *testing.T) {
	idx, rgb := screen.IndexColour, screen.RGBColour
	tests := []struct {
		style screen.Style
		want  string
	}{
		{screen.Style{Fg: idx(7), Bg: idx(0)}, "\x1b[0m"},
		{screen.Style{Fg: idx(9), Bg: idx(4), Attr: screen.Underline | screen.Blink}, "\x1b[0;4;5;91;44m"},
		{screen.Style{Fg: idx(0), Bg: idx(15)}, "\x1b[0;30;107m"},
		{screen.Style{Fg: idx(8), Bg: idx(7)}, "\x1b[0;90;47m"},
		{screen.Style{Fg: idx(200), Bg: idx(16)}, "\x1b[0;38;5;200;48;5;16m"},
		{screen.Style{Fg: rgb(1, 2, 3), Bg: rgb(255, 0, 16)}, "\x1b[0;38;2;1;2;3;48;2;255;0;16m"},
	}
	for _, tt := range tests {
		if got := string(appendSGR(nil, tt.style)); got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.style, got, tt.want)
		}
	}
	// Each style reads back: Import compares the grids.
	importOK(t, withSauce("\x1b[4;5;31;44mA\x1b[0;38;5;200;48;5;16mB\x1b[0;38;2;1;2;3;48;2;4;5;6mC\x1b[0;97;107mD", 4, 1, 0))
}

// TestImportRows checks a wide grapheme, the trailing cells and the trailing rows.
func TestImportRows(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
		data string // the output data after the BOM
	}{
		{"wide grapheme", withSauce("漢x", 4, 1, 0), "漢x\n"},
		{"trailing space with a background", withSauce("a\x1b[44m \x1b[0m  ", 6, 1, 0), "a\x1b[0;44m \n"},
		{"trailing rows", withSauce("a", 4, 3, 0), "a\n\n\n"},
		{"style over LF", withSauce("\x1b[44ma\r\nb", 4, 2, 0), "\x1b[0;44ma\nb\n"},
		{"no rows", withSauce("", 4, 0, 0), ""},
	}
	for _, tt := range tests {
		out := importOK(t, tt.in)
		data, _, _ := bytes.Cut(out, []byte{0x1a})
		if got := string(data); got != bom+tt.data {
			t.Errorf("%s: got %q, want %q", tt.name, got, bom+tt.data)
		}
	}
}

func TestImportSauce(t *testing.T) {
	// UTF-8 stays, CP437 goes through the table, trailing spaces and NUL go.
	rec := sauceRec(4, 1, 0x0c)
	copy(rec[7:], "Привет")
	copy(rec[42:], "\x80bc")
	copy(rec[62:], "ab\x00\x00")
	copy(rec[106:], "IBM VGA  ")
	r := outRecord(t, importOK(t, append([]byte("ab\x1a"), rec...)))
	if string(r[7:42]) != pad("Привет", 35) || string(r[42:62]) != pad("Çbc", 20) ||
		string(r[62:82]) != pad("ab", 20) || string(r[82:90]) != pad("", 8) ||
		string(r[106:128]) != "IBM VGA"+strings.Repeat("\x00", 15) || r[105] != 0x0c {
		t.Errorf("record %q", r)
	}

	// Without SAUCE: empty fields, width 80.
	r = outRecord(t, importOK(t, []byte("ab")))
	if string(r[7:90]) != pad("", 83) || string(r[106:128]) != strings.Repeat("\x00", 22) ||
		binary.LittleEndian.Uint16(r[96:]) != 80 || binary.LittleEndian.Uint16(r[98:]) != 1 ||
		r[94] != 1 || r[95] != 1 || r[104] != 0 || r[105] != 0 {
		t.Errorf("no SAUCE: record %q", r)
	}

	// Another SAUCE type: the text fields stay, TInfoS and TFlags do not.
	rec = sauceRec(4, 1, 0x0c)
	rec[95] = 0
	copy(rec[7:], "T")
	copy(rec[106:], "IBM VGA")
	r = outRecord(t, importOK(t, append([]byte("ab\x1a"), rec...)))
	if string(r[7:42]) != pad("T", 35) || string(r[106:128]) != strings.Repeat("\x00", 22) || r[105] != 0 {
		t.Errorf("other SAUCE: record %q", r)
	}
}

func TestImportCut(t *testing.T) {
	// A CP437 title of 35 bytes: 0x80 becomes 2 bytes of UTF-8.
	rec := sauceRec(4, 1, 0)
	copy(rec[7:], strings.Repeat("a", 34)+"\x80")
	out, ds := Import("a.ans", append([]byte("ab\x1a"), rec...))
	if got := diagText(ds); got != "a.ans:1:1: warning: SAUCE title is over 35 bytes, the import cuts it" {
		t.Errorf("title: %q", got)
	}
	if r := outRecord(t, out); string(r[7:42]) != pad(strings.Repeat("a", 34), 35) {
		t.Errorf("title %q", r[7:42])
	}

	// Moebius comment lines of 64 characters and 65 bytes: the cut keeps "é" whole.
	// The cut lines give one warning.
	line := strings.Repeat("a", 62) + "e\u0301"
	rec = sauceRec(4, 1, 0)
	rec[104] = 3
	out, ds = Import("a.ans", append([]byte("ab\x1aCOMNT"+line+pad("b", 64)+line), rec...))
	if got := diagText(ds); got != "a.ans:1:1: warning: SAUCE has 2 comment lines over 64 bytes, the import cuts them" {
		t.Errorf("comment: %q", got)
	}
	cut := pad(strings.Repeat("a", 62), 64)
	if !bytes.Contains(out, []byte("COMNT"+cut+pad("b", 64)+cut+"SAUCE00")) {
		t.Errorf("comment: %q", out)
	}

	// Over 255 comment lines: the first 255 stay.
	rec = sauceRec(4, 1, 0)
	out, ds = Import("a.ans", append([]byte("ab\x1aCOMNT"+strings.Repeat(pad("c", 64), 256)), rec...))
	if got := diagText(ds); got != "a.ans:1:1: warning: SAUCE has over 255 comment lines, the import keeps the first 255" {
		t.Errorf("comments: %q", got)
	}
	if r := outRecord(t, out); r[104] != 255 || !bytes.Contains(out, []byte("\x1aCOMNT"+strings.Repeat(pad("c", 64), 255)+"SAUCE00")) {
		t.Errorf("comments: %d lines", r[104])
	}
}

// TestImportChange checks the compare of the grids: two regional indicators
// in two cells join into one flag grapheme in the flat form.
func TestImportChange(t *testing.T) {
	out, ds := Import("a.ans", withSauce("x\U0001F1E6\x1b[0m\U0001F1E7", 8, 1, 0))
	want := `a.ans:1:2: the flat form changes the cell: "🇦" 7/0 becomes "🇦🇧" 7/0`
	if got := diagText(ds); out != nil || got != want {
		t.Errorf("output %q, diagnostics %q; want %q", out, got, want)
	}
}

// TestImportErrors checks that an error of the play gives no output.
func TestImportErrors(t *testing.T) {
	out, ds := Import("a.ans", withSauce("a\x1b[5Zb", 4, 1, 0))
	if got := diagText(ds); out != nil || got != `a.ans:1:2: unknown sequence "ESC[5Z"` {
		t.Errorf("output %q, diagnostics %q", out, got)
	}
}

// TestImportSize checks an output over the file limit from an input under it:
// each style change becomes "ESC [ 0 ; ... m" with the RGB background.
func TestImportSize(t *testing.T) {
	data := "\x1b[48;2;100;100;100m" + strings.Repeat("\x1b[31m\xdb\x1b[32m\xdb", 20000)
	in := cp437File(data, 200, 200, 0)
	out, ds := Import("a.ans", in)
	if got := diagText(ds); len(in) > screen.MaxFileBytes || out != nil || got != "a.ans:1:1: the flat form is over 1048576 bytes" {
		t.Errorf("input %d bytes, output %d bytes, diagnostics %q", len(in), len(out), got)
	}
}

// TestImportUTF8ans checks that a file of Moebius "Export As UTF-8" is an error (section 3).
func TestImportUTF8ans(t *testing.T) {
	for _, name := range []string{"a.utf8ans", "a.UTF8ANS"} {
		out, ds := Import(name, []byte("ab"))
		want := name + `:1:1: the import does not read Moebius "Export As UTF-8", import the file of Save`
		if got := diagText(ds); out != nil || got != want {
			t.Errorf("output %q, diagnostics %q", out, got)
		}
	}
}
