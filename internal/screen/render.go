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
		if hasStretch(s.Body) {
			p := countPass(s, env, theme, nil, 0)
			r.use(p, r.height-p.count)
		}
		r.yield = yield
		var moreCells []Cell
		for g, gw := range Graphemes(more) {
			moreCells = appendCell(moreCells, Grapheme{g, gw}, theme["more"])
		}
		r.more = r.row(moreCells)
		r.walk(s.Body, r.normal)
		r.noteWidths()
		if r.stopped {
			return
		}
		if r.shown && r.layout != nil {
			r.layout.add(LayoutNote{Kind: LayoutMore, Pos: r.start(), N: r.count})
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
	return countPass(s, env, theme, nil, 0).count
}

// RenderFile renders the page n of the file screen s (file-view, section 3.2): the rows above
// the file area, the rows of file and the rows below. The page is cut at term.rows.
// It starts at the terminal row 1. A row of file has at most A cells, the area width
// of file-view, section 3.1; a nil row is empty. The margin cells are spaces in normal.
func RenderFile(s *Screen, env *Env, theme map[string]Style, file []Row, n int) Page {
	r := renderFrame(s, env, theme)
	r.noteWidths()
	split := min(r.split, len(r.rows))
	rows := make([]Row, 0, len(r.rows)+len(file))
	rows = append(rows, r.rows[:split]...)
	left := make([]Cell, s.FileScreen.Left)
	for i := range left {
		left[i] = Cell{Grapheme: " ", Style: r.normal}
	}
	widest := 0
	for _, cells := range file {
		cells = append(left[:len(left):len(left)], cells...)
		widest = max(widest, ink(cells))
		rows = append(rows, r.row(cells))
	}
	rows = append(rows, r.rows[split:]...)
	if r.layout != nil {
		if widest > r.area {
			r.layout.add(LayoutNote{Kind: LayoutWidth, Pos: r.filePos, N: widest, M: r.area})
		}
		if n := r.count + len(file); n > r.height {
			r.layout.add(LayoutNote{Kind: LayoutHeight, Pos: r.start(), N: n, M: r.height})
		}
	}
	p := r.page(rows[:min(len(rows), r.height)])
	p.Num = n
	return p
}

// renderFrame renders the rows of the body of the file screen s without pages.
// It keeps the first term.rows rows and counts all rows. A spacer line gives 0 rows (section 6.6).
func renderFrame(s *Screen, env *Env, theme map[string]Style) *renderer {
	r := newRenderer(s, env, theme)
	if hasStretch(s.Body) {
		r.use(countPass(s, env, theme, nil, 0), 0)
	}
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
	return countPass(s, env, theme, list, rows).count
}

