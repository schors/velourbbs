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
	rng    *rand.Rand
	fx     []*runner     // in source order
	cols   int           // term.cols: a cell of the grid is row*cols + col
	normal screen.Colour // the foreground of normal: the colour of default in the mix
	now    time.Duration // the time of the last Cells
}

// runner is one effect and its state.
type runner struct {
	e      screen.Effect
	colors []screen.Colour
	peak   screen.Colour
	places []*place
	drawn  map[int]mark    // the cells that the effect draws
	order  []int           // the keys of drawn in the order of their first draw
	star   map[int]string  // the grapheme of each star of a twinkle
	runs   []*run          // the events of a glint or a sweep that run, in the order of their starts
	lit    map[int]float64 // the amount of the mix of each cell at the time of the layer
	litOrd []int           // the keys of lit in the order of the runs and of their cells
}

// mark is a drawn cell: its grapheme, and its foreground when ownFg is true. The other parts of the cell
// come from the cell below it at each frame, so a new base (a cursor move) shows through (section 6.3).
type mark struct {
	g     string
	fg    screen.Colour
	ownFg bool
}

// run is an event of a glint or a sweep that runs (live-screen, sections 6.7 and 6.8):
// its start and its end from the show, and its cells. left and stop are the columns
// where the centre of the band of a sweep starts and stops.
type run struct {
	start, end time.Duration
	left, stop int
	cells      []litCell
}

// litCell is a cell of a run: band is true for a cell of the band of a sweep, and glint is the time
// of its glint from the start of the run, or -1 without a glint.
type litCell struct {
	c, col int
	band   bool
	glint  time.Duration
}

// place is a place of an effect (live-screen, section 6.2) and the time of its next event.
type place struct {
	cells []int // the cells in the place; an event takes random cells by a partial shuffle
	point bool  // a point: each event takes its cell
	next  time.Duration
	done  bool // the place has no next event: the one event of a sweep without every started
}

