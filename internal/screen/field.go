package screen

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"newbbs/internal/key"
)

// Type is the type of a field (section 4.1).
type Type uint8

// The field types.
const (
	TypeBool Type = iota + 1
	TypeInt
	TypeString
)

func (t Type) String() string {
	switch t {
	case TypeBool:
		return "boolean"
	case TypeInt:
		return "integer"
	case TypeString:
		return "string"
	}
	return "type " + strconv.Itoa(int(t))
}

// termField is a term.* field of the registry. The zero value is item.<KEY>.
type termField uint8

const (
	termType termField = iota + 1
	termCols
	termRows
	termCharset
	termColors
)

// termFields is the registry of section 4.1 without item.<KEY>.
var termFields = [...]struct {
	path string
	typ  Type
}{
	termType:    {"term.type", TypeString},
	termCols:    {"term.cols", TypeInt},
	termRows:    {"term.rows", TypeInt},
	termCharset: {"term.charset", TypeString},
	termColors:  {"term.colors", TypeInt},
}

// Ref is a resolved field path: a term.* field or item.<KEY>.
type Ref struct {
	term termField
	key  string // the <KEY> of item.<KEY>, in upper case
}

// Type gives the type of the field.
func (r Ref) Type() Type {
	if r.isItem() {
		return TypeBool
	}
	return termFields[r.term].typ
}

// String gives the registry path.
func (r Ref) String() string {
	if r.isItem() {
		return "item." + r.key
	}
	return termFields[r.term].path
}

func (r Ref) isItem() bool { return r.term == 0 }

// Fields maps an alias to its field (section 4.2).
type Fields map[string]Ref

// reserved are the names that an alias cannot have (section 4.2).
var reserved = []string{"term", "item", "and", "or", "not"}

// NewAlias checks the alias name and its path against section 4.2
// and gives the field of the path.
func NewAlias(name, path string) (Ref, error) {
	switch {
	case !IsIdent(name):
		return Ref{}, fmt.Errorf("invalid alias %q", name)
	case slices.Contains(reserved, name):
		return Ref{}, fmt.Errorf("alias %q is a reserved word", name)
	case !strings.Contains(path, "."):
		return Ref{}, fmt.Errorf("alias %q: %q is not a registry path", name, path)
	}
	r, err := resolve(path, nil)
	if err != nil {
		return Ref{}, fmt.Errorf("alias %q: %w", name, err)
	}
	if r.isItem() {
		return Ref{}, fmt.Errorf("alias %q: %q is valid only in {#if}", name, path)
	}
	return r, nil
}

// LookupField resolves the path of a field token. item.<KEY> is an error.
func LookupField(path string, fields Fields) (Ref, error) {
	r, err := resolve(path, fields)
	if err == nil && r.isItem() {
		return Ref{}, fmt.Errorf("field %q is valid only in {#if}", path)
	}
	return r, err
}

// IsPath reports whether s matches the path syntax of section 4.1:
// [a-z][a-z0-9_]*(\.[A-Za-z0-9_]+)?
func IsPath(s string) bool {
	head, tail, dot := strings.Cut(s, ".")
	return IsIdent(head) && (!dot || isKey(tail))
}

// resolve checks the path syntax of section 4.1 and resolves an alias with fields.
func resolve(path string, fields Fields) (Ref, error) {
	if !IsPath(path) {
		return Ref{}, fmt.Errorf("invalid field path %q", path)
	}
	head, tail, dot := strings.Cut(path, ".")
	if !dot {
		if r, ok := fields[path]; ok {
			return r, nil
		}
	} else if head == "item" {
		return Ref{key: key.Upper(tail)}, nil
	}
	for i, f := range termFields {
		if i > 0 && f.path == path {
			return Ref{term: termField(i)}, nil
		}
	}
	return Ref{}, fmt.Errorf("unknown field %q", path)
}

// isKey reports whether s matches [A-Za-z0-9_]+.
func isKey(s string) bool {
	for i := range len(s) {
		if !isKeyByte(s[i]) {
			return false
		}
	}
	return s != ""
}

func isKeyByte(c byte) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' || c == '_'
}

// Term holds the term.* values of one session.
type Term struct {
	Type    string
	Cols    int
	Rows    int
	Charset string
	Colors  int
}

// The ranges of term.cols and term.rows (section 4.1).
const (
	minCols = 20
	maxCols = MaxWidth
	minRows = 10
	maxRows = 100
)

// Check reports an error when term.cols, term.rows or term.colors
// is outside the values of section 4.1.
func (t Term) Check() error {
	switch {
	case t.Cols < minCols || t.Cols > maxCols:
		return fmt.Errorf("term.cols is %d, must be from %d to %d", t.Cols, minCols, maxCols)
	case t.Rows < minRows || t.Rows > maxRows:
		return fmt.Errorf("term.rows is %d, must be from %d to %d", t.Rows, minRows, maxRows)
	case t.Colors != 16 && t.Colors != 256 && t.Colors != 1<<24:
		return fmt.Errorf("term.colors is %d, must be 16, 256 or 16777216", t.Colors)
	}
	return nil
}

// Env holds the field values of one render.
type Env struct {
	Term  Term
	Items map[string]bool // the keys of the enabled items
}

// value gives the value of r: a bool, an int or a string.
func (e *Env) value(r Ref) any {
	switch r.term {
	case termType:
		return e.Term.Type
	case termCols:
		return e.Term.Cols
	case termRows:
		return e.Term.Rows
	case termCharset:
		return e.Term.Charset
	case termColors:
		return e.Term.Colors
	}
	return e.Items[r.key]
}

// Text gives the graphemes of a field token: an integer in decimal,
// a string after CleanValue (sections 4.1 and 4.3). r is not item.<KEY>.
func (e *Env) Text(r Ref) []Grapheme {
	switch v := e.value(r).(type) {
	case int:
		return CleanValue(strconv.Itoa(v))
	case string:
		return CleanValue(v)
	}
	panic("screen: Text of " + r.String())
}
