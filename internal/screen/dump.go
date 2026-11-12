package screen

import (
	"bytes"
	"fmt"
	"io"
)

// WriteDump writes the cell dump of section 9.3 for a terminal of cols x rows.
func WriteDump(w io.Writer, cols, rows int, pages []Page) error {
	var b bytes.Buffer
	for i, page := range pages {
		fmt.Fprintf(&b, "page %d/%d %dx%d\n", i+1, len(pages), cols, rows)
		for _, row := range page {
			b.WriteByte('|')
			for _, c := range row {
				b.WriteString(c.Grapheme)
			}
			b.WriteString("|\n")
		}
		b.WriteString("styles\n")
		for n, row := range page {
			fmt.Fprintf(&b, "%d:", n+1)
			sep := " "
			for start := 0; start < len(row); {
				end := start + 1
				for end < len(row) && row[end].Style == row[start].Style {
					end++
				}
				fmt.Fprintf(&b, "%s%d-%d %s", sep, start+1, end, row[start].Style)
				sep = ", "
				start = end
			}
			b.WriteByte('\n')
		}
	}
	_, err := b.WriteTo(w)
	return err
}
