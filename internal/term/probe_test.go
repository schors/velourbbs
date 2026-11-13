package term

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"net"
	"strings"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"newbbs/internal/screen"
	"newbbs/internal/telnet"
)

// The telnet bytes of the fake client.
const (
	iac  = 255
	sb   = 250
	se   = 240
	will = 251
	wont = 252
	do   = 253
	dont = 254
)

// dsr is the cursor position request of the charset probe.
var dsr = []byte("\x1b[6n")

// client is the script of a fake telnet client.
type client struct {
	will  []byte   // the options of the client side that it enables: DO gives WILL
	do    []byte   // the options of the host side that it accepts: WILL gives DO
	naws  []byte   // the NAWS data after WILL NAWS, 4 bytes; nil sends no NAWS
	ttype []string // the names that the client sends for each SEND, each in its own subnegotiation
	cpr   string   // the reply to each ESC[6n, for example "1;2"; "" gives no reply
	early string   // data that the client sends with its first reply
	quiet bool     // the client sends nothing
}

// fake runs a client on the client end of net.Pipe. A reader goroutine answers
// the host and a writer goroutine sends the answers, so the pipe cannot block both sides.
// The tests run in a synctest bubble: synctest.Wait returns when the reader
// has handled all output of the host.
type fake struct {
	cl     client
	c      net.Conn
	send   chan []byte
	mu     sync.Mutex
	out    []byte // all bytes from the host
	pos    int    // the parsed bytes of out
	nsend  int    // the SEND requests
	first  bool   // the first reply was sent
	paused bool   // the reader stops
}

// start connects a fake client with the script cl. It gives the host end.
func start(t *testing.T, cl client) (net.Conn, *fake) {
	t.Helper()
	host, cc := net.Pipe()
	f := &fake{cl: cl, c: cc, send: make(chan []byte, 100)}
	go f.read()
	go func() {
		for b := range f.send {
			if _, err := cc.Write(b); err != nil {
				return
			}
		}
	}()
	t.Cleanup(func() {
		_ = host.Close()
		_ = cc.Close()
	})
	return host, f
}

func (f *fake) read() {
	defer close(f.send)
	buf := make([]byte, 1<<16)
	for {
		n, err := f.c.Read(buf)
		f.mu.Lock()
		f.out = append(f.out, buf[:n]...)
		reply := f.parse()
		paused := f.paused
		f.mu.Unlock()
		if err != nil || paused {
			return
		}
		if len(reply) > 0 && !f.cl.quiet {
			f.send <- reply
		}
	}
}

// parse answers the complete requests in out after pos.
func (f *fake) parse() []byte {
	var reply []byte
	if !f.first {
		f.first = true
		reply = append(reply, f.cl.early...)
	}
	for f.pos < len(f.out) {
		rest := f.out[f.pos:]
		switch {
		case rest[0] == iac && len(rest) >= 3 && (rest[1] == do || rest[1] == will):
			reply = append(reply, f.negotiate(rest[1], rest[2])...)
			f.pos += 3
		case rest[0] == iac && len(rest) >= 3 && (rest[1] == dont || rest[1] == wont):
			f.pos += 3
		case rest[0] == iac && len(rest) >= 2 && rest[1] == sb:
			end := bytes.Index(rest, []byte{iac, se})
			if end < 0 {
				return reply
			}
			if bytes.Equal(rest[2:end], []byte{telnet.OptTTYPE, 1}) {
				f.nsend++
				for _, name := range f.cl.ttype {
					reply = append(reply, iac, sb, telnet.OptTTYPE, 0)
					reply = append(reply, name...)
					reply = append(reply, iac, se)
				}
			}
			f.pos += end + 2
		case rest[0] == iac && len(rest) >= 2 && rest[1] == iac:
			f.pos += 2
		case rest[0] == iac:
			if len(rest) < 3 {
				return reply
			}
			f.pos += 2 // another command
		case bytes.HasPrefix(rest, dsr):
			if f.cl.cpr != "" {
				reply = append(reply, "\x1b["+f.cl.cpr+"R"...)
			}
			f.pos += len(dsr)
		case bytes.HasPrefix(dsr, rest):
			return reply // an incomplete request
		default:
			f.pos++
		}
	}
	return reply
}

