// Package telnet holds the telnet parser, the option states and the subnegotiations
// of the telnet-term specification, section 4. It has no network code:
// it takes bytes and gives data, events and replies.
package telnet

import (
	"errors"
	"time"
)

// The telnet options of the host.
const (
	OptBinary byte = 0
	OptEcho   byte = 1
	OptSGA    byte = 3
	OptTTYPE  byte = 24
	OptNAWS   byte = 31
)

// The telnet command bytes.
const (
	cmdSE   byte = 240
	cmdBRK  byte = 243
	cmdIP   byte = 244
	cmdAYT  byte = 246
	cmdSB   byte = 250
	cmdWILL byte = 251
	cmdWONT byte = 252
	cmdDO   byte = 253
	cmdDONT byte = 254
	cmdIAC  byte = 255
)

// The TTYPE subnegotiation codes.
const (
	ttypeIS   byte = 0
	ttypeSEND byte = 1
)

// MaxSub is the maximum data length of a subnegotiation, after IAC IAC gives 0xFF
// and without the option byte.
const MaxSub = 256

// SubTimeout is the time in which a subnegotiation must end.
const SubTimeout = 10 * time.Second

// The errors of Feed. Each ends the session.
var (
	ErrHangup     = errors.New("telnet: interrupt or break from the client")
	ErrSubTooLong = errors.New("telnet: subnegotiation too long")
	ErrSubTimeout = errors.New("telnet: subnegotiation did not end in time")
)

// Event is one of Option, NAWS and TTYPE.
type Event interface{ event() }

// Option tells that the state of an option changed: the client answered a request
// of the host, or the host agreed to a request of the client.
// Local is true for an option on the host side. On is the new state.
type Option struct {
	Opt   byte
	Local bool
	On    bool
}

// NAWS is a window size from the client. A value can be 0.
type NAWS struct{ Cols, Rows int }

// TTYPE is a terminal type name from the client, as it was sent.
type TTYPE struct{ Name string }

func (Option) event() {}
func (NAWS) event()   {}
func (TTYPE) event()  {}

// Result is the output of Feed. The caller writes Reply to the client
// before it gives more input to Feed.
type Result struct {
	Data   []byte
	Events []Event
	Reply  []byte
}

// optState is the state of one side of an option (RFC 1143). The host never asks
// to disable an option, so the states WANTNO and the queue bit are not necessary.
type optState uint8

const (
	stNo optState = iota
	stYes
	stWantYes
)

// The options that the host accepts on each side.
var (
	acceptLocal  = [256]bool{OptEcho: true, OptSGA: true, OptBinary: true}
	acceptRemote = [256]bool{OptSGA: true, OptBinary: true, OptNAWS: true, OptTTYPE: true}
)

// parseState is the state of the byte parser.
type parseState uint8

const (
	psData   parseState = iota
	psIAC               // after IAC
	psVerb              // after IAC and WILL, WONT, DO or DONT
	psSub               // in a subnegotiation
	psSubIAC            // after IAC in a subnegotiation
)

// Parser is the telnet parser of one connection.
type Parser struct {
	ps       parseState
	verb     byte
	local    [256]optState
	remote   [256]optState
	sub      []byte // the option byte, then the data
	subStart time.Time
}

// New gives a parser with all options disabled.
func New() *Parser {
	return &Parser{}
}

// Start gives the requests that the host sends at connect (telnet-term, section 4.1)
// and sets the state of these options to "want yes".
func (p *Parser) Start() []byte {
	reqs := []struct{ verb, opt byte }{
		{cmdWILL, OptEcho},
		{cmdWILL, OptSGA}, {cmdDO, OptSGA},
		{cmdWILL, OptBinary}, {cmdDO, OptBinary},
		{cmdDO, OptNAWS},
		{cmdDO, OptTTYPE},
	}
	var b []byte
	for _, r := range reqs {
		if r.verb == cmdWILL {
			p.local[r.opt] = stWantYes
		} else {
			p.remote[r.opt] = stWantYes
		}
		b = append(b, cmdIAC, r.verb, r.opt)
	}
	return b
}

// SendTTYPE gives the TTYPE SEND request.
func SendTTYPE() []byte {
	return []byte{cmdIAC, cmdSB, OptTTYPE, ttypeSEND, cmdIAC, cmdSE}
}

