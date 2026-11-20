package tomlcheck

import (
	"sort"

	"github.com/pelletier/go-toml/v2/unstable"
)

// Pos is a 1-based position in a document. Col counts bytes.
type Pos struct{ Line, Col int }

// Place is a key path in a document. It gives the positions of screen-render, section 3.1.
// The zero Place is a path that the document does not set, at 0:0.
type Place struct {
	n *node // nil when the document does not set the path
	// last is the position of the last table on the path that the document sets.
	last Pos
}

// node is a key of a document and its value.
type node struct {
	key   Pos              // the key name where the document sets the key first
	value Pos              // the value, or the table
	keys  map[string]*node // the keys of a table, else nil
	elems []*node          // the elements of an array or of an array of tables
	// header is true when a table header sets the position of the table.
	header bool
}

// At gives the place of the key path under p.
// A path into an array, or past a value that is not a table, is not set.
func (p Place) At(path ...string) Place {
	for _, k := range path {
		if p.n == nil {
			return p
		}
		c := p.n.keys[k]
		if c == nil {
			return Place{last: p.last}
		}
		p.n = c
		if c.keys != nil {
			p.last = c.value
		}
	}
	return p
}

// Elem gives the place of the 0-based element i of the array or the array of tables at p.
func (p Place) Elem(i int) Place {
	if p.n == nil || i < 0 || i >= len(p.n.elems) {
		return Place{last: p.last}
	}
	e := p.n.elems[i]
	if e.keys != nil {
		return Place{e, e.value}
	}
	return Place{e, p.last}
}

// Value gives the position of the value at p: its first byte, or the position of the table.
// A table has the position of its header, of the { of an inline table, or of the first
// dotted key that makes it. A path that is not set has the position of the last table on it.
// An array in an array has the position of the outer array: the parser gives no range
// of an array, and no key of the host has a nested array.
func (p Place) Value() Pos {
	if p.n == nil {
		return p.last
	}
	return p.n.value
}

// Key gives the position of the key name at p, where the document sets the key first.
// A path that is not set has the position of the last table on it.
func (p Place) Key() Pos {
	if p.n == nil {
		return p.last
	}
	return p.n.key
}

// Root gives the place of the root table of data. data has no syntax error.
// The root table has the position 1:1.
// The decode of go-toml gives no positions of keys and values, so the parser of go-toml gives them.
func Root(data []byte) Place {
	ix := index{data: data, lines: []int{0}}
	for i, c := range data {
		if c == '\n' {
			ix.lines = append(ix.lines, i+1)
		}
	}
	root := newTable(Pos{1, 1}, Pos{1, 1})
	cur := root
	var p unstable.Parser
	p.Reset(data)
	for p.NextExpression() {
		e := p.Expression()
		switch e.Kind {
		case unstable.KeyValue:
			ix.keyValue(cur, e)
		case unstable.Table:
			parts := keyParts(e)
			t, last := ix.parent(root, parts)
			cur = ix.child(t, last)
			if !cur.header {
				cur.value, cur.header = ix.header(parts[0], 1), true
			}
		case unstable.ArrayTable:
			parts := keyParts(e)
			t, last := ix.parent(root, parts)
			pos := ix.header(parts[0], 2)
			a := t.keys[string(last.Data)]
			if a == nil {
				a = &node{key: ix.pos(last.Raw.Offset), value: pos}
				t.keys[string(last.Data)] = a
			}
			cur = newTable(a.key, pos)
			a.elems = append(a.elems, cur)
		}
	}
	return Place{root, root.value}
}

// index builds the nodes of a document.
type index struct {
	data  []byte
	lines []int // the offset of each line
}

func newTable(key, value Pos) *node {
	return &node{key: key, value: value, keys: map[string]*node{}}
}

// pos gives the position of the byte offset off.
func (ix *index) pos(off uint32) Pos {
	o := int(off)
	line := sort.Search(len(ix.lines), func(i int) bool { return ix.lines[i] > o })
	return Pos{line, o - ix.lines[line-1] + 1}
}

func keyParts(e *unstable.Node) []*unstable.Node {
	var parts []*unstable.Node
	for it := e.Key(); it.Next(); {
		parts = append(parts, it.Node())
	}
	return parts
}

// parent gives the table of the dotted key parts under t, without the last part, and the last part.
func (ix *index) parent(t *node, parts []*unstable.Node) (*node, *unstable.Node) {
	for _, k := range parts[:len(parts)-1] {
		t = ix.child(t, k)
	}
	return t, parts[len(parts)-1]
}

// child gives the table of the key part k in t. It makes the table when t does not have it.
// For an array of tables it gives the last table, as TOML does.
func (ix *index) child(t *node, k *unstable.Node) *node {
	c := t.keys[string(k.Data)]
	if c == nil {
		pos := ix.pos(k.Raw.Offset)
		c = newTable(pos, pos)
		t.keys[string(k.Data)] = c
	}
	if len(c.elems) > 0 && c.keys == nil {
		c = c.elems[len(c.elems)-1]
	}
	if c.keys == nil {
		// Only a document with a syntax error has a value here.
		c.keys = map[string]*node{}
	}
	return c
}

// header gives the position of the first [ of the table header with the first key part k.
// n is the number of brackets: 1 for a table, 2 for an array of tables.
func (ix *index) header(k *unstable.Node, n int) Pos {
	i := int(k.Raw.Offset) - 1
	for i > 0 && (ix.data[i] == ' ' || ix.data[i] == '\t') {
		i--
	}
	return ix.pos(uint32(max(i-n+1, 0))) //nolint:gosec // a document is at most MaxBytes
}

// keyValue adds the key-value expression e to the table t.
func (ix *index) keyValue(t *node, e *unstable.Node) {
	t, last := ix.parent(t, keyParts(e))
	c := &node{key: ix.pos(last.Raw.Offset)}
	ix.value(c, e.Value(), ix.valueStart(last))
	t.keys[string(last.Data)] = c
}

// valueStart gives the offset of the value after the last key part k of a key-value:
// the parser gives no range of an array.
func (ix *index) valueStart(k *unstable.Node) uint32 {
	i := int(k.Raw.Offset + k.Raw.Length)
	for i < len(ix.data) && (ix.data[i] == ' ' || ix.data[i] == '\t' || ix.data[i] == '=') {
		i++
	}
	return uint32(i) //nolint:gosec // a document is at most MaxBytes
}

// value sets the position of the value v in c, and its keys or its elements.
// start is the offset of v when the parser gives no range of v.
func (ix *index) value(c *node, v *unstable.Node, start uint32) {
	switch v.Kind {
	case unstable.InlineTable:
		c.value = ix.pos(v.Raw.Offset)
		c.keys = map[string]*node{}
		for it := v.Children(); it.Next(); {
			if kv := it.Node(); kv.Kind == unstable.KeyValue {
				ix.keyValue(c, kv)
			}
		}
	case unstable.Array:
		c.value = ix.pos(start)
		for it := v.Children(); it.Next(); {
			el := it.Node()
			if el.Kind == unstable.Comment {
				continue
			}
			e := &node{key: c.key}
			// A nested array has the position of the outer array.
			ix.value(e, el, start)
			c.elems = append(c.elems, e)
		}
	default:
		c.value = ix.pos(v.Raw.Offset)
	}
}
