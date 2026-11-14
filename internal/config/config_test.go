package config

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"newbbs/internal/screen"
)

func style(t *testing.T, text string) screen.Style {
	t.Helper()
	spec, err := screen.ParseStyle(text)
	if err != nil {
		t.Fatal(err)
	}
	return spec.Over(screen.Style{})
}

func messages(ds screen.Diags) []string {
	var s []string
	for _, d := range ds {
		s = append(s, d.String())
	}
	return s
}

func TestDefaults(t *testing.T) {
	cfg, ds := Parse(nil)
	if len(ds) != 0 {
		t.Fatalf("diags: %q", messages(ds))
	}
	if cfg.Width != (screen.WidthLimits{Min: 40, Max: 80}) || cfg.More != "-- More --" {
		t.Errorf("cfg = %+v", cfg)
	}
	if len(cfg.Theme) != 2 || cfg.Theme["normal"] != style(t, "7/0") || cfg.Theme["more"] != style(t, "15/0") {
		t.Errorf("theme = %v", cfg.Theme)
	}
	if cfg.Start != "main" {
		t.Errorf("start %q", cfg.Start)
	}
}

func TestParse(t *testing.T) {
	data := `
[screen]
width = [64, 132]
more = "[ ещё ]"

[theme]
normal = "14"
title  = "15/4 bold"
prompt = "/1"
more = "-/-"

[menu]
  start = "welcome"

[fields]
w = "term.cols"
`
	cfg, ds := Parse([]byte(data))
	if len(ds) != 0 {
		t.Fatalf("diags: %q", messages(ds))
	}
	if cfg.Width != (screen.WidthLimits{Min: 64, Max: 132}) || cfg.More != "[ ещё ]" {
		t.Errorf("cfg = %+v", cfg)
	}
	if cfg.Start != "welcome" {
		t.Errorf("start %q", cfg.Start)
	}
	if len(cfg.Fields) != 1 || cfg.Fields["w"].String() != "term.cols" {
		t.Errorf("fields %v", cfg.Fields)
	}
	want := map[string]string{
		"normal": "14/0",      // the background from the default normal
		"title":  "15/4 bold", // the example of section 7.3
		"prompt": "14/1",      // the foreground from normal
		"more":   "-/-",
	}
	if len(cfg.Theme) != len(want) {
		t.Errorf("theme = %v", cfg.Theme)
	}
	for name, text := range want {
		if got := cfg.Theme[name]; got != style(t, text) {
			t.Errorf("theme %s = %v, want %s", name, got, text)
		}
	}
}

// A theme style takes the colours of normal, not its attributes.
func TestNormalAttributes(t *testing.T) {
	cfg, ds := Parse([]byte("[theme]\nnormal = \"2/3 bold\"\nx = \"5\"\n"))
	if len(ds) != 0 {
		t.Fatalf("diags: %q", messages(ds))
	}
	if cfg.Theme["normal"] != style(t, "2/3 bold") || cfg.Theme["x"] != style(t, "5/3") {
		t.Errorf("theme = %v", cfg.Theme)
	}
}

