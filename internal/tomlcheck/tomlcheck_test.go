package tomlcheck

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
)

type doc struct {
	Name  *string           `toml:"name"`
	Size  *[]int64          `toml:"size"`
	Tags  []string          `toml:"tags"`
	Map   map[string]string `toml:"map"`
	Table struct {
		N int64 `toml:"n"`
	} `toml:"table"`
	Item []struct {
		Key *string `toml:"key"`
	} `toml:"item"`
}

func decode(data string) (doc, []string) {
	var d doc
	var msgs []string
	for _, e := range Decode([]byte(data), &d) {
		msgs = append(msgs, fmt.Sprintf("%d:%d: %s", e.Line, e.Col, e.Msg))
	}
	return d, msgs
}

func TestDecode(t *testing.T) {
	d, msgs := decode("name = \"A\"\nsize = [1, 2]\ntags = [\"x\"]\n[map]\nk = \"B\"\n[table]\nn = 3\n[[item]]\nkey = \"C\"\n[[item]]\n")
	if len(msgs) != 0 {
		t.Fatalf("errors %q", msgs)
	}
	if *d.Name != "A" || !slices.Equal(*d.Size, []int64{1, 2}) || d.Map["k"] != "B" || d.Table.N != 3 ||
		len(d.Item) != 2 || *d.Item[0].Key != "C" || d.Item[1].Key != nil {
		t.Errorf("doc %+v", d)
	}
}

func TestErrors(t *testing.T) {
	tests := []struct {
		data string
		want []string
	}{
		{"name = ", []string{`1:7: expected value, not end of input`}},
		{"name = \"A\"\nname = \"B\"", []string{`2:1: key name is already defined`}},
		// A type error names the type of the field.
		{"name = 5", []string{`1:8: name: must be a string`}},
		{"name = true", []string{`1:8: name: must be a string`}},
		{"name = 1.5", []string{`1:8: name: must be a string`}},
		{"name = 1979-05-27", []string{`1:8: name: must be a string`}},
		{"name = {}", []string{`1:8: name: must be a string`}},
		{"[name]", []string{`1:2: name: must be a string`}},
		{"size = 1", []string{`1:8: size: must be an array of integers`}},
		{"size = [1, \"2\"]", []string{`1:12: size: must be an array of integers`}},
		{"tags = [1]", []string{`1:9: tags: must be an array of strings`}},
		{"map = 1", []string{`1:7: map: must be a table of strings`}},
		{"map = [{}]", []string{`1:7: map: must be a table of strings`}},
		{"[map]\nk = 1", []string{`2:5: map.k: must be a string`}},
		{"table = 1", []string{`1:9: table: must be a table`}},
		{"table.n = \"1\"", []string{`1:11: table.n: must be an integer`}},
		{"item = 1", []string{`1:8: item: must be an array of tables`}},
		{"item = [1]", []string{`1:9: item: must be an array of tables`}},
		{"[[item]]\nkey = 1", []string{`2:7: item.key: must be a string`}},
		{"[[item]]\nkey = {}", []string{`2:7: item.key: must be a string`}},
		{"table.n = 99999999999999999999", []string{`1:11: decimal number is too large to fit in a 64-bit signed integer`}},
		// All unknown keys.
		{"x = 1\n[table]\ny = 2\n[z]\n[[item]]\nw = 3", []string{
			`1:1: unknown key "x"`, `3:1: unknown key "table.y"`, `4:2: unknown key "z"`, `6:1: unknown key "item.w"`,
		}},
		// go-toml matches a key also without case and takes [x] as one table of an array:
		// such keys and tables have no position.
		{"NAME = \"A\"", []string{`1:1: unknown key "NAME"`}},
		{"name = \"A\"\nName = \"B\"", []string{`1:1: unknown key "Name"`}},
		{"[TABLE]\nn = 1", []string{`1:1: unknown key "TABLE"`}},
		{"[table]\nN = 1", []string{`1:1: unknown key "table.N"`}},
		{"[[ITEM]]", []string{`1:1: unknown key "ITEM"`}},
		{"[[item]]\nKEY = \"A\"\n[[item]]\nkey = \"B\"\nKey = \"C\"", []string{`1:1: unknown key "item.KEY"`, `1:1: unknown key "item.Key"`}},
		{"x = 1\n[item]\nkey = \"A\"", []string{`1:1: unknown key "x"`, `1:1: item: must be an array of tables`}},
		// A type error at such a key is an unknown key.
		{"NAME = 5", []string{`1:1: unknown key "NAME"`}},
		{"Table = {n = \"1\"}", []string{`1:1: unknown key "Table"`}},
		{"[table]\nN = \"x\"", []string{`1:1: unknown key "table.N"`}},
		{"[[ITEM]]\nkey = 1", []string{`1:1: unknown key "ITEM"`}},
		{"[[item]]\nKEY = 1", []string{`1:1: unknown key "item.KEY"`}},
		// In an inline table the key path ends at the table.
		{"item = [{key = 5}]", []string{`1:16: item: invalid type`}},
		{"item = {key = 5}", []string{`1:8: item: invalid type`}},
		{"table = {n = \"1\"}", []string{`1:14: table: invalid type`}},
		{"map = {k = 1}", []string{`1:12: map: invalid type`}},
		// A key path past a field that is not a table is an error of that field.
		{"name.x = 1", []string{`1:6: name: must be a string`}},
		{"[map]\na.b = \"X\"", []string{`2:3: map.a: must be a string`}},
		// A nested array has no position.
		{"size = [1, [2]]", []string{`1:1: size: must be an array of integers`}},
		{"ééé = 1", []string{`1:1: invalid character at start of key: U+00E9 'é'`}},
		// The order: a syntax error, a type error, all unknown keys.
		{"name = 5\nx = ", []string{`2:4: expected value, not end of input`}},
		{"x = 1\nsize = \"1\"\nname = 5", []string{`2:8: size: must be an array of integers`}},
	}
	for _, tt := range tests {
		_, msgs := decode(tt.data)
		if !slices.Equal(msgs, tt.want) {
			t.Errorf("%q:\n got %q\nwant %q", tt.data, msgs, tt.want)
		}
	}
}

func TestMaxBytes(t *testing.T) {
	data := strings.Repeat("#", MaxBytes)
	if _, msgs := decode(data); len(msgs) != 0 {
		t.Errorf("%d bytes: %q", len(data), msgs)
	}
	want := []string{"1:1: TOML document is over 65536 bytes"}
	if _, msgs := decode(data + "#"); !slices.Equal(msgs, want) {
		t.Errorf("%d bytes: %q, want %q", len(data)+1, msgs, want)
	}
}

// A decode error that is not a *toml.DecodeError has the position 1:1.
func TestOtherError(t *testing.T) {
	want := Error{1, 1, "x"}
	if got := syntaxError(nil, errors.New("x")); got != want {
		t.Errorf("syntaxError: got %+v, want %+v", got, want)
	}
	if got, ok := typeError(errors.New("x"), nil, nil); got != want || !ok {
		t.Errorf("typeError: got %+v, want %+v", got, want)
	}
}
