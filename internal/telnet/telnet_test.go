package telnet

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

var t0 = time.Unix(1000, 0)

// cat joins byte slices and strings into one input.
func cat(parts ...any) []byte {
	var b []byte
	for _, p := range parts {
		switch v := p.(type) {
		case byte:
			b = append(b, v)
		case int:
			b = append(b, byte(v))
		case rune:
			b = append(b, byte(v))
		case string:
			b = append(b, v...)
		case []byte:
			b = append(b, v...)
		default:
			panic("cat: bad part")
		}
	}
	return b
}

func mustFeed(t *testing.T, p *Parser, in []byte) Result {
	t.Helper()
	r, err := p.Feed(t0, in)
	if err != nil {
		t.Fatalf("Feed(% x): %v", in, err)
	}
	return r
}

func TestStart(t *testing.T) {
	want := cat(
		cmdIAC, cmdWILL, OptEcho,
		cmdIAC, cmdWILL, OptSGA, cmdIAC, cmdDO, OptSGA,
		cmdIAC, cmdWILL, OptBinary, cmdIAC, cmdDO, OptBinary,
		cmdIAC, cmdDO, OptNAWS,
		cmdIAC, cmdDO, OptTTYPE,
	)
	if got := New().Start(); !bytes.Equal(got, want) {
		t.Errorf("Start() = % x, want % x", got, want)
	}
}

// TestAccept sends each request to a new parser: the host accepts only the options
// of section 4.1 and refuses each other option.
func TestAccept(t *testing.T) {
	accepted := map[[2]byte]bool{
		{cmdDO, OptEcho}: true, {cmdDO, OptSGA}: true, {cmdDO, OptBinary}: true,
		{cmdWILL, OptSGA}: true, {cmdWILL, OptBinary}: true,
		{cmdWILL, OptNAWS}: true, {cmdWILL, OptTTYPE}: true,
	}
	agree := map[byte]byte{cmdWILL: cmdDO, cmdDO: cmdWILL}
	refuse := map[byte]byte{cmdWILL: cmdDONT, cmdDO: cmdWONT}
	for _, verb := range []byte{cmdWILL, cmdDO} {
		for _, opt := range []byte{OptBinary, OptEcho, OptSGA, OptTTYPE, OptNAWS, 99} {
			r := mustFeed(t, New(), cat(cmdIAC, verb, opt))
			want, events := cat(cmdIAC, refuse[verb], opt), []Event(nil)
			if accepted[[2]byte{verb, opt}] {
				want = cat(cmdIAC, agree[verb], opt)
				events = []Event{Option{Opt: opt, Local: verb == cmdDO, On: true}}
			}
			if !bytes.Equal(r.Reply, want) || !reflect.DeepEqual(r.Events, events) {
				t.Errorf("IAC %d %d: reply = % x, events = %v, want % x, %v",
					verb, opt, r.Reply, r.Events, want, events)
			}
		}
	}
}

