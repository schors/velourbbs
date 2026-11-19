// Package key holds the key names and the case rule of bbt-format, section 8.1,
// and the Russian layout map of menu-nav, section 3.2.
// A section without a specification name is a section of menu-nav.
package key

import (
	"strconv"
	"strings"
)

// Any is the key that matches each key without an enabled item of its own.
const Any = "ANY"

// The keys of the cursor of a list screen with select (list-view, section 8).
const (
	Up    = "UP"
	Down  = "DOWN"
	Enter = "ENTER"
)

// IsCursor reports whether k is a key of the cursor.
func IsCursor(k string) bool { return k == Up || k == Down || k == Enter }

// named are the named keys of bbt-format, section 8.1, without F1–F12.
var named = []string{
	"ESC", "ENTER", "SPACE", "TAB", "BS", "DEL", "UP", "DOWN", "LEFT", "RIGHT",
	"HOME", "END", "PGUP", "PGDN",
}

// Upper applies the case rule: it gives s with the ASCII letters in upper case.
func Upper(s string) string {
	return strings.Map(func(r rune) rune {
		if 'a' <= r && r <= 'z' {
			return r - 'a' + 'A'
		}
		return r
	}, s)
}

// Parse gives the key of an item after the case rule.
// ok is false when s is not a key name of bbt-format, section 8.1, or ANY.
func Parse(s string) (k string, ok bool) {
	k = Upper(s)
	if len(k) == 1 {
		c := k[0]
		return k, 'A' <= c && c <= 'Z' || '0' <= c && c <= '9'
	}
	if n, f := strings.CutPrefix(k, "F"); f {
		for i := 1; i <= 12; i++ {
			if n == strconv.Itoa(i) {
				return k, true
			}
		}
	}
	for _, name := range named {
		if k == name {
			return k, true
		}
	}
	return k, k == Any
}

// latin maps the lower-case Cyrillic letters of section 3.2 to their Latin letters.
var latin = map[rune]rune{
	'й': 'Q', 'ц': 'W', 'у': 'E', 'к': 'R', 'е': 'T', 'н': 'Y', 'г': 'U', 'ш': 'I', 'щ': 'O', 'з': 'P',
	'ф': 'A', 'ы': 'S', 'в': 'D', 'а': 'F', 'п': 'G', 'р': 'H', 'о': 'J', 'л': 'K', 'д': 'L',
	'я': 'Z', 'ч': 'X', 'с': 'C', 'м': 'V', 'и': 'B', 'т': 'N', 'ь': 'M',
}

// Latin gives the upper-case Latin letter on the key of the Cyrillic letter r
// in the ЙЦУКЕН layout, for both cases (section 3.2).
// ok is false for a letter without a Latin letter and for each other rune.
func Latin(r rune) (l rune, ok bool) {
	if 'А' <= r && r <= 'Я' {
		r += 'а' - 'А'
	}
	l, ok = latin[r]
	return l, ok
}

// Map gives the key of the menu match for a key from the terminal (section 3.3, step 1):
// a Cyrillic letter gives its Latin letter (section 3.2), and the case rule applies.
// k is a key name of bbt-format, section 8.1, or one character.
func Map(k string) string {
	if r := []rune(k); len(r) == 1 {
		if l, ok := Latin(r[0]); ok {
			return string(l)
		}
	}
	return Upper(k)
}
