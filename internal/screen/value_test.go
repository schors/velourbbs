package screen

import (
	"slices"
	"strings"
	"testing"
)

func TestCleanValue(t *testing.T) {
	long := "e" + strings.Repeat("\u0301", 16) // 33 bytes, one grapheme
	tests := []struct {
		name, value string
		want        []Grapheme
	}{
		{"plain", "ab", []Grapheme{{"a", 1}, {"b", 1}}},
		{"wide", "中", []Grapheme{{"中", 2}}},
		{"control", "a\tb\x7f\u202e", []Grapheme{{"a", 1}, {"b", 1}}},
		{"zero width", "\u0301a\u200bb", []Grapheme{{"a", 1}, {"b", 1}}},
		{"long zero width", "a\x01" + strings.Repeat("\u0301", 17) + "b", []Grapheme{{"a", 1}, {"b", 1}}},
		{"long grapheme", "a" + long + "b", []Grapheme{{"a", 1}, {"\uFFFD", 1}, {"b", 1}}},
		{"invalid UTF-8", "a\xff\xfeb", []Grapheme{{"a", 1}, {"\uFFFD", 1}, {"\uFFFD", 1}, {"b", 1}}},
		{"empty", "", nil},
	}
	for _, tt := range tests {
		if got := CleanValue(tt.value); !slices.Equal(got, tt.want) {
			t.Errorf("%s: got %+v, want %+v", tt.name, got, tt.want)
		}
	}
}

func TestFitValue(t *testing.T) {
	tests := []struct {
		name, value string
		width       int
		right       bool
		want        string
	}{
		{"pad left aligned", "ab", 5, false, "ab   "},
		{"pad right aligned", "ab", 5, true, "   ab"},
		{"exact", "abc", 3, false, "abc"},
		{"cut left aligned", "abcdef", 4, false, "abcd"},
		{"cut right aligned", "abcdef", 4, true, "abcd"},
		{"wide at edge", "ab中", 3, false, "ab "},
		{"wide at edge right aligned", "ab中", 3, true, "ab "},
		{"wide fits", "a中b", 3, true, "a中"},
		{"wide pad right aligned", "中", 3, true, " 中"},
		{"combining", "e\u0301xy", 2, false, "e\u0301x"},
		{"empty", "", 3, true, "   "},
	}
	for _, tt := range tests {
		got := FitValue(CleanValue(tt.value), tt.width, tt.right)
		width := 0
		for _, g := range got {
			width += g.Width
		}
		if text(got) != tt.want || width != tt.width {
			t.Errorf("%s: got %q of width %d, want %q of width %d", tt.name, text(got), width, tt.want, tt.width)
		}
	}
}
