package bbt_test

import (
	"encoding/binary"
	"fmt"
	"io/fs"
	"reflect"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/schors/velourbbs/internal/bbt"
	"github.com/schors/velourbbs/internal/menu"
	"github.com/schors/velourbbs/internal/screen"
)

func style(text string) screen.Style {
	spec, err := screen.ParseStyle(text)
	if err != nil {
		panic(err)
	}
	return spec.Over(screen.Style{})
}

var (
	theme = map[string]screen.Style{
		"normal": style("7/0"), "more": style("15/0"), "title": style("15/4 bold"), "prompt": style("14/0"),
	}
	defaultWidth = screen.WidthLimits{Min: 40, Max: 80}
	// aliases is the [fields] of screens/screens.toml in the tests.
	aliases = screen.Fields{
		"w": alias("w", "term.cols"), "t": alias("t", "term.type"),
		"clock": alias("clock", `bbs.time live "15:04:05"`), "up": alias("up", "bbs.uptime live"),
		"name": alias("name", "bbs.name live"),
	}
)

func alias(name, path string) screen.Alias {
	r, err := screen.NewAlias(name, path)
	if err != nil {
		panic(err)
	}
	return r
}

// sauceArt gives an .ans file in the flat form: a BOM, data, 0x1A and a character SAUCE of the width.
func sauceArt(data string, width int) string {
	r := make([]byte, 128)
	copy(r, "SAUCE00")
	r[94], r[95] = 1, 1
	binary.LittleEndian.PutUint16(r[96:], uint16(width))
	return "\xef\xbb\xbf" + data + "\x1a" + string(r)
}

// load loads the files, given by their names in screens/.
func load(files map[string]string) (map[string]*screen.Screen, []string) {
	fsys := fstest.MapFS{}
	for name, data := range files {
		fsys["screens/"+name] = &fstest.MapFile{Data: []byte(data)}
	}
	sets, ds := bbt.Load(fsys, bbt.Settings{Theme: theme, Layout: screen.Layout{Width: defaultWidth}, Fields: aliases})
	screens := sets["en"]
	var msgs []string
	for _, d := range ds {
		msgs = append(msgs, d.String())
	}
	return screens, msgs
}

func mustLoad(t *testing.T, files map[string]string) map[string]*screen.Screen {
	t.Helper()
	screens, msgs := load(files)
	if screens == nil || len(msgs) > 0 {
		t.Fatalf("diags:\n%s", strings.Join(msgs, "\n"))
	}
	return screens
}

// show writes segments in a short form for the tests.
func show(segs []screen.Seg) string {
	base := style("1/2 underline")
	var out []string
	for _, s := range segs {
		switch s := s.(type) {
		case screen.Text:
			var b strings.Builder
			for _, g := range s.Graphemes {
				b.WriteString(g.Text)
			}
			out = append(out, fmt.Sprintf("%q", b.String()))
		case screen.Field:
			w := ""
			if s.Width > 0 {
				w = fmt.Sprintf(":%d", s.Width)
				if s.Right {
					w += "R"
				}
			}
			if s.Live {
				w += " live"
			}
			if s.Layout != "" {
				w += fmt.Sprintf(" %q", s.Layout)
			}
			out = append(out, "{"+s.Ref.String()+w+"}")
		case screen.SetStyle:
			if s.Name != "" {
				out = append(out, "!"+s.Name)
			} else {
				out = append(out, "!"+s.Spec.Over(base).String())
			}
		case screen.Stretch:
			out = append(out, fmt.Sprintf("~%q", s.Fill.Text))
		case screen.Spacer:
			out = append(out, "~~")
		case screen.If:
			out = append(out, "if("+show(s.Then)+")else("+show(s.Else)+")")
		case screen.Include:
			out = append(out, s.File+"("+show(s.Body)+")")
		case screen.ListLine:
			out = append(out, "list("+show(s.Row)+")")
		case screen.ArtInsert:
			var refs []string
			for _, f := range s.Art.Fields {
				refs = append(refs, f.Ref.String())
			}
			out = append(out, fmt.Sprintf("art%d%v", s.Art.Width, refs))
		case screen.FileArea:
			out = append(out, "{>file}")
		case screen.LineEnd:
			if s.Block {
				out = append(out, "/b")
			} else {
				out = append(out, "/")
			}
		}
	}
	return strings.Join(out, " ")
}

func TestExample(t *testing.T) {
	screens := mustLoad(t, map[string]string{
		"main.bbt": `+++
width = [64, 100]
[[item]]
key = "A"
do = "goto about"
+++
{#if term.cols >= 100}
{>main_wide.ans}
{#else}
{>main.ans}
{/if}
{!title}Welcome! Your terminal has {w} columns.
{!prompt}[A]bout [G]oodbye
`,
		"main_wide.ans": sauceArt("x", 100),
		"main.ans":      sauceArt("{w}", 64),
	})
	s := screens["main"]
	if s.Width != (screen.WidthLimits{Min: 64, Max: 100}) {
		t.Errorf("width %v", s.Width)
	}
	if !reflect.DeepEqual(s.Items, []screen.Item{{
		Key: "A", Do: screen.Action{Op: screen.Goto, Screen: "about"},
	}}) {
		t.Errorf("items %+v", s.Items)
	}
	want := `if(/b art100[])else(/b art64[term.cols]) /b !title "Welcome! Your terminal has " {term.cols} " columns." / !prompt "[A]bout [G]oodbye" /`
	if got := show(s.Body); got != want {
		t.Errorf("body\n got %s\nwant %s", got, want)
	}
	if s.File != "screens/main.bbt" || s.Name != "main" {
		t.Errorf("screen %+v", s)
	}
}

