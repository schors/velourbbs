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

// dsr is the cursor position request of the CPR probes.
var dsr = []byte("\x1b[6n")

// client is the script of a fake telnet client.
type client struct {
	will  []byte   // the options of the client side that it enables: DO gives WILL
	do    []byte   // the options of the host side that it accepts: WILL gives DO
	naws  []byte   // the NAWS data after WILL NAWS, 4 bytes; nil sends no NAWS
	ttype []string // the name for each SEND; the last name repeats
	cpr   []string // the reply to each ESC[6n, for example "1;2"; "" gives no reply
	early string   // data that the client sends with its first reply
	quiet bool     // the client sends nothing
}

// fake runs a client on the client end of net.Pipe. A reader goroutine answers
// the host and a writer goroutine sends the answers, so the pipe cannot block both sides.
type fake struct {
	cl     client
	c      net.Conn
	send   chan []byte
	mu     sync.Mutex
	out    []byte // all bytes from the host
	pos    int    // the parsed bytes of out
	nsend  int    // the SEND requests
	ncpr   int    // the ESC[6n requests
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
			if bytes.Equal(rest[2:end], []byte{telnet.OptTTYPE, 1}) && len(f.cl.ttype) > 0 {
				name := f.cl.ttype[min(f.nsend, len(f.cl.ttype)-1)]
				f.nsend++
				reply = append(reply, iac, sb, telnet.OptTTYPE, 0)
				reply = append(reply, name...)
				reply = append(reply, iac, se)
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
			if f.ncpr < len(f.cl.cpr) && f.cl.cpr[f.ncpr] != "" {
				reply = append(reply, "\x1b["+f.cl.cpr[f.ncpr]+"R"...)
			}
			f.ncpr++
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

// output gives all bytes from the host.
func (f *fake) output() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return string(f.out)
}

// waitOutput waits until the output from the host ends with suffix.
// The reader appends the bytes of a host write after the write returns.
func (f *fake) waitOutput(t *testing.T, suffix string) {
	t.Helper()
	for end := time.Now().Add(time.Second); !strings.HasSuffix(f.output(), suffix); time.Sleep(time.Millisecond) {
		if time.Now().After(end) {
			t.Errorf("output %q does not end with %q", f.output(), suffix)
			return
		}
	}
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

// probeLog runs the probe and gives its result and the log.
func probeLog(t *testing.T, cl client, to Timeouts) (*Terminal, *fake, string, error) {
	t.Helper()
	host, f := start(t, cl)
	var log bytes.Buffer
	term, err := Probe(host, time.Now(), to, slog.New(slog.NewTextHandler(&log, nil)).With("addr", "192.0.2.1"))
	return term, f, log.String(), err
}

// The options that the clients of section 3 accept.
var (
	allOpts  = []byte{telnet.OptNAWS, telnet.OptTTYPE, telnet.OptSGA, telnet.OptBinary}
	hostOpts = []byte{telnet.OptEcho, telnet.OptSGA, telnet.OptBinary}
)

func TestProbeClients(t *testing.T) {
	tests := []struct {
		name  string
		cl    client
		want  screen.Term
		cprs  int // the ESC[6n requests
		probe string
	}{
		{"PuTTY",
			client{will: allOpts, do: hostOpts, naws: naws(120, 40), ttype: []string{"XTERM"}, cpr: []string{"1;2"}},
			screen.Term{Type: "xterm", Cols: 120, Rows: 40, Charset: "utf-8", Colors: 16}, 1, charsetProbe},
		{"Tera Term in line mode",
			client{
				will: []byte{telnet.OptNAWS, telnet.OptTTYPE, telnet.OptBinary},
				do:   []byte{telnet.OptEcho, telnet.OptBinary},
				naws: naws(80, 24), ttype: []string{"xterm"}, cpr: []string{"1;2"},
			},
			screen.Term{Type: "xterm", Cols: 80, Rows: 24, Charset: "utf-8", Colors: 16}, 1, charsetProbe},
		{"inetutils telnet",
			client{
				will: allOpts, do: hostOpts, naws: naws(200, 60),
				ttype: []string{"XTERM-256COLOR", "XTERM", "XTERM"}, cpr: []string{"3;2"},
			},
			screen.Term{Type: "xterm-256color", Cols: 200, Rows: 60, Charset: "utf-8", Colors: 256}, 1, charsetProbe},
		// MTTS gives UTF-8 and NAWS gives the size: no CPR probe.
		{"MTTS",
			client{
				will: allOpts, do: hostOpts, naws: naws(100, 30),
				ttype: []string{"MUDLET", "ANSI-TRUECOLOR", "MTTS 269"},
			},
			screen.Term{Type: "mudlet", Cols: 100, Rows: 30, Charset: "utf-8", Colors: ColorsTrue}, 0, ""},
		// MTTS with UTF-8 and no NAWS: only the size probe.
		{"MTTS without NAWS",
			client{will: []byte{telnet.OptTTYPE}, ttype: []string{"MTTS 4"}, cpr: []string{"40;120"}},
			screen.Term{Type: "ansi", Cols: 120, Rows: 40, Charset: "utf-8", Colors: 16}, 1, sizeProbe},
		// MTTS without the UTF-8 bit and no NAWS: both probes.
		{"MTTS without UTF-8",
			client{
				will: []byte{telnet.OptTTYPE}, ttype: []string{"MTTS 8"}, cpr: []string{"1;2", "50;132"},
			},
			screen.Term{Type: "ansi", Cols: 132, Rows: 50, Charset: "utf-8", Colors: 256}, 2, charsetProbe + sizeProbe},
		// A NAWS with 0 gives no size, so the size probe gives it.
		{"NAWS with 0",
			client{will: allOpts, naws: naws(0, 24), ttype: []string{"VT100"}, cpr: []string{"1;2", "30;100"}},
			screen.Term{Type: "vt100", Cols: 100, Rows: 30, Charset: "utf-8", Colors: 16}, 2, charsetProbe + sizeProbe},
		// The fourth name ends TTYPE. A key and a CPR reply arrive before step 4:
		// the key goes to the queue and the reply is dropped.
		{"four names and early input",
			client{
				will: allOpts, naws: naws(90, 30), ttype: []string{"A", "B", "C", "D-256color", "E-direct"},
				cpr: []string{"1;2"}, early: "x\x1b[1;3R",
			},
			screen.Term{Type: "a", Cols: 90, Rows: 30, Charset: "utf-8", Colors: 256}, 1, charsetProbe},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			term, f, log, err := probeLog(t, tt.cl, testTimeouts)
			if err != nil {
				t.Fatal(err)
			}
			if d := time.Since(start); d > replyWait/2 {
				t.Errorf("the probe took %v; want no wait", d)
			}
			if got := term.Term(); got != tt.want {
				t.Errorf("Term() = %+v, want %+v", got, tt.want)
			}
			out := f.output()
			if n := strings.Count(out, string(dsr)); n != tt.cprs {
				t.Errorf("%d CPR probes, want %d: %q", n, tt.cprs, out)
			}
			if tt.probe != "" && !strings.Contains(out, tt.probe) {
				t.Errorf("output %q has no %q", out, tt.probe)
			}
			if strings.Contains(out, charsetProbe) != strings.Contains(tt.probe, charsetProbe) {
				t.Errorf("output %q: charset probe %v, want %v", out, !strings.Contains(tt.probe, charsetProbe), strings.Contains(tt.probe, charsetProbe))
			}
			if !strings.Contains(log, "type="+tt.want.Type) {
				t.Errorf("log %q", log)
			}
			if tt.cl.early != "" {
				if k, err := term.Key(); k != "x" || err != nil {
					t.Errorf("Key() = %q, %v; want x", k, err)
				}
			}
		})
	}
}