// negotiate answers DO or WILL for the option opt.
func (f *fake) negotiate(verb, opt byte) []byte {
	if verb == do {
		if bytes.IndexByte(f.cl.will, opt) < 0 {
			return []byte{iac, wont, opt}
		}
		b := []byte{iac, will, opt}
		if opt == telnet.OptNAWS && f.cl.naws != nil {
			b = append(b, iac, sb, telnet.OptNAWS)
			b = append(b, f.cl.naws...)
			b = append(b, iac, se)
		}
		return b
	}
	if bytes.IndexByte(f.cl.do, opt) < 0 {
		return []byte{iac, dont, opt}
	}
	return []byte{iac, do, opt}
}

// write sends b to the host.
func (f *fake) write(b string) {
	f.send <- []byte(b)
}

// output gives all bytes from the host after the reader handled them.
func (f *fake) output() string {
	synctest.Wait()
	f.mu.Lock()
	defer f.mu.Unlock()
	return string(f.out)
}

// pause stops the reader after its current read. The client then does not read.
func (f *fake) pause() {
	f.mu.Lock()
	f.paused = true
	f.mu.Unlock()
}

// naws gives the data of a NAWS subnegotiation.
func naws(cols, rows int) []byte {
	return []byte{byte(cols >> 8), byte(cols), byte(rows >> 8), byte(rows)}
}

var testTimeouts = Timeouts{Probe: 5 * time.Second, Idle: time.Minute, Output: time.Second}

// probeLog runs the probe and gives its result, the log and the probe time.
func probeLog(t *testing.T, cl client, to Timeouts) (*Terminal, *fake, string, time.Duration, error) {
	t.Helper()
	host, f := start(t, cl)
	var log bytes.Buffer
	start := time.Now()
	term, err := Probe(host, start, to, slog.New(slog.NewTextHandler(&log, nil)).With("addr", "192.0.2.1"))
	return term, f, log.String(), time.Since(start), err
}

// The options that the clients of section 3 accept.
var (
	allOpts  = []byte{telnet.OptNAWS, telnet.OptTTYPE, telnet.OptSGA, telnet.OptBinary}
	hostOpts = []byte{telnet.OptEcho, telnet.OptSGA, telnet.OptBinary}
)

