// Package term holds the terminal of the telnet-term specification: the probe and
// the terminal values (sections 5 and 6), the page output, the colour reduction
// and the encoder (section 7), the key decoder, the key queue and the paste (section 8),
// the terminal of the line input (docs/specs/line-input.md) and the line form of
// velourbbs preview (docs/specs/screen-render.md, section 3.7).
// Specification: docs/specs/telnet-term.md.
package term

import (
	"strconv"
	"unicode/utf8"

	"github.com/schors/velourbbs/internal/screen"
)

// AppendPage appends the output of page p (section 7.1) to dst for a terminal
// with term.colors = colors. screen.ColorsTrue gives no reduction, screen.Colors256
// gives the 256-colour rule and each other value gives the 16-colour rule (section 7.3).
// vgaRGB is true when the palette of the terminal is not VGA (section 5.7):
// then a VGA colour goes out as its RGB value.
// The caller doubles IAC after it (section 4.3).
func AppendPage(dst []byte, p screen.Page, colors int, vgaRGB bool) []byte {
	e := encoder{dst: append(dst, "\x1b[?2026h"...), colors: colors, vgaRGB: vgaRGB}
	if p.Fill == screen.FillErase {
		e.model, e.out = p.Normal, reduceStyle(p.Normal, colors, vgaRGB)
		e.dst = appendSGR(e.dst, e.out)
	} else {
		e.dst = append(e.dst, "\x1b[0m"...)
	}
	e.dst = append(e.dst, "\x1b[2J"...)
	fillRow := func(row int) {
		if p.Fill != screen.FillSpaces {
			return
		}
		e.dst = appendPos(e.dst, row)
		e.spaces(p)
	}
	// The terminal rows from the top (section 7.1, step 2).
	for row := range p.Top {
		fillRow(row + 1)
	}
	var line screen.Row
	for i := range p.Rows {
		row := p.Top + i + 1
		line = p.Line(i, line)
		e.dst = appendPos(e.dst, row)
		e.row(line, func(dst []byte, col int) []byte { return appendCUP(dst, row, col) })
	}
	for row := p.Top + len(p.Rows); row < p.Height; row++ {
		fillRow(row + 1)
	}
	return append(e.dst, "\x1b[0m\x1b[?2026l"...)
}

// CursorPlace gives the cursor place of the page p, 1-based (section 7.5): the last terminal row
// that the output of p writes, at the column term.cols, or row 1, column 1 when it writes no row.
func CursorPlace(p screen.Page) (row, col int) {
	switch {
	case p.Fill == screen.FillSpaces && p.Height > 0:
		return p.Height, p.Cols
	case len(p.Rows) > 0:
		return min(p.Top+len(p.Rows), p.Height), p.Cols
	}
	return 1, 1
}

// AppendUpdate appends the update of section 7.5 to dst: the runs of the cells of the grid g
// that differ from the shown grid prev, then the move to the cursor place row, col.
// g and prev have the same count of rows and of cells in each row (screen.Page.Grid).
// Without a changed cell, it appends nothing. The other arguments are as of AppendPage.
func AppendUpdate(dst []byte, g, prev []screen.Row, row, col, colors int, vgaRGB bool) []byte {
	e := encoder{dst: dst, colors: colors, vgaRGB: vgaRGB}
	changed := false
	for i, line := range g {
		for _, r := range runs(line, prev[i]) {
			if !changed {
				e.dst = append(e.dst, "\x1b[?2026h\x1b[?25l"...)
				changed = true
			}
			n := i + 1
			e.dst = appendCUP(e.dst, n, r.start+1)
			e.cells(line, r.start, r.end, func(dst []byte, col int) []byte { return appendCUP(dst, n, col) })
		}
	}
	if !changed {
		return dst
	}
	e.dst = append(e.dst, "\x1b[0m"...)
	e.dst = appendCUP(e.dst, row, col)
	return append(e.dst, "\x1b[?25h\x1b[?2026l"...)
}

// run is the cells [start, end) of a row.
type run struct{ start, end int }

// maxGap is the most equal cells between two runs of one run (section 7.5).
const maxGap = 3

// runs gives the runs of the row line that differ from the shown row old (section 7.5), in order.
// A cell differs in its grapheme or its style.
func runs(line, old screen.Row) []run {
	var out []run
	for j := range line {
		if line[j].Grapheme == old[j].Grapheme && line[j].Style == old[j].Style {
			continue
		}
		if n := len(out); n > 0 && j-out[n-1].end <= maxGap {
			out[n-1].end = j + 1
			continue
		}
		out = append(out, run{j, j + 1})
	}
	for i := range out {
		out[i] = grow(line, out[i])
	}
	// A run that grows over the next run takes it.
	merged := out[:0]
	for _, r := range out {
		if n := len(merged); n > 0 && r.start <= merged[n-1].end {
			merged[n-1].end = max(merged[n-1].end, r.end)
			continue
		}
		merged = append(merged, r)
	}
	return merged
}