func TestProbeStart(t *testing.T) {
	// Step 2: the requests of section 4.1 in one write, then TTYPE SEND after WILL TTYPE
	// and after each name until the end of TTYPE.
	_, f, _, err := probeLog(t, client{will: allOpts, naws: naws(80, 24), ttype: []string{"A", "B", "B"}, cpr: []string{"1;2"}}, testTimeouts)
	if err != nil {
		t.Fatal(err)
	}
	out := f.output()
	p := telnet.New()
	if !strings.HasPrefix(out, string(p.Start())) {
		t.Errorf("output %q does not start with the requests", out)
	}
	// SEND after WILL TTYPE, after A and after B. The second B ends TTYPE.
	if n := strings.Count(out, string(telnet.SendTTYPE())); n != 3 {
		t.Errorf("%d SEND requests, want 3", n)
	}
}

func TestProbeLog(t *testing.T) {
	cl := client{
		will: allOpts, naws: naws(100, 30), ttype: []string{"X\x1bTERM-256color", "MTTS 13"},
	}
	_, _, log, err := probeLog(t, cl, testTimeouts)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{
		"msg=probe", "addr=192.0.2.1", `ttype="[XTERM-256color MTTS 13]"`, "mtts=13", "naws=100x30",
		"cpr=[]", "type=xterm-256color", "cols=100", "rows=30", "charset=utf-8", "colors=256", "time=",
	} {
		if !strings.Contains(log, s) {
			t.Errorf("log %q has no %q", log, s)
		}
	}
	if strings.Count(log, "\n") != 1 {
		t.Errorf("log %q is not one line", log)
	}
}

