package term

import (
	"slices"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/schors/velourbbs/internal/key"
	"github.com/schors/velourbbs/internal/screen"
)

var t0 = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// decode feeds in at t0 and then ends the wait.
func decode(in string) ([]string, []Reply) {
	var d Decoder
	keys, replies := d.Feed(t0, []byte(in))
	k, r := d.Feed(t0.Add(SeqWait), nil)
	return append(keys, k...), append(replies, r...)
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
	// Other control characters (bbt-format, section 2.1) are dropped.
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

// TestKeyNames checks that each key name of the decoder is a key name of bbt-format,
// section 8.1, in upper case: the menu matches the keys by these names.
func TestKeyNames(t *testing.T) {
	var names []string
	for _, k := range csiKeys {
		names = append(names, k)
	}
	for _, k := range ss3Keys {
		names = append(names, k)
	}
	// ESC alone ends at the timeout. ESC before another character ends at this character.
	for _, in := range []string{"\x1b", "\x1bx", "\r", "\x08", "\x7f", "\t", " ", "\x1b[[A", "\x1b[[B", "\x1b[[C", "\x1b[[D", "\x1b[[E"} {
		keys, _ := decode(in)
		if len(keys) == 0 {
			t.Fatalf("decode(%q) gives no key", in)
		}
		names = append(names, keys[0])
	}
	// Each name is a named key: key.Parse also accepts one letter or digit.
	for _, n := range names {
		if k, ok := key.Parse(n); !ok || k != n || len(n) < 2 {
			t.Errorf("key.Parse(%q) = %q, %v", n, k, ok)
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

	// The 128-byte limit, ESC included.
	n128 := "\x1b[" + strings.Repeat("1", 125) + "~"
	checkKeys(t, n128+"z", "z")
	long := "\x1b[" + strings.Repeat("1;", 70)
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
		// The wait restarts after each byte of a sequence: 150 ms after the byte
		// that starts it, 1 s after each later byte.
		{[]step{{0, "\x1b[1"}, {100 * ms, "5"}, {200 * ms, "~"}}, []string{"F5"}},
		{[]step{{0, "\x1b["}, {100 * ms, "["}, {200 * ms, "A"}}, []string{"F1"}},
		{[]step{{0, "\x1b["}, {149 * ms, "1"}, {1148 * ms, "5~"}}, []string{"F5"}},
		{[]step{{0, "\x1b[1"}, {999 * ms, "5~"}}, []string{"F5"}},
		{[]step{{0, "\x1b[["}, {999 * ms, "A"}}, []string{"F1"}},
		// A started sequence is dropped. The next bytes are new input.
		{[]step{{0, "\x1b["}, {150 * ms, "A"}}, []string{"A"}},
		{[]step{{0, "\x1b[1"}, {time.Second, ""}}, nil},
		{[]step{{0, "\x1b[1"}, {time.Second, "5~"}}, []string{"5", "~"}},
		{[]step{{0, "\x1bO"}, {150 * ms, "A"}}, []string{"A"}},
		{[]step{{0, "\x1b[["}, {time.Second, "A"}}, []string{"A"}},
		{[]step{{0, "\x1b[" + strings.Repeat("1", 140)}, {time.Second, "2"}}, []string{"2"}},
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
	// 1 s after a later byte of a sequence or a control string.
	for _, in := range []string{"\x1b[1", "\x1b[[", "\x1b]4", "\x1bPx\x1b", "\x1bP\xce", "\x1b]\xff", "\x1b[\xd1"} {
		d.Feed(t0, []byte(in))
		if dl, ok := d.Deadline(); !ok || !dl.Equal(t0.Add(SeqWait)) {
			t.Errorf("%q: Deadline = %v, %v; want %v, true", in, dl, ok, t0.Add(SeqWait))
		}
		d.Feed(t0.Add(SeqWait), nil)
	}
	// 150 ms after the byte that starts a control string.
	for _, in := range []string{"\x1b]", "\x1bP", "\x1b_"} {
		d.Feed(t0, []byte(in))
		if dl, ok := d.Deadline(); !ok || !dl.Equal(t0.Add(EscWait)) {
			t.Errorf("%q: Deadline = %v, %v; want %v, true", in, dl, ok, t0.Add(EscWait))
		}
		d.Feed(t0.Add(EscWait), nil)
	}
}

// Late replies to the queries of section 5.7.
var lateReplies = []string{
	"\x1b]4;1;rgb:cdcd/0000/0000\x1b\\",         // OSC 4 with ST
	"\x1b]4;15;rgb:ff/ff/ff\x07",                // OSC 4 with BEL
	"\x1bP1$r0;38:2::1:2:3m\x1b\\",              // DECRQSS
	"\x1bP0$r\x1b\\",                            // DECRQSS, not valid
	"\x1bP>|XTerm(412)\x1b\\",                   // XTVERSION
	"\x1bP>|tmux 3.6\x07",                       // XTVERSION with BEL
	"\x1b_Gi=1;OK\x1b\\",                        // APC
	"\x1bP>|Ωmega ы\x1b\\",                      // non-ASCII text
	"\x1bPΩ\x1b\\",                              // non-ASCII directly after the introducer
	"\x1b]\xff4\x07",                            // a byte that the UTF-8 decoder drops
	"\x1b]" + strings.Repeat("x", 200) + "\x07", // longer than 128 bytes
}

func TestControlString(t *testing.T) {
	for _, s := range lateReplies {
		checkKeys(t, s)
		checkKeys(t, "a"+s+"b", "a", "b")
		checkKeys(t, s+"\x1b[A", "UP")
	}
	// Other control characters do not end a control string.
	checkKeys(t, "\x1b]0;a\rb\x00c\x18d\x1a\x1b\\z", "z")
	// ESC followed by a byte other than \\ ends the string. ESC is new input.
	checkKeys(t, "\x1b]4;1;rgb\x1b[A", "UP")
	checkKeys(t, "\x1bP1$r\x1bx", "ESC", "x")
	checkKeys(t, "\x1b_abc\x1b\x1b\\", "ESC", "ESC", "\\")

	// A reply in two parts gives no keys: the parts arrive within 150 ms after
	// the introducer and within 1 s after a later byte.
	for _, s := range lateReplies {
		for i := 1; i < len(s); i++ {
			gap := SeqWait - time.Millisecond
			if i <= 2 {
				gap = EscWait - time.Millisecond
			}
			var d Decoder
			k1, _ := d.Feed(t0, []byte(s[:i]))
			k2, _ := d.Feed(t0.Add(gap), []byte(s[i:]+"z"))
			if got := append(k1, k2...); !slices.Equal(got, []string{"z"}) {
				t.Errorf("%q split at %d: %q, want z", s, i, got)
			}
		}
	}

	// Alt+], Alt+P and Alt+_ give no key. A key after 150 ms gives the key.
	for _, s := range []string{"\x1b]", "\x1bP", "\x1b_"} {
		var d Decoder
		k1, _ := d.Feed(t0, []byte(s))
		k2, _ := d.Feed(t0.Add(EscWait), []byte("a"))
		if got := append(k1, k2...); !slices.Equal(got, []string{"a"}) {
			t.Errorf("%q, a after 150 ms: %q, want a", s, got)
		}
	}
	// A control string that stops at its ESC is dropped at the timeout.
	checkKeys(t, "\x1b]4;1\x1b")
}

func TestReplies(t *testing.T) {
	cases := map[string][]Reply{
		"\x1b[12;80R":                       {{'[', "12;80R", false}},
		"\x1b[?61;4;6c":                     {{'[', "?61;4;6c", false}},
		"\x1b[>0;136;0c":                    {{'[', ">0;136;0c", false}},
		"\x1b[1;5A":                         {{'[', "1;5A", false}},
		"\x1b]4;1;rgb:cdcd/0000/0000\x1b\\": {{']', "4;1;rgb:cdcd/0000/0000", false}},
		"\x1b]4;15;rgb:ff/ff/ff\x07":        {{']', "4;15;rgb:ff/ff/ff", true}},
		"\x1bP1$r0;38:2::1:2:3m\x1b\\":      {{'P', "1$r0;38:2::1:2:3m", false}},
		"\x1bP>|Ωmega ы\x07":                {{'P', ">|Ωmega ы", true}},
		"\x1b_Gi=1;OK\x1b\\":                {{'_', "Gi=1;OK", false}},
		"\x1b]0;a\rb\x1b\\":                 {{']', "0;a\rb", false}},
		// Keys, a broken string and long input give no reply.
		"\x1b[A\x1bOP\x1b[[A":                       nil,
		"\x1b]4;1;rgb\x1bx":                         nil,
		"\x1b]" + strings.Repeat("x", 200) + "\x07": nil,
		"\x1b[" + strings.Repeat("1", maxSeq) + "c": nil,
		// The limit of 128 bytes counts the terminator.
		"\x1bP>|" + strings.Repeat("x", 122) + "\x1b\\": {{'P', ">|" + strings.Repeat("x", 122), false}},
		"\x1bP>|" + strings.Repeat("x", 123) + "\x1b\\": nil,
		"\x1b]" + strings.Repeat("x", 125) + "\x07":     {{']', strings.Repeat("x", 125), true}},
		"\x1b]" + strings.Repeat("x", 126) + "\x07":     nil,
		"\x1b[" + strings.Repeat("1", 125) + "c":        {{'[', strings.Repeat("1", 125) + "c", false}},
		"\x1b[" + strings.Repeat("1", 126) + "c":        nil,
	}
	for in, want := range cases {
		if _, got := decode(in); !slices.Equal(got, want) {
			t.Errorf("decode(%q) replies = %+v, want %+v", in, got, want)
		}
	}

	keys, replies := decode("a\x1b[3;2Rb\x1bP0$r\x1b\\c")
	if !slices.Equal(keys, []string{"a", "b", "c"}) || !slices.Equal(replies, []Reply{{'[', "3;2R", false}, {'P', "0$r", false}}) {
		t.Errorf("keys and replies: %q, %+v", keys, replies)
	}
}

// FuzzDecoderFeed checks that input in one Feed and one byte in each Feed at one time
// give the same keys and replies, and that each key and reply is valid.
func FuzzDecoderFeed(f *testing.F) {
	for _, s := range []string{
		"a\x1b[3;2Rb\x1b[24;80R", "\x1b[99999;1R\x1b[100000;1R", "\r\n\r\x00\n\r",
		"\x1b[A\x1bOP\x1b[[A\x1b[15~\x1b", "ы\xd0\x1b[1;2 R\x7f\t ", "\x1b[" + strings.Repeat("1", maxSeq),
		"a\x1b]4;1;rgb:cd/00/00\x07b\x1bP>|XTerm(412)\x1b\\c", "\x1b]4\x1bx\x1b_\x1b\x1b\\",
	} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, in []byte) {
		keys, replies := decode(string(in))

		var d Decoder
		var gotKeys []string
		var gotReplies []Reply
		for i := range in {
			k, r := d.Feed(t0, in[i:i+1])
			gotKeys, gotReplies = append(gotKeys, k...), append(gotReplies, r...)
		}
		k, r := d.Feed(t0.Add(SeqWait), nil)
		gotKeys, gotReplies = append(gotKeys, k...), append(gotReplies, r...)
		if !slices.Equal(gotKeys, keys) || !slices.Equal(gotReplies, replies) {
			t.Fatalf("one byte in each Feed: %q %+v; one Feed: %q %+v", gotKeys, gotReplies, keys, replies)
		}
		for _, k := range keys {
			if !isKey(k) {
				t.Fatalf("key %q", k)
			}
		}
		for _, r := range replies {
			if !strings.ContainsRune("[]P_", rune(r.Intro)) || len(r.Body) > maxSeq-2 {
				t.Fatalf("reply %+v", r)
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