func TestTokens(t *testing.T) {
	tests := []struct {
		body string
		want string // the body, or the first diagnostic
	}{
		{"a{{b}", `"a{b}" /`},
		{"{{{term.cols}", `"{" {term.cols} /`},
		{"{1} {} {Ж} {", `"{1} {} {Ж} {" /`},
		// Stretch tokens and spacer lines (section 6.6).
		{"{~.}", `~"." /`},
		{"{~ }", `~" " /`},
		{"a{~}b{~─}", `"a" ~" " "b" ~"─" /`},
		{"{~}}", `~" " "}" /`},
		{"{~e\u0301}", "~\"e\u0301\" /"},
		{"{~}\n {~}\n{~}x", `~~ " " ~" " / ~" " "x" /`},
		{"{#if term.cols > 1}\n{~}\n{/if}", `if(/b ~~)else() /b`},
		{"{#if term.cols > 1}{~.}{/if}", `if(~".")else() /b`},
		{"ab{~..}", `screens/s.bbt:1:3: invalid stretch token "{~..}", want {~} or {~<c>} with one character of width 1`},
		{"{~世}", `screens/s.bbt:1:1: invalid stretch token "{~世}", want {~} or {~<c>} with one character of width 1`},
		{"{~\t}", `screens/s.bbt:1:3: control character U+0009`},
		{"{~\u0301}", "screens/s.bbt:1:1: invalid stretch token \"{~\u0301}\", want {~} or {~<c>} with one character of width 1"},
		{"a {-}", `screens/s.bbt:1:3: unknown token "{-}"`},
		{"{%x", `screens/s.bbt:1:1: token has no "}"`},
		{"a}b", `"a}b" /`},
		{"", ``},
		{"{term.type      }", `{term.type:17} /`},
		{"{   term.cols}x", `{term.cols:14R} "x" /`},
		{"{!}{!14}{!/3}{!title}", `!normal !14/2 underline !1/3 underline !title /`},
		{"{#if term.cols > 1}a{#else}b{/if}c", `if("a")else("b") "c" /b`},
		{"{#if term.cols > 1}\n{#if term.rows > 1}a{/if}\n{/if}", `if(/b if("a")else() /b)else() /b`},
		{"{term.cols", `screens/s.bbt:1:1: token has no "}"`},
		{"ab{ term.cols }", `screens/s.bbt:1:3: field token "{ term.cols }" has spaces on both sides`},
		{"{ }", `screens/s.bbt:1:1: field token "{ }" has spaces on both sides`},
		{"{Term}", `screens/s.bbt:1:1: invalid field path "Term"`},
		{"{nope}", `screens/s.bbt:1:1: unknown field "nope"`},
		{"{item.A}", `screens/s.bbt:1:1: field "item.A" is valid only in {#if}`},
		{"яя{nope}", `screens/s.bbt:1:3: unknown field "nope"`},
		{"{#if term.cols > 1}", `screens/s.bbt:1:1: {#if} has no {/if}`},
		{"a\n{/if}", `screens/s.bbt:2:1: {/if} without {#if}`},
		{"{#else}", `screens/s.bbt:1:1: {#else} without {#if}`},
		{"{#if term.cols>1}{#else}{#else}{/if}", `screens/s.bbt:1:25: second {#else} in one {#if}`},
		{"{#if}{/if}", `screens/s.bbt:1:1: expected a field, found end of condition`},
		{"{#iff}", `screens/s.bbt:1:1: unknown token "{#iff}"`},
		{"{@m}x{/@}", `screens/s.bbt:1:1: marks are not supported`},
		{"{!nope}", `screens/s.bbt:1:1: unknown style "nope"`},
		{"{!14 bold}", `screens/s.bbt:1:1: style token "{!14 bold}" has attributes`},
		{"{!300}", `screens/s.bbt:1:1: invalid style "300": invalid colour "300"`},
		{"a\tb", `screens/s.bbt:1:2: control character U+0009`},
		{"a\rb", `screens/s.bbt:1:2: control character U+000D`},
		{"a\r", `screens/s.bbt:1:2: control character U+000D`},
		{"{term.cols}\u0301", `screens/s.bbt:1:12: zero-width grapheme U+0301`},
		{"x{>_f}", `screens/s.bbt:1:2: an include or an art insert must be the only content of its line`},
		{"\xff", `screens/s.bbt:1:1: invalid UTF-8`},
		{"ab\nя\xff", `screens/s.bbt:2:2: invalid UTF-8`},
	}
	for _, tt := range tests {
		screens, msgs := load(map[string]string{"s.bbt": tt.body})
		got := ""
		if screens != nil {
			got = show(screens["s"].Body)
		} else if len(msgs) > 0 {
			got = msgs[0]
		}
		if got != tt.want {
			t.Errorf("%q:\n got %s\nwant %s\nall %q", tt.body, got, tt.want, msgs)
		}
	}
}

func TestFileForm(t *testing.T) {
	// A BOM, CR LF and no LF at the end.
	screens := mustLoad(t, map[string]string{"s.bbt": "\ufeff+++\r\nwidth = [10, 20]\r\n+++\r\na\r\nb"})
	if got := show(screens["s"].Body); got != `"a" / "b" /` || screens["s"].Width.Max != 20 {
		t.Errorf("body %s, width %v", got, screens["s"].Width)
	}
}

