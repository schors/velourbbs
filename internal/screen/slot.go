package screen

import (
	"strconv"
	"strings"
	"sync"
	"time"
)

// Slot is the run of cells that the render gives a live field on a page (live-screen, section 3.1).
type Slot struct {
	Row, Col int // the row in Page.Rows and the first cell of the slot in the row, from 0
	Width    int // the cells of the slot; a cut at W or at term.cols shows fewer
	// Skip is the cells of the slot before the run: a word wider than W breaks the slot over rows
	// (bbt-format, section 6.3). Cells is the cells of the run.
	Skip, Cells int
	Ref         Ref
	Layout      string // the layout of bbs.time of the token or of the alias, else ""
	Right       bool   // the value aligns right
}

// durationWidth is the largest width of session.online and bbs.uptime: the width of "9999d 23:59:59"
// (live-screen, section 3.1).
const durationWidth = 14

// slotWidth gives the largest width of the value of the live field ref with the layout layout
// (live-screen, section 3.1). ref ticks. A layout "" is the layout of Session.Time.
func (e *Env) slotWidth(ref Ref, layout string) int {
	switch ref.session {
	case bbsTime:
		if layout == "" {
			layout = e.Session.Layout
		}
		return layoutWidth(layout, e.Session.Clock.Location())
	case bbsSessions:
		return len(strconv.Itoa(e.Session.MaxConns))
	}
	return durationWidth
}

// layoutWidth gives the sum of the largest widths of the elements of the layout of the Go package time
// in the zone loc (live-screen, section 3.1). It splits the layout as the package time does.
func layoutWidth(layout string, loc *time.Location) int {
	w := 0
	for layout != "" {
		lit, n, elem, rest := nextElement(layout)
		for _, gw := range Graphemes(lit) {
			w += gw
		}
		switch {
		case elem == "MST":
			w += zoneWidth(loc)
		case n > 0:
			w += n
		}
		layout = rest
	}
	return w
}

// elementWidths gives the largest width of each element of the package time that has no special rule
// in element. A longer element comes before an element that is its prefix.
var elementWidths = []struct {
	text  string
	width int
}{
	{"-07:00:00", 9}, {"-070000", 7}, {"-07:00", 6}, {"-0700", 5}, {"-07", 3},
	{"Z07:00:00", 9}, {"Z070000", 7}, {"Z07:00", 6}, {"Z0700", 5}, {"Z07", 3},
	{"2006", 4}, {"002", 3}, {"__2", 3},
	{"01", 2}, {"02", 2}, {"03", 2}, {"04", 2}, {"05", 2}, {"06", 2}, {"15", 2}, {"_2", 2},
	{"PM", 2}, {"pm", 2},
	{"1", 2}, {"2", 2}, {"3", 2}, {"4", 2}, {"5", 2},
}

// nextElement finds the first element of the layout. It gives the literal text before it,
// the largest width n of the element, its text elem and the rest after it.
// The element "MST" has n 0. Without an element, lit is the layout.
// The rules follow the function nextStdChunk of the package time.
func nextElement(layout string) (lit string, n int, elem, rest string) {
	for i := range len(layout) {
		s := layout[i:]
		if strings.HasPrefix(s, "_2006") { // a literal "_", then the year
			return layout[:i+1], 4, "2006", s[5:]
		}
		if e, n, ok := element(s); ok {
			return layout[:i], n, e, s[len(e):]
		}
	}
	return layout, 0, "", ""
}

// element gives the element of the package time at the start of s and its largest width.
func element(s string) (string, int, bool) {
	switch {
	case strings.HasPrefix(s, "January"):
		return "January", 9, true
	case strings.HasPrefix(s, "Monday"):
		return "Monday", 9, true
	case strings.HasPrefix(s, "Jan") && !startsLower(s[3:]):
		return "Jan", 3, true
	case strings.HasPrefix(s, "Mon") && !startsLower(s[3:]):
		return "Mon", 3, true
	case strings.HasPrefix(s, "MST"):
		return "MST", 0, true
	case s[0] == '.' || s[0] == ',': // fractional seconds: a run of 0 or of 9 that no digit follows
		if len(s) < 2 || s[1] != '0' && s[1] != '9' {
			return "", 0, false
		}
		j := 2
		for j < len(s) && s[j] == s[1] {
			j++
		}
		if j < len(s) && '0' <= s[j] && s[j] <= '9' {
			return "", 0, false
		}
		return s[:j], j, true // the digits and the separator
	}
	for _, e := range elementWidths {
		if strings.HasPrefix(s, e.text) {
			return e.text, e.width, true
		}
	}
	return "", 0, false
}

// startsLower reports whether s starts with a lower-case ASCII letter.
func startsLower(s string) bool {
	return s != "" && 'a' <= s[0] && s[0] <= 'z'
}

// zoneWidths caches the result of zoneWidth by zone.
var zoneWidths sync.Map // *time.Location to int

// zoneWidth gives the width of the widest abbreviation of the zone loc from 2000 to 2100
// (live-screen, section 3.1). A period without an abbreviation shows the form -0700.
func zoneWidth(loc *time.Location) int {
	if w, ok := zoneWidths.Load(loc); ok {
		return w.(int)
	}
	w := 0
	end := time.Date(2101, 1, 1, 0, 0, 0, 0, loc)
	for t := time.Date(2000, 1, 1, 0, 0, 0, 0, loc); t.Before(end); {
		name, _ := t.Zone()
		if name == "" {
			name = "-0700"
		}
		n := 0
		for _, gw := range Graphemes(name) {
			n += gw
		}
		w = max(w, n)
		_, next := t.ZoneBounds()
		if !next.After(t) { // a zero next: the period has no end
			break
		}
		t = next
	}
	zoneWidths.Store(loc, w)
	return w
}
