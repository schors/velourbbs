package screen

import (
	"cmp"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// op is a comparison operator of section 5.
type op uint8

const (
	opNone op = iota
	opEq
	opNe
	opLt
	opLe
	opGt
	opGe
)

// ops lists the operators. Each two-byte operator comes before its one-byte prefix.
var ops = []struct {
	text string
	op   op
}{{"==", opEq}, {"!=", opNe}, {"<=", opLe}, {">=", opGe}, {"<", opLt}, {">", opGt}}

// term is one unary of section 5: a boolean path, "not" path, or a comparison.
type term struct {
	ref Ref
	not bool
	op  op
	n   int    // the literal of an integer field
	s   string // the literal of a string field
}

// Cond is a parsed condition (section 5): a list of "or" alternatives,
// each a list of "and" terms. The grammar has no parentheses, so this form is the full tree.
type Cond struct {
	or [][]term
}

// ParseCond parses the condition s, resolves each path with fields and checks the types.
// items permits item.<KEY>: true for {#if}, false for when (section 4.1).
// The caller checks the keys with CheckItems.
func ParseCond(s string, fields Fields, items bool) (*Cond, error) {
	p := condParser{s: s, fields: fields, items: items}
	c := &Cond{}
	var and []term
	for {
		t, err := p.unary()
		if err != nil {
			return nil, err
		}
		and = append(and, t)
		p.skipSpaces()
		if p.i == len(p.s) {
			c.or = append(c.or, and)
			return c, nil
		}
		switch w := p.word(); w {
		case "and":
		case "or":
			c.or = append(c.or, and)
			and = nil
		default:
			return nil, fmt.Errorf("expected \"and\" or \"or\", found %s", p.found(w))
		}
	}
}

// CheckItems reports each item.<KEY> of c that is not in keys.
func (c *Cond) CheckItems(keys map[string]bool) []error {
	var errs []error
	for _, and := range c.or {
		for _, t := range and {
			if t.ref.isItem() && !keys[t.ref.name] {
				errs = append(errs, fmt.Errorf("unknown field %q", t.ref.String()))
			}
		}
	}
	return errs
}

// Args gives the argument names of the arg.<name> fields of c.
func (c *Cond) Args() []string {
	var names []string
	for _, and := range c.or {
		for _, t := range and {
			if t.ref.kind == refArg {
				names = append(names, t.ref.name)
			}
		}
	}
	return names
}

// FileField gives the path of the first field of a file screen in c, else "" (file-view, section 3.4).
func (c *Cond) FileField() string { return c.find(Ref.IsFile) }

// ListField gives the path of the first list.* field in c, else "" (list-view, section 6).
func (c *Cond) ListField() string { return c.find(Ref.IsList) }

// Refs gives the fields of c.
func (c *Cond) Refs() []Ref {
	var refs []Ref
	for _, and := range c.or {
		for _, t := range and {
			refs = append(refs, t.ref)
		}
	}
	return refs
}

// find gives the path of the first field of c for which is gives true, else "".
func (c *Cond) find(is func(Ref) bool) string {
	for _, and := range c.or {
		for _, t := range and {
			if is(t.ref) {
				return t.ref.String()
			}
		}
	}
	return ""
}

// Eval evaluates c with the values of env.
func (c *Cond) Eval(env *Env) bool {
	for _, and := range c.or {
		if allTrue(and, env) {
			return true
		}
	}
	return false
}

func allTrue(and []term, env *Env) bool {
	for _, t := range and {
		if !t.eval(env) {
			return false
		}
	}
	return true
}

func (t term) eval(env *Env) bool {
	switch v := env.value(t.ref).(type) {
	case bool:
		return v != t.not
	case int:
		c := cmp.Compare(v, t.n)
		switch t.op {
		case opEq:
			return c == 0
		case opNe:
			return c != 0
		case opLt:
			return c < 0
		case opLe:
			return c <= 0
		case opGt:
			return c > 0
		case opGe:
			return c >= 0
		}
	case string:
		return (v == t.s) == (t.op == opEq)
	}
	panic("screen: invalid condition term " + t.ref.String())
}

// condParser reads a condition from left to right. i is the byte offset of the next part.
type condParser struct {
	s      string
	i      int
	fields Fields
	items  bool
}

func (p *condParser) skipSpaces() {
	for p.i < len(p.s) && p.s[p.i] == ' ' {
		p.i++
	}
}

// word reads a path or a keyword: the bytes [A-Za-z0-9_.].
func (p *condParser) word() string {
	start := p.i
	for p.i < len(p.s) && (isKeyByte(p.s[p.i]) || p.s[p.i] == '.') {
		p.i++
	}
	return p.s[start:p.i]
}

// found describes the word w that the parser read, or the rest of s when w is empty.
func (p *condParser) found(w string) string {
	switch {
	case w != "":
		return strconv.Quote(w)
	case p.i == len(p.s):
		return "end of condition"
	}
	return strconv.Quote(p.s[p.i:])
}

func (p *condParser) unary() (term, error) {
	p.skipSpaces()
	w := p.word()
	if w == "not" {
		p.skipSpaces()
		ref, err := p.path(p.word())
		if err != nil {
			return term{}, err
		}
		if typ := ref.Type(); typ != TypeBool && typ != 0 {
			return term{}, fmt.Errorf("\"not\" needs a boolean field, %q is %s", ref.String(), typ)
		}
		return term{ref: ref, not: true}, nil
	}
	ref, err := p.path(w)
	if err != nil {
		return term{}, err
	}
	p.skipSpaces()
	t := term{ref: ref, op: p.op()}
	if ref.Type() == 0 { // a row.<field> that no provider gives: the load reports it
		if t.op != opNone {
			p.skipSpaces()
			return t, p.literal(&t)
		}
		return t, nil
	}
	if t.op == opNone {
		if ref.Type() != TypeBool {
			return term{}, fmt.Errorf("%s field %q needs an operator", ref.Type(), ref.String())
		}
		return t, nil
	}
	if ref.Type() == TypeBool {
		return term{}, fmt.Errorf("boolean field %q takes no operator", ref.String())
	}
	p.skipSpaces()
	return t, p.literal(&t)
}

// path resolves the word w as a field path.
func (p *condParser) path(w string) (Ref, error) {
	if w == "" || w == "and" || w == "or" || w == "not" {
		return Ref{}, fmt.Errorf("expected a field, found %s", p.found(w))
	}
	ref, err := resolve(w, p.fields)
	if err == nil && ref.isItem() && !p.items {
		return Ref{}, fmt.Errorf("field %q is valid only in {#if}", w)
	}
	return ref, err
}

// op reads an operator, or gives opNone when no operator follows.
// A '=' or '!' that starts no operator is left for the caller to report.
func (p *condParser) op() op {
	for _, o := range ops {
		if strings.HasPrefix(p.s[p.i:], o.text) {
			p.i += len(o.text)
			return o.op
		}
	}
	return opNone
}

// literal reads the literal of t and checks its type against the field of t.
func (p *condParser) literal(t *term) error {
	typ := t.ref.Type()
	if p.i < len(p.s) && p.s[p.i] == '"' {
		s, err := p.str()
		switch {
		case err != nil:
			return err
		case typ != TypeString && typ != 0:
			return fmt.Errorf("%s field %q takes an integer, found a string", typ, t.ref.String())
		case t.op != opEq && t.op != opNe:
			return fmt.Errorf("string field %q takes only == or !=", t.ref.String())
		}
		t.s = s
		return nil
	}
	n, err := p.integer()
	switch {
	case err != nil:
		return err
	case typ != TypeInt && typ != 0:
		return fmt.Errorf("%s field %q takes a string, found an integer", typ, t.ref.String())
	}
	t.n = n
	return nil
}

// str reads a string literal. s[i] is the opening '"'.
func (p *condParser) str() (string, error) {
	end := strings.IndexAny(p.s[p.i+1:], "\"}")
	switch {
	case end < 0:
		return "", errors.New("string without closing '\"'")
	case p.s[p.i+1+end] == '}':
		return "", errors.New("'}' in a string")
	}
	s := p.s[p.i+1 : p.i+1+end]
	p.i += end + 2
	return s, nil
}

// integer reads an integer literal: [-]digit{digit}.
func (p *condParser) integer() (int, error) {
	start := p.i
	if p.i < len(p.s) && p.s[p.i] == '-' {
		p.i++
	}
	digits := p.i
	for p.i < len(p.s) && '0' <= p.s[p.i] && p.s[p.i] <= '9' {
		p.i++
	}
	if p.i == digits {
		p.i = start
		return 0, fmt.Errorf("expected an integer or a string, found %s", p.found(""))
	}
	n, err := strconv.Atoi(p.s[start:p.i])
	if err != nil {
		return 0, fmt.Errorf("integer %s is out of range", p.s[start:p.i])
	}
	return n, nil
}
