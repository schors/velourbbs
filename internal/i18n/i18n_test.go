package i18n

import (
	"maps"
	"strings"
	"testing"
	"testing/fstest"
)

// A test reads each dictionary of the binary (session-language, section 7.1).
func TestDictionaries(t *testing.T) {
	if _, err := load(files); err != nil {
		t.Fatal(err)
	}
}

// The host shows the translation of the session language, else the English text (session-language, section 7).
func TestText(t *testing.T) {
	for _, tt := range []struct{ lang, want string }{
		{"ru", "-- Дальше --"},
		{"lt", "-- Toliau --"},
		{"en", "-- More --"},
		{"de", "-- More --"},
	} {
		if got := Text(tt.lang, More); got != tt.want {
			t.Errorf("Text(%q, More) = %q, want %q", tt.lang, got, tt.want)
		}
	}
}

func TestLoadName(t *testing.T) {
	_, err := load(fstest.MapFS{"po/RU.po": {Data: []byte("")}})
	if want := `po/RU.po: "RU" is not a language code`; err == nil || err.Error() != want {
		t.Errorf("err = %v, want %s", err, want)
	}
}

// lines joins the lines of a dictionary.
func lines(ls ...string) []byte {
	return []byte(strings.Join(ls, "\n"))
}

// The rules of the dictionary format (session-language, section 7.1).
func TestParse(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want map[string]string // Nil is no translation.
	}{
		{"empty file", nil, nil},
		{"entry", lines(`msgid "-- More --"`, `msgstr "-- Дальше --"`), map[string]string{More: "-- Дальше --"}},
		{"header, empty lines and comments", lines(
			`# Russian translation.`,
			`#, fuzzy`,
			`msgid ""`,
			`msgstr ""`,
			`"Content-Type: text/plain; charset=UTF-8\n"`,
			``,
			`#. A comment for translators.`,
			`#: internal/config/config.go:91`,
			`#| msgid "-- more --"`,
			`#, c-format`,
			`msgid "-- More --"`,
			`msgstr "[дальше]"`,
			``,
			`#~ msgid "-- Old --"`,
			`#~ msgstr "-- Старое --"`,
			``,
		), map[string]string{More: "[дальше]"}},
		{"continued strings", lines(
			`msgid ""`,
			`"-- More"`,
			`" --"`,
			`msgstr "-- "`,
			`"Дальше"`,
			`" --"`,
		), map[string]string{More: "-- Дальше --"}},
		{"escapes", lines(`msgid "-- More --"`, `msgstr "\"\\\""`), map[string]string{More: `"\"`}},
		{"spaces and CRLF", lines("  msgid\t\"-- More --\"  \r", "msgstr\"[x]\"\r"), map[string]string{More: "[x]"}},
		{"fuzzy", lines(`#, fuzzy, c-format`, `msgid "-- More --"`, `msgstr "[x]"`), nil},
		{"fuzzy with spaces", lines(`#,c-format , fuzzy `, `msgid "-- More --"`, `msgstr "[x]"`), nil},
		{"fuzzy of the entry before", lines(
			`#, fuzzy`, `msgid ""`, `msgstr ""`, ``, `msgid "-- More --"`, `msgstr "[x]"`,
		), map[string]string{More: "[x]"}},
		{"empty msgstr", lines(`msgid "-- More --"`, `msgstr ""`), nil},
		{"fuzzy breaks no rule", lines(`#, fuzzy`, `msgid "-- More --"`, `msgstr "\t"`), nil},
		{"fuzzy of an obsolete entry", lines(
			`#, fuzzy`, `#~| msgid "-- Older --"`, `#~ msgid "-- Old --"`, `#~ msgstr "x"`, ``,
			`msgid "-- More --"`, `msgstr "[y]"`,
		), map[string]string{More: "[y]"}},
	}
	for _, tt := range tests {
		got, err := Parse("ru.po", tt.data)
		if err != nil {
			t.Errorf("%s: %v", tt.name, err)
			continue
		}
		if tt.want == nil {
			tt.want = map[string]string{}
		}
		if !maps.Equal(got, tt.want) {
			t.Errorf("%s: got %q, want %q", tt.name, got, tt.want)
		}
	}
}

