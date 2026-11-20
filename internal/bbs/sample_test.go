package bbs

import (
	"maps"
	"strings"
	"testing"
)

// The sample values follow the {arg.<name>} values and skip an argument that is already
// on the path (screen-render, section 3.4).
func TestSamples(t *testing.T) {
	// item gives an item of the key k with the do and the args.
	item := func(k, do, args string) string {
		return "[[item]]\nkey = \"" + k + "\"\ndo = \"" + do + "\"\nargs = { " + args + " }\n"
	}
	screen := func(items ...string) string {
		return "+++\n" + strings.Join(items, "") + "[[item]]\nkey = \"ESC\"\ndo = \"back\"\n+++\n{arg.x}\n"
	}
	m, msgs := loadLangs(t, nil, map[string]string{
		// a and b give x to each other: the first action to a is the cycle.
		"screens/a.bbt": screen(item("1", "gosub b", `x = "{arg.x}"`)),
		"screens/b.bbt": screen(item("1", "gosub a", `x = "{arg.x}"`), item("2", "gosub c", `y = "{arg.x}"`)),
		"screens/c.bbt": "+++\n[[item]]\nkey = \"ESC\"\ndo = \"back\"\n+++\n{arg.y}\n",
		// d and e give z only to each other: z has no value.
		"screens/d.bbt":    "+++\n" + item("1", "gosub e", `z = "{arg.z}"`) + "+++\n{arg.z}\n",
		"screens/e.bbt":    "+++\n" + item("1", "gosub d", `z = "{arg.z}"`) + "+++\n{arg.z}\n",
		"screens/main.bbt": "+++\n" + item("1", "gosub a", `x = "m"`) + "[[item]]\nkey = \"Q\"\ndo = \"logoff\"\n+++\n",
	})
	if m == nil {
		t.Fatalf("diags:\n%s", strings.Join(msgs, "\n"))
	}
	s := NewSamples(m)
	for _, tt := range []struct {
		screen, arg, want string
		ok                bool
	}{
		{"a", "x", "m", true},
		{"b", "x", "m", true},
		{"c", "y", "m", true},
		{"d", "z", "", false},
	} {
		if v, ok := s.Arg(tt.screen, tt.arg); v != tt.want || ok != tt.ok {
			t.Errorf("Arg(%s, %s) = %q, %v; want %q, %v", tt.screen, tt.arg, v, ok, tt.want, tt.ok)
		}
	}
	if got, want := s.Fields(), map[string]string{"x": "m", "y": "m"}; !maps.Equal(got, want) {
		t.Errorf("Fields() = %v, want %v", got, want)
	}
}
