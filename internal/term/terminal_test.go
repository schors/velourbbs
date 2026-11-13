package term

import (
	"errors"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"newbbs/internal/screen"
	"newbbs/internal/telnet"
)

// mttsClient answers the probe at once: MTTS gives UTF-8 and NAWS gives 100x30.
var mttsClient = client{will: allOpts, do: hostOpts, naws: naws(100, 30), ttype: []string{"ANSI", "MTTS 13"}}

// connect runs the probe with mttsClient.
func connect(t *testing.T, to Timeouts) (*Terminal, *fake) {
	t.Helper()
	term, f, _, err := probeLog(t, mttsClient, to)
	if err != nil {
		t.Fatal(err)
	}
	return term, f
}

// keys reads n keys.
func keys(t *testing.T, term *Terminal, n int) []string {
	t.Helper()
	var ks []string
	for range n {
		k, err := term.Key()
		if err != nil {
			t.Fatalf("Key() after %q: %v", ks, err)
		}
		ks = append(ks, k)
	}
	return ks
}

func TestTerminalKeys(t *testing.T) {
	term, f := connect(t, testTimeouts)
	// Enter as CR NUL and CR LF, IAC IAC in the data (0xFF is not UTF-8, so it is dropped),
	// an escape sequence split across writes and a CPR reply after the probe.
	f.write("\r\x00a\xff\xff\r\n\x1b[")
	f.write("A\x1b[5;3R\x1b[6~")
	if got, want := keys(t, term, 5), []string{"ENTER", "a", "ENTER", "UP", "PGDN"}; !slices.Equal(got, want) {
		t.Errorf("keys %q, want %q", got, want)
	}
	// ESC alone gives ESC after the 150 ms wait.
	start := time.Now()
	f.write("\x1b")
	if got := keys(t, term, 1); got[0] != "ESC" || time.Since(start) < EscWait {
		t.Errorf("keys %q after %v, want ESC after %v", got, time.Since(start), EscWait)
	}
}

func TestTerminalBusy(t *testing.T) {
	// The host is busy for longer than 150 ms between two Key calls. The rest of the sequence
	// arrived in time, so it gives UP and not ESC and separate keys.
	term, f := connect(t, testTimeouts)
	f.write("x\x1b")
	go func() {
		time.Sleep(50 * time.Millisecond)
		f.write("[A")
	}()
	got := keys(t, term, 1)
	time.Sleep(3 * EscWait)
	f.write("y")
	if got = append(got, keys(t, term, 2)...); !slices.Equal(got, []string{"x", "UP", "y"}) {
		t.Errorf("keys %q, want x UP y", got)
	}
}

func TestTerminalQueueIdle(t *testing.T) {
	// A key that the full queue drops resets the idle timeout (section 8.3).
	to := testTimeouts
	to.Idle = 300 * time.Millisecond
	term, f := connect(t, to)
	start := time.Now()
	// The ESC makes Key read on with a full queue. After 150 ms it gives the key ESC,
	// the full queue drops it, and the idle timeout restarts.
	f.write(strings.Repeat("a", maxKeys) + "\x1b")
	keys(t, term, maxKeys)
	if k, err := term.Key(); !errors.Is(err, ErrIdle) {
		t.Fatalf("Key() = %q, %v; want ErrIdle", k, err)
	}
	if d := time.Since(start); d < EscWait+to.Idle || d > EscWait+to.Idle+200*time.Millisecond {
		t.Errorf("ErrIdle after %v, want %v", d, EscWait+to.Idle)
	}
}

func TestTerminalQueue(t *testing.T) {
	to := testTimeouts
	to.Idle = 200 * time.Millisecond
	term, f := connect(t, to)
	// One read gives 70 keys. The queue keeps 64.
	f.write(strings.Repeat("a", maxKeys) + "bcdefg")
	if got := keys(t, term, maxKeys); slices.ContainsFunc(got, func(k string) bool { return k != "a" }) {
		t.Errorf("keys %q, want %d times a", got, maxKeys)
	}
	if k, err := term.Key(); !errors.Is(err, ErrIdle) {
		t.Errorf("Key() = %q, %v; want ErrIdle", k, err)
	}
}

func TestTerminalIdle(t *testing.T) {
	to := testTimeouts
	to.Idle = 300 * time.Millisecond
	term, f := connect(t, to)
	start := time.Now()
	// A key resets the idle timeout.
	time.Sleep(200 * time.Millisecond)
	f.write("a")
	keys(t, term, 1)
	// NAWS and NOP do not reset it.
	f.write("\xff\xfa\x1f" + string(naws(132, 50)) + "\xff\xf0")
	time.Sleep(100 * time.Millisecond)
	f.write("\xff\xf1")
	if k, err := term.Key(); !errors.Is(err, ErrIdle) {
		t.Fatalf("Key() = %q, %v; want ErrIdle", k, err)
	}
	if d := time.Since(start); d < 500*time.Millisecond || d > 700*time.Millisecond {
		t.Errorf("ErrIdle after %v, want 500ms", d)
	}
	// The NAWS update changed the size.
	if got := term.Term(); got.Cols != 132 || got.Rows != 50 {
		t.Errorf("Term() = %+v, want 132x50", got)
	}
}