func TestFrontMatter(t *testing.T) {
	// Each error has the position of its key or value (screen-render, section 3.1).
	tests := []struct {
		files map[string]string
		want  []string
	}{
		{map[string]string{"main.bbt": "+++\n\n\n[[item]]\nkey = \"S\"\ndo = \"logoff\"\nwhen = \"term.colours == 1\"\n+++\n"},
			[]string{`screens/main.bbt:7:8: item "S": when: unknown field "term.colours"`}},
		{map[string]string{"s.bbt": "+++\nwidth = 5\ncolor = 1\n+++\n"}, []string{`screens/s.bbt:2:9: width: must be an array of integers`}},
		{map[string]string{"s.bbt": "+++\nwidth = [0, 300]\nnarrow = \"small\"\n+++\n"}, []string{`screens/s.bbt:3:1: unknown key "narrow"`}},
		{map[string]string{"s.bbt": "+++\nwidth = [0, 300]\n+++\n"},
			[]string{`screens/s.bbt:2:9: width: width [0, 300] is not in 1 <= min <= max <= 255`}},
		{map[string]string{"s.bbt": "+++\n\nwidth = [1]\n+++\n"}, []string{`screens/s.bbt:3:9: width: must be an array of 2 integers`}},
		{map[string]string{"s.bbt": "+++\nalign = \"middle\"\nvalign = \"center\"\nfill = \"clear\"\n+++\n"}, []string{
			`screens/s.bbt:2:9: align: invalid value "middle", want "left", "center" or "right"`,
			`screens/s.bbt:3:10: valign: invalid value "center", want "top", "middle" or "bottom"`,
			`screens/s.bbt:4:8: fill: invalid value "clear", want "none", "spaces" or "erase"`,
		}},
		{map[string]string{"s.bbt": "+++\nfill = 1\n+++\n"}, []string{`screens/s.bbt:2:8: fill: must be a string`}},
		{map[string]string{"s.bbt": "+++\ngutter = 101\n+++\n"}, []string{`screens/s.bbt:2:10: gutter: 101, must be from 0 to 100`}},
		{map[string]string{"s.bbt": "+++\ngutter = -1\n+++\n"}, []string{`screens/s.bbt:2:10: gutter: -1, must be from 0 to 100`}},
		{map[string]string{"s.bbt": "+++\ngutter = \"6\"\n+++\n"}, []string{`screens/s.bbt:2:10: gutter: must be an integer`}},
		// The column counts code points.
		{map[string]string{"s.bbt": "+++\nitem = [{ key = \"яя\", do = 1 }]\n+++\n"}, []string{`screens/s.bbt:2:28: item: invalid type`}},
		// An alias of screens/screens.toml in when: a type error at the value of when.
		{map[string]string{"s.bbt": "+++\n[[item]]\nkey = \"A\"\ndo = \"logoff\"\nwhen = \"t > 1\"\n+++\n"},
			[]string{`screens/s.bbt:5:8: item "A": when: string field "term.type" takes a string, found an integer`}},
		{map[string]string{"s.bbt": "+++\n[fields]\nk = \"item.A\"\n[[item]]\ndo = 1\nx = \"y\"\n+++\n"}, []string{`screens/s.bbt:5:6: item.do: must be a string`}},
		// The aliases are in screens/screens.toml: [fields] in a front matter is an unknown key (section 3.2).
		{map[string]string{"s.bbt": "+++\n[fields]\nk = \"item.A\"\n[[item]]\ny = 1\nx = 2\n+++\n"}, []string{
			`screens/s.bbt:2:2: unknown key "fields"`,
			`screens/s.bbt:5:1: unknown key "item.y"`,
			`screens/s.bbt:6:1: unknown key "item.x"`,
		}},
		{map[string]string{"s.bbt": "+++\nwidth = [1, 0]\n[[item]]\n+++\n"}, []string{
			// A missing key at the [[item]] header.
			`screens/s.bbt:2:9: width: width [1, 0] is not in 1 <= min <= max <= 255`,
			`screens/s.bbt:3:1: item 1: key is missing`,
			`screens/s.bbt:3:1: item 1: do is missing`,
		}},
		// A key that matches only without case and a table in place of [[item]] have the position of the key.
		{map[string]string{"s.bbt": "+++\n[[item]]\nKEY = \"a\"\nkey = \"b\"\ndo = \"logoff\"\n+++\n"}, []string{`screens/s.bbt:3:1: unknown key "item.KEY"`}},
		{map[string]string{"s.bbt": "+++\n[item]\nkey = \"a\"\ndo = \"logoff\"\n+++\n"}, []string{`screens/s.bbt:2:2: item: must be an array of tables`}},
		{map[string]string{"s.bbt": "+++\nwidth = [\n+++\n"}, []string{`screens/s.bbt:2:9: array is incomplete`}},
		{map[string]string{"s.bbt": "+++\n" + strings.Repeat("#\n", 33<<10) + "+++\n"}, []string{`screens/s.bbt:2:1: TOML document is over 65536 bytes`}},
		{map[string]string{"s.bbt": "+++\nwidth = [1, 2]\n"}, []string{`screens/s.bbt:1:1: front matter has no closing "+++"`}},
		{map[string]string{"s.bbt": "{>_f}", "_f.bbt": "+++\n+++\nx"}, []string{`screens/_f.bbt:1:1: a fragment cannot have front matter`}},
		{map[string]string{"Main.bbt": "", "_x.bbt": "\xff", "x.txt": "", "a.bbt.ans": ""}, []string{
			`screens/Main.bbt:1:1: invalid screen name "Main"`,
			`screens/_x.bbt:1:1: warning: no screen uses this file`,
			`screens/a.bbt.ans:1:1: warning: no screen uses this file`,
		}},
	}
	for _, tt := range tests {
		_, msgs := load(tt.files)
		if !slices.Equal(msgs, tt.want) {
			t.Errorf("%q:\n got %q\nwant %q", tt.files, msgs, tt.want)
		}
	}
}

