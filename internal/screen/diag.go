package screen

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Diag is an error or a warning in a file (screen-render, section 3.1).
// Row and Col are 1-based.
type Diag struct {
	File     string
	Row, Col int
	Warning  bool
	Msg      string
}

// String gives the form "<file>:<row>:<col>: [warning: ]<message>".
// The file name and the message can hold untrusted data, so String escapes
// each byte of invalid UTF-8 and each rune that is not printable, as Go does in a quoted string.
// Then a diagnostic cannot write a control sequence or a second line.
func (d Diag) String() string {
	var s string
	if d.Warning {
		s = fmt.Sprintf("%s:%d:%d: warning: %s", d.File, d.Row, d.Col, d.Msg)
	} else {
		s = fmt.Sprintf("%s:%d:%d: %s", d.File, d.Row, d.Col, d.Msg)
	}
	return escape(s)
}

// escape replaces each byte of invalid UTF-8 and each rune that is not printable
// with its Go escape sequence.
func escape(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		r, n := utf8.DecodeRuneInString(s[i:])
		switch {
		case r == utf8.RuneError && n == 1:
			fmt.Fprintf(&b, `\x%02x`, s[i])
		case !strconv.IsPrint(r):
			q := strconv.QuoteRune(r)
			b.WriteString(q[1 : len(q)-1])
		default:
			b.WriteString(s[i : i+n])
		}
		i += n
	}
	return b.String()
}

// Diags collects the diagnostics of a load.
type Diags []Diag

// Errorf adds an error.
func (ds *Diags) Errorf(file string, row, col int, format string, args ...any) {
	*ds = append(*ds, Diag{File: file, Row: row, Col: col, Msg: fmt.Sprintf(format, args...)})
}

// Warnf adds a warning.
func (ds *Diags) Warnf(file string, row, col int, format string, args ...any) {
	*ds = append(*ds, Diag{File: file, Row: row, Col: col, Warning: true, Msg: fmt.Sprintf(format, args...)})
}

// HasErrors reports whether ds has an error.
func (ds Diags) HasErrors() bool {
	return slices.ContainsFunc(ds, func(d Diag) bool { return !d.Warning })
}

// Sort orders ds by file, row and column. Diagnostics at one position keep their order.
func (ds Diags) Sort() {
	slices.SortStableFunc(ds, func(a, b Diag) int {
		return cmp.Or(cmp.Compare(a.File, b.File), cmp.Compare(a.Row, b.Row), cmp.Compare(a.Col, b.Col))
	})
}

// MaxFileDiags is the limit of errors and warnings of one file (screen-render, section 3.1).
const MaxFileDiags = 100

// FileDiags collects the diagnostics of one file up to MaxFileDiags.
// After the limit it adds one error "too many diagnostics" and drops the rest.
type FileDiags struct {
	File  string
	Diags Diags
}

// Errorf adds an error.
func (f *FileDiags) Errorf(row, col int, format string, args ...any) {
	if f.room(row, col) {
		f.Diags.Errorf(f.File, row, col, format, args...)
	}
}

// Warnf adds a warning.
func (f *FileDiags) Warnf(row, col int, format string, args ...any) {
	if f.room(row, col) {
		f.Diags.Warnf(f.File, row, col, format, args...)
	}
}

// full reports whether the file has too many diagnostics.
func (f *FileDiags) full() bool { return len(f.Diags) > MaxFileDiags }

// room reports whether one more diagnostic fits. At the limit it adds the last error.
func (f *FileDiags) room(row, col int) bool {
	switch {
	case f.full():
		return false
	case len(f.Diags) == MaxFileDiags:
		f.Diags.Errorf(f.File, row, col, "too many diagnostics")
		return false
	}
	return true
}
