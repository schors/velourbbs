package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/schors/velourbbs/internal/screen"
)

func style(t *testing.T, text string) screen.Style {
	t.Helper()
	spec, err := screen.ParseStyle(text)
	if err != nil {
		t.Fatal(err)
	}
	return spec.Over(screen.Style{})
}

// dark gives the dark theme of colors.
func dark(cfg Config, colors int) map[string]screen.Style {
	return cfg.SessionTheme(ThemeDark, screen.Term{Colors: colors})
}

// light gives the light theme of colors.
func light(cfg Config, colors int) map[string]screen.Style {
	return cfg.SessionTheme(ThemeLight, screen.Term{Colors: colors})
}

func messages(ds screen.Diags) []string {
	var s []string
	for _, d := range ds {
		s = append(s, d.String())
	}
	return s
}

// A missing screens/screens.toml gives the defaults with no diagnostics (screens-toml, section 2).
func TestDefaults(t *testing.T) {
	cfg, ds := Load(fstest.MapFS{"bbs.toml": {}})
	if len(ds) != 0 {
		t.Fatalf("diags: %q", messages(ds))
	}
	if cfg.Width != (screen.WidthLimits{Min: 40, Max: 80}) || cfg.More != nil || cfg.MoreText("en") != "-- More --" {
		t.Errorf("cfg = %+v", cfg)
	}
	if cfg.Lang != "en" || !slices.Equal(cfg.Langs, []string{"en"}) || len(cfg.LangMore) != 0 {
		t.Errorf("lang %q, langs %q, lang more %v", cfg.Lang, cfg.Langs, cfg.LangMore)
	}
	if cfg.Align != screen.AlignCenter || cfg.VAlign != screen.VAlignTop || cfg.Fill != screen.FillNone || cfg.Gutter != 0 {
		t.Errorf("layout = %+v", cfg.Layout)
	}
	// The base theme of screens-toml, section 3.2, with the md_* styles of file-view, section 4.
	want := map[string]string{
		"normal": "7/0", "more": "15/0", "select": "0/7", "md_h1": "15/0 bold", "md_h2": "15/0 bold", "md_h3": "15/0",
		"md_strong": "15/0 bold", "md_em": "14/0", "md_code": "10/0", "md_link": "14/0 underline",
		"md_mark": "11/0", "md_frame": "8/0", "md_dim": "8/0",
	}
	if len(cfg.Theme) != len(want) {
		t.Errorf("theme = %v", cfg.Theme)
	}
	for name, text := range want {
		if got := cfg.Theme[name]; got != style(t, text) {
			t.Errorf("theme %s = %v, want %s", name, got, text)
		}
	}
	if cfg.Start != "main" || len(cfg.Fields) != 0 || !reflect.DeepEqual(cfg.Telnet, defaultTelnet) || !reflect.DeepEqual(cfg.Limits, defaultLimits) ||
		cfg.Time != "2006-01-02 15:04:05" || cfg.Zone != time.Local || cfg.Name != "Kind Brothers BBS" {
		t.Errorf("cfg = %+v", cfg)
	}
	if cfg.ThemeChoice != ThemeDark {
		t.Errorf("theme choice = %q", cfg.ThemeChoice)
	}
	for _, colors := range []int{16, 256, 16777216} {
		if !reflect.DeepEqual(dark(cfg, colors), cfg.Theme) {
			t.Errorf("dark theme %d = %v", colors, dark(cfg, colors))
		}
		// Without light tables, the light theme is the dark theme: normal of the base theme is 7/0.
		if !reflect.DeepEqual(light(cfg, colors), cfg.Theme) {
			t.Errorf("light theme %d = %v", colors, light(cfg, colors))
		}
	}
}

func TestParse(t *testing.T) {
	data := `
[screen]
width = [64, 132]
more = "[ ещё ]"

[theme.default]
normal = "14"
title  = "15/4 bold"
prompt = "/1"
more = "-/-"

[menu]
  start = "welcome"

[fields]
w = "term.cols"
`
	cfg, ds := parseScreens([]byte(data))
	if len(ds) != 0 {
		t.Fatalf("diags: %q", messages(ds))
	}
	if cfg.Width != (screen.WidthLimits{Min: 64, Max: 132}) || cfg.MoreText("en") != "[ ещё ]" {
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
		"title":  "15/4 bold", // the example of screens-toml, section 4
		"prompt": "14/1",      // the foreground from normal
		"more":   "-/-",
		"select": "0/7",
	}
	if len(cfg.Theme) != len(want)+len(mdStyles) {
		t.Errorf("theme = %v", cfg.Theme)
	}
	for name, text := range want {
		if got := cfg.Theme[name]; got != style(t, text) {
			t.Errorf("theme %s = %v, want %s", name, got, text)
		}
	}
}

