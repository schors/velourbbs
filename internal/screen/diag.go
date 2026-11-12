package screen

import (
	"cmp"
	"fmt"
	"slices"
)

// Diag is an error or a warning in a file (section 9.1).
// Row and Col are 1-based.
type Diag struct {
	File     string
	Row, Col int
	Warning  bool
	Msg      string
}

// String gives the form "<file>:<row>:<col>: [warning: ]<message>".
func (d Diag) String() string {
	if d.Warning {
		return fmt.Sprintf("%s:%d:%d: warning: %s", d.File, d.Row, d.Col, d.Msg)
	}
	return fmt.Sprintf("%s:%d:%d: %s", d.File, d.Row, d.Col, d.Msg)
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
