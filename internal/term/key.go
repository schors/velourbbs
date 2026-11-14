package term

import (
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/schors/velourbbs/internal/screen"
)

// EscWait is the wait for the next byte after ESC and after each byte
// of an escape sequence (section 8.2).
const EscWait = 150 * time.Millisecond

// maxSeq is the maximum length of an escape sequence, ESC included (section 8.2).
const maxSeq = 32

// CPR is a cursor position report ESC[<row>;<col>R with 1–5 digits in each number
// (section 5.4). A number can be 0.
type CPR struct{ Row, Col int }

// csiKeys maps the bytes of a CSI sequence after ESC[ to the key (section 8.2).
var csiKeys = map[string]string{
	"A": "UP", "B": "DOWN", "C": "RIGHT", "D": "LEFT",
	"H": "HOME", "1~": "HOME", "7~": "HOME",
	"F": "END", "4~": "END", "8~": "END",
	"3~": "DEL", "5~": "PGUP", "6~": "PGDN",
	"11~": "F1", "12~": "F2", "13~": "F3", "14~": "F4",
	"15~": "F5", "17~": "F6", "18~": "F7", "19~": "F8",
	"20~": "F9", "21~": "F10", "23~": "F11", "24~": "F12",
}

// ss3Keys maps the final byte of an SS3 sequence to the key (section 8.2).
var ss3Keys = map[rune]string{
	'A': "UP", 'B': "DOWN", 'C': "RIGHT", 'D': "LEFT",
	'H': "HOME", 'F': "END",
	'P': "F1", 'Q': "F2", 'R': "F3", 'S': "F4",
}

// seqState is the escape sequence state of the key decoder.
type seqState uint8

const (
	seqNone    seqState = iota
	seqESC              // after ESC
	seqCSI              // in a CSI sequence
	seqBracket          // after CSI [ (Linux console F1–F5)
	seqSS3              // after ESC O
)

// Decoder is the key decoder of one connection (section 8.1, steps 2–4, and section 8.2).
// It takes the data bytes after the telnet parser. Its zero value is ready.
//
// A key is a key name of bbt-format, section 8.1, for example ENTER or F1,
// or one character, for example "a" or "ы".
type Decoder struct {
	afterCR bool              // the previous byte was CR (section 8.1, step 2)
	utf     [utf8.UTFMax]byte // an incomplete UTF-8 character
	nutf    int               // the bytes in utf
	st      seqState          // the escape sequence state
	seq     [maxSeq]byte      // the bytes of the sequence, ESC included
	nseq    int               // the bytes in seq
	long    bool              // the sequence is longer than maxSeq
	last    time.Time         // the arrival of the last byte while a sequence waits
	keys    []string          // the output of Feed
	cprs    []CPR             // the output of Feed
}

// Deadline gives the time at which the wait for the next byte after ESC
// or after a byte of a sequence ends. ok is false when no ESC or sequence waits.
// At this time the caller calls Feed with no input.
func (d *Decoder) Deadline() (t time.Time, ok bool) {
	if d.st == seqNone {
		return time.Time{}, false
	}
	return d.last.Add(EscWait), true
}

// Feed decodes the data bytes that arrived at the time now. All bytes of one call
// arrive together. First Feed ends the wait when now is at or after Deadline:
// ESC alone gives ESC, and a started sequence is dropped.
// The CPR replies are for the probe. After the probe, the caller drops them.
func (d *Decoder) Feed(now time.Time, in []byte) (keys []string, cprs []CPR) {
	d.keys, d.cprs = nil, nil
	if dl, ok := d.Deadline(); ok && !now.Before(dl) {
		if d.st == seqESC {
			d.emit("ESC")
		}
		d.st = seqNone
	}
	for _, b := range in {
		d.enter(b)
	}
	if len(in) > 0 && d.st != seqNone {
		d.last = now
	}
	return d.keys, d.cprs
}

