package artimport

import (
	"bytes"
	"unicode/utf8"

	"github.com/schors/velourbbs/internal/art"
	"github.com/schors/velourbbs/internal/screen"
)

// The start style of the play and the style of a cell that the play does not write.
var (
	startPen  = pen{fg: screen.IndexColour(7), bg: screen.IndexColour(0)}
	blankCell = screen.Cell{Grapheme: " ", Style: screen.Style{Fg: screen.IndexColour(7), Bg: screen.IndexColour(0)}}
)

// pen is the current SGR state of the play.
type pen struct {
	fg, bg                          screen.Colour
	bold, underline, blink, reverse bool
}

// style gives the style of a written cell by the rules of section 5.
func (p pen) style(ice bool) screen.Style {
	s := screen.Style{Fg: p.fg, Bg: p.bg}
	if p.bold && s.Fg.Kind == screen.Index && s.Fg.Value < 8 {
		s.Fg.Value += 8
	}
	if p.blink {
		if !ice {
			s.Attr |= screen.Blink
		} else if s.Bg.Kind == screen.Index && s.Bg.Value < 8 {
			s.Bg.Value += 8
		}
	}
	if p.underline {
		s.Attr |= screen.Underline
	}
	if p.reverse {
		s.Fg, s.Bg = s.Bg, s.Fg
	}
	return s
}

// player plays the data of an .ans file into a grid (section 5).
type player struct {
	d         *screen.FileDiags
	width     int
	ice       bool
	wrapReset bool // UTF-8 input: a wrap resets the style (section 5)

	rows     []screen.Row // a nil row has no writes
	drawn    [][]bool     // the cells that a grapheme wrote, for the overwrite warning (section 6)
	height   int          // the last written row
	frame    int          // the last row written after the last CSI 2 J: the grid height
	cells    int          // the written and erased cells (section 8)
	animated bool         // the play gave the overwrite warning

	row, col           int // 1-based; LF can move row past art.MaxRows
	pending            bool
	wrap               bool
	pen                pen
	last               string // the last grapheme, for CSI n b
	lastWidth          int
	savedRow, savedCol int
	stopped            bool
	params             []int
}

func newPlayer(d *screen.FileDiags, width int, ice, wrapReset bool) *player {
	return &player{
		d: d, width: width, ice: ice, wrapReset: wrapReset,
		row: 1, col: 1, wrap: true, pen: startPen, savedRow: 1, savedCol: 1,
	}
}

// play plays data. It returns false when the play stopped at an error.
func (p *player) play(data []byte) bool {
	for i := 0; i < len(data) && !p.stopped; {
		switch c := data[i]; {
		case c == '\n':
			p.lf()
			i++
		case c == '\r':
			p.doWrap()
			p.col = 1
			i++
		case c == 0x1b:
			i = p.escape(data, i)
		case isC0(c):
			p.errorf("%v", screen.CheckGrapheme(string(rune(c)), 1))
			i++
		default:
			j := i + 1
			for j < len(data) && !isC0(data[j]) {
				j++
			}
			p.text(data[i:j])
			i = j
		}
	}
	return !p.stopped
}

func isC0(c byte) bool { return c < 0x20 || c == 0x7f }

// pos gives the grid position of the next grapheme.
func (p *player) pos() (int, int) {
	if p.pending {
		return p.row + 1, 1
	}
	return p.row, p.col
}

func (p *player) errorf(format string, args ...any) {
	row, col := p.pos()
	p.d.Errorf(row, col, format, args...)
}

// stop adds an error and stops the play.
func (p *player) stop(format string, args ...any) {
	p.errorf(format, args...)
	p.stopped = true
}

// spend counts n written or erased cells. Over MaxPlayCells it stops the play.
func (p *player) spend(n int) bool {
	p.cells += n
	if p.cells > MaxPlayCells {
		p.stop("art play is over %d written and erased cells", MaxPlayCells)
		return false
	}
	return true
}