func TestItems(t *testing.T) {
	screens := mustLoad(t, map[string]string{"s.bbt": `+++
[[item]]
key = "esc"
do = "goto main"
[[item]]
key = "any"
do = "logoff"
when = "w > 80"
[[item]]
key = "1"
do = "theme dark"
[[item]]
key = "2"
do = "theme light"
[[item]]
key = "3"
do = "theme auto"
[[item]]
key = "4"
do = "gosub info"
args = { f = "a" }
[[item]]
key = "5"
do = "back"
[[item]]
key = "6"
do = "top"
+++
`})
	items := screens["s"].Items
	if len(items) != 8 || items[1].When == nil {
		t.Fatalf("items %+v", items)
	}
	var got []screen.Item
	for _, it := range items {
		got = append(got, screen.Item{Key: it.Key, Do: it.Do})
	}
	want := []screen.Item{
		{Key: "ESC", Do: screen.Action{Op: screen.Goto, Screen: "main"}},
		{Key: "ANY", Do: screen.Action{Op: screen.Logoff}},
		{Key: "1", Do: screen.Action{Op: screen.Theme, Theme: "dark"}},
		{Key: "2", Do: screen.Action{Op: screen.Theme, Theme: "light"}},
		{Key: "3", Do: screen.Action{Op: screen.Theme, Theme: "auto"}},
		{Key: "4", Do: screen.Action{Op: screen.Gosub, Screen: "info", Args: map[string]string{"f": "a"}}},
		{Key: "5", Do: screen.Action{Op: screen.Back}},
		{Key: "6", Do: screen.Action{Op: screen.Top}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("items %+v, want %+v", got, want)
	}
}

func TestItemErrors(t *testing.T) {
	// Each case is one [[item]] table; the front matter starts on line 2 of the file.
	tests := []struct {
		items []string // the lines of each [[item]] table
		want  []string
	}{
		// An invalid key or do has the position of its value. An item without a valid key has its number.
		{[]string{`key = "?"`, `do = "logoff"`}, []string{`screens/s.bbt:3:7: item 1: key: invalid key "?"`}},
		{[]string{`key = "ы"`, `do = "logoff"`}, []string{`screens/s.bbt:3:7: item 1: key: invalid key "ы"`}},
		{[]string{`key = "F13"`, `do = "x"`}, []string{
			`screens/s.bbt:3:7: item 1: key: invalid key "F13"`,
			`screens/s.bbt:4:6: item 1: do: unknown action "x"`,
		}},
		// A type error hides the other errors.
		{[]string{`do = "x"`, `key = 1`}, []string{`screens/s.bbt:4:7: item.key: must be a string`}},
		{[]string{`key = true`}, []string{`screens/s.bbt:3:7: item.key: must be a string`}},
		{[]string{`key = 1.5`}, []string{`screens/s.bbt:3:7: item.key: must be a string`}},
		{[]string{`key = {}`}, []string{`screens/s.bbt:3:7: item.key: must be a string`}},
		{[]string{`do = "logoff"`, `[item.key]`}, []string{`screens/s.bbt:4:2: item.key: must be a string`}},
		{[]string{`key = "esc"`, `do = "Goto main"`}, []string{`screens/s.bbt:4:6: item "ESC": do: unknown action "Goto"`}},
		{[]string{`key = "A"`, `do = ""`}, []string{`screens/s.bbt:4:6: item "A": do: unknown action ""`}},
		{[]string{`key = "A"`, `do = "goto"`}, []string{`screens/s.bbt:4:6: item "A": do: goto needs a screen name`}},
		{[]string{`key = "A"`, `do = "goto  main"`}, []string{`screens/s.bbt:4:6: item "A": do: invalid screen name " main"`}},
		{[]string{`key = "A"`, `do = "goto Main"`}, []string{`screens/s.bbt:4:6: item "A": do: invalid screen name "Main"`}},
		{[]string{`key = "A"`, `do = "goto _f"`}, []string{`screens/s.bbt:4:6: item "A": do: invalid screen name "_f"`}},
		{[]string{`key = "A"`, `do = "gosub"`}, []string{`screens/s.bbt:4:6: item "A": do: gosub needs a screen name`}},
		{[]string{`key = "A"`, `do = "gosub Main"`}, []string{`screens/s.bbt:4:6: item "A": do: invalid screen name "Main"`}},
		{[]string{`key = "A"`, `do = "back main"`}, []string{`screens/s.bbt:4:6: item "A": do: back takes no argument`}},
		{[]string{`key = "A"`, `do = "top "`}, []string{`screens/s.bbt:4:6: item "A": do: top takes no argument`}},
		{[]string{`key = "A"`, `do = "Back"`}, []string{`screens/s.bbt:4:6: item "A": do: unknown action "Back"`}},
		{[]string{`key = "A"`, `do = "back"`, `args = { f = "a" }`}, []string{
			`screens/s.bbt:5:8: item "A": args: only goto and gosub take arguments`,
		}},
		{[]string{`key = "A"`, `do = "top"`, `args = { f = "a" }`}, []string{
			`screens/s.bbt:5:8: item "A": args: only goto and gosub take arguments`,
		}},
		{[]string{`key = "A"`, `do = "logoff now"`}, []string{`screens/s.bbt:4:6: item "A": do: logoff takes no argument`}},
		{[]string{`key = "A"`, `do = "logoff "`}, []string{`screens/s.bbt:4:6: item "A": do: logoff takes no argument`}},
		{[]string{`key = "A"`, `do = "theme"`}, []string{`screens/s.bbt:4:6: item "A": do: invalid theme action "theme"`}},
		{[]string{`key = "A"`, `do = "theme blue"`}, []string{`screens/s.bbt:4:6: item "A": do: invalid theme action "theme blue"`}},
		{[]string{`key = "A"`, `do = "theme Dark"`}, []string{`screens/s.bbt:4:6: item "A": do: invalid theme action "theme Dark"`}},
		{[]string{`key = "A"`, `do = "theme  dark"`}, []string{`screens/s.bbt:4:6: item "A": do: invalid theme action "theme  dark"`}},
		{[]string{`key = "A"`, `do = "theme dark"`, `args = { f = "a" }`}, []string{
			`screens/s.bbt:5:8: item "A": args: only goto and gosub take arguments`,
		}},
		{[]string{`key = "a"`, `do = 1`, `when = "item.A"`, `x = 1`}, []string{`screens/s.bbt:4:6: item.do: must be a string`}},
		// Each when error has the position of the value of when, also a syntax error.
		// A missing key has the position of the [[item]] header.
		{[]string{`key = "a"`, `do = "logoff"`, `when = "item.A"`}, []string{
			`screens/s.bbt:5:8: item "A": when: field "item.A" is valid only in {#if}`,
		}},
		{[]string{`key = "a"`, `when = "term.cols >"`}, []string{
			`screens/s.bbt:2:1: item "A": do is missing`,
			`screens/s.bbt:4:8: item "A": when: expected an integer or a string, found end of condition`,
		}},
		// The args errors of section 8.4 have the position of the value of args or of the argument,
		// an invalid argument name the position of its key.
		{[]string{`key = "A"`, `do = "logoff"`, `args = { file = "a.md" }`}, []string{
			`screens/s.bbt:5:8: item "A": args: only goto and gosub take arguments`,
		}},
		{[]string{`key = "A"`, `do = "x"`, `args = { file = "a.md" }`}, []string{
			`screens/s.bbt:4:6: item "A": do: unknown action "x"`,
		}},
		{[]string{`key = "A"`, `do = "x"`, `args = { File = "{" }`}, []string{
			`screens/s.bbt:4:6: item "A": do: unknown action "x"`,
			`screens/s.bbt:5:10: item "A": args: invalid argument name "File"`,
		}},
		{[]string{`key = "A"`, `do = "logoff"`, `args = { f = "\u202e" }`}, []string{
			`screens/s.bbt:5:8: item "A": args: only goto and gosub take arguments`,
			`screens/s.bbt:5:14: item "A": args: f: value has a control character`,
		}},
		{[]string{`key = "A"`, `do = "goto r"`, `args = { File = "a.md", 2f = "b", "a-b" = "c" }`}, []string{
			`screens/s.bbt:5:10: item "A": args: invalid argument name "File"`,
			`screens/s.bbt:5:25: item "A": args: invalid argument name "2f"`,
			`screens/s.bbt:5:35: item "A": args: invalid argument name "a-b"`,
		}},
		{[]string{`key = "A"`, `do = "goto r"`, `args = { a = "x{y", b = "x\ty", c = "x\u0085" }`}, []string{
			`screens/s.bbt:5:14: item "A": args: a: value has "{"`,
			`screens/s.bbt:5:25: item "A": args: b: value has a control character`,
			`screens/s.bbt:5:37: item "A": args: c: value has a control character`,
		}},
		// A value is a literal, {arg.<name>} or {in}: each other token is an error (bbt-format, section 8.5).
		// {in} is valid only in an item with ask.
		{[]string{`key = "A"`, `do = "goto r"`, `args = { a = "{in}", b = "{row.path}", c = "{arg.x  }", d = "{t}", e = "{arg.x", F = "{arg.x}", g = "{in }" }`}, []string{
			`screens/s.bbt:5:14: item "A": args: a: {in} is valid only in an item with ask`,
			`screens/s.bbt:5:26: item "A": args: b: invalid token "{row.path}", want a value, {arg.<name>} or {in}`,
			`screens/s.bbt:5:44: item "A": args: c: invalid token "{arg.x  }", want a value, {arg.<name>} or {in}`,
			`screens/s.bbt:5:61: item "A": args: d: invalid token "{t}", want a value, {arg.<name>} or {in}`,
			`screens/s.bbt:5:72: item "A": args: e: invalid token "{arg.x", want a value, {arg.<name>} or {in}`,
			`screens/s.bbt:5:82: item "A": args: invalid argument name "F"`,
			`screens/s.bbt:5:101: item "A": args: g: invalid token "{in }", want a value, {arg.<name>} or {in}`,
		}},
		// The errors of ask and ask_max (section 8.6). The prompt has the rules of a text.
		{[]string{`key = "A"`, `ask = "N\u0009: "`, `do = "goto r"`, `args = { n = "{in}" }`}, []string{
			`screens/s.bbt:4:7: item "A": ask: control character U+0009`,
		}},
		{[]string{`key = "A"`, `ask = "\u200b"`, `do = "goto r"`, `args = { n = "{in}" }`}, []string{
			`screens/s.bbt:4:7: item "A": ask: zero-width grapheme U+200B`,
		}},
		{[]string{`key = "A"`, `ask = "a` + strings.Repeat("\u0301", 20) + `"`, `do = "goto r"`, `args = { n = "{in}" }`}, []string{
			`screens/s.bbt:4:7: item "A": ask: grapheme over 32 bytes: 41 bytes`,
		}},
		{[]string{`key = "A"`, `ask = "N: "`, `ask_max = 0`, `do = "goto r"`, `args = { n = "{in}" }`}, []string{
			`screens/s.bbt:5:11: item "A": ask_max: 0, must be at least 1`,
		}},
		{[]string{`key = "A"`, `ask_max = 5`, `do = "goto r"`}, []string{
			`screens/s.bbt:4:11: item "A": ask_max: the item has no ask`,
		}},
		{[]string{`key = "A"`, `ask = "N: "`, `do = "back"`}, []string{
			`screens/s.bbt:4:7: item "A": ask: the action must be goto or gosub`,
		}},
		{[]string{`key = "A"`, `ask = "N: "`, `do = "goto r"`, `args = { n = "5" }`}, []string{
			`screens/s.bbt:4:7: item "A": ask: args has no value {in}`,
		}},
		{[]string{`key = "A"`, `ask = "N: "`, `do = "goto r"`}, []string{
			`screens/s.bbt:4:7: item "A": ask: args has no value {in}`,
		}},
		// The min width 40 leaves the free cells 40 - 38 - 1 = 1 after a prompt of 38 cells, and none after 39.
		{[]string{`key = "A"`, `ask = "` + strings.Repeat("ы", 37) + `語"`, `do = "goto r"`, `args = { n = "{in}" }`}, []string{
			`screens/s.bbt:4:7: item "A": ask: the prompt is 39 cells wide, the min width 40 leaves no free cell`,
		}},
		{[]string{`key = "A"`, `ask = 1`, `do = "goto r"`}, []string{`screens/s.bbt:4:7: item.ask: must be a string`}},
		{[]string{`key = "A"`, `ask = "N: "`, `ask_max = "5"`, `do = "goto r"`}, []string{
			`screens/s.bbt:5:11: item.ask_max: must be an integer`,
		}},
		{[]string{`key = "A"`, `do = "goto r"`, `args = { file = 1 }`}, []string{
			`screens/s.bbt:5:17: item.args: must be a table of strings`,
		}},
		{[]string{`key = "A"`, `do = "goto r"`, `args = "file"`}, []string{
			`screens/s.bbt:5:8: item.args: must be a table of strings`,
		}},
	}
	for _, tt := range tests {
		data := "+++\n[[item]]\n" + strings.Join(tt.items, "\n") + "\n+++\n"
		_, msgs := load(map[string]string{"s.bbt": data})
		if !slices.Equal(msgs, tt.want) {
			t.Errorf("%q:\n got %q\nwant %q", tt.items, msgs, tt.want)
		}
	}

	// Duplicate keys after the case rule, also ANY.
	data := `+++
[[item]]
key = "s"
do = "logoff"
[[item]]
key = "any"
do = "logoff"
[[item]]
key = "S"
do = "logoff"
[[item]]
key = "ANY"
do = "logoff"
+++
`
	want := []string{
		`screens/s.bbt:9:7: item "S": key: duplicate key`,
		`screens/s.bbt:12:7: item "ANY": key: duplicate key`,
	}
	if _, msgs := load(map[string]string{"s.bbt": data}); !slices.Equal(msgs, want) {
		t.Errorf("duplicates:\n got %q\nwant %q", msgs, want)
	}
}

func TestIncludes(t *testing.T) {
	screens := mustLoad(t, map[string]string{
		"main.bbt": "{!14}a\n{>_a}\nb\n{>_a}",
		"_a.bbt":   "{!title}{w}\n{>_b}\n",
		"_b.bbt":   "{#if term.cols > 1}\nc\n{/if}\n",
	})
	body := screens["main"].Body
	want := `!14/2 underline "a" / screens/_a.bbt(!title {term.cols} / screens/_b.bbt(if(/b "c" /)else() /b)) "b" / screens/_a.bbt(!title {term.cols} / screens/_b.bbt(if(/b "c" /)else() /b))`
	if got := show(body); got != want {
		t.Errorf("body\n got %s\nwant %s", got, want)
	}
	// An alias of screens/screens.toml means the same in each screen, fragment and art.
	screens = mustLoad(t, map[string]string{
		"s.bbt": "{>_f}", "t.bbt": "{>_f}\n{>a.ans}", "_f.bbt": "{w}\n{#if t == \"x\"}{/if}\n{>a.ans}", "a.ans": sauceArt("{t}", 80),
	})
	for _, name := range []string{"s", "t"} {
		if got := show(screens[name].Body); !strings.HasPrefix(got, `screens/_f.bbt({term.cols} / if()else() /b art80[term.type])`) {
			t.Errorf("%s: %s", name, got)
		}
	}

	// The style after an include: the style token of the fragment is inside the include,
	// so the render restores the style before it. The two includes share the body.
	a1, a2 := body[3].(screen.Include), body[6].(screen.Include)
	if _, ok := a1.Body[0].(screen.SetStyle); !ok || &a1.Body[0] != &a2.Body[0] {
		t.Errorf("includes %v %v", a1, a2)
	}

	tests := []struct {
		files map[string]string
		want  []string
	}{
		{map[string]string{"s.bbt": "{>_a}", "_a.bbt": "{>_b}", "_b.bbt": "x\n{>_a}"},
			[]string{`screens/_b.bbt:2:1: include "screens/_a.bbt" makes a cycle`}},
		{map[string]string{"s.bbt": "{>_s}", "_s.bbt": "{>_s}"},
			[]string{`screens/_s.bbt:1:1: include "screens/_s.bbt" makes a cycle`}},
		{map[string]string{"s.bbt": "{>main}\n{>Bad}"},
			[]string{`screens/s.bbt:1:1: include "main": a fragment name starts with "_"`, `screens/s.bbt:2:1: invalid include "{>Bad}"`}},
		{map[string]string{"s.bbt": "{>_x}\n{>x.ans}", "t.bbt": "{>_x}"},
			[]string{`screens/s.bbt:1:1: fragment "screens/_x.bbt" does not exist`, `screens/s.bbt:2:1: art "screens/x.ans" does not exist`, `screens/t.bbt:1:1: fragment "screens/_x.bbt" does not exist`}},
		// The errors of a fragment that two screens use, once.
		{map[string]string{"s.bbt": "{>_f}", "t.bbt": "{>_f}", "_f.bbt": "{>_x}\n{>x.ans}\n{>_f}"},
			[]string{`screens/_f.bbt:1:1: fragment "screens/_x.bbt" does not exist`, `screens/_f.bbt:2:1: art "screens/x.ans" does not exist`, `screens/_f.bbt:3:1: include "screens/_f.bbt" makes a cycle`}},
		// A front matter error: the field and include errors stay.
		{map[string]string{"s.bbt": "+++\nx = 1\n+++\n{w}{nope}\n{>_x}"},
			[]string{`screens/s.bbt:2:1: unknown key "x"`, `screens/s.bbt:4:4: unknown field "nope"`, `screens/s.bbt:5:1: fragment "screens/_x.bbt" does not exist`}},
		// A field error in a fragment or in art that two screens use: once, with no screen name.
		{map[string]string{"s.bbt": "{>_f}", "t.bbt": "{>_f}\n{>a.ans}", "_f.bbt": "{bad}\n{#if x}\n{>a.ans}", "a.ans": sauceArt("{w} {bad}", 80)},
			[]string{`screens/_f.bbt:1:1: unknown field "bad"`, `screens/_f.bbt:2:1: unknown field "x"`, `screens/_f.bbt:2:1: {#if} has no {/if}`, `screens/a.ans:1:5: unknown field "bad"`}},
		// A fragment that cannot be read exists: its error, once.
		{map[string]string{"s.bbt": "{>_f}", "t.bbt": "{>_f}", "_f.bbt": strings.Repeat("x", screen.MaxFileBytes+1)},
			[]string{`screens/_f.bbt:1:1: file is over 1048576 bytes`}},
	}
	for _, tt := range tests {
		_, msgs := load(tt.files)
		if !slices.Equal(msgs, tt.want) {
			t.Errorf("%q:\n got %q\nwant %q", tt.files, msgs, tt.want)
		}
	}
}

func TestLimits(t *testing.T) {
	// Each include counts its line and the 101 lines of the fragment.
	frag := strings.Repeat("x\n", 101)
	main := strings.Repeat("{>_f}\n", 100)
	_, msgs := load(map[string]string{"s.bbt": main, "_f.bbt": frag})
	want := []string{`screens/s.bbt:99:1: body is over 10000 lines after includes`}
	if !slices.Equal(msgs, want) {
		t.Errorf("lines: %q", msgs)
	}

	// A fan-out of includes: 2^40 lines after includes, each fragment counted once.
	files := map[string]string{"s.bbt": "a\n{>_f0}"}
	for i := range 40 {
		files[fmt.Sprintf("_f%d.bbt", i)] = fmt.Sprintf("{>_f%d}\n{>_f%d}\n", i+1, i+1)
	}
	files["_f40.bbt"] = "x"
	_, msgs = load(files)
	if want := []string{`screens/s.bbt:2:1: body is over 10000 lines after includes`}; !slices.Equal(msgs, want) {
		t.Errorf("fan-out: %q", msgs)
	}

	// A spacer line counts its line.
	_, msgs = load(map[string]string{"s.bbt": strings.Repeat("{~}\n", bbt.MaxBodyLines+1)})
	if want := []string{`screens/s.bbt:10001:1: body is over 10000 lines after includes`}; !slices.Equal(msgs, want) {
		t.Errorf("spacer lines: %q", msgs)
	}

	// Art counts its line and its rows.
	_, msgs = load(map[string]string{"s.bbt": strings.Repeat("{>a.ans}\n", 11), "a.ans": sauceArt(strings.Repeat("x\n", 999)+"x", 80)})
	if want := []string{`screens/s.bbt:10:1: body is over 10000 lines after includes`}; !slices.Equal(msgs, want) {
		t.Errorf("art: %q", msgs)
	}

	// A fan-out of empty fragments: each include line counts.
	_, msgs = load(map[string]string{"s.bbt": strings.Repeat("{>_a}\n", 101), "_a.bbt": strings.Repeat("{>_b}\n", 99), "_b.bbt": ""})
	if want := []string{`screens/s.bbt:101:1: body is over 10000 lines after includes`}; !slices.Equal(msgs, want) {
		t.Errorf("empty fragments: %q", msgs)
	}

	// A chain of includes: each include line counts, so the nesting stops at the line limit.
	// The screen pass does not read the last fragment (screen-render, section 3.1).
	files = map[string]string{"s.bbt": "{>_f1}"}
	for i := 1; i <= bbt.MaxBodyLines+1; i++ {
		files[fmt.Sprintf("_f%d.bbt", i)] = fmt.Sprintf("{>_f%d}", i+1)
	}
	_, msgs = load(files)
	want = []string{`screens/_f10001.bbt:1:1: warning: no screen uses this file`, `screens/s.bbt:1:1: body is over 10000 lines after includes`}
	if !slices.Equal(msgs, want) {
		t.Errorf("chain: %q", msgs)
	}

	// The screen pass stops at the limit, so it does not read _g.bbt:
	// the warning "no screen uses this file" is false (screen-render, section 3.1).
	_, msgs = load(map[string]string{"s.bbt": strings.Repeat("{>_f}\n", 100) + "{>_g}", "_f.bbt": frag, "_g.bbt": "x"})
	want = []string{`screens/_g.bbt:1:1: warning: no screen uses this file`, `screens/s.bbt:99:1: body is over 10000 lines after includes`}
	if !slices.Equal(msgs, want) {
		t.Errorf("after a stop: %q", msgs)
	}

	// After 100 diagnostics, the load goes on: the include after them gives the error of the fragment.
	_, msgs = load(map[string]string{"s.bbt": strings.Repeat("\t", 200) + "\n{>_f}", "_f.bbt": "{nope}"})
	if len(msgs) != 102 || msgs[0] != `screens/_f.bbt:1:1: unknown field "nope"` ||
		msgs[101] != "screens/s.bbt:1:101: too many diagnostics" {
		t.Errorf("diags: %d, first %q, last %q", len(msgs), msgs[0], msgs[len(msgs)-1])
	}

	_, msgs = load(map[string]string{"s.bbt": strings.Repeat("x", screen.MaxFileBytes+1)})
	if want := []string{`screens/s.bbt:1:1: file is over 1048576 bytes`}; !slices.Equal(msgs, want) {
		t.Errorf("file: %q", msgs)
	}
}

// Each value of align, valign and fill in the front matter overrides the key of [screen].
// A screen without the keys takes the keys of [screen] (section 3.2).
func TestLayout(t *testing.T) {
	base := screen.Layout{Width: defaultWidth, Align: screen.AlignRight, VAlign: screen.VAlignBottom, Fill: screen.FillErase, Gutter: 6}
	tests := []struct {
		front string
		want  func(*screen.Layout)
	}{
		{"", func(*screen.Layout) {}},
		{`align = "left"`, func(l *screen.Layout) { l.Align = screen.AlignLeft }},
		{`align = "center"`, func(l *screen.Layout) { l.Align = screen.AlignCenter }},
		{`align = "right"`, func(l *screen.Layout) { l.Align = screen.AlignRight }},
		{`valign = "top"`, func(l *screen.Layout) { l.VAlign = screen.VAlignTop }},
		{`valign = "middle"`, func(l *screen.Layout) { l.VAlign = screen.VAlignMiddle }},
		{`valign = "bottom"`, func(l *screen.Layout) { l.VAlign = screen.VAlignBottom }},
		{`fill = "none"`, func(l *screen.Layout) { l.Fill = screen.FillNone }},
		{`fill = "spaces"`, func(l *screen.Layout) { l.Fill = screen.FillSpaces }},
		{`fill = "erase"`, func(l *screen.Layout) { l.Fill = screen.FillErase }},
		{`gutter = 0`, func(l *screen.Layout) { l.Gutter = 0 }},
		{`gutter = 100`, func(l *screen.Layout) { l.Gutter = 100 }},
	}
	for _, tt := range tests {
		fsys := fstest.MapFS{"screens/s.bbt": {Data: []byte("+++\n" + tt.front + "\n+++\nx\n")}}
		sets, ds := bbt.Load(fsys, bbt.Settings{Theme: theme, Layout: base, Fields: aliases})
		screens := sets["en"]
		if len(ds) != 0 {
			t.Fatalf("%q: diags %v", tt.front, ds)
		}
		want := base
		tt.want(&want)
		if got := screens["s"].Layout; got != want {
			t.Errorf("%q: layout %+v, want %+v", tt.front, got, want)
		}
	}
}

func TestWidth(t *testing.T) {
	// The art does not set the width limits (section 6.1).
	screens := mustLoad(t, map[string]string{
		"a.bbt": "{>x.ans}", "b.bbt": "+++\nwidth = [20, 30]\n+++\n{>x.ans}",
		"x.ans": sauceArt("x", 30),
	})
	if w := screens["a"].Width; w != defaultWidth {
		t.Errorf("a: %v", w)
	}
	if w := screens["b"].Width; w != (screen.WidthLimits{Min: 20, Max: 30}) {
		t.Errorf("b: %v", w)
	}

	// A screen without width takes [screen] width of screens/screens.toml.
	fsys := fstest.MapFS{
		"screens/a.bbt": {Data: []byte("{>x.ans}\n{>y.ans}")},
		"screens/x.ans": {Data: []byte(sauceArt("x", 50))},
		"screens/y.ans": {Data: []byte(sauceArt("y", 132))},
	}
	sets, ds := bbt.Load(fsys, bbt.Settings{Theme: theme, Layout: screen.Layout{Width: screen.WidthLimits{Min: 20, Max: 60}}, Fields: aliases})
	if w := sets["en"]["a"].Width; w != (screen.WidthLimits{Min: 20, Max: 60}) {
		t.Errorf("[screen] width: %v", w)
	}
	if want := `screens/a.bbt:2:1: warning: art "y.ans" is 132 cells wide, max width is 60`; len(ds) != 1 || ds[0].String() != want {
		t.Errorf("[screen] width: diags %v", ds)
	}

	tests := []struct {
		files map[string]string
		want  []string
	}{
		// Art of two widths and no width: the default limits, a warning for the wider art.
		{map[string]string{"s.bbt": "{#if term.cols > 100}\n{>x.ans}\n{#else}\n{>y.ans}\n{/if}", "x.ans": sauceArt("x", 132), "y.ans": sauceArt("y", 80)},
			[]string{`screens/s.bbt:2:1: warning: art "x.ans" is 132 cells wide, max width is 80`}},
		{map[string]string{"about.bbt": "+++\nwidth = [40, 100]\n+++\n{>logo.ans}\n{>_f}", "_f.bbt": "{>logo.ans}", "logo.ans": sauceArt("x", 132)},
			[]string{`screens/_f.bbt:1:1: warning: screen "about": art "logo.ans" is 132 cells wide, max width is 100`, `screens/about.bbt:4:1: warning: art "logo.ans" is 132 cells wide, max width is 100`}},
		{map[string]string{"s.bbt": "{>a.ans}", "a.ans": sauceArt("{nope}", 80)},
			[]string{`screens/a.ans:1:1: unknown field "nope"`}},
		{map[string]string{"s.bbt": "x", "_f.bbt": "{>a.ans}", "a.ans": "x", "b.ans": "\xff"},
			[]string{`screens/_f.bbt:1:1: warning: no screen uses this file`, `screens/a.ans:1:1: warning: no screen uses this file`, `screens/b.ans:1:1: warning: no screen uses this file`}},
		{map[string]string{"s.bbt": "{>a.ans}", "a.ans": sauceArt("\xff", 80)},
			[]string{`screens/a.ans:1:1: invalid UTF-8`}},
	}
	for _, tt := range tests {
		_, msgs := load(tt.files)
		if !slices.Equal(msgs, tt.want) {
			t.Errorf("%q:\n got %q\nwant %q", tt.files, msgs, tt.want)
		}
	}
}

func TestItemPaths(t *testing.T) {
	mustLoad(t, map[string]string{
		"main.bbt": "+++\n[[item]]\nkey = \"A\"\ndo = \"logoff\"\n+++\n{#if item.A}a{/if}\n{>_f}",
		"_f.bbt":   "{#if item.A}a{/if}",
	})

	tests := []struct {
		files map[string]string
		want  []string
	}{
		{map[string]string{"goodbye.bbt": "+++\nwidth = [40, 80]\n[[item]]\nkey = \"A\"\ndo = \"logoff\"\n+++\n"},
			[]string{`screens/goodbye.bbt:3:1: item: the goodbye screen cannot have items`}},
		{map[string]string{"shutdown.bbt": "+++\n[[item]]\nkey = \"A\"\ndo = \"logoff\"\n+++\n"},
			[]string{`screens/shutdown.bbt:2:1: item: the shutdown screen cannot have items`}},
		{map[string]string{
			"main.bbt": "+++\n[[item]]\nkey = \"A\"\ndo = \"logoff\"\n+++\n{#if item.A and item.B}a{/if}\n{>_f}",
			"_f.bbt":   "{#if item.C}a{/if}",
		}, []string{`screens/_f.bbt:1:1: screen "main": unknown field "item.C"`, `screens/main.bbt:6:1: unknown field "item.B"`}},
		{map[string]string{"s.bbt": "+++\n[[item]]\nkey = \"A\"\ndo = \"logoff\"\n+++\n{#if item.a or item.b}a{/if}"},
			[]string{`screens/s.bbt:6:1: unknown field "item.B"`}}, // the case rule: item.a is item.A
	}
	for _, tt := range tests {
		_, msgs := load(tt.files)
		if !slices.Equal(msgs, tt.want) {
			t.Errorf("%q:\n got %q\nwant %q", tt.files, msgs, tt.want)
		}
	}
}

func TestItemArgs(t *testing.T) {
	screens := mustLoad(t, map[string]string{"s.bbt": `+++
[[item]]
key = "1"
do = "goto r"
args = { file = "gold.md", n = "" }
[[item]]
key = "2"
do = "goto r"
[item.args]
file = "золото.md"
[[item]]
key = "3"
do = "goto r"
+++
`})
	var got []map[string]string
	for _, it := range screens["s"].Items {
		got = append(got, it.Do.Args)
	}
	want := []map[string]string{{"file": "gold.md", "n": ""}, {"file": "золото.md"}, nil}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("args %q, want %q", got, want)
	}
}