// RenderList renders the list page n of the list screen s with the page size p (list-view, section 7).
// env has the list values of page n. The page is cut at term.rows. A page with fewer rows
// has the vertical position of section 6.5.
func RenderList(s *Screen, env *Env, theme map[string]Style, list []ListRow, n, p int) Page {
	r := renderList(s, env, theme, list, p)
	r.noteWidths()
	if r.layout != nil && r.count > r.height {
		r.layout.add(LayoutNote{Kind: LayoutHeight, Pos: r.start(), N: r.count, M: r.height})
	}
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
// It keeps the first term.rows rows and counts all rows. The spacer lines take the free rows
// after the page size (section 6.6).
func renderList(s *Screen, env *Env, theme map[string]Style, list []ListRow, p int) *renderer {
	r := newRenderer(s, env, theme)
	if hasStretch(s.Body) {
		c := countPass(s, env, theme, list, p)
		r.use(c, r.height-c.count)
	}
	r.frame = true
	r.list, r.pageSize = list, p
	r.walk(s.Body, r.normal)
	return r
}

// RenderWidth gives the render width W for cols columns (section 6.1).
func (l Layout) RenderWidth(cols int) int {
	return min(max(cols-2*l.Gutter, l.Width.Min), l.Width.Max)
}

// Margin gives the left margin of the screen area for cols columns (section 6.1).
func (l Layout) Margin(cols int) int {
	f := cols - l.RenderWidth(cols)
	if f <= 0 {
		return 0
	}
	switch g := min(l.Gutter, f); l.Align {
	case AlignLeft:
		return g
	case AlignRight:
		return f - g
	}
	return f / 2
}

// hasStretch reports whether segs have a stretch token or a spacer line (section 6.6).
func hasStretch(segs []Seg) bool {
	for _, seg := range segs {
		switch seg := seg.(type) {
		case Stretch, Spacer:
			return true
		case If:
			if hasStretch(seg.Then) || hasStretch(seg.Else) {
				return true
			}
		case Include:
			if hasStretch(seg.Body) {
				return true
			}
		case ListLine:
			if hasStretch(seg.Row) {
				return true
			}
		}
	}
	return false
}

// plan is the result of the count pass of a screen (section 6.6).
type plan struct {
	count   int    // the rows without the spacer lines, at most MaxRenderRows
	stretch []bool // for each render line: the render reaches a stretch token in it
	spacers int    // the spacer lines that the render reaches
	values  map[Ref][]Grapheme
}

// countPass counts the rows of s without the spacer lines and finds the render lines
// with a stretch token. The list and the page size p are for a list screen.
// A render line ends at each LineEnd that the render reaches: a false block can join source lines.
func countPass(s *Screen, env *Env, theme map[string]Style, list []ListRow, p int) *plan {
	r := newRenderer(s, env, theme)
	r.counting = &plan{values: r.values}
	r.layout = nil
	r.list, r.pageSize = list, p
	r.wrapper = newWrapper(r.w, func([]Cell) { r.pending++ })
	r.walk(s.Body, r.normal)
	r.counting.count = r.count
	return r.counting
}

// use makes the render use the count pass p: the stretch lines of p, and free rows for the spacer lines.
func (r *renderer) use(p *plan, free int) {
	r.plan, r.free, r.values = p, max(free, 0), p.values
	r.startLine()
}

// newRenderer gives a renderer of s for the terminal of env.
func newRenderer(s *Screen, env *Env, theme map[string]Style) *renderer {
	cols := env.Term.Cols
	w := s.RenderWidth(cols)
	margin := s.Margin(cols)
	r := &renderer{
		env: env, theme: theme, normal: theme["normal"], w: w, cols: cols,
		area: min(w, cols), margin: margin, height: env.Term.Rows, fill: s.Fill,
		values: map[Ref][]Grapheme{}, file: s.File, widthMax: s.Width.Max, layout: env.Layout,
	}
	if s.List != nil {
		r.rectRows, r.selectable = s.List.Rows, s.List.Select != nil
	}
	r.wrapper = newWrapper(w, r.textRow)
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
	slots  []Slot // the slot of each Cell.Slot - 1, without its run (live-screen, section 3.1)

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
	// The rows of the list rectangle in art: the key rows. selectable is true in a screen
	// with select (list-view, sections 8 and 9).
	rectRows   int
	selectable bool

	// The count pass (section 6.6). counting is the result in progress, else nil.
	// pending is the rows of the wrap of the current line, and reached is true when the line has a stretch token.
	counting *plan
	pending  int
	reached  bool

	// The render after a count pass: the stretch lines and the free rows of the spacer lines.
	// plan is nil when the screen has no stretch token and no spacer line.
	plan       *plan
	free       int
	spacer     int  // the spacer lines so far
	line       int  // the render line
	stretching bool // the render line has a stretch token: it does not wrap
	sl         stretchLine

	// The layout warnings (screen-render, section 3.6): nil without Env.Layout and in the count pass.
	layout   *LayoutLog
	file     string     // the screen file
	widthMax int        // max of the width limits
	filePos  Pos        // the position of the file area
	lineRows int        // the rows of the wrap of the current source line
	lineCuts []wideRow  // the rows of the current source line that the cut makes shorter
	cuts     []wideNote // the rows that the cut makes shorter, in row order

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
		case Stretch:
			switch {
			case r.counting != nil:
				r.reached = true
			case r.stretching:
				r.sl.stretch(seg.Fill, cur, r.w)
			}
		case Spacer:
			r.spacerLine()
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
			r.split, r.filePos = r.count, seg.Pos
		case ListLine:
			r.listLine(seg, cur)
		case LineEnd:
			r.lineEnd(seg)
		}
	}
	return cur
}

