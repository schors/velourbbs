package menu

import (
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/schors/velourbbs/internal/bbt"
	"github.com/schors/velourbbs/internal/config"
	"github.com/schors/velourbbs/internal/screen"
)

// fakeTerm is a Terminal without Events. It logs each page as "page: <row 1>".
type fakeTerm struct {
	cols    []int // the cols of each Term call; the last value stays
	keys    []string
	log     []string
	pageErr error // the error of Page
}

func (f *fakeTerm) Term() screen.Term {
	c := f.cols[0]
	if len(f.cols) > 1 {
		f.cols = f.cols[1:]
	}
	return screen.Term{Type: "ansi", Cols: c, Rows: 10, Charset: "utf-8", Colors: 16}
}

func (f *fakeTerm) Key() (string, error) {
	if len(f.keys) == 0 {
		return "", io.EOF
	}
	k := f.keys[0]
	f.keys = f.keys[1:]
	return k, nil
}

func (f *fakeTerm) Page(p screen.Page) error {
	var row strings.Builder
	for _, c := range p.Line(0, nil) {
		row.WriteString(c.Grapheme)
	}
	f.log = append(f.log, "page: "+strings.TrimRight(row.String(), " "))
	return f.pageErr
}

// eventsTerm is a fakeTerm with Events. It logs the events as strings.
type eventsTerm struct{ *fakeTerm }

func (f eventsTerm) Screen(name string) error {
	f.log = append(f.log, "screen "+name)
	return nil
}

func (f eventsTerm) KeyEvent(k string, ignored bool) error {
	if ignored {
		k += " ignored"
	}
	f.log = append(f.log, "key "+k)
	return nil
}

func (f eventsTerm) Logoff() error {
	f.log = append(f.log, "logoff")
	return nil
}

// sessionFiles are the screens of the session tests. Each case can change them.
var sessionFiles = map[string]string{
	"main": `+++
[[item]]
key = "S"
do = "goto info"
when = "term.cols >= 80"
[[item]]
key = "L"
do = "goto long"
[[item]]
key = "G"
do = "logoff"
[[item]]
key = "ANY"
do = "goto main"
+++
{#if item.S}[S]ystem {/if}[G]oodbye
`,
	"info":    "+++\n[[item]]\nkey = \"Q\"\ndo = \"goto main\"\n+++\nInfo\n",
	"long":    "+++\n[[item]]\nkey = \"ANY\"\ndo = \"goto main\"\n+++\n" + lines(12),
	"goodbye": "Bye\n",
}

// lines gives n lines "1" to n.
func lines(n int) string {
	var b strings.Builder
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&b, "%d\n", i)
	}
	return b.String()
}

// noValues gives empty values of the session and BBS fields.
func noValues() screen.SessionValues { return screen.SessionValues{} }

// load loads the screens of files with the changes of change. A new name adds a file.
// An empty value removes a file.
func load(t *testing.T, change map[string]string) map[string]*screen.Screen {
	t.Helper()
	files := maps.Clone(sessionFiles)
	maps.Copy(files, change)
	fsys := fstest.MapFS{}
	for name, data := range files {
		if data != "" {
			fsys["screens/"+name+".bbt"] = &fstest.MapFile{Data: []byte(data)}
		}
	}
	theme := map[string]screen.Style{"normal": {}, "more": {}}
	ss, ds := bbt.Load(fsys, theme, screen.Layout{Width: screen.WidthLimits{Min: 20, Max: 80}}, nil)
	ds = Check(config.Config{Start: "main"}, ss, ds)
	if len(ds) > 0 {
		t.Fatalf("load: %v", ds)
	}
	return ss
}

