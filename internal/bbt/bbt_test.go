package bbt_test

import (
	"encoding/binary"
	"fmt"
	"io/fs"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"newbbs/internal/bbt"
	"newbbs/internal/screen"
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
)

// sauceArt gives an .ans file with a character SAUCE of the width.
func sauceArt(data string, width int) string {
	r := make([]byte, 128)
	copy(r, "SAUCE00")
	r[94], r[95] = 1, 1
	binary.LittleEndian.PutUint16(r[96:], uint16(width))
	return data + "\x1a" + string(r)
}

// load loads the files, given by their names in screens/.
func load(files map[string]string) (map[string]*screen.Screen, []string) {
	fsys := fstest.MapFS{}
	for name, data := range files {
		fsys["screens/"+name] = &fstest.MapFile{Data: []byte(data)}
	}
	screens, ds := bbt.Load(fsys, theme, defaultWidth)
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
			out = append(out, "{"+s.Ref.String()+w+"}")
		case screen.SetStyle:
			out = append(out, "!"+s.Apply(base).String())
		case screen.If:
			out = append(out, "if("+show(s.Then)+")else("+show(s.Else)+")")
		case screen.Include:
			out = append(out, s.File+"("+show(s.Body)+")")
		case screen.ArtInsert:
			var refs []string
			for _, r := range s.Fields {
				refs = append(refs, r.String())
			}
			out = append(out, fmt.Sprintf("art%d%v", s.Art.Width, refs))
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
[fields]
w = "term.cols"
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
	if !slices.Equal(s.Items, []screen.Item{{
		Key: "A", Do: screen.Action{Op: screen.Goto, Screen: "about"},
	}}) {
		t.Errorf("items %+v", s.Items)
	}
	want := `if(/b art100[])else(/b art64[term.cols]) /b !15/4 bold "Welcome! Your terminal has " {term.cols} " columns." / !14/0 "[A]bout [G]oodbye" /`
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
		{"{1} {-} {Ж} {", `"{1} {-} {Ж} {" /`},
		{"a}b", `"a}b" /`},
		{"", ``},
		{"{term.type      }", `{term.type:17} /`},
		{"{   term.cols}x", `{term.cols:14R} "x" /`},
		{"{!}{!14}{!/3}{!title}", `!7/0 !14/2 underline !1/3 underline !15/4 bold /`},
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
		{strings.Repeat("{#if term.cols>1}", 101) + "x", `screens/s.bbt:1:1701: {#if} is nested over 100 levels`},
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
	// A TOML error has the position of its key or value; each other error has 2:1 (section 9.1).
	tests := []struct {
		files map[string]string
		want  []string
	}{
		{map[string]string{"main.bbt": "+++\n[fields]\nw = \"term.cols\"\n\n\n[[item]]\nkey = \"S\"\ndo = \"logoff\"\nwhen = \"term.colours == 1\"\n+++\n"},
			[]string{`screens/main.bbt:2:1: item "S": when: unknown field "term.colours"`}},
		{map[string]string{"s.bbt": "+++\nwidth = 5\ncolor = 1\n+++\n"}, []string{`screens/s.bbt:2:9: width: must be an array of integers`}},
		{map[string]string{"s.bbt": "+++\nwidth = [0, 300]\nnarrow = \"small\"\n+++\n"}, []string{`screens/s.bbt:3:1: unknown key "narrow"`}},
		{map[string]string{"s.bbt": "+++\nwidth = [0, 300]\n+++\n"},
			[]string{`screens/s.bbt:2:1: width: width [0, 300] is not in 1 <= min <= max <= 255`}},
		{map[string]string{"s.bbt": "+++\n\nwidth = [1]\n+++\n"}, []string{`screens/s.bbt:2:1: width: must be an array of 2 integers`}},
		{map[string]string{"s.bbt": "+++\nfields = { \"яя\" = \"term.cols\", and = \"term.rows\" }\n+++\n"},
			[]string{`screens/s.bbt:2:1: fields: alias "and" is a reserved word`, `screens/s.bbt:2:1: fields: invalid alias "яя"`}},
		// The column counts code points.
		{map[string]string{"s.bbt": "+++\nfields = { \"яя\" = \"term.cols\", w = 5 }\n+++\n"}, []string{`screens/s.bbt:2:36: fields: invalid type of a value`}},
		{map[string]string{"s.bbt": "+++\n[fields]\nk = \"item.A\"\n[[item]]\ndo = 1\nx = \"y\"\n+++\n"}, []string{`screens/s.bbt:5:6: item.do: must be a string`}},
		{map[string]string{"s.bbt": "+++\n[fields]\nk = \"item.A\"\n[[item]]\ny = 1\nx = 2\n+++\n"}, []string{
			`screens/s.bbt:5:1: unknown key "item.y"`,
			`screens/s.bbt:6:1: unknown key "item.x"`,
		}},
		{map[string]string{"s.bbt": "+++\n[fields]\nb = \"term.x\"\na = \"term.y\"\n[[item]]\n+++\n"}, []string{
			`screens/s.bbt:2:1: fields: alias "a": unknown field "term.y"`,
			`screens/s.bbt:2:1: fields: alias "b": unknown field "term.x"`,
			`screens/s.bbt:2:1: item 1: key is missing`,
			`screens/s.bbt:2:1: item 1: do is missing`,
		}},
		// A key that matches only without case and a table in place of [[item]] have 2:1.
		{map[string]string{"s.bbt": "+++\n[[item]]\nKEY = \"a\"\nkey = \"b\"\ndo = \"logoff\"\n+++\n"}, []string{`screens/s.bbt:2:1: unknown key "item.KEY"`}},
		{map[string]string{"s.bbt": "+++\n[item]\nkey = \"a\"\ndo = \"logoff\"\n+++\n"}, []string{`screens/s.bbt:2:1: item: must be an array of tables`}},
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
when = "term.cols > 80"
+++
`})
	items := screens["s"].Items
	if len(items) != 2 || items[1].When == nil {
		t.Fatalf("items %+v", items)
	}
	got := []screen.Item{
		{Key: items[0].Key, Do: items[0].Do},
		{Key: items[1].Key, Do: items[1].Do},
	}
	want := []screen.Item{
		{Key: "ESC", Do: screen.Action{Op: screen.Goto, Screen: "main"}},
		{Key: "ANY", Do: screen.Action{Op: screen.Logoff}},
	}
	if !slices.Equal(got, want) {
		t.Errorf("items %+v, want %+v", got, want)
	}
}

func TestItemErrors(t *testing.T) {
	// Each case is one [[item]] table; the front matter starts on line 2 of the file.
	tests := []struct {
		items []string // the lines of each [[item]] table
		want  []string
	}{
		{[]string{`key = "?"`, `do = "logoff"`}, []string{`screens/s.bbt:3:7: item.key: invalid key "?"`}},
		{[]string{`key = "ы"`, `do = "logoff"`}, []string{`screens/s.bbt:3:7: item.key: invalid key "ы"`}},
		{[]string{`key = "F13"`, `do = "x"`}, []string{`screens/s.bbt:3:7: item.key: invalid key "F13"`}},
		// A type error comes before an invalid value.
		{[]string{`do = "x"`, `key = 1`}, []string{`screens/s.bbt:4:7: item.key: must be a string`}},
		{[]string{`key = true`}, []string{`screens/s.bbt:3:7: item.key: must be a string`}},
		{[]string{`key = 1.5`}, []string{`screens/s.bbt:3:7: item.key: must be a string`}},
		{[]string{`key = {}`}, []string{`screens/s.bbt:3:7: item.key: must be a string`}},
		{[]string{`do = "logoff"`, `[item.key]`}, []string{`screens/s.bbt:4:2: item.key: must be a string`}},
		{[]string{`key = "esc"`, `do = "Goto main"`}, []string{`screens/s.bbt:4:6: item.do: unknown action "Goto"`}},
		{[]string{`key = "A"`, `do = ""`}, []string{`screens/s.bbt:4:6: item.do: unknown action ""`}},
		{[]string{`key = "A"`, `do = "goto"`}, []string{`screens/s.bbt:4:6: item.do: goto needs a screen name`}},
		{[]string{`key = "A"`, `do = "goto  main"`}, []string{`screens/s.bbt:4:6: item.do: invalid screen name " main"`}},
		{[]string{`key = "A"`, `do = "goto Main"`}, []string{`screens/s.bbt:4:6: item.do: invalid screen name "Main"`}},
		{[]string{`key = "A"`, `do = "goto _f"`}, []string{`screens/s.bbt:4:6: item.do: invalid screen name "_f"`}},
		{[]string{`key = "A"`, `do = "logoff now"`}, []string{`screens/s.bbt:4:6: item.do: logoff takes no argument`}},
		{[]string{`key = "A"`, `do = "logoff "`}, []string{`screens/s.bbt:4:6: item.do: logoff takes no argument`}},
		{[]string{`key = "a"`, `do = 1`, `when = "item.A"`, `x = 1`}, []string{`screens/s.bbt:4:6: item.do: must be a string`}},
		// Each when error is an other error at 2:1, also a syntax error.
		{[]string{`key = "a"`, `do = "logoff"`, `when = "item.A"`}, []string{
			`screens/s.bbt:2:1: item "A": when: field "item.A" is valid only in {#if}`,
		}},
		{[]string{`key = "a"`, `when = "term.cols >"`}, []string{
			`screens/s.bbt:2:1: item "A": do is missing`,
			`screens/s.bbt:2:1: item "A": when: expected an integer or a string, found end of condition`,
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
		`screens/s.bbt:2:1: item "S": key: duplicate key`,
		`screens/s.bbt:2:1: item "ANY": key: duplicate key`,
	}
	if _, msgs := load(map[string]string{"s.bbt": data}); !slices.Equal(msgs, want) {
		t.Errorf("duplicates:\n got %q\nwant %q", msgs, want)
	}
}

func TestIncludes(t *testing.T) {
	screens := mustLoad(t, map[string]string{
		"main.bbt": "+++\n[fields]\nw = \"term.cols\"\n+++\n{!14}a\n{>_a}\nb\n{>_a}",
		"_a.bbt":   "{!title}{w}\n{>_b}\n",
		"_b.bbt":   "{#if term.cols > 1}\nc\n{/if}\n",
	})
	body := screens["main"].Body
	want := `!14/2 underline "a" / screens/_a.bbt(!15/4 bold {term.cols} / screens/_b.bbt(if(/b "c" /)else() /b)) "b" / screens/_a.bbt(!15/4 bold {term.cols} / screens/_b.bbt(if(/b "c" /)else() /b))`
	if got := show(body); got != want {
		t.Errorf("body\n got %s\nwant %s", got, want)
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
		{map[string]string{"s.bbt": "{>_f}", "t.bbt": "{>_f}", "_f.bbt": "{bad}\n{"},
			[]string{`screens/_f.bbt:1:1: screen "s": unknown field "bad"`, `screens/_f.bbt:1:1: screen "t": unknown field "bad"`}},
		{map[string]string{"s.bbt": "+++\n[fields]\nw = \"term.cols\"\n+++\n{>_f}", "t.bbt": "{>_f}", "_f.bbt": "{w}"},
			[]string{`screens/_f.bbt:1:1: screen "t": unknown field "w"`}},
		{map[string]string{"s.bbt": "{>_f}", "_f.bbt": "{#if x}"},
			[]string{`screens/_f.bbt:1:1: {#if} has no {/if}`, `screens/_f.bbt:1:1: screen "s": unknown field "x"`}},
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

	// Art counts its line and its rows.
	_, msgs = load(map[string]string{"s.bbt": strings.Repeat("{>a.ans}\n", 11), "a.ans": strings.Repeat("x\n", 999) + "x"})
	if want := []string{`screens/s.bbt:10:1: body is over 10000 lines after includes`}; !slices.Equal(msgs, want) {
		t.Errorf("art: %q", msgs)
	}

	// A fan-out of empty fragments: each include line counts.
	_, msgs = load(map[string]string{"s.bbt": strings.Repeat("{>_a}\n", 101), "_a.bbt": strings.Repeat("{>_b}\n", 99), "_b.bbt": ""})
	if want := []string{`screens/s.bbt:101:1: body is over 10000 lines after includes`}; !slices.Equal(msgs, want) {
		t.Errorf("empty fragments: %q", msgs)
	}

	// A chain of 100 includes is valid, 101 is over the nesting limit.
	for _, n := range []int{100, 101} {
		files := map[string]string{"s.bbt": "a\n{>_f1}", fmt.Sprintf("_f%d.bbt", n): "x"}
		for i := 1; i < n; i++ {
			files[fmt.Sprintf("_f%d.bbt", i)] = fmt.Sprintf("{>_f%d}", i+1)
		}
		_, msgs = load(files)
		var want []string
		if n == 101 {
			want = []string{`screens/s.bbt:2:1: includes are nested over 100 levels`}
		}
		if !slices.Equal(msgs, want) {
			t.Errorf("chain of %d: %q", n, msgs)
		}
	}

	long := strings.Repeat("x", 600_000)
	_, msgs = load(map[string]string{"s.bbt": "a\n{>_f}\n{>_f}\n", "_f.bbt": long})
	if want := []string{`screens/s.bbt:3:1: body is over 1048576 bytes after includes`}; !slices.Equal(msgs, want) {
		t.Errorf("bytes: %q", msgs)
	}

	// The screens together: 16 screens are at the limit, the 17th passes it.
	for _, tt := range []struct{ frag, want string }{
		{strings.Repeat("\n", bbt.MaxBodyLines-1), `screens/s16.bbt:1:1: screens are over 160000 lines after includes`},
		{strings.Repeat("x", 1_000_000), `screens/s16.bbt:1:1: screens are over 16777216 bytes after includes`},
	} {
		files := map[string]string{"_f.bbt": tt.frag, "t.bbt": "x"}
		for i := range 17 {
			files[fmt.Sprintf("s%02d.bbt", i)] = "{>_f}"
		}
		_, msgs = load(files)
		if want := []string{tt.want}; !slices.Equal(msgs, want) {
			t.Errorf("total: %q", msgs)
		}
	}

	_, msgs = load(map[string]string{"s.bbt": strings.Repeat("\t", 200)})
	if len(msgs) != 101 || msgs[100] != "screens/s.bbt:1:101: too many diagnostics" {
		t.Errorf("diags: %d, last %q", len(msgs), msgs[len(msgs)-1])
	}

	_, msgs = load(map[string]string{"s.bbt": strings.Repeat("x", screen.MaxFileBytes+1)})
	if want := []string{`screens/s.bbt:1:1: file is over 1048576 bytes`}; !slices.Equal(msgs, want) {
		t.Errorf("file: %q", msgs)
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

	// A screen without width takes [screen] width of bbs.toml.
	fsys := fstest.MapFS{
		"screens/a.bbt": {Data: []byte("{>x.ans}\n{>y.ans}")},
		"screens/x.ans": {Data: []byte(sauceArt("x", 50))},
		"screens/y.ans": {Data: []byte(sauceArt("y", 132))},
	}
	screens, ds := bbt.Load(fsys, theme, screen.WidthLimits{Min: 20, Max: 60})
	if w := screens["a"].Width; w != (screen.WidthLimits{Min: 20, Max: 60}) {
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
		{map[string]string{"s.bbt": "{#if term.cols > 100}\n{>x.ans}\n{#else}\n{>y.ans}\n{/if}", "x.ans": sauceArt("x", 132), "y.ans": "y"},
			[]string{`screens/s.bbt:2:1: warning: art "x.ans" is 132 cells wide, max width is 80`}},
		{map[string]string{"about.bbt": "+++\nwidth = [40, 100]\n+++\n{>logo.ans}\n{>_f}", "_f.bbt": "{>logo.ans}", "logo.ans": sauceArt("x", 132)},
			[]string{`screens/_f.bbt:1:1: warning: screen "about": art "logo.ans" is 132 cells wide, max width is 100`, `screens/about.bbt:4:1: warning: art "logo.ans" is 132 cells wide, max width is 100`}},
		{map[string]string{"s.bbt": "{>a.ans}", "a.ans": "{nope}"},
			[]string{`screens/a.ans:1:1: screen "s": unknown field "nope"`}},
		{map[string]string{"s.bbt": "x", "_f.bbt": "{>a.ans}", "a.ans": "x", "b.ans": "\xff"},
			[]string{`screens/_f.bbt:1:1: warning: no screen uses this file`, `screens/a.ans:1:1: warning: no screen uses this file`, `screens/b.ans:1:1: warning: no screen uses this file`}},
		{map[string]string{"s.bbt": "{>a.ans}", "a.ans": "\xff"},
			[]string{`screens/a.ans:1:1: invalid UTF-8: art must be UTF-8, CP437 art needs a conversion`}},
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
			[]string{`screens/goodbye.bbt:2:1: the goodbye screen cannot have items`}},
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
	_, ds := bbt.Load(fsys, theme, defaultWidth)
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

// FuzzLoad loads a screen with a fragment and an art file.
func FuzzLoad(f *testing.F) {
	f.Add([]byte("+++\n[fields]\nw = \"term.cols\"\n+++\n{#if w > 1}{>_f}{/if}\n{>_f}\n{w   }"), []byte("{!title}{w}\n{>a.ans}"))
	f.Add([]byte("{>_f}\n{{{ x}{#else}"), []byte("{>_f}"))
	f.Fuzz(func(t *testing.T, s, frag []byte) {
		fsys := fstest.MapFS{
			"screens/s.bbt":  {Data: s},
			"screens/_f.bbt": {Data: frag},
			"screens/a.ans":  {Data: []byte("{w}")},
		}
		screens, ds := bbt.Load(fsys, theme, defaultWidth)
		if ds.HasErrors() {
			return
		}
		if screens["s"] == nil {
			t.Fatal("no screen without errors")
		}
		checkSegs(t, screens["s"].Body)
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
			if s.Art == nil || len(s.Fields) != len(s.Art.Fields) {
				t.Fatal("art insert without its fields")
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
