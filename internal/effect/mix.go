package effect

import (
	"math"

	"github.com/schors/velourbbs/internal/screen"
)

// mix gives the foreground of the mix of the colour c with the colour p at the amount a, from 0 to 1
// (live-screen, section 6.3): c + (p - c) × a for each RGB channel, rounded. normal is the foreground
// of normal: the value of default.
func mix(c, p screen.Colour, a float64, normal screen.Colour) screen.Colour {
	from, to := rgb(c, normal), rgb(p, normal)
	var out uint32
	for shift := 16; shift >= 0; shift -= 8 {
		x, y := float64(from>>shift&0xff), float64(to>>shift&0xff)
		out |= uint32(math.Floor(x+(y-x)*a+0.5)) << shift
	}
	return screen.Colour{Kind: screen.RGB, Value: out}
}

// rgb gives the value 0xRRGGBB of the colour c: an index or a VGA colour by the palette,
// default by the colour normal, or aaaaaa when normal is default too.
func rgb(c, normal screen.Colour) uint32 {
	switch c.Kind {
	case screen.Index, screen.VGA:
		return screen.Palette(c.Value)
	case screen.RGB:
		return c.Value
	}
	if normal.Kind == screen.Default {
		return 0xaaaaaa
	}
	return rgb(normal, normal)
}
