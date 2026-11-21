// Package effect gives the layer of the effects of a shown page (docs/specs/live-screen.md, section 6):
// the places, the events and the cells of each effect. It has no I/O: it takes the time from the show
// and the random source as values.
package effect

import (
	"math"
	"math/rand/v2"
	"slices"
	"time"

	"github.com/schors/velourbbs/internal/screen"
)

// Layer is the layer of the effects of one show (live-screen, sections 5 and 6).
// It keeps the state of the effects from the show to the next show.
type Layer struct {
	rng  *rand.Rand
	fx   []*runner // in source order
	cols int       // term.cols: a cell of the grid is row*cols + col
}

// runner is one effect and its state.
type runner struct {
	e      screen.Effect
	colors []screen.Colour
	places []*place
	drawn  map[int]screen.Cell // the cells that the effect draws
	order  []int               // the keys of drawn in the order of their first draw
	star   map[int]string      // the grapheme of each star of a twinkle
}

// place is a place of an effect (live-screen, section 6.2) and the time of its next event.
type place struct {
	cells []int // the cells in the place; an event takes random cells by a partial shuffle
	point bool  // a point: each event takes its cell
	next  time.Duration
}

// New gives the layer of the effects effs on the base page p at the show (live-screen, section 6).
// theme is the session theme of the colour values; rng is the random source of the session.
// The layer of the time 0 draws the cells of the show.
func New(effs []screen.Effect, p screen.Page, theme map[string]screen.Style, rng *rand.Rand) *Layer {
	l := &Layer{rng: rng, cols: p.Cols}
	grid := p.Grid()
	for _, e := range effs {
		r := &runner{e: e, drawn: map[int]screen.Cell{}}
		for _, c := range e.Colors {
			r.colors = append(r.colors, screen.ColourValue(c, theme))
		}
		l.fx = append(l.fx, r)
		if e.Do != screen.Twinkle && e.Do != screen.Candle {
			continue // the other effects come in other blocks
		}
		for _, cells := range places(e, p, grid) {
			r.places = append(r.places, &place{cells: cells, point: e.At.Kind == screen.AtPoint || e.At.Kind == screen.AtPoints})
		}
		l.show(len(l.fx)-1, grid)
		for _, pl := range r.places {
			pl.next = e.Delay + l.every(e)
		}
	}
	return l
}

// show draws the cells of the effect i at the show (sections 6.4 and 6.6) and drops the places without cells.
// A twinkle rectangle takes its stars from the base grid.
func (l *Layer) show(i int, grid []screen.Row) {
	r := l.fx[i]
	r.star = map[int]string{}
	var live []*place
	for _, pl := range r.places {
		if r.e.Do == screen.Twinkle && !pl.point { // the stars: the cells with a grapheme of chars
			pl.cells = slices.DeleteFunc(pl.cells, func(c int) bool {
				return !slices.Contains(r.e.Chars, l.cell(grid, c).Grapheme)
			})
		}
		if len(pl.cells) == 0 {
			continue
		}
		live = append(live, pl)
		for _, c := range pl.cells {
			switch {
			case r.e.Do == screen.Candle:
				below := l.below(i, c, grid)
				below.Grapheme, below.Style.Fg = r.e.Chars[0], r.colors[0]
				r.draw(c, below)
			case pl.point && !slices.Contains(r.e.Chars, l.below(i, c, grid).Grapheme):
				r.star[c] = r.e.Chars[0]
				r.draw(c, l.paint(r, r.e.Chars[0], l.below(i, c, grid)))
			case pl.point:
				r.star[c] = l.below(i, c, grid).Grapheme
			default: // a star of a rectangle shows the cell below it until its first change
				r.star[c] = l.cell(grid, c).Grapheme
			}
		}
	}
	r.places = live
}

// every gives the time from an event to the next: every plus a random value from -spread to spread
// (section 6.3). Without spread, it takes no random value.
func (l *Layer) every(e screen.Effect) time.Duration {
	if e.Spread == 0 {
		return e.Every
	}
	return e.Every + time.Duration(l.rng.Int64N(int64(2*e.Spread)+1)) - e.Spread
}

