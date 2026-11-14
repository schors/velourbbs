// Package artimport plays the ANSI art of the art editors into a grid of cells.
// Specification: docs/specs/art-import.md.
package artimport

import (
	"bytes"

	"newbbs/internal/art"
	"newbbs/internal/screen"
)

// MaxPlayCells is the limit of written and erased cells of the play: 16 grids
// of 255 columns and 1000 rows.
const MaxPlayCells = 16 * screen.MaxWidth * art.MaxRows

// defaultWidth is the art width without a SAUCE width.
const defaultWidth = 80

// read reads the file and plays its data into a grid. The art has no fields.
// The art is nil when the diagnostics have an error.
func read(name string, file []byte) (*screen.Art, screen.Diags) {
	d := &screen.FileDiags{File: name}
	if len(file) > screen.MaxFileBytes {
		d.Errorf(1, 1, "file is over %d bytes", screen.MaxFileBytes)
		return nil, d.Diags
	}
	data, rec := art.SplitSauce(file)
	data = bytes.TrimSuffix(data, []byte{0x1a})
	hasBOM := bytes.HasPrefix(data, screen.BOM)
	data = bytes.TrimPrefix(data, screen.BOM)

	width, minHeight, ice := defaultWidth, 0, false
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

	// The Moebius form plays at the largest width and takes the art width from the rows.
	moebius := !hasBOM && rec == nil
	p := newPlayer(d, width, ice, moebius)
	if moebius {
		p.width = screen.MaxWidth
	}
	if !p.play(data) {
		return nil, d.Diags
	}
	if moebius {
		width = max(defaultWidth, p.maxCol)
	}
	a := &screen.Art{Width: width, Height: max(p.height, minHeight)}
	a.Rows = p.grid(a.Height, width)
	if d.Diags.HasErrors() {
		return nil, d.Diags
	}
	return a, d.Diags
}