func TestProbeNotUTF8(t *testing.T) {
	cl := client{will: []byte{telnet.OptTTYPE}, ttype: []string{"SYNCTERM"}, cpr: []string{"1;3", "25;80"}}
	term, f, log, err := probeLog(t, cl, testTimeouts)
	if term != nil || !errors.Is(err, ErrNotUTF8) {
		t.Fatalf("Probe() = %v, %v; want ErrNotUTF8", term, err)
	}
	f.waitOutput(t, notUTF8Text)
	for _, s := range []string{`charset="not utf-8"`, "cpr=\"[1;3 25;80]\"", "naws=-", "mtts=-"} {
		if !strings.Contains(log, s) {
			t.Errorf("log %q has no %q", log, s)
		}
	}
}

func TestProbeNoReplies(t *testing.T) {
	// probe_timeout ends the wait before the 1 s.
	to := testTimeouts
	to.Probe = 300 * time.Millisecond
	start := time.Now()
	term, f, _, err := probeLog(t, client{quiet: true}, to)
	if err != nil {
		t.Fatal(err)
	}
	if d := time.Since(start); d < to.Probe || d > replyWait {
		t.Errorf("the probe took %v; want %v", d, to.Probe)
	}
	want := screen.Term{Type: "ansi", Cols: 80, Rows: 25, Charset: "utf-8", Colors: 16}
	if got := term.Term(); got != want {
		t.Errorf("Term() = %+v, want %+v", got, want)
	}
	// After probe_timeout, the host sends no CPR probe.
	if strings.Contains(f.output(), string(dsr)) {
		t.Errorf("output %q has a CPR probe", f.output())
	}
}

func TestProbeReplyWait(t *testing.T) {
	// WILL NAWS without a NAWS: step 3 waits 1 s for the NAWS. Then the size probe.
	start := time.Now()
	cl := client{will: []byte{telnet.OptNAWS}, cpr: []string{"1;2", "24;80"}}
	term, _, _, err := probeLog(t, cl, testTimeouts)
	if err != nil {
		t.Fatal(err)
	}
	if d := time.Since(start); d < replyWait || d > 2*replyWait {
		t.Errorf("the probe took %v; want %v", d, replyWait)
	}
	if got := term.Term(); got.Cols != 80 || got.Rows != 24 {
		t.Errorf("Term() = %+v, want 80x24", got)
	}
}

func TestProbeHangup(t *testing.T) {
	host, f := start(t, client{quiet: true})
	go func() { _ = f.c.Close() }()
	if _, err := Probe(host, time.Now(), testTimeouts, slog.New(slog.DiscardHandler)); !errors.Is(err, io.ErrClosedPipe) && !errors.Is(err, io.EOF) {
		t.Errorf("Probe() error = %v, want a closed connection", err)
	}
}

func TestProbeLate(t *testing.T) {
	// probe_timeout passed before the probe (the PROXY header took the time).
	// The requests go out before Probe returns, so before the first screen.
	host, f := start(t, client{quiet: true})
	term, err := Probe(host, time.Now().Add(-testTimeouts.Probe), testTimeouts, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatal(err)
	}
	f.waitOutput(t, string(telnet.New().Start()))
	if got := term.Term(); got.Cols != 80 || got.Rows != 25 {
		t.Errorf("Term() = %+v, want 80x25", got)
	}
}