// lineEnd ends the render line. A line with a stretch token gives one row (section 6.6).
func (r *renderer) lineEnd(e LineEnd) {
	switch {
	case r.counting != nil:
		r.endLine(e.Block)
		n := r.pending
		if r.reached {
			n = 1
		}
		r.counting.stretch = append(r.counting.stretch, r.reached)
		r.pending, r.reached = 0, false
		r.countRows(n)
	case r.stretching:
		cells := r.sl.row(r.w)
		r.cut(r.count, r.sl.ink(cells, r.w))
		r.addRow(cells)
		r.sl.reset()
	default:
		r.endLine(e.Block)
	}
	if r.layout != nil {
		if r.lineRows > 1 {
			r.layout.add(LayoutNote{Kind: LayoutWrap, Pos: e.Pos, N: r.lineRows})
		}
		for _, c := range r.lineCuts {
			r.cuts = append(r.cuts, wideNote{e.Pos, c})
		}
		r.lineRows, r.lineCuts = 0, r.lineCuts[:0]
	}
	r.line++
	r.startLine()
}

// startLine starts the render line r.line.
func (r *renderer) startLine() {
	r.stretching = r.plan != nil && r.line < len(r.plan.stretch) && r.plan.stretch[r.line]
}

// spacerLine adds the rows of a spacer line: its share of the free rows (section 6.6).
func (r *renderer) spacerLine() {
	if r.counting != nil {
		r.counting.spacers++
		return
	}
	if r.plan == nil {
		return
	}
	n := share(r.free, r.plan.spacers, r.spacer)
	r.spacer++
	for range n {
		r.addRow(nil)
	}
}

// share gives the share of the item i, from 0, of n items in free units:
// free / n, rounded down, and one more for each of the last free mod n items (section 6.6).
func share(free, n, i int) int {
	k := free / n
	if i >= n-free%n {
		k++
	}
	return k
}

// countRows adds n rows to the count. The count stops at MaxRenderRows.
func (r *renderer) countRows(n int) {
	r.count = min(r.count+n, MaxRenderRows)
	r.done = r.count == MaxRenderRows
}

// stretchLine is a row with stretch tokens before the share of its free cells (section 6.6).
// It keeps at most w + 1 cells: the row cuts the rest. It keeps at most the last 2w tokens:
// with more than w tokens, only the last w tokens can get a cell.
type stretchLine struct {
	cells []Cell
	width int // the cells of the row without its stretch tokens
	full  int // the cells of the row up to its last grapheme that is not a space, without its stretch tokens
	toks  []stretchTok
	n     int // the stretch tokens of the row
}

// stretchTok is a stretch token at the position at of the cells, with its fill cell.
type stretchTok struct {
	at   int
	cell Cell
}

// add adds the grapheme g in style to the row of the width w.
func (l *stretchLine) add(g Grapheme, style Style, w int) { l.addSlot(g, style, 0, w) }

// addSlot adds the grapheme g in style with the mark slot (Cell.Slot) to the row of the width w.
func (l *stretchLine) addSlot(g Grapheme, style Style, slot int32, w int) {
	l.width += g.Width
	if g.Text != " " {
		l.full = l.width
	}
	if len(l.cells) <= w {
		l.cells = appendSlot(l.cells, g, style, slot)
	}
}

// field adds a fixed-width field to the row of the width w: the graphemes gs of the cut field
// and the cells n of the field before the cut, up to its last grapheme that is not a space.
// The field before the cut is wider than gs.
func (l *stretchLine) field(gs []Grapheme, n, cells int, style Style, slot int32, w int) {
	start := l.width
	for _, g := range gs {
		l.addSlot(g, style, slot, w)
	}
	l.width = start + cells
	if n > 0 {
		l.full = start + n
	}
}

// ink gives the cells of the row before the cut, up to its last grapheme that is not a space.
// cells is the row of the width w. When the row is wider than w, the stretch tokens have 0 cells.
func (l *stretchLine) ink(cells []Cell, w int) int {
	if l.width > w {
		return l.full
	}
	return ink(cells)
}

// stretch adds a stretch token with the fill g in style to the row of the width w.
func (l *stretchLine) stretch(g Grapheme, style Style, w int) {
	if len(l.toks) == 2*w {
		l.toks = l.toks[:copy(l.toks, l.toks[w:])]
	}
	l.toks = append(l.toks, stretchTok{len(l.cells), Cell{Grapheme: g.Text, Style: style}})
	l.n++
}