func TestOptions(t *testing.T) {
	const other = 99
	tests := []struct {
		name   string
		start  bool // call Start first
		pre    []byte
		in     []byte
		reply  []byte
		events []Event
	}{
		{
			name:   "accepted client option, disabled: enable and agree",
			in:     cat(cmdIAC, cmdWILL, OptNAWS),
			reply:  cat(cmdIAC, cmdDO, OptNAWS),
			events: []Event{Option{Opt: OptNAWS, On: true}},
		},
		{
			name:   "accepted host option, disabled: enable and agree",
			in:     cat(cmdIAC, cmdDO, OptEcho),
			reply:  cat(cmdIAC, cmdWILL, OptEcho),
			events: []Event{Option{Opt: OptEcho, Local: true, On: true}},
		},
		{
			name:  "client ECHO: refuse",
			in:    cat(cmdIAC, cmdWILL, OptEcho),
			reply: cat(cmdIAC, cmdDONT, OptEcho),
		},
		{
			name:  "other client option: refuse",
			in:    cat(cmdIAC, cmdWILL, other),
			reply: cat(cmdIAC, cmdDONT, other),
		},
		{
			name:  "host NAWS: refuse",
			in:    cat(cmdIAC, cmdDO, OptNAWS),
			reply: cat(cmdIAC, cmdWONT, OptNAWS),
		},
		{
			name:  "other host option: refuse",
			in:    cat(cmdIAC, cmdDO, other),
			reply: cat(cmdIAC, cmdWONT, other),
		},
		{
			name:   "answer WILL to DO of the host: no reply",
			start:  true,
			in:     cat(cmdIAC, cmdWILL, OptNAWS),
			events: []Event{Option{Opt: OptNAWS, On: true}},
		},
		{
			name:   "answer WONT to DO of the host: no reply",
			start:  true,
			in:     cat(cmdIAC, cmdWONT, OptTTYPE),
			events: []Event{Option{Opt: OptTTYPE}},
		},
		{
			name:   "answer DO to WILL of the host: no reply",
			start:  true,
			in:     cat(cmdIAC, cmdDO, OptEcho),
			events: []Event{Option{Opt: OptEcho, Local: true, On: true}},
		},
		{
			name:   "answer DONT to WILL of the host: no reply",
			start:  true,
			in:     cat(cmdIAC, cmdDONT, OptSGA),
			events: []Event{Option{Opt: OptSGA, Local: true}},
		},
		{
			name: "enable request in state YES: no reply",
			pre:  cat(cmdIAC, cmdWILL, OptNAWS, cmdIAC, cmdDO, OptEcho),
			in:   cat(cmdIAC, cmdWILL, OptNAWS, cmdIAC, cmdDO, OptEcho),
		},
		{
			name:   "client disables an enabled option: disable and agree",
			pre:    cat(cmdIAC, cmdWILL, OptNAWS),
			in:     cat(cmdIAC, cmdWONT, OptNAWS),
			reply:  cat(cmdIAC, cmdDONT, OptNAWS),
			events: []Event{Option{Opt: OptNAWS}},
		},
		{
			name:   "client disables an enabled host option: disable and agree",
			pre:    cat(cmdIAC, cmdDO, OptEcho),
			in:     cat(cmdIAC, cmdDONT, OptEcho),
			reply:  cat(cmdIAC, cmdWONT, OptEcho),
			events: []Event{Option{Opt: OptEcho, Local: true}},
		},
		{
			name: "disable request in state NO: no reply",
			in: cat(cmdIAC, cmdWONT, OptNAWS, cmdIAC, cmdDONT, OptEcho,
				cmdIAC, cmdWONT, other, cmdIAC, cmdDONT, other),
		},
		{
			name:  "after a refusal of the client, the option is disabled",
			pre:   cat(cmdIAC, cmdWONT, OptNAWS),
			start: true,
			in:    cat(cmdIAC, cmdWONT, OptNAWS),
		},
		{
			name:   "client enables an option later that it refused",
			start:  true,
			pre:    cat(cmdIAC, cmdWONT, OptNAWS),
			in:     cat(cmdIAC, cmdWILL, OptNAWS),
			reply:  cat(cmdIAC, cmdDO, OptNAWS),
			events: []Event{Option{Opt: OptNAWS, On: true}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := New()
			if tc.start {
				p.Start()
			}
			mustFeed(t, p, tc.pre)
			r := mustFeed(t, p, tc.in)
			if !bytes.Equal(r.Reply, tc.reply) {
				t.Errorf("reply = % x, want % x", r.Reply, tc.reply)
			}
			if !reflect.DeepEqual(r.Events, tc.events) {
				t.Errorf("events = %v, want %v", r.Events, tc.events)
			}
			if len(r.Data) != 0 {
				t.Errorf("data = % x, want none", r.Data)
			}
		})
	}
}

