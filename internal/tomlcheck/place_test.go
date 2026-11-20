package tomlcheck

import "testing"

func TestPlace(t *testing.T) {
	data := []byte(`# a comment
name = "A"
size = [ 1, 2 ]
"q.k" = 1
[table]
n = 3
a.b.c = "X"
[table.sub]
[up.down]
x = 1
[up]
[[item]]
key = "B"
[[item]]
  key = 'C'
  in = { k = { m = "D" }, l = [7] }
[ map ]
list = [{ k = "E" }, { k = "F" }]
nest = [[1],
 [2]]
`)
	root := Root(data)
	tests := []struct {
		name      string
		at        Place
		key, want Pos
	}{
		// A value at its first byte, an array at its [.
		{"value", root.At("name"), Pos{2, 1}, Pos{2, 8}},
		{"array", root.At("size"), Pos{3, 1}, Pos{3, 8}},
		{"array element", root.At("size").Elem(1), Pos{3, 1}, Pos{3, 13}},
		{"nested array", root.At("map", "nest").Elem(1), Pos{19, 1}, Pos{19, 8}},
		{"element of a nested array", root.At("map", "nest").Elem(1).Elem(0), Pos{19, 1}, Pos{20, 3}},
		{"quoted key", root.At("q.k"), Pos{4, 1}, Pos{4, 9}},
		{"value in a table", root.At("table", "n"), Pos{6, 1}, Pos{6, 5}},
		// A table at the [ of its header.
		{"table", root.At("table"), Pos{5, 2}, Pos{5, 1}},
		{"dotted header", root.At("table", "sub"), Pos{8, 8}, Pos{8, 1}},
		{"header with spaces", root.At("map"), Pos{17, 3}, Pos{17, 1}},
		// A table at the first dotted key that makes it, else at its header.
		{"dotted key", root.At("table", "a"), Pos{7, 1}, Pos{7, 1}},
		{"dotted key 2", root.At("table", "a", "b"), Pos{7, 3}, Pos{7, 3}},
		{"dotted key value", root.At("table", "a", "b", "c"), Pos{7, 5}, Pos{7, 9}},
		{"implicit table with a header", root.At("up"), Pos{9, 2}, Pos{11, 1}},
		// An array of tables: each table at its header.
		{"array of tables", root.At("item"), Pos{12, 3}, Pos{12, 1}},
		{"table of an array", root.At("item").Elem(1), Pos{12, 3}, Pos{14, 1}},
		{"key in a table of an array", root.At("item").Elem(1).At("key"), Pos{15, 3}, Pos{15, 9}},
		// An inline table at its {.
		{"inline table", root.At("item").Elem(1).At("in"), Pos{16, 3}, Pos{16, 8}},
		{"nested inline table", root.At("item").Elem(1).At("in", "k", "m"), Pos{16, 16}, Pos{16, 20}},
		{"array in an inline table", root.At("item").Elem(1).At("in", "l"), Pos{16, 27}, Pos{16, 31}},
		{"inline tables in an array", root.At("map", "list").Elem(1), Pos{18, 1}, Pos{18, 22}},
		{"key in an inline table of an array", root.At("map", "list").Elem(1).At("k"), Pos{18, 24}, Pos{18, 28}},
		// A key that is not set at the last table on the path that the document sets.
		{"missing key", root.At("table", "x"), Pos{5, 1}, Pos{5, 1}},
		{"missing table", root.At("table", "x", "y"), Pos{5, 1}, Pos{5, 1}},
		{"missing key in a dotted table", root.At("table", "a", "z"), Pos{7, 1}, Pos{7, 1}},
		{"missing key in a table of an array", root.At("item").Elem(0).At("do"), Pos{12, 1}, Pos{12, 1}},
		{"missing table of an array", root.At("item").Elem(2), Pos{1, 1}, Pos{1, 1}},
		{"path into an array of tables", root.At("item", "key"), Pos{1, 1}, Pos{1, 1}},
		{"path past a value", root.At("table", "n", "x"), Pos{5, 1}, Pos{5, 1}},
		{"missing key in an inline table", root.At("item").Elem(1).At("in", "k", "x"), Pos{16, 14}, Pos{16, 14}},
		// The root table and a key that is not set in it at 1:1.
		{"root", root, Pos{1, 1}, Pos{1, 1}},
		{"missing key in the root", root.At("menu", "start"), Pos{1, 1}, Pos{1, 1}},
	}
	for _, tt := range tests {
		if got := tt.at.Key(); got != tt.key {
			t.Errorf("%s: key %v, want %v", tt.name, got, tt.key)
		}
		if got := tt.at.Value(); got != tt.want {
			t.Errorf("%s: value %v, want %v", tt.name, got, tt.want)
		}
	}
}

// The whole document has the position 1:1, also when it is empty.
func TestPlaceEmpty(t *testing.T) {
	for _, data := range []string{"", "\n\n", "# x"} {
		root := Root([]byte(data))
		if got := root.At("menu", "start").Value(); got != (Pos{1, 1}) {
			t.Errorf("%q: %v, want 1:1", data, got)
		}
	}
}

// Decode gives the root of the document.
func TestDecodeRoot(t *testing.T) {
	var d doc
	root, errs := Decode([]byte("[map]\nk = \"B\""), &d)
	if errs != nil {
		t.Fatalf("errors %v", errs)
	}
	if got := root.At("map", "k").Value(); got != (Pos{2, 5}) {
		t.Errorf("map.k: %v, want 2:5", got)
	}
	if root, _ := Decode([]byte("x = "), &d); root != (Place{}) {
		t.Errorf("syntax error: root %+v, want the zero Place", root)
	}
}
