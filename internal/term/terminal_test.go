package term

import (
	"context"
	"errors"
	"os"
	"slices"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/schors/velourbbs/internal/screen"
	"github.com/schors/velourbbs/internal/telnet"
)

// puttyClient answers the probe at once: XTERM gives 256 colours and NAWS gives 100x30.
var puttyClient = client{will: allOpts, do: hostOpts, naws: naws(100, 30), ttype: []string{"XTERM"}, cpr: "1;2"}

// connect runs the probe with puttyClient. The caller runs in a synctest bubble.
func connect(t *testing.T, to Timeouts) (*Terminal, *fake) {
	t.Helper()
	term, f, _, _, err := probeLog(t, puttyClient, to)
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
		k, err := term.Key(t.Context())
		if err != nil {
			t.Fatalf("Key() after %q: %v", ks, err)
		}
		ks = append(ks, k)
	}
	return ks
}

func TestTerminalKeys(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
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
		if got := keys(t, term, 1); got[0] != "ESC" || time.Since(start) != EscWait {
			t.Errorf("keys %q after %v, want ESC after %v", got, time.Since(start), EscWait)
		}
	})
}

func TestTerminalBusy(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
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
	})
}

func TestTerminalQueueIdle(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// A key that the full queue drops resets the idle timeout (section 8.3).
		to := testTimeouts
		to.Idle = 300 * time.Millisecond
		term, f := connect(t, to)
		start := time.Now()
		// The ESC makes Key read on with a full queue. After 150 ms it gives the key ESC,
		// the full queue drops it, and the idle timeout restarts.
		f.write(strings.Repeat("a", maxKeys) + "\x1b")
		keys(t, term, maxKeys)
		if k, err := term.Key(t.Context()); !errors.Is(err, ErrIdle) {
			t.Fatalf("Key() = %q, %v; want ErrIdle", k, err)
		}
		if d := time.Since(start); d != EscWait+to.Idle {
			t.Errorf("ErrIdle after %v, want %v", d, EscWait+to.Idle)
		}
	})
}

func TestTerminalQueue(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		to := testTimeouts
		to.Idle = 200 * time.Millisecond
		term, f := connect(t, to)
		// One read gives maxKeys + 6 keys. The queue keeps maxKeys.
		f.write(strings.Repeat("a", maxKeys) + "bcdefg")
		if got := keys(t, term, maxKeys); slices.ContainsFunc(got, func(k string) bool { return k != "a" }) {
			t.Errorf("keys %q, want %d times a", got, maxKeys)
		}
		if k, err := term.Key(t.Context()); !errors.Is(err, ErrIdle) {
			t.Errorf("Key() = %q, %v; want ErrIdle", k, err)
		}
	})
}

func TestTerminalIdle(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
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
		if k, err := term.Key(t.Context()); !errors.Is(err, ErrIdle) {
			t.Fatalf("Key() = %q, %v; want ErrIdle", k, err)
		}
		if d := time.Since(start); d != 500*time.Millisecond {
			t.Errorf("ErrIdle after %v, want 500ms", d)
		}
		// The NAWS update changed the size.
		if got := term.Term(); got.Cols != 132 || got.Rows != 50 {
			t.Errorf("Term() = %+v, want 132x50", got)
		}
	})
}

func TestTerminalNAWS(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
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
	})
}

func TestTerminalHangup(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		term, f := connect(t, testTimeouts)
		f.write("a\xff\xf4")
		if k, err := term.Key(t.Context()); !errors.Is(err, telnet.ErrHangup) {
			t.Errorf("Key() = %q, %v; want ErrHangup", k, err)
		}
	})
}

func TestTerminalReply(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// The reply to AYT and to an unknown option goes out before the next read.
		term, f := connect(t, testTimeouts)
		f.write("\xff\xf6\xff\xfb\x05")
		f.write("k")
		keys(t, term, 1)
		if out := f.output(); !strings.HasSuffix(out, "\x07\xff\xfe\x05") {
			t.Errorf("output %q does not end with the replies", out)
		}
	})
}

func TestTerminalPage(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		term, f := connect(t, testTimeouts)
		page := rows(screen.Row{{Grapheme: "ÿ", Style: screen.Style{Fg: screen.RGBColour(1, 2, 3)}}})
		if err := term.Page(page); err != nil {
			t.Fatal(err)
		}
		// XTERM gives 256 colours, so the RGB colour becomes an index.
		if out := f.output(); !strings.HasSuffix(out, string(AppendPage(nil, page, screen.Colors256, false))) {
			t.Errorf("output %q does not end with the page", out)
		}
	})
}

// A cursor move sends only the changed cells: Update writes the runs of the grid that differ
// from the shown grid, and an update without a change writes nothing (section 7.5).
func TestTerminalUpdate(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		term, f := connect(t, testTimeouts)
		page := func(sel int) screen.Page {
			var rs []screen.Row
			for i := range 4 {
				mark := " "
				if i == sel {
					mark = ">"
				}
				rs = append(rs, screen.Row{{Grapheme: mark}, {Grapheme: string(rune('a' + i))}})
			}
			return screen.Page{Rows: rs, Cols: 2, Height: 5}
		}
		if err := term.Page(page(0)); err != nil {
			t.Fatal(err)
		}
		before := len(f.output())
		for _, p := range []screen.Page{page(1), page(1), page(2)} {
			if err := term.Update(p); err != nil {
				t.Fatal(err)
			}
		}
		want := "\x1b[?2026h\x1b[?25l\x1b[1;1H \x1b[2;1H>\x1b[0m\x1b[4;2H\x1b[?25h\x1b[?2026l" +
			"\x1b[?2026h\x1b[?25l\x1b[2;1H \x1b[3;1H>\x1b[0m\x1b[4;2H\x1b[?25h\x1b[?2026l"
		if got := f.output()[before:]; got != want {
			t.Errorf("output of the updates\n%q\nwant\n%q", got, want)
		}
	})
}

