package term

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/schors/velourbbs/internal/screen"
	"github.com/schors/velourbbs/internal/telnet"
)

// paletteSize is the number of OSC 4 queries: the indexes 0–15 (section 5.7).
const paletteSize = 16

// probeWrite is the write of step 4 of section 5.1: the charset probe (section 5.4),
// the terminal queries in the order of section 5.7 and the erase.
var probeWrite = func() string {
	var b strings.Builder
	b.WriteString("\râ\x1b[6n")                            // charset probe
	b.WriteString("\x1b[38;2;1;2;3m\x1bP$qm\x1b\\\x1b[0m") // DECRQSS for SGR
	b.WriteString("\x1b[>0q")                              // XTVERSION
	b.WriteString("\x1b[>c")                               // DA2
	for n := range paletteSize {
		fmt.Fprintf(&b, "\x1b]4;%d;?\x1b\\", n) // OSC 4
	}
	b.WriteString("\x1b[c")   // DA1
	b.WriteString("\r\x1b[K") // erase
	return b.String()
}()

// CPR is a cursor position report ESC[<row>;<col>R with 1–5 digits in each number
// (section 5.4). A number can be 0.
type CPR struct{ Row, Col int }

// replies holds the replies of the probe. A reply to a query of section 5.7 is its
// Reply.Body; an empty string is no reply.
type replies struct {
	ttype      string      // the first TTYPE name, as sent
	ttypeGot   bool        // the first TTYPE name arrived
	naws       telnet.NAWS // the last NAWS without 0 in a dimension; zero is no size
	cpr        CPR         // the reply to the charset probe
	cprGot     bool        // the reply to the charset probe arrived
	decrqss    string
	xtversion  string
	da2        string
	da1        string
	palette    [paletteSize]uint32 // the colours of the OSC 4 replies as 0xRRGGBB
	paletteGot [paletteSize]bool   // the OSC 4 reply of the index arrived
}

// add takes the reply rp of step 4 of section 5.1. It drops a second reply
// and a reply that does not match section 5.7. It reports whether rp is the first
// reply to its query.
func (r *replies) add(rp Reply) bool {
	first := func(p *string) bool {
		if *p != "" {
			return false
		}
		*p = rp.Body
		return true
	}
	b := rp.Body
	switch rp.Intro {
	case '[':
		if c, ok := parseCPR(b); ok {
			if r.cprGot {
				return false
			}
			r.cpr, r.cprGot = c, true
			return true
		}
		params, ok := strings.CutSuffix(b, "c")
		switch {
		case !ok:
		case strings.HasPrefix(params, "?") && isParams(params[1:]):
			return first(&r.da1)
		case strings.HasPrefix(params, ">") && isParams(params[1:]):
			return first(&r.da2)
		}
	case 'P':
		switch {
		case b == "0$r" || isSGRReply(b):
			return first(&r.decrqss)
		case strings.HasPrefix(b, ">|"):
			return first(&r.xtversion)
		}
	case ']':
		n, c, ok := parseOSC4(b)
		if !ok || r.paletteGot[n] {
			return false
		}
		r.palette[n], r.paletteGot[n] = c, true
		return true
	}
	return false
}

// done reports whether the wait of step 4 ends: the DA1 reply arrived,
// or a CPR reply that means "not UTF-8" (section 5.1).
func (r *replies) done() bool {
	return r.da1 != "" || !r.utf8()
}

// utf8 reports whether the charset probe gives UTF-8 (section 5.4).
func (r *replies) utf8() bool {
	return !r.cprGot || r.cpr.Col != 3
}

// values gives the terminal values of section 6. utf8 is false when the charset probe
// gives "not UTF-8" (section 5.4).
func (r *replies) values() (t screen.Term, utf8 bool) {
	t = screen.Term{Charset: "utf-8"}
	t.Type, t.Colors = TypeColors(r.ttype)
	if isRGBReply(r.decrqss) || inClientTable(r.da2) {
		t.Colors = screen.ColorsTrue
	}
	if xt, ok := strings.CutPrefix(r.xtversion, ">|"); ok {
		t.Name = CleanName(xt)
	}
	if hasSixel(r.da1) {
		t.Images = screen.ImagesSixel
	}
	t.Cols, t.Rows = r.naws.Cols, r.naws.Rows
	if t.Cols == 0 {
		t.Cols, t.Rows = defaultCols, defaultRows
	}
	t.Cols, t.Rows = clamp(t.Cols, t.Rows)
	return t, r.utf8()
}

// logReply gives the reply b for the log line (section 5.6).
func logReply(b string) string {
	if b == "" {
		return "-"
	}
	return CleanName(b)
}

