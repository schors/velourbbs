package screen

import (
	"slices"
	"testing"
)

func TestCondEval(t *testing.T) {
	fields := Fields{"w": {term: termCols}, "t": {term: termType}}
	env := &Env{
		Term:  Term{Type: "ansi", Cols: 80, Rows: 25, Charset: "utf-8", Colors: 16},
		Items: map[string]bool{"A": true, "B": true},
		Args:  map[string]string{"file": "gold.md"},
	}
	tests := []struct {
		cond string
		want bool
	}{
		{"term.cols == 80", true},
		{"term.cols == 81", false},
		{"term.cols != 81", true},
		{"term.cols != 80", false},
		{"term.cols < 81", true},
		{"term.cols < 80", false},
		{"term.cols <= 80", true},
		{"term.cols <= 79", false},
		{"term.cols > 79", true},
		{"term.cols > 80", false},
		{"term.cols >= 80", true},
		{"term.cols >= 81", false},
		{"term.cols > -1", true},
		{"term.cols>=80", true},
		{"w==80", true},
		{`term.type == "ansi"`, true},
		{`term.type == "vt100"`, false},
		{`term.type != "vt100"`, true},
		{`term.type != "ansi"`, false},
		{`t=="ansi"and w==80`, true},
		{`term.charset == ""`, false},
		{"item.A", true},
		{"item.a", true}, // the case rule
		{"item.C", false},
		{"not item.A", false},
		{"not item.C", true},
		{"item.A and item.C", false},
		{"item.C or item.A", true},
		{"item.C or item.A and item.C", false}, // and before or: C or (A and C)
		{"item.A or item.C and item.C", true},  // A or (C and C)
		{"item.A and item.C or item.B", true},  // (A and C) or B
		{"item.C and item.A or item.C and item.B", false},
		{"  item.A  and  not  item.C  ", true},
		{`arg.file == "gold.md"`, true},
		{`arg.file != "gold.md"`, false},
		{`arg.other == ""`, true},
	}
	for _, tt := range tests {
		c, err := ParseCond(tt.cond, fields, true)
		if err != nil {
			t.Errorf("%q: %v", tt.cond, err)
			continue
		}
		if got := c.Eval(env); got != tt.want {
			t.Errorf("%q: got %v, want %v", tt.cond, got, tt.want)
		}
	}
}

func TestCondErrors(t *testing.T) {
	fields := Fields{"w": {term: termCols}}
	tests := []struct {
		cond, err string
	}{
		// Type errors of section 5.
		{"term.cols", `integer field "term.cols" needs an operator`},
		{"term.type", `string field "term.type" needs an operator`},
		{"not term.cols", `"not" needs a boolean field, "term.cols" is integer`},
		{"item.A == 1", `boolean field "item.A" takes no operator`},
		{`term.cols == "80"`, `integer field "term.cols" takes an integer, found a string`},
		{"term.type == 1", `string field "term.type" takes a string, found an integer`},
		{`term.type < "a"`, `string field "term.type" takes only == or !=`},
		{`term.type >= "a"`, `string field "term.type" takes only == or !=`},
		// Paths.
		{"term.colours == 1", `unknown field "term.colours"`},
		{"x", `unknown field "x"`},
		{"notitem.A", `unknown field "notitem.A"`},
		{"Term.cols == 1", `invalid field path "Term.cols"`},
		// Syntax.
		{"", "expected a field, found end of condition"},
		{"   ", "expected a field, found end of condition"},
		{"and", `expected a field, found "and"`},
		{"not", "expected a field, found end of condition"},
		{"not not item.A", `expected a field, found "not"`},
		{"item.A and", "expected a field, found end of condition"},
		{"item.A or or item.B", `expected a field, found "or"`},
		{"item.A item.B", `expected "and" or "or", found "item.B"`},
		{"item.A = 1", `expected "and" or "or", found "= 1"`},
		{"term.cols =1", `integer field "term.cols" needs an operator`},
		{"term.cols ==", "expected an integer or a string, found end of condition"},
		{"term.cols == - 1", `expected an integer or a string, found "- 1"`},
		{"term.cols == 1x", `expected "and" or "or", found "x"`},
		{"term.cols == 99999999999999999999", "integer 99999999999999999999 is out of range"},
		{`term.type == "ansi`, `string without closing '"'`},
		{`term.type == "a}b"`, `'}' in a string`},
		{"(item.A)", `expected a field, found "(item.A)"`},
	}
	for _, tt := range tests {
		_, err := ParseCond(tt.cond, fields, true)
		if err == nil || err.Error() != tt.err {
			t.Errorf("%q: got error %v, want %q", tt.cond, err, tt.err)
		}
	}
}

func TestCondItems(t *testing.T) {
	// when: item.<KEY> is valid only in {#if}.
	if _, err := ParseCond("item.A", nil, false); err == nil || err.Error() != `field "item.A" is valid only in {#if}` {
		t.Errorf("when with item.A: got error %v", err)
	}
	if _, err := ParseCond("term.cols > 1", nil, false); err != nil {
		t.Errorf("when: %v", err)
	}

	// A screen checks its conditions against its own items.
	c, err := ParseCond("item.A or item.s", nil, true)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		keys map[string]bool
		want []string
	}{
		// The case rule: item.s is item.S.
		{"all keys", map[string]bool{"A": true, "S": true}, nil},
		{"one key", map[string]bool{"A": true}, []string{`unknown field "item.S"`}},
		{"no items", nil, []string{`unknown field "item.A"`, `unknown field "item.S"`}},
	}
	for _, tt := range tests {
		var got []string
		for _, err := range c.CheckItems(tt.keys) {
			got = append(got, err.Error())
		}
		if !slices.Equal(got, tt.want) {
			t.Errorf("%s: got %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestCondArgs(t *testing.T) {
	c, err := ParseCond(`arg.a == "x" and term.cols > 1 or f != "" and item.A`, Fields{"f": {arg: "b"}}, true)
	if err != nil {
		t.Fatal(err)
	}
	if got := c.Args(); !slices.Equal(got, []string{"a", "b"}) {
		t.Errorf("got %q", got)
	}
}