func TestRun(t *testing.T) {
	tests := []struct {
		name    string
		change  map[string]string
		cols    []int
		keys    []string
		want    []string
		wantErr error
	}{{
		name: "key match",
		cols: []int{80},
		keys: []string{"s", "X", "q"},
		want: []string{
			"screen main", "page: [S]ystem [G]oodbye",
			"key S", "screen info", "page: Info",
			"key X ignored",
			"key Q", "screen main", "page: [S]ystem [G]oodbye",
		},
		wantErr: io.EOF,
	}, {
		name:    "Russian layout",
		cols:    []int{80},
		keys:    []string{"ы", "ж"},
		want:    []string{"screen main", "page: [S]ystem [G]oodbye", "key S", "screen info", "page: Info", "key ж ignored"},
		wantErr: io.EOF,
	}, {
		// S is disabled at 40 columns, so ANY matches it and renders main again
		// at the new size.
		name: "disabled item, ANY and goto to the current screen",
		cols: []int{40, 80},
		keys: []string{"S"},
		want: []string{
			"screen main", "page: [G]oodbye",
			"key S", "screen main", "page: [S]ystem [G]oodbye",
		},
		wantErr: io.EOF,
	}, {
		name: "More rows",
		cols: []int{80},
		keys: []string{"L", "s", "B"},
		want: []string{
			"screen main", "page: [S]ystem [G]oodbye",
			"key L", "screen long", "page: 1",
			"key S", "page: 10",
			"key B", "screen main", "page: [S]ystem [G]oodbye",
		},
		wantErr: io.EOF,
	}, {
		name:    "keys end on a More row",
		cols:    []int{80},
		keys:    []string{"L"},
		want:    []string{"screen main", "page: [S]ystem [G]oodbye", "key L", "screen long", "page: 1"},
		wantErr: io.EOF,
	}, {
		// Run does not wait for a key after goodbye: X stays.
		name: "logoff with goodbye",
		cols: []int{80},
		keys: []string{"G", "X"},
		want: []string{"screen main", "page: [S]ystem [G]oodbye", "key G", "logoff", "screen goodbye", "page: Bye"},
	}, {
		name:   "logoff with More rows in goodbye",
		change: map[string]string{"goodbye": lines(12)},
		cols:   []int{80},
		keys:   []string{"G", "X", "Y"},
		want: []string{
			"screen main", "page: [S]ystem [G]oodbye",
			"key G", "logoff", "screen goodbye", "page: 1",
			"key X", "page: 10",
		},
	}, {
		// The goto gives its args to the screen. A goto to the current screen
		// with other args shows it with the new values.
		name: "arguments",
		change: map[string]string{
			"main": `+++
[[item]]
key = "1"
do = "goto view"
args = { f = "gold.md" }
[[item]]
key = "2"
do = "goto view"
args = { f = "silver.md" }
+++
Main
`,
			"view": `+++
[[item]]
key = "N"
do = "goto view"
args = { f = "next.md" }
[[item]]
key = "Q"
do = "goto main"
when = 'arg.f != "next.md"'
+++
File {arg.f}
`,
		},
		cols: []int{80},
		keys: []string{"1", "Q", "2", "N", "Q"},
		want: []string{
			"screen main", "page: Main",
			"key 1", "screen view", "page: File gold.md",
			"key Q", "screen main", "page: Main",
			"key 2", "screen view", "page: File silver.md",
			"key N", "screen view", "page: File next.md",
			"key Q ignored",
		},
		wantErr: io.EOF,
	}, {
		name:   "logoff without goodbye",
		change: map[string]string{"goodbye": ""},
		cols:   []int{80},
		keys:   []string{"G"},
		want:   []string{"screen main", "page: [S]ystem [G]oodbye", "key G", "logoff"},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := load(t, tt.change)
			ft := &fakeTerm{cols: tt.cols, keys: slices.Clone(tt.keys)}
			err := Run(eventsTerm{ft}, config.Config{More: "-- More --"}, ss, "main", nil, noValues)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("error %v, want %v", err, tt.wantErr)
			}
			if !slices.Equal(ft.log, tt.want) {
				t.Errorf("log:\n%s\nwant:\n%s", strings.Join(ft.log, "\n"), strings.Join(tt.want, "\n"))
			}
		})
	}
}

// An error of the terminal ends the session.
func TestRunTermError(t *testing.T) {
	errClosed := errors.New("closed")
	ft := &fakeTerm{cols: []int{80}, keys: []string{"G"}, pageErr: errClosed}
	err := Run(ft, config.Config{}, load(t, nil), "main", nil, noValues)
	if !errors.Is(err, errClosed) || len(ft.log) != 1 {
		t.Errorf("error %v, log %q; want %v after one page", err, ft.log, errClosed)
	}
}

// A terminal without Events gets the pages only.
func TestRunNoEvents(t *testing.T) {
	ft := &fakeTerm{cols: []int{80}, keys: []string{"L", "S", "B", "G", "X"}}
	if err := Run(ft, config.Config{More: "-- More --"}, load(t, nil), "main", nil, noValues); err != nil {
		t.Fatal(err)
	}
	want := []string{"page: [S]ystem [G]oodbye", "page: 1", "page: 10", "page: [S]ystem [G]oodbye", "page: Bye"}
	if !slices.Equal(ft.log, want) {
		t.Errorf("log %q, want %q", ft.log, want)
	}
}

// styleTerm is a Terminal with the term.colors colors and the term.background background.
// It keeps the style of the first cell of the last page and sends the key G once.
type styleTerm struct {
	colors     int
	background string
	style      string
	sent       bool
}