// The errors of the dictionary format (session-language, section 7.1).
func TestParseErrors(t *testing.T) {
	entry := `msgid "-- More --"`
	tests := []struct {
		data []byte
		want string
	}{
		{lines(`msgctxt "menu"`, entry, `msgstr ""`), `ru.po:1: keyword "msgctxt" is not supported`},
		{lines(entry, `msgid_plural "-- More --"`), `ru.po:2: keyword "msgid_plural" is not supported`},
		{lines(entry, `msgstr[0] ""`), `ru.po:2: keyword "msgstr[0]" is not supported`},
		{lines(entry, `msgstr ""`, `= x`), `ru.po:3: invalid line`},
		{lines(`msgid -- More --`), `ru.po:1: no string`},
		{lines(`msgid "-- More --`), `ru.po:1: string without end quote`},
		{lines(`msgid "-- More --\"`), `ru.po:1: string without end quote`},
		{lines(`msgid "-- More --\`), `ru.po:1: string without end quote`},
		{lines(`msgid "-- More --" "x"`), `ru.po:1: text after the string`},
		{lines(`msgid "-- More --" # x`), `ru.po:1: text after the string`},
		{lines(`"-- More --"`), `ru.po:1: continued string with no string before it`},
		{lines(entry, `# x`, `" --"`, `msgstr ""`), `ru.po:3: continued string with no string before it`},
		{lines(entry, ``, `" --"`, `msgstr ""`), `ru.po:3: continued string with no string before it`},
		{lines(entry), `ru.po:1: msgid without msgstr`},
		{lines(``, entry, `msgid ""`, `msgstr ""`), `ru.po:2: msgid without msgstr`},
		{lines(`msgstr ""`), `ru.po:1: msgstr without msgid`},
		{lines(entry, `msgstr ""`, `msgstr ""`), `ru.po:3: msgstr without msgid`},
		{lines(entry, `msgstr "\x41"`), `ru.po:2: escape \x is not supported`},
		{lines(entry, `msgstr "\Д"`), `ru.po:2: escape \Д is not supported`},
		{lines(entry, "msgstr \"\xff\""), `ru.po:2: invalid UTF-8`},
		{lines(`# `+"\xc0", entry, `msgstr ""`), `ru.po:1: invalid UTF-8`},
		{lines(entry, `msgstr ""`, entry, `msgstr "[x]"`), `ru.po:3: duplicate msgid "-- More --"`},
		{lines(`msgid ""`, `msgstr ""`, `msgid ""`, `msgstr ""`), `ru.po:3: duplicate msgid ""`},
		{lines(`msgid "-- Next --"`, `msgstr "[x]"`), `ru.po:1: msgid "-- Next --" is not an engine text`},
		{lines(`#, fuzzy`, `msgid "-- Next --"`, `msgstr ""`), `ru.po:2: msgid "-- Next --" is not an engine text`},
		{lines(entry, `msgstr "[x]\n"`), `ru.po:2: msgstr: control character U+000A`},
		{lines(entry, `msgstr ""`, `"\t"`), `ru.po:2: msgstr: control character U+0009`},
		{lines(entry, "msgstr \"a\u200b\""), `ru.po:2: msgstr: zero-width grapheme U+200B`},
		{lines(entry, `msgstr "e`+strings.Repeat("\u0301", 16)+`"`), `ru.po:2: msgstr: grapheme over 32 bytes: 33 bytes`},
	}
	for _, tt := range tests {
		got, err := Parse("ru.po", tt.data)
		if err == nil || err.Error() != tt.want {
			t.Errorf("%q: got %q, %v; want %s", tt.data, got, err, tt.want)
		}
	}
}