// text plays a run of bytes without C0 controls.
func (p *player) text(run []byte) {
	valid := run[:screen.ValidPrefix(run)]
	for g, w := range screen.Graphemes(string(valid)) {
		if p.stopped {
			return
		}
		if err := screen.CheckGrapheme(g, w); err != nil {
			p.errorf("%v", err)
			continue
		}
		p.put(g, w)
	}
	if len(valid) < len(run) && !p.stopped {
		p.d.Diags = nil // invalid UTF-8 is the only error of the file (section 4, step 4)
		p.stop("invalid UTF-8")
	}
}

// put writes the grapheme g of the width w at the cursor.
func (p *player) put(g string, w int) {
	if w > p.width {
		p.errorf("wide grapheme in art of width %d", p.width)
		return
	}
	if p.pending || w == 2 && p.col == p.width && p.wrap {
		p.wrapRow()
	}
	if w == 2 && p.col == p.width {
		p.col-- // wrap off: the wide grapheme takes the last two cells
	}
	if p.row > art.MaxRows {
		p.stop("art is over %d rows", art.MaxRows)
		return
	}
	if !p.spend(w) {
		return
	}
	style := p.pen.style(p.ice)
	r := p.open(p.col, p.col+w-1)
	r[p.col-1] = screen.Cell{Grapheme: g, Style: style}
	if w == 2 {
		r[p.col] = screen.Cell{Style: style}
	}
	fillFlags(p.drawn[p.row-1][p.col-1 : p.col-1+w])
	p.last, p.lastWidth = g, w
	if p.col+w > p.width {
		p.col, p.pending = p.width, p.wrap
	} else {
		p.col += w
	}
}

// open prepares the cells a..b of the cursor row for a write. It allocates the row.
// When the write cuts a wide grapheme, the other half becomes a space in its style.
func (p *player) open(a, b int) screen.Row {
	r := p.rowAt(p.row)
	p.overwrite(p.row, a, b)
	if a > 1 && r[a-1].Grapheme == "" {
		r[a-2].Grapheme = " "
	}
	if b < len(r) && r[b].Grapheme == "" {
		r[b].Grapheme = " "
	}
	p.height = max(p.height, p.row)
	p.frame = max(p.frame, p.row)
	return r
}

// overwrite gives the warning of section 6 at the first write to a cell of a..b
// in the row n that a grapheme wrote. The warning has the position of the cell.
func (p *player) overwrite(n, a, b int) {
	if p.animated {
		return
	}
	for i, drawn := range p.drawn[n-1][a-1 : b] {
		if drawn {
			p.animated = true
			p.d.Warnf(n, a+i, "art overwrites a cell, the import keeps the last frame")
			return
		}
	}
}

// rowAt gives the row n. It allocates the row at the first write.
func (p *player) rowAt(n int) screen.Row {
	for len(p.rows) < n {
		p.rows = append(p.rows, nil)
		p.drawn = append(p.drawn, nil)
	}
	if p.rows[n-1] == nil {
		p.rows[n-1] = newRow(p.width, blankCell)
		p.drawn[n-1] = make([]bool, p.width)
	}
	return p.rows[n-1]
}

func fillFlags(f []bool) {
	for i := range f {
		f[i] = true
	}
}

func newRow(width int, c screen.Cell) screen.Row {
	r := make(screen.Row, width)
	fillCells(r, c)
	return r
}

func fillCells(r screen.Row, c screen.Cell) {
	for i := range r {
		r[i] = c
	}
}

// grid gives the rows 1..height of the art. The rows without writes share one blank row.
func (p *player) grid(height int) []screen.Row {
	rows := make([]screen.Row, height)
	var blankRow screen.Row
	for i := range rows {
		if i < len(p.rows) && p.rows[i] != nil {
			rows[i] = p.rows[i]
			continue
		}
		if blankRow == nil {
			blankRow = newRow(p.width, blankCell)
		}
		rows[i] = blankRow
	}
	return rows
}

func (p *player) lf() {
	p.row, p.col, p.pending = p.row+1, 1, false
}

// doWrap does a pending wrap.
func (p *player) doWrap() {
	if p.pending {
		p.wrapRow()
	}
}

// wrapRow moves the cursor to column 1 of the next row. In UTF-8 input it resets the style.
func (p *player) wrapRow() {
	p.row, p.col, p.pending = p.row+1, 1, false
	if p.wrapReset {
		p.pen = startPen
	}
}