// confirm gives the answers of a naive client that confirms each request that it gets,
// also a request that matches its state.
func confirm(out []byte) []byte {
	answer := map[byte]byte{cmdWILL: cmdDO, cmdDO: cmdWILL, cmdWONT: cmdDONT, cmdDONT: cmdWONT}
	var b []byte
	for i := 0; i+2 < len(out); i++ {
		if out[i] == cmdIAC {
			if a, ok := answer[out[i+1]]; ok {
				b = append(b, cmdIAC, a, out[i+2])
				i += 2
			}
		}
	}
	return b
}

func TestNoLoop(t *testing.T) {
	p := New()
	out := p.Start()
	// The client also sends its own requests, accepted and refused, each twice.
	reqs := cat(
		cmdIAC, cmdWILL, OptEcho, cmdIAC, cmdDO, OptNAWS, cmdIAC, cmdWILL, 99, cmdIAC, cmdDO, 99,
		cmdIAC, cmdWILL, OptNAWS, cmdIAC, cmdDO, OptEcho, cmdIAC, cmdWONT, OptSGA,
	)
	in := cat(reqs, reqs)
	for round := 0; ; round++ {
		in = append(in, confirm(out)...)
		if len(in) == 0 {
			break
		}
		if round == 4 {
			t.Fatalf("the negotiation does not end: host sends % x", out)
		}
		out = mustFeed(t, p, in).Reply
		in = nil
	}
}

func TestCommands(t *testing.T) {
	tests := []struct {
		name  string
		in    []byte
		data  string
		reply []byte
	}{
		{"IAC IAC", cat("a", cmdIAC, cmdIAC, "b"), "a\xffb", nil},
		{"AYT", cat("a", cmdIAC, cmdAYT, "b"), "ab", []byte{0x07}},
		{
			"other commands",
			cat("a", cmdIAC, 241, cmdIAC, 242, cmdIAC, 245, cmdIAC, 247, cmdIAC, 248, cmdIAC, 249,
				cmdIAC, cmdSE, cmdIAC, 0, cmdIAC, "x", "b"),
			"ab", nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := mustFeed(t, New(), tc.in)
			if string(r.Data) != tc.data {
				t.Errorf("data = %q, want %q", r.Data, tc.data)
			}
			if !bytes.Equal(r.Reply, tc.reply) {
				t.Errorf("reply = % x, want % x", r.Reply, tc.reply)
			}
		})
	}
}

func TestHangup(t *testing.T) {
	for _, c := range []byte{cmdIP, cmdBRK} {
		if _, err := New().Feed(t0, cat("a", cmdIAC, c, "b")); !errors.Is(err, ErrHangup) {
			t.Errorf("IAC %d: err = %v, want ErrHangup", c, err)
		}
		in := cat(cmdIAC, cmdSB, OptNAWS, 0, cmdIAC, c, "b")
		if _, err := New().Feed(t0, in); !errors.Is(err, ErrHangup) {
			t.Errorf("IAC %d in a subnegotiation: err = %v, want ErrHangup", c, err)
		}
	}
}

