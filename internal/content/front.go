package content

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/pelletier/go-toml/v2/unstable"
	yaml "go.yaml.in/yaml/v4"

	"github.com/schors/velourbbs/internal/screen"
	"github.com/schors/velourbbs/internal/tomlcheck"
)

// MaxFrontBytes is the size limit of a front matter (section 8).
const MaxFrontBytes = 64 << 10

// parseFront finds and parses the front matter of a Markdown file (section 5.1).
// It gives the meta fields and the lines of the body. ok is false after an error.
func parseFront(d *screen.FileDiags, lines []string) (meta map[string]string, body []string, ok bool) {
	if len(lines) == 0 {
		return nil, lines, true
	}
	var parse func(*screen.FileDiags, []byte) (map[string]string, bool)
	end := -1
	switch {
	case isDelim(lines[0], "+++"):
		end = findDelim(lines, "+++")
		if end < 0 {
			d.Errorf(1, 1, `front matter: no end line "+++"`)
			return nil, nil, false
		}
		parse = parseTOML
	case isDelim(lines[0], "---") && len(lines) > 1 && lines[1] != "":
		end = findDelim(lines, "---", "...")
		parse = parseYAML
	}
	if end < 0 {
		return nil, lines, true
	}
	data := []byte(strings.Join(lines[1:end], "\n"))
	if len(data) > MaxFrontBytes {
		d.Errorf(1, 1, "front matter: over %d bytes", MaxFrontBytes)
		return nil, nil, false
	}
	meta, ok = parse(d, data)
	return meta, lines[end+1:], ok
}

// isDelim reports whether line is the delimiter line delim. It can have spaces at its end.
func isDelim(line, delim string) bool {
	return strings.TrimRight(line, " ") == delim
}

// findDelim gives the index of the first line after the first line that is one of delims, or -1.
func findDelim(lines []string, delims ...string) int {
	for i := 1; i < len(lines); i++ {
		for _, delim := range delims {
			if isDelim(lines[i], delim) {
				return i
			}
		}
	}
	return -1
}

// frontErrorf reports an error at the line of the front matter data. Line 1 of data is line 2 of the file.
func frontErrorf(d *screen.FileDiags, line, col int, format string, args ...any) {
	d.Errorf(line+1, col, "front matter: "+format, args...)
}

// parseTOML parses a TOML front matter. Each top-level key with a scalar value gives a meta field.
func parseTOML(d *screen.FileDiags, data []byte) (map[string]string, bool) {
	m, errs := tomlcheck.Map(data)
	for _, e := range errs {
		frontErrorf(d, e.Line, e.RuneCol(data), "%s", e.Msg)
	}
	if errs != nil {
		return nil, false
	}
	numbers := tomlNumbers(data)
	meta := map[string]string{}
	for k, v := range m {
		if !screen.IsIdent(k) {
			continue
		}
		switch v := v.(type) {
		case string:
			meta[k] = v
		case bool:
			meta[k] = strconv.FormatBool(v)
		case int64, float64:
			meta[k] = numbers[k]
		case toml.LocalDate:
			meta[k] = v.String()
		case toml.LocalDateTime:
			meta[k] = v.String()
		case toml.LocalTime:
			meta[k] = v.String()
		case time.Time:
			meta[k] = v.Format(time.RFC3339Nano)
		}
	}
	return meta, true
}

// tomlNumbers gives the integers and floats of the top-level keys of a valid TOML document
// as the document writes them. The decode gives only their values.
func tomlNumbers(data []byte) map[string]string {
	numbers := map[string]string{}
	var p unstable.Parser
	p.Reset(data)
	for p.NextExpression() {
		e := p.Expression()
		if e.Kind == unstable.Table || e.Kind == unstable.ArrayTable {
			break
		}
		if e.Kind != unstable.KeyValue {
			continue
		}
		key := e.Key()
		if !key.Next() {
			continue
		}
		name := string(key.Node().Data)
		if key.Next() {
			continue // a dotted key sets a table
		}
		if v := e.Value(); v.Kind == unstable.Integer || v.Kind == unstable.Float {
			numbers[name] = string(v.Data)
		}
	}
	return numbers
}

// parseYAML parses a YAML front matter. Each top-level key with a scalar value gives a meta field.
func parseYAML(d *screen.FileDiags, data []byte) (map[string]string, bool) {
	var docs []yaml.Node
	if err := yaml.Load(data, &docs, yaml.WithAllDocuments()); err != nil {
		line, col, msg := 1, 1, err.Error()
		var le *yaml.LoadError
		if errors.As(err, &le) {
			msg = le.Message
			if le.Mark.Line > 0 {
				line, col = le.Mark.Line, max(le.Mark.Column, 1)
			}
		}
		frontErrorf(d, line, col, "%s", msg)
		return nil, false
	}
	meta := map[string]string{}
	switch {
	case len(docs) == 0:
		return meta, true // only comments
	case len(docs) > 1:
		frontErrorf(d, docs[1].Line, 1, "more than one YAML document")
		return nil, false
	}
	top := docs[0].Content[0]
	if top.Kind != yaml.MappingNode {
		frontErrorf(d, top.Line, top.Column, "not a mapping")
		return nil, false
	}
	ok := true
	seen := map[string]bool{}
	for i := 0; i+1 < len(top.Content); i += 2 {
		k, v := top.Content[i], top.Content[i+1]
		if k.Kind != yaml.ScalarNode {
			continue
		}
		if seen[k.Value] {
			frontErrorf(d, k.Line, k.Column, "duplicate key %q", k.Value)
			ok = false
			continue
		}
		seen[k.Value] = true
		if v.Kind == yaml.AliasNode {
			v = v.Alias
		}
		s, isField, err := yamlValue(v)
		switch {
		case err != nil:
			frontErrorf(d, v.Line, v.Column, "%s", err)
			ok = false
		case isField && screen.IsIdent(k.Value):
			meta[k.Value] = s
		}
	}
	if !ok {
		return nil, false
	}
	return meta, true
}

// yamlValue gives the meta field of the YAML node v by its tag (section 5.1).
// isField is false for a value that gives no field. The node load does not check
// a value with an explicit tag, for example "!!int abc"; the decode does.
func yamlValue(v *yaml.Node) (s string, isField bool, err error) {
	if v.Kind != yaml.ScalarNode {
		return "", false, nil
	}
	tag := v.ShortTag()
	switch tag {
	case "!!int", "!!float", "!!bool", "!!timestamp":
		var x any
		if err := v.Decode(&x); err != nil {
			var le *yaml.LoadError
			if errors.As(err, &le) {
				return "", false, errors.New(le.Message)
			}
			return "", false, err
		}
		if b, ok := x.(bool); ok {
			return strconv.FormatBool(b), true, nil
		}
	}
	switch tag {
	case "!!str", "!!int", "!!float":
		return v.Value, true, nil
	case "!!timestamp":
		s, ok := yamlTime(v.Value)
		return s, ok, nil
	}
	return "", false, nil
}

// The timestamp forms of the yaml resolver, and the form of the meta field for each.
var yamlTimes = []struct{ in, out string }{
	{"2006-1-2T15:4:5.999999999Z07:00", time.RFC3339Nano},
	{"2006-1-2t15:4:5.999999999Z07:00", time.RFC3339Nano},
	{"2006-1-2 15:4:5.999999999", "2006-01-02T15:04:05.999999999"},
	{"2006-1-2", time.DateOnly},
}

func yamlTime(s string) (string, bool) {
	for _, f := range yamlTimes {
		if t, err := time.Parse(f.in, s); err == nil {
			return t.Format(f.out), true
		}
	}
	return "", false
}
