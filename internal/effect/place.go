package effect

import "github.com/schors/velourbbs/internal/screen"

// rect is a rectangle of the grid: the first cell and the size, from 0.
type rect struct{ col, row, width, height int }

// places gives each place of the effect e on the page p with the grid grid (live-screen, section 6.2):
// one place for an at without points, else one place for each point. A cell is row*cols + col.
// A cell that the page does not show, a slot cell and a cell of a grapheme of width 2 are not in a place.
// The rectangle of a place is the rectangle of at, the art grid, the fragment rows or the grid,
// cut to the art grid or the fragment rows of in and to the grid.
// An effect with in on a page without the art or the fragment has no place.
func places(e screen.Effect, p screen.Page, grid []screen.Row) []*place {
	// frame is the cells where the coordinates count: the art grid, the fragment rows or the page,
	// in grid coordinates. bounds is the cut of the rectangle of a place, and shown is the part
	// of the grid that can be in a place.
	frame := rect{col: p.Margin, row: p.Top, width: p.Area, height: len(p.Rows)}
	bounds := rect{width: p.Cols, height: p.Height}
	shown := bounds
	if e.In != "" {
		i := -1
		for j, a := range p.Inserts {
			if a.Name == e.In {
				i = j
				break
			}
		}
		if i < 0 {
			return nil
		}
		a := p.Inserts[i]
		frame = rect{col: p.Margin, row: p.Top + a.Row, width: a.Width, height: a.Height}
		bounds = intersect(bounds, frame)
		// The art and the fragment show in the screen area and on the rows of the page.
		shown = intersect(bounds, rect{col: p.Margin, row: p.Top, width: p.Area, height: len(p.Rows)})
	}
	var out []*place
	add := func(r rect) {
		pl := &place{rect: intersect(r, bounds), point: e.At.Kind == screen.AtPoint || e.At.Kind == screen.AtPoints}
		r = intersect(r, shown)
		for row := r.row; row < r.row+r.height; row++ {
			for col := r.col; col < r.col+r.width; col++ {
				if free(grid[row], col) {
					pl.cells = append(pl.cells, row*p.Cols+col)
				}
			}
		}
		out = append(out, pl)
	}
	at := e.At
	switch at.Kind {
	case screen.AtAll: // the art grid or the fragment rows of in, else the grid
		if e.In != "" {
			add(frame)
		} else {
			add(shown)
		}
	case screen.AtRect:
		col, row := coord(at.Points[0].Col, frame.width), coord(at.Points[0].Row, frame.height)
		add(rect{col: frame.col + col, row: frame.row + row, width: at.Width, height: at.Height})
	default:
		for _, pt := range at.Points {
			add(rect{col: frame.col + coord(pt.Col, frame.width), row: frame.row + coord(pt.Row, frame.height), width: 1, height: 1})
		}
	}
	return out
}

// coord gives the coordinate c of a size n from 0: c counts from 1, a negative c from the end.
func coord(c, n int) int {
	if c < 0 {
		return n + c
	}
	return c - 1
}

// intersect gives the cells of both a and b.
func intersect(a, b rect) rect {
	col, row := max(a.col, b.col), max(a.row, b.row)
	end, bottom := min(a.col+a.width, b.col+b.width), min(a.row+a.height, b.row+b.height)
	return rect{col: col, row: row, width: max(0, end-col), height: max(0, bottom-row)}
}

// free reports whether the cell col of row can be in a place: not a slot cell
// and not a cell of a grapheme of width 2 (live-screen, section 6.2).
func free(row screen.Row, col int) bool {
	c := row[col]
	wide := c.Grapheme == "" || col+1 < len(row) && row[col+1].Grapheme == ""
	return c.Slot == 0 && !wide
}
