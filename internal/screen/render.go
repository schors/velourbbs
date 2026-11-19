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
		r := newRenderer(s, env, theme)
		r.yield = yield
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

// FrameRows gives the count of the rows above and below the file area of the file screen s
// (file-view, section 3.2). The count stops at MaxRenderRows.
func FrameRows(s *Screen, env *Env, theme map[string]Style) int {
	return renderFrame(s, env, theme).count
}

// RenderFile renders the page n of the file screen s (file-view, section 3.2): the rows above
// the file area, the rows of file and the rows below. The page is cut at term.rows.
// It starts at the terminal row 1. A row of file has at most A cells, the area width
// of file-view, section 3.1; a nil row is empty. The margin cells are spaces in normal.
func RenderFile(s *Screen, env *Env, theme map[string]Style, file []Row, n int) Page {
	r := renderFrame(s, env, theme)
	split := min(r.split, len(r.rows))
	rows := make([]Row, 0, len(r.rows)+len(file))
	rows = append(rows, r.rows[:split]...)
	left := make([]Cell, s.FileScreen.Left)
	for i := range left {
		left[i] = Cell{" ", r.normal}
	}
	for _, cells := range file {
		rows = append(rows, r.row(append(left[:len(left):len(left)], cells...)))
	}
	rows = append(rows, r.rows[split:]...)
	p := r.page(rows[:min(len(rows), r.height)])
	p.Num = n
	return p
}

// renderFrame renders the rows of the body of the file screen s without pages.
// It keeps the first term.rows rows and counts all rows.
func renderFrame(s *Screen, env *Env, theme map[string]Style) *renderer {
	r := newRenderer(s, env, theme)
	r.frame = true
	// A render that stops at MaxRenderRows before the file area puts the file rows after the kept rows.
	r.split = MaxRenderRows
	r.walk(s.Body, r.normal)
	return r
}

// ListRows gives the rows R of page 1 of the list screen s with the page size rows
// (list-view, section 7). env has the list values of page 1 with this page size.
// The count stops at MaxRenderRows.
func ListRows(s *Screen, env *Env, theme map[string]Style, list []ListRow, rows int) int {
	return renderList(s, env, theme, list, rows).count
}

// RenderList renders the list page n of the list screen s with the page size p (list-view, section 7).
// env has the list values of page n. The page is cut at term.rows. A page with fewer rows
// has the vertical position of section 6.5.
func RenderList(s *Screen, env *Env, theme map[string]Style, list []ListRow, n, p int) Page {
	r := renderList(s, env, theme, list, p)
	pg := r.page(r.rows)
	pg.Num = n
	switch s.VAlign {
	case VAlignMiddle:
		pg.Top = (r.height - len(r.rows)) / 2
	case VAlignBottom:
		pg.Top = r.height - len(r.rows)
	}
	return pg
}

// renderList renders the rows of the list screen s with the page size p, without pages.
// It keeps the first term.rows rows and counts all rows.
func renderList(s *Screen, env *Env, theme map[string]Style, list []ListRow, p int) *renderer {
	r := newRenderer(s, env, theme)
	r.frame = true
	r.list, r.pageSize = list, p
	r.walk(s.Body, r.normal)
	return r
}

// RenderWidth gives the render width W for cols columns (section 6.1).
func (l Layout) RenderWidth(cols int) int {
	return min(max(cols, l.Width.Min), l.Width.Max)
}

// newRenderer gives a renderer of s for the terminal of env.
func newRenderer(s *Screen, env *Env, theme map[string]Style) *renderer {
	cols := env.Term.Cols
	w := s.RenderWidth(cols)
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
		values: map[Ref][]Grapheme{},
	}
	r.wrapper = newWrapper(w, r.addRow)
	return r
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
	pages   int  // the pages so far

	// The frame render of a file screen or a list screen: no pages. rows keeps the first term.rows rows.
	frame bool
	split int // the rows before the file area

	// The list of a list screen and the page size P (list-view, section 7).
	list     []ListRow
	pageSize int

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
		case FileArea:
			r.split = r.count
		case ListLine:
			r.listLine(seg, cur)
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

// listLine adds the P rows of a list line: one row for each row of the list page,
// then empty rows up to P (list-view, section 5). A row does not wrap: row cuts it.
// Each row starts with the style cur; the list line does not change the style after it.
func (r *renderer) listLine(l ListLine, cur Style) {
	first := (r.env.List.Page - 1) * r.pageSize
	for i := range r.pageSize {
		if r.done {
			return
		}
		n := first + i
		if n >= len(r.list) {
			r.addRow(nil)
			continue
		}
		env := *r.env
		env.Row = &RowValues{Num: n + 1, Selected: n == r.env.List.Cursor, Row: r.list[n]}
		cells, _ := r.rowText(l.Row, cur, &env, nil)
		r.addRow(cells)
	}
}

// rowText appends the cells of the row text segs with the style cur and the values of env
// to cells. It stops after W cells. It gives the cells and the style after them.
func (r *renderer) rowText(segs []Seg, cur Style, env *Env, cells []Cell) ([]Cell, Style) {
	for _, seg := range segs {
		if len(cells) > r.w {
			break
		}
		switch seg := seg.(type) {
		case Text:
			for _, g := range seg.Graphemes {
				cells = appendCell(cells, g, cur)
			}
		case Field:
			var gs []Grapheme
			if seg.Ref.Row() != "" { // the value of each row: not in the cache of the render
				gs = env.Text(seg.Ref)
			} else {
				gs = r.value(seg.Ref)
			}
			if seg.Width > 0 {
				gs = FitValue(gs, min(seg.Width, r.w), seg.Right)
			}
			for _, g := range gs {
				cells = appendCell(cells, g, cur)
			}
		case SetStyle:
			cur = seg.Apply(cur, r.theme)
		case If:
			if seg.Cond.Eval(env) {
				cells, cur = r.rowText(seg.Then, cur, env, cells)
			} else {
				cells, cur = r.rowText(seg.Else, cur, env, cells)
			}
		}
	}
	return cells, cur
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
	if r.frame {
		if len(r.rows) < r.height {
			r.rows = append(r.rows, r.row(cells))
		}
		r.count++
		r.done = r.count == MaxRenderRows
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

// page gives the next page with rows.
func (r *renderer) page(rows []Row) Page {
	r.pages++
	return Page{Num: r.pages, Rows: rows, Cols: r.cols, Height: r.height, Margin: r.margin, Normal: r.normal, Fill: r.fill}
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