// moveTo moves the cursor inside the grid width and the row limit.
func (p *player) moveTo(row, col int) {
	p.row = min(max(row, 1), art.MaxRows)
	p.moveCol(col)
}

// moveCol moves the cursor inside the grid width. The row stays, also past the row limit.
func (p *player) moveCol(col int) {
	p.col = min(max(col, 1), p.width)
}

// restore moves the cursor to the saved position. The row can be past the row limit.
func (p *player) restore() {
	p.doWrap()
	p.row, p.col = p.savedRow, p.savedCol
}

// escape plays the ESC at data[i] and gives the index after it.
func (p *player) escape(data []byte, i int) int {
	if i+1 == len(data) {
		p.errorf(`incomplete sequence "ESC"`)
		return i + 1
	}
	switch data[i+1] {
	case '[':
		seq, next, ok := art.ScanCSI(data, i+2)
		if !ok {
			p.errorf("incomplete sequence %s", seq.Text)
		} else {
			p.control(seq)
		}
		return next // after an incomplete sequence, the byte at next plays
	case '7':
		p.doWrap()
		p.savedRow, p.savedCol = p.row, p.col
		return i + 2
	case '8':
		p.restore()
		return i + 2
	}
	next := data[i+1 : i+2]
	if r, n := utf8.DecodeRune(data[i+1:]); r != utf8.RuneError {
		next = data[i+1 : i+1+n]
	}
	p.errorf("unknown sequence %q", "ESC"+string(next))
	return i + 1 // the byte after ESC plays
}

// control does the effect of a CSI sequence (section 5).
func (p *player) control(seq art.CSI) {
	params := seq.Params
	var marker byte
	if len(params) > 0 && params[0] >= 0x3c {
		marker, params = params[0], params[1:]
	}
	ps, ok := art.AppendParams(p.params[:0], params)
	p.params = ps
	switch inter := string(seq.Inter); {
	case !ok:
	case marker == 0 && inter == "":
		if p.plain(seq.Final, ps, len(params) == 0, seq) {
			return
		}
	case marker == '?' && inter == "" && (seq.Final == 'h' || seq.Final == 'l'):
		if len(ps) == 1 && (ps[0] == 33 || ps[0] == 7) {
			if ps[0] == 33 {
				p.ice = seq.Final == 'h'
			} else {
				p.wrap = seq.Final == 'h'
			}
			return
		}
	case marker == 0 && inter == " " && seq.Final == 'D':
		if len(ps) == 2 && ps[0] == 0 {
			return // the font page of icy_draw: no effect
		}
	}
	p.errorf("unknown sequence %s", seq.Text)
}

// plain does a CSI sequence without a marker and intermediate bytes. empty reports that
// the sequence has no parameter bytes. It reports false for a sequence that is not
// in the table of section 5.
func (p *player) plain(final byte, ps []int, empty bool, seq art.CSI) bool {
	n := max(ps[0], 1)
	switch final {
	case 'A', 'B', 'C', 'D', 'b', 'J', 'K':
		if len(ps) > 1 {
			return false
		}
	case 'H', 'f':
		if len(ps) > 2 {
			return false
		}
	case 's', 'u':
		if !empty {
			return false
		}
	}
	switch final {
	case 'A':
		p.doWrap()
		p.moveTo(p.row-n, p.col)
	case 'B':
		p.doWrap()
		p.moveTo(p.row+n, p.col)
	case 'C':
		p.doWrap()
		p.moveCol(p.col + n)
	case 'D':
		p.doWrap()
		p.moveCol(p.col - n)
	case 'H', 'f':
		p.doWrap()
		p.moveTo(n, max(arg(ps, 1), 1))
	case 'J':
		if ps[0] != 2 {
			return false
		}
		p.clear()
	case 'K':
		if ps[0] > 2 {
			return false
		}
		p.erase(ps[0])
	case 's':
		p.doWrap()
		p.savedRow, p.savedCol = p.row, p.col
	case 'u':
		p.restore()
	case 'b':
		p.repeat(n)
	case 't':
		p.rgb(ps, seq.Text)
	case 'm':
		p.sgr(ps, seq)
	default:
		return false
	}
	return true
}

