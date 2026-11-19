package term

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strings"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/schors/velourbbs/internal/screen"
	"github.com/schors/velourbbs/internal/telnet"
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

// dsr is the cursor position request of a CPR probe, da1 is the DA1 query.
var (
	dsr = []byte("\x1b[6n")
	da1 = []byte("\x1b[c")
)

// client is the script of a fake telnet client.
type client struct {
	will  []byte   // the options of the client side that it enables: DO gives WILL
	do    []byte   // the options of the host side that it accepts: WILL gives DO
	naws  []byte   // the NAWS data after WILL NAWS, 4 bytes; nil sends no NAWS
	ttype []string // the names that the client sends for each SEND, each in its own subnegotiation
	cpr   string   // the reply to each ESC[6n, for example "1;2"; "" gives no reply
	cprs  []string // when not nil, the reply to the n-th ESC[6n in place of cpr; "" gives no reply
	term  string   // the replies to the terminal queries, sent at the DA1 query
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
	ndsr   int    // the ESC[6n requests
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
			cpr := f.cl.cpr
			if f.cl.cprs != nil {
				cpr = ""
				if f.ndsr < len(f.cl.cprs) {
					cpr = f.cl.cprs[f.ndsr]
				}
			}
			f.ndsr++
			if cpr != "" {
				reply = append(reply, "\x1b["+cpr+"R"...)
			}
			f.pos += len(dsr)
		case bytes.HasPrefix(rest, da1):
			reply = append(reply, f.cl.term...)
			f.pos += len(da1)
		case bytes.HasPrefix(dsr, rest) || bytes.HasPrefix(da1, rest):
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
	term, err := Probe(t.Context(), host, start, to, slog.New(slog.NewTextHandler(&log, nil)).With("addr", "192.0.2.1"))
	return term, f, log.String(), time.Since(start), err
}

// The options that the clients of section 3 accept.
var (
	allOpts  = []byte{telnet.OptNAWS, telnet.OptTTYPE, telnet.OptSGA, telnet.OptBinary}
	hostOpts = []byte{telnet.OptEcho, telnet.OptSGA, telnet.OptBinary}
)

// osc4 gives the OSC 4 replies to the indexes 0–15 with the VGA colours (section 7.3)
// in 4-digit form and the terminator end.
func osc4(end string) string {
	var b strings.Builder
	for n, c := range vga {
		fmt.Fprintf(&b, "\x1b]4;%d;rgb:%02x%02[2]x/%02x%02[3]x/%02x%02[4]x%s", n, c>>16, c>>8&0xff, c&0xff, end)
	}
	return b.String()
}

// osc gives the OSC reply with the body body and the terminator end.
func osc(body, end string) string {
	return "\x1b]" + body + end
}

