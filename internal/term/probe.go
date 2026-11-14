package term

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/schors/velourbbs/internal/screen"
	"github.com/schors/velourbbs/internal/telnet"
)

// replyWait is the time without a reply that ends a wait of the probe (section 5.1).
const replyWait = time.Second

// charsetProbe is the charset probe of section 5.4 and the erase after it.
const charsetProbe = "\râ\x1b[6n\r\x1b[K"

// notUTF8Text is the text of section 5.5.
const notUTF8Text = "\r\n" +
	"This BBS needs a UTF-8 terminal. Your terminal uses another charset.\r\n" +
	"Set the charset (encoding) of your client to UTF-8 and connect again.\r\n" +
	"SyncTERM has no UTF-8. Use PuTTY, Tera Term or telnet in a UTF-8 terminal.\r\n"

// ErrNotUTF8 tells that the charset probe gave "not UTF-8" (section 5.5).
var ErrNotUTF8 = errors.New("term: the terminal is not UTF-8")

// probe is the state of the probe of one connection.
type probe struct {
	t   *Terminal
	end time.Time // accept + probe_timeout
	r   replies

	nawsAnswered, nawsOn, nawsGot bool
	ttypeAnswered, ttypeOn        bool
	ttypeSent                     bool // the host sent TTYPE SEND
	cprSent                       bool // the host sent the charset probe
}

// Probe does steps 2–5 of section 5.1 on the connection c, without the first screen.
// accepted is the time of accept. log has the address of the client.
// After "not UTF-8", Probe writes the text of section 5.5 and gives ErrNotUTF8.
// After each error, the caller closes c.
func Probe(c Conn, accepted time.Time, to Timeouts, log *slog.Logger) (*Terminal, error) {
	t := &Terminal{c: c, to: to, p: telnet.New(), buf: make([]byte, readSize)}
	pr := &probe{t: t, end: accepted.Add(to.Probe)}

	// Steps 2 and 3.
	t.out = t.p.Start()
	if err := pr.wait(false, pr.repliesDone); err != nil {
		return nil, err
	}

	// Step 4.
	if time.Now().Before(pr.end) {
		pr.cprSent = true
		t.out = append(t.out, charsetProbe...)
		if err := pr.wait(true, func() bool { return pr.r.cprGot }); err != nil {
			return nil, err
		}
	}

	// Step 5.
	term, utf8 := pr.r.values()
	pr.log(log, accepted, term, utf8)
	if !utf8 {
		t.out = append(t.out, notUTF8Text...)
		if err := t.write(t.out); err != nil {
			return nil, err
		}
		return nil, ErrNotUTF8
	}
	// The output of the probe goes out before the first screen.
	if len(t.out) > 0 {
		err := t.write(t.out)
		t.out = t.out[:0]
		if err != nil {
			return nil, err
		}
	}
	t.term = term
	t.lastKey = time.Now()
	return t, nil
}

// repliesDone reports whether all replies of step 3 arrived (section 5.1).
func (pr *probe) repliesDone() bool {
	return pr.nawsAnswered && (!pr.nawsOn || pr.nawsGot) &&
		pr.ttypeAnswered && (!pr.ttypeOn || pr.r.ttypeGot)
}

// wait reads the input until done gives true, or until no reply arrives for replyWait,
// or until the end of the probe (section 5.1). In step 4 (cpr is true), only a CPR reply
// is a reply. In step 3, only a reply of step 3.
func (pr *probe) wait(cpr bool, done func() bool) error {
	last := time.Now()
	for !done() {
		dl := last.Add(replyWait)
		if pr.end.Before(dl) {
			dl = pr.end
		}
		if !time.Now().Before(dl) {
			return nil
		}
		events, cprs, err := pr.t.read(dl)
		if err != nil {
			return err
		}
		if step3, step4 := pr.handle(events, cprs); cpr && step4 || !cpr && step3 {
			last = time.Now()
		}
	}
	return nil
}

// handle handles the events and the CPR replies of one read. step3 is true when
// a reply of step 3 arrived, step4 when a CPR reply arrived (section 5.1).
func (pr *probe) handle(events []telnet.Event, cprs []CPR) (step3, step4 bool) {
	for _, e := range events {
		switch e := e.(type) {
		case telnet.Option:
			switch {
			case e.Local:
			case e.Opt == telnet.OptNAWS:
				step3 = step3 || !pr.nawsAnswered
				pr.nawsAnswered, pr.nawsOn = true, e.On
			case e.Opt == telnet.OptTTYPE:
				step3 = step3 || !pr.ttypeAnswered
				pr.ttypeAnswered, pr.ttypeOn = true, e.On
				if e.On && !pr.ttypeSent {
					pr.ttypeSent = true
					pr.t.out = append(pr.t.out, telnet.SendTTYPE()...)
				}
			}
		case telnet.NAWS:
			step3 = step3 || !pr.nawsGot
			pr.nawsGot = true
			if e.Cols > 0 && e.Rows > 0 {
				pr.r.naws = e
			}
		case telnet.TTYPE:
			if !pr.r.ttypeGot {
				step3 = true
				pr.r.ttype, pr.r.ttypeGot = e.Name, true
			}
		}
	}
	// A CPR reply before step 4 is dropped.
	if len(cprs) > 0 && pr.cprSent && !pr.r.cprGot {
		pr.r.cpr, pr.r.cprGot, step4 = cprs[0], true, true
	}
	return step3, step4
}

// log writes the log line of section 5.6.
func (pr *probe) log(log *slog.Logger, accepted time.Time, term screen.Term, utf8 bool) {
	ttype := "-"
	if pr.r.ttypeGot {
		ttype = cleanName(pr.r.ttype)
	}
	naws := "-"
	if pr.r.naws.Cols > 0 {
		naws = fmt.Sprintf("%dx%d", pr.r.naws.Cols, pr.r.naws.Rows)
	}
	cpr := "-"
	if pr.r.cprGot {
		cpr = fmt.Sprintf("%d;%d", pr.r.cpr.Row, pr.r.cpr.Col)
	}
	charset := term.Charset
	if !utf8 {
		charset = "not utf-8"
	}
	log.Info("probe",
		"ttype", ttype, "naws", naws, "cpr", cpr,
		"type", term.Type, "cols", term.Cols, "rows", term.Rows,
		"charset", charset, "colors", term.Colors,
		"time", time.Since(accepted))
}
