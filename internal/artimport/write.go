package artimport

import (
	"bytes"
	"encoding/binary"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"newbbs/internal/art"
	"newbbs/internal/screen"
)

// maxComments is the largest count of COMNT lines: the SAUCE field has one byte.
const maxComments = 255

// The text fields of the output SAUCE record: the name, the offset and the size.
// TInfoS is a ZString and is padded with NUL; the other fields are padded with spaces.
var sauceFields = []struct {
	name       string
	at, size   int
	get        func(*art.Sauce) []byte
	nulPadding bool
}{
	{"title", 7, 35, func(s *art.Sauce) []byte { return s.Title }, false},
	{"author", 42, 20, func(s *art.Sauce) []byte { return s.Author }, false},
	{"group", 62, 20, func(s *art.Sauce) []byte { return s.Group }, false},
	{"date", 82, 8, func(s *art.Sauce) []byte { return s.Date }, false},
	{"font", 106, 22, func(s *art.Sauce) []byte { return s.TInfoS }, true},
}

// Import converts the input file to the flat form (section 9).
// name is the file in the diagnostics. The output is nil when the diagnostics have an error.
func Import(name string, file []byte) ([]byte, screen.Diags) {
	if strings.EqualFold(filepath.Ext(name), ".utf8ans") {
		var ds screen.Diags
		ds.Errorf(name, 1, 1, `the import does not read Moebius "Export As UTF-8", import the file of Save`)
		return nil, ds
	}
	res, ds := read(name, file)
	if res == nil {
		return nil, ds
	}
	d := &screen.FileDiags{File: name, Diags: ds}
	out := encode(d, res)
	if len(out) > screen.MaxFileBytes {
		d.Errorf(1, 1, "the flat form is over %d bytes", screen.MaxFileBytes)
		return nil, d.Diags
	}
	back, bds := art.ReadGrid(name, out)
	if back == nil {
		d.Diags = append(d.Diags, bds...)
		return nil, d.Diags
	}
	compare(d, res.art, back)
	if d.Diags.HasErrors() {
		return nil, d.Diags
	}
	return out, d.Diags
}

// encode writes the grid of res in the flat form with its SAUCE record (section 9).
func encode(d *screen.FileDiags, res *result) []byte {
	a := res.art
	out := append([]byte(nil), screen.BOM...)
	cur := blankCell.Style
	for _, row := range a.Rows {
		end := len(row)
		for end > 0 && row[end-1] == blankCell {
			end--
		}
		for _, c := range row[:end] {
			if c.Grapheme == "" {
				continue // the second cell of a wide grapheme
			}
			if c.Style != cur {
				out = appendSGR(out, c.Style)
				cur = c.Style
			}
			out = append(out, c.Grapheme...)
		}
		out = append(out, '\n')
	}
	size := len(out)
	out = append(out, 0x1a)
	lines := comments(d, res.sauce)
	if len(lines) > 0 {
		out = append(out, "COMNT"...)
		for _, l := range lines {
			out = append(out, l...)
		}
	}
	return append(out, sauceRecord(d, res, size, len(lines))...)
}

// appendSGR appends "ESC [ 0 ; <params> m" for the style s (section 9, step 3).
func appendSGR(out []byte, s screen.Style) []byte {
	out = append(out, "\x1b[0"...)
	if s.Attr&screen.Underline != 0 {
		out = append(out, ";4"...)
	}
	if s.Attr&screen.Blink != 0 {
		out = append(out, ";5"...)
	}
	if s.Fg != screen.IndexColour(7) {
		out = appendColour(out, s.Fg, 30, 90, 38)
	}
	if s.Bg != screen.IndexColour(0) {
		out = appendColour(out, s.Bg, 40, 100, 48)
	}
	return append(out, 'm')
}

