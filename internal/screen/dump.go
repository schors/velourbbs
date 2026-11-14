package screen

import (
	"bytes"
	"fmt"
	"io"
	"iter"
)

// WriteDump writes the cell dump of section 9.3 of pages for a terminal of term.rows = rows.
func WriteDump(w io.Writer, rows int, pages iter.Seq[Page]) error {
	n := 0
	for p := range pages {
		n++
		if err := WritePage(w, rows, n, p); err != nil {
			return err
		}
	}
	return nil
}

// WritePage writes the part of the cell dump for page n, 1-based, of a terminal
// of term.rows = rows (section 9.3).
func WritePage(w io.Writer, rows, n int, p Page) error {
	var b bytes.Buffer
	fmt.Fprintf(&b, "page %d %dx%d\n", n, p.Cols, rows)
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
		fmt.Fprintf(&b, "%d:", i+1)
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