// ask gives the prompt and ask_max. {in} can be the value of more than one argument.
// The prompt has no tokens, and a prompt of 38 cells leaves 1 free cell at the min width 40 (section 8.6).
func TestItemAsk(t *testing.T) {
	screens := mustLoad(t, map[string]string{"s.bbt": `+++
[[item]]
key = "J"
ask = "{n}: "
ask_max = 5
do = "gosub r"
args = { area = "a", n = "{in}", m = "{in}" }
[[item]]
key = "K"
ask = "` + strings.Repeat("ы", 36) + `語"
do = "goto r"
args = { n = "{in}" }
[[item]]
key = "L"
ask = ""
do = "goto r"
args = { n = "{in}" }
+++
`})
	items := screens["s"].Items
	if a := items[0].Ask; a == nil || *a != (screen.Ask{Prompt: "{n}: ", Max: 5}) {
		t.Errorf("ask of J %+v", a)
	}
	if got := items[0].Do; !reflect.DeepEqual(got.Args, map[string]string{"area": "a"}) || !slices.Equal(got.In, []string{"m", "n"}) {
		t.Errorf("args of J %q, in %q", got.Args, got.In)
	}
	if a := items[1].Ask; a == nil || a.Width() != 38 || a.Max != 0 {
		t.Errorf("ask of K %+v", a)
	}
	if a := items[2].Ask; a == nil || a.Prompt != "" {
		t.Errorf("ask of L %+v", a)
	}
}

