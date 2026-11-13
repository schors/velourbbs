package screen

import (
	"fmt"
	"strings"
	"testing"
)

func TestLookupField(t *testing.T) {
	fields := Fields{"w": {term: termCols}}
	tests := []struct {
		path, want, err string
	}{
		{path: "term.type", want: "term.type"},
		{path: "term.cols", want: "term.cols"},
		{path: "term.rows", want: "term.rows"},
		{path: "term.charset", want: "term.charset"},
		{path: "term.colors", want: "term.colors"},
		{path: "w", want: "term.cols"},
		{path: "term.colours", err: `unknown field "term.colours"`},
		{path: "x", err: `unknown field "x"`},
		{path: "term", err: `unknown field "term"`},
		{path: "user.name", err: `unknown field "user.name"`},
		{path: "Term.cols", err: `invalid field path "Term.cols"`},
		{path: "term.", err: `invalid field path "term."`},
		{path: "term.a.b", err: `invalid field path "term.a.b"`},
		{path: "", err: `invalid field path ""`},
		{path: "item.S", err: `field "item.S" is valid only in {#if}`},
	}
	for _, tt := range tests {
		r, err := LookupField(tt.path, fields)
		switch {
		case tt.err != "":
			if err == nil || err.Error() != tt.err {
				t.Errorf("%q: got error %v, want %q", tt.path, err, tt.err)
			}
		case err != nil:
			t.Errorf("%q: %v", tt.path, err)
		case r.String() != tt.want:
			t.Errorf("%q: got %s, want %s", tt.path, r, tt.want)
		}
	}
}

func TestNewAlias(t *testing.T) {
	tests := []struct {
		name, path, err string
	}{
		{name: "w", path: "term.cols"},
		{name: "t_2", path: "term.type"},
		{name: "W", path: "term.cols", err: `invalid alias "W"`},
		{name: "2w", path: "term.cols", err: `invalid alias "2w"`},
		{name: "term", path: "term.cols", err: `alias "term" is a reserved word`},
		{name: "item", path: "term.cols", err: `alias "item" is a reserved word`},
		{name: "and", path: "term.cols", err: `alias "and" is a reserved word`},
		{name: "or", path: "term.cols", err: `alias "or" is a reserved word`},
		{name: "not", path: "term.cols", err: `alias "not" is a reserved word`},
		{name: "w", path: "x", err: `alias "w": "x" is not a registry path`},
		{name: "w", path: "term.colours", err: `alias "w": unknown field "term.colours"`},
		{name: "w", path: "item.S", err: `alias "w": "item.S" is valid only in {#if}`},
	}
	for _, tt := range tests {
		r, err := NewAlias(tt.name, tt.path)
		switch {
		case tt.err != "":
			if err == nil || err.Error() != tt.err {
				t.Errorf("%s = %q: got error %v, want %q", tt.name, tt.path, err, tt.err)
			}
		case err != nil:
			t.Errorf("%s = %q: %v", tt.name, tt.path, err)
		case r.String() != tt.path:
			t.Errorf("%s = %q: got %s", tt.name, tt.path, r)
		}
	}
}

func TestEnvText(t *testing.T) {
	env := &Env{Term: Term{Type: "xterm\x1b[31m", Cols: 132, Rows: -1, Charset: "utf-8", Colors: 256}}
	tests := []struct {
		path, want string
	}{
		{"term.type", "xterm[31m"},
		{"term.cols", "132"},
		{"term.rows", "-1"},
		{"term.charset", "utf-8"},
		{"term.colors", "256"},
	}
	for _, tt := range tests {
		r, err := LookupField(tt.path, nil)
		if err != nil {
			t.Fatal(err)
		}
		if got := text(env.Text(r)); got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.path, got, tt.want)
		}
	}
}

// text joins the graphemes of gs.
func text(gs []Grapheme) string {
	var b strings.Builder
	for _, g := range gs {
		b.WriteString(g.Text)
	}
	return b.String()
}

func TestTermCheck(t *testing.T) {
	ok := Term{Cols: 80, Rows: 25, Colors: 16}
	tests := []struct {
		change func(*Term)
		err    string
	}{
		{func(*Term) {}, ""},
		{func(t *Term) { t.Cols, t.Rows, t.Colors = 20, 10, 256 }, ""},
		{func(t *Term) { t.Cols, t.Rows, t.Colors = 255, 100, 16777216 }, ""},
		{func(t *Term) { t.Cols = 19 }, "term.cols is 19, must be from 20 to 255"},
		{func(t *Term) { t.Cols = 256 }, "term.cols is 256, must be from 20 to 255"},
		{func(t *Term) { t.Rows = 9 }, "term.rows is 9, must be from 10 to 100"},
		{func(t *Term) { t.Rows = 101 }, "term.rows is 101, must be from 10 to 100"},
		{func(t *Term) { t.Colors = 8 }, "term.colors is 8, must be 16, 256 or 16777216"},
	}
	for _, tt := range tests {
		term := ok
		tt.change(&term)
		err := term.Check()
		if got := fmt.Sprint(err); err == nil && tt.err != "" || err != nil && got != tt.err {
			t.Errorf("%+v: error %v, want %q", term, err, tt.err)
		}
	}
}
