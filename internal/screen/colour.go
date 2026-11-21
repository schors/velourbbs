package screen

import (
	"fmt"
	"strings"
)

// ColourKind tells how a Colour is given.
type ColourKind uint8

// The colour kinds of section 7.1.
const (
	Default ColourKind = iota // the colour of the terminal
	Index                     // an index from 0 to 255
	RGB                       // a 24-bit colour
	VGA                       // a colour 0–15 of art with the VGA value
)

// Colour is a foreground or a background colour.
// Value is the index for Index and VGA, and 0xRRGGBB for RGB.
type Colour struct {
	Kind  ColourKind
	Value uint32
}

// IndexColour returns the colour with the index n.
func IndexColour(n uint8) Colour { return Colour{Kind: Index, Value: uint32(n)} }

// VGAColour returns the VGA colour with the index n, 0–15.
func VGAColour(n uint8) Colour { return Colour{Kind: VGA, Value: uint32(n)} }

// RGBColour returns the colour with the components r, g and b.
func RGBColour(r, g, b uint8) Colour {
	return Colour{Kind: RGB, Value: uint32(r)<<16 | uint32(g)<<8 | uint32(b)}
}

// String writes the colour in the style syntax of section 7.2.
// A VGA colour gives its index (screen-render, section 3.3).
func (c Colour) String() string {
	switch c.Kind {
	case Index, VGA:
		return fmt.Sprint(c.Value)
	case RGB:
		return fmt.Sprintf("#%06x", c.Value)
	default:
		return "-"
	}
}

// Attr is a set of text attributes.
type Attr uint8

// The attributes of section 7.1, in the order of the style syntax.
const (
	Bold Attr = 1 << iota
	Underline
	Blink
)

var attrNames = [...]struct {
	attr Attr
	name string
}{{Bold, "bold"}, {Underline, "underline"}, {Blink, "blink"}}

// Style is a foreground colour, a background colour and attributes.
type Style struct {
	Fg, Bg Colour
	Attr   Attr
}

// String writes the style in the syntax of section 7.2 with both colours.
// Example: "15/4 bold".
func (s Style) String() string {
	var b strings.Builder
	b.WriteString(s.Fg.String())
	b.WriteByte('/')
	b.WriteString(s.Bg.String())
	for _, a := range attrNames {
		if s.Attr&a.attr != 0 {
			b.WriteByte(' ')
			b.WriteString(a.name)
		}
	}
	return b.String()
}

// vga holds the values of the indexes 0–15 (telnet-term, section 7.3).
var vga = [16]uint32{
	0x000000, 0xaa0000, 0x00aa00, 0xaa5500, 0x0000aa, 0xaa00aa, 0x00aaaa, 0xaaaaaa,
	0x555555, 0xff5555, 0x55ff55, 0xffff55, 0x5555ff, 0xff55ff, 0x55ffff, 0xffffff,
}

// cube holds the component levels of the xterm colour cube.
var cube = [6]uint32{0, 95, 135, 175, 215, 255}

// Palette gives the value 0xRRGGBB of the index n, 0–255: VGA for 0–15, xterm for 16–255
// (telnet-term, section 7.3).
func Palette(n uint32) uint32 {
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
