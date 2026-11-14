package artimport

import (
	"fmt"

	"newbbs/internal/art"
	"newbbs/internal/screen"
)

// maxParam is the largest value of a CSI parameter (section 5).
const maxParam = 9999

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
	d       *screen.FileDiags
	width   int
	ice     bool
	moebius bool // the Moebius form: a row over the width is an error
	skipLF  bool // the Moebius form before the first grapheme: LF does nothing

	rows   []screen.Row // a nil row has no writes
	height int          // the last written row
	maxCol int          // the last column of a written grapheme
	cells  int          // the written and erased cells (section 8)

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

func newPlayer(d *screen.FileDiags, width int, ice, moebius bool) *player {
	return &player{
		d: d, width: width, ice: ice, moebius: moebius, skipLF: moebius,
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

func (p *player) warnf(format string, args ...any) {
	row, col := p.pos()
	p.d.Warnf(row, col, format, args...)
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
		p.stop("invalid UTF-8: art must be UTF-8, CP437 art needs a conversion")
	}
}

// put writes the grapheme g of the width w at the cursor.
func (p *player) put(g string, w int) {
	p.skipLF = false
	if w > p.width {
		p.errorf("wide grapheme in art of width %d", p.width)
		return
	}
	if p.pending || w == 2 && p.col == p.width && p.wrap {
		if p.moebius {
			p.d.Errorf(p.row, p.col, "row is over %d columns", p.width)
			p.stopped = true
			return
		}
		p.row, p.col, p.pending = p.row+1, 1, false
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
	p.maxCol = max(p.maxCol, p.col+w-1)
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
	if a > 1 && r[a-1].Grapheme == "" {
		r[a-2].Grapheme = " "
	}
	if b < len(r) && r[b].Grapheme == "" {
		r[b].Grapheme = " "
	}
	p.height = max(p.height, p.row)
	return r
}

// rowAt gives the row n. It allocates the row at the first write.
func (p *player) rowAt(n int) screen.Row {
	for len(p.rows) < n {
		p.rows = append(p.rows, nil)
	}
	if p.rows[n-1] == nil {
		p.rows[n-1] = newRow(p.width, blankCell)
	}
	return p.rows[n-1]
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

// grid gives the rows of the art. The rows without writes share one blank row.
func (p *player) grid(height, width int) []screen.Row {
	rows := make([]screen.Row, height)
	var blankRow screen.Row
	for i := range rows {
		if i < len(p.rows) && p.rows[i] != nil {
			rows[i] = p.rows[i][:width:width]
			continue
		}
		if blankRow == nil {
			blankRow = newRow(width, blankCell)
		}
		rows[i] = blankRow
	}
	return rows
}

func (p *player) lf() {
	if p.skipLF {
		return
	}
	p.row, p.col, p.pending = p.row+1, 1, false
}

// doWrap does a pending wrap.
func (p *player) doWrap() {
	if p.pending {
		p.row, p.col, p.pending = p.row+1, 1, false
	}
}

// moveTo moves the cursor inside the grid width and the row limit.
func (p *player) moveTo(row, col int) {
	p.row = min(max(row, 1), art.MaxRows)
	p.col = min(max(col, 1), p.width)
}

// escape plays the ESC at data[i] and gives the index after it.
func (p *player) escape(data []byte, i int) int {
	if i+1 == len(data) {
		return i + 1
	}
	switch data[i+1] {
	case '[':
		return p.csi(data, i+2)
	case '7':
		p.doWrap()
		p.savedRow, p.savedCol = p.row, p.col
		return i + 2
	case '8':
		p.doWrap()
		p.moveTo(p.savedRow, p.savedCol)
		return i + 2
	}
	return i + 1 // the loader drops the ESC and plays the byte
}

// csi plays the CSI sequence after "ESC [" at data[i] (ECMA-48, section 5.4)
// and gives the index after it.
func (p *player) csi(data []byte, i int) int {
	j := i
	for j < len(data) && 0x30 <= data[j] && data[j] <= 0x3f {
		j++
	}
	k := j
	for k < len(data) && 0x20 <= data[k] && data[k] <= 0x2f {
		k++
	}
	if k == len(data) || data[k] < 0x40 || data[k] > 0x7e {
		p.warnf("incomplete sequence %s", seqText(data[i:k]))
		return k // the byte at k plays
	}
	p.control(data[i:j], data[j:k], data[k], seqText(data[i:k+1]))
	return k + 1
}

// seqText gives a CSI sequence for a message. It cuts a long sequence.
func seqText(b []byte) string {
	const maxLen = 24
	if len(b) > maxLen {
		return fmt.Sprintf("%q", "ESC["+string(b[:maxLen])+"...")
	}
	return fmt.Sprintf("%q", "ESC["+string(b))
}

// control does the effect of a CSI sequence (section 5).
func (p *player) control(params, inter []byte, final byte, text string) {
	var marker byte
	if len(params) > 0 && params[0] >= 0x3c {
		marker, params = params[0], params[1:]
	}
	ps, ok := p.parseParams(params)
	// Only the mode sequences take a marker, and only "?".
	mode := final == 'h' || final == 'l'
	known := len(inter) == 0 && (marker == 0 && !mode || marker == '?' && mode)
	if !ok || !known {
		p.warnf("unknown sequence %s", text)
		return
	}
	n := max(arg(ps, 0), 1)
	switch final {
	case 'A':
		p.doWrap()
		p.moveTo(p.row-n, p.col)
	case 'B':
		p.doWrap()
		p.moveTo(p.row+n, p.col)
	case 'C':
		p.doWrap()
		p.moveTo(p.row, p.col+n)
	case 'D':
		p.doWrap()
		p.moveTo(p.row, p.col-n)
	case 'H', 'f':
		p.doWrap()
		p.moveTo(n, max(arg(ps, 1), 1))
	case 'J':
		if len(ps) != 1 || ps[0] != 2 {
			p.warnf("unknown sequence %s", text)
			return
		}
		p.clear()
	case 'K':
		if len(ps) != 1 || ps[0] > 2 {
			p.warnf("unknown sequence %s", text)
			return
		}
		p.erase(ps[0])
	case 's':
		p.doWrap()
		p.savedRow, p.savedCol = p.row, p.col
	case 'u':
		p.doWrap()
		p.moveTo(p.savedRow, p.savedCol)
	case 'b':
		p.repeat(n)
	case 'h', 'l':
		for _, v := range ps {
			switch v {
			case 33:
				p.ice = final == 'h'
			case 7:
				p.wrap = final == 'h'
			default:
				p.warnf("unknown mode %d in %s", v, text)
			}
		}
	case 'm':
		p.sgr(ps, text)
	default:
		p.warnf("unknown sequence %s", text)
	}
}

// parseParams parses the parameters of a CSI sequence. A missing parameter is 0.
// A parameter over maxParam counts as maxParam. It reports false for a byte
// other than a digit or ";".
func (p *player) parseParams(b []byte) ([]int, bool) {
	ps := p.params[:0]
	v := 0
	for _, c := range b {
		switch {
		case c == ';':
			ps = append(ps, v)
			v = 0
		case '0' <= c && c <= '9':
			v = min(v*10+int(c-'0'), maxParam)
		default:
			return nil, false
		}
	}
	p.params = append(ps, v)
	return p.params, true
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
func (p *player) clear() {
	if !p.spend(p.height * p.width) {
		return
	}
	c := screen.Cell{Grapheme: " ", Style: p.eraseStyle()}
	for n := 1; n <= p.height; n++ {
		fillCells(p.rowAt(n), c)
	}
	p.row, p.col, p.pending = 1, 1, false
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

// sgr does CSI ... m.
func (p *player) sgr(ps []int, text string) {
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
			c, n, ok := extColour(ps[k+1:])
			if !ok {
				p.warnf("invalid colour in %s", text)
				return
			}
			if v == 38 {
				p.pen.fg = c
			} else {
				p.pen.bg = c
			}
			k += n
		default:
			p.warnf("unknown SGR parameter %d in %s", v, text)
		}
	}
}

// extColour parses the parameters after 38 or 48: "5;n" or "2;r;g;b".
// It gives the colour and the count of parameters.
func extColour(ps []int) (screen.Colour, int, bool) {
	switch {
	case len(ps) >= 2 && ps[0] == 5 && ps[1] <= 255:
		return screen.IndexColour(uint8(ps[1])), 2, true
	case len(ps) >= 4 && ps[0] == 2 && ps[1] <= 255 && ps[2] <= 255 && ps[3] <= 255:
		return screen.RGBColour(uint8(ps[1]), uint8(ps[2]), uint8(ps[3])), 4, true
	}
	return screen.Colour{}, 0, false
}
