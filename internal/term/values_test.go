package term

import (
	"slices"
	"strings"
	"testing"

	"newbbs/internal/screen"
	"newbbs/internal/telnet"
)

func TestTTYPELoop(t *testing.T) {
	// Section 5.3: the end after 4 names, or at a repeat of the first or the previous name.
	tests := []struct {
		in    []string
		names []string
		sends int // the number of names after which the host sends SEND again
	}{
		{[]string{"XTERM", "XTERM"}, []string{"XTERM"}, 1},
		{[]string{"A", "B", "B"}, []string{"A", "B"}, 2},
		{[]string{"A", "B", "C", "A"}, []string{"A", "B", "C"}, 3},
		{[]string{"A", "B", "C", "D"}, []string{"A", "B", "C", "D"}, 3},
		// After the end, the names are ignored.
		{[]string{"A", "B", "C", "D", "E"}, []string{"A", "B", "C", "D"}, 3},
		{[]string{"A", "A", "B"}, []string{"A"}, 1},
		// A name that repeats an older name in the middle does not end TTYPE.
		{[]string{"A", "B", "C", "B", "B"}, []string{"A", "B", "C", "B"}, 3},
	}
	for _, tt := range tests {
		var l ttypeLoop
		sends := 0
		for _, n := range tt.in {
			if l.add(n) {
				sends++
			}
		}
		if !slices.Equal(l.names, tt.names) || !l.ended || sends != tt.sends {
			t.Errorf("%q: names %q, ended %v, sends %d; want %q, true, %d",
				tt.in, l.names, l.ended, sends, tt.names, tt.sends)
		}
	}
}

func TestParseMTTS(t *testing.T) {
	ok := map[string]int{
		"MTTS 4": 4, "mtts 269": 269, "Mtts 0": 0, "MTTS 123456789": 123456789,
	}
	for in, want := range ok {
		if bits, ok := parseMTTS(in); !ok || bits != want {
			t.Errorf("parseMTTS(%q) = %d, %v; want %d, true", in, bits, ok, want)
		}
	}
	for _, in := range []string{
		"MTTS", "MTTS ", "MTTS 1234567890", "MTTS  4", "MTTS 4 ", "MTTS -4", "MTTS x",
		"MTTS4", "XTERM", "MTTſ 4", "MTTS\u00a04", "mtts\t4",
	} {
		if _, ok := parseMTTS(in); ok {
			t.Errorf("parseMTTS(%q) is ok, want not an MTTS name", in)
		}
	}
}

