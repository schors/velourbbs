package screen

import "iter"

// MaxRenderRows is the row limit of a rendered screen (section 9).
const MaxRenderRows = 10000

// Render renders the screen s for the terminal of env and gives its pages (section 6).
// env.Term is in the ranges of section 4.1. env.Items holds the enabled items.
// theme is the session theme: it has normal, more and each name of the style tokens of s.
// more is the text of the More row.
// The render holds one page: it renders the next rows after the consumer takes a page.
func Render(s *Screen, env *Env, theme map[string]Style, more string) iter.Seq[Page] {
	return func(yield func(Page) bool) {
		cols := env.Term.Cols
		w := min(max(cols, s.Width.Min), s.Width.Max)
		margin := max(cols-w, 0) // align = "right" (section 6.1)
		switch s.Align {
		case AlignCenter:
			margin /= 2
		case AlignLeft:
			margin = 0
		}
		r := &renderer{
			env: env, theme: theme, normal: theme["normal"], w: w, cols: cols,
			area: min(w, cols), margin: margin, height: env.Term.Rows, fill: s.Fill,
			values: map[Ref][]Grapheme{}, yield: yield,
		}
		r.wrapper = newWrapper(w, r.addRow)
		var moreCells []Cell
		for g, gw := range Graphemes(more) {
			moreCells = appendCell(moreCells, Grapheme{g, gw}, theme["more"])
		}
		r.more = r.row(moreCells)
		r.walk(s.Body, r.normal)
		if r.stopped {
			return
		}
		p := r.page(r.rows)
		if !r.shown { // the only page (section 6.5)
			switch s.VAlign {
			case VAlignMiddle:
				p.Top = (r.height - len(r.rows)) / 2
			case VAlignBottom:
				p.Top = r.height - len(r.rows)
			}
		}
		yield(p)
	}
}

// renderer holds the state of one render. Its wrapper lays out each source line as it comes
// (section 6.3).
type renderer struct {
	env    *Env
	theme  map[string]Style // the session theme
	normal Style
	w      int // the render width W
	cols   int // term.cols
	area   int // the cells of the screen area: W, or term.cols when it is smaller
	margin int // the left margin of the screen area
	fill   Fill
	values map[Ref][]Grapheme

	// The pages (section 6.4).
	height  int // term.rows
	more    Row // the More row
	rows    []Row
	shown   bool // a page with a More row is shown
	count   int  // the rendered rows
	yield   func(Page) bool
	stopped bool // the consumer stopped the render
	done    bool // the render stops: the rows are at MaxRenderRows, or the consumer stopped

	wrapper // the current source line
}

// walk renders segs with the style cur and gives the style after them.
func (r *renderer) walk(segs []Seg, cur Style) Style {
	for _, seg := range segs {
		if r.done {
			break
		}
		switch seg := seg.(type) {
		case Text:
			r.text(seg.Graphemes, cur)
		case Field:
			r.field(seg, cur)
		case SetStyle:
			cur = seg.Apply(cur, r.theme)
		case If:
			if seg.Cond.Eval(r.env) {
				cur = r.walk(seg.Then, cur)
			} else {
				cur = r.walk(seg.Else, cur)
			}
		case Include:
			r.walk(seg.Body, cur) // the style before the include stays (section 3.4)
		case ArtInsert:
			r.art(seg) // the art does not change the style of the text (section 7.3)
		case LineEnd:
			r.endLine(seg.Block)
		}
	}
	return cur
}

// value gives the graphemes of the field r, once for each render.
func (r *renderer) value(ref Ref) []Grapheme {
	gs, ok := r.values[ref]
	if !ok {
		gs = r.env.Text(ref)
		r.values[ref] = gs
	}
	return gs
}

// text adds graphemes to the line. A space is a break.
func (r *renderer) text(gs []Grapheme, style Style) {
	for _, g := range gs {
		r.add(lineCell{g, style, g.Text == " "})
	}
}

// field adds the value of a field token to the line (sections 4.3 and 6.3).
// A fixed-width field is one word of at most W cells.
func (r *renderer) field(f Field, style Style) {
	gs := r.value(f.Ref)
	if f.Width == 0 {
		r.text(gs, style)
		return
	}
	for _, g := range FitValue(gs, min(f.Width, r.w), f.Right) {
		r.add(lineCell{g, style, false})
	}
}

// art adds the rows of an art insert with the values of its fields (section 6.2 and art-format, section 4).
func (r *renderer) art(a ArtInsert) {
	fields := a.Art.Fields
	fi := 0
	for i, cells := range a.Art.Rows {
		if fi < len(fields) && fields[fi].Row == i+1 {
			cells = append([]Cell(nil), cells...) // the model is immutable
			for ; fi < len(fields) && fields[fi].Row == i+1; fi++ {
				f := fields[fi]
				var value []Cell
				for _, g := range FitValue(r.value(f.Ref), f.Width, f.Right) {
					value = appendCell(value, g, f.Style)
				}
				copy(cells[f.Col-1:], value)
			}
		}
		r.addRow(cells)
	}
}

// addRow adds a row from the cells of the screen area. The render stops at MaxRenderRows.
// The first page shows up to term.rows rows. When more rows follow, each page has
// term.rows - 1 rows and the More row, and the last page has the rest (section 6.4).
func (r *renderer) addRow(cells []Cell) {
	if r.done {
		return
	}
	r.rows = append(r.rows, r.row(cells))
	r.count++
	r.done = r.count == MaxRenderRows
	h := r.height - 1
	if len(r.rows) <= h || !r.shown && len(r.rows) <= r.height {
		return
	}
	r.shown = true
	out := append(r.rows[:h:h], r.more) // a new slice: r.rows keeps the rest
	if !r.yield(r.page(out)) {
		r.stopped, r.done = true, true
		return
	}
	r.rows = r.rows[:copy(r.rows, r.rows[h:])]
}

// page gives a page of rows.
func (r *renderer) page(rows []Row) Page {
	return Page{Rows: rows, Cols: r.cols, Height: r.height, Margin: r.margin, Normal: r.normal, Fill: r.fill}
}

// row gives a row of the screen area with cells. It cuts cells at the screen area; a cut through a wide grapheme leaves a space in its style.
// Each other cell is a space in normal (section 6.2).
func (r *renderer) row(cells []Cell) Row {
	row := make(Row, r.area)
	for i := range row {
		row[i] = Cell{" ", r.normal}
	}
	n := copy(row, cells)
	if n < len(cells) && n > 0 && cells[n].Grapheme == "" {
		row[n-1].Grapheme = " "
	}
	return row
}

// appendCell appends the cells of g: a wide grapheme takes 2 cells.
func appendCell(cells []Cell, g Grapheme, style Style) []Cell {
	cells = append(cells, Cell{g.Text, style})
	if g.Width == 2 {
		cells = append(cells, Cell{"", style})
	}
	return cells
}