// The replies of the terminals of section 5.7 in the order of the queries.
// The research of #31 and #49 gives them. The OSC 4 replies for 256 and 258 of xterm
// and Tera Term are special colours, of Windows Terminal reserved entries.
var (
	st     = "\x1b\\"
	termWT = "\x1bP1$r0;38:2::1:2:3m" + st + "\x1b[>0;10;1c" +
		osc("10;rgb:cccc/cccc/cccc", st) + osc("11;rgb:0c0c/0c0c/0c0c", st) + osc4(st) +
		osc("4;256;rgb:0000/0000/0000", st) + osc("4;258;rgb:0000/0000/0000", st) + "\x1b[?61;4;6;7;14;21;22;23;24;28;32;42;52c"
	termXterm = "\x1bP1$r0;38:2::1:2:3m" + st + "\x1bP>|XTerm(412)" + st + "\x1b[>41;412;0c" +
		osc("10;rgb:0000/0000/0000", "\a") + osc("11;rgb:ffff/ffff/ffff", "\a") + osc4("\a") +
		osc("4;256;rgb:0000/0000/0000", "\a") + osc("4;258;rgb:0000/0000/0000", "\a") + "\x1b[?64;1;2;6;9;15;17;18;21;22;28c"
	termXterm16 = "\x1bP1$r0;38:5:16m" + st + "\x1bP>|XTerm(412)" + st + "\x1b[>41;412;0c" +
		osc("10;rgb:0000/0000/0000", st) + osc("11;rgb:ffff/ffff/ffff", st) + osc4(st) +
		osc("4;256;rgb:0000/0000/0000", st) + osc("4;258;rgb:0000/0000/0000", st) + "\x1b[?64;1;2;6;9;15;17;18;21;22;28c"
	termTmux = "\x1bP0$r" + st + "\x1bP>|tmux 3.6" + st + "\x1b[>84;0;0c" +
		osc("10;rgb:d0d0/d0d0/d0d0", st) + osc("11;rgb:1c1c/1c1c/1c1c", st) + osc4(st) + "\x1b[?1;2;4c"
	termITerm2 = "\x1bP1$r0;38:2:1:1:2:3m" + st + "\x1bP>|iTerm2 3.7.7" + st + "\x1b[>64;2500;0c" +
		osc("10;rgb:c7c7/c7c7/c7c7", st) + osc("11;rgb:0000/0000/0000", st) + osc4(st) +
		osc("4;256;rgb:0000/0000/0000", st) + osc("4;258;rgb:0000/0000/0000", st) + "\x1b[?64;1;2;4;6;17;18;21;22;52c"
	termGhostty = "\x1bP1$r0;38:2::1:2:3m" + st + "\x1bP>|ghostty 1.4.0" + st + "\x1b[>1;10;0c" +
		osc("10;rgb:ffff/ffff/ffff", st) + osc("11;rgb:2828/2c2c/3434", st) + osc4(st) + "\x1b[?62;22;52c"
	termPuTTY = "\x1b[>0;136;0c" + osc4("\a") +
		osc("4;256;rgb:bbbb/bbbb/bbbb", "\a") + osc("4;258;rgb:0000/0000/0000", "\a") + "\x1b[?6c"
	termTeraTerm = "\x1bP1$r0;30m" + st + "\x1b[>32;331;0c" +
		osc("10;rgb:ffff/ffff/ffff", st) + osc("11;rgb:0000/0000/0000", st) + osc4(st) +
		osc("4;256;rgb:ffff/ffff/ffff", st) + osc("4;258;rgb:0000/0000/ffff", st) + "\x1b[?1;2c"
)

