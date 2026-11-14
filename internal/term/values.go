package term

import (
	"strings"

	"github.com/schors/velourbbs/internal/screen"
	"github.com/schors/velourbbs/internal/telnet"
)

// typeLen is the maximum number of graphemes of a TTYPE name (section 6).
const typeLen = 40

// The default size (section 5.2).
const (
	defaultCols = 80
	defaultRows = 24
)

// names16 are the names of rule 3 of the colour depth (section 6).
var names16 = map[string]bool{
	"ansi": true, "dumb": true, "vt52": true, "vt100": true, "vt102": true, "vt220": true,
	"vt320": true, "rxvt": true, "xterm-color": true, "eterm-color": true, "cygwin": true,
}

// cleanName removes the control characters of a TTYPE name and cuts it
// to typeLen graphemes (section 6). It keeps the case.
func cleanName(name string) string {
	var b strings.Builder
	for i, g := range screen.CleanValue(name) {
		if i == typeLen {
			break
		}
		b.WriteString(g.Text)
	}
	return b.String()
}

// TypeColors gives term.type and term.colors from the terminal name name (section 6):
// the clean name in lower case, or ansi for an empty name, and its colour depth.
func TypeColors(name string) (typ string, colors int) {
	typ = "ansi"
	if c := cleanName(name); c != "" {
		typ = strings.ToLower(c)
	}
	return typ, colorDepth(typ)
}

// colorDepth gives term.colors for term.type typ (section 6).
func colorDepth(typ string) int {
	switch {
	case strings.HasSuffix(typ, "-direct") || strings.Contains(typ, "truecolor") || strings.Contains(typ, "24bit"):
		return screen.ColorsTrue
	case strings.Contains(typ, "256color"):
		return screen.Colors256
	case strings.HasSuffix(typ, "-16color") || strings.HasSuffix(typ, "-8color") || names16[typ]:
		return screen.Colors16
	}
	return screen.Colors256
}

// clamp gives the size cols x rows in the range of section 5.2.
func clamp(cols, rows int) (int, int) {
	return min(max(cols, screen.MinCols), screen.MaxCols), min(max(rows, screen.MinRows), screen.MaxRows)
}

// replies holds the replies of the probe.
type replies struct {
	ttype    string      // the first TTYPE name, as sent
	ttypeGot bool        // the first TTYPE name arrived
	naws     telnet.NAWS // the last NAWS without 0 in a dimension; zero is no size
	cpr      CPR         // the reply to the charset probe
	cprGot   bool        // the reply to the charset probe arrived
}

// values gives the terminal values of section 6. utf8 is false when the charset probe
// gives "not UTF-8" (section 5.4).
func (r *replies) values() (t screen.Term, utf8 bool) {
	t = screen.Term{Charset: "utf-8"}
	t.Type, t.Colors = TypeColors(r.ttype)
	utf8 = !r.cprGot || r.cpr.Col != 3
	t.Cols, t.Rows = r.naws.Cols, r.naws.Rows
	if t.Cols == 0 {
		t.Cols, t.Rows = defaultCols, defaultRows
	}
	t.Cols, t.Rows = clamp(t.Cols, t.Rows)
	return t, utf8
}
