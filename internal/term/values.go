package term

import (
	"strconv"
	"strings"

	"newbbs/internal/screen"
	"newbbs/internal/telnet"
)

// The MTTS bits of section 5.3.
const (
	mttsUTF8 = 4
	mtts256  = 8
	mttsTrue = 256
)

// maxNames is the number of TTYPE names after which TTYPE ends (section 5.3).
const maxNames = 4

// typeLen is the maximum number of graphemes of a TTYPE name (section 6).
const typeLen = 40

// The default size (section 5.2).
const (
	defaultCols = 80
	defaultRows = 25
)

// ttypeLoop holds the TTYPE names of the probe (section 5.3).
type ttypeLoop struct {
	names []string // the names as sent, without the name that repeats
	ended bool
}

// add adds a TTYPE name. It gives true when the host sends SEND again.
// After the end of TTYPE, it ignores the name.
func (l *ttypeLoop) add(name string) bool {
	if l.ended {
		return false
	}
	if n := len(l.names); n > 0 && (name == l.names[0] || name == l.names[n-1]) {
		l.ended = true
		return false
	}
	l.names = append(l.names, name)
	l.ended = len(l.names) == maxNames
	return !l.ended
}

// parseMTTS gives the bits of an MTTS name: MTTS in any case, a space
// and 1–9 decimal digits (section 5.3). ok is false for each other name.
func parseMTTS(name string) (bits int, ok bool) {
	// A 5-byte prefix with a non-ASCII character has fewer than 5 characters,
	// so EqualFold does not match it.
	if len(name) < 6 || len(name) > 14 || !strings.EqualFold(name[:5], "MTTS ") {
		return 0, false
	}
	digits := name[5:] // 1–9 bytes
	for _, c := range digits {
		if c < '0' || c > '9' {
			return 0, false
		}
	}
	bits, _ = strconv.Atoi(digits)
	return bits, true
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

// colorDepth gives term.colors (section 6). names are the terminal names as sent,
// not cut. hasMTTS tells that a TTYPE name gave the MTTS bits mtts.
func colorDepth(names []string, mtts int, hasMTTS bool) int {
	if hasMTTS {
		switch {
		case mtts&mttsTrue != 0:
			return ColorsTrue
		case mtts&mtts256 != 0:
			return Colors256
		}
		return Colors16
	}
	depth := Colors16
	for _, n := range names {
		n = strings.ToLower(n)
		switch {
		case strings.HasSuffix(n, "-direct") || strings.Contains(n, "truecolor") || strings.Contains(n, "24bit"):
			return ColorsTrue
		case strings.Contains(n, "256color"):
			depth = Colors256
		}
	}
	return depth
}

// clamp gives the size cols x rows in the range of section 5.2.
func clamp(cols, rows int) (int, int) {
	return min(max(cols, screen.MinCols), screen.MaxCols), min(max(rows, screen.MinRows), screen.MaxRows)
}

// replies holds the replies of the probe.
type replies struct {
	ttype     ttypeLoop
	naws      telnet.NAWS // the last NAWS without 0 in a dimension; zero is no size
	charset   bool        // the host sent the charset probe
	sizeProbe bool        // the host sent the size probe
	cprs      []CPR       // the CPR replies, in the order of the probes
}

// mtts gives the bits of the first MTTS name. ok is false without an MTTS name.
func (r *replies) mtts() (bits int, ok bool) {
	for _, n := range r.ttype.names {
		if bits, ok := parseMTTS(n); ok {
			return bits, true
		}
	}
	return 0, false
}

// values gives the terminal values of section 6. utf8 is false when the charset probe
// gives "not UTF-8" (section 5.4).
func (r *replies) values() (t screen.Term, utf8 bool) {
	var names []string // the terminal names
	for _, n := range r.ttype.names {
		if _, ok := parseMTTS(n); !ok {
			names = append(names, n)
		}
	}
	mtts, hasMTTS := r.mtts()
	t = screen.Term{Type: "ansi", Charset: "utf-8", Colors: colorDepth(names, mtts, hasMTTS)}
	for _, n := range names {
		if c := cleanName(n); c != "" {
			t.Type = strings.ToLower(c)
			break
		}
	}

	utf8 = true
	cprs := r.cprs
	if r.charset && len(cprs) > 0 {
		utf8 = cprs[0].Col != 3
		cprs = cprs[1:]
	}
	t.Cols, t.Rows = r.naws.Cols, r.naws.Rows
	if t.Cols == 0 && r.sizeProbe && len(cprs) > 0 && cprs[0].Row > 0 && cprs[0].Col > 0 {
		t.Cols, t.Rows = cprs[0].Col, cprs[0].Row
	}
	if t.Cols == 0 {
		t.Cols, t.Rows = defaultCols, defaultRows
	}
	t.Cols, t.Rows = clamp(t.Cols, t.Rows)
	return t, utf8
}
