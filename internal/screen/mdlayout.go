package screen

import (
	"iter"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/yuin/goldmark/v2/ast"
	east "github.com/yuin/goldmark/v2/extension/ast"
)

// MarkdownLayout gives the layout of the Markdown file c in rows of w cells
// (file-view, section 5). theme has normal and the md_* styles.
// The restart point of a row is the index of its top-level block.
// The index after the last block is the restart point of the notes.
func MarkdownLayout(c *Content, w int, theme map[string]Style) FileLayout {
	m := &mdLayout{src: c.Body, w: w, theme: theme, noteOf: map[string]int{}}
	for b := c.Doc.FirstChild(); b != nil; b = b.NextSibling() {
		m.blocks = append(m.blocks, b)
	}
	// A note number is the number of the address in the order of the note numbers in the text.
	_ = ast.Walk(c.Doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			if addr, ok := linkAddress(n, m.src); ok && hasScheme(addr) && m.noteOf[addr] == 0 {
				m.notes = append(m.notes, addr)
				m.noteOf[addr] = len(m.notes)
			}
		}
		return ast.WalkContinue, nil
	})
	m.first = len(m.blocks)
	for row := range m.rows(0) {
		m.first = row.From
		break
	}
	return m.rows
}

// mdLayout is the layout of a Markdown file.
type mdLayout struct {
	src    []byte
	w      int
	theme  map[string]Style
	blocks []ast.Node     // the top-level blocks
	notes  []string       // the addresses of the notes, in the order of their numbers
	noteOf map[string]int // the note number of an address
	first  int            // the index of the first top-level block that gives rows
}

// linkAddress gives the address of a link or an image.
func linkAddress(n ast.Node, src []byte) (string, bool) {
	switch n := n.(type) {
	case *ast.Link:
		return n.Destination.Value(src), true
	case *ast.Image:
		return n.Destination.Value(src), true
	}
	return "", false
}

// hasScheme reports whether the address starts with a scheme of CommonMark:
// 2 to 32 characters, a letter first, then letters, digits, "+", "." or "-", then ":".
func hasScheme(addr string) bool {
	end := strings.IndexByte(addr, ':')
	if end < 2 || end > 32 {
		return false
	}
	for i := range end {
		c := addr[i]
		letter := 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z'
		other := '0' <= c && c <= '9' || c == '+' || c == '.' || c == '-'
		if !letter && (i == 0 || !other) {
			return false
		}
	}
	return true
}

// rows gives the rows from the top-level block with the index from.
func (m *mdLayout) rows(from int) iter.Seq[FileRow] {
	return func(yield func(FileRow) bool) {
		r := mdRun{mdLayout: m, yield: yield}
		// The separator before a block is a row of that block.
		r.sep = from > m.first
		for i := from; i < len(m.blocks) && !r.stopped; i++ {
			r.from = i
			r.blockSep(m.blocks[i], "normal", 0, false)
		}
		if !r.stopped && len(m.notes) > 0 {
			r.from = len(m.blocks)
			r.notesRows()
		}
	}
}

// prefixPart is the cells before the rows of a block: a list marker or a bar.
// The first row of the part has first; the other rows have rest.
type prefixPart struct {
	first, rest []Cell
	used        bool
}

// mdRun is one run of the layout from a restart point.
type mdRun struct {
	*mdLayout
	yield   func(FileRow) bool
	stopped bool
	from    int // the restart point of the current top-level block
	count   int // the rows that are not separators

	parts    []prefixPart
	sep      bool // a separator row comes before the next row
	sepDepth int  // the parts of the separator row

	pre, row []Cell
}

// style gives the theme style name.
func (r *mdRun) style(name string) Style {
	return r.theme[name]
}

// prefixWidth gives the width of the prefix: at most W / 2 (file-view, section 5.2).
func (r *mdRun) prefixWidth() int {
	n := 0
	for _, p := range r.parts {
		n += len(p.rest)
	}
	return min(n, r.w/2)
}

// space gives the cells of a row after the prefix.
func (r *mdRun) space() int {
	return r.w - r.prefixWidth()
}

// prefix puts the prefix of a row of the first depth parts in r.pre.
// A row that is not a separator uses the first cells of a part once.
// A longer prefix than W / 2 keeps its last W / 2 cells.
func (r *mdRun) prefix(depth int, sep bool) []Cell {
	r.pre = r.pre[:0]
	for i := range r.parts[:depth] {
		p := &r.parts[i]
		if sep || p.used {
			r.pre = append(r.pre, p.rest...)
		} else {
			r.pre = append(r.pre, p.first...)
			p.used = true
		}
	}
	return r.pre[max(0, len(r.pre)-r.w/2):]
}

