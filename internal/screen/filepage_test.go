package screen

import (
	"iter"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// flagLayout gives a layout of one row for each letter of kinds: "r" a row, "b" a row
// of a level-1 heading (Break and Keep), "k" a heading row (Keep), "s" a Sep row.
// The restart point of a row is its index, and its cell is the letter and the index.
func flagLayout(kinds string) FileLayout {
	return func(from int) iter.Seq[FileRow] {
		return func(yield func(FileRow) bool) {
			for i := from; i < len(kinds); i++ {
				k := kinds[i]
				row := FileRow{Cells: []Cell{{Grapheme: string(k) + strconv.Itoa(i)}}, From: i,
					Break: k == 'b', Keep: k == 'b' || k == 'k', Sep: k == 's'}
				if !yield(row) {
					return
				}
			}
		}
	}
}

// pageText gives the rows of each file page of l: the rows separated with spaces,
// the pages with "|". A nil row is "-".
func pageText(l FileLayout, h int) string {
	var pages []string
	for _, p := range SplitPages(l, h) {
		var rows []string
		for _, r := range PageRows(l, p, h) {
			if r == nil {
				rows = append(rows, "-")
				continue
			}
			var b strings.Builder
			for _, c := range r {
				b.WriteString(c.Grapheme)
			}
			rows = append(rows, b.String())
		}
		pages = append(pages, strings.Join(rows, " "))
	}
	return strings.Join(pages, "|")
}

func TestSplitPagesFlags(t *testing.T) {
	for _, tc := range []struct {
		name, kinds string
		h           int
		want        string
	}{
		{"full pages", "rrrrr", 2, "r0 r1|r2 r3|r4 -"},
		{"no rows", "", 2, "- -"},
		{"break at the first row", "bsrr", 4, "b0 s1 r2 r3"},
		{"break later", "rsbsr", 4, "r0 s1 - -|b2 s3 r4 -"},
		{"break on an empty page", "rrbr", 2, "r0 r1|b2 r3"},
		{"group fits", "rrksr", 5, "r0 r1 k2 s3 r4"},
		{"group moves", "rrrksr", 4, "r0 r1 r2 -|k3 s4 r5 -"},
		{"group on a full page", "rrksr", 2, "r0 r1|k2 s3|r4 -"},
		{"group longer than h", "rkksr", 2, "r0 k1|k2 s3|r4 -"},
		{"chain of headings moves", "rrrkskssr", 6, "r0 r1 r2 - - -|k3 s4 k5 s6 s7 r8"},
		{"break after a group", "rksbsr", 4, "r0 k1 s2 -|b3 s4 r5 -"},
		{"group at the end", "rk", 4, "r0 k1 - -"},
		{"level-1 heading of two rows", "rbkr", 4, "r0 - - -|b1 k2 r3 -"},
		{"sep at the start of a page", "rrsr", 2, "r0 r1|r3 -"},
		{"sep at the start of the file", "sr", 2, "r1 -"},
		{"sep at the end", "rrs", 2, "r0 r1"},
	} {
		if got := pageText(flagLayout(tc.kinds), tc.h); got != tc.want {
			t.Errorf("%s: %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestSplitPagesRowLimit(t *testing.T) {
	kinds := strings.Repeat("r", MaxFileRows+5)
	pages := SplitPages(flagLayout(kinds), 3)
	n := 0
	for _, p := range pages {
		n += p.N
	}
	if n != MaxFileRows {
		t.Fatalf("%d rows, want %d", n, MaxFileRows)
	}
	last := pages[len(pages)-1]
	if want := (FilePage{From: MaxFileRows - 1, N: 1}); last != want {
		t.Errorf("last page %+v, want %+v", last, want)
	}
	if got := PageRows(flagLayout(kinds), last, 3); !reflect.DeepEqual(got[1:], []Row{nil, nil}) {
		t.Errorf("last page rows %v", got)
	}
}