// appendColour appends the SGR parameters of the colour c: base+n for the index n 0–7,
// bright+n-8 for 8–15, "ext;5;n" for 16–255 and "ext;2;r;g;b" for RGB.
func appendColour(out []byte, c screen.Colour, base, bright, ext int) []byte {
	out = append(out, ';')
	switch v := int(c.Value); {
	case c.Kind == screen.RGB:
		out = strconv.AppendInt(out, int64(ext), 10)
		out = append(out, ";2"...)
		for _, shift := range []int{16, 8, 0} {
			out = append(out, ';')
			out = strconv.AppendInt(out, int64(v>>shift&0xff), 10)
		}
	case v < 8:
		out = strconv.AppendInt(out, int64(base+v), 10)
	case v < 16:
		out = strconv.AppendInt(out, int64(bright+v-8), 10)
	default:
		out = strconv.AppendInt(out, int64(ext), 10)
		out = append(out, ";5;"...)
		out = strconv.AppendInt(out, int64(v), 10)
	}
	return out
}

// comments gives the COMNT lines of the output: each line decoded and padded
// to 64 bytes, at most maxComments lines.
func comments(d *screen.FileDiags, rec *art.Sauce) [][]byte {
	if rec == nil {
		return nil
	}
	lines := rec.Lines
	if len(lines) > maxComments {
		d.Warnf(1, 1, "SAUCE has over %d comment lines, the import keeps the first %d", maxComments, maxComments)
		lines = lines[:maxComments]
	}
	out := make([][]byte, len(lines))
	cut := 0
	for i, l := range lines {
		out[i] = make([]byte, comntLine)
		if putText(out[i], decodeText(l), ' ') {
			cut++
		}
	}
	if cut > 0 {
		d.Warnf(1, 1, "SAUCE has %d comment lines over %d bytes, the import cuts them", cut, comntLine)
	}
	return out
}

// sauceRecord gives the SAUCE record of the output (section 9). size is the bytes
// of the output before the 0x1A, n is the count of COMNT lines.
func sauceRecord(d *screen.FileDiags, res *result, size, n int) []byte {
	r := make([]byte, 128)
	copy(r, "SAUCE00")
	rec := res.sauce
	char := rec != nil && rec.DataType == 1 && rec.FileType == 1
	for _, f := range sauceFields {
		pad := byte(' ')
		if f.nulPadding {
			pad = 0
		}
		var text string
		if rec != nil && (f.name != "font" || char) {
			text = decodeText(f.get(rec))
		}
		if putText(r[f.at:f.at+f.size], text, pad) {
			d.Warnf(1, 1, "SAUCE %s is over %d bytes, the import cuts it", f.name, f.size)
		}
	}
	binary.LittleEndian.PutUint32(r[90:], uint32(size))
	r[94], r[95] = 1, 1
	binary.LittleEndian.PutUint16(r[96:], uint16(res.art.Width))
	binary.LittleEndian.PutUint16(r[98:], uint16(res.art.Height))
	r[104] = byte(n)
	if char {
		r[105] = rec.TFlags & 0x1e // bits 1–4; bit 0 (iCE colours) is 0
	}
	return r
}

// decodeText reads a SAUCE text field: without trailing spaces and NUL bytes,
// as UTF-8 when it is valid UTF-8, else by the CP437 table (section 9).
func decodeText(raw []byte) string {
	raw = bytes.TrimRight(raw, " \x00")
	if utf8.Valid(raw) {
		return string(raw)
	}
	return string(decodeCP437(raw))
}

// putText writes text to the field dst and pads it with pad. A text longer than dst
// is cut after the last whole grapheme that fits; putText reports the cut.
func putText(dst []byte, text string, pad byte) bool {
	cut := len(text) > len(dst)
	if cut {
		n := 0
		for g := range screen.Graphemes(text) {
			if n+len(g) > len(dst) {
				break
			}
			n += len(g)
		}
		text = text[:n]
	}
	n := copy(dst, text)
	for i := n; i < len(dst); i++ {
		dst[i] = pad
	}
	return cut
}

// compare checks that the grid of the output equals the grid of the play.
// The first different cell is an error at its position.
func compare(d *screen.FileDiags, play, back *screen.Art) {
	if back.Height != play.Height {
		d.Errorf(min(back.Height, play.Height)+1, 1, "the flat form has %d rows, the play has %d", back.Height, play.Height)
		return
	}
	for i, row := range play.Rows {
		for j, c := range row {
			if b := back.Rows[i][j]; b != c {
				d.Errorf(i+1, j+1, "the flat form changes the cell: %q %s becomes %q %s",
					c.Grapheme, c.Style, b.Grapheme, b.Style)
				return
			}
		}
	}
}
