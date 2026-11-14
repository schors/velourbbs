package art

import (
	"bytes"

	"newbbs/internal/screen"
)

// The start style and the cell that the data does not write (section 8.3).
var (
	startStyle = screen.Style{Fg: screen.IndexColour(7), Bg: screen.IndexColour(0)}
	blankCell  = screen.Cell{Grapheme: " ", Style: startStyle}
)

// reader reads the data of an .ans file in the flat form into a grid (section 8.3).
type reader struct {
	d     *screen.FileDiags
	width int

	rows    []screen.Row // the ended rows; a nil row has no writes
	row     screen.Row   // the current row; nil before its first write
	n, col  int          // the 1-based row number and the column of the next grapheme
	used    bool         // the current row has a byte
	over    bool         // the current row is over the width
	style   screen.Style // the SGR state
	stopped bool
}

func newReader(d *screen.FileDiags, width int) *reader {
	return &reader{d: d, width: width, n: 1, col: 1, style: startStyle}
}

// read reads data. It returns false when the read stopped at an error.
func (r *reader) read(data []byte) bool {
	for i := 0; i < len(data) && !r.stopped; {
		if r.n > MaxRows {
			r.stop("art is over %d rows", MaxRows)
			break
		}
		r.used = true
		switch c := data[i]; {
		case c == '\n':
			r.endRow()
			i++
		case c == 0x1b:
			i = r.escape(data, i)
		case isC0(c):
			r.errorf("%v", screen.CheckGrapheme(string(rune(c)), 1))
			i++
		default:
			j := i + 1
			for j < len(data) && !isC0(data[j]) {
				j++
			}
			r.text(data[i:j])
			i = j
		}
	}
	if r.used && !r.stopped {
		r.endRow() // the last row has no LF
	}
	return !r.stopped
}

func isC0(c byte) bool { return c < 0x20 || c == 0x7f }

func (r *reader) errorf(format string, args ...any) {
	r.d.Errorf(r.n, r.col, format, args...)
}

// stop adds an error and stops the read.
func (r *reader) stop(format string, args ...any) {
	r.errorf(format, args...)
	r.stopped = true
}

func (r *reader) endRow() {
	r.rows = append(r.rows, r.row)
	r.row, r.n, r.col, r.used, r.over = nil, r.n+1, 1, false, false
}

// text reads a run of bytes without C0 controls.
func (r *reader) text(run []byte) {
	valid := run[:screen.ValidPrefix(run)]
	for g, w := range screen.Graphemes(string(valid)) {
		if err := screen.CheckGrapheme(g, w); err != nil {
			r.errorf("%v", err)
			continue
		}
		r.put(g, w)
	}
	if len(valid) < len(run) {
		r.d.Diags = nil // invalid UTF-8 is the only error of the file (section 8.2, step 5)
		r.stop("invalid UTF-8")
	}
}

// put writes the grapheme g of the width w at the next column of the row.
// A grapheme past the art width is one error for the row. After it, the row
// counts the columns for the positions and writes nothing.
func (r *reader) put(g string, w int) {
	if !r.over && r.col+w-1 > r.width {
		r.errorf("row is over %d columns", r.width)
		r.over = true
	}
	if r.over {
		r.col += w
		return
	}
	if r.row == nil {
		r.row = newRow(r.width)
	}
	r.row[r.col-1] = screen.Cell{Grapheme: g, Style: r.style}
	if w == 2 {
		r.row[r.col] = screen.Cell{Style: r.style}
	}
	r.col += w
}

func newRow(width int) screen.Row {
	row := make(screen.Row, width)
	for i := range row {
		row[i] = blankCell
	}
	return row
}

// grid gives the rows of the art. The rows without writes share one blank row.
func (r *reader) grid(height int) []screen.Row {
	rows := make([]screen.Row, height)
	var blankRow screen.Row
	for i := range rows {
		if i < len(r.rows) && r.rows[i] != nil {
			rows[i] = r.rows[i]
			continue
		}
		if blankRow == nil {
			blankRow = newRow(r.width)
		}
		rows[i] = blankRow
	}
	return rows
}

// escape reads the ESC at data[i] and gives the index after the sequence.
// The only sequence is SGR.
func (r *reader) escape(data []byte, i int) int {
	if i+1 == len(data) || data[i+1] != '[' {
		r.errorf(`ESC without "["`)
		return i + 1 // the byte after ESC reads
	}
	seq, next, ok := ScanCSI(data, i+2)
	if !ok {
		r.errorf("incomplete sequence %s", seq.Text)
		return next
	}
	ps, ok := AppendParams(nil, seq.Params)
	if seq.Final != 'm' || len(seq.Inter) > 0 || !ok {
		r.errorf("unknown sequence %s", seq.Text)
		return next
	}
	r.sgr(ps, bytes.Split(seq.Params, []byte{';'}), seq.Text)
	return next
}

// sgr does the SGR parameters ps (section 8.3). raw holds the digits of each parameter
// for the messages. After an error, the rest of the sequence does nothing.
func (r *reader) sgr(ps []int, raw [][]byte, text string) {
	s := &r.style
	for k := 0; k < len(ps); k++ {
		switch v := ps[k]; {
		case v == 0:
			*s = startStyle
		case v == 4:
			s.Attr |= screen.Underline
		case v == 24:
			s.Attr &^= screen.Underline
		case v == 5:
			s.Attr |= screen.Blink
		case v == 25:
			s.Attr &^= screen.Blink
		case 30 <= v && v <= 37:
			s.Fg = screen.IndexColour(uint8(v - 30))
		case 90 <= v && v <= 97:
			s.Fg = screen.IndexColour(uint8(v - 90 + 8))
		case v == 39:
			s.Fg = screen.IndexColour(7)
		case 40 <= v && v <= 47:
			s.Bg = screen.IndexColour(uint8(v - 40))
		case 100 <= v && v <= 107:
			s.Bg = screen.IndexColour(uint8(v - 100 + 8))
		case v == 49:
			s.Bg = screen.IndexColour(0)
		case v == 38 || v == 48:
			c, n, ok := ExtColour(ps[k+1:])
			if !ok {
				r.errorf("invalid colour in %s", text)
				return
			}
			if v == 38 {
				s.Fg = c
			} else {
				s.Bg = c
			}
			k += n
		default:
			r.errorf("unknown SGR parameter %s in %s", raw[k], text)
			return
		}
	}
}
