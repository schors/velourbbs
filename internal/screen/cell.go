package screen

import "fmt"

// Cell is one cell of the screen: one grapheme and a style (section 7.1).
// The second cell of a wide grapheme has an empty Grapheme.
type Cell struct {
	Grapheme string
	Style    Style
}

// Row is one row of cells.
type Row []Cell

// Page is the rows that the terminal shows at one time (section 6.4).
type Page []Row

// MaxWidth is the largest render width and art width (sections 3.2 and 8.2).
const MaxWidth = 255

// WidthLimits is the [min, max] range of the render width (section 6.1).
type WidthLimits struct {
	Min, Max int
}

// Check reports an error unless 1 <= Min <= Max <= MaxWidth.
func (w WidthLimits) Check() error {
	if w.Min < 1 || w.Min > w.Max || w.Max > MaxWidth {
		return fmt.Errorf("width [%d, %d] is not in 1 <= min <= max <= %d", w.Min, w.Max, MaxWidth)
	}
	return nil
}