func TestProbeClients(t *testing.T) {
	term := func(typ string, cols, rows, colors int, name, images string) screen.Term {
		return screen.Term{Type: typ, Cols: cols, Rows: rows, Charset: "utf-8", Colors: colors, Name: name, Images: images, Ambiguous: 1}
	}
	inet := func(typ, replies string) client {
		return client{will: allOpts, do: hostOpts, naws: naws(80, 24), ttype: []string{typ}, cpr: "1;2", term: replies}
	}
	tests := []struct {
		name string
		cl   client
		want screen.Term
	}{
		// The terminals of section 5.7. inetutils telnet sends $TERM in upper case.
		{"Windows Terminal", inet("XTERM-256COLOR", termWT), colours(term("xterm-256color", 80, 24, screen.ColorsTrue, "", "sixel"), "#cccccc", "#0c0c0c", "dark")},
		{"xterm", inet("XTERM-256COLOR", termXterm), colours(term("xterm-256color", 80, 24, screen.ColorsTrue, "XTerm(412)", ""), "#000000", "#ffffff", "light")},
		{"xterm without direct colour", inet("XTERM", termXterm16), colours(term("xterm", 80, 24, 256, "XTerm(412)", ""), "#000000", "#ffffff", "light")},
		{"tmux", inet("TMUX-256COLOR", termTmux), colours(term("tmux-256color", 80, 24, screen.ColorsTrue, "tmux 3.6", "sixel"), "#d0d0d0", "#1c1c1c", "dark")},
		{"iTerm2", inet("XTERM-256COLOR", termITerm2), colours(term("xterm-256color", 80, 24, screen.ColorsTrue, "iTerm2 3.7.7", "sixel"), "#c7c7c7", "#000000", "dark")},
		{"Ghostty", inet("XTERM-GHOSTTY", termGhostty), colours(term("xterm-ghostty", 80, 24, screen.ColorsTrue, "ghostty 1.4.0", ""), "#ffffff", "#282c34", "dark")},
		{"PuTTY",
			client{will: allOpts, do: hostOpts, naws: naws(120, 40), ttype: []string{"XTERM"}, cpr: "1;2", term: termPuTTY},
			colours(term("xterm", 120, 40, screen.ColorsTrue, "", ""), "#bbbbbb", "#000000", "dark")},
		// PuTTY with "Treat CJK ambiguous characters as wide": the ambiguous probe
		// moves the cursor by two cells (section 5.4).
		{"PuTTY with wide ambiguous characters",
			client{will: allOpts, do: hostOpts, naws: naws(120, 40), ttype: []string{"XTERM"}, cprs: []string{"1;2", "1;3"}, term: termPuTTY},
			ambiguous(colours(term("xterm", 120, 40, screen.ColorsTrue, "", ""), "#bbbbbb", "#000000", "dark"), 2)},
		{"Tera Term in line mode",
			client{
				will: []byte{telnet.OptNAWS, telnet.OptTTYPE, telnet.OptBinary},
				do:   []byte{telnet.OptEcho, telnet.OptBinary},
				naws: naws(80, 24), ttype: []string{"xterm"}, cpr: "1;2", term: termTeraTerm,
			},
			colours(term("xterm", 80, 24, 256, "", ""), "#ffffff", "#000000", "dark")},
		// With only a DA1 reply, the TTYPE name gives the values.
		// The host takes the first name and ignores the later names.
		{"inetutils telnet",
			client{
				will: allOpts, do: hostOpts, naws: naws(200, 60),
				ttype: []string{"XTERM-256COLOR", "XTERM-DIRECT"}, cpr: "3;2", term: "\x1b[?1;2c",
			},
			term("xterm-256color", 200, 60, 256, "", "")},
		{"direct colour", client{will: allOpts, naws: naws(80, 25), ttype: []string{"xterm-direct"}, cpr: "1;2", term: "\x1b[?1;2c"},
			term("xterm-direct", 80, 25, screen.ColorsTrue, "", "")},
		// MTTS is a plain name.
		{"MTTS", client{will: allOpts, naws: naws(100, 30), ttype: []string{"MTTS 13"}, cpr: "1;2", term: "\x1b[?1;2c"},
			term("mtts 13", 100, 30, 256, "", "")},
		// A NAWS with 0 gives no size.
		{"NAWS with 0", client{will: allOpts, naws: naws(0, 30), ttype: []string{"VT100"}, cpr: "1;2", term: "\x1b[?1;2c"},
			term("vt100", 80, 24, 16, "", "")},
		{"no TTYPE", client{will: []byte{telnet.OptNAWS}, naws: naws(90, 30), cpr: "1;2", term: "\x1b[?1;2c"},
			term("ansi", 90, 30, 16, "", "")},
		// A key and replies arrive before step 4: the key goes to the queue
		// and the replies are dropped.
		{"early input",
			client{
				will: allOpts, naws: naws(90, 30), ttype: []string{"A"}, cpr: "1;2", term: "\x1b[?1;2c",
				early: "x\x1b[1;3R\x1b[>0;136;0c\x1b[?1;4c",
			},
			term("a", 90, 30, 256, "", "")},
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
			// One SEND after WILL TTYPE, and one write of step 4.
			out := f.output()
			if n, want := f.nsend, min(len(tt.cl.ttype), 1); n != want {
				t.Errorf("%s: %d SEND requests, want %d", tt.name, n, want)
			}
			if n := strings.Count(out, string(dsr)); n != 2 || !strings.Contains(out, probeWrite) {
				t.Errorf("%s: output %q, want one write of step 4", tt.name, out)
			}
			if tt.cl.early != "" {
				if k, err := term.Key(t.Context()); k != "x" || err != nil {
					t.Errorf("%s: Key() = %q, %v; want x", tt.name, k, err)
				}
			}
		})
	}
}

