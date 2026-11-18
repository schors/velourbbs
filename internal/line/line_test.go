package line

import (
	"fmt"
	"strings"
	"testing"

	"github.com/schors/velourbbs/internal/screen"
)

// typed gives a line in a field of M = max after the keys ks, outside a paste.
func typed(t *testing.T, max int, ks ...string) *Line {
	t.Helper()
	l := New(Field{Row: 1, Col: 1, Max: max}, 80, 24)
	for _, k := range ks {
		l.Key(k, false)
	}
	return l
}

// state writes the line with | at the cursor position.
func state(l *Line) string {
	r := []rune(l.Text())
	return string(r[:l.cur]) + "|" + string(r[l.cur:])
}

// TestKeys covers each key of section 4 at the start, in the middle and at the end
// of the line.
func TestKeys(t *testing.T) {
	tests := []struct {
		name string
		keys []string // after "abc"
		want string
		res  Result // the result of the last key
	}{
		{"char at end", []string{"d"}, "abcd|", Changed},
		{"char in middle", []string{"LEFT", "d"}, "abd|c", Changed},
		{"char at start", []string{"HOME", "d"}, "d|abc", Changed},
		{"space at end", []string{"SPACE"}, "abc |", Changed},
		{"space in middle", []string{"LEFT", "SPACE"}, "ab |c", Changed},
		{"space at start", []string{"HOME", "SPACE"}, " |abc", Changed},
		{"left at end", []string{"LEFT"}, "ab|c", Changed},
		{"left in middle", []string{"LEFT", "LEFT"}, "a|bc", Changed},
		{"left at start", []string{"HOME", "LEFT"}, "|abc", None},
		{"right at end", []string{"RIGHT"}, "abc|", None},
		{"right in middle", []string{"HOME", "RIGHT"}, "a|bc", Changed},
		{"right at start", []string{"HOME", "RIGHT", "RIGHT"}, "ab|c", Changed},
		{"home at end", []string{"HOME"}, "|abc", Changed},
		{"home in middle", []string{"LEFT", "HOME"}, "|abc", Changed},
		{"home at start", []string{"HOME", "HOME"}, "|abc", None},
		{"end at end", []string{"END"}, "abc|", None},
		{"end in middle", []string{"LEFT", "END"}, "abc|", Changed},
		{"end at start", []string{"HOME", "END"}, "abc|", Changed},
		{"bs at end", []string{"BS"}, "ab|", Changed},
		{"bs in middle", []string{"LEFT", "BS"}, "a|c", Changed},
		{"bs at start", []string{"HOME", "BS"}, "|abc", None},
		{"del at end", []string{"DEL"}, "abc|", None},
		{"del in middle", []string{"LEFT", "DEL"}, "ab|", Changed},
		{"del at start", []string{"HOME", "DEL"}, "|bc", Changed},
		{"enter at end", []string{"ENTER"}, "abc|", Enter},
		{"enter in middle", []string{"LEFT", "ENTER"}, "ab|c", Enter},
		{"enter at start", []string{"HOME", "ENTER"}, "|abc", Enter},
		{"esc at end", []string{"ESC"}, "abc|", Cancel},
		{"esc in middle", []string{"LEFT", "ESC"}, "ab|c", Cancel},
		{"esc at start", []string{"HOME", "ESC"}, "|abc", Cancel},
	}
	ignored := []string{"TAB", "UP", "DOWN", "PGUP", "PGDN"}
	for i := 1; i <= 12; i++ {
		ignored = append(ignored, fmt.Sprintf("F%d", i))
	}
	for _, k := range ignored {
		for _, pos := range []string{"HOME", "LEFT", "END"} {
			tests = append(tests, struct {
				name string
				keys []string
				want string
				res  Result
			}{k + " after " + pos, []string{pos, k}, map[string]string{"HOME": "|abc", "LEFT": "ab|c", "END": "abc|"}[pos], None})
		}
	}
	for _, tt := range tests {
		l := typed(t, 10, "a", "b", "c")
		var res Result
		for _, k := range tt.keys {
			res = l.Key(k, false)
		}
		if got := state(l); got != tt.want || res != tt.res {
			t.Errorf("%s: %q, %d; want %q, %d", tt.name, got, res, tt.want, tt.res)
		}
	}
}

// TestMax covers the length limit M at the start, in the middle and at the end.
func TestMax(t *testing.T) {
	for _, pos := range []string{"HOME", "LEFT", "END"} {
		l := typed(t, 3, "a", "b", "c", pos)
		before := state(l)
		if res := l.Key("d", false); res != None || state(l) != before {
			t.Errorf("d after %s in a full line: %q, %d; want %q, None", pos, state(l), res, before)
		}
	}
	l := typed(t, 1, "a", "b")
	if got := l.Text(); got != "a" {
		t.Errorf("M = 1: line %q, want a", got)
	}
}

