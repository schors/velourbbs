package content

import (
	"strconv"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/pelletier/go-toml/v2/unstable"

	"github.com/schors/velourbbs/internal/screen"
	"github.com/schors/velourbbs/internal/tomlcheck"
)

// MaxFrontBytes is the size limit of a front matter (section 8).
const MaxFrontBytes = 64 << 10

// parseFront finds and parses the front matter of a Markdown file (section 5.1).
// It gives the meta fields and the lines of the body. ok is false after an error.
func parseFront(d *screen.FileDiags, lines []string) (meta map[string]string, body []string, ok bool) {
	switch {
	case len(lines) > 1 && isDelim(lines[0], "---") && lines[1] != "" && findDelim(lines, "---", "...") > 0:
		d.Errorf(1, 1, `front matter: YAML is not supported, use TOML between "+++" lines`)
		return nil, nil, false
	case len(lines) == 0 || !isDelim(lines[0], "+++"):
		return nil, lines, true
	}
	end := findDelim(lines, "+++")
	if end < 0 {
		d.Errorf(1, 1, `front matter: no end line "+++"`)
		return nil, nil, false
	}
	data := []byte(strings.Join(lines[1:end], "\n"))
	if len(data) > MaxFrontBytes {
		d.Errorf(1, 1, "front matter: over %d bytes", MaxFrontBytes)
		return nil, nil, false
	}
	meta, ok = parseTOML(d, data)
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

// parseTOML parses a TOML front matter. Each top-level key with a scalar value gives a meta field.
func parseTOML(d *screen.FileDiags, data []byte) (map[string]string, bool) {
	m, errs := tomlcheck.MapFront(data)
	for _, e := range errs {
		d.Errorf(e.Line, e.Col, "front matter: %s", e.Msg)
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
