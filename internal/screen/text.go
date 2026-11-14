package screen

import (
	"errors"
	"fmt"
	"iter"
	"strings"

	"github.com/clipperhouse/displaywidth"
)

// MaxGraphemeBytes is the size limit of a grapheme (sections 2.1 and 9).
const MaxGraphemeBytes = 32

// The errors of CheckGrapheme. A file reports each of them as an error.
// A value removes a grapheme with ErrControl or ErrZeroWidth
// and replaces a grapheme with ErrLongGrapheme (section 4.3).
var (
	ErrControl      = errors.New("control character")
	ErrZeroWidth    = errors.New("zero-width grapheme")
	ErrLongGrapheme = errors.New("grapheme over 32 bytes")
)

// IsControl reports whether r is a control character of section 2.1.
func IsControl(r rune) bool {
	return r <= 0x1F || 0x7F <= r && r <= 0x9F ||
		0x202A <= r && r <= 0x202E || 0x2066 <= r && r <= 0x2069
}

// Graphemes returns each grapheme of s with its width in cells.
// The width is 0, 1 or 2. CheckGrapheme rejects the width 0.
func Graphemes(s string) iter.Seq2[string, int] {
	return func(yield func(string, int) bool) {
		g := displaywidth.StringGraphemes(s)
		for g.Next() {
			if !yield(g.Value(), g.Width()) {
				return
			}
		}
	}
}

// CheckGrapheme checks the grapheme g of the width w against section 2.1.
// The error wraps ErrControl, ErrLongGrapheme or ErrZeroWidth.
func CheckGrapheme(g string, w int) error {
	switch {
	case strings.ContainsFunc(g, IsControl):
		return fmt.Errorf("%w %s", ErrControl, codePoints(g))
	case len(g) > MaxGraphemeBytes:
		return fmt.Errorf("%w: %d bytes", ErrLongGrapheme, len(g))
	case w == 0:
		return fmt.Errorf("%w %s", ErrZeroWidth, codePoints(g))
	}
	return nil
}

// codePoints writes g as code points. Example: "U+0009".
func codePoints(g string) string {
	var b strings.Builder
	for i, r := range []rune(g) {
		if i > 0 {
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, "U+%04X", r)
	}
	return b.String()
}

// IsIdent reports whether s matches [a-z][a-z0-9_]*: a theme style name,
// an alias or the first part of a field path.
func IsIdent(s string) bool {
	for i := range len(s) {
		c := s[i]
		ok := 'a' <= c && c <= 'z' || i > 0 && ('0' <= c && c <= '9' || c == '_')
		if !ok {
			return false
		}
	}
	return s != ""
}