// grow grows the run r of the row line until the rules of section 7.5 hold:
// it takes each wide grapheme at its edges, the grapheme of more than one code point before it,
// and the cell after its last grapheme when that grapheme has more than one code point.
func grow(line screen.Row, r run) run {
	for {
		start, end := r.start, r.end
		if line[r.start].Grapheme == "" { // the second cell of a wide grapheme
			r.start--
		}
		if r.end < len(line) && line[r.end].Grapheme == "" {
			r.end++
		}
		if r.start > 0 {
			k := r.start - 1
			if line[k].Grapheme == "" {
				k--
			}
			if long(line[k].Grapheme) {
				r.start = k
			}
		}
		if r.end < len(line) && long(line[lastGrapheme(line, r.end)].Grapheme) {
			r.end++
		}
		if r.start == start && r.end == end {
			return r
		}
	}
}

// lastGrapheme gives the index of the first cell of the grapheme that ends before the cell end.
func lastGrapheme(line screen.Row, end int) int {
	k := end - 1
	if line[k].Grapheme == "" {
		k--
	}
	return k
}

// long reports whether the grapheme g has more than one code point: it needs the re-sync of section 7.4.
func long(g string) bool { return utf8.RuneCountInString(g) > 1 }

// AppendLines appends the line form of page p (screen-render, section 3.7) to dst
// for a terminal with term.colors = colors. The arguments are as of AppendPage.
// A VGA colour goes out as its RGB value.
func AppendLines(dst []byte, p screen.Page, colors int) []byte {
	return appendLines(dst, p.Header(), p, colors)
}

// AppendGridLines appends the grid of page p in the line form (live-screen, section 7) to dst:
// the first line of the grid dump, then each terminal row as a row of a page.
// The arguments are as of AppendLines.
func AppendGridLines(dst []byte, p screen.Page, colors int) []byte {
	return appendLines(dst, p.GridHeader(), p.GridPage(), colors)
}

// appendLines appends the line form of page p with the first line header.
func appendLines(dst []byte, header string, p screen.Page, colors int) []byte {
	dst = append(append(dst, header...), '\n')
	e := encoder{dst: dst, colors: colors, vgaRGB: true}
	// endRow resets the style: each row starts with default colours without attributes.
	endRow := func() {
		e.dst = append(e.dst, "\x1b[0m\n"...)
		e.model, e.out = screen.Style{}, screen.Style{}
	}
	fillRow := func() {
		if p.Fill != screen.FillNone {
			e.spaces(p)
		}
		endRow()
	}
	for range p.Top {
		fillRow()
	}
	var line screen.Row
	for i := range p.Rows {
		line = p.Line(i, line)
		e.row(line, appendCHA)
		endRow()
	}
	for range p.Height - p.Top - len(p.Rows) {
		fillRow()
	}
	return e.dst
}

// encoder appends cells to dst with the SGR of section 7.2 and the colour reduction of section 7.3.
type encoder struct {
	dst        []byte
	colors     int
	vgaRGB     bool
	model, out screen.Style // the model style and the output style of the previous cell
}

// cell appends the cell c: the SGR when its output style differs, then its grapheme.
func (e *encoder) cell(c screen.Cell) {
	if c.Style != e.model {
		e.model = c.Style
		if s := reduceStyle(e.model, e.colors, e.vgaRGB); s != e.out {
			e.out = s
			e.dst = appendSGR(e.dst, e.out)
		}
	}
	e.dst = appendUTF8(e.dst, c.Grapheme)
}

// spaces appends the cells of a fill row: p.Cols spaces in p.Normal.
func (e *encoder) spaces(p screen.Page) {
	for range p.Cols {
		e.cell(screen.Cell{Grapheme: " ", Style: p.Normal})
	}
}

// row appends the cells of line with the re-sync of section 7.4.
// resync appends the move to the 1-based column col of the row.
func (e *encoder) row(line screen.Row, resync func(dst []byte, col int) []byte) {
	e.cells(line, 0, len(line), resync)
}

// cells appends the cells [start, end) of line with the re-sync of section 7.4: a grapheme
// of more than one code point that a cell of the row follows. The re-sync of the last grapheme
// before end writes a space over the cell end.
func (e *encoder) cells(line screen.Row, start, end int, resync func(dst []byte, col int) []byte) {
	for j := start; j < end; j++ {
		c := line[j]
		if c.Grapheme == "" { // the second cell of a wide grapheme gives no output
			continue
		}
		e.cell(c)
		// next is the index of the cell after the grapheme.
		next := j + 1
		if next < len(line) && line[next].Grapheme == "" {
			next++
		}
		if next < len(line) && long(c.Grapheme) {
			e.dst = append(e.dst, ' ')
			e.dst = resync(e.dst, next+1)
		}
	}
}

// appendPos appends ESC[<row>;1H.
func appendPos(dst []byte, row int) []byte {
	return appendCUP(dst, row, 1)
}

