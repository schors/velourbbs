package screen

import "iter"

// MaxRenderRows is the row limit of a rendered screen (section 10).
const MaxRenderRows = 10000

// Render renders the screen s for the terminal of env and gives its pages (section 6).
// env.Term is in the ranges of section 4.1. env.Items holds the enabled items.
// theme has the styles normal and more. more is the text of the More row.
// The render holds one page: it renders the next rows after the consumer takes a page,
// and it reuses the rows of the page. So the rows of a page are valid until the next page.
func Render(s *Screen, env *Env, theme map[string]Style, more string) iter.Seq[Page] {
	return func(yield func(Page) bool) {
		cols := env.Term.Cols
		w := min(max(cols, s.Width.Min), s.Width.Max)
		r := &renderer{
			env: env, normal: theme["normal"], w: w, cols: cols,
			area: min(w, cols), margin: max(cols-w, 0) / 2, height: env.Term.Rows,
			values: map[Ref][]Grapheme{}, lead: true, yield: yield,
		}
		var moreCells []Cell
		for g, gw := range Graphemes(more) {
			moreCells = appendCell(moreCells, Grapheme{g, gw}, theme["more"])
		}
		r.more = r.row(moreCells)
		r.walk(s.Body, r.normal)
		if !r.stopped {
			yield(r.page(r.rows))
		}
	}
}

// renderer holds the state of one render. It lays out each source line as it comes
// (section 6.3): the buffers of a line have at most W cells.
type renderer struct {
	env    *Env
	normal Style
	w      int // the render width W
	cols   int // term.cols
	area   int // the cells of the screen area: W, or term.cols when it is smaller
	margin int // the left margin of the screen area
	values map[Ref][]Grapheme

	// The pages (section 6.4).
	height  int // term.rows
	more    Row // the More row
	rows    []Row
	out     []Row // the rows of a page with the More row
	free    []Row // the rows of the pages that the consumer took, for reuse
	shown   bool  // a page with a More row is shown
	count   int   // the rendered rows
	yield   func(Page) bool
	stopped bool // the consumer stopped the render
	done    bool // the render stops: the rows are at MaxRenderRows, or the consumer stopped

	// The current source line.
	cells  []Cell     // the cells of the current row
	col    int        // the width of cells
	lead   bool       // the line has only spaces so far: they join the first word
	spaces []lineCell // the break spaces before the word, at most W
	word   []lineCell // the current word, at most W cells
	ww     int        // the width of word
	long   bool       // the word is wider than W: its graphemes go on the rows as they come
}

// lineCell is one grapheme of a source line before the wrap.
type lineCell struct {
	g     Grapheme
	style Style
	brk   bool // a space where the line can break
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
			cur = seg.Apply(cur)
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

// add adds one grapheme to the line (section 6.3).
func (r *renderer) add(c lineCell) {
	switch {
	case r.done:
	case r.lead && c.g.Text == " ":
		// The spaces at the start of the line stay up to W and join the first word.
		if r.ww < r.w {
			r.word = append(r.word, c)
			r.ww++
		}
	case c.brk:
		r.endWord()
		if len(r.spaces) < r.w { // more spaces always make a break, which removes them
			r.spaces = append(r.spaces, c)
		}
	case r.long:
		r.lead = false
		r.place(c)
	default:
		r.lead = false
		r.word = append(r.word, c)
		r.ww += c.g.Width
		if r.ww > r.w {
			// A word wider than W starts on a new row and breaks at a grapheme boundary.
			// The spaces at the break are removed.
			r.long = true
			if r.col > 0 {
				r.newRow()
			}
			for _, c := range r.word {
				r.place(c)
			}
			r.spaces, r.word, r.ww = r.spaces[:0], r.word[:0], 0
		}
	}
}

// endWord puts the word and the spaces before it on the row. A word that does not fit
// starts a new row; the spaces at the break are removed.
func (r *renderer) endWord() {
	if r.long {
		r.long = false
		return
	}
	if len(r.word) == 0 {
		return
	}
	if r.col+len(r.spaces)+r.ww > r.w {
		if r.col > 0 {
			r.newRow()
		}
	} else {
		for _, c := range r.spaces {
			r.place(c)
		}
	}
	for _, c := range r.word {
		r.place(c)
	}
	r.spaces, r.word, r.ww = r.spaces[:0], r.word[:0], 0
}

// endLine ends the source line. A line with a block token and only spaces gives no row
// (section 6.2). The spaces at the end of the line stay up to W.
func (r *renderer) endLine(block bool) {
	if !r.lead || !block {
		r.endWord()
		for _, c := range r.spaces[:min(len(r.spaces), r.w-r.col)] {
			r.place(c)
		}
		r.newRow()
	}
	r.cells, r.col = r.cells[:0], 0
	r.spaces, r.word, r.ww = r.spaces[:0], r.word[:0], 0
	r.lead, r.long = true, false
}

// place puts one grapheme on the row. A wide grapheme that does not fit goes to the next row;
// when W is 1, it becomes a space.
func (r *renderer) place(c lineCell) {
	if r.col+c.g.Width > r.w {
		if r.col > 0 {
			r.newRow()
		}
		if c.g.Width > r.w {
			c.g = space
		}
	}
	r.cells = appendCell(r.cells, c.g, c.style)
	r.col += c.g.Width
}

// newRow ends the current row of the line.
func (r *renderer) newRow() {
	r.addRow(r.cells)
	r.cells, r.col = r.cells[:0], 0
}

// art adds the rows of an art insert with the values of its fields (sections 6.2 and 8.4).
func (r *renderer) art(a ArtInsert) {
	fields := a.Art.Fields
	fi := 0
	for i, cells := range a.Art.Rows {
		if fi < len(fields) && fields[fi].Row == i+1 {
			cells = append([]Cell(nil), cells...) // the model is immutable
			for ; fi < len(fields) && fields[fi].Row == i+1; fi++ {
				f := fields[fi]
				var value []Cell
				for _, g := range FitValue(r.value(a.Fields[fi]), f.Width, f.Right) {
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
	r.out = append(append(r.out[:0], r.rows[:h]...), r.more)
	if !r.yield(r.page(r.out)) {
		r.stopped, r.done = true, true
		return
	}
	r.free = append(r.free, r.rows[:h]...)
	r.rows = r.rows[:copy(r.rows, r.rows[h:])]
}

// page gives a page of rows.
func (r *renderer) page(rows []Row) Page {
	return Page{Rows: rows, Cols: r.cols, Margin: r.margin, Fill: r.normal}
}

// row gives a row of the screen area with cells. It reuses a row of r.free.
// It cuts cells at the screen area; a cut through a wide grapheme leaves a space in its style.
// Each other cell is a space in normal (section 6.2).
func (r *renderer) row(cells []Cell) Row {
	var row Row
	if n := len(r.free); n > 0 {
		row, r.free = r.free[n-1], r.free[:n-1]
	} else {
		row = make(Row, r.area)
	}
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
