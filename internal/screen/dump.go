package screen

import (
	"bytes"
	"fmt"
	"io"
	"iter"
)

// WriteDump writes the cell dump of screen-render, section 3.3, of pages.
func WriteDump(w io.Writer, pages iter.Seq[Page]) error {
	for p := range pages {
		if err := WritePage(w, p); err != nil {
			return err
		}
	}
	return nil
}

// Header gives the first line of the part of page p, without the line end:
// the page number, the terminal size and fill (screen-render, sections 3.3 and 3.7).
func (p Page) Header() string {
	return fmt.Sprintf("page %d %dx%d fill %s", p.Num, p.Cols, p.Height, p.Fill)
}

// WritePage writes the part of the cell dump for the page p (screen-render, section 3.3).
func WritePage(w io.Writer, p Page) error {
	return writePart(w, p.Header(), p)
}

// GridPage gives the grid of p (Page.Grid) as a page: its rows are all terminal rows,
// without a margin. The line form of velourbbs preview writes it (live-screen, section 7).
func (p Page) GridPage() Page {
	return Page{Num: p.Num, Rows: p.Grid(), Cols: p.Cols, Height: p.Height, Normal: p.Normal, Fill: p.Fill}
}

// GridHeader gives the first line of the grid dump of p, without the line end (live-screen, section 7).
func (p Page) GridHeader() string {
	return fmt.Sprintf("grid %dx%d", p.Cols, p.Height)
}

// WriteGrid writes the grid dump of the grid of p (live-screen, section 7): the cell dump
// of all terminal rows with the first line GridHeader.
func WriteGrid(w io.Writer, p Page) error {
	return writePart(w, p.GridHeader(), p.GridPage())
}

// writePart writes a part of the cell dump: the line header, then the rows of p and their styles.
func writePart(w io.Writer, header string, p Page) error {
	var b bytes.Buffer
	b.WriteString(header)
	b.WriteByte('\n')
	var line Row
	for i := range p.Rows {
		line = p.Line(i, line)
		b.WriteByte('|')
		for _, c := range line {
			b.WriteString(c.Grapheme)
		}
		b.WriteString("|\n")
	}
	b.WriteString("styles\n")
	for i := range p.Rows {
		line = p.Line(i, line)
		fmt.Fprintf(&b, "%d:", p.Top+i+1) // the terminal row
		sep := " "
		for start := 0; start < len(line); {
			end := start + 1
			for end < len(line) && dumpStyle(line[end].Style) == dumpStyle(line[start].Style) {
				end++
			}
			fmt.Fprintf(&b, "%s%d-%d %s", sep, start+1, end, line[start].Style)
			sep = ", "
			start = end
		}
		b.WriteByte('\n')
	}
	_, err := b.WriteTo(w)
	return err
}

// dumpStyle gives the style s with each VGA colour as its index.
// So a VGA colour and the same index are one run (section 3.3).
func dumpStyle(s Style) Style {
	for _, c := range []*Colour{&s.Fg, &s.Bg} {
		if c.Kind == VGA {
			c.Kind = Index
		}
	}
	return s
}
