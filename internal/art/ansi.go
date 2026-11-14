package art

import (
	"fmt"

	"newbbs/internal/screen"
)

// MaxParam is the largest value of a CSI parameter. A larger parameter counts as MaxParam.
const MaxParam = 9999

// CSI is a CSI sequence after "ESC [" (ECMA-48, section 5.4).
type CSI struct {
	Params, Inter []byte
	Final         byte
	Text          string // the sequence for a message
}

// ScanCSI delimits the CSI sequence that starts after "ESC [" at data[i]: parameter bytes,
// intermediate bytes and a final byte. It gives the index after the sequence.
// When a byte outside these ranges or the end of the data comes first, the sequence
// is incomplete: ok is false and the index is that of the byte, which plays.
func ScanCSI(data []byte, i int) (seq CSI, next int, ok bool) {
	j := i
	for j < len(data) && 0x30 <= data[j] && data[j] <= 0x3f {
		j++
	}
	k := j
	for k < len(data) && 0x20 <= data[k] && data[k] <= 0x2f {
		k++
	}
	if k == len(data) || data[k] < 0x40 || data[k] > 0x7e {
		return CSI{Text: SeqText(data[i:k])}, k, false
	}
	return CSI{Params: data[i:j], Inter: data[j:k], Final: data[k], Text: SeqText(data[i : k+1])}, k + 1, true
}

// SeqText gives a CSI sequence for a message. It cuts a long sequence.
func SeqText(b []byte) string {
	const maxLen = 24
	if len(b) > maxLen {
		return fmt.Sprintf("%q", "ESC["+string(b[:maxLen])+"...")
	}
	return fmt.Sprintf("%q", "ESC["+string(b))
}

// AppendParams parses the parameters b: digits, separated by ";". It appends them to dst.
// A missing parameter is 0. It reports false for another byte.
func AppendParams(dst []int, b []byte) ([]int, bool) {
	v := 0
	for _, c := range b {
		switch {
		case c == ';':
			dst = append(dst, v)
			v = 0
		case '0' <= c && c <= '9':
			v = min(v*10+int(c-'0'), MaxParam)
		default:
			return dst, false
		}
	}
	return append(dst, v), true
}

// ExtColour parses the SGR parameters after 38 or 48: "5;n" or "2;r;g;b".
// It gives the colour and the count of parameters.
func ExtColour(ps []int) (screen.Colour, int, bool) {
	switch {
	case len(ps) >= 2 && ps[0] == 5 && ps[1] <= 255:
		return screen.IndexColour(uint8(ps[1])), 2, true
	case len(ps) >= 4 && ps[0] == 2 && ps[1] <= 255 && ps[2] <= 255 && ps[3] <= 255:
		return screen.RGBColour(uint8(ps[1]), uint8(ps[2]), uint8(ps[3])), 4, true
	}
	return screen.Colour{}, 0, false
}
