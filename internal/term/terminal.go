package term

import (
	"errors"
	"io"
	"os"
	"sync/atomic"
	"time"

	"github.com/schors/velourbbs/internal/screen"
	"github.com/schors/velourbbs/internal/telnet"
)

// Conn is the connection of a terminal. net.Conn has these methods.
// A read after the read deadline gives an error that matches os.ErrDeadlineExceeded.
type Conn interface {
	io.ReadWriter
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
}

// Timeouts are the timeouts of [telnet] (bbs-toml, section 3.1).
type Timeouts struct {
	Probe  time.Duration // probe_timeout
	Idle   time.Duration // idle_timeout
	Output time.Duration // output_timeout
}

// maxKeys is the size of the key queue (section 8.3).
const maxKeys = 64

// readSize is the size of the read buffer.
const readSize = 4096

// ErrIdle tells that no key arrived within the idle timeout.
var ErrIdle = errors.New("term: idle timeout")

// ErrStop tells that the host stops (telnet-term, section 10.2).
var ErrStop = errors.New("term: stop")

// Terminal is the terminal of menu-nav, section 8, on a telnet connection.
// Probe gives it. One goroutine uses it. Another goroutine can call Stop.
type Terminal struct {
	c       Conn
	to      Timeouts
	p       *telnet.Parser
	d       Decoder
	term    screen.Term
	keys    []string  // the key queue
	lastKey time.Time // the start of the idle timeout
	out     []byte    // the replies that wait for the next read
	buf     []byte    // the read buffer
	stopped atomic.Bool
}

// Stop makes Key give ErrStop: a Key that waits for the input ends,
// and each later Key ends at once (section 10.2). A Page that writes completes.
func (t *Terminal) Stop() {
	t.stopped.Store(true)
	// The read checks stopped after it sets its deadline, so a read
	// that starts before this deadline also ends.
	_ = t.c.SetReadDeadline(time.Now())
}

// Term gives the terminal fields.
func (t *Terminal) Term() screen.Term {
	return t.term
}

// Key gives the next key of the queue. With an empty queue, it reads the input
// until a key arrives. While the key decoder waits for the next byte of a sequence,
// it also reads: a busy host must not end the wait of bytes that arrived in time.
// It gives ErrIdle at the idle timeout, ErrStop after Stop and the errors of the connection
// and of the telnet parser.
func (t *Terminal) Key() (string, error) {
	for {
		if t.stopped.Load() {
			return "", ErrStop
		}
		_, waiting := t.d.Deadline()
		if len(t.keys) > 0 && !waiting {
			break
		}
		idle := t.lastKey.Add(t.to.Idle)
		if !time.Now().Before(idle) {
			return "", ErrIdle
		}
		events, _, err := t.read(idle)
		if err != nil {
			return "", err
		}
		// After the probe, the terminal drops the CPR replies, the option events
		// and the TTYPE names.
		for _, e := range events {
			if n, ok := e.(telnet.NAWS); ok && n.Cols > 0 && n.Rows > 0 {
				t.term.Cols, t.term.Rows = clamp(n.Cols, n.Rows)
			}
		}
	}
	k := t.keys[0]
	t.keys = t.keys[1:]
	return k, nil
}

// Page writes the page p in one write (sections 7.1 and 9.4).
func (t *Terminal) Page(p screen.Page) error {
	return t.write(telnet.Escape(nil, AppendPage(nil, p, t.term.Colors)))
}

// read writes the waiting replies, then reads the input once, until the deadline dl
// at the latest. It gives the telnet events and the CPR replies. The keys go to the queue.
// The wait of the key decoder and the 10 s of a subnegotiation also end a read,
// so a read can give nothing.
func (t *Terminal) read(dl time.Time) ([]telnet.Event, []CPR, error) {
	if len(t.out) > 0 {
		err := t.write(t.out)
		t.out = t.out[:0]
		if err != nil {
			return nil, nil, err
		}
	}
	if d, ok := t.d.Deadline(); ok && d.Before(dl) {
		dl = d
	}
	if d, ok := t.p.Deadline(); ok && d.Before(dl) {
		dl = d
	}
	if err := t.c.SetReadDeadline(dl); err != nil {
		return nil, nil, err
	}
	if t.stopped.Load() {
		return nil, nil, ErrStop
	}
	n, err := t.c.Read(t.buf)
	if err != nil && !errors.Is(err, os.ErrDeadlineExceeded) {
		return nil, nil, err
	}
	now := time.Now()
	r, err := t.p.Feed(now, t.buf[:n])
	if err != nil {
		return nil, nil, err
	}
	t.out = append(t.out, r.Reply...)
	keys, cprs := t.d.Feed(now, r.Data)
	if len(keys) > 0 {
		t.lastKey = now
	}
	for _, k := range keys {
		if len(t.keys) < maxKeys {
			t.keys = append(t.keys, k)
		}
	}
	return r.Events, cprs, nil
}

// write writes b with the deadline output_timeout.
func (t *Terminal) write(b []byte) error {
	if err := t.c.SetWriteDeadline(time.Now().Add(t.to.Output)); err != nil {
		return err
	}
	_, err := t.c.Write(b)
	return err
}
