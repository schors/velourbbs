// Package artimport plays the ANSI art of the art editors into a grid of cells.
// Specification: docs/specs/art-import.md.
package artimport

import (
	"bytes"
	"unicode/utf16"
	"unicode/utf8"

	"newbbs/internal/art"
	"newbbs/internal/screen"
)

// MaxPlayCells is the limit of written and erased cells of the play: 16 grids
// of 255 columns and 1000 rows.
const MaxPlayCells = 16 * screen.MaxWidth * art.MaxRows

// defaultWidth is the art width without a SAUCE width.
const defaultWidth = 80

// comntLine is the length of a COMNT line in characters.
const comntLine = 64

// result is the grid of the play and the SAUCE record of the input.
type result struct {
	art   *screen.Art // the grid; the art has no fields
	sauce *art.Sauce  // nil when the input has no SAUCE record
}

// read reads the input file (section 4) and plays its data into a grid (section 5).
// The result is nil when the diagnostics have an error.
func read(name string, file []byte) (*result, screen.Diags) {
	d := &screen.FileDiags{File: name}
	if len(file) > screen.MaxFileBytes {
		d.Errorf(1, 1, "file is over %d bytes", screen.MaxFileBytes)
		return nil, d.Diags
	}
	data, rec := art.SplitSauce(file)
	data, rest, _ := bytes.Cut(data, []byte{0x1a})
	if rec != nil && rec.Lines == nil && bytes.HasPrefix(rest, []byte("COMNT")) {
		rec.Lines, rest = moebiusLines(rest[len("COMNT"):]), nil
	}
	width, minHeight, ice := defaultWidth, 0, false
	if rec != nil && rec.Comments > 0 && rec.Lines == nil {
		d.Errorf(1, 1, "SAUCE has %d comment lines, but no COMNT block", rec.Comments)
	}
	if rec != nil && rec.DataType == 1 && rec.FileType == 1 { // a character SAUCE
		if rec.TInfo1 > 0 {
			width = rec.TInfo1
		}
		minHeight = rec.TInfo2
		ice = rec.TFlags&1 != 0
	}
	if width > screen.MaxWidth {
		d.Errorf(1, 1, "SAUCE width %d is over %d", width, screen.MaxWidth)
	}
	if minHeight > art.MaxRows {
		d.Errorf(1, 1, "SAUCE height %d is over %d rows", minHeight, art.MaxRows)
	}
	if d.Diags.HasErrors() {
		return nil, d.Diags
	}

	isUTF8 := bytes.HasPrefix(data, screen.BOM)
	if isUTF8 {
		data = data[len(screen.BOM):]
	} else {
		data = decodeCP437(data)
	}
	p := newPlayer(d, width, ice, isUTF8)
	if !p.play(data) {
		return nil, d.Diags
	}
	if len(rest) > 0 {
		p.errorf("data after 0x1A")
	}
	a := &screen.Art{Width: width, Height: max(p.frame, minHeight)}
	a.Rows = p.grid(a.Height)
	if d.Diags.HasErrors() {
		return nil, d.Diags
	}
	return &result{art: a, sauce: rec}, d.Diags
}

// moebiusLines splits the text of a COMNT block of Moebius into lines of 64 characters
// (section 4, step 1). Moebius counts the characters in UTF-16 code units.
// A block of invalid UTF-8 splits into lines of 64 bytes.
func moebiusLines(text []byte) [][]byte {
	var lines [][]byte
	for len(text) > 0 {
		n, units := 0, 0
		for n < len(text) && units < comntLine {
			r, size := utf8.DecodeRune(text[n:])
			if r == utf8.RuneError && size == 1 {
				units++ // a byte of invalid UTF-8 counts as one character
			} else {
				units += utf16.RuneLen(r)
			}
			n += size
		}
		lines = append(lines, text[:n])
		text = text[n:]
	}
	return lines
}
