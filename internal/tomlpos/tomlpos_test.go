package tomlpos

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func TestPos(t *testing.T) {
	data := strings.Join([]string{
		`width = [64, 100]`,           // 1
		`[fields]`,                    // 2
		`  w = "term.cols"`,           // 3
		`[[item]]`,                    // 4
		`key = "A"`,                   // 5
		`[[item]]`,                    // 6
		`key = "B"`,                   // 7
		`[item.sub]`,                  // 8
		`x = 1`,                       // 9
		`inline = { a.b = 1, c = 2 }`, // 10
		`list = [ {k = 1}, 2 ]`,       // 11
		`"quoted key" = 3`,            // 12
	}, "\r\n")
	d, err := Parse([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		path []string
		want Pos
	}{
		{[]string{"width"}, Pos{1, 1}},
		{[]string{"width", "1"}, Pos{1, 14}},
		{[]string{"fields"}, Pos{2, 2}},
		{[]string{"fields", "w"}, Pos{3, 3}},
		{[]string{"fields", "missing"}, Pos{2, 2}},
		{[]string{"item"}, Pos{4, 3}},
		{[]string{"item", "0", "key"}, Pos{5, 1}},
		{[]string{"item", "1"}, Pos{6, 3}},
		{[]string{"item", "1", "key"}, Pos{7, 1}},
		{[]string{"item", "1", "sub", "x"}, Pos{9, 1}},
		{[]string{"item", "1", "sub", "inline", "a"}, Pos{10, 12}},
		{[]string{"item", "1", "sub", "inline", "a", "b"}, Pos{10, 14}},
		{[]string{"item", "1", "sub", "inline", "c"}, Pos{10, 21}},
		{[]string{"item", "1", "sub", "list", "0", "k"}, Pos{11, 11}},
		{[]string{"item", "1", "sub", "list", "1"}, Pos{11, 19}},
		{[]string{"item", "1", "sub", "quoted key"}, Pos{12, 1}},
		{[]string{"none"}, Pos{1, 1}},
	}
	for _, tt := range tests {
		if got := d.Pos(tt.path...); got != tt.want {
			t.Errorf("Pos(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
	items := d.Values["item"].([]any)
	if len(items) != 2 || items[1].(map[string]any)["key"] != "B" {
		t.Errorf("item = %v", items)
	}
}

// A table that only a dotted key or a dotted header defines has the position
// of its key part.
func TestImplicitTables(t *testing.T) {
	data := "a = 1\nserver.port = 1\n[foo.bar]\nx = 1\n[[p.q]]\n[[p.q]]\n"
	d, err := Parse([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		path []string
		want Pos
	}{
		{[]string{"server"}, Pos{2, 1}},
		{[]string{"server", "port"}, Pos{2, 8}},
		{[]string{"foo"}, Pos{3, 2}},
		{[]string{"foo", "bar"}, Pos{3, 6}},
		{[]string{"foo", "bar", "x"}, Pos{4, 1}},
		{[]string{"p"}, Pos{5, 3}},
		{[]string{"p", "q", "1"}, Pos{6, 3}},
	}
	for _, tt := range tests {
		if got := d.Pos(tt.path...); got != tt.want {
			t.Errorf("Pos(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

// The memory of Parse grows linearly with the depth of nesting.
// A quadratic growth gives a ratio near 16 for 4 times the depth.
func TestDeepNesting(t *testing.T) {
	docs := map[string]func(int) string{
		"arrays": func(n int) string {
			return "a = " + strings.Repeat("[", n) + strings.Repeat("]", n)
		},
		"dotted header": func(n int) string {
			return "[" + strings.Repeat("a.", n) + "a]\n" + keys(n)
		},
	}
	for name, doc := range docs {
		small, large := allocBytes(t, doc(1000)), allocBytes(t, doc(4000))
		if ratio := float64(large) / float64(small); ratio > 8 {
			t.Errorf("%s: 4 times the depth uses %.1f times the memory", name, ratio)
		}
	}
}

// keys returns n lines with different keys.
func keys(n int) string {
	var b strings.Builder
	for i := range n {
		fmt.Fprintf(&b, "k%d = 1\n", i)
	}
	return b.String()
}

func allocBytes(t *testing.T, doc string) uint64 {
	t.Helper()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	if _, err := Parse([]byte(doc)); err != nil {
		t.Fatal(err)
	}
	runtime.ReadMemStats(&after)
	return after.TotalAlloc - before.TotalAlloc
}

func TestParseError(t *testing.T) {
	tests := []struct {
		data string
		want Pos
	}{
		{"a = 1\nb = \n", Pos{2, 5}},
		{"a = 1\na = 2\n", Pos{2, 1}},
		{"[t]\n[t]\n", Pos{2, 2}},
	}
	for _, tt := range tests {
		_, err := Parse([]byte(tt.data))
		var e *Error
		if !errors.As(err, &e) {
			t.Errorf("Parse(%q) = %v, want an *Error", tt.data, err)
			continue
		}
		if e.Pos != tt.want || e.Msg == "" || strings.HasPrefix(e.Msg, "toml:") {
			t.Errorf("Parse(%q) = %v at %v, want a message at %v", tt.data, e.Msg, e.Pos, tt.want)
		}
	}
}

func TestEmpty(t *testing.T) {
	d, err := Parse(nil)
	if err != nil || d.Values == nil || len(d.Values) != 0 {
		t.Errorf("Parse(nil) = %v, %v", d, err)
	}
}