// Each value of align, valign and fill, and the limits of gutter (screens-toml, section 3.1).
func TestLayout(t *testing.T) {
	tests := []struct {
		data string
		want screen.Layout
	}{
		{`align = "left"`, screen.Layout{Align: screen.AlignLeft}},
		{`align = "center"`, screen.Layout{Align: screen.AlignCenter}},
		{`align = "right"`, screen.Layout{Align: screen.AlignRight}},
		{`valign = "top"`, screen.Layout{VAlign: screen.VAlignTop}},
		{`valign = "middle"`, screen.Layout{VAlign: screen.VAlignMiddle}},
		{`valign = "bottom"`, screen.Layout{VAlign: screen.VAlignBottom}},
		{`fill = "none"`, screen.Layout{Fill: screen.FillNone}},
		{`fill = "spaces"`, screen.Layout{Fill: screen.FillSpaces}},
		{`fill = "erase"`, screen.Layout{Fill: screen.FillErase}},
		{`gutter = 0`, screen.Layout{}},
		{`gutter = 100`, screen.Layout{Gutter: 100}},
	}
	for _, tt := range tests {
		cfg, ds := parseScreens([]byte("[screen]\n" + tt.data))
		if len(ds) != 0 {
			t.Errorf("%s: diags %q", tt.data, messages(ds))
		}
		tt.want.Width = defaultLayout.Width
		if cfg.Layout != tt.want {
			t.Errorf("%s: layout %+v, want %+v", tt.data, cfg.Layout, tt.want)
		}
	}
}

// A theme style takes the colours of normal, not its attributes.
func TestNormalAttributes(t *testing.T) {
	cfg, ds := parseScreens([]byte("[theme.default]\nnormal = \"2/3 bold\"\nx = \"5\"\n"))
	if len(ds) != 0 {
		t.Fatalf("diags: %q", messages(ds))
	}
	if cfg.Theme["normal"] != style(t, "2/3 bold") || cfg.Theme["x"] != style(t, "5/3") {
		t.Errorf("theme = %v", cfg.Theme)
	}
}

// A reference gives the style of its name in the same theme, with its attributes added
// (screens-toml, section 3.2).
func TestReferences(t *testing.T) {
	data := strings.Join([]string{
		`[theme.default]`,
		`normal = "#c0c0c0/#101020"`,
		`title  = "#60a0d0 bold"`,
		`md_h1  = "title underline"`,
		`md_h2  = "md_h1"`,
		`md_h3  = "md_h2 blink"`,
		`[theme.16]`,
		`normal = "7/0"`,
		`title  = "15 bold"`,
		`[theme.256]`,
		`md_dim = "title"`,
		`[theme.light.default]`,
		`normal = "#202020/#f0f0e8"`,
		`title  = "#004080 bold"`,
		`[theme.light.16]`,
		`normal = "0/7"`,
		`[theme.light.256]`,
		`md_h2 = "normal"`,
	}, "\n")
	cfg, ds := parseScreens([]byte(data))
	if len(ds) != 0 {
		t.Fatalf("diags: %q", messages(ds))
	}
	tests := []struct {
		theme map[string]screen.Style
		name  string
		want  string
	}{
		{cfg.Theme, "md_h1", "#60a0d0/#101020 bold underline"},
		// The depth theme changes the style of the name.
		{dark(cfg, 16), "md_h1", "15/0 bold underline"},
		{dark(cfg, 16), "md_h2", "15/0 bold underline"},
		{dark(cfg, 16), "md_h3", "15/0 bold underline blink"},
		{dark(cfg, 256), "md_h3", "#60a0d0/#101020 bold underline blink"},
		// A reference in a depth theme.
		{dark(cfg, 256), "md_dim", "#60a0d0/#101020 bold"},
		{dark(cfg, 16), "md_dim", "8/0"},
		// The light tables change the style of the name.
		{light(cfg, 16), "md_h1", "#004080/7 bold underline"},
		{light(cfg, 16777216), "md_h3", "#004080/#f0f0e8 bold underline blink"},
		// A light table changes a reference, and the chain after it.
		{light(cfg, 256), "md_h2", "#202020/#f0f0e8"},
		{light(cfg, 256), "md_h3", "#202020/#f0f0e8 blink"},
		{light(cfg, 256), "md_dim", "#004080/#f0f0e8 bold"},
	}
	for _, tt := range tests {
		if got := tt.theme[tt.name]; got != style(t, tt.want) {
			t.Errorf("%s = %v, want %s", tt.name, got, tt.want)
		}
	}
}