// rgb does CSI 0;r;g;b t and CSI 1;r;g;b t of PabloDraw: the background and the foreground.
func (p *player) rgb(ps []int, text string) {
	if ps[0] > 1 {
		p.errorf("unknown sequence %s", text)
		return
	}
	if len(ps) != 4 || ps[1] > 255 || ps[2] > 255 || ps[3] > 255 {
		p.errorf("invalid colour in %s", text)
		return
	}
	c := screen.RGBColour(uint8(ps[1]), uint8(ps[2]), uint8(ps[3]))
	if ps[0] == 1 {
		p.pen.fg = c
	} else {
		p.pen.bg = c
	}
}

func arg(ps []int, i int) int {
	if i < len(ps) {
		return ps[i]
	}
	return 0
}

// eraseStyle gives the style of an erased cell: the foreground 7,
// the background of a written cell, no attributes.
func (p *player) eraseStyle() screen.Style {
	return screen.Style{Fg: screen.IndexColour(7), Bg: p.pen.style(p.ice).Bg}
}

// erase does CSI n K: 0 to the end of the row, 1 to the start, 2 the whole row.
func (p *player) erase(mode int) {
	if p.row > art.MaxRows {
		p.stop("art is over %d rows", art.MaxRows)
		return
	}
	a, b := 1, p.width
	switch mode {
	case 0:
		a = p.col
	case 1:
		b = p.col
	}
	if !p.spend(b - a + 1) {
		return
	}
	fillCells(p.open(a, b)[a-1:b], screen.Cell{Grapheme: " ", Style: p.eraseStyle()})
}

// clear does CSI 2 J: it erases the rows 1 to the height and moves the cursor to row 1, column 1.
// The grid height starts again from 0.
func (p *player) clear() {
	if !p.spend(p.height * p.width) {
		return
	}
	c := screen.Cell{Grapheme: " ", Style: p.eraseStyle()}
	for n := 1; n <= p.height; n++ {
		fillCells(p.rowAt(n), c)
		p.overwrite(n, 1, p.width)
	}
	p.row, p.col, p.pending, p.frame = 1, 1, false, 0
}

// repeat does CSI n b. It stops at the row limit.
func (p *player) repeat(n int) {
	if p.last == "" {
		return
	}
	for ; n > 0 && !p.stopped; n-- {
		next := p.row
		if p.pending || p.lastWidth == 2 && p.col == p.width && p.wrap {
			next++
		}
		if next > art.MaxRows {
			return
		}
		p.put(p.last, p.lastWidth)
	}
}

// sgr does CSI ... m. An unknown parameter is an error; the rest of the sequence plays.
// After an invalid 38 or 48, the rest does nothing.
func (p *player) sgr(ps []int, seq art.CSI) {
	var raw [][]byte // the digits of each parameter, for the messages
	for k := 0; k < len(ps); k++ {
		switch v := ps[k]; {
		case v == 0:
			p.pen = startPen
		case v == 1 || v == 22:
			p.pen.bold = v == 1
		case v == 4 || v == 24:
			p.pen.underline = v == 4
		case v == 5 || v == 25:
			p.pen.blink = v == 5
		case v == 7 || v == 27:
			p.pen.reverse = v == 7
		case 30 <= v && v <= 37:
			p.pen.fg = screen.IndexColour(uint8(v - 30))
		case 90 <= v && v <= 97:
			p.pen.fg = screen.IndexColour(uint8(v - 90 + 8))
		case v == 39:
			p.pen.fg = screen.IndexColour(7)
		case 40 <= v && v <= 47:
			p.pen.bg = screen.IndexColour(uint8(v - 40))
		case 100 <= v && v <= 107:
			p.pen.bg = screen.IndexColour(uint8(v - 100 + 8))
		case v == 49:
			p.pen.bg = screen.IndexColour(0)
		case v == 38 || v == 48:
			c, n, ok := art.ExtColour(ps[k+1:])
			if !ok {
				p.errorf("invalid colour in %s", seq.Text)
				return
			}
			if v == 38 {
				p.pen.fg = c
			} else {
				p.pen.bg = c
			}
			k += n
		default:
			if raw == nil {
				raw = bytes.Split(seq.Params, []byte{';'})
			}
			p.errorf("unknown SGR parameter %s in %s", raw[k], seq.Text)
		}
	}
}