// Next gives the time from the show of the next event, and false when the layer has no event.
func (l *Layer) Next() (time.Duration, bool) {
	_, pl := l.first()
	if pl == nil {
		return 0, false
	}
	return pl.next, true
}

// first gives the effect and the place of the next event: the earliest start, then the first effect,
// then the first place. pl is nil without an event.
func (l *Layer) first() (i int, pl *place) {
	for j, r := range l.fx {
		for _, p := range r.places {
			if pl == nil || p.next < pl.next {
				i, pl = j, p
			}
		}
	}
	return i, pl
}

// Cells runs the events up to the time t from the show and gives the layer: the cells that the effects
// change on the grid of the base, in the order of the effects (Page.Layer). base is the grid
// of the current base (screen.Page.Grid).
func (l *Layer) Cells(t time.Duration, base []screen.Row) []screen.GridCell {
	for {
		i, pl := l.first()
		if pl == nil || pl.next > t {
			break
		}
		l.event(i, pl, base)
		pl.next += l.every(l.fx[i].e)
	}
	var out []screen.GridCell
	for _, r := range l.fx {
		for _, c := range r.order {
			out = append(out, screen.GridCell{Row: c / l.cols, Col: c % l.cols, Cell: r.drawn[c]})
		}
	}
	return out
}

// event runs an event of the effect i in the place pl (sections 6.4 and 6.6).
// A rectangle takes share of its cells, rounded, but at least 1.
func (l *Layer) event(i int, pl *place, grid []screen.Row) {
	r := l.fx[i]
	n := 1
	if !pl.point {
		n = max(1, int(math.Floor(r.e.Share*float64(len(pl.cells))+0.5)))
	}
	for k := range n { // a partial shuffle: the random cells of one event are different cells
		j := k + l.rng.IntN(len(pl.cells)-k)
		pl.cells[k], pl.cells[j] = pl.cells[j], pl.cells[k]
		c := pl.cells[k]
		below := l.below(i, c, grid)
		switch r.e.Do {
		case screen.Twinkle:
			g := l.another(r.e.Chars, r.star[c])
			r.star[c] = g
			r.draw(c, l.paint(r, g, below))
		case screen.Candle:
			below.Grapheme = r.e.Chars[l.rng.IntN(len(r.e.Chars))]
			below.Style.Fg = r.colors[l.rng.IntN(len(r.colors))]
			r.draw(c, below)
		}
	}
}

// another gives a random grapheme of chars other than g.
func (l *Layer) another(chars []string, g string) string {
	var other []string
	for _, c := range chars {
		if c != g {
			other = append(other, c)
		}
	}
	return other[l.rng.IntN(len(other))]
}

// paint gives the cell c with the drawn grapheme g (section 6.3): the foreground is a random value
// of colors when the effect has colors, else the foreground of c. The background and the attributes stay.
func (l *Layer) paint(r *runner, g string, c screen.Cell) screen.Cell {
	c.Grapheme = g
	if len(r.colors) > 0 {
		c.Style.Fg = r.colors[l.rng.IntN(len(r.colors))]
	}
	return c
}

// draw sets the cell c of the effect to v.
func (r *runner) draw(c int, v screen.Cell) {
	if _, ok := r.drawn[c]; !ok {
		r.order = append(r.order, c)
	}
	v.Slot = 0
	r.drawn[c] = v
}

// cell gives the cell c of grid.
func (l *Layer) cell(grid []screen.Row, c int) screen.Cell { return grid[c/l.cols][c%l.cols] }

// below gives the cell c after the effects before the effect i (section 6.3).
func (l *Layer) below(i, c int, grid []screen.Row) screen.Cell {
	v := l.cell(grid, c)
	for _, r := range l.fx[:i] {
		if d, ok := r.drawn[c]; ok {
			v = d
		}
	}
	return v
}
