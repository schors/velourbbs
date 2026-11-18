package term

import (
	"slices"
	"strings"
	"testing"
)

func TestParseCPR(t *testing.T) {
	cases := map[string]*CPR{
		"12;80R":    {12, 80},
		"0;0R":      {0, 0},
		"99999;1R":  {99999, 1},
		"007;02R":   {7, 2},
		"100000;1R": nil,
		"1;100000R": nil,
		";5R":       nil,
		"5;R":       nil,
		"5R":        nil,
		"1;2;3R":    nil,
		"?1;2R":     nil,
		"1:2R":      nil,
		"1;2 R":     nil,
		"1;2c":      nil,
	}
	for in, want := range cases {
		got, ok := parseCPR(in)
		if ok != (want != nil) || ok && got != *want {
			t.Errorf("parseCPR(%q) = %v, %v; want %v", in, got, ok, want)
		}
	}
}

func TestRGBReply(t *testing.T) {
	// Section 5.7: the forms of the RGB reply.
	for _, b := range []string{
		"1$r0;38:2::1:2:3m", "1$r0;38:2:1:1:2:3m", "1$r38:2:1:2:3m", "1$r0;38;2;1;2;3m",
		"1$r38;2;1;2;3m", "1$r1;38:2:12:1:2:3;4m",
	} {
		if !isRGBReply(b) {
			t.Errorf("isRGBReply(%q) = false", b)
		}
	}
	for _, b := range []string{
		"1$r0;30m", "1$r0;38:5:16m", "0$r", "1$r0;38;2;1;2m", "1$r0;38;2;1;2;4m",
		"1$r0;38:2:x:1:2:3m", "1$r0;38:2::1:2:3", "1$r48:2::1:2:3m", "1$r0;38;2:1;2;3m",
		"1$r0;38:2::1:2:3:4m", "2$r0;38:2::1:2:3m",
	} {
		if isRGBReply(b) {
			t.Errorf("isRGBReply(%q) = true", b)
		}
	}
}

func TestOSC4(t *testing.T) {
	cases := map[string]struct {
		n  int
		c  uint32
		ok bool
	}{
		"4;0;rgb:0000/0000/0000": {0, 0x000000, true},
		"4;15;rgb:ffff/ff/f":     {15, 0xffffff, true},
		"4;1;rgb:8080/80/8":      {1, 0x808088, true},
		"4;3;rgb:AAAA/5555/0":    {3, 0xaa5500, true},
		"4;7;rgb:7/77/777":       {7, 0x777777, true},
		"4;16;rgb:0/0/0":         {},
		"4;1;rgb:0/0":            {},
		"4;1;rgb:0/0/0/0":        {},
		"4;1;rgb:00000/0/0":      {},
		"4;1;rgb://0":            {},
		"4;1;rgb:g/0/0":          {},
		"4;1;rgb:+f/0/0":         {},
		"4;1;rgbi:1/0/0":         {},
		"4;;rgb:0/0/0":           {},
		"4;1":                    {},
		"10;rgb:0/0/0":           {},
	}
	for in, want := range cases {
		n, c, ok := parseOSC4(in)
		if ok != want.ok || ok && (n != want.n || c != want.c) {
			t.Errorf("parseOSC4(%q) = %d, %06x, %v; want %d, %06x, %v", in, n, c, ok, want.n, want.c, want.ok)
		}
	}
}

