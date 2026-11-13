package key

import "testing"

func TestParse(t *testing.T) {
	valid := []string{
		"ESC", "ENTER", "SPACE", "TAB", "BS", "DEL", "UP", "DOWN", "LEFT", "RIGHT",
		"HOME", "END", "PGUP", "PGDN",
		"F1", "F2", "F3", "F4", "F5", "F6", "F7", "F8", "F9", "F10", "F11", "F12",
		"ANY",
	}
	for c := 'A'; c <= 'Z'; c++ {
		valid = append(valid, string(c))
	}
	for c := '0'; c <= '9'; c++ {
		valid = append(valid, string(c))
	}
	for _, k := range valid {
		if got, ok := Parse(k); got != k || !ok {
			t.Errorf("Parse(%q) = %q, %v; want %q, true", k, got, ok, k)
		}
	}

	// The case rule.
	for in, want := range map[string]string{"s": "S", "esc": "ESC", "Esc": "ESC", "f12": "F12", "any": "ANY", "pgDn": "PGDN"} {
		if got, ok := Parse(in); got != want || !ok {
			t.Errorf("Parse(%q) = %q, %v; want %q, true", in, got, ok, want)
		}
	}

	// A Cyrillic letter is not a key name: the map of section 3.2 is only for the input.
	for _, k := range []string{"", "?", " ", "AB", "F0", "F13", "F01", "ESCAPE", "ы", "Ы", "ESC ", "_"} {
		if _, ok := Parse(k); ok {
			t.Errorf("Parse(%q): ok, want invalid", k)
		}
	}
}

func TestUpper(t *testing.T) {
	for in, want := range map[string]string{"": "", "abc": "ABC", "F1": "F1", "ы": "ы", "?": "?", "aЫz": "AЫZ"} {
		if got := Upper(in); got != want {
			t.Errorf("Upper(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLatin(t *testing.T) {
	// The table of section 3.2, row by row.
	rows := []struct{ cyr, lat string }{
		{"йцукенгшщз", "QWERTYUIOP"},
		{"фывапролд", "ASDFGHJKL"},
		{"ячсмить", "ZXCVBNM"},
	}
	for _, row := range rows {
		cyr, lat := []rune(row.cyr), []rune(row.lat)
		for i, c := range cyr {
			for _, r := range []rune{c, c - 'а' + 'А'} {
				if got, ok := Latin(r); got != lat[i] || !ok {
					t.Errorf("Latin(%q) = %q, %v; want %q, true", r, got, ok, lat[i])
				}
			}
		}
	}

	// The letters without a Latin letter, and other runes.
	for _, r := range "хъжэёбюХЪЖЭЁБЮaZ1?ї" {
		if got, ok := Latin(r); ok {
			t.Errorf("Latin(%q) = %q, true; want false", r, got)
		}
	}
}
