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
	cfg, ds := Parse("bbs.toml", nil)
	if len(ds) != 0 {
		t.Fatalf("diags: %q", messages(ds))
	}
	if cfg.Width != (screen.WidthLimits{Min: 40, Max: 80}) || cfg.More != "-- More --" {
		t.Errorf("cfg = %+v", cfg)
	}
	if len(cfg.Theme) != 2 || cfg.Theme["normal"] != style(t, "7/0") || cfg.Theme["more"] != style(t, "15/0") {
		t.Errorf("theme = %v", cfg.Theme)
	}
	if cfg.Start != "main" || cfg.StartPos != (screen.Pos{File: "bbs.toml", Line: 1, Col: 1}) {
		t.Errorf("start %q at %v", cfg.Start, cfg.StartPos)
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
`
	cfg, ds := Parse("bbs.toml", []byte(data))
	if len(ds) != 0 {
		t.Fatalf("diags: %q", messages(ds))
	}
	if cfg.Width != (screen.WidthLimits{Min: 64, Max: 132}) || cfg.More != "[ ещё ]" {
		t.Errorf("cfg = %+v", cfg)
	}
	if cfg.Start != "welcome" || cfg.StartPos != (screen.Pos{File: "bbs.toml", Line: 13, Col: 3}) {
		t.Errorf("start %q at %v", cfg.Start, cfg.StartPos)
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
	cfg, ds := Parse("bbs.toml", []byte("[theme]\nnormal = \"2/3 bold\"\nx = \"5\"\n"))
	if len(ds) != 0 {
		t.Fatalf("diags: %q", messages(ds))
	}
	if cfg.Theme["normal"] != style(t, "2/3 bold") || cfg.Theme["x"] != style(t, "5/3") {
		t.Errorf("theme = %v", cfg.Theme)
	}
}

func TestErrors(t *testing.T) {
	data := strings.Join([]string{
		`server = 1`,                 // 1
		`[screen]`,                   // 2
		`width = [80, 40]`,           // 3
		`more = "a` + "\u200b" + `"`, // 4
		`rows = 25`,                  // 5
		`[theme]`,                    // 6
		`Title = "1"`,                // 7
		`prompt = "14 bolt"`,         // 8
		`menu = 3`,                   // 9
		`normal = "7/0"`,             // 10
	}, "\n")
	cfg, ds := Parse("bbs.toml", []byte(data))
	want := []string{
		`bbs.toml:1:1: unknown key "server"`,
		`bbs.toml:3:1: screen.width: width [80, 40] is not in 1 <= min <= max <= 255`,
		`bbs.toml:4:1: screen.more: zero-width grapheme U+200B`,
		`bbs.toml:5:1: unknown key "screen.rows"`,
		`bbs.toml:7:1: theme: invalid style name "Title"`,
		`bbs.toml:8:1: theme.prompt: invalid style "14 bolt": unknown attribute "bolt"`,
		`bbs.toml:9:1: theme.menu: must be a string`,
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

func TestTypeErrors(t *testing.T) {
	tests := []struct {
		data string
		want string
	}{
		{"screen = 1", `bbs.toml:1:1: screen: must be a table`},
		{"theme = []", `bbs.toml:1:1: theme: must be a table`},
		{"menu = \"main\"", `bbs.toml:1:1: menu: must be a table`},
		{"[menu]\nstart = 1", `bbs.toml:2:1: menu: start: must be a string`},
		{"[menu]\nfirst = \"main\"", `bbs.toml:2:1: unknown key "menu.first"`},
		{"[screen]\nwidth = [1]", `bbs.toml:2:1: screen.width: must be an array of 2 integers`},
		{"[screen]\nwidth = [1, \"2\"]", `bbs.toml:2:1: screen.width: must be an array of 2 integers`},
		{"[screen]\nmore = 1", `bbs.toml:2:1: screen.more: must be a string`},
		{"[screen]\nmore = \"\t\"", `bbs.toml:2:1: screen.more: control character U+0009`},
		{"[screen]\nwidth = ", `bbs.toml:2:8: expected value, not end of input`},
		{"a.b = 1", `bbs.toml:1:1: unknown key "a"`},
		{"\n[foo.bar]\nx = 1", `bbs.toml:2:2: unknown key "foo"`},
	}
	for _, tt := range tests {
		_, ds := Parse("bbs.toml", []byte(tt.data))
		if got := messages(ds); !slices.Equal(got, []string{tt.want}) {
			t.Errorf("Parse(%q) = %q, want %q", tt.data, got, tt.want)
		}
	}
}

func TestDiagLimit(t *testing.T) {
	var b strings.Builder
	for i := range 2 * screen.MaxFileDiags {
		fmt.Fprintf(&b, "k%d = 1\n", i)
	}
	_, ds := Parse("bbs.toml", []byte(b.String()))
	got := messages(ds)
	if len(got) != screen.MaxFileDiags+1 || got[0] != `bbs.toml:1:1: unknown key "k0"` ||
		got[len(got)-1] != "bbs.toml:101:1: too many diagnostics" {
		t.Errorf("%d diags, first %q, last %q", len(got), got[0], got[len(got)-1])
	}
}
