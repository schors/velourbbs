package screen

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestGraphemes(t *testing.T) {
	type gw struct {
		g string
		w int
	}
	tests := []struct {
		name string
		text string
		want []gw
	}{
		{"ascii", "ab", []gw{{"a", 1}, {"b", 1}}},
		{"wide", "中a", []gw{{"中", 2}, {"a", 1}}},
		{"combining", "e\u0301\u0302x", []gw{{"e\u0301\u0302", 1}, {"x", 1}}},
		{"emoji modifier", "👍🏽", []gw{{"👍🏽", 2}}},
		{"flag", "🇷🇺", []gw{{"🇷🇺", 2}}},
		{"ambiguous is narrow", "§", []gw{{"§", 1}}},
	}
	for _, tt := range tests {
		var got []gw
		for g, w := range Graphemes(tt.text) {
			got = append(got, gw{g, w})
		}
		if !slices.Equal(got, tt.want) {
			t.Errorf("%s: got %+v, want %+v", tt.name, got, tt.want)
		}
	}
}

func TestCheckGrapheme(t *testing.T) {
	long := "e" + strings.Repeat("\u0301", 16) // 33 bytes, one grapheme
	tests := []struct {
		text string
		want error
	}{
		{"a", nil},
		{"中", nil},
		{"e\u0301", nil},
		{"e" + strings.Repeat("\u0301", 15), nil}, // 31 bytes
		{long, ErrLongGrapheme},
		{"\t", ErrControl},
		{"\x7f", ErrControl},
		{"\u0085", ErrControl},
		{"\u202e", ErrControl},
		{"\u2066", ErrControl},
		{"\u200b", ErrZeroWidth},
		{"\u00ad", ErrZeroWidth},
		{"\u0301", ErrZeroWidth},
	}
	for _, tt := range tests {
		n := 0
		for g, w := range Graphemes(tt.text) {
			n++
			if err := CheckGrapheme(g, w); !errors.Is(err, tt.want) {
				t.Errorf("CheckGrapheme(%+q) = %v, want %v", g, err, tt.want)
			}
		}
		if n != 1 {
			t.Errorf("%+q has %d graphemes, want 1", tt.text, n)
		}
	}
}

func TestCheckGraphemeMessage(t *testing.T) {
	if got, want := CheckGrapheme("\t", 0).Error(), "control character U+0009"; got != want {
		t.Errorf("message = %q, want %q", got, want)
	}
}

func TestIsControl(t *testing.T) {
	for _, r := range []rune{0, 0x1f, 0x7f, 0x9f, 0x202a, 0x202e, 0x2066, 0x2069} {
		if !IsControl(r) {
			t.Errorf("IsControl(%U) = false", r)
		}
	}
	for _, r := range []rune{' ', '~', 0xa0, 0x2029, 0x202f, 0x2065, 0x206a} {
		if IsControl(r) {
			t.Errorf("IsControl(%U) = true", r)
		}
	}
}

func TestIsIdent(t *testing.T) {
	for _, s := range []string{"a", "normal", "a_1", "z9"} {
		if !IsIdent(s) {
			t.Errorf("IsIdent(%q) = false", s)
		}
	}
	for _, s := range []string{"", "1a", "_a", "A", "aB", "a-b", "a.b", "é"} {
		if IsIdent(s) {
			t.Errorf("IsIdent(%q) = true", s)
		}
	}
}

func TestWidthLimits(t *testing.T) {
	for _, w := range []WidthLimits{{1, 1}, {40, 80}, {255, 255}} {
		if err := w.Check(); err != nil {
			t.Errorf("%v: %v", w, err)
		}
	}
	for _, w := range []WidthLimits{{0, 80}, {81, 80}, {40, 256}} {
		if err := w.Check(); err == nil {
			t.Errorf("%v: want an error", w)
		}
	}
}
