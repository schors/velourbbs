package term

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/schors/velourbbs/internal/screen"
	"github.com/schors/velourbbs/internal/telnet"
)

// paletteSize is the number of palette queries: the OSC 4 indexes 0–15 (section 5.7).
const paletteSize = 16

// The OSC 4 indexes of the default foreground and background (section 5.7).
const (
	indexFg = 256
	indexBg = 258
)

// probeWrite is the write of step 4 of section 5.1: the CPR probes (section 5.4),
// the terminal queries in the order of section 5.7 and the erase.
var probeWrite = func() string {
	var b strings.Builder
	b.WriteString("\râ\x1b[6n")                            // charset probe
	b.WriteString("\rж\x1b[6n")                            // ambiguous probe
	b.WriteString("\x1b[38;2;1;2;3m\x1bP$qm\x1b\\\x1b[0m") // DECRQSS for SGR
	b.WriteString("\x1b[>0q")                              // XTVERSION
	b.WriteString("\x1b[>c")                               // DA2
	b.WriteString("\x1b]10;?\x1b\\")                       // OSC 10
	b.WriteString("\x1b]11;?\x1b\\")                       // OSC 11
	for n := range paletteSize {
		fmt.Fprintf(&b, "\x1b]4;%d;?\x1b\\", n) // OSC 4
	}
	fmt.Fprintf(&b, "\x1b]4;%d;?\x1b\\\x1b]4;%d;?\x1b\\", indexFg, indexBg)
	b.WriteString("\x1b[c")   // DA1
	b.WriteString("\r\x1b[K") // erase
	return b.String()
}()

// CPR is a cursor position report ESC[<row>;<col>R with 1–5 digits in each number
// (section 5.4). A number can be 0.
type CPR struct{ Row, Col int }

// colour is the colour of an OSC reply (section 5.7).
type colour struct {
	rgb uint32 // 0xRRGGBB
	got bool   // the reply arrived
}

// hex gives the colour as #rrggbb, or empty without a reply (section 6).
func (c colour) hex() string {
	if !c.got {
		return ""
	}
	return fmt.Sprintf("#%06x", c.rgb)
}

// log gives the colour for the log line: rrggbb, or - without a reply (section 5.6).
func (c colour) log() string {
	if !c.got {
		return "-"
	}
	return fmt.Sprintf("%06x", c.rgb)
}

// replies holds the replies of the probe. A reply to a query of section 5.7 is its
// Reply.Body; an empty string is no reply.
type replies struct {
	ttype     string      // the first TTYPE name, as sent
	ttypeGot  bool        // the first TTYPE name arrived
	naws      telnet.NAWS // the last NAWS without 0 in a dimension; zero is no size
	cpr       CPR         // the reply to the charset probe
	cprGot    bool        // the reply to the charset probe arrived
	amb       CPR         // the reply to the ambiguous probe
	ambGot    bool        // the reply to the ambiguous probe arrived
	decrqss   string
	xtversion string
	da2       string
	da1       string
	palette   [paletteSize]colour // the OSC 4 replies for 0–15
	fg, bg    colour              // the OSC 10 and OSC 11 replies
	fg4, bg4  colour              // the OSC 4 replies for 256 and 258
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
			switch {
			case !r.cprGot:
				r.cpr, r.cprGot = c, true
			case !r.ambGot:
				r.amb, r.ambGot = c, true
			default:
				return false
			}
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
		case rp.BEL: // a DCS reply ends with ST
		case b == "0$r" || isSGRReply(b):
			return first(&r.decrqss)
		case strings.HasPrefix(b, ">|"):
			return first(&r.xtversion)
		}
	case ']':
		p, rgb, ok := r.oscColour(b)
		if !ok || p.got {
			return false
		}
		*p = colour{rgb, true}
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
	t.Type, t.Colors = TypeColors(r.ttype)
	if isRGBReply(r.decrqss) || isPuTTY(r.da2) || isTmux(r.da2) {
		t.Colors = screen.ColorsTrue
	}
	if xt, ok := strings.CutPrefix(r.xtversion, ">|"); ok {
		t.Name = CleanName(xt)
	}
	t.Ambiguous = 1
	if r.ambGot && r.amb.Col == 3 {
		t.Ambiguous = 2
	}
	t.Cols, t.Rows = r.naws.Cols, r.naws.Rows
	if t.Cols == 0 {
		t.Cols, t.Rows = defaultCols, defaultRows
	}
	t.Cols, t.Rows = clamp(t.Cols, t.Rows)
	bg := r.bg
	if isPuTTY(r.da2) && !bg.got {
		bg = r.bg4
	}
	t.BgColor = bg.hex()
	t.Background = Background(t.BgColor)
	return t, r.utf8()
}

