// Package art loads .ans files into the art model of internal/screen.
// Specification: docs/specs/screen-format.md, section 8.
package art

import (
	"bytes"
	"io/fs"

	"newbbs/internal/screen"
)

// MaxRows is the row limit of the art grid (section 10).
const MaxRows = 1000

// MaxPlayCells is the limit of written and erased cells of the play: 16 grids
// of 255 columns and 1000 rows (section 10).
const MaxPlayCells = 16 * screen.MaxWidth * MaxRows

// defaultWidth is the art width without a SAUCE width (section 8.2).
const defaultWidth = 80

// Load reads the file name of fsys and parses it. name is also the file in the diagnostics.
// The art is nil when the diagnostics have an error.
func Load(fsys fs.FS, name string) (*screen.Art, screen.Diags) {
	data, err := screen.ReadFile(fsys, name)
	if err != nil {
		var ds screen.Diags
		ds.Errorf(name, 1, 1, "%v", err)
		return nil, ds
	}
	return Parse(name, data)
}

// Parse loads the contents of an .ans file (section 8).
// The art is nil when the diagnostics have an error.
func Parse(name string, file []byte) (*screen.Art, screen.Diags) {
	d := &screen.FileDiags{File: name}
	if len(file) > screen.MaxFileBytes {
		d.Errorf(1, 1, "file is over %d bytes", screen.MaxFileBytes)
		return nil, d.Diags
	}
	data, rec := splitSauce(file)
	hasBOM := bytes.HasPrefix(data, screen.BOM)
	data = bytes.TrimPrefix(data, screen.BOM)

	width, minHeight, ice := defaultWidth, 0, false
	if rec != nil && rec.dataType == 1 && rec.fileType == 1 { // a character SAUCE
		if rec.tInfo1 > 0 {
			width = rec.tInfo1
		}
		minHeight = rec.tInfo2
		ice = rec.tFlags&1 != 0
	}
	if width > screen.MaxWidth {
		d.Errorf(1, 1, "SAUCE width %d is over %d", width, screen.MaxWidth)
	}
	if minHeight > MaxRows {
		d.Errorf(1, 1, "SAUCE height %d is over %d rows", minHeight, MaxRows)
	}
	if d.Diags.HasErrors() {
		return nil, d.Diags
	}

	// The Moebius form (section 8.2, step 6) plays at the largest width
	// and takes the art width from the rows.
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
	art := &screen.Art{Width: width, Height: max(p.height, minHeight)}
	art.Rows = p.grid(art.Height, width)
	art.Fields = findFields(d, art.Rows)
	if d.Diags.HasErrors() {
		return nil, d.Diags
	}
	return art, d.Diags
}
