package screen

// wrapper wraps the graphemes of source lines into rows of w cells (section 6.3).
// It gives each full row to emit. The cells of the row are valid only in the call of emit.
// The buffers of a line have at most w cells.
type wrapper struct {
	w    int
	emit func(cells []Cell)

	cells  []Cell     // the cells of the current row
	col    int        // the width of cells
	lead   bool       // the line has only spaces so far: they join the first word
	spaces []lineCell // the break spaces before the word, at most w
	word   []lineCell // the current word, at most w cells
	ww     int        // the width of word
	long   bool       // the word is wider than w: its graphemes go on the rows as they come
}

// lineCell is one grapheme of a source line before the wrap.
type lineCell struct {
	g     Grapheme
	style Style
	brk   bool // a space where the line can break
}

// newWrapper gives a wrapper of the width w at the start of a line.
func newWrapper(w int, emit func(cells []Cell)) wrapper {
	return wrapper{w: w, emit: emit, lead: true}
}

// add adds one grapheme to the line.
func (r *wrapper) add(c lineCell) {
	switch {
	case r.lead && c.g.Text == " ":
		// The spaces at the start of the line stay up to w and join the first word.
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
			// A word wider than w starts on a new row and breaks at a grapheme boundary.
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
func (r *wrapper) endWord() {
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
// (section 6.2). The spaces at the end of the line stay up to w.
func (r *wrapper) endLine(block bool) {
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
// when w is 1, it becomes a space.
func (r *wrapper) place(c lineCell) {
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
func (r *wrapper) newRow() {
	r.emit(r.cells)
	r.cells, r.col = r.cells[:0], 0
}