// TestProbePalette: the OSC 4 replies give the output of a VGA colour (section 7.3).
// A VGA palette and no OSC 4 reply give the index. A palette that is not VGA gives RGB.
func TestProbePalette(t *testing.T) {
	campbell := strings.Replace(osc4(st), "rgb:aaaa/0000/0000", "rgb:c5c5/0f0f/1f1f", 1)
	tests := []struct {
		name, replies, want string
	}{
		{"VGA palette", osc4(st) + "\x1b[?1;2c", "\x1b[0;31;40m"},
		{"palette that is not VGA", campbell + "\x1b[?1;2c", "\x1b[0;38;2;170;0;0;48;2;0;0;0m"},
		{"no OSC 4 reply", "\x1b[?1;2c", "\x1b[0;31;40m"},
	}
	for _, tt := range tests {
		synctest.Test(t, func(t *testing.T) {
			cl := client{will: allOpts, naws: naws(80, 24), ttype: []string{"xterm-direct"}, cpr: "1;2", term: tt.replies}
			term, f, _, _, err := probeLog(t, cl, testTimeouts)
			if err != nil {
				t.Fatalf("%s: %v", tt.name, err)
			}
			s := screen.Style{Fg: screen.VGAColour(1), Bg: screen.VGAColour(0)}
			if err := term.Page(screen.Page{Rows: []screen.Row{{{Grapheme: "a", Style: s}}}, Cols: 1}); err != nil {
				t.Fatal(err)
			}
			if out := f.output(); !strings.HasSuffix(out, tt.want+"a\x1b[0m\x1b[?2026l") {
				t.Errorf("%s: output %q does not end with %q", tt.name, out, tt.want+"a")
			}
		})
	}
}

func TestProbeWrite(t *testing.T) {
	// Section 5.7: the CPR probes first, the queries in one write, DA1 last, then the erase.
	want := "\râ\x1b[6n\rж\x1b[6n\x1b[38;2;1;2;3m\x1bP$qm\x1b\\\x1b[0m\x1b[>0q\x1b[>c" +
		"\x1b]10;?\x1b\\\x1b]11;?\x1b\\" +
		"\x1b]4;0;?\x1b\\\x1b]4;1;?\x1b\\\x1b]4;2;?\x1b\\\x1b]4;3;?\x1b\\" +
		"\x1b]4;4;?\x1b\\\x1b]4;5;?\x1b\\\x1b]4;6;?\x1b\\\x1b]4;7;?\x1b\\" +
		"\x1b]4;8;?\x1b\\\x1b]4;9;?\x1b\\\x1b]4;10;?\x1b\\\x1b]4;11;?\x1b\\" +
		"\x1b]4;12;?\x1b\\\x1b]4;13;?\x1b\\\x1b]4;14;?\x1b\\\x1b]4;15;?\x1b\\" +
		"\x1b]4;256;?\x1b\\\x1b]4;258;?\x1b\\\x1b[c\r\x1b[K"
	if probeWrite != want {
		t.Errorf("probeWrite = %q, want %q", probeWrite, want)
	}
}