func TestCleanName(t *testing.T) {
	long := strings.Repeat("ы", 50)
	tests := map[string]string{
		"XTERM":                  "XTERM",
		"x\x01te\x1brm\u202e":    "xterm",
		"\x00\x07":               "",
		"ab\xffc":                "ab�c",
		long:                     strings.Repeat("ы", typeLen),
		strings.Repeat("é", 45): strings.Repeat("é", typeLen),
	}
	for in, want := range tests {
		if got := cleanName(in); got != want {
			t.Errorf("cleanName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestColorDepth(t *testing.T) {
	tests := []struct {
		names   []string
		mtts    int
		hasMTTS bool
		want    int
	}{
		// With MTTS.
		{nil, 4 | 256, true, ColorsTrue},
		{nil, 8 | 256, true, ColorsTrue},
		{nil, 8, true, Colors256},
		{nil, 4, true, Colors16},
		// MTTS wins over the names.
		{[]string{"xterm-direct"}, 4, true, Colors16},
		// Without MTTS: the largest depth of the names.
		{nil, 0, false, Colors16},
		{[]string{"xterm"}, 0, false, Colors16},
		{[]string{"xterm-256color"}, 0, false, Colors256},
		{[]string{"xterm", "xterm-256color"}, 0, false, Colors256},
		{[]string{"xterm-direct"}, 0, false, ColorsTrue},
		{[]string{"xterm-256color", "konsole-direct"}, 0, false, ColorsTrue},
		{[]string{"x-truecolor"}, 0, false, ColorsTrue},
		{[]string{"x24bit"}, 0, false, ColorsTrue},
		{[]string{"xterm-direct2"}, 0, false, Colors16},
	}
	for _, tt := range tests {
		if got := colorDepth(tt.names, tt.mtts, tt.hasMTTS); got != tt.want {
			t.Errorf("colorDepth(%q, %d, %v) = %d, want %d", tt.names, tt.mtts, tt.hasMTTS, got, tt.want)
		}
	}
}

func TestValues(t *testing.T) {
	tests := []struct {
		name string
		r    replies
		want screen.Term
		utf8 bool
	}{
		{"no replies", replies{},
			screen.Term{Type: "ansi", Cols: 80, Rows: 25, Charset: "utf-8", Colors: 16}, true},
		{"first terminal name in lower case",
			replies{ttype: ttypeLoop{names: []string{"MTTS 8", "\x01", "XTerm-256Color", "VT100"}}},
			screen.Term{Type: "xterm-256color", Cols: 80, Rows: 25, Charset: "utf-8", Colors: 256}, true},
		// The colour check uses the whole name, not the name cut to 40 graphemes.
		{"long name", replies{ttype: ttypeLoop{names: []string{strings.Repeat("x", 50) + "-DIRECT"}}},
			screen.Term{Type: strings.Repeat("x", 40), Cols: 80, Rows: 25, Charset: "utf-8", Colors: ColorsTrue}, true},
		{"colours from names without case",
			replies{ttype: ttypeLoop{names: []string{"XTERM", "XTERM-DIRECT"}}},
			screen.Term{Type: "xterm", Cols: 80, Rows: 25, Charset: "utf-8", Colors: ColorsTrue}, true},
		{"first MTTS name",
			replies{ttype: ttypeLoop{names: []string{"MTTS 256", "MTTS 8"}}},
			screen.Term{Type: "ansi", Cols: 80, Rows: 25, Charset: "utf-8", Colors: ColorsTrue}, true},
		{"NAWS size", replies{naws: telnet.NAWS{Cols: 132, Rows: 50}},
			screen.Term{Type: "ansi", Cols: 132, Rows: 50, Charset: "utf-8", Colors: 16}, true},
		{"NAWS wins over CPR",
			replies{naws: telnet.NAWS{Cols: 100, Rows: 30}, sizeProbe: true, cprs: []CPR{{40, 120}}},
			screen.Term{Type: "ansi", Cols: 100, Rows: 30, Charset: "utf-8", Colors: 16}, true},
		{"CPR size", replies{sizeProbe: true, cprs: []CPR{{40, 120}}},
			screen.Term{Type: "ansi", Cols: 120, Rows: 40, Charset: "utf-8", Colors: 16}, true},
		{"CPR size with a 0", replies{sizeProbe: true, cprs: []CPR{{0, 120}}},
			screen.Term{Type: "ansi", Cols: 80, Rows: 25, Charset: "utf-8", Colors: 16}, true},
		{"clamp low", replies{naws: telnet.NAWS{Cols: 5, Rows: 3}},
			screen.Term{Type: "ansi", Cols: 20, Rows: 10, Charset: "utf-8", Colors: 16}, true},
		{"clamp high", replies{sizeProbe: true, cprs: []CPR{{999, 999}}},
			screen.Term{Type: "ansi", Cols: 255, Rows: 100, Charset: "utf-8", Colors: 16}, true},
		{"charset column 2", replies{charset: true, cprs: []CPR{{1, 2}}},
			screen.Term{Type: "ansi", Cols: 80, Rows: 25, Charset: "utf-8", Colors: 16}, true},
		{"charset column 3", replies{charset: true, cprs: []CPR{{1, 3}}},
			screen.Term{Type: "ansi", Cols: 80, Rows: 25, Charset: "utf-8", Colors: 16}, false},
		{"charset no reply", replies{charset: true},
			screen.Term{Type: "ansi", Cols: 80, Rows: 25, Charset: "utf-8", Colors: 16}, true},
		{"charset other column", replies{charset: true, cprs: []CPR{{1, 4}}},
			screen.Term{Type: "ansi", Cols: 80, Rows: 25, Charset: "utf-8", Colors: 16}, true},
		{"charset and size", replies{charset: true, sizeProbe: true, cprs: []CPR{{5, 3}, {30, 90}}},
			screen.Term{Type: "ansi", Cols: 90, Rows: 30, Charset: "utf-8", Colors: 16}, false},
		// The first reply belongs to the first probe.
		{"one reply for two probes", replies{charset: true, sizeProbe: true, cprs: []CPR{{30, 90}}},
			screen.Term{Type: "ansi", Cols: 80, Rows: 25, Charset: "utf-8", Colors: 16}, true},
	}
	for _, tt := range tests {
		got, utf8 := tt.r.values()
		if got != tt.want || utf8 != tt.utf8 {
			t.Errorf("%s: values() = %+v, %v; want %+v, %v", tt.name, got, utf8, tt.want, tt.utf8)
		}
		if err := got.Check(); err != nil {
			t.Errorf("%s: %v", tt.name, err)
		}
	}
}