// TestChars covers the characters of section 5.
func TestChars(t *testing.T) {
	tests := []struct {
		name   string
		before []string
		k      string
		want   string
	}{
		{"control character", nil, "\x07", ""},
		{"C1 control character", nil, "\u0085", ""},
		{"bidi control", nil, "\u202e", ""},
		{"wide character", nil, "中", ""},
		{"emoji", nil, "😀", ""},
		{"zero width space", nil, "\u200b", ""},
		{"combining mark after a letter", []string{"a"}, "\u0301", "a"},
		{"variation selector", []string{"a"}, "\ufe0f", "a"},
		{"U+0903 after a letter", []string{"\u0915"}, "\u0903", "\u0915"},
		{"U+0903 at the start", nil, "\u0903", ""},
		{"U+0903 at the start after a letter", []string{"\u0915", "HOME"}, "\u0903", "\u0915"},
		{"Prepend U+0D4E at the end", []string{"a"}, "\u0d4e", "a"},
		{"Prepend U+0D4E in the middle", []string{"a", "HOME"}, "\u0d4e", "a"},
		{"valid U+FFFD", []string{"a"}, "\ufffd", "a\ufffd"},
		{"East Asian Ambiguous", []string{"a"}, "\u00b1", "a\u00b1"},
		{"Cyrillic letter", []string{"a"}, "ы", "aы"},
	}
	for _, tt := range tests {
		l := typed(t, 10, tt.before...)
		l.Key(tt.k, false)
		if got := l.Text(); got != tt.want {
			t.Errorf("%s: line %q, want %q", tt.name, got, tt.want)
		}
	}
}

// TestRemoveJoin covers BS and DEL in U+1161 x U+1161: the removal of x joins the two U+1161.
func TestRemoveJoin(t *testing.T) {
	l := typed(t, 10, "\u1161", "x", "\u1161")
	if got := l.Text(); got != "\u1161x\u1161" {
		t.Fatalf("line %q, want \u1161x\u1161", got)
	}
	l.Key("LEFT", false)
	if res := l.Key("BS", false); res != None || state(l) != "\u1161x|\u1161" {
		t.Errorf("BS: %q, %d; want no change", state(l), res)
	}
	l.Key("LEFT", false)
	if res := l.Key("DEL", false); res != None || state(l) != "\u1161|x\u1161" {
		t.Errorf("DEL: %q, %d; want no change", state(l), res)
	}
}

// TestMasked covers the masked mode: the draw shows *, the caller gets the line.
func TestMasked(t *testing.T) {
	l := New(Field{Row: 2, Col: 3, Max: 5, Masked: true}, 80, 24)
	for _, k := range []string{"p", "ы", "SPACE", "1", "LEFT"} {
		l.Key(k, false)
	}
	d := l.Draw()
	if d.Cells != "****  " || d.Cursor != 3 {
		t.Errorf("draw %q, cursor %d; want \"****  \", 3", d.Cells, d.Cursor)
	}
	if got := l.Text(); got != "pы 1" {
		t.Errorf("line %q, want \"pы 1\"", got)
	}
}

// TestDraw covers the cells of a draw: the line, then spaces up to M+1 cells.
func TestDraw(t *testing.T) {
	f := Field{Row: 2, Col: 3, Max: 4, Style: screen.Style{Attr: screen.Bold}}
	l := New(f, 80, 24)
	if d := l.Draw(); d != (Draw{Field: f, Cells: "     ", Cursor: 0}) {
		t.Errorf("empty draw %+v", d)
	}
	for _, k := range []string{"a", "ы", "c", "d"} {
		l.Key(k, false)
	}
	if d := l.Draw(); d.Cells != "aыcd " || d.Cursor != 4 {
		t.Errorf("full draw %q, cursor %d; want \"aыcd \", 4", d.Cells, d.Cursor)
	}
}

// TestEmpty covers an empty ENTER and ESC.
func TestEmpty(t *testing.T) {
	l := typed(t, 5)
	if res := l.Key("ENTER", false); res != Enter || l.Text() != "" {
		t.Errorf("ENTER: %d, %q; want Enter and an empty line", res, l.Text())
	}
	if res := typed(t, 5).Key("ESC", false); res != Cancel {
		t.Errorf("ESC: %d, want Cancel", res)
	}
}

// TestPaste covers the keys of a paste (section 6).
func TestPaste(t *testing.T) {
	l := typed(t, 10, "a", "b")
	for _, k := range []string{"c", "SPACE", "ESC", "BS", "DEL", "LEFT", "HOME", "TAB"} {
		l.Key(k, true)
	}
	if got := state(l); got != "abc |" {
		t.Errorf("line after a paste %q, want \"abc |\"", got)
	}
	if res := l.Key("ENTER", true); res != Enter {
		t.Errorf("ENTER in a paste: %d, want Enter", res)
	}
}

// TestFit covers the panic of a field that does not fit the screen (section 3).
func TestFit(t *testing.T) {
	fits := []Field{{Row: 24, Col: 1, Max: 79}, {Row: 1, Col: 70, Max: 10}}
	for _, f := range fits {
		New(f, 80, 24)
	}
	bad := []Field{
		{Row: 25, Col: 1, Max: 10}, {Row: 1, Col: 1, Max: 80}, {Row: 1, Col: 71, Max: 10},
		{Row: 0, Col: 1, Max: 1}, {Row: 1, Col: 0, Max: 1}, {Row: 1, Col: 1, Max: 0},
	}
	for _, f := range bad {
		func() {
			defer func() {
				if v := recover(); v == nil || !strings.Contains(v.(string), "does not fit") {
					t.Errorf("New(%+v) recover %v, want a panic", f, v)
				}
			}()
			New(f, 80, 24)
		}()
	}
}
