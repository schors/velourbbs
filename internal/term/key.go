package term

import (
	"time"
	"unicode/utf8"

	"github.com/schors/velourbbs/internal/screen"
)

// EscWait is the wait for the next byte after ESC and after the byte that starts
// a sequence or a control string (section 8.2).
const EscWait = 150 * time.Millisecond

// SeqWait is the wait for the next byte after each later byte of a sequence
// or a control string (section 8.2).
const SeqWait = time.Second

// maxSeq is the maximum length of a sequence or a control string in bytes,
// ESC included (section 8.2).
const maxSeq = 128

// Reply is a complete CSI sequence that is not a key, or a complete control string.
// The probe takes the replies to its queries from them (sections 5.4 and 5.7).
type Reply struct {
	Intro byte   // the byte after ESC: '[' (CSI), ']' (OSC), 'P' (DCS) or '_' (APC)
	Body  string // the bytes after Intro, without the terminator of a control string
	BEL   bool   // the control string ends with BEL, not with ST
}

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
	seqNone      seqState = iota
	seqESC                // after ESC
	seqCSI                // in a CSI sequence
	seqBracket            // after CSI [ (Linux console F1–F5)
	seqSS3                // after ESC O
	seqString             // in an OSC, DCS or APC control string
	seqStringESC          // after ESC in a control string
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
	seq     [maxSeq]byte      // the bytes of the sequence or the control string, ESC included
	nseq    int               // the bytes in seq
	long    bool              // the sequence or the control string is longer than maxSeq
	later   bool              // a byte arrived after the byte that starts the sequence
	last    time.Time         // the arrival of the last byte while a sequence waits
	keys    []string          // the output of Feed
	replies []Reply           // the output of Feed
}

// Deadline gives the time at which the wait for the next byte after ESC
// or after a byte of a sequence or a control string ends. ok is false when no ESC,
// sequence or control string waits. At this time the caller calls Feed with no input.
func (d *Decoder) Deadline() (t time.Time, ok bool) {
	if d.st == seqNone {
		return time.Time{}, false
	}
	if d.later {
		return d.last.Add(SeqWait), true
	}
	return d.last.Add(EscWait), true
}

// Feed decodes the data bytes that arrived at the time now. All bytes of one call
// arrive together. First Feed ends the wait when now is at or after Deadline:
// ESC alone gives ESC, and a started sequence or control string is dropped.
// The replies are for the probe. After the probe, the caller drops them.
func (d *Decoder) Feed(now time.Time, in []byte) (keys []string, replies []Reply) {
	d.keys, d.replies = nil, nil
	if dl, ok := d.Deadline(); ok && !now.Before(dl) {
		if d.st == seqESC {
			d.emit("ESC")
		}
		d.st = seqNone
	}
	for _, b := range in {
		// Each byte counts, also a byte of an incomplete character and a byte
		// that the UTF-8 decoder drops.
		if d.st != seqNone && d.st != seqESC {
			d.later = true
		}
		d.enter(b)
	}
	if len(in) > 0 && d.st != seqNone {
		d.last = now
	}
	return d.keys, d.replies
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

// key is step 4 of section 8.1 with the escape sequences and the control strings
// of section 8.2.
func (d *Decoder) key(r rune) {
	if d.st == seqStringESC {
		d.st = seqNone
		if r == '\\' { // ST ends the control string
			d.reply(d.nseq-1, 1, false) // the ESC of ST is in seq, \ is not
			return
		}
		// Another byte ends the control string. ESC is new input.
		d.startESC()
	}
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
		case ']', 'P', '_':
			d.st = seqString
			d.push(r)
			return
		}
		d.st = seqNone
		d.emit("ESC")
	case seqString:
		switch r {
		case 0x07: // BEL ends the control string
			d.st = seqNone
			d.reply(d.nseq, 1, true)
		case 0x1B:
			d.st = seqStringESC
			d.push(r)
		default:
			d.push(r)
		}
		return
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
		d.startESC()
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
// A sequence that is not a key is a reply.
func (d *Decoder) csiEnd(body string) {
	if k, ok := csiKeys[body]; ok {
		d.emit(k)
		return
	}
	d.reply(d.nseq, 0, false)
}

// reply gives the complete sequence or control string as a reply: its body is
// seq[2:end]. term is the bytes of the terminator that are not in seq, bel tells
// that BEL ends a control string. A reply longer than maxSeq with its terminator is dropped.
func (d *Decoder) reply(end, term int, bel bool) {
	if !d.long && d.nseq+term <= maxSeq {
		d.replies = append(d.replies, Reply{Intro: d.seq[1], Body: string(d.seq[2:end]), BEL: bel})
	}
}

// isFinal reports whether r is a final byte of a CSI or SS3 sequence.
func isFinal(r rune) bool {
	return 0x40 <= r && r <= 0x7E
}

// startESC starts a new sequence with ESC.
func (d *Decoder) startESC() {
	d.st, d.nseq, d.long, d.later = seqESC, 0, false, false
	d.push(0x1B)
}

// push adds the UTF-8 bytes of r to the sequence or the control string.
// After maxSeq bytes it only marks it as too long, so it is dropped up to its end.
func (d *Decoder) push(r rune) {
	n := utf8.RuneLen(r)
	if d.long || d.nseq+n > maxSeq {
		d.long = true
		return
	}
	d.nseq += utf8.EncodeRune(d.seq[d.nseq:], r)
}

func (d *Decoder) emit(k string) {
	d.keys = append(d.keys, k)
}