// New gives the layer of the effects effs on the base page p at the show (live-screen, section 6).
// theme is the session theme of the colour values; rng is the random source of the session.
// The layer of the time 0 draws the cells of the show.
func New(effs []screen.Effect, p screen.Page, theme map[string]screen.Style, rng *rand.Rand) *Layer {
	l := &Layer{rng: rng, cols: p.Cols, normal: theme["normal"].Fg}
	grid := p.Grid()
	for _, e := range effs {
		r := &runner{e: e, drawn: map[int]mark{}, lit: map[int]float64{}}
		for _, c := range e.Colors {
			r.colors = append(r.colors, screen.ColourValue(c, theme))
		}
		if e.Peak != "" {
			r.peak = screen.ColourValue(e.Peak, theme)
		}
		l.fx = append(l.fx, r)
		if e.Do == screen.Sparkle || e.Do == screen.Lightning {
			continue // these effects come in another block
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
// A twinkle rectangle takes its stars from the base grid. A rectangle without stars keeps its events:
// they find no cell.
func (l *Layer) show(i int, grid []screen.Row) {
	r := l.fx[i]
	r.star = map[int]string{}
	var live []*place
	for _, pl := range r.places {
		if len(pl.cells) == 0 {
			continue
		}
		live = append(live, pl)
		if r.e.Do == screen.Glint || r.e.Do == screen.Sweep { // they draw nothing at the show
			continue
		}
		if r.e.Do == screen.Twinkle && !pl.point { // the stars: the cells with a grapheme of chars
			pl.cells = slices.DeleteFunc(pl.cells, func(c int) bool {
				return !slices.Contains(r.e.Chars, l.cell(grid, c).Grapheme)
			})
		}
		for _, c := range pl.cells {
			switch {
			case r.e.Do == screen.Candle:
				r.draw(c, mark{g: r.e.Chars[0], fg: r.colors[0], ownFg: true})
			case pl.point && !slices.Contains(r.e.Chars, l.below(i, c, grid).Grapheme):
				r.star[c] = r.e.Chars[0]
				r.draw(c, l.paint(r, r.e.Chars[0]))
			case pl.point: // the star is the cell below it at the show: it stays over the effects before
				below := l.below(i, c, grid)
				r.star[c] = below.Grapheme
				r.draw(c, mark{g: below.Grapheme, fg: below.Style.Fg, ownFg: true})
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

// step is the time between two frames while an event of a glint or a sweep runs (live-screen, section 5).
const step = 50 * time.Millisecond

// Next gives the time from the show when the layer changes next, and false when it does not change:
// the start of the next event, the end of an event that runs, or step after the last Cells
// while an event runs (live-screen, section 5).
func (l *Layer) Next() (time.Duration, bool) {
	next, ok := time.Duration(0), false
	if _, pl := l.first(); pl != nil {
		next, ok = pl.next, true
	}
	for _, r := range l.fx {
		for _, rn := range r.runs {
			if d := min(rn.end, l.now+step); !ok || d < next {
				next, ok = d, true
			}
		}
	}
	return next, ok
}

// first gives the effect and the place of the next event: the earliest start, then the first effect,
// then the first place. pl is nil without an event.
func (l *Layer) first() (i int, pl *place) {
	for j, r := range l.fx {
		for _, p := range r.places {
			if !p.done && (pl == nil || p.next < pl.next) {
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
		if l.fx[i].e.Every == 0 { // a sweep without every has one event
			pl.done = true
			continue
		}
		pl.next += l.every(l.fx[i].e)
	}
	l.now = t
	var out []screen.GridCell
	for i, r := range l.fx {
		r.light(t)
		for _, c := range slices.Concat(r.order, r.litOrd) {
			row, col := c/l.cols, c%l.cols
			if !free(base[row], col) { // a new base can have a slot or a wide grapheme here
				continue
			}
			out = append(out, screen.GridCell{Row: row, Col: col, Cell: l.over(i, c, base)})
		}
	}
	return out
}

// over gives the cell c of the effect i over the cell below it.
func (l *Layer) over(i, c int, grid []screen.Row) screen.Cell {
	return l.apply(l.fx[i], c, l.below(i, c, grid))
}

// apply gives the cell v after the effect r on the cell c: a drawn grapheme, or the mix of its foreground.
func (l *Layer) apply(r *runner, c int, v screen.Cell) screen.Cell {
	if m, ok := r.drawn[c]; ok {
		v.Grapheme, v.Slot = m.g, 0
		if m.ownFg {
			v.Style.Fg = m.fg
		}
	}
	if a, ok := r.lit[c]; ok {
		v.Style.Fg = mix(v.Style.Fg, r.peak, a, l.normal)
	}
	return v
}

// event runs an event of the effect i in the place pl (sections 6.4 and 6.6–6.8).
// A rectangle of a twinkle or a candle takes share of its cells, rounded, but at least 1.
func (l *Layer) event(i int, pl *place, grid []screen.Row) {
	r := l.fx[i]
	switch r.e.Do {
	case screen.Glint:
		l.glint(i, pl, grid)
		return
	case screen.Sweep:
		l.sweep(i, pl, grid)
		return
	}
	if len(pl.cells) == 0 { // a twinkle rectangle without stars
		return
	}
	n := 1
	if !pl.point {
		n = max(1, int(math.Floor(r.e.Share*float64(len(pl.cells))+0.5)))
	}
	for k := range n { // a partial shuffle: the random cells of one event are different cells
		j := k + l.rng.IntN(len(pl.cells)-k)
		pl.cells[k], pl.cells[j] = pl.cells[j], pl.cells[k]
		c := pl.cells[k]
		switch r.e.Do {
		case screen.Twinkle:
			g := l.another(r.e.Chars, r.star[c])
			r.star[c] = g
			r.draw(c, l.paint(r, g))
		case screen.Candle:
			g := r.e.Chars[l.rng.IntN(len(r.e.Chars))]
			r.draw(c, mark{g: g, fg: r.colors[l.rng.IntN(len(r.colors))], ownFg: true})
		}
	}
}

// nonSpace gives the cells of the place pl that are not spaces after the effects before the effect i.
// A cell that is now a slot or a part of a wide grapheme (a new base) is not in the place.
func (l *Layer) nonSpace(i int, pl *place, grid []screen.Row) []int {
	var out []int
	for _, c := range pl.cells {
		if free(grid[c/l.cols], c%l.cols) && l.below(i, c, grid).Grapheme != " " {
			out = append(out, c)
		}
	}
	return out
}

// glint starts a glint of the effect i in the place pl (section 6.7): one random cell of the cells
// that are not spaces. A point is its cell.
func (l *Layer) glint(i int, pl *place, grid []screen.Row) {
	r := l.fx[i]
	cells := l.nonSpace(i, pl, grid)
	if len(cells) == 0 {
		return
	}
	c := cells[0]
	if !pl.point {
		c = cells[l.rng.IntN(len(cells))]
	}
	r.runs = append(r.runs, &run{start: pl.next, end: pl.next + r.e.Time, cells: []litCell{{c: c, glint: 0}}})
}

// sweep starts a sweep of the effect i in the place pl (section 6.8). The centre of the band starts
// at the column L - width and stops at the column R + width. The last lit cell of each row
// starts a glint when the centre is at its column.
func (l *Layer) sweep(i int, pl *place, grid []screen.Row) {
	r := l.fx[i]
	cells := l.nonSpace(i, pl, grid)
	if len(cells) == 0 {
		return
	}
	left, right := l.cols, -1
	last := map[int]int{} // the last lit column of each row
	for _, c := range cells {
		row, col := c/l.cols, c%l.cols
		left, right = min(left, col), max(right, col)
		last[row] = max(last[row], col)
	}
	rn := &run{start: pl.next, left: left - r.e.Width, stop: right + r.e.Width}
	end := r.at(rn.stop - rn.left)
	for _, c := range cells {
		row, col := c/l.cols, c%l.cols
		lc := litCell{c: c, col: col, band: true, glint: -1}
		if last[row] == col {
			lc.glint = r.at(col - rn.left)
			end = max(end, lc.glint+r.e.Time)
		}
		rn.cells = append(rn.cells, lc)
	}
	rn.end = pl.next + end
	r.runs = append(r.runs, rn)
}

// at gives the time of a sweep that its centre takes to run n columns.
func (r *runner) at(n int) time.Duration {
	return time.Duration(math.Round(float64(n) / r.e.Speed * float64(time.Second)))
}

// light drops the events of the effect that end at the time t or before, and sets the amount
// of each cell at the time t. On a cell of two events, the event that started later wins.
func (r *runner) light(t time.Duration) {
	r.runs = slices.DeleteFunc(r.runs, func(rn *run) bool { return rn.end <= t })
	clear(r.lit)
	r.litOrd = r.litOrd[:0]
	for _, rn := range r.runs {
		for _, lc := range rn.cells {
			a := r.amount(rn, lc, t)
			if a <= 0 {
				continue
			}
			if _, ok := r.lit[lc.c]; !ok {
				r.litOrd = append(r.litOrd, lc.c)
			}
			r.lit[lc.c] = a
		}
	}
}

// amount gives the amount of the mix of the cell lc of the event rn at the time t from the show:
// the larger amount of the band and of the glint (sections 6.7 and 6.8).
func (r *runner) amount(rn *run, lc litCell, t time.Duration) float64 {
	t -= rn.start
	a := 0.0
	if lc.band {
		centre := float64(rn.left) + min(t.Seconds()*r.e.Speed, float64(rn.stop-rn.left))
		if x := math.Abs(float64(lc.col) - centre); x < float64(r.e.Width) {
			a = 1 - x/float64(r.e.Width)
		}
	}
	if lc.glint >= 0 && t >= lc.glint && t < lc.glint+r.e.Time {
		a = max(a, 1-float64(t-lc.glint)/float64(r.e.Time))
	}
	return a
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

// paint gives the mark of the drawn grapheme g (section 6.3): its foreground is a random value
// of colors when the effect has colors, else the foreground of the cell below it.
func (l *Layer) paint(r *runner, g string) mark {
	m := mark{g: g}
	if len(r.colors) > 0 {
		m.fg, m.ownFg = r.colors[l.rng.IntN(len(r.colors))], true
	}
	return m
}

// draw sets the cell c of the effect to m.
func (r *runner) draw(c int, m mark) {
	if _, ok := r.drawn[c]; !ok {
		r.order = append(r.order, c)
	}
	r.drawn[c] = m
}

// cell gives the cell c of grid.
func (l *Layer) cell(grid []screen.Row, c int) screen.Cell { return grid[c/l.cols][c%l.cols] }

// below gives the cell c after the effects before the effect i (section 6.3).
func (l *Layer) below(i, c int, grid []screen.Row) screen.Cell {
	v := l.cell(grid, c)
	for _, r := range l.fx[:i] {
		v = l.apply(r, c, v)
	}
	return v
}
