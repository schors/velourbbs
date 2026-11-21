// Package art loads .ans files in the flat form into the art model of internal/screen.
// Specification: docs/specs/art-format.md.
package art

import (
	"bytes"

	"github.com/schors/velourbbs/internal/screen"
)

// MaxRows is the row limit of the art grid (section 5).
const MaxRows = 1000

// errNotFlat is the error of the steps 1, 2 and 4 of section 2.
const errNotFlat = "not in the flat form, convert it with velourbbs art import"

// Parse loads the contents of an .ans file in the flat form and resolves
// its fields with the aliases fields of the screen (bbt-format, section 4.2).
// name is the file in the diagnostics. The art is nil when the diagnostics have an error.
func Parse(name string, file []byte, fields screen.Fields) (*screen.Art, screen.Diags) {
	d := &screen.FileDiags{File: name}
	art := readGrid(d, file)
	if art == nil {
		return nil, d.Diags
	}
	art.Fields, art.Boxes = findTokens(d, art.Rows, fields)
	art.List = findList(d, art.Fields)
	checkBoxes(d, art.Boxes, art.List)
	if d.Diags.HasErrors() {
		return nil, d.Diags
	}
	return art, d.Diags
}

// ReadGrid reads the contents of an .ans file in the flat form by sections 2 and 3.
// The art has no fields. It is nil when the diagnostics have an error.
func ReadGrid(name string, file []byte) (*screen.Art, screen.Diags) {
	d := &screen.FileDiags{File: name}
	art := readGrid(d, file)
	if d.Diags.HasErrors() {
		return nil, d.Diags
	}
	return art, d.Diags
}

// readGrid reads file by sections 2 and 3. The art is nil after an error that stops the read.
func readGrid(d *screen.FileDiags, file []byte) *screen.Art {
	if len(file) > screen.MaxFileBytes {
		d.Errorf(1, 1, "file is over %d bytes", screen.MaxFileBytes)
		return nil
	}
	data, rec := SplitSauce(file)
	data = bytes.TrimSuffix(data, []byte{0x1a})
	if !isFlatSauce(rec) || !bytes.HasPrefix(data, screen.BOM) {
		d.Errorf(1, 1, errNotFlat)
		return nil
	}
	data = data[len(screen.BOM):]

	r := newReader(d, rec.TInfo1)
	if !r.read(data) {
		return nil
	}
	art := &screen.Art{Width: rec.TInfo1, Height: max(len(r.rows), rec.TInfo2)}
	art.Rows = r.grid(art.Height)
	return art
}

// isFlatSauce checks the SAUCE record by section 2, steps 1 and 2.
func isFlatSauce(rec *Sauce) bool {
	return rec != nil && rec.DataType == 1 && rec.FileType == 1 &&
		1 <= rec.TInfo1 && rec.TInfo1 <= screen.MaxWidth &&
		rec.TInfo2 <= MaxRows && rec.TFlags&1 == 0
}
