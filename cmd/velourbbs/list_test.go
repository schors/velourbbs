package main

import (
	"os"
	"path/filepath"
	"testing"
)

// listGoldens is the golden dumps of the test BBS testdata/list (list-view, section 12):
// the file name in testdata/list and the arguments of velourbbs preview.
var listGoldens = []struct {
	name string
	args []string
}{
	// The first page and the last page with empty rows, a list line in a block.
	{"parts-80x24.dump", []string{"parts", "--set", "arg.dir=parts"}},
	// A list line in a fragment, wider than W: each row is cut at W.
	{"parts-40x10.dump", []string{"parts", "--set", "arg.dir=parts", "--size", "40x10"}},
	// An empty list: one page with P empty rows and list.empty.
	{"empty-80x24.dump", []string{"parts", "--set", "arg.dir=empty"}},
	// P smaller than rows: rows = 12 and 16 rows give P = 6 at 80x10.
	{"news-80x10.dump", []string{"news", "--size", "80x10"}},
	// Each page action; the disabled page next on the last page and page prev on the first page;
	// back to the page of the frame; theme on page 2 keeps the page; an empty list.
	{"keys-40x10.dump", []string{"main", "--size", "40x10", "--keys", "1,PGDN,PGDN,PGUP,PGUP,END,HOME,END,N,ESC,T,ESC,2,PGDN,ESC,G"}},
	// The cursor of select: UP and DOWN on a page, over a page border and at the ends of the list;
	// a page action puts the cursor on the first row; ENTER to a file screen and back to the same cursor;
	// ENTER and DOWN in an empty list do nothing.
	{"cursor-40x10.dump", []string{"main", "--size", "40x10", "--keys",
		"1,DOWN,UP,UP,DOWN,DOWN,DOWN,DOWN,UP,END,DOWN,DOWN,DOWN,ENTER,ESC,HOME,ESC,2,ENTER,DOWN,ESC,G"}},
	// The rows of a --set value that no goto gives: ENTER shows the file of the row.
	{"set-40x10.dump", []string{"parts", "--set", "arg.dir=more", "--size", "40x10", "--keys", "ENTER,ESC"}},
	// The list row in art (list-view, section 9): the first page and the last page with an empty
	// rectangle row; the selected row in select; the frame and the fields outside the rectangle stay.
	// The box {#if item.PGDN}[PgDn]{/if} of the bottom border is true on page 1 and false on page 2
	// (art-format, section 5).
	{"shelf-80x24.dump", []string{"shelf", "--set", "arg.dir=parts"}},
	// P smaller than rows: the art does not shrink, so P = 3 leaves the fourth rectangle row empty.
	{"shelf-80x10.dump", []string{"shelf", "--set", "arg.dir=parts", "--size", "80x10"}},
	// An empty list: each rectangle row is spaces in the styles of the list row.
	{"shelf-empty-40x24.dump", []string{"shelf", "--set", "arg.dir=empty", "--size", "40x24"}},
	// The cursor in art: UP at the start and DOWN at the end of the list do nothing; DOWN and UP
	// over a page border; ENTER on row 7 to a file screen and back to the same cursor; PGUP.
	{"shelf-keys-40x24.dump", []string{"main", "--size", "40x24", "--keys",
		"4,UP,DOWN,DOWN,DOWN,DOWN,UP,PGDN,DOWN,DOWN,DOWN,ENTER,ESC,PGUP,ESC,G"}},
}

func TestPreviewList(t *testing.T) {
	out, errOut, status := runIn(t, "list", "check")
	if out != "" || errOut != "" || status != 0 {
		t.Errorf("check: stdout %q, stderr %q, status %d; want no output and status 0", out, errOut, status)
	}
	for _, g := range listGoldens {
		t.Run(g.name, func(t *testing.T) {
			file := filepath.Join("testdata", "list", g.name)
			want, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			out, errOut, status := runIn(t, "list", append([]string{"preview"}, g.args...)...)
			if errOut != "" || status != 0 || out != string(want) {
				t.Errorf("%s: stderr %q, status %d, got\n%s\nwant\n%s", file, errOut, status, out, want)
			}
		})
	}
}

// An error of the provider for a --set value is an invalid argument (screen-render, section 3.2).
func TestPreviewListArg(t *testing.T) {
	out, errOut, status := runIn(t, "list", "preview", "parts", "--set", "arg.dir=nope")
	want := "velourbbs preview: data: args: dir: directory \"nope\" does not exist\n"
	if out != "" || errOut != want || status != exitError {
		t.Errorf("stdout %q, stderr %q, status %d; want stderr %q, status %d", out, errOut, status, want, exitError)
	}
}
