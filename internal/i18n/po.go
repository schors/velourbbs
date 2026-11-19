package i18n

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// entry is a msgid with its msgstr.
type entry struct {
	id, str         string
	idLine, strLine int  // The lines of msgid and msgstr; strLine is 0 before msgstr.
	fuzzy           bool // The flag fuzzy in the #, comment before msgid.
}

// parser reads a dictionary line by line.
type parser struct {
	name  string
	ln    int // The number of the current line.
	dict  map[string]string
	seen  map[string]bool // Each msgid.
	e     *entry          // The current entry, else nil.
	fuzzy bool            // The flag fuzzy of the next entry.
	cont  *string         // The string that a continued string continues, else nil.
}

// Parse reads the dictionary data of the file name: the PO subset of session-language, section 7.1.
// It gives the translations of the engine texts. An error names the file and the line.
func Parse(name string, data []byte) (map[string]string, error) {
	p := parser{name: name, dict: map[string]string{}, seen: map[string]bool{}}
	for i, line := range strings.Split(string(data), "\n") {
		p.ln = i + 1
		if err := p.line(strings.Trim(line, " \t\r")); err != nil {
			return nil, err
		}
	}
	if err := p.end(); err != nil {
		return nil, err
	}
	return p.dict, nil
}

// line reads the current line s without the spaces at its ends.
func (p *parser) line(s string) error {
	if !utf8.ValidString(s) {
		return p.errorf(p.ln, "invalid UTF-8")
	}
	switch {
	case s == "":
		p.cont = nil
	case s[0] == '#':
		if flags, ok := strings.CutPrefix(s, "#,"); ok {
			for f := range strings.SplitSeq(flags, ",") {
				p.fuzzy = p.fuzzy || strings.TrimSpace(f) == "fuzzy"
			}
		}
		p.cont = nil
	case s[0] == '"':
		if p.cont == nil {
			return p.errorf(p.ln, "continued string with no string before it")
		}
		t, err := p.unquote(s)
		if err != nil {
			return err
		}
		*p.cont += t
	default:
		i := strings.IndexAny(s, " \t\"")
		if i < 0 {
			i = len(s)
		}
		kw, rest := s[:i], strings.TrimLeft(s[i:], " \t")
		switch kw {
		case "msgid":
			return p.msgid(rest)
		case "msgstr":
			return p.msgstr(rest)
		}
		if 'a' <= s[0] && s[0] <= 'z' {
			return p.errorf(p.ln, "keyword %q is not supported", kw)
		}
		return p.errorf(p.ln, "invalid line")
	}
	return nil
}

// msgid ends the current entry and starts an entry with the string s.
func (p *parser) msgid(s string) error {
	if err := p.end(); err != nil {
		return err
	}
	t, err := p.unquote(s)
	if err != nil {
		return err
	}
	p.e = &entry{id: t, idLine: p.ln, fuzzy: p.fuzzy}
	p.fuzzy = false
	p.cont = &p.e.id
	return nil
}

// msgstr sets the translation of the current entry to the string s.
func (p *parser) msgstr(s string) error {
	if p.e == nil || p.e.strLine != 0 {
		return p.errorf(p.ln, "msgstr without msgid")
	}
	t, err := p.unquote(s)
	if err != nil {
		return err
	}
	p.e.str, p.e.strLine = t, p.ln
	p.cont = &p.e.str
	return nil
}

// end ends the current entry and adds its translation.
func (p *parser) end() error {
	e := p.e
	if e == nil {
		return nil
	}
	p.e, p.cont = nil, nil
	if e.strLine == 0 {
		return p.errorf(e.idLine, "msgid without msgstr")
	}
	if p.seen[e.id] {
		return p.errorf(e.idLine, "duplicate msgid %q", e.id)
	}
	p.seen[e.id] = true
	if e.id == "" { // The header.
		return nil
	}
	check, ok := rules[e.id]
	if !ok {
		return p.errorf(e.idLine, "msgid %q is not an engine text", e.id)
	}
	if e.fuzzy || e.str == "" {
		return nil
	}
	if err := check(e.str); err != nil {
		return p.errorf(e.strLine, "msgstr: %v", err)
	}
	p.dict[e.id] = e.str
	return nil
}

// errorf gives an error at the line of the file.
func (p *parser) errorf(line int, format string, args ...any) error {
	return fmt.Errorf("%s:%d: %s", p.name, line, fmt.Sprintf(format, args...))
}

// unquote reads the string s of the current line with its quotes and gives its text.
func (p *parser) unquote(s string) (string, error) {
	t, err := unquote(s)
	if err != nil {
		return "", p.errorf(p.ln, "%v", err)
	}
	return t, nil
}

// unquote reads the string s with its quotes and gives its text.
func unquote(s string) (string, error) {
	if s == "" || s[0] != '"' {
		return "", errors.New("no string")
	}
	var b strings.Builder
	for i := 1; i < len(s); i++ {
		switch c := s[i]; c {
		case '"':
			if i+1 < len(s) {
				return "", errors.New("text after the string")
			}
			return b.String(), nil
		case '\\':
			if i+1 == len(s) {
				return "", errors.New("string without end quote")
			}
			i++
			switch e := s[i]; e {
			case '\\', '"':
				b.WriteByte(e)
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			default:
				r, _ := utf8.DecodeRuneInString(s[i:])
				return "", fmt.Errorf("escape \\%c is not supported", r)
			}
		default:
			b.WriteByte(c)
		}
	}
	return "", errors.New("string without end quote")
}
