package term

import (
	"strings"

	"github.com/schors/velourbbs/internal/screen"
)

// typeLen is the maximum number of graphemes of a client text: a TTYPE name,
// term.name and a reply in the log (sections 5.6 and 6).
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

// CleanName removes the control characters of a client text and cuts it
// to typeLen graphemes (sections 5.6 and 6). It keeps the case.
func CleanName(name string) string {
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
	if c := CleanName(name); c != "" {
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