// Wait gives no key at its deadline and a key that comes before it. A key in the queue comes first.
// The deadlines do not reset the idle timeout (menu-nav, section 8, and section 8.3).
func TestTerminalWait(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		to := testTimeouts
		to.Idle = 2500 * time.Millisecond
		term, f := connect(t, to)
		start := time.Now()
		if k, ok, err := term.Wait(t.Context(), start.Add(time.Second)); k != "" || ok || err != nil {
			t.Fatalf("Wait() = %q, %v, %v; want no key", k, ok, err)
		}
		if d := time.Since(start); d != time.Second {
			t.Errorf("Wait ended after %v, want 1s", d)
		}
		f.write("ab")
		if k, ok, err := term.Wait(t.Context(), start.Add(2*time.Second)); k != "a" || !ok || err != nil {
			t.Fatalf("Wait() = %q, %v, %v; want a", k, ok, err)
		}
		if k, ok, err := term.Wait(t.Context(), start); k != "b" || !ok || err != nil {
			t.Fatalf("Wait() after the deadline with a key in the queue = %q, %v, %v; want b", k, ok, err)
		}
		// The key "a" came at 1 s: the idle timeout ends at 3.5 s, after the deadlines of each 1 s.
		var err error
		for n := 2; err == nil; n++ {
			_, _, err = term.Wait(t.Context(), start.Add(time.Duration(n)*time.Second))
		}
		if !errors.Is(err, ErrIdle) || time.Since(start) != 3500*time.Millisecond {
			t.Errorf("Wait: %v after %v, want ErrIdle after 3.5s", err, time.Since(start))
		}
	})
}

// The end of the context ends a Wait with a deadline at once (telnet-term, section 10.2).
func TestTerminalWaitStop(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		term, _ := connect(t, testTimeouts)
		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan error, 1)
		go func() {
			_, _, err := term.Wait(ctx, time.Now().Add(time.Second))
			done <- err
		}()
		time.Sleep(300 * time.Millisecond)
		start := time.Now()
		cancel()
		if err := <-done; !errors.Is(err, ErrStop) || time.Since(start) != 0 {
			t.Errorf("Wait: %v after %v, want ErrStop at once", err, time.Since(start))
		}
	})
}

func TestTerminalOutputTimeout(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// A client that does not read ends the session after output_timeout.
		to := testTimeouts
		to.Output = 200 * time.Millisecond
		term, f := connect(t, to)
		f.pause()
		page := rows(screen.Row{{Grapheme: "a"}})
		start := time.Now()
		var err error
		for range 3 {
			if err = term.Page(page); err != nil {
				break
			}
		}
		if !errors.Is(err, os.ErrDeadlineExceeded) {
			t.Fatalf("Page() error = %v, want a deadline error", err)
		}
		if d := time.Since(start); d != to.Output {
			t.Errorf("the error after %v, want %v", d, to.Output)
		}
	})
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
		if _, _, err := term.read(t.Context(), far); err != nil {
			t.Fatal(err)
		}
	}
	sub, ok := term.p.Deadline()
	if !ok || !slices.Equal(c.deadlines, []time.Time{far, sub}) {
		t.Errorf("read deadlines %v, want %v and %v", c.deadlines, far, sub)
	}
}

// The stop function of onDone waits for a call of f that started, also when ctx ended
// before onDone: a late deadline must not follow the wait.
func TestOnDoneWaits(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		ended := false
		stop := onDone(ctx, func() {
			time.Sleep(time.Second)
			ended = true
		})
		stop()
		if !ended {
			t.Error("stop returned before the end of f")
		}
	})
}

// The end of the context ends a Key that waits and each later Key with ErrStop at once,
// also with keys in the queue (session, section 6). A Page after the end writes.
func TestTerminalStop(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		term, f := connect(t, testTimeouts)
		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan error, 1)
		go func() {
			_, err := term.Key(ctx)
			done <- err
		}()
		synctest.Wait()
		start := time.Now()
		cancel()
		if err := <-done; !errors.Is(err, ErrStop) || time.Since(start) != 0 {
			t.Errorf("waiting Key: %v after %v, want ErrStop at once", err, time.Since(start))
		}
		// One read puts "a", "b" and "c" into the queue.
		f.write("abc")
		if got := keys(t, term, 1); got[0] != "a" || len(term.keys) != 2 {
			t.Fatalf("key %q, queue %v; want a and 2 keys in the queue", got, term.keys)
		}
		if _, err := term.Key(ctx); !errors.Is(err, ErrStop) {
			t.Errorf("Key with keys in the queue after the end: %v, want ErrStop", err)
		}
		if got := keys(t, term, 2); !slices.Equal(got, []string{"b", "c"}) {
			t.Errorf("keys %q with a new context, want the queue \"b\", \"c\"", got)
		}
		before := len(f.output())
		if err := term.Page(screen.Page{Rows: []screen.Row{{{Grapheme: "x"}}}, Cols: 1, Height: 1}); err != nil {
			t.Fatal(err)
		}
		if out := f.output(); !strings.Contains(out[before:], "x") {
			t.Errorf("output after Page %q, want x", out[before:])
		}
	})
}
