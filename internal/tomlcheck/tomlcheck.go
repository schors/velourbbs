// Package tomlcheck decodes a TOML document of the host: bbs.toml, screens/screens.toml
// or a front matter.
// It gives the errors of docs/specs/screen-render.md, section 3.1.
//
// go-toml v2 has two gaps that Decode closes:
//   - A key matches a field also without case, and DisallowUnknownFields does not stop it.
//   - A table header [x] fills a slice of structs as one element.
//
// So a decode into map[string]any finds them, and Root gives their positions.
// The same decode finds the syntax errors before the strict decode.
package tomlcheck

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/pelletier/go-toml/v2"
)

// MaxBytes is the size limit of a TOML document (bbt-format, section 9, and bbs-toml,
// section 2). The time of a decode grows with the square of the size: each unknown key error
// scans the document.
const MaxBytes = 64 << 10

// Error is an error of a document. Line and Col are 1-based; Col counts bytes.
type Error struct {
	Line, Col int
	Msg       string
}

// RuneCol gives the column of e in code points. data is the document of e.
func (e Error) RuneCol(data []byte) int {
	off, ok := lineStart(data, e.Line)
	if !ok || e.Line < 1 {
		return e.Col
	}
	line := data[off:]
	if i := bytes.IndexByte(line, '\n'); i >= 0 {
		line = line[:i]
	}
	return runeCol(line, e.Col)
}

// runeCol gives the column in code points of the byte column col in line.
func runeCol(line []byte, col int) int {
	return utf8.RuneCount(line[:min(max(col-1, 0), len(line))]) + 1
}

// lineStart gives the offset of the 1-based line in data.
func lineStart(data []byte, line int) (int, bool) {
	off := 0
	for range line - 1 {
		i := bytes.IndexByte(data[off:], '\n')
		if i < 0 {
			return 0, false
		}
		off += i + 1
	}
	return off, true
}

// Decode decodes data into shape, a pointer to a struct.
// shape has a field with a plain type for each key: string, int64, a slice, a map, a struct.
// Decode gives the first of these: an error of the size, of the syntax or of a type;
// all unknown keys and tables in place of an array of tables.
// The message of a type error starts with the dotted key path.
// Without an error of the size or of the syntax, it also gives the root of data (Root).
func Decode(data []byte, shape any) (Place, []Error) {
	return decodeDoc(newIndex(data, false), shape)
}

// DecodeFront decodes the front matter data of a file as Decode does.
// The errors and the places give positions in the file: line 1 of data is line 2 of the file,
// and a column counts code points (screen-render, section 3.1).
func DecodeFront(data []byte, shape any) (Place, []Error) {
	return decodeDoc(newIndex(data, true), shape)
}

func decodeDoc(ix *index, shape any) (Place, []Error) {
	data := ix.data
	m, errs := Map(data)
	if errs != nil {
		return Place{}, ix.fileErrors(errs)
	}
	root := ix.root()
	t := reflect.TypeOf(shape)
	err := toml.NewDecoder(bytes.NewReader(data)).DisallowUnknownFields().Decode(shape)
	var unknown *toml.StrictMissingError
	if err != nil && !errors.As(err, &unknown) {
		if e, ok := typeError(err, t, m); ok {
			return root, ix.fileErrors([]Error{e})
		}
		// The key of the error matches a field only without case: exact reports it.
	}
	if unknown != nil {
		for _, e := range unknown.Errors {
			line, col := e.Position()
			errs = append(errs, Error{line, col, fmt.Sprintf("unknown key %q", strings.Join(e.Key(), "."))})
		}
	}
	errs = ix.fileErrors(errs)
	// The places of root give positions in the file.
	exact(t, m, root, nil, &errs)
	return root, errs
}

// fileErrors gives the errors with their positions in the file (index.file).
func (ix *index) fileErrors(errs []Error) []Error {
	for i, e := range errs {
		p := ix.file(Pos{e.Line, e.Col})
		errs[i].Line, errs[i].Col = p.Line, p.Col
	}
	return errs
}

// MapFront decodes the front matter data of a file as Map does.
// The errors give positions in the file (DecodeFront).
func MapFront(data []byte) (map[string]any, []Error) {
	m, errs := Map(data)
	return m, newIndex(data, true).fileErrors(errs)
}

// Map decodes data into a map. It gives the error of the size or of the syntax.
func Map(data []byte) (map[string]any, []Error) {
	if len(data) > MaxBytes {
		return nil, []Error{{1, 1, fmt.Sprintf("TOML document is over %d bytes", MaxBytes)}}
	}
	var m map[string]any
	if err := toml.Unmarshal(data, &m); err != nil {
		return nil, []Error{syntaxError(data, err)}
	}
	return m, nil
}

// exact reports each key of the table m that matches a field of the struct type t
// only without case, and each table in place of an array of tables.
// at is the place of m. Each error has the position of the key name.
// go-toml reports the other unknown keys.
func exact(t reflect.Type, m map[string]any, at Place, path []string, errs *[]Error) {
	t = deref(t)
	if t.Kind() != reflect.Struct { // a map takes each key
		return
	}
	errorf := func(at Place, format string, args ...any) {
		pos := at.Key()
		*errs = append(*errs, Error{pos.Line, pos.Col, fmt.Sprintf(format, args...)})
	}
	for _, k := range slices.Sorted(maps.Keys(m)) {
		p := append(slices.Clip(path), k)
		f, ok := field(t, k)
		if !ok {
			if foldField(t, k) {
				errorf(at.At(k), "unknown key %q", strings.Join(p, "."))
			}
			continue
		}
		ft := deref(f.Type)
		switch v := m[k].(type) {
		case map[string]any:
			switch {
			case ft.Kind() == reflect.Struct:
				exact(ft, v, at.At(k), p, errs)
			case isTables(ft):
				errorf(at.At(k), "%s: must be an array of tables", strings.Join(p, "."))
			}
		case []any:
			if isTables(ft) {
				for i, el := range v {
					if el, ok := el.(map[string]any); ok {
						exact(ft.Elem(), el, at.At(k).Elem(i), p, errs)
					}
				}
			}
		}
	}
}

