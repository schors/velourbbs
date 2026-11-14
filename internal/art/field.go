package art

import (
	"strings"

	"newbbs/internal/screen"
)

// findFields finds the field tokens in the rows of the art (section 8.4).
// A token is "{", optional spaces, a path, optional spaces and "}" on one row.
// Each other "{" is a plain character.
func findFields(d *screen.FileDiags, rows []screen.Row) []screen.ArtField {
	var fields []screen.ArtField
	for i, row := range rows {
		for c := 0; c < len(row); c++ {
			if row[c].Grapheme != "{" {
				continue
			}
			path, start, end, ok := scanToken(row, c)
			if !ok {
				continue
			}
			if checkToken(d, i+1, row[c:end+1], c+1) {
				fields = append(fields, screen.ArtField{
					Row: i + 1, Col: c + 1, Width: end - c + 1, Path: path,
					Right: start > c+1, Style: row[c].Style,
				})
			}
			c = end
		}
	}
	return fields
}

// scanToken scans the token at row[c] == "{". It gives the path, the index of its
// first cell and the index of "}".
func scanToken(row screen.Row, c int) (path string, start, end int, ok bool) {
	j := skipSpaces(row, c+1)
	start = j
	var b strings.Builder
	for ; j < len(row) && isPathCell(row[j].Grapheme); j++ {
		b.WriteString(row[j].Grapheme)
	}
	path = b.String()
	j = skipSpaces(row, j)
	if j == len(row) || row[j].Grapheme != "}" || !screen.IsPath(path) {
		return "", 0, 0, false
	}
	return path, start, j, true
}

func skipSpaces(row screen.Row, j int) int {
	for j < len(row) && row[j].Grapheme == " " {
		j++
	}
	return j
}

// isPathCell reports whether g is one character of the path syntax.
func isPathCell(g string) bool {
	if len(g) != 1 {
		return false
	}
	c := g[0]
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' || c == '_' || c == '.'
}

// checkToken checks the colours of the token cells at the row and the column col.
// All cells must have one background. All non-space cells must have one foreground.
func checkToken(d *screen.FileDiags, row int, cells screen.Row, col int) bool {
	var text strings.Builder
	for _, c := range cells {
		text.WriteString(c.Grapheme)
	}
	ok := true
	for i, c := range cells {
		if c.Style.Bg != cells[0].Style.Bg {
			d.Errorf(row, col, "token %q changes the background: %s at col %d, %s at col %d",
				text.String(), cells[0].Style.Bg, col, c.Style.Bg, col+i)
			ok = false
			break
		}
	}
	first := -1
	for i, c := range cells {
		if c.Grapheme == " " {
			continue
		}
		if first < 0 {
			first = i
		} else if c.Style.Fg != cells[first].Style.Fg {
			d.Errorf(row, col, "token %q changes the foreground: %s at col %d, %s at col %d",
				text.String(), cells[first].Style.Fg, col+first, c.Style.Fg, col+i)
			ok = false
			break
		}
	}
	return ok
}