func TestProbeClients(t *testing.T) {
	term := func(typ string, cols, rows, colors int) screen.Term {
		return screen.Term{Type: typ, Cols: cols, Rows: rows, Charset: "utf-8", Colors: colors}
	}
	tests := []struct {
		name string
		cl   client
		want screen.Term
	}{
		{"PuTTY",
			client{will: allOpts, do: hostOpts, naws: naws(120, 40), ttype: []string{"XTERM"}, cpr: "1;2"},
			term("xterm", 120, 40, 256)},
		{"Tera Term in line mode",
			client{
				will: []byte{telnet.OptNAWS, telnet.OptTTYPE, telnet.OptBinary},
				do:   []byte{telnet.OptEcho, telnet.OptBinary},
				naws: naws(80, 24), ttype: []string{"xterm"}, cpr: "1;2",
			},
			term("xterm", 80, 24, 256)},
		// The host takes the first name and ignores the later names.
		{"inetutils telnet",
			client{
				will: allOpts, do: hostOpts, naws: naws(200, 60),
				ttype: []string{"XTERM-256COLOR", "XTERM-DIRECT"}, cpr: "3;2",
			},
			term("xterm-256color", 200, 60, 256)},
		{"direct colour", client{will: allOpts, naws: naws(80, 25), ttype: []string{"xterm-direct"}, cpr: "1;2"},
			term("xterm-direct", 80, 25, screen.ColorsTrue)},
		// MTTS is a plain name.
		{"MTTS", client{will: allOpts, naws: naws(100, 30), ttype: []string{"MTTS 13"}, cpr: "1;2"},
			term("mtts 13", 100, 30, 256)},
		// A NAWS with 0 gives no size.
		{"NAWS with 0", client{will: allOpts, naws: naws(0, 24), ttype: []string{"VT100"}, cpr: "1;2"},
			term("vt100", 80, 25, 16)},
		{"no TTYPE", client{will: []byte{telnet.OptNAWS}, naws: naws(90, 30), cpr: "1;2"},
			term("ansi", 90, 30, 16)},
		// A key and a CPR reply arrive before step 4: the key goes to the queue
		// and the reply is dropped.
		{"early input",
			client{will: allOpts, naws: naws(90, 30), ttype: []string{"A"}, cpr: "1;2", early: "x\x1b[1;3R"},
			term("a", 90, 30, 256)},
	}
	for _, tt := range tests {
		synctest.Test(t, func(t *testing.T) {
			term, f, _, d, err := probeLog(t, tt.cl, testTimeouts)
			if err != nil {
				t.Fatalf("%s: %v", tt.name, err)
			}
			if d != 0 {
				t.Errorf("%s: the probe took %v; want no wait", tt.name, d)
			}
			if got := term.Term(); got != tt.want {
				t.Errorf("%s: Term() = %+v, want %+v", tt.name, got, tt.want)
			}
			// One SEND after WILL TTYPE, and one charset probe.
			out := f.output()
			if n, want := f.nsend, min(len(tt.cl.ttype), 1); n != want {
				t.Errorf("%s: %d SEND requests, want %d", tt.name, n, want)
			}
			if n := strings.Count(out, string(dsr)); n != 1 || !strings.Contains(out, charsetProbe) {
				t.Errorf("%s: output %q, want one charset probe", tt.name, out)
			}
			if tt.cl.early != "" {
				if k, err := term.Key(); k != "x" || err != nil {
					t.Errorf("%s: Key() = %q, %v; want x", tt.name, k, err)
				}
			}
		})
	}
}

func TestProbeStart(t *testing.T) {
	// Step 2: the requests of section 4.1 in one write, then TTYPE SEND after WILL TTYPE.
	synctest.Test(t, func(t *testing.T) {
		_, f, _, _, err := probeLog(t, client{will: allOpts, naws: naws(80, 24), ttype: []string{"A"}, cpr: "1;2"}, testTimeouts)
		if err != nil {
			t.Fatal(err)
		}
		if out := f.output(); !strings.HasPrefix(out, string(telnet.New().Start())) {
			t.Errorf("output %q does not start with the requests", out)
		}
	})
}

func TestProbeSendOnce(t *testing.T) {
	// The client enables TTYPE, disables it and enables it again. The host sends SEND once.
	synctest.Test(t, func(t *testing.T) {
		cl := client{
			will: allOpts, naws: naws(80, 24), ttype: []string{"A"}, cpr: "1;2",
			early: "\xff\xfb\x18\xff\xfc\x18\xff\xfb\x18",
		}
		term, f, _, _, err := probeLog(t, cl, testTimeouts)
		if err != nil {
			t.Fatal(err)
		}
		f.output()
		if f.nsend != 1 || term.Term().Type != "a" {
			t.Errorf("%d SEND requests, type %q; want 1, a", f.nsend, term.Term().Type)
		}
	})
}

func TestProbeLog(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		cl := client{will: allOpts, naws: naws(100, 30), ttype: []string{"X\x1bTERM-256color", "VT100"}, cpr: "1;2"}
		_, _, log, _, err := probeLog(t, cl, testTimeouts)
		if err != nil {
			t.Fatal(err)
		}
		want := "level=INFO msg=probe addr=192.0.2.1 ttype=XTERM-256color naws=100x30 cpr=1;2 " +
			"type=xterm-256color cols=100 rows=30 charset=utf-8 colors=256 time=0s\n"
		if _, line, _ := strings.Cut(log, " "); line != want {
			t.Errorf("log %q, want %q after the time", log, want)
		}
	})
}

