package screen

import "fmt"

// LayoutKind is a kind of layout warning (screen-render, section 3.6).
type LayoutKind int

// The kinds of layout warnings.
const (
	LayoutWrap LayoutKind = iota
	LayoutMore
	LayoutHeight
	LayoutWidth
	LayoutArt
)

// LayoutNote is one layout warning of a render.
// N and M are the values of the message: Wrap has N, More has N, Height has N and M = term.rows,
// Width and Art have N cells and M, the cells of the screen area.
type LayoutNote struct {
	Kind LayoutKind
	Pos  Pos
	N, M int
	Art  string // the art name of Art
}

// Message gives the text of the warning without the size.
func (n LayoutNote) Message() string {
	switch n.Kind {
	case LayoutWrap:
		return fmt.Sprintf("the line wraps to %d rows", n.N)
	case LayoutMore:
		return fmt.Sprintf("%d rows, the screen gets a More row", n.N)
	case LayoutHeight:
		return fmt.Sprintf("%d rows, the rows after %d are cut", n.N, n.M)
	case LayoutWidth:
		return fmt.Sprintf("the row is %d cells wide, the screen shows %d", n.N, n.M)
	}
	return fmt.Sprintf("art %q is %d cells wide, the screen shows %d", n.Art, n.N, n.M)
}

// LayoutLog collects the layout notes of the renders of an Env with this log, in the order of the render.
type LayoutLog struct {
	Notes []LayoutNote
}

func (l *LayoutLog) add(n LayoutNote) {
	n.Pos.Col = 1
	l.Notes = append(l.Notes, n)
}

// ink gives the cells of cells up to the last cell that is not a space.
func ink(cells []Cell) int {
	for i := len(cells) - 1; i >= 0; i-- {
		if cells[i].Grapheme != " " {
			return i + 1
		}
	}
	return 0
}

// inkGraphemes gives the cells of gs up to the last grapheme that is not a space.
func inkGraphemes(gs []Grapheme) int {
	n, w := 0, 0
	for _, g := range gs {
		w += g.Width
		if g.Text != " " {
			n = w
		}
	}
	return n
}
