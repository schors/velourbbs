package term

import (
	"strings"
	"testing"

	"github.com/schors/velourbbs/internal/screen"
	"github.com/schors/velourbbs/internal/telnet"
)

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
		if got := CleanName(in); got != want {
			t.Errorf("CleanName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestColorDepth(t *testing.T) {
	// Section 6: the first rule that matches term.type.
	tests := map[string]int{
		// Rule 1.
		"xterm-direct": screen.ColorsTrue, "konsole-truecolor": screen.ColorsTrue,
		"x24bit": screen.ColorsTrue, "xterm-direct-256color": screen.Colors256,
		// Rule 2.
		"xterm-256color": screen.Colors256, "screen-256color-bce": screen.Colors256,
		// Rule 3.
		"ansi": screen.Colors16, "dumb": screen.Colors16, "vt52": screen.Colors16,
		"vt100": screen.Colors16, "vt102": screen.Colors16, "vt220": screen.Colors16,
		"vt320": screen.Colors16, "rxvt": screen.Colors16, "xterm-color": screen.Colors16,
		"eterm-color": screen.Colors16, "cygwin": screen.Colors16,
		"xterm-16color": screen.Colors16, "rxvt-8color": screen.Colors16,
		// Rule 4.
		"xterm": screen.Colors256, "linux": screen.Colors256, "syncterm": screen.Colors256,
		"mtts 13": screen.Colors256, "vt100-am": screen.Colors256, "xterm-direct2": screen.Colors256,
	}
	for typ, want := range tests {
		if got := colorDepth(typ); got != want {
			t.Errorf("colorDepth(%q) = %d, want %d", typ, got, want)
		}
	}
}

func TestTypeColors(t *testing.T) {
	tests := []struct {
		name, typ string
		colors    int
	}{
		{"VT100", "vt100", screen.Colors16},
		{"XTERM-256COLOR", "xterm-256color", screen.Colors256},
		{"", "ansi", screen.Colors16},
		{"\x01\x1b", "ansi", screen.Colors16},
	}
	for _, tt := range tests {
		if typ, colors := TypeColors(tt.name); typ != tt.typ || colors != tt.colors {
			t.Errorf("TypeColors(%q) = %q, %d; want %q, %d", tt.name, typ, colors, tt.typ, tt.colors)
		}
	}
}

func TestValues(t *testing.T) {
	term := func(typ string, cols, rows, colors int) screen.Term {
		return screen.Term{Type: typ, Cols: cols, Rows: rows, Charset: "utf-8", Colors: colors, Ambiguous: 1}
	}
	wide := term("ansi", 80, 24, 16)
	wide.Ambiguous = 2
	tests := []struct {
		name string
		r    replies
		want screen.Term
		utf8 bool
	}{
		{"no replies", replies{}, term("ansi", 80, 24, 16), true},
		{"name in lower case", replies{ttype: "XTERM", ttypeGot: true}, term("xterm", 80, 24, 256), true},
		{"name in upper case for 16", replies{ttype: "VT100", ttypeGot: true}, term("vt100", 80, 24, 16), true},
		{"empty name", replies{ttype: "\x01", ttypeGot: true}, term("ansi", 80, 24, 16), true},
		// The colour rules apply to term.type: the name after the cut to 40 graphemes.
		{"long name", replies{ttype: strings.Repeat("x", 50) + "-DIRECT", ttypeGot: true},
			term(strings.Repeat("x", 40), 80, 24, 256), true},
		{"NAWS size", replies{naws: telnet.NAWS{Cols: 132, Rows: 50}}, term("ansi", 132, 50, 16), true},
		{"clamp low", replies{naws: telnet.NAWS{Cols: 5, Rows: 3}}, term("ansi", 20, 10, 16), true},
		{"clamp high", replies{naws: telnet.NAWS{Cols: 999, Rows: 999}}, term("ansi", 255, 100, 16), true},
		{"charset column 2", replies{cpr: CPR{1, 2}, cprGot: true}, term("ansi", 80, 24, 16), true},
		{"charset column 3", replies{cpr: CPR{1, 3}, cprGot: true}, term("ansi", 80, 24, 16), false},
		{"charset other column", replies{cpr: CPR{1, 4}, cprGot: true}, term("ansi", 80, 24, 16), true},
		{"ambiguous column 3", replies{cpr: CPR{1, 2}, cprGot: true, amb: CPR{1, 3}, ambGot: true}, wide, true},
		{"ambiguous column 2", replies{cpr: CPR{1, 2}, cprGot: true, amb: CPR{1, 2}, ambGot: true}, term("ansi", 80, 24, 16), true},
		{"ambiguous other column", replies{cpr: CPR{1, 2}, cprGot: true, amb: CPR{1, 4}, ambGot: true}, term("ansi", 80, 24, 16), true},
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
