package screen

import (
	"iter"
	"strconv"
)

// TabWidth is the step of a TAB in a line of a file (file-view, sections 5.4 and 6).
const TabWidth = 8

// TextLayout gives the layout of the plain text or source file c in rows of w cells
// (file-view, section 6). first and last are the 1-based lines of the key lines;
// 0 and 0 show all lines. numbers is the key numbers. theme has normal, md_code and md_dim.
// The restart point of a row is the index of its line in the shown lines.
func TextLayout(c *Content, first, last int, numbers bool, w int, theme map[string]Style) FileLayout {
	lines, start, top := c.Lines, 1, len(c.Lines)
	if first > 0 {
		lines, start, top = lines[first-1:last], first, last
	}
	t := textLayout{lines: lines, start: start, source: c.Kind == Source,
		normal: theme["normal"], code: theme["md_code"], dim: theme["md_dim"]}
	if numbers {
		// The number, then one space. The column is at most W / 2, as a prefix (section 5.2).
		if nw := len(strconv.Itoa(top)) + 1; nw <= w/2 {
			t.nw = nw
		}
	}
	t.w = w - t.nw
	return t.rows
}

// textLayout is the layout of a plain text or source file.
type textLayout struct {
	lines             []string // the shown lines
	start             int      // the number of the first shown line
	source            bool
	w                 int // the cells of a row after the number column
	nw                int // the cells of the number column, or 0 without numbers
	normal, code, dim Style
}

// rows gives the rows from the line with the index from.
func (t textLayout) rows(from int) iter.Seq[FileRow] {
	return func(yield func(FileRow) bool) {
		var (
			i       int    // the index of the line
			lineRow bool   // the next row is the first row of the line i
			row     []Cell // the number column and the cells of a row
			lcs     []lineCell
			stopped bool
		)
		emit := func(cells []Cell) {
			if stopped {
				return
			}
			row = t.number(row[:0], i, lineRow)
			row = append(row, cells...)
			lineRow = false
			stopped = !yield(FileRow{Cells: row, From: i})
		}
		wr := newWrapper(t.w, emit)
		var cells []Cell
		for i = from; i < len(t.lines) && !stopped; i++ {
			lineRow = true
			if t.source {
				lcs = expandTabs(lcs[:0], t.lines[i], t.code)
				cells = cutRow(cells[:0], lcs, t.w, t.dim)
				emit(cells)
				continue
			}
			lcs = expandTabs(lcs[:0], t.lines[i], t.normal)
			for _, c := range lcs {
				wr.add(c)
			}
			wr.endLine(false)
		}
	}
}

// number appends the number column of a row of the line with the index i:
// on the first row of the line the number right-aligned and a space, else spaces.
func (t textLayout) number(cells []Cell, i int, first bool) []Cell {
	if t.nw == 0 {
		return cells
	}
	n := ""
	if first {
		n = strconv.Itoa(t.start + i)
	}
	for range t.nw - 1 - len(n) {
		cells = append(cells, Cell{Grapheme: " ", Style: t.dim})
	}
	for _, d := range n {
		cells = append(cells, Cell{Grapheme: string(d), Style: t.dim})
	}
	return append(cells, Cell{Grapheme: " ", Style: t.dim})
}

// expandTabs appends the graphemes of line in style to lcs. A TAB becomes spaces up to the next
// multiple of TabWidth cells from the start of the line (file-view, sections 5.4 and 6).
// A space is a break space.
func expandTabs(lcs []lineCell, line string, style Style) []lineCell {
	col := 0
	for g, w := range Graphemes(line) {
		if g == "\t" {
			w = TabWidth - col%TabWidth
			for range w {
				lcs = append(lcs, lineCell{g: space, style: style, brk: true})
			}
		} else {
			lcs = append(lcs, lineCell{g: Grapheme{g, w}, style: style, brk: g == " "})
		}
		col += w
	}
	return lcs
}

// cutRow appends the row of lcs in width cells to cells: a code row or a table row
// (file-view, sections 5.3 and 5.4). A row wider than width keeps the graphemes in its first
// width - 3 cells, pads the rest with spaces and ends with the cut mark "..." in dim.
// When width is less than 3, the cut mark is cut to width.
func cutRow(cells []Cell, lcs []lineCell, width int, dim Style) []Cell {
	total := 0
	for _, c := range lcs {
		total += c.g.Width
	}
	keep := width
	if total > width {
		keep = max(width-3, 0)
	}
	used, i := 0, 0
	for ; i < len(lcs) && used+lcs[i].g.Width <= keep; i++ {
		cells = appendCell(cells, lcs[i].g, lcs[i].style)
		used += lcs[i].g.Width
	}
	if total <= width {
		return cells
	}
	// A cut wide grapheme leaves a space in its style.
	for ; used < keep; used++ {
		cells = append(cells, Cell{Grapheme: " ", Style: lcs[i].style})
	}
	for range width - keep {
		cells = append(cells, Cell{Grapheme: ".", Style: dim})
	}
	return cells
}
