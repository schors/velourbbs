// Package tomlpos decodes a TOML document and gives the position of each key.
//
// toml.Decoder stops at the first type error and gives no position for a map key,
// but `bbs check` reports all errors with positions (docs/specs/screen-format.md,
// section 9.1). So this package takes the values from toml.Unmarshal and the key
// positions from the unstable parser of go-toml/v2. It is the only user of that
// parser (section 12).
package tomlpos

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/pelletier/go-toml/v2/unstable"
)

// Pos is a 1-based line and column. The column counts bytes.
type Pos struct {
	Line, Col int
}

// Error is a TOML syntax or decode error.
type Error struct {
	Pos Pos
	Msg string
}

func (e *Error) Error() string { return fmt.Sprintf("%d:%d: %s", e.Pos.Line, e.Pos.Col, e.Msg) }

// Doc is a decoded document.
type Doc struct {
	// Values has the types of toml.Unmarshal into map[string]any:
	// string, int64, float64, bool, a time type, []any and map[string]any.
	// An array of tables is []any of map[string]any.
	Values map[string]any
	root   node
}

// node is a key, a table or an array element with its position.
// A tree of nodes keeps the cost of each key independent of its depth.
type node struct {
	pos    Pos
	kids   map[string]*node
	tables int // the count of tables in an array of tables
}

// child returns the child with the name. A new child gets the position pos.
func (n *node) child(name string, pos Pos) *node {
	c, ok := n.kids[name]
	if !ok {
		if n.kids == nil {
			n.kids = map[string]*node{}
		}
		c = &node{pos: pos}
		n.kids[name] = c
	}
	return c
}

// Pos returns the position of the key with the path, or of its longest prefix
// with a position. An array element has its index as a path part.
// Example: Pos("item", "0", "key") for the first [[item]] table.
func (d *Doc) Pos(path ...string) Pos {
	p := Pos{Line: 1, Col: 1}
	n := &d.root
	for _, part := range path {
		c, ok := n.kids[part]
		if !ok {
			break
		}
		n, p = c, c.pos
	}
	return p
}

// MaxKeys is the key limit of a document (section 10). Each part of a dotted key
// and of a table header counts. toml.Unmarshal checks the keys for duplicates
// in a time that grows with the square of the key count.
const MaxKeys = 1000

// Parse decodes data. The error is an *Error.
// It counts the keys before toml.Unmarshal and stops after MaxKeys.
func Parse(data []byte) (*Doc, error) {
	d := &Doc{}
	w := walker{root: &d.root, table: &d.root, lines: lineStarts(data)}
	w.p.Reset(data)
	for w.over == nil && w.p.NextExpression() {
		w.expression(w.p.Expression())
	}
	if w.over != nil {
		return nil, &Error{*w.over, fmt.Sprintf("too many keys: the limit is %d", MaxKeys)}
	}
	if err := toml.Unmarshal(data, &d.Values); err != nil {
		var de *toml.DecodeError
		if errors.As(err, &de) {
			line, col := de.Position()
			return nil, &Error{Pos{line, col}, strings.TrimPrefix(de.Error(), "toml: ")}
		}
		return nil, &Error{Pos{1, 1}, err.Error()}
	}
	if err := w.p.Error(); err != nil {
		// toml.Unmarshal accepted the document, so this is a defect of go-toml.
		return nil, &Error{Pos{1, 1}, err.Error()}
	}
	if d.Values == nil {
		d.Values = map[string]any{}
	}
	return d, nil
}

type walker struct {
	p     unstable.Parser
	root  *node
	table *node // the current table
	lines []int // the offset of each line start
	keys  int   // the count of key parts
	over  *Pos  // the first key part over MaxKeys
}

func (w *walker) expression(e *unstable.Node) {
	switch e.Kind {
	case unstable.Table:
		w.table = w.key(w.root, e, true)
	case unstable.ArrayTable:
		array := w.key(w.root, e, false)
		first := e.Key()
		first.Next()
		w.table = array.child(strconv.Itoa(array.tables), w.pos(first.Node()))
		array.tables++
	case unstable.KeyValue:
		w.keyValue(w.table, e)
	}
}

// key walks the key parts of e from n and records each of them.
// After an array of tables, the walk goes into its last table, as TOML requires.
// Example: after [[item]] twice, [item.sub] is item, 1, sub.
// For the header of an array of tables, last is false: the walk stays on the array.
func (w *walker) key(n *node, e *unstable.Node, last bool) *node {
	for it := e.Key(); it.Next(); {
		k := it.Node()
		pos := w.pos(k)
		if w.keys++; w.keys > MaxKeys && w.over == nil {
			w.over = &pos
		}
		n = n.child(string(k.Data), pos)
		if n.tables > 0 && (last || !it.IsLast()) {
			n = n.kids[strconv.Itoa(n.tables-1)]
		}
	}
	return n
}

func (w *walker) keyValue(table *node, e *unstable.Node) {
	w.value(w.key(table, e, true), e.Value())
}

// value records the keys of inline tables and the elements of arrays.
func (w *walker) value(n *node, v *unstable.Node) {
	switch v.Kind {
	case unstable.InlineTable:
		for it := v.Children(); it.Next(); {
			w.keyValue(n, it.Node())
		}
	case unstable.Array:
		i := 0
		for it := v.Children(); it.Next(); i++ {
			el := it.Node()
			pos := n.pos // a nested array has no range of its own
			if el.Raw.Length > 0 {
				pos = w.pos(el)
			}
			w.value(n.child(strconv.Itoa(i), pos), el)
		}
	}
}

func (w *walker) pos(n *unstable.Node) Pos {
	off := int(n.Raw.Offset)
	line, _ := slices.BinarySearch(w.lines, off+1)
	return Pos{Line: line, Col: off - w.lines[line-1] + 1}
}

func lineStarts(data []byte) []int {
	starts := []int{0}
	for i, c := range data {
		if c == '\n' {
			starts = append(starts, i+1)
		}
	}
	return starts
}
