package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// format is the value of format in the header (section 2).
const format = "velourbbs-store"

// timeLayout is the form of a time: RFC 3339 in UTC, with seconds and no fraction (section 3).
const timeLayout = "2006-01-02T15:04:05Z"

// schema is the kinds and the migrations of one format version (sections 3 and 6).
type schema struct {
	kinds []kind // in the order of the write
	// migrations[n-1] changes the records of version n to version n+1.
	migrations []migration
}

// version gives the format version of the schema.
func (s *schema) version() int { return len(s.migrations) + 1 }

// kind is a type of record (section 3).
type kind struct {
	name   string
	fields []field // in the order of the write
	limit  int     // the maximum number of records, at least 1
	check  func(Record) error
}

// field is a field of a kind.
type field struct {
	name string
	typ  fieldType
}

// fieldType is the type of the value of a field.
type fieldType int

// The field types and their Go values in a Record.
const (
	typeString fieldType = iota // string
	typeInt                     // int64
	typeTime                    // time.Time
)

// raw is a record of the file before the checks of its kind: the line in the file
// and the fields as encoding/json decodes them with UseNumber.
type raw struct {
	Line   int
	Fields map[string]any
}

// migration changes the records of one version to the next version (section 6).
type migration func([]raw) []raw

// kind gives the kind name, or nil.
func (s *schema) kind(name string) *kind {
	i := slices.IndexFunc(s.kinds, func(k kind) bool { return k.name == name })
	if i < 0 {
		return nil
	}
	return &s.kinds[i]
}

// field gives the field name of k, or nil. The case of the name counts (section 4).
func (k *kind) field(name string) *field {
	i := slices.IndexFunc(k.fields, func(f field) bool { return f.name == name })
	if i < 0 {
		return nil
	}
	return &k.fields[i]
}

// header checks the header line and gives the format version of the file (section 2).
func (s *schema) header(line []byte) (int, error) {
	invalid := fmt.Errorf(`invalid header, want {"format":%q,"version":<n>}`, format)
	f, err := parseLine(line)
	if err != nil || len(f) != 2 || f["format"] != format {
		return 0, invalid
	}
	n, ok := f["version"].(json.Number)
	if !ok {
		return 0, invalid
	}
	v, err := strconv.ParseInt(string(n), 10, 64)
	if err != nil || v < 1 {
		return 0, invalid
	}
	if v > int64(s.version()) {
		return 0, fmt.Errorf("version %d is newer than this velourbbs (%d)", v, s.version())
	}
	return int(v), nil
}

// parseLine parses a line that is one JSON object. A duplicate field is an error.
func parseLine(line []byte) (map[string]any, error) {
	if !utf8.Valid(line) {
		return nil, errors.New("invalid UTF-8")
	}
	notObject := func(err error) error {
		if err != nil {
			return fmt.Errorf("not one JSON object: %v", err)
		}
		return errors.New("not one JSON object")
	}
	dec := json.NewDecoder(bytes.NewReader(line))
	dec.UseNumber()
	if t, err := dec.Token(); err != nil || t != json.Delim('{') {
		return nil, notObject(err)
	}
	fields := map[string]any{}
	for dec.More() {
		t, err := dec.Token()
		if err != nil {
			return nil, notObject(err)
		}
		name := t.(string) // a token before a value of an object is its name
		var v any
		if err := dec.Decode(&v); err != nil {
			return nil, notObject(err)
		}
		if _, dup := fields[name]; dup {
			return nil, fmt.Errorf("duplicate field %q", name)
		}
		fields[name] = v
	}
	if t, err := dec.Token(); err != nil || t != json.Delim('}') {
		return nil, notObject(err)
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, notObject(err)
	}
	return fields, nil
}

// record checks the fields of a record with its kind (section 4) and gives the record.
// It gives all errors of the record.
func (s *schema) record(fields map[string]any) (Record, []string) {
	v, ok := fields["kind"]
	if !ok {
		return Record{}, []string{`missing field "kind"`}
	}
	name, ok := v.(string)
	if !ok {
		return Record{}, []string{`field "kind": must be a string`}
	}
	k := s.kind(name)
	if k == nil {
		return Record{}, []string{fmt.Sprintf("unknown kind %q", name)}
	}
	var errs []string
	for _, n := range slices.Sorted(maps.Keys(fields)) {
		if n != "kind" && k.field(n) == nil {
			errs = append(errs, fmt.Sprintf("unknown field %q", n))
		}
	}
	rec := Record{Kind: name, Fields: make(map[string]any, len(k.fields))}
	for _, f := range k.fields {
		v, ok := fields[f.name]
		if !ok {
			errs = append(errs, fmt.Sprintf("missing field %q", f.name))
			continue
		}
		val, err := f.typ.value(v)
		if err != nil {
			errs = append(errs, fmt.Sprintf("field %q: %v", f.name, err))
			continue
		}
		rec.Fields[f.name] = val
	}
	if len(errs) == 0 && k.check != nil {
		if err := k.check(rec); err != nil {
			errs = append(errs, err.Error())
		}
	}
	return rec, errs
}

