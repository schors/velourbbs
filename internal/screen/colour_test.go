package screen

import "testing"

func TestPalette(t *testing.T) {
	for n, want := range map[uint32]uint32{
		0: 0x000000, 3: 0xaa5500, 15: 0xffffff,
		16: 0x000000, 17: 0x00005f, 52: 0x5f0000, 208: 0xff8700, 231: 0xffffff,
		232: 0x080808, 244: 0x808080, 255: 0xeeeeee,
	} {
		if got := Palette(n); got != want {
			t.Errorf("Palette(%d) = %06x, want %06x", n, got, want)
		}
	}
}
