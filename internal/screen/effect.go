package screen

import "time"

// EffectKind is the key do of an effect (live-screen, section 6.1).
type EffectKind uint8

// The effects of live-screen, sections 6.4–6.9.
const (
	Twinkle EffectKind = iota + 1
	Sparkle
	Candle
	Glint
	Sweep
	Lightning
)

// Point is a cell [col, row] of a place, from 1. A negative coordinate counts from the end
// (live-screen, section 6.2).
type Point struct{ Col, Row int }

// AtKind is the kind of the key at of an effect (live-screen, section 6.2).
type AtKind uint8

// The kinds of at. AtAll is no at: the whole art or grid.
const (
	AtAll AtKind = iota
	AtPoint
	AtRect
	AtPoints
)

// At is the key at of an effect. Points holds the point, the first cell of the rectangle
// or the points. Width and Height are the size of a rectangle.
type At struct {
	Kind          AtKind
	Points        []Point
	Width, Height int
}

// Effect is one [[effect]] table of the front matter (live-screen, section 6).
// Each key holds its value or its default. Every is 0 for one event: a sweep without every.
type Effect struct {
	Do                   EffectKind
	In                   string // the art name of in, else "": the page
	At                   At
	Delay, Every, Spread time.Duration
	Chars                []string // the graphemes of chars
	Colors               []string // the colour values of colors (ColourValue)
	Peak                 string   // a colour value
	Share                float64
	Over                 bool
	Count                int
	Time                 time.Duration
	Width                int
	Speed                float64
}

// ColourValue gives the colour of the colour value v in the session theme theme (live-screen, section 6.1):
// the foreground of the style v, or of the theme style with the name v. The load checks v.
func ColourValue(v string, theme map[string]Style) Colour {
	if IsIdent(v) {
		return theme[v].Fg
	}
	s, _ := ParseStyle(v)
	return s.Fg
}
