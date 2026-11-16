package screen

import (
	"iter"
	"slices"
)

// MaxFileRows is the row limit of the layout of a content file (file-view, sections 3.2 and 8).
const MaxFileRows = 10000

// FileRow is one row of the layout of a content file (file-view, section 3.2).
// Cells is valid only in the call of the consumer.
// From is the restart point of the row: FileLayout(From) gives the rows from the first row
// with this From. The rows of one restart point follow each other.
// The flags are the page break rules of file-view, section 5.6.
type FileRow struct {
	Cells []Cell
	From  int
	// Break marks the first row of a level-1 heading: a file page starts at it,
	// except at the first file row. The other rows of the heading have only Keep.
	Break bool
	Keep  bool // a heading row: on one file page with the next row that is not Sep
	Sep   bool // an empty row between blocks: a file page does not start with it
}

// FileLayout gives the rows of a content file from the restart point from.
type FileLayout func(from int) iter.Seq[FileRow]

// FilePage is a file page: the N rows after the first Skip rows from the restart point From.
type FilePage struct {
	From, Skip, N int
}

// filePos is the position of a file row: its restart point and the rows before it from that point.
type filePos struct {
	from, skip int
	sep        bool
}

// SplitPages splits the rows of l into file pages of at most h rows
// (file-view, sections 3.2 and 5.6). It reads only the flags of the rows.
// It stops after MaxFileRows rows. A file with no rows has one file page with no rows.
func SplitPages(l FileLayout, h int) []FilePage {
	s := splitter{h: h, pages: []FilePage{{}}}
	var pos filePos
	count := 0
	for row := range l(0) {
		if count == MaxFileRows {
			break
		}
		if count > 0 && row.From == pos.from {
			pos.skip++
		} else {
			pos = filePos{from: row.From}
		}
		pos.sep = row.Sep
		count++
		s.add(row, pos)
	}
	s.flush()
	return s.pages
}

// splitter holds the state of SplitPages. The last page is the open page.
type splitter struct {
	h     int
	pages []FilePage
	group []filePos // the open group of Keep rows and the Sep rows in it
}

// add adds a row at pos.
func (s *splitter) add(row FileRow, pos filePos) {
	if row.Break {
		s.flush()
		s.newPage()
	}
	if row.Keep || row.Sep && len(s.group) > 0 {
		s.group = append(s.group, pos)
		return
	}
	if len(s.group) > 0 {
		// A heading and the first row of the next block are on one file page.
		n, g := s.pages[len(s.pages)-1].N, len(s.group)+1
		if n+g > s.h && g <= s.h {
			s.newPage()
		}
		s.flush()
	}
	s.place(pos)
}

// flush places the rows of the open group one by one.
func (s *splitter) flush() {
	for _, pos := range s.group {
		s.place(pos)
	}
	s.group = s.group[:0]
}

// newPage starts a new page when the open page has rows.
func (s *splitter) newPage() {
	if s.pages[len(s.pages)-1].N > 0 {
		s.pages = append(s.pages, FilePage{})
	}
}

// place puts the row at pos on the open page, or on a new page when the open page is full.
// A page does not start with a Sep row: the row is dropped.
func (s *splitter) place(pos filePos) {
	p := &s.pages[len(s.pages)-1]
	full := p.N == s.h
	if pos.sep && (p.N == 0 || full) {
		return
	}
	if full {
		s.pages = append(s.pages, FilePage{})
		p = &s.pages[len(s.pages)-1]
	}
	if p.N == 0 {
		p.From, p.Skip = pos.from, pos.skip
	}
	p.N++
}

// PageRows gives the h rows of the file page p of l: the rows of the page,
// then nil rows up to h. It lays out the rows from the restart point of the page.
func PageRows(l FileLayout, p FilePage, h int) []Row {
	rows := make([]Row, 0, h)
	if p.N > 0 {
		skip := p.Skip
		for row := range l(p.From) {
			if skip > 0 {
				skip--
				continue
			}
			rows = append(rows, slices.Clone(Row(row.Cells)))
			if len(rows) == p.N {
				break
			}
		}
	}
	return append(rows, make([]Row, h-len(rows))...)
}