func TestTerminalNAWS(t *testing.T) {
	term, f := connect(t, testTimeouts)
	for _, tt := range []struct {
		cols, rows int
		want       [2]int
	}{
		{132, 50, [2]int{132, 50}},
		{0, 40, [2]int{132, 50}}, // 0 in a dimension: no change
		{90, 0, [2]int{132, 50}},
		{300, 5, [2]int{255, 10}}, // the clamp
		{1, 200, [2]int{20, 100}},
	} {
		// The key after the NAWS shows that Key read the NAWS.
		f.write("\xff\xfa\x1f" + string(naws(tt.cols, tt.rows)) + "\xff\xf0k")
		keys(t, term, 1)
		if got := term.Term(); got.Cols != tt.want[0] || got.Rows != tt.want[1] {
			t.Errorf("NAWS %dx%d: Term() = %+v, want %dx%d", tt.cols, tt.rows, got, tt.want[0], tt.want[1])
		}
		if err := term.Term().Check(); err != nil {
			t.Error(err)
		}
	}
}

func TestTerminalHangup(t *testing.T) {
	term, f := connect(t, testTimeouts)
	f.write("a\xff\xf4")
	if k, err := term.Key(); !errors.Is(err, telnet.ErrHangup) {
		t.Errorf("Key() = %q, %v; want ErrHangup", k, err)
	}
}

func TestTerminalReply(t *testing.T) {
	// The reply to AYT and to an unknown option goes out before the next read.
	term, f := connect(t, testTimeouts)
	f.write("\xff\xf6\xff\xfb\x05")
	f.write("k")
	keys(t, term, 1)
	f.waitOutput(t, "\x07\xff\xfe\x05")
}

func TestTerminalPage(t *testing.T) {
	term, f := connect(t, testTimeouts)
	page := screen.Page{{
		{Grapheme: "ÿ", Style: screen.Style{Fg: screen.RGBColour(1, 2, 3)}},
	}}
	if err := term.Page("main", []screen.Page{page}, 0); err != nil {
		t.Fatal(err)
	}
	// MTTS 13 gives 256 colours, so the RGB colour becomes an index.
	f.waitOutput(t, string(AppendPage(nil, page, Colors256)))
	if term.KeyEvent("a", false) != nil || term.Logoff() != nil {
		t.Error("KeyEvent or Logoff gives an error")
	}
}

func TestTerminalOutputTimeout(t *testing.T) {
	// A client that does not read ends the session after output_timeout.
	to := testTimeouts
	to.Output = 200 * time.Millisecond
	term, f := connect(t, to)
	f.pause()
	page := []screen.Page{{{{Grapheme: "a"}}}}
	start := time.Now()
	var err error
	for range 3 {
		if err = term.Page("main", page, 0); err != nil {
			break
		}
	}
	if !errors.Is(err, os.ErrDeadlineExceeded) {
		t.Fatalf("Page() error = %v, want a deadline error", err)
	}
	if d := time.Since(start); d < to.Output || d > 2*to.Output {
		t.Errorf("the error after %v, want %v", d, to.Output)
	}
}

// deadlineConn gives the bytes of in on the first read, then deadline errors.
// It records the read deadlines.
type deadlineConn struct {
	in        []byte
	deadlines []time.Time
}

func (c *deadlineConn) Read(b []byte) (int, error) {
	if c.in == nil {
		return 0, os.ErrDeadlineExceeded
	}
	n := copy(b, c.in)
	c.in = nil
	return n, nil
}

func (c *deadlineConn) Write(b []byte) (int, error) { return len(b), nil }

func (c *deadlineConn) SetReadDeadline(t time.Time) error {
	c.deadlines = append(c.deadlines, t)
	return nil
}

func (c *deadlineConn) SetWriteDeadline(time.Time) error { return nil }

func TestReadDeadline(t *testing.T) {
	// An open subnegotiation ends a read at its 10 s deadline (telnet-term, section 4.2).
	c := &deadlineConn{in: []byte("\xff\xfa\x1f\x00")}
	term := &Terminal{c: c, p: telnet.New(), buf: make([]byte, readSize)}
	far := time.Now().Add(time.Hour)
	for range 2 {
		if _, _, err := term.read(far); err != nil {
			t.Fatal(err)
		}
	}
	sub, ok := term.p.Deadline()
	if !ok || !slices.Equal(c.deadlines, []time.Time{far, sub}) {
		t.Errorf("read deadlines %v, want %v and %v", c.deadlines, far, sub)
	}
}
