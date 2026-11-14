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

// Page is the rows that the terminal shows at one time (section 6.4).
// A row holds the cells of the screen area. Line gives the row at term.cols cells.
type Page struct {
	Rows   []Row
	Cols   int   // term.cols
	Margin int   // the left margin of the screen area
	Fill   Style // the style of the cells outside the screen area: normal
}

// Line gives row i at p.Cols cells in buf: each cell outside the screen area
// is a space in p.Fill (section 6.2). It reuses the memory of buf.
func (p Page) Line(i int, buf Row) Row {
	buf = buf[:0]
	for range p.Margin {
		buf = append(buf, Cell{" ", p.Fill})
	}
	buf = append(buf, p.Rows[i]...)
	for len(buf) < p.Cols {
		buf = append(buf, Cell{" ", p.Fill})
	}
	return buf
}

// MaxWidth is the largest render width and art width (sections 3.2 and 8.2).
const MaxWidth = 255

// WidthLimits is the [min, max] range of the render width (section 6.1).
type WidthLimits struct {
	Min, Max int
}

// NewWidth gives the width limits of the TOML value a: an array of 2 integers.
// bbs.toml and the front matter use it.
func NewWidth(a []int64) (WidthLimits, error) {
	if len(a) != 2 {
		return WidthLimits{}, errors.New("must be an array of 2 integers")
	}
	if a[0] < 1 || a[0] > a[1] || a[1] > MaxWidth {
		return WidthLimits{}, fmt.Errorf("width [%d, %d] is not in 1 <= min <= max <= %d", a[0], a[1], MaxWidth)
	}
	return WidthLimits{Min: int(a[0]), Max: int(a[1])}, nil
}
