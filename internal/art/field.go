package art

import (
	"strings"

	"github.com/schors/velourbbs/internal/screen"
)

// findTokens finds the field tokens (section 4) and the boxes (section 5) in the rows of the art.
// It resolves the paths with the aliases fields. A field token is "{", optional spaces, a path,
// the optional live mark, optional spaces and "}" on one row (scanToken). A "{" followed by "#", "/", "!" or ">" starts a token
// that ends at the first "}" on the row. Each other "{" is a plain character.
func findTokens(d *screen.FileDiags, rows []screen.Row, fields screen.Fields) ([]screen.ArtField, []screen.ArtBox) {
	var found []screen.ArtField
	var boxes []screen.ArtBox
	for i, row := range rows {
		var open *screen.ArtBox // the box in progress on this row
		for c := 0; c < len(row); c++ {
			if row[c].Grapheme != "{" {
				continue
			}
			if c+1 < len(row) && strings.Contains("#/!>", row[c+1].Grapheme) && row[c+1].Grapheme != "" {
				end := closing(row, c)
				if end < 0 {
					d.Errorf(i+1, c+1, `token has no "}"`)
					break
				}
				open, boxes = blockToken(d, row, i+1, c, end, open, boxes, fields)
				c = end
				continue
			}
			text, start, end, ok := scanToken(row, c)
			if !ok {
				continue
			}
			valid := checkToken(d, i+1, row[c:end+1], c+1)
			ref, m, err := screen.LookupField(text, fields)
			if err != nil {
				d.Errorf(i+1, c+1, "%v", err)
				valid = false
			}
			if valid {
				f := screen.ArtField{
					Row: i + 1, Col: c + 1, Width: end - c + 1, Ref: ref,
					Right: start > c+1, Style: row[c].Style,
				}
				if m.Live && ref.Ticks() { // the mark on another field gives no slot (live-screen, section 3)
					f.Live, f.Layout = true, m.Layout
				}
				found = append(found, f)
			}
			c = end
		}
		if open != nil {
			d.Errorf(open.Row, open.Col, "{#if} has no {/if} on its row")
		}
	}
	return found, boxes
}

// blockToken handles the token row[c:end+1] that starts with "{#", "{/", "{!" or "{>" on the row n
// (section 5). open is the box in progress, else nil. It gives the box in progress and the boxes.
// A box with an error of its condition is not in boxes.
func blockToken(d *screen.FileDiags, row screen.Row, n, c, end int, open *screen.ArtBox,
	boxes []screen.ArtBox, fields screen.Fields) (*screen.ArtBox, []screen.ArtBox) {
	text := cellText(row[c+1 : end])
	isIf := text == "#if" || strings.HasPrefix(text, "#if ")
	if isIf || text == "/if" { // a block token has the colours of a token (section 4)
		checkToken(d, n, row[c:end+1], c+1)
	}
	switch {
	case isIf:
		if open != nil {
			d.Errorf(n, c+1, "{#if} in a box: boxes do not nest")
			return open, boxes
		}
		cond, err := screen.ParseCond(text[len("#if"):], fields, true)
		if err != nil {
			d.Errorf(n, c+1, "%v", err)
		} else {
			for _, r := range cond.Refs() {
				if r.Row() != "" {
					d.Errorf(n, c+1, "field %q is not valid in a box condition", r.String())
					cond = nil
					break
				}
			}
		}
		return &screen.ArtBox{Row: n, Col: c + 1, Content: end + 2, Cond: cond, Style: row[c].Style}, boxes
	case text == "/if":
		if open == nil {
			d.Errorf(n, c+1, "{/if} without {#if}")
			return nil, boxes
		}
		open.Len, open.End = c+1-open.Content, end+1
		if open.Cond != nil {
			boxes = append(boxes, *open)
		}
		return nil, boxes
	}
	d.Errorf(n, c+1, "token %q is not valid in art", "{"+text+"}")
	return open, boxes
}

// closing gives the index of the first "}" after row[c], else -1.
func closing(row screen.Row, c int) int {
	for j := c + 1; j < len(row); j++ {
		if row[j].Grapheme == "}" {
			return j
		}
	}
	return -1
}

// cellText gives the text of cells.
func cellText(cells screen.Row) string {
	var b strings.Builder
	for _, c := range cells {
		b.WriteString(c.Grapheme)
	}
	return b.String()
}

// scanToken scans the field token at row[c] == "{": a path with optional spaces around it,
// or a path with " live" after it up to the first "}" (live-screen, section 3).
// It gives the text of the token without the spaces around it, the index of the first cell
// of the path and the index of "}".
func scanToken(row screen.Row, c int) (text string, start, end int, ok bool) {
	j := skipSpaces(row, c+1)
	start = j
	var b strings.Builder
	for ; j < len(row) && isPathCell(row[j].Grapheme); j++ {
		b.WriteString(row[j].Grapheme)
	}
	path := b.String()
	if !screen.IsPath(path) {
		return "", 0, 0, false
	}
	if k := skipSpaces(row, j); k < len(row) && row[k].Grapheme == "}" {
		return path, start, k, true
	}
	if mark := cellText(row[j:min(j+6, len(row))]); mark != " live " && mark != " live}" {
		return "", 0, 0, false
	}
	end = closing(row, j)
	if end < 0 {
		return "", 0, 0, false
	}
	return strings.TrimRight(cellText(row[start:end]), " "), start, end, true
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

// checkBoxes reports each box in the list row l (section 5).
func checkBoxes(d *screen.FileDiags, boxes []screen.ArtBox, l *screen.ArtList) {
	for _, b := range boxes {
		if l != nil && b.Row == l.Row {
			d.Errorf(b.Row, b.Col, "box in the list row")
		}
	}
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