func TestProbeLateDA1(t *testing.T) {
	// The replies before DA1 restart the 1 s wait. Without DA1, the wait ends
	// 1 s after the last reply, and the host keeps the replies.
	synctest.Test(t, func(t *testing.T) {
		cl := client{will: allOpts, naws: naws(80, 24), ttype: []string{"XTERM"}, cpr: "1;2"}
		host, f := start(t, cl)
		go func() {
			time.Sleep(900 * time.Millisecond)
			f.write("\x1bP>|tmux 3.6\x1b\\")
			time.Sleep(900 * time.Millisecond)
			f.write("\x1b[>84;0;0c")
			time.Sleep(900 * time.Millisecond)
			f.write("\x1b[>84;0;0c") // a second reply is not a reply
		}()
		start := time.Now()
		term, err := Probe(t.Context(), host, start, testTimeouts, slog.New(slog.DiscardHandler))
		if err != nil {
			t.Fatal(err)
		}
		if d := time.Since(start); d != 2800*time.Millisecond {
			t.Errorf("the probe took %v; want 2.8s", d)
		}
		if got := term.Term(); got.Name != "tmux 3.6" || got.Colors != screen.ColorsTrue {
			t.Errorf("Term() = %+v, want tmux 3.6 with 16777216 colours", got)
		}
	})
}

// colours sets term.fg_color, term.bg_color and term.background of t.
func colours(t screen.Term, fg, bg, background string) screen.Term {
	t.FgColor, t.BgColor, t.Background = fg, bg, background
	return t
}

// ambiguous gives t with term.ambiguous n.
func ambiguous(t screen.Term, n int) screen.Term {
	t.Ambiguous = n
	return t
}

// TestProbeAmbiguous: the second CPR reply is the reply to the ambiguous probe (section 5.4).
func TestProbeAmbiguous(t *testing.T) {
	tests := []struct {
		name string
		cprs []string
		term string // the replies at the DA1 query
		want int
	}{
		{"column 3", []string{"1;2", "1;3"}, "\x1b[?1;2c", 2},
		{"column 3 in another row", []string{"5;2", "5;3"}, "\x1b[?1;2c", 2},
		{"column 2", []string{"1;2", "1;2"}, "\x1b[?1;2c", 1},
		{"another column", []string{"1;2", "1;4"}, "\x1b[?1;2c", 1},
		{"no reply", []string{"1;2"}, "\x1b[?1;2c", 1},
		// A third CPR reply is dropped: it changes neither the ambiguous width nor the charset.
		{"third reply", []string{"1;2", "1;2"}, "\x1b[1;3R\x1b[?1;2c", 1},
		{"third reply after column 3", []string{"1;2", "1;3"}, "\x1b[1;2R\x1b[?1;2c", 2},
	}
	for _, tt := range tests {
		synctest.Test(t, func(t *testing.T) {
			cl := client{will: allOpts, naws: naws(80, 24), ttype: []string{"xterm"}, cprs: tt.cprs, term: tt.term}
			term, _, _, _, err := probeLog(t, cl, testTimeouts)
			if err != nil {
				t.Fatalf("%s: %v", tt.name, err)
			}
			if got := term.Term(); got.Ambiguous != tt.want || got.Charset != "utf-8" {
				t.Errorf("%s: Term() = %+v, want term.ambiguous %d and utf-8", tt.name, got, tt.want)
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
		cl := client{
			will: allOpts, naws: naws(100, 30), ttype: []string{"X\x1bTERM-256color", "VT100"}, cprs: []string{"1;2", "1;3"},
			term: "\x1bP1$r0;38:2::1:2:3m\x1b\\\x1bP>|X\x01Term(412)" + strings.Repeat("ы", 40) + "\x1b\\" +
				"\x1b]11;rgb:fe/fe/fe\x07\x1b]4;1;rgb:cdcd/0/80\x07\x1b]4;15;rgb:f/ff/fff\x1b\\\x1b]4;258;rgb:0/0/0\x07\x1b[?64;1;4c",
		}
		_, _, log, _, err := probeLog(t, cl, testTimeouts)
		if err != nil {
			t.Fatal(err)
		}
		want := "level=INFO msg=probe addr=192.0.2.1 ttype=XTERM-256color naws=100x30 cpr=1;2 cpr_ambiguous=1;3 " +
			"decrqss=1$r0;38:2::1:2:3m xtversion=>|XTerm(412)" + strings.Repeat("ы", 28) + " da2=- da1=?64;1;4c " +
			"palette=-,cd0080,-,-,-,-,-,-,-,-,-,-,-,-,-,ffffff osc10=- osc11=fefefe osc4_256=- osc4_258=000000 " +
			"type=xterm-256color cols=100 rows=30 charset=utf-8 colors=16777216 " +
			"name=XTerm(412)" + strings.Repeat("ы", 30) + " images=sixel ambiguous=2 fg_color=\"\" bg_color=#fefefe background=light time=0s\n"
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
		want := screen.Term{Type: "ansi", Cols: 80, Rows: 24, Charset: "utf-8", Colors: 16, Ambiguous: 1}
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
		if got := term.Term(); got.Cols != 80 || got.Rows != 24 {
			t.Errorf("Term() = %+v, want 80x24", got)
		}
	})
}

func TestProbeHangup(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		host, f := start(t, client{quiet: true})
		go func() { _ = f.c.Close() }()
		if _, err := Probe(t.Context(), host, time.Now(), testTimeouts, slog.New(slog.DiscardHandler)); !errors.Is(err, io.ErrClosedPipe) && !errors.Is(err, io.EOF) {
			t.Errorf("Probe() error = %v, want a closed connection", err)
		}
	})
}

