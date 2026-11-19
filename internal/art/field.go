package art

import (
	"strings"

	"github.com/schors/velourbbs/internal/screen"
)

// findFields finds the field tokens in the rows of the art (section 4)
// and resolves their paths with the aliases fields.
// A token is "{", optional spaces, a path, optional spaces and "}" on one row.
// Each other "{" is a plain character.
func findFields(d *screen.FileDiags, rows []screen.Row, fields screen.Fields) []screen.ArtField {
	var found []screen.ArtField
	for i, row := range rows {
		for c := 0; c < len(row); c++ {
			if row[c].Grapheme != "{" {
				continue
			}
			path, start, end, ok := scanToken(row, c)
			if !ok {
				continue
			}
			valid := checkToken(d, i+1, row[c:end+1], c+1)
			ref, err := screen.LookupField(path, fields)
			if err != nil {
				d.Errorf(i+1, c+1, "%v", err)
				valid = false
			}
			if valid {
				found = append(found, screen.ArtField{
					Row: i + 1, Col: c + 1, Width: end - c + 1, Ref: ref,
					Right: start > c+1, Style: row[c].Style,
				})
			}
			c = end
		}
	}
	return found
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

// findList finds the list row: the art row with row.* fields (list-view, section 9).
// A row.* field on another row is an error, once for each row.
func findList(d *screen.FileDiags, fields []screen.ArtField) *screen.ArtList {
	var l *screen.ArtList
	bad := 0 // the last row with an error
	for _, f := range fields {
		switch {
		case f.Ref.Row() == "":
		case l == nil:
			l = &screen.ArtList{Row: f.Row, First: f.Col, Last: f.Col + f.Width - 1}
		case f.Row == l.Row:
			l.Last = f.Col + f.Width - 1
		case f.Row != bad:
			d.Errorf(f.Row, f.Col, "field %q: art has at most one list row, the list row is row %d", f.Ref.String(), l.Row)
			bad = f.Row
		}
	}
	return l
}