// appendCUP appends ESC[<row>;<col>H.
func appendCUP(dst []byte, row, col int) []byte {
	dst = append(dst, "\x1b["...)
	dst = strconv.AppendInt(dst, int64(row), 10)
	dst = append(dst, ';')
	dst = strconv.AppendInt(dst, int64(col), 10)
	return append(dst, 'H')
}

// appendCHA appends ESC[<col>G.
func appendCHA(dst []byte, col int) []byte {
	dst = append(dst, "\x1b["...)
	dst = strconv.AppendInt(dst, int64(col), 10)
	return append(dst, 'G')
}

// appendUTF8 is the encoder of section 7.4. The model holds valid UTF-8
// (bbt-format, sections 2, 3.1 and 4.3, and art-format, section 2), so it copies the bytes.
func appendUTF8(dst []byte, g string) []byte {
	return append(dst, g...)
}

// appendSGR appends ESC[0;<params>m, or ESC[0m without parameters (section 7.2).
func appendSGR(dst []byte, s screen.Style) []byte {
	dst = append(dst, "\x1b[0"...)
	for _, a := range [...]struct {
		attr screen.Attr
		p    string
	}{{screen.Bold, ";1"}, {screen.Underline, ";4"}, {screen.Blink, ";5"}} {
		if s.Attr&a.attr != 0 {
			dst = append(dst, a.p...)
		}
	}
	dst = appendColour(dst, s.Fg, 30, 90, "38")
	dst = appendColour(dst, s.Bg, 40, 100, "48")
	return append(dst, 'm')
}

// appendColour appends the parameter of colour c with a leading ';'.
// low is the base of the indexes 0–7, high of 8–15, ext is 38 or 48.
func appendColour(dst []byte, c screen.Colour, low, high int64, ext string) []byte {
	switch {
	case c.Kind == screen.Index && c.Value < 8:
		dst = append(dst, ';')
		return strconv.AppendInt(dst, low+int64(c.Value), 10)
	case c.Kind == screen.Index && c.Value < 16:
		dst = append(dst, ';')
		return strconv.AppendInt(dst, high+int64(c.Value)-8, 10)
	case c.Kind == screen.Index:
		dst = append(dst, ';')
		dst = append(dst, ext...)
		dst = append(dst, ";5;"...)
		return strconv.AppendUint(dst, uint64(c.Value), 10)
	case c.Kind == screen.RGB:
		dst = append(dst, ';')
		dst = append(dst, ext...)
		dst = append(dst, ";2;"...)
		dst = strconv.AppendUint(dst, uint64(c.Value>>16&0xff), 10)
		dst = append(dst, ';')
		dst = strconv.AppendUint(dst, uint64(c.Value>>8&0xff), 10)
		dst = append(dst, ';')
		return strconv.AppendUint(dst, uint64(c.Value&0xff), 10)
	}
	return dst // default
}

// reduceStyle reduces both colours of s to colors.
func reduceStyle(s screen.Style, colors int, vgaRGB bool) screen.Style {
	s.Fg = reduce(s.Fg, colors, vgaRGB)
	s.Bg = reduce(s.Bg, colors, vgaRGB)
	return s
}

// reduce reduces colour c to colors (section 7.3). A VGA colour first becomes
// its RGB value when vgaRGB is true, else its index.
func reduce(c screen.Colour, colors int, vgaRGB bool) screen.Colour {
	if c.Kind == screen.VGA {
		if vgaRGB {
			c = screen.Colour{Kind: screen.RGB, Value: screen.Palette(c.Value)}
		} else {
			c.Kind = screen.Index
		}
	}
	switch {
	case colors == screen.ColorsTrue || c.Kind == screen.Default:
		return c
	case colors == screen.Colors256:
		if c.Kind == screen.RGB {
			return screen.Colour{Kind: screen.Index, Value: nearest(c.Value, 16, 255)}
		}
		return c
	case c.Kind == screen.RGB:
		return screen.Colour{Kind: screen.Index, Value: nearest(c.Value, 0, 15)}
	case c.Value >= 16:
		return screen.Colour{Kind: screen.Index, Value: nearest(screen.Palette(c.Value), 0, 15)}
	}
	return c
}

// nearest gives the index from lo to hi with the smallest squared RGB distance to rgb.
// On a tie, the smaller index wins.
func nearest(rgb, lo, hi uint32) uint32 {
	best, bestD := lo, -1
	for n := lo; n <= hi; n++ {
		if d := distance(rgb, screen.Palette(n)); bestD < 0 || d < bestD {
			best, bestD = n, d
		}
	}
	return best
}

// distance gives the squared RGB distance of a and b, both 0xRRGGBB.
func distance(a, b uint32) int {
	d := 0
	for shift := 0; shift <= 16; shift += 8 {
		x := int(a>>shift&0xff) - int(b>>shift&0xff)
		d += x * x
	}
	return d
}