// emit gives one row: the separator before it, if one is pending, then the prefix and cells.
func (r *mdRun) emit(cells []Cell, row FileRow) {
	if r.stopped {
		return
	}
	if r.sep {
		r.sep = false
		r.row = append(r.row[:0], r.prefix(r.sepDepth, true)...)
		if !r.yield(FileRow{Cells: r.row, From: r.from, Sep: true}) {
			r.stopped = true
			return
		}
	}
	r.row = append(r.row[:0], r.prefix(len(r.parts), false)...)
	r.row = append(r.row, cells...)
	row.Cells, row.From = r.row, r.from
	r.count++
	if !r.yield(row) {
		r.stopped = true
	}
}

// push adds a prefix part of the cells first and rest.
func (r *mdRun) push(first, rest []Cell) {
	r.parts = append(r.parts, prefixPart{first: first, rest: rest})
}

// pop removes the last prefix part.
func (r *mdRun) pop() {
	r.parts = r.parts[:len(r.parts)-1]
}

// blockSep lays out the block n and, when it gave rows, sets the separator before
// the next block of its parent. tight is true in a tight list (file-view, section 5.2).
func (r *mdRun) blockSep(n ast.Node, base string, level int, tight bool) {
	n0 := r.count
	r.block(n, base, level)
	if r.count > n0 {
		r.sep, r.sepDepth = !tight, len(r.parts)
	}
}

// children lays out the child blocks of n.
func (r *mdRun) children(n ast.Node, base string, level int, tight bool) {
	for c := n.FirstChild(); c != nil && !r.stopped; c = c.NextSibling() {
		r.blockSep(c, base, level, tight)
	}
}

// block lays out the block n (file-view, section 5.2). base is the theme style of the text,
// level is the number of the lists around n.
func (r *mdRun) block(n ast.Node, base string, level int) {
	switch n := n.(type) {
	case *ast.Paragraph:
		r.text(n, base, 0)
	case *ast.Heading:
		r.text(n, headingStyle(n.Level), n.Level)
	case *ast.List:
		r.list(n, base, level+1)
	case *ast.Blockquote:
		r.push(r.bar("md_mark"), r.bar("md_mark"))
		r.children(n, "md_em", level, false)
		r.pop()
	case *ast.CodeBlock:
		r.code(n)
	case *ast.ThematicBreak:
		r.thematicBreak()
	case *east.Table:
		r.table(n, base)
	case *ast.HTMLBlock:
		r.html(n)
	case *ast.LinkReferenceDefinition:
		// No rows.
	default:
		r.children(n, base, level, false)
	}
}

// headingStyle gives the theme style of a heading of the level (file-view, section 4).
func headingStyle(level int) string {
	switch level {
	case 1:
		return "md_h1"
	case 2:
		return "md_h2"
	}
	return "md_h3"
}

// bar gives the cells of a bar in the style: "│" and a space (file-view, section 5.2).
func (r *mdRun) bar(style string) []Cell {
	return []Cell{{"│", r.style(style)}, {" ", r.style("normal")}}
}

// spaces gives n spaces in normal.
func (r *mdRun) spaces(n int) []Cell {
	cells := make([]Cell, n)
	for i := range cells {
		cells[i] = Cell{" ", r.style("normal")}
	}
	return cells
}

// list lays out the list n of the level (file-view, section 5.2).
func (r *mdRun) list(n *ast.List, base string, level int) {
	markStyle := r.style("md_frame")
	if level == 1 {
		markStyle = r.style("md_mark")
	}
	width := 1 // the width of the marker
	if n.IsOrdered() {
		width = len(strconv.Itoa(n.Start+n.ChildCount()-1)) + 1
	}
	i := 0
	for item := n.FirstChild(); item != nil && !r.stopped; item = item.NextSibling() {
		mark := "-"
		switch {
		case n.IsOrdered():
			mark = strconv.Itoa(n.Start+i) + "."
		case level == 1:
			mark = "∙"
		}
		first := r.spaces(2 + width - utf8.RuneCountInString(mark))
		for _, c := range mark {
			first = append(first, Cell{string(c), markStyle})
		}
		first = append(first, Cell{" ", r.style("normal")})
		n0 := r.count
		r.push(first, r.spaces(len(first)))
		r.children(item, base, level, n.IsTight)
		if r.count == n0 {
			r.emit(nil, FileRow{}) // an empty item has its marker
		}
		r.pop()
		// Between the items: a separator in a loose list.
		r.sep, r.sepDepth = !n.IsTight, len(r.parts)
		i++
	}
}