func TestSub(t *testing.T) {
	sb := func(opt byte, data ...any) []byte {
		return cat(cmdIAC, cmdSB, opt, cat(data...), cmdIAC, cmdSE)
	}
	tests := []struct {
		name   string
		start  bool // the client enables NAWS and TTYPE first
		in     []byte
		data   string
		reply  []byte
		events []Event
	}{
		{
			name:   "NAWS",
			start:  true,
			in:     sb(OptNAWS, 0, 80, 0, 24),
			events: []Event{NAWS{Cols: 80, Rows: 24}},
		},
		{
			name:   "NAWS with IAC IAC",
			start:  true,
			in:     sb(OptNAWS, 1, cmdIAC, cmdIAC, 0, 0),
			events: []Event{NAWS{Cols: 511, Rows: 0}},
		},
		{name: "NAWS with 3 bytes", start: true, in: sb(OptNAWS, 0, 80, 0)},
		{name: "NAWS with 5 bytes", start: true, in: sb(OptNAWS, 0, 80, 0, 24, 0)},
		{name: "NAWS not enabled", in: sb(OptNAWS, 0, 80, 0, 24)},
		{
			name:   "TTYPE IS",
			start:  true,
			in:     sb(OptTTYPE, ttypeIS, "XTERM"),
			events: []Event{TTYPE{Name: "XTERM"}},
		},
		{
			name:   "TTYPE IS with an empty name",
			start:  true,
			in:     sb(OptTTYPE, ttypeIS),
			events: []Event{TTYPE{}},
		},
		{name: "TTYPE SEND from the client", start: true, in: sb(OptTTYPE, ttypeSEND)},
		{name: "TTYPE with no data", start: true, in: sb(OptTTYPE)},
		{name: "TTYPE not enabled", in: sb(OptTTYPE, ttypeIS, "XTERM")},
		{name: "other option", start: true, in: sb(OptSGA, 1, 2, 3)},
		{name: "no option byte", start: true, in: cat("a", cmdIAC, cmdSB, cmdIAC, cmdSE, "b"), data: "ab"},
		{
			name:  "IAC WILL drops the subnegotiation",
			start: true,
			in: cat(cmdIAC, cmdSB, OptTTYPE, ttypeIS, "ab", cmdIAC, cmdWILL, OptEcho,
				"cd", cmdIAC, cmdSE),
			data:  "cd",
			reply: cat(cmdIAC, cmdDONT, OptEcho),
		},
		{
			name:   "IAC SB starts a new subnegotiation",
			start:  true,
			in:     cat(cmdIAC, cmdSB, OptTTYPE, ttypeIS, "ab", sb(OptNAWS, 0, 80, 0, 24)),
			events: []Event{NAWS{Cols: 80, Rows: 24}},
		},
		{
			name:  "IAC AYT drops the subnegotiation",
			start: true,
			in:    cat(cmdIAC, cmdSB, OptNAWS, 0, 80, cmdIAC, cmdAYT, 0, 24, cmdIAC, cmdSE),
			data:  "\x00\x18",
			reply: []byte{0x07},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := New()
			if tc.start {
				mustFeed(t, p, cat(cmdIAC, cmdWILL, OptNAWS, cmdIAC, cmdWILL, OptTTYPE))
			}
			r := mustFeed(t, p, tc.in)
			if string(r.Data) != tc.data {
				t.Errorf("data = %q, want %q", r.Data, tc.data)
			}
			if !bytes.Equal(r.Reply, tc.reply) {
				t.Errorf("reply = % x, want % x", r.Reply, tc.reply)
			}
			if !reflect.DeepEqual(r.Events, tc.events) {
				t.Errorf("events = %v, want %v", r.Events, tc.events)
			}
		})
	}
}

func TestSubLimit(t *testing.T) {
	enable := cat(cmdIAC, cmdWILL, OptTTYPE)
	name := strings.Repeat("x", MaxSub-1)
	tests := []struct {
		name string
		in   []byte
		err  error
	}{
		{"256 bytes", cat(cmdIAC, cmdSB, OptTTYPE, ttypeIS, name, cmdIAC, cmdSE), nil},
		{"257 bytes", cat(cmdIAC, cmdSB, OptTTYPE, ttypeIS, name, "x", cmdIAC, cmdSE), ErrSubTooLong},
		{
			"256 bytes of IAC IAC",
			cat(cmdIAC, cmdSB, OptTTYPE, bytes.Repeat([]byte{cmdIAC, cmdIAC}, MaxSub), cmdIAC, cmdSE),
			nil,
		},
		{
			"257 bytes of IAC IAC",
			cat(cmdIAC, cmdSB, OptTTYPE, bytes.Repeat([]byte{cmdIAC, cmdIAC}, MaxSub+1)),
			ErrSubTooLong,
		},
		{
			"257 bytes of an option that is not enabled",
			cat(cmdIAC, cmdSB, 99, bytes.Repeat([]byte{0}, MaxSub+1)),
			ErrSubTooLong,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := New()
			mustFeed(t, p, enable)
			r, err := p.Feed(t0, tc.in)
			if !errors.Is(err, tc.err) {
				t.Fatalf("err = %v, want %v", err, tc.err)
			}
			if err == nil && tc.in[2] == OptTTYPE && tc.in[3] == ttypeIS {
				want := []Event{TTYPE{Name: name}}
				if !reflect.DeepEqual(r.Events, want) {
					t.Errorf("events = %v, want %v", r.Events, want)
				}
			}
		})
	}
}