// row gives the cells of the row of the width w: the stretch tokens share the free cells.
// The cells are valid until the next reset.
func (l *stretchLine) row(w int) []Cell {
	free := w - l.width
	if free <= 0 || len(l.toks) == 0 {
		return l.cells
	}
	out := make([]Cell, 0, w)
	prev := 0
	first := l.n - len(l.toks) // the tokens before l.toks have 0 cells
	for i, t := range l.toks {
		out = append(out, l.cells[prev:t.at]...)
		for range share(free, l.n, first+i) {
			out = append(out, t.cell)
		}
		prev = t.at
	}
	return append(out, l.cells[prev:]...)
}

// reset empties the row.
func (l *stretchLine) reset() {
	l.cells, l.width, l.full, l.toks, l.n = l.cells[:0], 0, 0, l.toks[:0], 0
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
	if r.counting != nil {
		r.countRows(r.pageSize)
		return
	}
	first := (r.env.List.Page - 1) * r.pageSize
	var row stretchLine
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
		r.rowText(l.Row, cur, &env, &row)
		cells := row.row(r.w)
		if n := row.ink(cells, r.w); r.layout != nil && n > r.area {
			r.cuts = append(r.cuts, wideNote{l.Pos, wideRow{r.count, n}})
		}
		r.addRow(cells)
		row.reset()
	}
}

// rowText adds the cells of the row text segs with the style cur and the values of env
// to the row l. It stops after W cells. It gives the style after them.
func (r *renderer) rowText(segs []Seg, cur Style, env *Env, l *stretchLine) Style {
	for _, seg := range segs {
		if len(l.cells) > r.w && r.layout == nil { // the layout warnings measure the whole row
			break
		}
		switch seg := seg.(type) {
		case Text:
			for _, g := range seg.Graphemes {
				l.add(g, cur, r.w)
			}
		case Field:
			gs, width, slot := r.fieldText(seg.Ref, env), seg.Width, int32(0)
			if seg.Live {
				gs, width, slot = r.live(seg, env)
			}
			if width > r.w {
				full := FitValue(gs, width, seg.Right)
				l.field(FitValue(gs, r.w, seg.Right), inkGraphemes(full), width, cur, slot, r.w)
				continue
			}
			if width > 0 {
				gs = FitValue(gs, width, seg.Right)
			}
			for _, g := range gs {
				l.addSlot(g, cur, slot, r.w)
			}
		case SetStyle:
			cur = seg.Apply(cur, r.theme)
		case Stretch:
			l.stretch(seg.Fill, cur, r.w)
		case If:
			if seg.Cond.Eval(env) {
				cur = r.rowText(seg.Then, cur, env, l)
			} else {
				cur = r.rowText(seg.Else, cur, env, l)
			}
		}
	}
	return cur
}

// text adds graphemes to the line. A space is a break.
func (r *renderer) text(gs []Grapheme, style Style) {
	for _, g := range gs {
		if r.stretching {
			r.sl.add(g, style, r.w)
		} else {
			r.add(lineCell{g: g, style: style, brk: g.Text == " "})
		}
	}
}

// field adds the value of a field token to the line (sections 4.3 and 6.3).
// A fixed-width field is one word of at most W cells.
// A live field is a slot: one word of its slot width (live-screen, section 3.1).
func (r *renderer) field(f Field, style Style) {
	gs, width, slot := r.value(f.Ref), f.Width, int32(0)
	if f.Live {
		gs, width, slot = r.live(f, r.env)
	} else if width == 0 {
		r.text(gs, style)
		return
	}
	cut := FitValue(gs, min(width, r.w), f.Right)
	if width > r.w { // the field is one word of W cells: its row is the field (section 6.3)
		n := inkGraphemes(FitValue(gs, width, f.Right))
		if r.stretching {
			r.sl.field(cut, n, width, style, slot, r.w)
			return
		}
		// The field starts a new row after other cells (section 6.3).
		row := r.count
		if r.col > 0 {
			row++
		}
		r.cut(row, n)
	}
	for _, g := range cut {
		if r.stretching {
			r.sl.addSlot(g, style, slot, r.w)
		} else {
			r.add(lineCell{g: g, style: style, slot: slot})
		}
	}
}

