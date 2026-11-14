package term

import (
	"slices"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"newbbs/internal/screen"
)

var t0 = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// decode feeds in at t0 and then ends the wait.
func decode(in string) ([]string, []CPR) {
	var d Decoder
	keys, cprs := d.Feed(t0, []byte(in))
	k, c := d.Feed(t0.Add(EscWait), nil)
	return append(keys, k...), append(cprs, c...)
}

func checkKeys(t *testing.T, in string, want ...string) {
	t.Helper()
	if got, _ := decode(in); !slices.Equal(got, want) {
		t.Errorf("decode(%q) = %q, want %q", in, got, want)
	}
}

func TestKeyTable(t *testing.T) {
	// The table of section 8.1.
	checkKeys(t, "\x08", "BS")
	checkKeys(t, "\x7f", "BS")
	checkKeys(t, "\t", "TAB")
	checkKeys(t, " ", "SPACE")
	checkKeys(t, "\x1b", "ESC")
	// Other control characters (screen-format, section 2.1) are dropped.
	for _, c := range []string{"\x00", "\x01", "\x07", "\x1f", "\u0080", "\u009f", "\u202a", "\u202e", "\u2066", "\u2069"} {
		checkKeys(t, c)
	}
	// Other characters.
	for _, c := range []string{"a", "Z", "5", "?", "~", "ы", "Ы", "â", "€", "\ufffd", "😀"} {
		checkKeys(t, c, c)
	}
	checkKeys(t, "x\x01y", "x", "y")
}

func TestEnter(t *testing.T) {
	for in, n := range map[string]int{
		"\r\n": 1, "\r\x00": 1, "\r": 1, "\n": 1,
		"\r\n\n": 2, "\r\x00\x00": 1, "\r\r": 2, "\n\n": 2, "\n\r": 2, "\n\r\n": 2,
	} {
		checkKeys(t, in, slices.Repeat([]string{"ENTER"}, n)...)
	}
	checkKeys(t, "\rx\n", "ENTER", "x", "ENTER")
	checkKeys(t, "\x00\n", "ENTER")

	// CR at the end of one call and LF or NUL in the next.
	for _, b := range []string{"\n", "\x00"} {
		var d Decoder
		k1, _ := d.Feed(t0, []byte("\r"))
		k2, _ := d.Feed(t0.Add(time.Second), []byte(b+"a"))
		if got := append(k1, k2...); !slices.Equal(got, []string{"ENTER", "a"}) {
			t.Errorf("CR, then %q a: %q", b, got)
		}
	}
}

func TestUTF8(t *testing.T) {
	checkKeys(t, "a\xffb", "a", "b")
	checkKeys(t, "\xc3a", "a")            // an incomplete character
	checkKeys(t, "\xc0\xaf")              // overlong
	checkKeys(t, "\xed\xa0\x80")          // a surrogate
	checkKeys(t, "\xf4\x90\x80\x80")      // over U+10FFFF
	checkKeys(t, "\xe2\x82\r", "ENTER")   // an incomplete character before CR
	checkKeys(t, "\x80\xbf\xc3\xa4", "ä") // continuation bytes alone

	// A character split across calls.
	var d Decoder
	var got []string
	for _, b := range []byte("€") {
		k, _ := d.Feed(t0, []byte{b})
		got = append(got, k...)
	}
	if !slices.Equal(got, []string{"€"}) {
		t.Errorf("split €: %q", got)
	}
}

func TestEscapeTable(t *testing.T) {
	// The table of section 8.2, row by row.
	rows := map[string][]string{
		"UP":    {"\x1b[A", "\x1bOA"},
		"DOWN":  {"\x1b[B", "\x1bOB"},
		"RIGHT": {"\x1b[C", "\x1bOC"},
		"LEFT":  {"\x1b[D", "\x1bOD"},
		"HOME":  {"\x1b[H", "\x1bOH", "\x1b[1~", "\x1b[7~"},
		"END":   {"\x1b[F", "\x1bOF", "\x1b[4~", "\x1b[8~"},
		"DEL":   {"\x1b[3~"},
		"PGUP":  {"\x1b[5~"},
		"PGDN":  {"\x1b[6~"},
		"F1":    {"\x1bOP", "\x1b[11~", "\x1b[[A"},
		"F2":    {"\x1bOQ", "\x1b[12~", "\x1b[[B"},
		"F3":    {"\x1bOR", "\x1b[13~", "\x1b[[C"},
		"F4":    {"\x1bOS", "\x1b[14~", "\x1b[[D"},
		"F5":    {"\x1b[15~", "\x1b[[E"},
		"F6":    {"\x1b[17~"},
		"F7":    {"\x1b[18~"},
		"F8":    {"\x1b[19~"},
		"F9":    {"\x1b[20~"},
		"F10":   {"\x1b[21~"},
		"F11":   {"\x1b[23~"},
		"F12":   {"\x1b[24~"},
	}
	for key, seqs := range rows {
		for _, s := range seqs {
			checkKeys(t, s, key)
			checkKeys(t, "a"+s+"b", "a", key, "b")
		}
	}
}

