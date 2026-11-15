package screen

import (
	"bytes"
	"fmt"
	"io"
	"iter"
)

// WriteDump writes the cell dump of screen-render, section 3.3, of pages.
func WriteDump(w io.Writer, pages iter.Seq[Page]) error {
	n := 0
	for p := range pages {
		n++
		if err := WritePage(w, n, p); err != nil {
			return err
		}
	}
	return nil
}

// WritePage writes the part of the cell dump for page n, 1-based (screen-render, section 3.3).
func WritePage(w io.Writer, n int, p Page) error {
	var b bytes.Buffer
	fmt.Fprintf(&b, "page %d %dx%d fill %s\n", n, p.Cols, p.Height, p.Fill)
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
			for end < len(line) && line[end].Style == line[start].Style {
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