// The arguments of a screen come from its body, its fragments, its art,
// its {#if} and its when, also through an alias (section 8.5).
func TestScreenArgs(t *testing.T) {
	fsys := fstest.MapFS{
		"screens/s.bbt": {Data: []byte(`+++
[[item]]
key = "A"
do = "logoff"
when = 'arg.w == "1"'
+++
{arg.body} {f}
{#if arg.cond == "x"}{/if}
{>_frag}
{>a.ans}
{>_frag}
`)},
		"screens/_frag.bbt": {Data: []byte("{arg.frag}\n")},
		"screens/a.ans":     {Data: []byte(sauceArt(`{arg.art} {#if arg.box == "1"}x{/if}`, 40))},
		"screens/t.bbt":     {Data: []byte("+++\n[[item]]\nkey = \"A\"\ndo = \"logoff\"\n+++\n{term.cols}\n")},
	}
	fields := screen.Fields{"f": alias("f", "arg.alias")}
	sets, ds := bbt.Load(fsys, bbt.Settings{Theme: theme, Layout: screen.Layout{Width: defaultWidth}, Fields: fields})
	screens := sets["en"]
	if ds.HasErrors() {
		t.Fatalf("diags %v", ds)
	}
	want := []string{"alias", "art", "body", "box", "cond", "frag", "w"}
	if got := screens["s"].Args; !slices.Equal(got, want) {
		t.Errorf("args %q, want %q", got, want)
	}
	if got := screens["t"].Args; len(got) != 0 {
		t.Errorf("t: args %q", got)
	}
}