func TestSubTimeout(t *testing.T) {
	p := New()
	if _, ok := p.Deadline(); ok {
		t.Fatal("Deadline: ok before a subnegotiation")
	}
	mustFeed(t, p, cat(cmdIAC, cmdWILL, OptNAWS))
	if _, err := p.Feed(t0, cat(cmdIAC, cmdSB, OptNAWS, 0)); err != nil {
		t.Fatal(err)
	}
	d, ok := p.Deadline()
	if !ok || !d.Equal(t0.Add(SubTimeout)) {
		t.Fatalf("Deadline() = %v, %v, want %v, true", d, ok, t0.Add(SubTimeout))
	}
	if _, err := p.Feed(d, cat(80)); err != nil {
		t.Fatalf("Feed at the deadline: %v", err)
	}
	if _, err := p.Feed(d.Add(time.Nanosecond), cat(0)); !errors.Is(err, ErrSubTimeout) {
		t.Fatalf("Feed after the deadline: err = %v, want ErrSubTimeout", err)
	}

	// A subnegotiation that ends closes the deadline.
	p = New()
	mustFeed(t, p, cat(cmdIAC, cmdSB, OptNAWS, 0, cmdIAC, cmdSE))
	if _, ok := p.Deadline(); ok {
		t.Fatal("Deadline: ok after the end of the subnegotiation")
	}
	if _, err := p.Feed(t0.Add(time.Hour), cat("a")); err != nil {
		t.Fatalf("Feed after the end: %v", err)
	}
}

// TestSplit feeds an IAC in the middle of a multibyte character and of an escape sequence,
// and splits the input at each byte.
func TestSplit(t *testing.T) {
	in := cat(
		cmdIAC, cmdWILL, OptNAWS,
		0xD0, cmdIAC, cmdAYT, 0x9F, // "П"
		0xD0, cmdIAC, cmdSB, OptNAWS, 0, 80, 0, 24, cmdIAC, cmdSE, 0x9F,
		0x1B, cmdIAC, 241, '[', cmdIAC, cmdIAC, 'A',
	)
	wantData := "ПП\x1b[\xffA"
	wantReply := cat(cmdIAC, cmdDO, OptNAWS, 0x07)
	wantEvents := []Event{Option{Opt: OptNAWS, On: true}, NAWS{Cols: 80, Rows: 24}}

	for _, step := range []int{len(in), 1, 2, 3} {
		p := New()
		var got Result
		for i := 0; i < len(in); i += step {
			r := mustFeed(t, p, in[i:min(i+step, len(in))])
			got.Data = append(got.Data, r.Data...)
			got.Reply = append(got.Reply, r.Reply...)
			got.Events = append(got.Events, r.Events...)
		}
		if string(got.Data) != wantData {
			t.Errorf("step %d: data = %q, want %q", step, got.Data, wantData)
		}
		if !bytes.Equal(got.Reply, wantReply) {
			t.Errorf("step %d: reply = % x, want % x", step, got.Reply, wantReply)
		}
		if !reflect.DeepEqual(got.Events, wantEvents) {
			t.Errorf("step %d: events = %v, want %v", step, got.Events, wantEvents)
		}
	}
}