// text lays out the inline elements of the paragraph or heading n in the theme style base.
// level is the level of a heading, or 0 for a paragraph.
// A block with no cells gives no rows.
func (r *mdRun) text(n ast.Node, base string, level int) {
	var flags FileRow
	if level > 0 {
		flags.Keep, flags.Break = true, level == 1
		if level <= 2 {
			flags.NewSection, flags.Section = true, InlineText(n, r.src)
		}
	}
	wr := newWrapper(r.space(), func(cells []Cell) {
		r.emit(cells, flags)
		flags.Break, flags.NewSection, flags.Section = false, false, ""
	})
	// A line break waits for the next cells: a break next to elements with no cells
	// gives no empty row and no space at the start of a row.
	var (
		gave  bool // the block has cells
		soft  bool // a soft break waits
		softS Style
		hards int // the hard breaks that wait
	)
	add := func(s string, st Style) {
		lcs := appendValue(nil, s, st)
		if len(lcs) == 0 {
			return
		}
		for range hards {
			wr.endLine(false)
		}
		if soft && hards == 0 {
			wr.add(lineCell{space, softS, true})
		}
		soft, hards, gave = false, 0, true
		for _, c := range lcs {
			wr.add(c)
		}
	}
	walkInline(n, r.src, func(kind inlineKind, text, style string) {
		if style == "" {
			style = base
		}
		st := r.style(style)
		switch kind {
		case inlineText:
			add(text, st)
		case inlineSoft:
			soft, softS = gave, st
		case inlineHard:
			if gave {
				hards++
			}
		case inlineLinkEnd:
			if k := r.noteOf[text]; k > 0 {
				add("["+strconv.Itoa(k)+"]", r.style("md_dim"))
			}
		}
	})
	if gave {
		wr.endLine(false)
	}
}

// addText adds the graphemes of the value s in st to the line of wr (bbt-format, section 4.3).
// A TAB is a space. A space is a break space. It reports whether s gave graphemes.
func addText(wr *wrapper, s string, st Style) bool {
	lcs := appendValue(nil, s, st)
	for _, c := range lcs {
		wr.add(c)
	}
	return len(lcs) > 0
}

// lines splits a block value into its lines, without the line ends.
func lines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(s, "\n"), "\n")
}

// code lays out the code block n: a code row for each line after a bar (file-view, section 5.4).
func (r *mdRun) code(n *ast.CodeBlock) {
	r.push(r.bar("md_frame"), r.bar("md_frame"))
	var (
		lcs   []lineCell
		cells []Cell
	)
	for _, line := range lines(n.Value.Str(r.src)) {
		lcs = expandTabs(lcs[:0], line, r.style("md_code"))
		cells = cutRow(cells[:0], lcs, r.space(), r.style("md_dim"))
		r.emit(cells, FileRow{})
	}
	r.pop()
}

// thematicBreak lays out a thematic break, cut to the space.
func (r *mdRun) thematicBreak() {
	var cells []Cell
	for _, c := range []rune("∙   ∙   ∙")[:min(9, r.space())] {
		cells = append(cells, Cell{string(c), r.style("md_mark")})
	}
	r.emit(cells, FileRow{})
}

// html lays out the HTML block n: each line wraps as a text line in normal.
// A comment with only spaces after its "-->" gives no rows.
func (r *mdRun) html(n *ast.HTMLBlock) {
	text := HTMLText(n, r.src)
	if strings.TrimLeft(text, " \t\n") == "" {
		return
	}
	wr := newWrapper(r.space(), func(cells []Cell) { r.emit(cells, FileRow{}) })
	for _, line := range lines(text) {
		addText(&wr, line, r.style("normal"))
		wr.endLine(false)
	}
}