// live gives the graphemes of the value of the live field f in env, the width of its slot
// and a new mark of the slot (live-screen, section 3.1): the width of the token, else the largest width.
// bbs.time with a layout shows Session.Clock in the layout; a zero instant shows Session.Time.
func (r *renderer) live(f Field, env *Env) ([]Grapheme, int, int32) {
	var gs []Grapheme
	if f.Layout != "" && !env.Session.Clock.IsZero() {
		gs = CleanValue(env.Session.Clock.Format(f.Layout))
	} else {
		gs = r.fieldText(f.Ref, env)
	}
	width := f.Width
	if width == 0 {
		width = env.slotWidth(f.Ref, f.Layout)
	}
	r.slots = append(r.slots, Slot{Width: min(width, r.w), Ref: f.Ref, Layout: f.Layout, Right: f.Right})
	return gs, width, int32(len(r.slots))
}

// art adds the rows of an art insert with the values of its fields (section 6.2 and art-format, section 4).
// In a list render, the list rectangle shows the list page (list-view, section 9).
func (r *renderer) art(a ArtInsert) {
	if r.counting != nil {
		r.countRows(len(a.Art.Rows))
		return
	}
	if r.layout != nil && a.Art.Width > r.area && r.area < r.widthMax {
		r.layout.add(LayoutNote{Kind: LayoutArt, Pos: a.Pos, N: a.Art.Width, M: r.area, Art: a.Name})
	}
	ls := a.Art.List
	if r.env.List == nil {
		ls = nil
	}
	for i, cells := range a.Art.Rows {
		cells = r.artFields(a.Art, i+1, cells, r.env)
		if ls != nil && i+1 >= ls.Row && i+1 < ls.Row+r.rectRows {
			cells = r.listRect(a.Art, i+1-ls.Row, cells)
		}
		r.addRow(cells)
	}
}

// artFields gives the cells of the art row n with the values of its fields in env, then with its boxes
// (art-format, sections 4 and 5). Without a row in env, the row.* fields keep their cells.
func (r *renderer) artFields(a *Art, n int, cells Row, env *Env) Row {
	copied := false
	own := func() {
		if !copied {
			cells = append(Row(nil), cells...) // the model is immutable
			copied = true
		}
	}
	for _, f := range a.Fields {
		if f.Row != n || f.Ref.Row() != "" && env.Row == nil {
			continue
		}
		own()
		gs, slot := r.fieldText(f.Ref, env), int32(0)
		if f.Live { // the slot has the width of the token (live-screen, section 3.1)
			gs, _, slot = r.live(Field{Ref: f.Ref, Width: f.Width, Right: f.Right, Live: true, Layout: f.Layout}, env)
		}
		var value []Cell
		for _, g := range FitValue(gs, f.Width, f.Right) {
			value = appendSlot(value, g, f.Style, slot)
		}
		copy(cells[f.Col-1:], value)
	}
	for _, b := range a.Boxes {
		if b.Row != n {
			continue
		}
		own()
		box := cells[b.Col-1 : b.End]
		shown := 0
		if b.Cond.Eval(env) { // the box content moves to the first cell of the box
			shown = copy(box, cells[b.Content-1:b.Content-1+b.Len])
		}
		for i := shown; i < len(box); i++ {
			box[i] = Cell{Grapheme: " ", Style: b.Style}
		}
	}
	return cells
}

// listRect gives the cells of the art row i of the list rectangle, from 0, with the row i of the list page.
// A rectangle row without a row of the list page is spaces in the styles of the list row cells.
// The selected row takes the style select (list-view, section 9).
func (r *renderer) listRect(a *Art, i int, cells Row) Row {
	cells = append(Row(nil), cells...) // the model is immutable
	ls := a.List
	from, to := ls.First-1, ls.Last
	listRow := a.Rows[ls.Row-1]
	n := (r.env.List.Page-1)*r.pageSize + i
	if i >= r.pageSize || n >= len(r.list) {
		for c := from; c < to; c++ {
			cells[c] = Cell{Grapheme: " ", Style: listRow[c].Style}
		}
		return cells
	}
	env := *r.env
	env.Row = &RowValues{Num: n + 1, Selected: n == r.env.List.Cursor, Row: r.list[n]}
	copy(cells[from:to], r.artFields(a, ls.Row, listRow, &env)[from:to])
	if r.selectable && env.Row.Selected {
		for c := from; c < to; c++ {
			cells[c].Style = r.theme["select"]
		}
	}
	return cells
}

// fieldText gives the graphemes of the field ref in env. The value of a row.* field
// is not in the cache of the render: it differs for each row.
func (r *renderer) fieldText(ref Ref, env *Env) []Grapheme {
	if ref.Row() != "" {
		return env.Text(ref)
	}
	return r.value(ref)
}