// An unused .bbt or .ans file that is not a regular file is an error (section 2).
func TestNotRegular(t *testing.T) {
	fsys := fstest.MapFS{
		"bbs.toml":          {},
		"screens/s.bbt":     {Data: []byte("x")},
		"screens/_f.bbt":    {Mode: fs.ModeNamedPipe},
		"screens/a.ans":     {Mode: fs.ModeNamedPipe},
		"screens/d.bbt/x":   {},
		"screens/dir/x.txt": {},
		"screens/notes.txt": {Mode: fs.ModeNamedPipe},
	}
	_, ds := bbt.Load(fsys, bbt.Settings{Theme: theme, Layout: screen.Layout{Width: defaultWidth}, Fields: aliases})
	var got []string
	for _, d := range ds {
		got = append(got, d.String())
	}
	want := []string{
		"screens/_f.bbt:1:1: not a regular file",
		"screens/a.ans:1:1: not a regular file",
		"screens/d.bbt:1:1: not a regular file",
	}
	if !slices.Equal(got, want) {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

// FuzzLoad loads a screen with a fragment and an art file. It renders a screen
// without errors on the smallest terminal and reads all pages.
func FuzzLoad(f *testing.F) {
	f.Add([]byte("{#if w > 1}{>_f}{/if}\n{>_f}\n{w   }"), []byte("{!title}{w}\n{>a.ans}"))
	f.Add([]byte("{>_f}\n{{{ x}{#else}"), []byte("{>_f}"))
	f.Add([]byte("+++\n[[item]]\nkey = \"A\"\ndo = \"logoff\"\nwhen = \"term.colors == 16\"\n+++\n"+
		"{#if item.A and t != \"x\"}{      t} {w}{/if}\n{>a.ans}\n"+strings.Repeat("{>_f}\n", 12)), []byte("{!more}line {t}"))
	f.Fuzz(func(t *testing.T, s, frag []byte) {
		fsys := fstest.MapFS{
			"screens/s.bbt":  {Data: s},
			"screens/_f.bbt": {Data: frag},
			"screens/a.ans":  {Data: []byte(sauceArt("{w}", 80))},
		}
		sets, ds := bbt.Load(fsys, bbt.Settings{Theme: theme, Layout: screen.Layout{Width: defaultWidth}, Fields: aliases})
		screens := sets["en"]
		if ds.HasErrors() {
			return
		}
		if screens["s"] == nil {
			t.Fatal("no screen without errors")
		}
		checkSegs(t, screens["s"].Body)
		env, _ := menu.Enabled(screens["s"], screen.Env{Term: screen.Term{
			Type: "ansi", Cols: screen.MinCols, Rows: screen.MinRows, Colors: screen.Colors16,
		}})
		for range screen.Render(screens["s"], env, theme, "more") {
		}
	})
}

// checkSegs checks that a model without errors is complete.
func checkSegs(t *testing.T, segs []screen.Seg) {
	for _, s := range segs {
		switch s := s.(type) {
		case screen.If:
			if s.Cond == nil {
				t.Fatal("if without a condition")
			}
			checkSegs(t, s.Then)
			checkSegs(t, s.Else)
		case screen.Include:
			checkSegs(t, s.Body)
		case screen.ArtInsert:
			if s.Art == nil {
				t.Fatal("art insert without art")
			}
			for _, f := range s.Art.Fields {
				if f.Ref.Type() == 0 {
					t.Fatal("art field without a type")
				}
			}
		case screen.Field:
			if s.Ref.Type() == 0 {
				t.Fatal("field without a type")
			}
		case screen.Text:
			if len(s.Graphemes) == 0 {
				t.Fatal("empty text")
			}
		}
	}
}
