package screen

import (
	"errors"
	"unicode/utf8"
)

// Grapheme is one grapheme of a value and its width in cells.
type Grapheme struct {
	Text  string
	Width int
}

// space is the grapheme that pads a fixed-width field.
var space = Grapheme{" ", 1}

// CleanValue applies the rules of section 4.3 to a value. It replaces each invalid
// UTF-8 byte with U+FFFD. It removes the control characters and the graphemes of width 0.
// It replaces each other grapheme of more than 32 bytes with U+FFFD.
//
// The result is a list of graphemes, not a string: a removal can join
// the graphemes on its two sides into a new grapheme when the string is segmented again.
func CleanValue(s string) []Grapheme {
	if !utf8.ValidString(s) {
		s = string([]rune(s)) // each invalid byte becomes U+FFFD
	}
	var gs []Grapheme
	for g, w := range Graphemes(s) {
		switch err := CheckGrapheme(g, w); {
		case err == nil:
			gs = append(gs, Grapheme{g, w})
		case w > 0 && errors.Is(err, ErrLongGrapheme):
			gs = append(gs, Grapheme{string(utf8.RuneError), 1})
		}
	}
	return gs
}

// FitValue gives the graphemes of a fixed-width field of width cells (section 4.3).
// It cuts the end of gs at a grapheme boundary. A wide grapheme that does not fit
// becomes a space. It pads a short value with spaces on the right,
// or on the left when right is true.
func FitValue(gs []Grapheme, width int, right bool) []Grapheme {
	out := make([]Grapheme, 0, width)
	used, cut := 0, false
	for _, g := range gs {
		if used+g.Width > width {
			cut = true
			break
		}
		out = append(out, g)
		used += g.Width
	}
	pad := make([]Grapheme, width-used)
	for i := range pad {
		pad[i] = space
	}
	if right && !cut {
		return append(pad, out...)
	}
	return append(out, pad...)
}