// value gives the Go value of a decoded JSON value v. null is a wrong type.
func (t fieldType) value(v any) (any, error) {
	switch t {
	case typeString:
		if s, ok := v.(string); ok {
			return s, nil
		}
		return nil, errors.New("must be a string")
	case typeInt:
		if n, ok := v.(json.Number); ok {
			if i, err := strconv.ParseInt(string(n), 10, 64); err == nil {
				return i, nil
			}
		}
		return nil, errors.New("must be an integer")
	default:
		if s, ok := v.(string); ok {
			if tm, err := time.Parse(timeLayout, s); err == nil && tm.Format(timeLayout) == s {
				return tm, nil
			}
		}
		return nil, fmt.Errorf("must be a time as %s", timeLayout)
	}
}

// normalize checks the record r of a change (section 4) and gives it as a read of its line gives it.
func (s *schema) normalize(r Record) (Record, error) {
	k := s.kind(r.Kind)
	if k == nil {
		return Record{}, fmt.Errorf("unknown kind %q", r.Kind)
	}
	line, err := k.encode(r)
	if err != nil {
		return Record{}, err
	}
	fields, err := parseLine(line)
	if err != nil {
		return Record{}, err
	}
	rec, errs := s.record(fields)
	if len(errs) > 0 {
		return Record{}, errors.New(strings.Join(errs, "; "))
	}
	return rec, nil
}

// limit drops the oldest records of each kind over its limit (section 3).
func (s *schema) limit(recs map[string][]Record) {
	for _, k := range s.kinds {
		if rs := recs[k.name]; len(rs) > k.limit {
			recs[k.name] = rs[len(rs)-k.limit:]
		}
	}
}

// encode gives the file of recs: the header and the records grouped by kind (sections 2 and 5).
func (s *schema) encode(recs map[string][]Record) ([]byte, error) {
	b := fmt.Appendf(nil, `{"format":%q,"version":%d}`+"\n", format, s.version())
	for i := range s.kinds {
		k := &s.kinds[i]
		for _, r := range recs[k.name] {
			line, err := k.encode(r)
			if err != nil {
				return nil, err
			}
			b = append(append(b, line...), '\n')
		}
	}
	if len(b) > MaxFileBytes {
		return nil, errTooLarge
	}
	return b, nil
}

// encode gives the line of the record r of k with no line end: kind first,
// then the fields in their order. A time is in UTC with no fraction.
func (k *kind) encode(r Record) ([]byte, error) {
	for _, n := range slices.Sorted(maps.Keys(r.Fields)) {
		if n == "kind" || k.field(n) == nil {
			return nil, fmt.Errorf("unknown field %q", n)
		}
	}
	b := appendString([]byte(`{"kind":`), k.name)
	for _, f := range k.fields {
		v, ok := r.Fields[f.name]
		if !ok {
			return nil, fmt.Errorf("missing field %q", f.name)
		}
		b = append(appendString(append(b, ','), f.name), ':')
		switch f.typ {
		case typeString:
			s, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("field %q: must be a string", f.name)
			}
			if !utf8.ValidString(s) {
				return nil, fmt.Errorf("field %q: invalid UTF-8", f.name)
			}
			b = appendString(b, s)
		case typeInt:
			i, ok := v.(int64)
			if !ok {
				return nil, fmt.Errorf("field %q: must be an int64", f.name)
			}
			b = strconv.AppendInt(b, i, 10)
		default:
			t, ok := v.(time.Time)
			if !ok {
				return nil, fmt.Errorf("field %q: must be a time.Time", f.name)
			}
			b = appendString(b, t.UTC().Format(timeLayout))
		}
	}
	return append(b, '}'), nil
}

// appendString appends the JSON string of s. It does not escape <, > and &,
// and it keeps non-ASCII text as UTF-8 except U+2028 and U+2029 (section 5).
func appendString(b []byte, s string) []byte {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(s) // a string always encodes
	return append(b, bytes.TrimSuffix(buf.Bytes(), []byte("\n"))...)
}
