package screen

import (
	"errors"
	"fmt"
)

// Cell is one cell of the screen: one grapheme and a style (section 7.1).
// The second cell of a wide grapheme has an empty Grapheme.
type Cell struct {
	Grapheme string
	Style    Style
}

// Row is one row of cells.
type Row []Cell

// Page is the rows that the terminal shows at one time (sections 6.4 and 6.5).
// A row holds the cells of the screen area. Line gives the row at term.cols cells.
// The terminal rows above and below the rows are the fill rows.
type Page struct {
	Num    int // the page number from 1: file.page on a file screen (screen-render, section 3.3)
	Rows   []Row
	Cols   int   // term.cols
	Height int   // term.rows
	Top    int   // the fill rows above the rows: the first row is the terminal row Top + 1
	Margin int   // the left margin of the screen area
	Normal Style // the style normal: the cells outside the screen area and the fill rows
	Fill   Fill  // the cells of the fill rows
}

// Line gives row i at p.Cols cells in buf: each cell outside the screen area
// is a space in p.Normal (section 6.2). It reuses the memory of buf.
func (p Page) Line(i int, buf Row) Row {
	buf = buf[:0]
	for range p.Margin {
		buf = append(buf, Cell{" ", p.Normal})
	}
	buf = append(buf, p.Rows[i]...)
	for len(buf) < p.Cols {
		buf = append(buf, Cell{" ", p.Normal})
	}
	return buf
}

// MaxWidth is the largest render width and art width (section 3.2 and art-format, section 2).
const MaxWidth = 255

// WidthLimits is the [min, max] range of the render width (section 6.1).
type WidthLimits struct {
	Min, Max int
}

// NewWidth gives the width limits of the TOML value a: an array of 2 integers.
// screens/screens.toml and the front matter use it.
func NewWidth(a []int64) (WidthLimits, error) {
	if len(a) != 2 {
		return WidthLimits{}, errors.New("must be an array of 2 integers")
	}
	if a[0] < 1 || a[0] > a[1] || a[1] > MaxWidth {
		return WidthLimits{}, fmt.Errorf("width [%d, %d] is not in 1 <= min <= max <= %d", a[0], a[1], MaxWidth)
	}
	return WidthLimits{Min: int(a[0]), Max: int(a[1])}, nil
}