// The end of the context ends the probe at once with ErrStop and no more output
// (session, section 6). A context that ended before the probe gives no output.
func TestProbeStop(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		host, f := start(t, client{quiet: true})
		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan error, 1)
		go func() {
			_, err := Probe(ctx, host, time.Now(), testTimeouts, slog.New(slog.DiscardHandler))
			done <- err
		}()
		synctest.Wait()
		start := time.Now()
		cancel()
		if err := <-done; !errors.Is(err, ErrStop) || time.Since(start) != 0 {
			t.Errorf("Probe() error = %v after %v, want ErrStop at once", err, time.Since(start))
		}
		if out := f.output(); out != string(telnet.New().Start()) {
			t.Errorf("output %q, want only the requests", out)
		}
	})
	synctest.Test(t, func(t *testing.T) {
		host, f := start(t, client{quiet: true})
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		if _, err := Probe(ctx, host, time.Now(), testTimeouts, slog.New(slog.DiscardHandler)); !errors.Is(err, ErrStop) {
			t.Errorf("Probe() after the end error = %v, want ErrStop", err)
		}
		synctest.Wait()
		if out := f.output(); out != "" {
			t.Errorf("output %q after the end, want none", out)
		}
	})
}

// The end of the context also ends a write of the probe to a client that does not read.
func TestProbeStopWrite(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		host, cc := net.Pipe()
		defer func() { _ = cc.Close() }()
		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan error, 1)
		go func() {
			_, err := Probe(ctx, host, time.Now(), testTimeouts, slog.New(slog.DiscardHandler))
			done <- err
		}()
		synctest.Wait()
		start := time.Now()
		cancel()
		if err := <-done; !errors.Is(err, ErrStop) || time.Since(start) != 0 {
			t.Errorf("Probe() error = %v after %v, want ErrStop at once", err, time.Since(start))
		}
	})
}

func TestProbeLate(t *testing.T) {
	// probe_timeout passed before the probe (the PROXY header took the time).
	// The requests go out before Probe returns, so before the first screen.
	synctest.Test(t, func(t *testing.T) {
		host, f := start(t, client{quiet: true})
		term, err := Probe(t.Context(), host, time.Now().Add(-testTimeouts.Probe), testTimeouts, slog.New(slog.DiscardHandler))
		if err != nil {
			t.Fatal(err)
		}
		if out := f.output(); out != string(telnet.New().Start()) {
			t.Errorf("output %q, want only the requests", out)
		}
		if got := term.Term(); got.Cols != 80 || got.Rows != 24 {
			t.Errorf("Term() = %+v, want 80x24", got)
		}
	})
}