func TestRepliesAdd(t *testing.T) {
	var r replies
	add := func(intro byte, body string, want bool) {
		t.Helper()
		if got := r.add(Reply{intro, body, false}); got != want {
			t.Errorf("add(%c %q) = %v, want %v", intro, body, got, want)
		}
	}
	// A reply that does not match section 5.7 is dropped.
	for _, b := range []string{"?c", "?1;;2c", ">c", ">1;xc", "1;5A", "?1;2", "=1c"} {
		add('[', b, false)
	}
	for _, b := range []string{"1$r", "1$rxm", "2$r", "0$rm", "|x", "+r", ""} {
		add('P', b, false)
	}
	add('_', ">|x", false)
	// A DCS reply ends with ST. An OSC 4 reply can end with BEL.
	if r.add(Reply{'P', "1$r0;38;2;1;2;3m", true}) || r.add(Reply{'P', ">|x", true}) {
		t.Error("add() takes a DCS reply with BEL")
	}
	add(']', "11;rgb:0/0/0", false)
	if r != (replies{}) {
		t.Fatalf("replies after invalid replies: %+v", r)
	}

	// The first reply to each query is a reply. A second reply is dropped.
	add('[', "1;2R", true)
	add('[', "1;3R", false)
	add('P', "0$r", true)
	add('P', "1$r0;38:2::1:2:3m", false)
	add('P', ">|", true)
	add('P', ">|x", false)
	add('[', ">0;136;0c", true)
	add('[', ">84c", false)
	add(']', "4;2;rgb:0/0/0", true)
	add(']', "4;2;rgb:f/f/f", false)
	if !r.add(Reply{']', "4;3;rgb:f/f/f", true}) {
		t.Error("add() drops an OSC 4 reply with BEL")
	}
	if r.done() {
		t.Error("done() before DA1")
	}
	add('[', "?1;4c", true)
	add('[', "?1c", false)
	if !r.done() {
		t.Error("done() = false after DA1")
	}
	if r.cpr != (CPR{1, 2}) || r.decrqss != "0$r" || r.xtversion != ">|" || r.da2 != ">0;136;0c" || r.da1 != "?1;4c" {
		t.Errorf("replies %+v", r)
	}
	if got := r.logPalette(); got != "-,-,000000,ffffff,-,-,-,-,-,-,-,-,-,-,-,-" {
		t.Errorf("logPalette() = %q", got)
	}
	// The DA2 reply of PuTTY gives 16777216. The empty XTVERSION text is no name.
	if tm, _ := r.values(); tm.Colors != 16777216 || tm.Name != "" || tm.Images != "sixel" {
		t.Errorf("values() = %+v", tm)
	}

	// A CPR with the column 3 ends the wait.
	r = replies{}
	add('[', "1;3R", true)
	if !r.done() {
		t.Error("done() = false after a CPR with the column 3")
	}
}

func TestQueryValues(t *testing.T) {
	tests := []struct {
		r            replies
		colors       int
		name, images string
	}{
		{replies{}, 16, "", ""},
		{replies{ttype: "xterm-256color", decrqss: "1$r0;30m"}, 256, "", ""},
		{replies{ttype: "xterm", decrqss: "1$r0;38:2::1:2:3m"}, 16777216, "", ""},
		{replies{ttype: "vt100", da2: ">0;136;0c"}, 16777216, "", ""},
		{replies{ttype: "vt100", da2: ">0;136;1c"}, 16, "", ""},
		{replies{ttype: "screen", da2: ">84;0;0c"}, 16777216, "", ""},
		{replies{ttype: "screen", da2: ">841;0;0c"}, 256, "", ""},
		// The parameters are numbers.
		{replies{ttype: "screen", da2: ">084;0;0c"}, 16777216, "", ""},
		{replies{ttype: "vt100", da2: ">00;136;000c"}, 16777216, "", ""},
		{replies{da1: "?61;04c"}, 16, "", "sixel"},
		{replies{da1: "?61;99999999999999999999;14c"}, 16, "", ""},
		{replies{xtversion: ">|\x1b[1m" + strings.Repeat("x", 50)}, 16, "[1m" + strings.Repeat("x", 37), ""},
		{replies{xtversion: ">|\x01"}, 16, "", ""},
		{replies{da1: "?4c"}, 16, "", ""},
		{replies{da1: "?62;14;4c"}, 16, "", "sixel"},
		{replies{da1: "?62;44c"}, 16, "", ""},
	}
	for _, tt := range tests {
		got, _ := tt.r.values()
		if got.Colors != tt.colors || got.Name != tt.name || got.Images != tt.images {
			t.Errorf("values(%+v) = %d %q %q, want %d %q %q",
				tt.r, got.Colors, got.Name, got.Images, tt.colors, tt.name, tt.images)
		}
	}
	if s := slices.Collect(strings.SplitSeq(new(replies).logPalette(), ",")); len(s) != paletteSize {
		t.Errorf("logPalette() of no replies has %d items", len(s))
	}
}

// TestNotVGA: the palette is not VGA when an OSC 4 reply differs from the VGA
// colour of its index (section 5.7). No reply gives a VGA palette.
func TestNotVGA(t *testing.T) {
	tests := []struct {
		name    string
		replies []string
		want    bool
	}{
		{"no reply", nil, false},
		{"VGA", []string{"4;0;rgb:00/00/00", "4;1;rgb:aaaa/0000/0000", "4;15;rgb:f/f/f"}, false},
		{"not VGA", []string{"4;0;rgb:00/00/00", "4;1;rgb:c5/0f/1f"}, true},
		{"not VGA at the last index", []string{"4;15;rgb:f2/f2/f2"}, true},
	}
	for _, tt := range tests {
		var r replies
		for _, b := range tt.replies {
			if !r.add(Reply{']', b, true}) {
				t.Fatalf("%s: add(%q) = false", tt.name, b)
			}
		}
		if got := r.notVGA(); got != tt.want {
			t.Errorf("%s: notVGA() = %v, want %v", tt.name, got, tt.want)
		}
	}
}
