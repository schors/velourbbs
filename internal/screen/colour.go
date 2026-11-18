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