// logPalette gives the palette for the log line (section 5.6).
func (r *replies) logPalette() string {
	s := make([]string, paletteSize)
	for n := range s {
		s[n] = "-"
		if r.paletteGot[n] {
			s[n] = fmt.Sprintf("%06x", r.palette[n])
		}
	}
	return strings.Join(s, ",")
}

// parseCPR parses the body of a CPR reply (section 5.4).
func parseCPR(b string) (CPR, bool) {
	params, ok := strings.CutSuffix(b, "R")
	if !ok {
		return CPR{}, false
	}
	row, col, ok := strings.Cut(params, ";")
	if !ok || !isNumber(row) || !isNumber(col) {
		return CPR{}, false
	}
	r, _ := strconv.Atoi(row)
	c, _ := strconv.Atoi(col)
	return CPR{Row: r, Col: c}, true
}

// isNumber reports whether s has 1–5 decimal digits and nothing else (section 5.4).
func isNumber(s string) bool {
	return len(s) >= 1 && len(s) <= 5 && isDigits(s)
}

// isDigits reports whether s has one or more decimal digits and nothing else.
func isDigits(s string) bool {
	return s != "" && strings.Trim(s, "0123456789") == ""
}

// isParams reports whether s is numbers separated by ';' (section 5.7).
func isParams(s string) bool {
	for p := range strings.SplitSeq(s, ";") {
		if !isDigits(p) {
			return false
		}
	}
	return true
}

// sgrParams gives the parameters of a DECRQSS reply with 1$r, or false.
func sgrParams(b string) (string, bool) {
	p, ok := strings.CutPrefix(b, "1$r")
	if !ok {
		return "", false
	}
	p, ok = strings.CutSuffix(p, "m")
	return p, ok && strings.Trim(p, "0123456789;:") == ""
}

// isSGRReply reports whether b is the body of a DECRQSS reply with 1$r (section 5.7).
func isSGRReply(b string) bool {
	_, ok := sgrParams(b)
	return ok
}

// isRGBReply reports whether the DECRQSS reply b is the RGB reply (section 5.7).
func isRGBReply(b string) bool {
	p, ok := sgrParams(b)
	if !ok {
		return false
	}
	ps := strings.Split(p, ";")
	for i, s := range ps {
		if i+5 <= len(ps) && strings.Join(ps[i:i+5], ";") == "38;2;1;2;3" {
			return true
		}
		sub := strings.Split(s, ":")
		if s == "38:2:1:2:3" || len(sub) == 6 && sub[0] == "38" && sub[1] == "2" &&
			(sub[2] == "" || isDigits(sub[2])) && strings.Join(sub[3:], ":") == "1:2:3" {
			return true
		}
	}
	return false
}

// inClientTable reports whether the DA2 reply b is in the client table of section 6.
func inClientTable(b string) bool {
	p, ok := strings.CutPrefix(b, ">")
	if !ok {
		return false
	}
	p = strings.TrimSuffix(p, "c")
	class, _, _ := strings.Cut(p, ";")
	return p == "0;136;0" || class == "84"
}

// hasSixel reports whether the DA1 reply b has the feature 4 (section 6).
func hasSixel(b string) bool {
	p, ok := strings.CutPrefix(b, "?")
	if !ok {
		return false
	}
	ps := strings.Split(strings.TrimSuffix(p, "c"), ";")
	for _, f := range ps[1:] {
		if f == "4" {
			return true
		}
	}
	return false
}

// parseOSC4 parses the body of an OSC 4 reply (section 5.7): the index n
// and the colour c as 0xRRGGBB.
func parseOSC4(b string) (n int, c uint32, ok bool) {
	rest, ok := strings.CutPrefix(b, "4;")
	if !ok {
		return 0, 0, false
	}
	idx, spec, ok := strings.Cut(rest, ";")
	if !ok || !isNumber(idx) {
		return 0, 0, false
	}
	n, _ = strconv.Atoi(idx)
	spec, ok = strings.CutPrefix(spec, "rgb:")
	parts := strings.Split(spec, "/")
	if !ok || n >= paletteSize || len(parts) != 3 {
		return 0, 0, false
	}
	for _, s := range parts {
		if len(s) < 1 || len(s) > 4 {
			return 0, 0, false
		}
		v, err := strconv.ParseUint(s, 16, 16)
		if err != nil {
			return 0, 0, false
		}
		// round(v × 255 / (16^d − 1))
		maxV := uint64(1)<<(4*len(s)) - 1
		c = c<<8 | uint32((2*v*255+maxV)/(2*maxV))
	}
	return n, c, true
}