// logReply gives the reply b for the log line (section 5.6).
func logReply(b string) string {
	if b == "" {
		return "-"
	}
	return CleanName(b)
}

// notVGA reports whether the palette is not VGA: an OSC 4 reply has a colour
// other than the VGA colour of its index (section 5.7).
func (r *replies) notVGA() bool {
	for n := range paletteSize {
		if c := r.palette[n]; c.got && c.rgb != screen.Palette(uint32(n)) {
			return true
		}
	}
	return false
}

// logPalette gives the palette for the log line (section 5.6).
func (r *replies) logPalette() string {
	s := make([]string, paletteSize)
	for n, c := range r.palette {
		s[n] = c.log()
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

// daParams gives the numbers of the DA1 or DA2 reply b with the prefix prefix
// (section 5.7). A number too large for an int is -1.
func daParams(b, prefix string) []int {
	p, ok := strings.CutPrefix(b, prefix)
	if !ok {
		return nil
	}
	var ns []int
	for s := range strings.SplitSeq(strings.TrimSuffix(p, "c"), ";") {
		n, err := strconv.Atoi(s)
		if err != nil {
			n = -1
		}
		ns = append(ns, n)
	}
	return ns
}

// isPuTTY reports whether the DA2 reply b is the reply of PuTTY in the client table
// of section 6.
func isPuTTY(b string) bool {
	return slices.Equal(daParams(b, ">"), []int{0, 136, 0})
}

// isTmux reports whether the DA2 reply b is the reply of tmux in the client table
// of section 6.
func isTmux(b string) bool {
	ns := daParams(b, ">")
	return len(ns) > 0 && ns[0] == 84
}

// oscColour parses the body b of an OSC 4, OSC 10 or OSC 11 reply (section 5.7).
// It gives the place of the reply in r and the colour as 0xRRGGBB.
func (r *replies) oscColour(b string) (p *colour, rgb uint32, ok bool) {
	cmd, spec, _ := strings.Cut(b, ";")
	switch cmd {
	case "10":
		p = &r.fg
	case "11":
		p = &r.bg
	case "4":
		idx, rest, found := strings.Cut(spec, ";")
		if !found || !isNumber(idx) {
			return nil, 0, false
		}
		switch n, _ := strconv.Atoi(idx); {
		case n < paletteSize:
			p = &r.palette[n]
		case n == indexFg:
			p = &r.fg4
		case n == indexBg:
			p = &r.bg4
		default:
			return nil, 0, false
		}
		spec = rest
	default:
		return nil, 0, false
	}
	rgb, ok = parseRGB(spec)
	return p, rgb, ok
}

// parseRGB parses the colour rgb:<r>/<g>/<b> of an OSC reply (section 5.7)
// and gives it as 0xRRGGBB.
func parseRGB(spec string) (c uint32, ok bool) {
	spec, ok = strings.CutPrefix(spec, "rgb:")
	parts := strings.Split(spec, "/")
	if !ok || len(parts) != 3 {
		return 0, false
	}
	for _, s := range parts {
		if len(s) < 1 || len(s) > 4 {
			return 0, false
		}
		v, err := strconv.ParseUint(s, 16, 16)
		if err != nil {
			return 0, false
		}
		// round(v × 255 / (16^d − 1))
		maxV := uint64(1)<<(4*len(s)) - 1
		c = c<<8 | uint32((2*v*255+maxV)/(2*maxV))
	}
	return c, true
}
