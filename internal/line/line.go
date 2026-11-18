// Package line holds the line input: the field, the line, the edit keys,
// the grapheme rule, the keys of a paste and the cells of a draw. It has no I/O.
// Specification: docs/specs/line-input.md.
package line

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/schors/velourbbs/internal/screen"
)

// Field is the place and the form of the line on the terminal screen (section 3).
type Field struct {
	Row, Col int          // the terminal row and column of the first cell, 1-based
	Max      int          // the most characters of the line, M >= 1
	Style    screen.Style // the style of the cells
	Masked   bool         // each character shows as *
}

// Result is the effect of a key.
type Result uint8

const (
	None    Result = iota // no change of the line and the cursor position
	Changed               // the line or the cursor position changed: draw the field again
	Enter                 // the input ends; the caller gets the line
	Cancel                // the input is cancelled; the caller gets no line
)

// Line is the state of one line input.
type Line struct {
	f    Field
	text []rune // each rune is one grapheme of width 1 (section 5)
	cur  int    // the cursor position, 0..len(text)
}

// New gives an empty line in the field f on a screen of cols x rows cells.
// It panics when the field does not fit the screen: this is a host bug (section 3).
func New(f Field, cols, rows int) *Line {
	if f.Row < 1 || f.Col < 1 || f.Max < 1 || f.Col+f.Max > cols || f.Row > rows {
		panic(fmt.Sprintf("line: field %+v does not fit %dx%d", f, cols, rows))
	}
	return &Line{f: f, text: make([]rune, 0, f.Max)}
}

// Key applies the key k of the key decoder (section 4). paste is true for a key
// of a paste: then only a character, SPACE and ENTER act (section 6).
func (l *Line) Key(k string, paste bool) Result {
	if k == "SPACE" {
		k = " "
	}
	if r, size := utf8.DecodeRuneInString(k); size == len(k) && r != utf8.RuneError {
		return l.insert(r)
	}
	if k == "ENTER" {
		return Enter
	}
	if paste {
		return None
	}
	switch k {
	case "ESC":
		return Cancel
	case "LEFT":
		return l.move(l.cur - 1)
	case "RIGHT":
		return l.move(l.cur + 1)
	case "HOME":
		return l.move(0)
	case "END":
		return l.move(len(l.text))
	case "BS":
		if l.cur > 0 && l.remove(l.cur-1) {
			l.cur--
			return Changed
		}
	case "DEL":
		if l.cur < len(l.text) && l.remove(l.cur) {
			return Changed
		}
	}
	return None
}

// Text gives the line.
func (l *Line) Text() string {
	return string(l.text)
}

// Draw is the field on the terminal screen (section 7).
type Draw struct {
	Field  Field
	Cells  string // the M+1 cells: one character of width 1 each
	Cursor int    // the cursor position
}

// Draw gives the cells of the field: each character of the line, or * when masked,
// then spaces.
func (l *Line) Draw() Draw {
	var b strings.Builder
	if l.f.Masked {
		b.WriteString(strings.Repeat("*", len(l.text)))
	} else {
		b.WriteString(string(l.text))
	}
	b.WriteString(strings.Repeat(" ", l.f.Max+1-len(l.text)))
	return Draw{Field: l.f, Cells: b.String(), Cursor: l.cur}
}

// move moves the cursor to the position p, if it is in the line.
func (l *Line) move(p int) Result {
	if p < 0 || p > len(l.text) || p == l.cur {
		return None
	}
	l.cur = p
	return Changed
}

// insert inserts r at the cursor position (section 5).
func (l *Line) insert(r rune) Result {
	if len(l.text) >= l.f.Max || screen.IsControl(r) {
		return None
	}
	next := slices.Insert(slices.Clone(l.text), l.cur, r)
	if !valid(next) {
		return None
	}
	l.text = next
	l.cur++
	return Changed
}

// remove removes the character at the index i. It makes no change when the line
// after the removal breaks the grapheme rule (section 5).
func (l *Line) remove(i int) bool {
	next := slices.Delete(slices.Clone(l.text), i, i+1)
	if !valid(next) {
		return false
	}
	l.text = next
	return true
}

// valid reports whether each character of text is one grapheme of width 1:
// the grapheme rule of section 5.
func valid(text []rune) bool {
	for g, w := range screen.Graphemes(string(text)) {
		if w != 1 || utf8.RuneCountInString(g) != 1 {
			return false
		}
	}
	return true
}