// Escape appends b to dst with each 0xFF doubled (telnet-term, section 4.3).
func Escape(dst, b []byte) []byte {
	for _, c := range b {
		dst = append(dst, c)
		if c == cmdIAC {
			dst = append(dst, cmdIAC)
		}
	}
	return dst
}

// Deadline gives the time at which an open subnegotiation must end.
// ok is false when no subnegotiation is open.
func (p *Parser) Deadline() (t time.Time, ok bool) {
	if p.ps != psSub && p.ps != psSubIAC {
		return time.Time{}, false
	}
	return p.subStart.Add(SubTimeout), true
}

// Feed parses the input that arrived at the time now. After an error,
// the caller ends the session.
func (p *Parser) Feed(now time.Time, in []byte) (Result, error) {
	var r Result
	if d, ok := p.Deadline(); ok && now.After(d) {
		return r, ErrSubTimeout
	}
	for _, b := range in {
		switch p.ps {
		case psData:
			if b == cmdIAC {
				p.ps = psIAC
			} else {
				r.Data = append(r.Data, b)
			}
		case psIAC:
			if err := p.command(now, b, &r); err != nil {
				return r, err
			}
		case psVerb:
			p.negotiate(p.verb, b, &r)
			p.ps = psData
		case psSub:
			if b == cmdIAC {
				p.ps = psSubIAC
			} else if err := p.subByte(b); err != nil {
				return r, err
			}
		case psSubIAC:
			switch b {
			case cmdIAC:
				p.ps = psSub
				if err := p.subByte(b); err != nil {
					return r, err
				}
			case cmdSE:
				p.ps = psData
				p.subEnd(&r)
			default:
				// IAC and another byte drop the subnegotiation and are a command.
				if err := p.command(now, b, &r); err != nil {
					return r, err
				}
			}
		}
	}
	return r, nil
}

// command handles the byte b after IAC outside a subnegotiation.
func (p *Parser) command(now time.Time, b byte, r *Result) error {
	p.ps = psData
	switch b {
	case cmdIAC:
		r.Data = append(r.Data, cmdIAC)
	case cmdWILL, cmdWONT, cmdDO, cmdDONT:
		p.verb = b
		p.ps = psVerb
	case cmdSB:
		p.ps = psSub
		p.sub = p.sub[:0]
		p.subStart = now
	case cmdAYT:
		r.Reply = append(r.Reply, 0x07)
	case cmdIP, cmdBRK:
		return ErrHangup
	}
	return nil
}

// negotiate handles a WILL, WONT, DO or DONT request for the option opt (telnet-term, section 4.1).
func (p *Parser) negotiate(verb, opt byte, r *Result) {
	local := verb == cmdDO || verb == cmdDONT
	on := verb == cmdWILL || verb == cmdDO
	st, accept, yes, no := &p.remote[opt], acceptRemote[opt], cmdDO, cmdDONT
	if local {
		st, accept, yes, no = &p.local[opt], acceptLocal[opt], cmdWILL, cmdWONT
	}
	switch {
	case on && *st == stNo && !accept:
		r.Reply = append(r.Reply, cmdIAC, no, opt)
		return
	case on && *st == stNo:
		*st = stYes
		r.Reply = append(r.Reply, cmdIAC, yes, opt)
	case !on && *st == stYes:
		*st = stNo
		r.Reply = append(r.Reply, cmdIAC, no, opt)
	case *st == stWantYes:
		// The answer to a request of the host.
		*st = stNo
		if on {
			*st = stYes
		}
	default:
		// The request matches the current state.
		return
	}
	r.Events = append(r.Events, Option{Opt: opt, Local: local, On: on})
}

// subByte appends a byte to the subnegotiation.
func (p *Parser) subByte(b byte) error {
	if len(p.sub) > MaxSub {
		return ErrSubTooLong
	}
	p.sub = append(p.sub, b)
	return nil
}

// subEnd handles a complete subnegotiation.
func (p *Parser) subEnd(r *Result) {
	if len(p.sub) == 0 || p.remote[p.sub[0]] != stYes {
		return
	}
	data := p.sub[1:]
	switch p.sub[0] {
	case OptNAWS:
		if len(data) == 4 {
			r.Events = append(r.Events, NAWS{
				Cols: int(data[0])<<8 | int(data[1]),
				Rows: int(data[2])<<8 | int(data[3]),
			})
		}
	case OptTTYPE:
		if len(data) > 0 && data[0] == ttypeIS {
			r.Events = append(r.Events, TTYPE{Name: string(data[1:])})
		}
	}
}