func TestStartAnswered(t *testing.T) {
	// The client agrees to each request of the host: the host does not reply.
	p := New()
	p.Start()
	r := mustFeed(t, p, cat(
		cmdIAC, cmdDO, OptEcho, cmdIAC, cmdDO, OptSGA, cmdIAC, cmdDO, OptBinary,
		cmdIAC, cmdWILL, OptSGA, cmdIAC, cmdWILL, OptBinary,
		cmdIAC, cmdWILL, OptNAWS, cmdIAC, cmdWILL, OptTTYPE,
	))
	if len(r.Reply) != 0 || len(r.Events) != 7 {
		t.Errorf("reply = % x, events = %v, want no reply and 7 events", r.Reply, r.Events)
	}
}

func TestSendTTYPE(t *testing.T) {
	want := cat(cmdIAC, cmdSB, OptTTYPE, ttypeSEND, cmdIAC, cmdSE)
	if got := SendTTYPE(); !bytes.Equal(got, want) {
		t.Errorf("SendTTYPE() = % x, want % x", got, want)
	}
}

func TestEscape(t *testing.T) {
	tests := []struct{ in, want string }{
		{"", ""},
		{"abc", "abc"},
		{"a\xffb", "a\xff\xffb"},
		{"\xff\xff", "\xff\xff\xff\xff"},
	}
	for _, tc := range tests {
		if got := Escape([]byte("x"), []byte(tc.in)); string(got) != "x"+tc.want {
			t.Errorf("Escape(%q) = %q, want %q", tc.in, got, "x"+tc.want)
		}
	}
}

// FuzzFeed checks that input in one Feed and one byte in each Feed give the same result.
func FuzzFeed(f *testing.F) {
	f.Add(cat(cmdIAC, cmdWILL, OptNAWS, cmdIAC, cmdSB, OptNAWS, 0, 80, 0, 24, cmdIAC, cmdSE))
	f.Add(cat(cmdIAC, cmdWILL, OptTTYPE, cmdIAC, cmdSB, OptTTYPE, ttypeIS, "xterm", cmdIAC, cmdSE))
	f.Add(cat(cmdIAC, cmdDO, OptEcho, cmdIAC, cmdDONT, OptEcho, cmdIAC, cmdWONT, OptSGA, "a", cmdIAC, cmdIAC))
	f.Add(cat(cmdIAC, cmdWILL, OptTTYPE, cmdIAC, cmdSB, OptTTYPE, ttypeIS, strings.Repeat("x", MaxSub)))
	f.Add(cat(0xD0, cmdIAC, cmdSB, OptNAWS, 0, cmdIAC, cmdAYT, 0x9F, cmdIAC, cmdIP))
	f.Fuzz(func(t *testing.T, in []byte) {
		p := New()
		p.Start()
		want, wantErr := p.Feed(t0, in)

		p = New()
		p.Start()
		var got Result
		var err error
		for i := range in {
			var r Result
			r, err = p.Feed(t0, in[i:i+1])
			got.Data = append(got.Data, r.Data...)
			got.Reply = append(got.Reply, r.Reply...)
			got.Events = append(got.Events, r.Events...)
			if err != nil {
				break
			}
		}
		if err != wantErr || !bytes.Equal(got.Data, want.Data) || !bytes.Equal(got.Reply, want.Reply) ||
			!reflect.DeepEqual(got.Events, want.Events) {
			t.Fatalf("one byte in each Feed: %v, %v; one Feed: %v, %v", got, err, want, wantErr)
		}
		for _, e := range want.Events {
			if tt, ok := e.(TTYPE); ok && len(tt.Name) > MaxSub {
				t.Fatalf("TTYPE name has %d bytes", len(tt.Name))
			}
		}
	})
}
