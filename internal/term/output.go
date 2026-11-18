// Package term holds the terminal of the telnet-term specification: the probe and
// the terminal values (sections 5 and 6), the page output, the colour reduction
// and the encoder (section 7), the key decoder and the key queue (section 8).
// Specification: docs/specs/telnet-term.md.
package term

import (
	"strconv"

	"github.com/schors/velourbbs/internal/screen"
)

// AppendPage appends the output of page p (section 7.1) to dst for a terminal
// with term.colors = colors. screen.ColorsTrue gives no reduction, screen.Colors256
// gives the 256-colour rule and each other value gives the 16-colour rule (section 7.3).
// vgaRGB is true when the palette of the terminal is not VGA (section 5.7):
// then a VGA colour goes out as its RGB value.
// The caller doubles IAC after it (section 4.3).
func AppendPage(dst []byte, p screen.Page, colors int, vgaRGB bool) []byte {
	var model, out screen.Style // the model style and the output style of the previous cell
	dst = append(dst, "\x1b[?2026h"...)
	if p.Fill == screen.FillErase {
		model, out = p.Normal, reduceStyle(p.Normal, colors, vgaRGB)
		dst = appendSGR(dst, out)
	} else {
		dst = append(dst, "\x1b[0m"...)
	}
	dst = append(dst, "\x1b[2J"...)
	appendCell := func(c screen.Cell) {
		if c.Style != model {
			model = c.Style
			if s := reduceStyle(model, colors, vgaRGB); s != out {
				out = s
				dst = appendSGR(dst, out)
			}
		}
		dst = appendUTF8(dst, c.Grapheme)
	}
	fillRow := func(row int) {
		if p.Fill != screen.FillSpaces {
			return
		}
		dst = appendPos(dst, row)
		for range p.Cols {
			appendCell(screen.Cell{Grapheme: " ", Style: p.Normal})
		}
	}
	// The terminal rows from the top (section 7.1, step 2).
	for row := range p.Top {
		fillRow(row + 1)
	}
	var line screen.Row
	for i := range p.Rows {
		dst = appendPos(dst, p.Top+i+1)
		line = p.Line(i, line)
		for _, c := range line {
			if c.Grapheme != "" { // the second cell of a wide grapheme gives no output
				appendCell(c)
			}
		}
	}
	for row := p.Top + len(p.Rows); row < p.Height; row++ {
		fillRow(row + 1)
	}
	return append(dst, "\x1b[0m\x1b[?2026l"...)
}

// appendPos appends ESC[<row>;1H.
func appendPos(dst []byte, row int) []byte {
	dst = append(dst, "\x1b["...)
	dst = strconv.AppendInt(dst, int64(row), 10)
	return append(dst, ";1H"...)
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