// Each other error has the position 1:1 (section 9.1).
func TestErrors(t *testing.T) {
	data := strings.Join([]string{
		`[screen]`,
		`width = [80, 40]`,
		`more = "a` + "\u200b" + `"`,
		`[theme]`,
		`Title = "1"`,
		`normal = "7/0"`,
	}, "\n")
	cfg, ds := Parse([]byte(data))
	want := []string{
		`bbs.toml:1:1: screen.width: width [80, 40] is not in 1 <= min <= max <= 255`,
		`bbs.toml:1:1: screen.more: zero-width grapheme U+200B`,
		`bbs.toml:1:1: theme: invalid style name "Title"`,
	}
	if got := messages(ds); !slices.Equal(got, want) {
		t.Errorf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	if !ds.HasErrors() {
		t.Error("HasErrors = false")
	}
	if cfg.Width != (screen.WidthLimits{Min: 40, Max: 80}) || cfg.More != "-- More --" {
		t.Errorf("an invalid value replaced a default: %+v", cfg)
	}
}

// A TOML document gives a syntax or type error; else all unknown keys;
// else its other errors (section 9.1).
func TestTOMLErrors(t *testing.T) {
	tests := []struct {
		data string
		want []string
	}{
		{"screen = 1", []string{`bbs.toml:1:10: screen: must be a table`}},
		{"theme = []", []string{`bbs.toml:1:9: theme: must be a table of strings`}},
		{"menu = \"main\"", []string{`bbs.toml:1:8: menu: must be a table`}},
		{"[menu]\nstart = 1", []string{`bbs.toml:2:9: menu.start: must be a string`}},
		{"[screen]\nwidth = [1]", []string{`bbs.toml:1:1: screen.width: must be an array of 2 integers`}},
		{"[screen]\nwidth = [1, \"2\"]", []string{`bbs.toml:2:13: screen.width: must be an array of integers`}},
		{"[screen]\nmore = 1", []string{`bbs.toml:2:8: screen.more: must be a string`}},
		{"[screen]\nmore = \"\t\"", []string{`bbs.toml:1:1: screen.more: control character U+0009`}},
		{"[screen]\nwidth = ", []string{`bbs.toml:2:8: expected value, not end of input`}},
		{"[theme]\nprompt = \"14 bolt\"", []string{`bbs.toml:1:1: theme.prompt: invalid style "14 bolt": unknown attribute "bolt"`}},
		{"[theme]\nprompt = 14", []string{`bbs.toml:2:10: theme.prompt: must be a string`}},
		{"a.b = 1", []string{`bbs.toml:1:1: unknown key "a.b"`}},
		{"[TELNET]\nlisten = \":1\"", []string{`bbs.toml:1:1: unknown key "TELNET"`}},
		{"[telnet]\nlisten = \":1\"\nLISTEN = \":2\"", []string{`bbs.toml:1:1: unknown key "telnet.LISTEN"`}},
		{"\n[foo.bar]\nx = 1", []string{`bbs.toml:2:2: unknown key "foo.bar"`}},
		// A type error comes before an unknown key, an unknown key before an other error.
		{"[theme]\nx = \"bad\"\ny = 1\nz = \"bad\"", []string{`bbs.toml:3:5: theme.y: must be a string`}},
		{"a = 1\n[theme]\nx = \"bad\"\nz = \"bad\"", []string{`bbs.toml:1:1: unknown key "a"`}},
		{"[theme]\nx = \"bad\"\nz = \"bad\"", []string{
			`bbs.toml:1:1: theme.x: invalid style "bad": invalid colour "bad"`,
			`bbs.toml:1:1: theme.z: invalid style "bad": invalid colour "bad"`,
		}},
		{"a = 1\n[screen]\nb = 2\nwidth = [1, 1]", []string{`bbs.toml:1:1: unknown key "a"`, `bbs.toml:3:1: unknown key "screen.b"`}},
		{"[fields]\nw = 1", []string{`bbs.toml:2:5: fields.w: must be a string`}},
		// Each invalid alias is an other error at 1:1 (section 4.2).
		{"fields = { \"яя\" = \"term.cols\", and = \"term.rows\", k = \"item.A\", u = \"term.x\", v = \"w\" }", []string{
			`bbs.toml:1:1: fields: alias "and" is a reserved word`,
			`bbs.toml:1:1: fields: alias "k": "item.A" is valid only in {#if}`,
			`bbs.toml:1:1: fields: alias "u": unknown field "term.x"`,
			`bbs.toml:1:1: fields: alias "v": "w" is not a registry path`,
			`bbs.toml:1:1: fields: invalid alias "яя"`,
		}},
	}
	for _, tt := range tests {
		_, ds := Parse([]byte(tt.data))
		if got := messages(ds); !slices.Equal(got, tt.want) {
			t.Errorf("Parse(%q) = %q, want %q", tt.data, got, tt.want)
		}
	}
}

// The loader removes a BOM (section 2).
func TestBOM(t *testing.T) {
	cfg, ds := Parse([]byte("\ufeff[menu]\nstart = \"x\"\n"))
	if len(ds) != 0 || cfg.Start != "x" {
		t.Errorf("start %q, diags %q", cfg.Start, messages(ds))
	}
}

func TestDiagLimit(t *testing.T) {
	var b strings.Builder
	for i := range 2 * screen.MaxFileDiags {
		fmt.Fprintf(&b, "k%d = 1\n", i)
	}
	_, ds := Parse([]byte(b.String()))
	got := messages(ds)
	if len(got) != screen.MaxFileDiags+1 || got[0] != `bbs.toml:1:1: unknown key "k0"` ||
		got[len(got)-1] != "bbs.toml:101:1: too many diagnostics" {
		t.Errorf("%d diags, first %q, last %q", len(got), got[0], got[len(got)-1])
	}
}