// table lays out the table n (file-view, section 5.3).
func (r *mdRun) table(n *east.Table, base string) {
	var (
		rows   [][][]lineCell // the cells of each row: the graphemes of each cell
		styles []Style        // the style of each row
		align  []east.Alignment
		widths []int
	)
	addRow := func(row ast.Node, style string) {
		var cells [][]lineCell
		st := r.style(style)
		for c := row.FirstChild(); c != nil; c = c.NextSibling() {
			cell, _ := c.(*east.TableCell)
			if cell == nil {
				continue
			}
			if len(rows) == 0 {
				align = append(align, cell.Alignment)
			}
			var lcs []lineCell
			walkInline(cell, r.src, func(kind inlineKind, text, s string) {
				cs := st
				if s != "" {
					cs = r.style(s)
				}
				switch kind {
				case inlineText:
					lcs = appendValue(lcs, text, cs)
				case inlineSoft, inlineHard:
					lcs = append(lcs, lineCell{space, cs, true})
				case inlineLinkEnd:
					if k := r.noteOf[text]; k > 0 {
						lcs = appendValue(lcs, "["+strconv.Itoa(k)+"]", r.style("md_dim"))
					}
				}
			})
			cells = append(cells, lcs)
		}
		for i, lcs := range cells[:min(len(cells), len(align))] {
			w := 0
			for _, c := range lcs {
				w += c.g.Width
			}
			if i == len(widths) {
				widths = append(widths, 0)
			}
			widths[i] = max(widths[i], w)
		}
		rows, styles = append(rows, cells), append(styles, st)
	}
	for part := n.FirstChild(); part != nil; part = part.NextSibling() {
		switch part := part.(type) {
		case *east.TableHeader:
			addRow(part, "md_strong")
		case *east.TableRow:
			addRow(part, base)
		case *east.TableBody:
			for row := part.FirstChild(); row != nil; row = row.NextSibling() {
				addRow(row, base)
			}
		}
	}
	total := 0
	for _, w := range widths {
		total += w
	}
	total += 2 * max(0, len(widths)-1)

	var (
		lcs   []lineCell
		cells []Cell
	)
	dim := r.style("md_dim")
	for i, row := range rows {
		lcs = lcs[:0]
		pad := lineCell{space, styles[i], true}
		for j, w := range widths {
			if j > 0 {
				lcs = append(lcs, pad, pad)
			}
			var cell []lineCell
			if j < len(row) {
				cell = row[j]
			}
			extra := w
			for _, c := range cell {
				extra -= c.g.Width
			}
			left := 0
			switch align[j] {
			case east.AlignRight:
				left = extra
			case east.AlignCenter:
				left = extra / 2
			}
			for range left {
				lcs = append(lcs, pad)
			}
			lcs = append(lcs, cell...)
			for range extra - left {
				lcs = append(lcs, pad)
			}
		}
		cells = cutRow(cells[:0], lcs, r.space(), dim)
		r.emit(cells, FileRow{})
		if i == 0 {
			// The rule under the header row.
			lcs = lcs[:0]
			for range total {
				lcs = append(lcs, lineCell{Grapheme{"─", 1}, r.style("md_frame"), false})
			}
			cells = cutRow(cells[:0], lcs, r.space(), dim)
			r.emit(cells, FileRow{})
		}
	}
}

// appendValue appends the graphemes of the value s in st to lcs (bbt-format, section 4.3).
// A TAB is a space.
func appendValue(lcs []lineCell, s string, st Style) []lineCell {
	for _, g := range CleanValue(strings.ReplaceAll(s, "\t", " ")) {
		lcs = append(lcs, lineCell{g, st, g.Text == " "})
	}
	return lcs
}

// notesRows lays out the notes: a separator, then "[n] " and the address for each note
// (file-view, section 5.5). The address wraps after "[n] ".
func (r *mdRun) notesRows() {
	r.sep, r.sepDepth = true, 0
	dim := r.style("md_dim")
	for i, addr := range r.notes {
		if r.stopped {
			return
		}
		// The address wraps after "[n] " when "[n] " fits in W / 2 as a prefix.
		// Else "[n] " is text, and the address wraps from the start of the row.
		num := "[" + strconv.Itoa(i+1) + "] "
		indent := len(num) <= r.w/2
		if indent {
			var first []Cell
			for _, c := range num {
				first = append(first, Cell{string(c), dim})
			}
			r.push(first, r.spaces(len(first)))
		}
		wr := newWrapper(r.space(), func(cells []Cell) { r.emit(cells, FileRow{}) })
		if !indent {
			addText(&wr, num, dim)
		}
		addText(&wr, addr, dim)
		wr.endLine(false)
		if indent {
			r.pop()
		}
	}
}