func (f *styleTerm) Term() screen.Term {
	return screen.Term{Type: "ansi", Cols: 80, Rows: 10, Charset: "utf-8", Colors: f.colors, Background: f.background}
}

func (f *styleTerm) Key() (string, error) {
	if f.sent {
		return "", io.EOF
	}
	f.sent = true
	return "G", nil
}

func (f *styleTerm) Page(p screen.Page) error {
	f.style = p.Line(0, nil)[0].Style.String()
	return nil
}

// The session renders with the session theme of the term.colors of the terminal.
func TestRunSessionTheme(t *testing.T) {
	fsys := fstest.MapFS{
		"bbs.toml": {},
		"screens/screens.toml": {Data: []byte(
			"[theme.default]\ntitle = \"15/4 bold\"\n[theme.16]\ntitle = \"15/1 bold\"\n")},
		"screens/main.bbt": {Data: []byte("+++\n[[item]]\nkey = \"G\"\ndo = \"logoff\"\n+++\n{!title}T\n")},
	}
	cfg, ds := config.Load(fsys)
	ss, lds := bbt.Load(fsys, cfg.Theme, cfg.Layout, cfg.Fields)
	if ds = append(ds, lds...); len(ds) > 0 {
		t.Fatalf("load: %v", ds)
	}
	for colors, want := range map[int]string{16: "15/1 bold", 256: "15/4 bold"} {
		st := &styleTerm{colors: colors}
		if err := Run(st, cfg, ss, "main", nil, noValues); err != nil {
			t.Fatal(err)
		}
		if st.style != want {
			t.Errorf("colors %d: style %s, want %s", colors, st.style, want)
		}
	}
}

// The session theme follows [screen] theme and term.background. An auto session
// without a reply of the terminal uses the dark theme (screens-toml, section 3.2).
func TestRunThemeChoice(t *testing.T) {
	tests := []struct{ choice, background, want string }{
		{"dark", "light", "15/4 bold"},
		{"light", "", "4/15 bold"},
		{"auto", "", "15/4 bold"},
		{"auto", "dark", "15/4 bold"},
		{"auto", "light", "4/15 bold"},
	}
	for _, tt := range tests {
		fsys := fstest.MapFS{
			"bbs.toml": {},
			"screens/screens.toml": {Data: []byte("[screen]\ntheme = \"" + tt.choice + "\"\n" +
				"[theme.default]\ntitle = \"15/4 bold\"\n[theme.light.default]\ntitle = \"4 bold\"\n")},
			"screens/main.bbt": {Data: []byte("+++\n[[item]]\nkey = \"G\"\ndo = \"logoff\"\n+++\n{!title}T\n")},
		}
		cfg, ds := config.Load(fsys)
		ss, lds := bbt.Load(fsys, cfg.Theme, cfg.Layout, cfg.Fields)
		if ds = append(ds, lds...); len(ds) > 0 {
			t.Fatalf("load: %v", ds)
		}
		st := &styleTerm{colors: 256, background: tt.background}
		if err := Run(st, cfg, ss, "main", nil, noValues); err != nil {
			t.Fatal(err)
		}
		if st.style != tt.want {
			t.Errorf("theme %s, background %q: style %s, want %s", tt.choice, tt.background, st.style, tt.want)
		}
	}
}