func TestEscapeDrop(t *testing.T) {
	// Alt+letter and other bytes after ESC.
	checkKeys(t, "\x1bx", "ESC", "x")
	checkKeys(t, "\x1bы", "ESC", "ы")
	checkKeys(t, "\x1b\x1b", "ESC", "ESC")
	checkKeys(t, "\x1b\r\n", "ESC", "ENTER")
	checkKeys(t, "\x1b\x1b[A", "ESC", "UP")

	// Sequences that are not in the table are dropped as a whole.
	for _, s := range []string{
		"\x1b[1;5A", "\x1b[1;2P", "\x1b[2~", "\x1b[16~", "\x1b[25~", "\x1b[1A", "\x1b[?1;2c",
		"\x1b[ A", "\x1b[E", "\x1b[1[", "\x1bOx", "\x1bOE", "\x1bO~",
	} {
		checkKeys(t, s)
		checkKeys(t, s+"z", "z")
	}

	// An invalid byte ends the sequence and is new input.
	checkKeys(t, "\x1b[1\ra", "ENTER", "a")
	checkKeys(t, "\x1b[1ы", "ы")
	checkKeys(t, "\x1b[1\x1b[B", "DOWN")
	checkKeys(t, "\x1bO\r", "ENTER")
	checkKeys(t, "\x1bO1", "1")
	checkKeys(t, "\x1b[[F", "F")
	checkKeys(t, "\x1b[[a", "a")
	checkKeys(t, "\x1b[[\x1b[[E", "F5")

	// The 32-byte limit, ESC included.
	n32 := "\x1b[" + strings.Repeat("1", 29) + "~"
	checkKeys(t, n32+"z", "z")
	long := "\x1b[" + strings.Repeat("1;", 20)
	checkKeys(t, long+"A"+"z", "z")
	checkKeys(t, long+"\r", "ENTER")
	checkKeys(t, long+"\x1b[A", "UP")
}

func TestEscWait(t *testing.T) {
	type step struct {
		at time.Duration
		in string
	}
	feed := func(steps ...step) []string {
		var d Decoder
		var got []string
		for _, s := range steps {
			var in []byte
			if s.in != "" {
				in = []byte(s.in)
			}
			k, _ := d.Feed(t0.Add(s.at), in)
			got = append(got, k...)
		}
		return got
	}
	ms := time.Millisecond
	cases := []struct {
		steps []step
		want  []string
	}{
		{[]step{{0, "\x1b"}, {149 * ms, "[A"}}, []string{"UP"}},
		{[]step{{0, "\x1b"}, {150 * ms, "[A"}}, []string{"ESC", "[", "A"}},
		{[]step{{0, "\x1b"}, {150 * ms, ""}}, []string{"ESC"}},
		{[]step{{0, "\x1b"}, {100 * ms, ""}, {149 * ms, "x"}}, []string{"ESC", "x"}},
		// The wait restarts after each byte of a sequence.
		{[]step{{0, "\x1b[1"}, {100 * ms, "5"}, {200 * ms, "~"}}, []string{"F5"}},
		{[]step{{0, "\x1b["}, {100 * ms, "["}, {200 * ms, "A"}}, []string{"F1"}},
		// A started sequence is dropped. The next bytes are new input.
		{[]step{{0, "\x1b[1"}, {150 * ms, ""}}, nil},
		{[]step{{0, "\x1b[1"}, {150 * ms, "5~"}}, []string{"5", "~"}},
		{[]step{{0, "\x1bO"}, {150 * ms, "A"}}, []string{"A"}},
		{[]step{{0, "\x1b[["}, {150 * ms, "A"}}, []string{"A"}},
		{[]step{{0, "\x1b[" + strings.Repeat("1", 40)}, {150 * ms, "2"}}, []string{"2"}},
		// A byte of an incomplete character also restarts the wait.
		{[]step{{0, "\x1b"}, {100 * ms, "\xd1"}, {200 * ms, "\x8b"}}, []string{"ESC", "ы"}},
	}
	for i, c := range cases {
		if got := feed(c.steps...); !slices.Equal(got, c.want) {
			t.Errorf("case %d: %q, want %q", i, got, c.want)
		}
	}
}

