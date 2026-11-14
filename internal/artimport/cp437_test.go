package artimport

import (
	"testing"

	"newbbs/internal/screen"
)

// TestCP437 checks each entry class of the table of section 7.
func TestCP437(t *testing.T) {
	glyphs := map[byte]rune{
		0x01: '☺', 0x02: '☻', 0x03: '♥', 0x04: '♦', 0x05: '♣', 0x06: '♠', 0x07: '•', 0x08: '◘',
		0x09: '○', 0x0B: '♂', 0x0C: '♀', 0x0E: '♫', 0x0F: '☼', 0x10: '►', 0x11: '◄', 0x12: '↕',
		0x13: '‼', 0x14: '¶', 0x15: '§', 0x16: '▬', 0x17: '↨', 0x18: '↑', 0x19: '↓', 0x1C: '∟',
		0x1D: '↔', 0x1E: '▲', 0x1F: '▼',
		0x7F: '⌂',
		// Spot values of 0x80–0xFF.
		0x80: 'Ç', 0x9B: '¢', 0xB0: '░', 0xC9: '╔', 0xDB: '█', 0xE1: 'ß', 0xFE: '■', 0xFF: ' ',
	}
	controls := map[byte]bool{'\n': true, '\r': true, 0x1a: true, 0x1b: true}
	seen := map[rune]int{}
	for b := range 256 {
		want, ok := glyphs[byte(b)]
		switch {
		case b == 0:
			ok, want = true, ' '
		case controls[byte(b)], 0x20 <= b && b <= 0x7e:
			ok, want = true, rune(b)
		case b < 0x80 && !ok:
			t.Errorf("byte %02X has no glyph in the test", b)
		}
		if ok && cp437[b] != want {
			t.Errorf("byte %02X = %U, want %U", b, cp437[b], want)
		}
	}
	for b := 1; b < 256; b++ {
		r := cp437[b]
		if controls[byte(b)] {
			continue
		}
		if c, ok := seen[r]; ok {
			t.Errorf("bytes %02X and %02X are both %U", c, b, r)
		}
		seen[r] = b
		for g, w := range screen.Graphemes(string(r)) {
			if err := screen.CheckGrapheme(g, w); err != nil || w != 1 {
				t.Errorf("byte %02X: %U width %d, %v", b, r, w, err)
			}
		}
	}
	if got := string(decodeCP437([]byte("\x00a\r\n\x1b[1m\xdb"))); got != " a\r\n\x1b[1m█" {
		t.Errorf("decode %q", got)
	}
}
