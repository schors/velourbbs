package term

import (
	"context"
	"errors"
	"io"
	"os"
	"time"

	"github.com/schors/velourbbs/internal/line"
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

// Timeouts are the timeouts of [telnet] and [limits] (bbs-toml, sections 3.1 and 3.3).
type Timeouts struct {
	Probe  time.Duration // probe_timeout
	Idle   time.Duration // idle_timeout
	Output time.Duration // output_timeout
}

// maxKeys is the size of the key queue (section 8.3).
const maxKeys = 256

// The bracketed paste mode on and off (line-input, section 6).
const (
	pasteOn  = "\x1b[?2004h"
	pasteOff = "\x1b[?2004l"
)

// readSize is the size of the read buffer.
const readSize = 4096

// ErrIdle tells that no key arrived within the idle timeout.
var ErrIdle = errors.New("term: idle timeout")

// ErrStop tells that the host stops (telnet-term, section 10.2): the context of a wait ended
// (session, section 6).
var ErrStop = errors.New("term: stop")

// Terminal is the terminal of menu-nav, section 8, on a telnet connection.
// Probe gives it. One goroutine uses it.
type Terminal struct {
	c       Conn
	to      Timeouts
	p       *telnet.Parser
	d       Decoder
	term    screen.Term
	vgaRGB  bool        // the palette is not VGA (section 5.7)
	keys    []queued    // the key queue
	lastKey time.Time   // the start of the idle timeout
	paste   bool        // the input is in a paste (section 8.4)
	drop    bool        // drop the rest of the paste
	mode    bool        // the bracketed paste mode is on
	out     []byte      // the replies that wait for the next read
	buf     []byte      // the read buffer
	shown   screen.Page // the page of the last Page or Update
}

// queued is a key of the queue with its paste mark.
type queued struct {
	k     string
	paste bool
}

// Term gives the terminal fields.
func (t *Terminal) Term() screen.Term {
	return t.term
}

// Key gives the next key of the queue that is not a key of a paste: outside the line
// input, the host drops each key of a paste (section 8.4). It gives the errors of next.
func (t *Terminal) Key(ctx context.Context) (string, error) {
	for {
		q, err := t.next(ctx)
		if err != nil {
			return "", err
		}
		if !q.paste {
			return q.k, nil
		}
	}
}

// ReadLine runs the line input in the field f (line-input, sections 6 and 7).
// It gives the line and true after ENTER, or false after ESC. It panics when the field
// does not fit the size of the page of the last Page or Update: the size of the current screen,
// also after a resize. It gives the errors of next and of the writes.
// Before an error of next, it turns the bracketed paste mode off (section 8.4).
func (t *Terminal) ReadLine(ctx context.Context, f line.Field) (string, bool, error) {
	l := line.New(f, t.shown.Cols, t.shown.Height)
	t.mode = true
	var out []byte // the draw that waits for the end of the queued keys
	if err := t.write(context.Background(), t.appendDraw([]byte(pasteOn), l.Draw())); err != nil {
		t.mode = false
		return "", false, err
	}
	for {
		if len(out) > 0 && !t.ready() {
			if err := t.write(context.Background(), out); err != nil {
				t.mode = false
				return "", false, err
			}
			out = out[:0]
		}
		q, err := t.next(ctx)
		if err != nil {
			t.endMode()
			return "", false, err
		}
		switch l.Key(q.k, q.paste) {
		case line.Changed:
			out = t.appendDraw(out[:0], l.Draw())
		case line.Enter:
			return l.Text(), true, t.endInput(out)
		case line.Cancel:
			return "", false, t.endInput(out)
		}
	}
}

// endInput writes the draw out, if any, and turns the paste mode off in one write.
// It clears the key queue and drops the rest of an open paste (line-input, section 6).
func (t *Terminal) endInput(out []byte) error {
	t.mode = false
	t.keys = t.keys[:0]
	t.drop = t.paste
	return t.write(context.Background(), append(out, pasteOff...))
}

// endMode turns the bracketed paste mode off at an end of the session in the line input.
// It ignores a write error (section 8.4).
func (t *Terminal) endMode() {
	if t.mode {
		t.mode = false
		_ = t.write(context.Background(), []byte(pasteOff))
	}
}

// appendDraw appends the bytes of the draw d (line-input, section 7) to dst.
func (t *Terminal) appendDraw(dst []byte, d line.Draw) []byte {
	f := d.Field
	dst = appendCUP(dst, f.Row, f.Col)
	dst = appendSGR(dst, reduceStyle(f.Style, t.term.Colors, t.vgaRGB))
	dst = appendUTF8(dst, d.Cells)
	dst = append(dst, "\x1b[0m"...)
	return appendCUP(dst, f.Row, f.Col+d.Cursor)
}

// ready reports whether next gives a key without a read.
func (t *Terminal) ready() bool {
	_, waiting := t.d.Deadline()
	return len(t.keys) > 0 && !waiting
}

// next gives the next key of the queue. With an empty queue, it reads the input
// until a key arrives. While the key decoder waits for the next byte of a sequence,
// it also reads: a busy host must not end the wait of bytes that arrived in time.
// It gives ErrIdle at the idle timeout, ErrStop at once after the end of ctx,
// also with keys in the queue, and the errors of the connection and of the telnet parser.
func (t *Terminal) next(ctx context.Context) (queued, error) {
	defer onDone(ctx, func() { _ = t.c.SetReadDeadline(time.Now()) })()
	for {
		if ctx.Err() != nil {
			return queued{}, ErrStop
		}
		if t.ready() {
			break
		}
		idle := t.lastKey.Add(t.to.Idle)
		if !time.Now().Before(idle) {
			return queued{}, ErrIdle
		}
		events, _, err := t.read(ctx, idle)
		if err != nil {
			return queued{}, err
		}
		// After the probe, the terminal drops the replies, the option events
		// and the TTYPE names.
		for _, e := range events {
			if n, ok := e.(telnet.NAWS); ok && n.Cols > 0 && n.Rows > 0 {
				t.term.Cols, t.term.Rows = clamp(n.Cols, n.Rows)
			}
		}
	}
	q := t.keys[0]
	t.keys = t.keys[1:]
	return q, nil
}

// Page writes the page p in one write (sections 7.1 and 9.4).
func (t *Terminal) Page(p screen.Page) error {
	t.shown = p
	return t.write(context.Background(), telnet.Escape(nil, AppendPage(nil, p, t.term.Colors, t.vgaRGB)))
}

// Update writes the rows of the page p that differ from the page of the last Page or Update,
// in one write (section 7.1). p has the size, the Top and the count of rows of that page.
func (t *Terminal) Update(p screen.Page) error {
	prev := t.shown
	t.shown = p
	return t.write(context.Background(), telnet.Escape(nil, AppendUpdate(nil, p, prev, t.term.Colors, t.vgaRGB)))
}

// onDone calls f in its own goroutine when ctx ends. It gives the function that stops this.
// The function waits for a call of f that started, so f has no effect after it returns:
// a late deadline of f must not end the close after output of the server.
func onDone(ctx context.Context, f func()) (stop func()) {
	done := make(chan struct{})
	stopF := context.AfterFunc(ctx, func() {
		defer close(done)
		f()
	})
	return func() {
		if !stopF() {
			<-done
		}
	}
}

// read writes the waiting replies, then reads the input once, until the deadline dl
// at the latest. It gives the telnet events and the replies. The keys go to the queue.
// The wait of the key decoder and the 10 s of a subnegotiation also end a read,
// so a read can give nothing. After the end of ctx, it gives ErrStop without a write.
// The caller makes the end of ctx end the read with onDone.
func (t *Terminal) read(ctx context.Context, dl time.Time) ([]telnet.Event, []Reply, error) {
	if ctx.Err() != nil {
		return nil, nil, ErrStop
	}
	if len(t.out) > 0 {
		err := t.write(ctx, t.out)
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
	// The read checks ctx after it sets its deadline, so a read that starts
	// before the deadline of onDone also ends.
	if ctx.Err() != nil {
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
	keys, replies := t.d.Feed(now, r.Data)
	for _, k := range keys {
		t.queue(now, k)
	}
	return r.Events, replies, nil
}

// queue adds the key k, which arrived at the time now, to the key queue
// with its paste mark (sections 8.3 and 8.4).
func (t *Terminal) queue(now time.Time, k string) {
	switch k {
	case PasteStart, PasteEnd:
		t.paste, t.drop = k == PasteStart, false
		return
	}
	t.lastKey = now
	switch {
	case t.paste && t.drop:
	case len(t.keys) < maxKeys:
		t.keys = append(t.keys, queued{k: k, paste: t.paste})
	case t.paste:
		t.drop = true
	}
}

// write writes b with the deadline output_timeout. After the end of ctx, it gives ErrStop.
// A write that must complete, as a page, takes context.Background().
func (t *Terminal) write(ctx context.Context, b []byte) error {
	if err := t.c.SetWriteDeadline(time.Now().Add(t.to.Output)); err != nil {
		return err
	}
	// The write checks ctx after it sets its deadline, so the deadline of an onDone
	// of the probe also ends a write that starts before it.
	if ctx.Err() != nil {
		return ErrStop
	}
	_, err := t.c.Write(b)
	return err
}