// enter is step 2 of section 8.1: each Enter form becomes one CR.
func (d *Decoder) enter(b byte) {
	if d.afterCR {
		d.afterCR = false
		if b == '\n' || b == 0 {
			return
		}
	}
	switch b {
	case '\r':
		d.afterCR = true
	case '\n':
		b = '\r'
	}
	d.decodeUTF8(b)
}

// decodeUTF8 is step 3 of section 8.1: it drops each byte that is not valid UTF-8.
func (d *Decoder) decodeUTF8(b byte) {
	d.utf[d.nutf] = b
	d.nutf++
	for d.nutf > 0 && utf8.FullRune(d.utf[:d.nutf]) {
		r, size := utf8.DecodeRune(d.utf[:d.nutf])
		if r != utf8.RuneError || size > 1 {
			d.key(r)
		}
		d.nutf = copy(d.utf[:], d.utf[size:d.nutf])
	}
}

// key is step 4 of section 8.1 with the escape sequences of section 8.2.
func (d *Decoder) key(r rune) {
	switch d.st {
	case seqESC:
		switch r {
		case '[':
			d.st = seqCSI
			d.push(r)
			return
		case 'O':
			d.st = seqSS3
			d.push(r)
			return
		}
		d.st = seqNone
		d.emit("ESC")
	case seqCSI:
		if d.csi(r) {
			return
		}
	case seqBracket:
		d.st = seqNone
		if 'A' <= r && r <= 'E' {
			d.emit("F" + string(r-'A'+'1'))
			return
		}
	case seqSS3:
		d.st = seqNone
		if isFinal(r) {
			if k, ok := ss3Keys[r]; ok {
				d.emit(k)
			}
			return
		}
	}
	// New input.
	switch {
	case r == 0x1B:
		d.st, d.nseq, d.long = seqESC, 0, false
		d.push(r)
	case r == '\r':
		d.emit("ENTER")
	case r == 0x08 || r == 0x7F:
		d.emit("BS")
	case r == '\t':
		d.emit("TAB")
	case r == ' ':
		d.emit("SPACE")
	case screen.IsControl(r):
		// Other control characters are dropped.
	default:
		d.emit(string(r))
	}
}

// csi handles the byte r in a CSI sequence. It gives false for an invalid byte:
// the sequence is dropped and the caller handles r as new input.
func (d *Decoder) csi(r rune) bool {
	switch {
	case 0x20 <= r && r <= 0x3F: // a parameter or intermediate byte
		d.push(r)
		return true
	case isFinal(r):
		d.st = seqNone
		d.push(r)
		switch {
		case d.long:
		case r == '[' && d.nseq == 3:
			d.st = seqBracket
		default:
			d.csiEnd(string(d.seq[2:d.nseq]))
		}
		return true
	}
	d.st = seqNone
	return false
}

// csiEnd handles a complete CSI sequence. body is the bytes after ESC[.
// A sequence that is not a key or a CPR reply is dropped.
func (d *Decoder) csiEnd(body string) {
	if k, ok := csiKeys[body]; ok {
		d.emit(k)
		return
	}
	params, ok := strings.CutSuffix(body, "R")
	if !ok {
		return
	}
	row, col, ok := strings.Cut(params, ";")
	if !ok || !isNumber(row) || !isNumber(col) {
		return
	}
	r, _ := strconv.Atoi(row)
	c, _ := strconv.Atoi(col)
	d.cprs = append(d.cprs, CPR{Row: r, Col: c})
}

// isNumber reports whether s has 1–5 decimal digits and nothing else (section 5.4).
func isNumber(s string) bool {
	if len(s) < 1 || len(s) > 5 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// isFinal reports whether r is a final byte of a CSI or SS3 sequence.
func isFinal(r rune) bool {
	return 0x40 <= r && r <= 0x7E
}

// push adds the byte r to the sequence. After maxSeq bytes it only marks
// the sequence as too long, so the sequence is dropped up to its end.
// r is ASCII.
func (d *Decoder) push(r rune) {
	if d.nseq == maxSeq {
		d.long = true
		return
	}
	d.seq[d.nseq] = byte(r)
	d.nseq++
}

func (d *Decoder) emit(k string) {
	d.keys = append(d.keys, k)
}