// A page action shows a page of the file screen and waits for keys again (section 5.1).
// The items take the fields of the page. Each other action leaves the screen.
func TestRunFile(t *testing.T) {
	fsys := fstest.MapFS{
		"screens/main.bbt": {Data: []byte("+++\n[[item]]\nkey = \"R\"\ndo = \"goto read\"\n" +
			"args = { file = \"a.txt\" }\n+++\nMain\n")},
		"screens/read.bbt": {Data: []byte(`+++
file = "{arg.file}"
[[item]]
key = "N"
do = "page next"
when = "not file.last"
[[item]]
key = "P"
do = "page prev"
[[item]]
key = "F"
do = "page first"
[[item]]
key = "L"
do = "page last"
[[item]]
key = "Q"
do = "goto main"
+++
{file.page}/{file.pages}{#if item.N} N{/if}
{>file}
`)},
		"content/a.txt": {Data: []byte(lines(20))},
	}
	theme := map[string]screen.Style{"normal": {}, "more": {}, "md_code": {}, "md_dim": {}}
	ss, ds := bbt.Load(fsys, theme, screen.Layout{Width: screen.WidthLimits{Min: 20, Max: 80}}, nil)
	ds = Check(config.Config{Start: "main"}, ss, ds)
	if len(ds) > 0 {
		t.Fatalf("load: %v", ds)
	}
	// 10 rows: H is 9, so 20 lines give 3 pages.
	ft := &fakeTerm{cols: []int{80}, keys: []string{"R", "N", "N", "N", "P", "P", "P", "L", "F", "Q"}}
	err := Run(eventsTerm{ft}, config.Config{Start: "main"}, ss, "main", nil, noValues)
	want := []string{
		"screen main", "page: Main", "key R", "screen read", "page: 1/3 N",
		"key N", "page: 2/3 N", "key N", "page: 3/3", "key N ignored",
		"key P", "page: 2/3 N", "key P", "page: 1/3 N", "key P", "page: 1/3 N",
		"key L", "page: 3/3", "key F", "page: 1/3 N", "key Q", "screen main", "page: Main",
	}
	if !errors.Is(err, io.EOF) || !slices.Equal(ft.log, want) {
		t.Errorf("err %v, log:\n%s\nwant:\n%s", err, strings.Join(ft.log, "\n"), strings.Join(want, "\n"))
	}

	// Pages gives each file page.
	var got []string
	for p := range Pages(config.Config{}, ss["read"], ft.Term(), map[string]string{"file": "a.txt"}, screen.SessionValues{}) {
		got = append(got, fmt.Sprintf("%d:%d", p.Num, len(p.Rows)))
	}
	if want := []string{"1:10", "2:10", "3:10"}; !slices.Equal(got, want) {
		t.Errorf("Pages %q, want %q", got, want)
	}
}

// The session takes the session and BBS fields once for each screen, before the when
// of the items, and for each page of a file screen (session-fields, section 2).
func TestRunValues(t *testing.T) {
	fsys := fstest.MapFS{
		"screens/main.bbt": {Data: []byte(`+++
[[item]]
key = "R"
do = "goto read"
when = "bbs.sessions == 1"
[[item]]
key = "Q"
do = "goto main"
+++
main {bbs.sessions}{#if item.R} R{/if}
`)},
		"screens/read.bbt": {Data: []byte(`+++
file = "a.txt"
[[item]]
key = "N"
do = "page next"
[[item]]
key = "Q"
do = "goto main"
+++
read {bbs.sessions} {file.page}
{>file}
`)},
		"content/a.txt": {Data: []byte(lines(20))},
	}
	theme := map[string]screen.Style{"normal": {}, "more": {}, "md_code": {}, "md_dim": {}}
	ss, ds := bbt.Load(fsys, theme, screen.Layout{Width: screen.WidthLimits{Min: 20, Max: 80}}, nil)
	ds = Check(config.Config{Start: "main"}, ss, ds)
	if len(ds) > 0 {
		t.Fatalf("load: %v", ds)
	}
	// Each call gives the next count. The file screen takes one count for its layout.
	n := 0
	values := func() screen.SessionValues {
		n++
		return screen.SessionValues{Sessions: n}
	}
	ft := &fakeTerm{cols: []int{80}, keys: []string{"R", "N", "Q", "R"}}
	err := Run(ft, config.Config{Start: "main"}, ss, "main", nil, values)
	want := []string{"page: main 1 R", "page: read 3 1", "page: read 4 2", "page: main 5"}
	if !errors.Is(err, io.EOF) || !slices.Equal(ft.log, want) {
		t.Errorf("err %v, log %q, want %q", err, ft.log, want)
	}
}

// Shutdown shows the first page of shutdown with the size of the terminal (section 5.2).
func TestShutdown(t *testing.T) {
	ss := load(t, map[string]string{"shutdown": "Restart {term.cols} {bbs.name}\n" + lines(12)})
	ft := &fakeTerm{cols: []int{40}}
	values := func() screen.SessionValues { return screen.SessionValues{Name: "Lamp"} }
	if err := Shutdown(ft, config.Config{More: "-- More --"}, ss, values); err != nil {
		t.Fatal(err)
	}
	if want := []string{"page: Restart 40 Lamp"}; !slices.Equal(ft.log, want) {
		t.Errorf("log %q, want %q", ft.log, want)
	}

	// Without the file, Shutdown shows nothing.
	ft = &fakeTerm{cols: []int{40}}
	if err := Shutdown(ft, config.Config{}, load(t, nil), noValues); err != nil || len(ft.log) > 0 {
		t.Errorf("no shutdown: error %v, log %q; want no error and no page", err, ft.log)
	}

	// Shutdown gives the error of the terminal.
	errClosed := errors.New("closed")
	ft = &fakeTerm{cols: []int{40}, pageErr: errClosed}
	if err := Shutdown(ft, config.Config{}, ss, values); !errors.Is(err, errClosed) {
		t.Errorf("error %v, want %v", err, errClosed)
	}
}