// addRow adds a row from the cells of the screen area. The render stops at MaxRenderRows.
// The first page shows up to term.rows rows. When more rows follow, each page has
// term.rows - 1 rows and the More row, and the last page has the rest (section 6.4).
func (r *renderer) addRow(cells []Cell) {
	if r.done {
		return
	}
	if r.counting != nil {
		r.countRows(1)
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

// textRow adds a row of the wrap of a text line. It counts the row for the layout warnings.
func (r *renderer) textRow(cells []Cell) {
	if r.layout != nil {
		r.lineRows++
		r.cut(r.count, ink(cells))
	}
	r.addRow(cells)
}

// wideRow is the row number row, from 0, of a render, and its cells n before the cut,
// up to its last grapheme that is not a space.
type wideRow struct {
	row, n int
}

// wideNote is a wideRow at the position pos.
type wideNote struct {
	pos Pos
	wideRow
}

// cut adds the row number row with n cells to the current source line when the cut makes it shorter.
func (r *renderer) cut(row, n int) {
	if r.layout != nil && n > r.area {
		r.lineCuts = append(r.lineCuts, wideRow{row, n})
	}
}

// noteWidths adds the warnings Width of the render: for each position, the widest row
// of the first page where the cut makes a row shorter (screen-render, section 3.6).
// A render with a More row has pages of term.rows - 1 rows; each other render is one page.
func (r *renderer) noteWidths() {
	type key struct {
		pos  Pos
		page int
	}
	var order []key
	widest := map[key]int{}
	for _, c := range r.cuts {
		k := key{c.pos, 0}
		if r.shown {
			k.page = c.row / (r.height - 1)
		}
		if _, ok := widest[k]; !ok {
			order = append(order, k)
		}
		widest[k] = max(widest[k], c.n)
	}
	for _, k := range order { // the rows are in order, so the first page of a position comes first
		r.layout.add(LayoutNote{Kind: LayoutWidth, Pos: k.pos, N: widest[k], M: r.area})
	}
}

// start gives the position of the first line of the screen file.
func (r *renderer) start() Pos {
	return Pos{File: r.file, Line: 1, Col: 1}
}

// page gives the next page with rows and their slots.
func (r *renderer) page(rows []Row) Page {
	r.pages++
	return Page{
		Num: r.pages, Rows: rows, Cols: r.cols, Height: r.height, Margin: r.margin, Normal: r.normal, Fill: r.fill,
		Slots: r.pageSlots(rows),
	}
}

// pageSlots gives the slots of rows: one for each run of cells with one mark (live-screen, section 3.1).
func (r *renderer) pageSlots(rows []Row) []Slot {
	if len(r.slots) == 0 {
		return nil
	}
	var out []Slot
	seen := map[int32]int{} // the cells of each slot in the runs before
	for i, row := range rows {
		last := int32(0)
		for j, c := range row {
			switch c.Slot {
			case 0:
			case last:
				out[len(out)-1].Cells++
			default:
				s := r.slots[c.Slot-1]
				s.Row, s.Col, s.Skip, s.Cells = i, j, seen[c.Slot], 1
				out = append(out, s)
			}
			if c.Slot != 0 {
				seen[c.Slot]++
			}
			last = c.Slot
		}
	}
	return out
}

// row gives a row of the screen area with cells. It cuts cells at the screen area; a cut through a wide grapheme leaves a space in its style.
// Each other cell is a space in normal (section 6.2).
func (r *renderer) row(cells []Cell) Row {
	row := make(Row, r.area)
	for i := range row {
		row[i] = Cell{Grapheme: " ", Style: r.normal}
	}
	n := copy(row, cells)
	if n < len(cells) && n > 0 && cells[n].Grapheme == "" {
		row[n-1].Grapheme = " "
	}
	return row
}

// appendCell appends the cells of g: a wide grapheme takes 2 cells.
func appendCell(cells []Cell, g Grapheme, style Style) []Cell { return appendSlot(cells, g, style, 0) }

// appendSlot appends the cells of g with the mark slot (Cell.Slot).
func appendSlot(cells []Cell, g Grapheme, style Style, slot int32) []Cell {
	cells = append(cells, Cell{Grapheme: g.Text, Style: style, Slot: slot})
	if g.Width == 2 {
		cells = append(cells, Cell{Style: style, Slot: slot})
	}
	return cells
}
