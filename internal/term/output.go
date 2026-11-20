// Package term holds the terminal of the telnet-term specification: the probe and
// the terminal values (sections 5 and 6), the page output, the colour reduction
// and the encoder (section 7), the key decoder, the key queue and the paste (section 8),
// the terminal of the line input (docs/specs/line-input.md) and the line form of
// velourbbs preview (docs/specs/screen-render.md, section 3.7).
// Specification: docs/specs/telnet-term.md.
package term

import (
	"slices"
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
	return appendPage(dst, p, nil, colors, vgaRGB)
}

// AppendUpdate appends the output of the update of the shown page prev to the page p (section 7.1):
// the rows of p whose cells differ from the row of prev, without the clear and without the fill rows.
// p and prev have the same size, the same Top and the same count of rows. The arguments are as of AppendPage.
func AppendUpdate(dst []byte, p, prev screen.Page, colors int, vgaRGB bool) []byte {
	return appendPage(dst, p, &prev, colors, vgaRGB)
}

// appendPage appends the output of the page p, or of its update when prev is the shown page.
func appendPage(dst []byte, p screen.Page, prev *screen.Page, colors int, vgaRGB bool) []byte {
	e := encoder{dst: append(dst, "\x1b[?2026h"...), colors: colors, vgaRGB: vgaRGB}
	if p.Fill == screen.FillErase {
		e.model, e.out = p.Normal, reduceStyle(p.Normal, colors, vgaRGB)
		e.dst = appendSGR(e.dst, e.out)
	} else {
		e.dst = append(e.dst, "\x1b[0m"...)
	}
	if prev == nil {
		e.dst = append(e.dst, "\x1b[2J"...)
	}
	fillRow := func(row int) {
		if p.Fill != screen.FillSpaces || prev != nil {
			return
		}
		e.dst = appendPos(e.dst, row)
		e.spaces(p)
	}
	// The terminal rows from the top (section 7.1, step 2).
	for row := range p.Top {
		fillRow(row + 1)
	}
	var line, old screen.Row
	for i := range p.Rows {
		row := p.Top + i + 1
		line = p.Line(i, line)
		if prev != nil {
			if old = prev.Line(i, old); slices.Equal(line, old) {
				continue
			}
		}
		e.dst = appendPos(e.dst, row)
		e.row(line, func(dst []byte, col int) []byte { return appendCUP(dst, row, col) })
	}
	for row := p.Top + len(p.Rows); row < p.Height; row++ {
		fillRow(row + 1)
	}
	return append(e.dst, "\x1b[0m\x1b[?2026l"...)
}

// AppendLines appends the line form of page p (screen-render, section 3.7) to dst
// for a terminal with term.colors = colors. The arguments are as of AppendPage.
// A VGA colour goes out as its RGB value.
func AppendLines(dst []byte, p screen.Page, colors int) []byte {
	dst = append(append(dst, p.Header()...), '\n')
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
	for j, c := range line {
		if c.Grapheme == "" { // the second cell of a wide grapheme gives no output
			continue
		}
		e.cell(c)
		// next is the index of the cell after the grapheme.
		next := j + 1
		if next < len(line) && line[next].Grapheme == "" {
			next++
		}
		if next < len(line) && utf8.RuneCountInString(c.Grapheme) > 1 {
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
			c = screen.Colour{Kind: screen.RGB, Value: vga[c.Value]}
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
		return screen.Colour{Kind: screen.Index, Value: nearest(palette(c.Value), 0, 15)}
	}
	return c
}

// nearest gives the index from lo to hi with the smallest squared RGB distance to rgb.
// On a tie, the smaller index wins.
func nearest(rgb, lo, hi uint32) uint32 {
	best, bestD := lo, -1
	for n := lo; n <= hi; n++ {
		if d := distance(rgb, palette(n)); bestD < 0 || d < bestD {
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

// vga holds the values of the indexes 0–15 (section 7.3).
var vga = [16]uint32{
	0x000000, 0xaa0000, 0x00aa00, 0xaa5500, 0x0000aa, 0xaa00aa, 0x00aaaa, 0xaaaaaa,
	0x555555, 0xff5555, 0x55ff55, 0xffff55, 0x5555ff, 0xff55ff, 0x55ffff, 0xffffff,
}

// cube holds the component levels of the xterm colour cube.
var cube = [6]uint32{0, 95, 135, 175, 215, 255}

// palette gives the value 0xRRGGBB of index n: VGA for 0–15, xterm for 16–255.
func palette(n uint32) uint32 {
	switch {
	case n < 16:
		return vga[n]
	case n < 232:
		n -= 16
		return cube[n/36]<<16 | cube[n/6%6]<<8 | cube[n%6]
	}
	v := 8 + 10*(n-232)
	return v<<16 | v<<8 | v
}