func TestProbeNotUTF8(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		cl := client{will: []byte{telnet.OptTTYPE}, ttype: []string{"SYNCTERM"}, cpr: "1;3"}
		term, f, log, _, err := probeLog(t, cl, testTimeouts)
		if term != nil || !errors.Is(err, ErrNotUTF8) {
			t.Fatalf("Probe() = %v, %v; want ErrNotUTF8", term, err)
		}
		if out := f.output(); !strings.HasSuffix(out, notUTF8Text) {
			t.Errorf("output %q does not end with the text of section 5.5", out)
		}
		for _, s := range []string{"ttype=SYNCTERM", `charset="not utf-8"`, "cpr=1;3", "naws=-"} {
			if !strings.Contains(log, s) {
				t.Errorf("log %q has no %q", log, s)
			}
		}
	})
}

func TestProbeNoReplies(t *testing.T) {
	// probe_timeout ends the wait before the 1 s.
	synctest.Test(t, func(t *testing.T) {
		to := testTimeouts
		to.Probe = 300 * time.Millisecond
		term, f, log, d, err := probeLog(t, client{quiet: true}, to)
		if err != nil {
			t.Fatal(err)
		}
		if d != to.Probe {
			t.Errorf("the probe took %v; want %v", d, to.Probe)
		}
		want := screen.Term{Type: "ansi", Cols: 80, Rows: 25, Charset: "utf-8", Colors: 16}
		if got := term.Term(); got != want {
			t.Errorf("Term() = %+v, want %+v", got, want)
		}
		// After probe_timeout, the host sends no CPR probe.
		if out := f.output(); strings.Contains(out, string(dsr)) {
			t.Errorf("output %q has a CPR probe", out)
		}
		if !strings.Contains(log, "ttype=- naws=- cpr=-") {
			t.Errorf("log %q has no empty replies", log)
		}
	})
}

func TestProbeReplyWait(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// WILL NAWS without a NAWS: step 3 waits 1 s for the NAWS.
		// No reply to the charset probe: step 4 waits 1 s more.
		term, _, _, d, err := probeLog(t, client{will: []byte{telnet.OptNAWS}}, testTimeouts)
		if err != nil {
			t.Fatal(err)
		}
		if d != 2*replyWait {
			t.Errorf("the probe took %v; want %v", d, 2*replyWait)
		}
		if got := term.Term(); got.Cols != 80 || got.Rows != 25 {
			t.Errorf("Term() = %+v, want 80x25", got)
		}
	})
}

func TestProbeHangup(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		host, f := start(t, client{quiet: true})
		go func() { _ = f.c.Close() }()
		if _, err := Probe(host, time.Now(), testTimeouts, slog.New(slog.DiscardHandler)); !errors.Is(err, io.ErrClosedPipe) && !errors.Is(err, io.EOF) {
			t.Errorf("Probe() error = %v, want a closed connection", err)
		}
	})
}

func TestProbeLate(t *testing.T) {
	// probe_timeout passed before the probe (the PROXY header took the time).
	// The requests go out before Probe returns, so before the first screen.
	synctest.Test(t, func(t *testing.T) {
		host, f := start(t, client{quiet: true})
		term, err := Probe(host, time.Now().Add(-testTimeouts.Probe), testTimeouts, slog.New(slog.DiscardHandler))
		if err != nil {
			t.Fatal(err)
		}
		if out := f.output(); out != string(telnet.New().Start()) {
			t.Errorf("output %q, want only the requests", out)
		}
		if got := term.Term(); got.Cols != 80 || got.Rows != 25 {
			t.Errorf("Term() = %+v, want 80x25", got)
		}
	})
}