// isTables reports whether t is a slice of structs or of maps: an array of tables.
func isTables(t reflect.Type) bool {
	if t.Kind() != reflect.Slice {
		return false
	}
	k := deref(t.Elem()).Kind()
	return k == reflect.Struct || k == reflect.Map
}

// syntaxError converts an error of the decode of data into a map.
func syntaxError(data []byte, err error) Error {
	var de *toml.DecodeError
	if !errors.As(err, &de) {
		return Error{1, 1, err.Error()}
	}
	line, col := de.Position()
	// go-toml gives the first byte of a character that is not ASCII.
	return Error{line, col, fixRune(strings.TrimPrefix(de.Error(), "toml: "), data, line, col)}
}

// typeError converts an error of the strict decode into the struct type t.
// m is the decode of the document into a map.
// ok is false when the key path has no field.
func typeError(err error, t reflect.Type, m map[string]any) (e Error, ok bool) {
	var de *toml.DecodeError
	if !errors.As(err, &de) {
		return Error{1, 1, err.Error()}, true
	}
	line, col := de.Position()
	ft, key := fieldType(t, de.Key())
	if ft == nil {
		return Error{}, false
	}
	msg := "must be " + describe(ft)
	if inTable(ft, lookup(m, key)) {
		// In an inline table the key path of go-toml ends at the table,
		// so the field of the error is not known.
		msg = "invalid type"
	}
	return Error{line, col, strings.Join(key, ".") + ": " + msg}, true
}

// inTable reports whether v is a table, or an array with a table, at a field of the type t:
// a struct, a map or a slice of structs. Then a type error at the key of t
// is in a value of the table.
func inTable(t reflect.Type, v any) bool {
	t = deref(t)
	if t.Kind() != reflect.Struct && t.Kind() != reflect.Map && !isTables(t) {
		return false
	}
	isTable := func(v any) bool { _, ok := v.(map[string]any); return ok }
	if a, ok := v.([]any); ok {
		return isTables(t) && slices.ContainsFunc(a, isTable)
	}
	return isTable(v)
}

// lookup gives the value at the key path in the table m, or nil.
// It does not go into an array of tables.
func lookup(m map[string]any, key []string) any {
	var v any = m
	for _, k := range key {
		t, ok := v.(map[string]any)
		if !ok {
			return nil
		}
		v = t[k]
	}
	return v
}

// fieldType gives the type of the field at the key path from the struct type t,
// and the part of the path up to the field. The walk stops at the first field that
// is not a table: a key path past it is an error of that field.
// An array of tables continues the path in its element. The type is nil when
// the path has no field.
func fieldType(t reflect.Type, key []string) (reflect.Type, []string) {
	for i, part := range key {
		t = deref(t)
		if isTables(t) {
			t = deref(t.Elem())
		}
		switch t.Kind() {
		case reflect.Map:
			t = t.Elem()
		case reflect.Struct:
			f, ok := field(t, part)
			if !ok {
				return nil, key
			}
			t = f.Type
		default:
			return t, key[:i]
		}
	}
	return t, key
}

func field(t reflect.Type, name string) (reflect.StructField, bool) {
	for i := range t.NumField() {
		f := t.Field(i)
		if tag, _, _ := strings.Cut(f.Tag.Get("toml"), ","); tag == name {
			return f, true
		}
	}
	return reflect.StructField{}, false
}

// foldField reports whether a field of the struct type t has the name without case.
func foldField(t reflect.Type, name string) bool {
	for i := range t.NumField() {
		if tag, _, _ := strings.Cut(t.Field(i).Tag.Get("toml"), ","); strings.EqualFold(tag, name) {
			return true
		}
	}
	return false
}

func deref(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

// describe names the TOML type of the field type t.
func describe(t reflect.Type) string {
	t = deref(t)
	switch t.Kind() {
	case reflect.String:
		return "a string"
	case reflect.Int64:
		return "an integer"
	case reflect.Slice:
		switch deref(t.Elem()).Kind() {
		case reflect.Struct, reflect.Map:
			return "an array of tables"
		case reflect.String:
			return "an array of strings"
		case reflect.Int64:
			return "an array of integers"
		}
	case reflect.Map:
		if t.Elem().Kind() == reflect.String {
			return "a table of strings"
		}
	}
	return "a table"
}

// fixRune puts the character at the position into msg in place of the rune
// of its first byte, as go-toml prints it.
func fixRune(msg string, data []byte, line, col int) string {
	off, ok := lineStart(data, line)
	if !ok {
		return msg
	}
	off += col - 1
	if off < 0 || off >= len(data) || data[off] < utf8.RuneSelf {
		return msg
	}
	r, _ := utf8.DecodeRune(data[off:])
	return strings.Replace(msg, fmt.Sprintf("%#U", rune(data[off])), fmt.Sprintf("%#U", r), 1)
}