func TestDeadline(t *testing.T) {
	var d Decoder
	if _, ok := d.Deadline(); ok {
		t.Error("Deadline: ok with no sequence")
	}
	d.Feed(t0, []byte("a\x1b["))
	if dl, ok := d.Deadline(); !ok || !dl.Equal(t0.Add(EscWait)) {
		t.Errorf("Deadline = %v, %v; want %v, true", dl, ok, t0.Add(EscWait))
	}
	d.Feed(t0.Add(100*time.Millisecond), nil) // no input does not restart the wait
	if dl, _ := d.Deadline(); !dl.Equal(t0.Add(EscWait)) {
		t.Errorf("Deadline after no input = %v", dl)
	}
	d.Feed(t0.Add(EscWait), nil)
	if _, ok := d.Deadline(); ok {
		t.Error("Deadline: ok after the wait")
	}
}

func TestCPR(t *testing.T) {
	cases := map[string][]CPR{
		"\x1b[12;80R":    {{12, 80}},
		"\x1b[0;0R":      {{0, 0}},
		"\x1b[1;5R":      {{1, 5}},
		"\x1b[99999;1R":  {{99999, 1}},
		"\x1b[007;02R":   {{7, 2}},
		"\x1b[100000;1R": nil,
		"\x1b[1;100000R": nil,
		"\x1b[;5R":       nil,
		"\x1b[5;R":       nil,
		"\x1b[5R":        nil,
		"\x1b[1;2;3R":    nil,
		"\x1b[?1;2R":     nil,
		"\x1b[1:2R":      nil,
		"\x1b[1;2 R":     nil,
		"\x1bOR":         nil,
	}
	for in, want := range cases {
		keys, got := decode(in)
		if !slices.Equal(got, want) {
			t.Errorf("decode(%q) CPR = %v, want %v", in, got, want)
		}
		if in != "\x1bOR" && len(keys) > 0 {
			t.Errorf("decode(%q) keys = %q, want none", in, keys)
		}
	}

	keys, cprs := decode("a\x1b[3;2Rb\x1b[24;80R")
	if !slices.Equal(keys, []string{"a", "b"}) || !slices.Equal(cprs, []CPR{{3, 2}, {24, 80}}) {
		t.Errorf("keys and CPR: %q, %v", keys, cprs)
	}
}

// FuzzDecoderFeed checks that input in one Feed and one byte in each Feed at one time
// give the same keys and CPRs, and that each key and CPR is valid.
func FuzzDecoderFeed(f *testing.F) {
	for _, s := range []string{
		"a\x1b[3;2Rb\x1b[24;80R", "\x1b[99999;1R\x1b[100000;1R", "\r\n\r\x00\n\r",
		"\x1b[A\x1bOP\x1b[[A\x1b[15~\x1b", "ы\xd0\x1b[1;2 R\x7f\t ", "\x1b[" + strings.Repeat("1", maxSeq),
	} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, in []byte) {
		keys, cprs := decode(string(in))

		var d Decoder
		var gotKeys []string
		var gotCPRs []CPR
		for i := range in {
			k, c := d.Feed(t0, in[i:i+1])
			gotKeys, gotCPRs = append(gotKeys, k...), append(gotCPRs, c...)
		}
		k, c := d.Feed(t0.Add(EscWait), nil)
		gotKeys, gotCPRs = append(gotKeys, k...), append(gotCPRs, c...)
		if !slices.Equal(gotKeys, keys) || !slices.Equal(gotCPRs, cprs) {
			t.Fatalf("one byte in each Feed: %q %v; one Feed: %q %v", gotKeys, gotCPRs, keys, cprs)
		}
		for _, k := range keys {
			if !isKey(k) {
				t.Fatalf("key %q", k)
			}
		}
		for _, c := range cprs {
			if c.Row < 0 || c.Row > 99999 || c.Col < 0 || c.Col > 99999 {
				t.Fatalf("CPR %v", c)
			}
		}
	})
}

// isKey reports whether k is a key name or one character that is not a control
// character (section 8.1).
func isKey(k string) bool {
	names := map[string]bool{"ENTER": true, "BS": true, "TAB": true, "SPACE": true, "ESC": true}
	for _, n := range csiKeys {
		names[n] = true
	}
	for _, n := range ss3Keys {
		names[n] = true
	}
	r, size := utf8.DecodeRuneInString(k)
	return names[k] || utf8.ValidString(k) && size == len(k) && size > 0 && r != ' ' && !screen.IsControl(r)
}