// An unknown name and a cycle are errors at the position of the reference (screens-toml, section 3.2).
func TestReferenceErrors(t *testing.T) {
	data := strings.Join([]string{
		`[theme.default]`,
		`a = "b"`,
		`b = "a"`,
		`c = "a bold"`,
		`d = "haed"`,
		`e = "a bolt"`,
		`f = "g"`,
		`g = "7"`,
		`h = "#x"`,
		`i = "h underline"`,
		`x = "y"`,
		`y = "z"`,
		`z = "x"`,
		`[theme.16]`,
		`g = "f"`,
		`y = "x"`,
		`[theme.light.default]`,
		`f = "d"`,
	}, "\n")
	cfg, ds := parseScreens([]byte(data))
	want := []string{
		`screens/screens.toml:6:5: theme.default.e: invalid style "a bolt": unknown attribute "bolt"`,
		`screens/screens.toml:9:5: theme.default.h: invalid style "#x": invalid colour "#x"`,
		`screens/screens.toml:5:5: theme.default.d: unknown style "haed"`,
		// The dark theme of 16 colours.
		`screens/screens.toml:2:5: theme.default.a: cycle: a -> b -> a`,
		`screens/screens.toml:3:5: theme.default.b: cycle: b -> a -> b`,
		// A cycle across two tables.
		`screens/screens.toml:7:5: theme.default.f: cycle: f -> g -> f`,
		`screens/screens.toml:15:5: theme.16.g: cycle: g -> f -> g`,
		// The path of x differs in the dark theme of 256 colours: the first theme gives it.
		`screens/screens.toml:11:5: theme.default.x: cycle: x -> y -> x`,
		`screens/screens.toml:16:5: theme.16.y: cycle: y -> x -> y`,
		// The dark theme of 256 colours. With 16 colours, z leads into a cycle.
		`screens/screens.toml:12:5: theme.default.y: cycle: y -> z -> x -> y`,
		`screens/screens.toml:13:5: theme.default.z: cycle: z -> x -> y -> z`,
	}
	if got := messages(ds); !slices.Equal(got, want) {
		t.Errorf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	tests := []struct {
		theme map[string]screen.Style
		name  string
		want  string
	}{
		// A style in a cycle and a reference that leads into a cycle get the colours of normal.
		{dark(cfg, 16), "c", "7/0 bold"},
		{dark(cfg, 16), "g", "7/0"},
		{dark(cfg, 16), "z", "7/0"},
		// A reference to a name with an invalid style.
		{dark(cfg, 256), "i", "7/0 underline"},
		{light(cfg, 16), "f", "7/0"},
	}
	for _, tt := range tests {
		if got := tt.theme[tt.name]; got != style(t, tt.want) {
			t.Errorf("%s = %v, want %s", tt.name, got, tt.want)
		}
	}
	if _, ok := cfg.Theme["d"]; ok {
		t.Error("a reference with an unknown name is in the theme")
	}
}

// Each other error has the position of its value, or of its key name (screen-render, section 3.1).
func TestErrors(t *testing.T) {
	data := strings.Join([]string{
		`[screen]`,
		`width = [80, 40]`,
		`more = "a` + "\u200b" + `"`,
		`align = "centre"`,
		`valign = "Top"`,
		`fill = ""`,
		`gutter = 101`,
		`theme = "Dark"`,
		`zone = "Mars/Olympus"`,
		`[theme.default]`,
		`Title = "1"`,
		`normal = "7/0"`,
		`bad = "#x"`,
		`[theme.16]`,
		`Title = "1"`,
		`title = "1"`,
		`bad = "2"`,
		`normal = "1 bolt"`,
		`[theme.256]`,
		`more = "1"`,
		`[theme.light.default]`,
		`x = "1"`,
		`normal = "1 bolt"`,
		`[theme.light.16]`,
		`title = "1"`,
		`[theme.light.256]`,
		`Bad = "1"`,
		`[menu]`,
		`start = "Main"`,
	}, "\n")
	cfg, ds := parseScreens([]byte(data))
	want := []string{
		`screens/screens.toml:2:9: screen.width: width [80, 40] is not in 1 <= min <= max <= 255`,
		`screens/screens.toml:4:9: screen.align: invalid value "centre", want "left", "center" or "right"`,
		`screens/screens.toml:5:10: screen.valign: invalid value "Top", want "top", "middle" or "bottom"`,
		`screens/screens.toml:6:8: screen.fill: invalid value "", want "none", "spaces" or "erase"`,
		`screens/screens.toml:7:10: screen.gutter: 101, must be from 0 to 100`,
		`screens/screens.toml:3:8: screen.more: zero-width grapheme U+200B`,
		`screens/screens.toml:9:8: screen.zone: unknown time zone "Mars/Olympus"`,
		`screens/screens.toml:8:9: screen.theme: invalid value "Dark", want "dark", "light" or "auto"`,
		`screens/screens.toml:29:9: menu.start: invalid screen name "Main"`,
		`screens/screens.toml:11:1: theme.default: invalid style name "Title"`,
		`screens/screens.toml:13:7: theme.default.bad: invalid style "#x": invalid colour "#x"`,
		// A name with an invalid style is a name of the theme: bad gives no error in theme.16.
		`screens/screens.toml:15:1: theme.16: invalid style name "Title"`,
		`screens/screens.toml:18:10: theme.16.normal: invalid style "1 bolt": unknown attribute "bolt"`,
		`screens/screens.toml:16:1: theme.16: unknown style "title"`,
		`screens/screens.toml:23:10: theme.light.default.normal: invalid style "1 bolt": unknown attribute "bolt"`,
		`screens/screens.toml:22:1: theme.light.default: unknown style "x"`,
		`screens/screens.toml:25:1: theme.light.16: unknown style "title"`,
		`screens/screens.toml:27:1: theme.light.256: invalid style name "Bad"`,
	}
	if got := messages(ds); !slices.Equal(got, want) {
		t.Errorf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	if !ds.HasErrors() {
		t.Error("HasErrors = false")
	}
	if cfg.Layout != defaultLayout || cfg.More != nil || cfg.Start != "main" || cfg.ThemeChoice != ThemeDark ||
		cfg.Zone != time.Local {
		t.Errorf("an invalid value replaced a default: %+v", cfg)
	}
	if dark(cfg, 16)["normal"] != style(t, "7/0") || dark(cfg, 16)["bad"] != style(t, "2/0") {
		t.Errorf("theme.16 = %v", dark(cfg, 16))
	}
	if dark(cfg, 256)["more"] != style(t, "1/0") {
		t.Errorf("theme.256 = %v", dark(cfg, 256))
	}
}

// A TOML document gives a syntax or type error; else all unknown keys;
// else its other errors (screen-render, section 3.1).
func TestTOMLErrors(t *testing.T) {
	tests := []struct {
		data string
		want []string
	}{
		{"screen = 1", []string{`1:10: screen: must be a table`}},
		{"theme = []", []string{`1:9: theme: must be a table`}},
		// In an inline table, the key ends at the table (screen-render, section 3.1).
		{"theme = { default = 1 }", []string{`1:21: theme: invalid type`}},
		{"[theme.default]\nx = 1", []string{`2:5: theme.default.x: must be a string`}},
		{"[theme.16]\nx = 1", []string{`2:5: theme.16.x: must be a string`}},
		{"[theme.32]\nx = \"1\"", []string{`1:2: unknown key "theme.32"`}},
		{"[theme]\nnormal = \"1\"", []string{`2:1: unknown key "theme.normal"`}},
		{"[theme.light.32]\nx = \"1\"", []string{`1:2: unknown key "theme.light.32"`}},
		{"[theme.light.dark]\nx = \"1\"", []string{`1:2: unknown key "theme.light.dark"`}},
		{"[theme.light]\nnormal = \"1\"", []string{`2:1: unknown key "theme.light.normal"`}},
		{"[theme.light.256]\nx = 1", []string{`2:5: theme.light.256.x: must be a string`}},
		{"[screen]\ntheme = 1", []string{`2:9: screen.theme: must be a string`}},
		{"[telnet]\nlisten = \":1\"", []string{`1:2: unknown key "telnet"`}},
		{"menu = \"main\"", []string{`1:8: menu: must be a table`}},
		{"[menu]\nstart = 1", []string{`2:9: menu.start: must be a string`}},
		{"[screen]\nwidth = [1]", []string{`2:9: screen.width: must be an array of 2 integers`}},
		{"[screen]\nwidth = [1, \"2\"]", []string{`2:13: screen.width: must be an array of integers`}},
		{"[screen]\nmore = 1", []string{`2:8: screen.more: must be a string`}},
		{"[screen]\nfill = 1", []string{`2:8: screen.fill: must be a string`}},
		{"[screen]\nmore = \"\t\"", []string{`2:8: screen.more: control character U+0009`}},
		{"[screen]\nwidth = ", []string{`2:8: expected value, not end of input`}},
		{"[theme.default]\nprompt = \"14 bolt\"", []string{`2:10: theme.default.prompt: invalid style "14 bolt": unknown attribute "bolt"`}},
		{"a.b = 1", []string{`1:1: unknown key "a.b"`}},
		{"[MENU]\nstart = \"x\"", []string{`1:2: unknown key "MENU"`}},
		{"[menu]\nstart = \"x\"\nSTART = \"y\"", []string{`3:1: unknown key "menu.START"`}},
		{"\n[foo.bar]\nx = 1", []string{`2:2: unknown key "foo.bar"`}},
		// A type error comes before an unknown key, an unknown key before an other error.
		{"[theme.default]\nx = \"#bad\"\ny = 1\nz = \"#bad\"", []string{`3:5: theme.default.y: must be a string`}},
		{"a = 1\n[theme.default]\nx = \"#bad\"\nz = \"#bad\"", []string{`1:1: unknown key "a"`}},
		{"[theme.default]\nx = \"#bad\"\nz = \"#bad\"", []string{
			`2:5: theme.default.x: invalid style "#bad": invalid colour "#bad"`,
			`3:5: theme.default.z: invalid style "#bad": invalid colour "#bad"`,
		}},
		{"a = 1\n[screen]\nb = 2\nwidth = [1, 1]", []string{`1:1: unknown key "a"`, `3:1: unknown key "screen.b"`}},
		{"[fields]\nw = 1", []string{`2:5: fields.w: must be a string`}},
		// An invalid alias name has the position of the key, an invalid path of the value
		// (screens-toml, section 3.3).
		{"fields = { \"яя\" = \"term.cols\", and = \"term.rows\", k = \"item.A\", u = \"term.x\", v = \"w\" }", []string{
			`1:34: fields: alias "and" is a reserved word`,
			`1:57: fields: alias "k": "item.A" is valid only in {#if}`,
			`1:71: fields: alias "u": unknown field "term.x"`,
			`1:85: fields: alias "v": "w" is not a registry path`,
			`1:12: fields: invalid alias "яя"`,
		}},
		{"fields = { row = \"term.cols\", list = \"term.rows\", p = \"list.page\", n = \"row.num\" }", []string{
			`1:31: fields: alias "list" is a reserved word`,
			`1:12: fields: alias "row" is a reserved word`,
		}},
	}
	for _, tt := range tests {
		_, ds := parseScreens([]byte(tt.data))
		var want []string
		for _, w := range tt.want {
			want = append(want, ScreensFile+":"+w)
		}
		if got := messages(ds); !slices.Equal(got, want) {
			t.Errorf("parseScreens(%q) = %q, want %q", tt.data, got, want)
		}
	}
}

// [screen] time and [bbs] name take any string (screens-toml, section 3.1, and bbs-toml, section 3.2).
func TestTimeAndName(t *testing.T) {
	cfg, ds := Load(fstest.MapFS{
		"bbs.toml":             {Data: []byte("[bbs]\nname = \"\"\n")},
		"screens/screens.toml": {Data: []byte("[screen]\ntime = \"02.01.2006 15:04\"\n")},
	})
	if len(ds) != 0 {
		t.Fatalf("diags: %q", messages(ds))
	}
	if cfg.Time != "02.01.2006 15:04" || cfg.Name != "" {
		t.Errorf("time %q, name %q", cfg.Time, cfg.Name)
	}
}

// [screen] zone takes Local, UTC or a name of the IANA database (screens-toml, section 3.1).
func TestZone(t *testing.T) {
	for _, name := range []string{"Local", "UTC", "Europe/Berlin", "Asia/Tokyo"} {
		cfg, ds := parseScreens([]byte("[screen]\nzone = \"" + name + "\"\n"))
		if len(ds) != 0 {
			t.Fatalf("%s: diags %q", name, messages(ds))
		}
		if got := cfg.Zone.String(); got != name {
			t.Errorf("zone %q = %q", name, got)
		}
	}
	for _, name := range []string{"", "Europe/Nowhere", "../etc/passwd", "/etc/localtime"} {
		cfg, ds := parseScreens([]byte("[screen]\nzone = \"" + name + "\"\n"))
		want := []string{ScreensFile + `:2:8: screen.zone: unknown time zone "` + name + `"`}
		if got := messages(ds); !slices.Equal(got, want) {
			t.Errorf("zone %q: got %q, want %q", name, got, want)
		}
		if cfg.Zone != time.Local {
			t.Errorf("zone %q replaced the default: %v", name, cfg.Zone)
		}
	}
}

// bbs.toml has only [bbs], [telnet] and [limits] (bbs-toml, section 2).
func TestHostErrors(t *testing.T) {
	tests := []struct{ data, want string }{
		{"[screen]\nwidth = [40, 80]", `bbs.toml:1:2: unknown key "screen"`},
		{"[theme]\ntitle = \"1\"", `bbs.toml:1:2: unknown key "theme"`},
		{"[fields]\nw = \"term.cols\"", `bbs.toml:1:2: unknown key "fields"`},
		{"[menu]\nstart = \"x\"", `bbs.toml:1:2: unknown key "menu"`},
		{"[TELNET]\nlisten = \":1\"", `bbs.toml:1:2: unknown key "TELNET"`},
		{"[telnet]\nlisten = \":1\"\nLISTEN = \":2\"", `bbs.toml:3:1: unknown key "telnet.LISTEN"`},
		{"telnet = 1", `bbs.toml:1:10: telnet: must be a table`},
		{"[bbs]\nname = 1", `bbs.toml:2:8: bbs.name: must be a string`},
		{"[bbs]\nnode = 1", `bbs.toml:2:1: unknown key "bbs.node"`},
	}
	for _, tt := range tests {
		if _, got := host(tt.data); !slices.Equal(got, []string{tt.want}) {
			t.Errorf("parseHost(%q) = %q, want %q", tt.data, got, tt.want)
		}
	}
}

// The examples of screens-toml, section 4, and of bbs-toml, section 4, give no diagnostics.
func TestExamples(t *testing.T) {
	screens := `
[screen]
width = [40, 100]
more = "-- More --"
valign = "middle"
fill = "spaces"

[theme.default]
normal = "7/0"
title  = "15/4 bold"
prompt = "14"

[theme.16]
title = "15/1 bold"

[fields]
w = "term.cols"

[menu]
start = "welcome"
`
	host := `
[bbs]
name = "Kind Brothers BBS"

[telnet]
listen = ":2323"
proxy = ["127.0.0.1"]
probe_timeout = 5

[limits]
max_conns = 64
max_conns_per_addr = 4
max_connects_per_min = 10
allow = ["192.0.2.10", "2001:db8::/48"]
idle_timeout = 600
output_timeout = 60
`
	cfg, ds := Load(fstest.MapFS{
		"bbs.toml":             {Data: []byte(host)},
		"screens/screens.toml": {Data: []byte(screens)},
	})
	if len(ds) != 0 {
		t.Fatalf("diags: %q", messages(ds))
	}
	if dark(cfg, 16)["title"] != style(t, "15/1 bold") || dark(cfg, 256)["title"] != style(t, "15/4 bold") {
		t.Errorf("title: 16 %v, 256 %v", dark(cfg, 16)["title"], dark(cfg, 256)["title"])
	}
	if cfg.VAlign != screen.VAlignMiddle || cfg.Fill != screen.FillSpaces {
		t.Errorf("layout = %+v", cfg.Layout)
	}
	if dark(cfg, 16)["prompt"] != style(t, "14/0") || cfg.Start != "welcome" || cfg.Telnet.Listen != ":2323" ||
		cfg.Name != "Kind Brothers BBS" {
		t.Errorf("cfg = %+v", cfg)
	}
}

// The example of screens-toml, section 3.2.
func TestDepthTheme(t *testing.T) {
	data := `
[theme.default]
normal = "#c0c0c0/#101020"
title  = "#60a0d0 bold"

[theme.16]
normal = "7/0"
title  = "15 bold"
`
	cfg, ds := parseScreens([]byte(data))
	if len(ds) != 0 {
		t.Fatalf("diags: %q", messages(ds))
	}
	want := map[int]string{
		16:       "15/0 bold",
		256:      "#60a0d0/#101020 bold",
		16777216: "#60a0d0/#101020 bold",
	}
	for colors, text := range want {
		if got := dark(cfg, colors)["title"]; got != style(t, text) {
			t.Errorf("title at %d = %v, want %s", colors, got, text)
		}
	}
	if cfg.Theme["title"] != style(t, "#60a0d0/#101020 bold") {
		t.Errorf("base title = %v", cfg.Theme["title"])
	}
}

// A style of the session theme takes its colours from normal of the session theme.
// A depth theme can set normal and more without [theme.default].
func TestDepthNormal(t *testing.T) {
	data := `
[theme.default]
title  = "/4 bold"
prompt = "14"

[theme.256]
normal = "#203040/#405060"
more   = "1"
`
	cfg, ds := parseScreens([]byte(data))
	if len(ds) != 0 {
		t.Fatalf("diags: %q", messages(ds))
	}
	want := map[string]string{
		"normal": "#203040/#405060",
		"more":   "1/#405060",
		"title":  "#203040/4 bold",
		"prompt": "14/#405060",
		"select": "0/7",
	}
	th := dark(cfg, 256)
	if len(th) != len(want)+len(mdStyles) {
		t.Errorf("theme.256 = %v", th)
	}
	for name, text := range want {
		if got := th[name]; got != style(t, text) {
			t.Errorf("%s = %v, want %s", name, got, text)
		}
	}
	if cfg.Theme["title"] != style(t, "7/4 bold") || cfg.Theme["more"] != style(t, "15/0") {
		t.Errorf("base theme = %v", cfg.Theme)
	}
}

// The example of screens-toml, section 3.2, with light tables.
func TestLightTheme(t *testing.T) {
	data := `
[screen]
theme = "auto"

[theme.default]
normal = "#c0c0c0/#101020"
title  = "#60a0d0 bold"
prompt = "14"

[theme.16]
normal = "7/0"
title  = "15 bold"

[theme.light.default]
normal = "#202020/#f0f0e8"
title  = "#004080 bold"

[theme.light.16]
normal = "0/7"

[theme.light.256]
prompt = "4"
`
	cfg, ds := parseScreens([]byte(data))
	if len(ds) != 0 {
		t.Fatalf("diags: %q", messages(ds))
	}
	if cfg.ThemeChoice != ThemeAuto {
		t.Errorf("theme choice = %q", cfg.ThemeChoice)
	}
	want := []struct {
		colors                int
		normal, title, prompt string
	}{
		{16, "0/7", "#004080/7 bold", "14/7"},
		{256, "#202020/#f0f0e8", "#004080/#f0f0e8 bold", "4/#f0f0e8"},
		{16777216, "#202020/#f0f0e8", "#004080/#f0f0e8 bold", "14/#f0f0e8"},
	}
	for _, w := range want {
		th := light(cfg, w.colors)
		for name, text := range map[string]string{"normal": w.normal, "title": w.title, "prompt": w.prompt} {
			if got := th[name]; got != style(t, text) {
				t.Errorf("light %d: %s = %v, want %s", w.colors, name, got, text)
			}
		}
	}
	// The light tables do not change the dark themes.
	if dark(cfg, 16)["title"] != style(t, "15/0 bold") || dark(cfg, 256)["prompt"] != style(t, "14/#101020") {
		t.Errorf("dark: 16 %v, 256 %v", dark(cfg, 16), dark(cfg, 256))
	}
}

// A light normal with one colour takes the other from 0/15 (screens-toml, section 3.2).
// A style of a light table replaces the whole style with the same name.
// With 16777216 colours, only [theme.light.default] applies.
func TestLightNormal(t *testing.T) {
	data := `
[theme.default]
title = "1"

[theme.light.default]
normal = "4"
title  = "/2"

[theme.light.16]
normal = "/7"
`
	cfg, ds := parseScreens([]byte(data))
	if len(ds) != 0 {
		t.Fatalf("diags: %q", messages(ds))
	}
	want := map[int]map[string]string{
		16:       {"normal": "0/7", "title": "0/2", "more": "15/0", "md_em": "14/7"},
		256:      {"normal": "4/15", "title": "4/2", "more": "15/0", "md_em": "14/15"},
		16777216: {"normal": "4/15", "title": "4/2", "more": "15/0", "md_em": "14/15"},
	}
	for colors, styles := range want {
		for name, text := range styles {
			if got := light(cfg, colors)[name]; got != style(t, text) {
				t.Errorf("light %d: %s = %v, want %s", colors, name, got, text)
			}
		}
	}
}

// The session theme is light for light, and for auto with term.background light
// (screens-toml, section 3.2).
func TestSessionTheme(t *testing.T) {
	cfg, ds := parseScreens([]byte("[theme.default]\nnormal = \"7/1\"\n[theme.light.default]\nnormal = \"0/7\"\n"))
	if len(ds) != 0 {
		t.Fatalf("diags: %q", messages(ds))
	}
	tests := []struct {
		choice     ThemeChoice
		background string
		want       string
	}{
		{ThemeDark, "", "7/1"}, {ThemeDark, "light", "7/1"},
		{ThemeLight, "", "0/7"}, {ThemeLight, "dark", "0/7"},
		{ThemeAuto, "", "7/1"}, {ThemeAuto, "dark", "7/1"}, {ThemeAuto, "light", "0/7"},
	}
	for _, tt := range tests {
		for _, colors := range []int{16, 256, 16777216} {
			th := cfg.SessionTheme(tt.choice, screen.Term{Colors: colors, Background: tt.background})
			if th["normal"] != style(t, tt.want) {
				t.Errorf("%s, background %q, colors %d: normal %v, want %s", tt.choice, tt.background, colors, th["normal"], tt.want)
			}
		}
	}
	for _, c := range []string{"dark", "light", "auto"} {
		cfg, ds := parseScreens([]byte("[screen]\ntheme = \"" + c + "\"\n"))
		if len(ds) != 0 || cfg.ThemeChoice != ThemeChoice(c) {
			t.Errorf("theme %q: choice %q, diags %q", c, cfg.ThemeChoice, messages(ds))
		}
	}
}

// Load reads both files (screens-toml, section 2).
func TestLoad(t *testing.T) {
	tests := []struct {
		fsys fstest.MapFS
		want []string
	}{
		{fstest.MapFS{"screens/screens.toml": {Data: []byte("[menu]\nstart = \"x\"\n")}}, []string{
			"bbs.toml:1:1: open bbs.toml: file does not exist",
		}},
		{fstest.MapFS{"bbs.toml": {}, "screens/screens.toml/x": {}}, []string{
			"screens/screens.toml:1:1: not a regular file",
		}},
		{fstest.MapFS{"bbs.toml": {Data: []byte("x = 1")}, "screens/screens.toml": {Data: []byte("y = 1")}}, []string{
			`screens/screens.toml:1:1: unknown key "y"`,
			`bbs.toml:1:1: unknown key "x"`,
		}},
	}
	for _, tt := range tests {
		_, ds := Load(tt.fsys)
		if got := messages(ds); !slices.Equal(got, tt.want) {
			t.Errorf("Load = %q, want %q", got, tt.want)
		}
	}
}

// A dangling symlink at screens/screens.toml exists and is not a regular file
// (screens-toml, section 2).
func TestLoadDanglingLink(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "bbs.toml"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "screens"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("missing.toml", filepath.Join(dir, "screens", "screens.toml")); err != nil {
		t.Fatal(err)
	}
	_, ds := Load(os.DirFS(dir))
	want := []string{"screens/screens.toml:1:1: not a regular file"}
	if got := messages(ds); !slices.Equal(got, want) {
		t.Errorf("Load = %q, want %q", got, want)
	}
}

// The loader removes a BOM (bbt-format, section 2).
func TestBOM(t *testing.T) {
	cfg, ds := parseScreens([]byte("\ufeff[menu]\nstart = \"x\"\n"))
	if len(ds) != 0 || cfg.Start != "x" {
		t.Errorf("start %q, diags %q", cfg.Start, messages(ds))
	}
}

func TestDiagLimit(t *testing.T) {
	var b strings.Builder
	for i := range 2 * screen.MaxFileDiags {
		fmt.Fprintf(&b, "k%d = 1\n", i)
	}
	_, ds := parseScreens([]byte(b.String()))
	got := messages(ds)
	if len(got) != screen.MaxFileDiags+1 || got[0] != `screens/screens.toml:1:1: unknown key "k0"` ||
		got[len(got)-1] != "screens/screens.toml:101:1: too many diagnostics" {
		t.Errorf("%d diags, first %q, last %q", len(got), got[0], got[len(got)-1])
	}
}

// The languages of the BBS and the text of the More row in each language
// (screens-toml, sections 3.1 and 3.5; session-language, sections 2 and 7.2).
func TestLang(t *testing.T) {
	data := `
[screen]
lang = "ru"
more = "[more]"

[lang.uk]
more = "[далі]"

[lang.de]
`
	cfg, ds := parseScreens([]byte(data))
	if len(ds) != 0 {
		t.Fatalf("diags: %q", messages(ds))
	}
	if cfg.Lang != "ru" || !slices.Equal(cfg.Langs, []string{"de", "ru", "uk"}) {
		t.Errorf("lang %q, langs %q", cfg.Lang, cfg.Langs)
	}
	for lang, want := range map[string]string{"uk": "[далі]", "de": "[more]", "ru": "[more]"} {
		if got := cfg.MoreText(lang); got != want {
			t.Errorf("more %s = %q, want %q", lang, got, want)
		}
	}
	cfg, _ = parseScreens([]byte("[lang.ru]\nmore = \"\""))
	if !slices.Equal(cfg.Langs, []string{"en", "ru"}) || cfg.MoreText("ru") != "" || cfg.MoreText("en") != "-- More --" {
		t.Errorf("langs %q, more ru %q, more en %q", cfg.Langs, cfg.MoreText("ru"), cfg.MoreText("en"))
	}
	// Without more, the engine text in the language (session-language, section 7.2).
	cfg, _ = parseScreens([]byte("[lang.ru]\n[lang.de]"))
	if cfg.MoreText("ru") != "-- Дальше --" || cfg.MoreText("de") != "-- More --" {
		t.Errorf("more ru %q, more de %q", cfg.MoreText("ru"), cfg.MoreText("de"))
	}
}

func TestLangErrors(t *testing.T) {
	data := strings.Join([]string{
		`[screen]`,
		`lang = "RU"`,
		`[lang.e]`,
		`[lang.de]`,
		`more = "a` + "\u200b" + `"`,
	}, "\n")
	cfg, ds := parseScreens([]byte(data))
	want := []string{
		`screens/screens.toml:2:8: screen.lang: invalid language code "RU"`,
		`screens/screens.toml:5:8: lang.de.more: zero-width grapheme U+200B`,
		`screens/screens.toml:3:7: lang: invalid language code "e"`,
	}
	if got := messages(ds); !slices.Equal(got, want) {
		t.Errorf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	if cfg.Lang != "en" || !slices.Equal(cfg.Langs, []string{"de", "en"}) || cfg.MoreText("de") != "-- More --" {
		t.Errorf("lang %q, langs %q, more de %q", cfg.Lang, cfg.Langs, cfg.MoreText("de"))
	}
	_, ds = parseScreens([]byte("[lang.ru]\nx = 1"))
	if got := messages(ds); !slices.Equal(got, []string{`screens/screens.toml:2:1: unknown key "lang.ru.x"`}) {
		t.Errorf("unknown key: %q", got)
	}
	cfg, ds = parseScreens([]byte("lang = 1"))
	if len(ds) != 1 || !slices.Equal(cfg.Langs, []string{"en"}) {
		t.Errorf("diags %q, langs %q", messages(ds), cfg.Langs)
	}
}
