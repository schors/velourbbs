package term

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"newbbs/internal/screen"
	"newbbs/internal/telnet"
)

// replyWait is the time without a reply that ends a wait of the probe (section 5.1).
const replyWait = time.Second

// The CPR probes of section 5.4.
const (
	charsetProbe = "\râ\x1b[6n\r\x1b[K"
	sizeProbe    = "\x1b[999C\x1b[999B\x1b[6n"
)

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
	mtts, ok := pr.r.mtts()
	pr.r.charset = !ok || mtts&mttsUTF8 == 0
	pr.r.sizeProbe = pr.r.naws.Cols == 0
	if !time.Now().Before(pr.end) {
		pr.r.charset, pr.r.sizeProbe = false, false
	}
	if pr.r.charset {
		t.out = append(t.out, charsetProbe...)
	}
	if pr.r.sizeProbe {
		t.out = append(t.out, sizeProbe...)
	}
	if err := pr.wait(true, func() bool { return len(pr.r.cprs) == pr.probes() }); err != nil {
		return nil, err
	}

	// Step 5.
	pr.r.ttype.ended = true
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

// probes gives the number of CPR probes that the host sent.
func (pr *probe) probes() int {
	n := 0
	for _, sent := range []bool{pr.r.charset, pr.r.sizeProbe} {
		if sent {
			n++
		}
	}
	return n
}

// repliesDone reports whether all replies of step 3 arrived (section 5.1).
func (pr *probe) repliesDone() bool {
	return pr.nawsAnswered && (!pr.nawsOn || pr.nawsGot) &&
		pr.ttypeAnswered && (!pr.ttypeOn || pr.r.ttype.ended)
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
				if e.On && !pr.r.ttype.ended {
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
			if pr.r.ttype.ended {
				continue
			}
			step3 = true
			if pr.r.ttype.add(e.Name) {
				pr.t.out = append(pr.t.out, telnet.SendTTYPE()...)
			}
		}
	}
	// A CPR reply before step 4 is dropped.
	for _, c := range cprs {
		if len(pr.r.cprs) < pr.probes() {
			pr.r.cprs = append(pr.r.cprs, c)
			step4 = true
		}
	}
	return step3, step4
}

// log writes the log line of section 5.6.
func (pr *probe) log(log *slog.Logger, accepted time.Time, term screen.Term, utf8 bool) {
	names := make([]string, len(pr.r.ttype.names))
	for i, n := range pr.r.ttype.names {
		names[i] = cleanName(n)
	}
	mtts := "-"
	if bits, ok := pr.r.mtts(); ok {
		mtts = fmt.Sprint(bits)
	}
	naws := "-"
	if pr.r.naws.Cols > 0 {
		naws = fmt.Sprintf("%dx%d", pr.r.naws.Cols, pr.r.naws.Rows)
	}
	cprs := make([]string, len(pr.r.cprs))
	for i, c := range pr.r.cprs {
		cprs[i] = fmt.Sprintf("%d;%d", c.Row, c.Col)
	}
	charset := term.Charset
	if !utf8 {
		charset = "not utf-8"
	}
	log.Info("probe",
		"ttype", names, "mtts", mtts, "naws", naws, "cpr", cprs,
		"type", term.Type, "cols", term.Cols, "rows", term.Rows,
		"charset", charset, "colors", term.Colors,
		"time", time.Since(accepted))
}
