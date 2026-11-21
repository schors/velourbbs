package bbt_test

import (
	"slices"
	"testing"

	"github.com/schors/velourbbs/internal/screen"
)

// The aliases of a front matter override the aliases of screens/screens.toml in the screen,
// its when, its fragments and its art (section 4.2). A fragment or an art takes the aliases
// of each screen that uses it.
func TestFrontFields(t *testing.T) {
	screens := mustLoad(t, map[string]string{
		"s.bbt": "+++\n[fields]\nw = \"term.rows\"\nx = 'bbs.time live \"15:04\"'\n" +
			"[[item]]\nkey = \"A\"\ndo = \"logoff\"\nwhen = \"w > 10\"\n+++\n{w}{x}\n{>_f}\n{>a.ans}",
		"t.bbt":  "{w}\n{>_f}\n{>a.ans}",
		"_f.bbt": "{w}\n{#if w > 10}{/if}\n{>_g}\n{>a.ans}",
		"_g.bbt": "{w      }",
		"a.ans":  sauceArt("{w} {#if w > 10}{/if}", 80),
	})
	tests := []struct{ name, field, want string }{
		{"s", "term.rows", `{term.rows} {bbs.time live "15:04"} / screens/_f.bbt({term.rows} / if()else() /b screens/_g.bbt({term.rows:9} /) art80[term.rows]) art80[term.rows]`},
		{"t", "term.cols", `{term.cols} / screens/_f.bbt({term.cols} / if()else() /b screens/_g.bbt({term.cols:9} /) art80[term.cols]) art80[term.cols]`},
	}
	for _, tt := range tests {
		body := screens[tt.name].Body
		if got := show(body); got != tt.want {
			t.Errorf("%s:\n got %s\nwant %s", tt.name, got, tt.want)
		}
		// The condition of the fragment and the box of the art.
		i := slices.IndexFunc(body, func(s screen.Seg) bool { _, ok := s.(screen.Include); return ok })
		cond := body[i].(screen.Include).Body[2].(screen.If).Cond
		box := body[len(body)-1].(screen.ArtInsert).Art.Boxes[0].Cond
		for _, c := range []*screen.Cond{cond, box} {
			if got := c.Refs()[0].String(); got != tt.field {
				t.Errorf("%s: condition field %s, want %s", tt.name, got, tt.field)
			}
		}
	}
	if got := screens["s"].Items[0].When.Refs()[0].String(); got != "term.rows" {
		t.Errorf("when: field %s, want term.rows", got)
	}
}

func TestFrontFieldErrors(t *testing.T) {
	tests := []struct {
		files map[string]string
		want  []string
	}{
		// The errors of an alias have the positions of screens/screens.toml (screen-render, section 3.1).
		{map[string]string{"s.bbt": "+++\nfields = { \"яя\" = \"term.cols\", and = \"term.rows\", k = \"item.A\", u = \"term.x\", v = \"w\" }\n+++\n"}, []string{
			`screens/s.bbt:2:12: fields: invalid alias "яя"`,
			`screens/s.bbt:2:32: fields: alias "and" is a reserved word`,
			`screens/s.bbt:2:55: fields: alias "k": "item.A" is valid only in {#if}`,
			`screens/s.bbt:2:69: fields: alias "u": unknown field "term.x"`,
			`screens/s.bbt:2:83: fields: alias "v": "w" is not a registry path`,
		}},
		{map[string]string{"s.bbt": "+++\n[fields]\nc = 'term.cols live \"15:04\"'\nw = 1\n+++\n"},
			[]string{`screens/s.bbt:4:5: fields.w: must be a string`}},
		// After an error of an alias, the body and the fragments use the valid aliases.
		{map[string]string{"s.bbt": "+++\n[fields]\nc = 'term.cols live \"15:04\"'\nx = \"term.type\"\n+++\n{c}{x}\n{>_f}", "_f.bbt": "{x}"}, []string{
			`screens/s.bbt:3:5: fields: alias "c": layout "15:04" on "term.cols": only bbs.time takes a layout`,
			`screens/s.bbt:6:1: unknown field "c"`,
		}},
		// An error of the aliases of the screen in its body.
		{map[string]string{"s.bbt": "+++\n[fields]\nc = 'bbs.time live'\n+++\n{c live}\n{#if c > 1}{/if}\n"}, []string{
			`screens/s.bbt:5:1: the live mark on the alias "c": an alias takes no mark`,
			`screens/s.bbt:6:1: alias "c" has the live mark: a condition takes the values of the show`,
		}},
		// A fragment and an art that two screens use: an error of each resolution once,
		// an error of the aliases of one screen with its name.
		{map[string]string{
			"s.bbt":  "+++\n[fields]\nx = \"term.type\"\n+++\n{>_f}\n{>a.ans}",
			"t.bbt":  "{>_f}\n{>a.ans}",
			"u.bbt":  "{>_f}",
			"_f.bbt": "{x} {bad}\n{#if x == \"a\"}{/if}",
			"a.ans":  sauceArt("{x} {bad}", 80),
		}, []string{
			`screens/_f.bbt:1:1: screen "t": unknown field "x"`,
			`screens/_f.bbt:1:1: screen "u": unknown field "x"`,
			`screens/_f.bbt:1:5: unknown field "bad"`,
			`screens/_f.bbt:2:1: screen "t": unknown field "x"`,
			`screens/_f.bbt:2:1: screen "u": unknown field "x"`,
			`screens/a.ans:1:1: screen "t": unknown field "x"`,
			`screens/a.ans:1:5: unknown field "bad"`,
		}},
		// After an unknown key, the body uses the aliases of [fields] without their errors.
		{map[string]string{"s.bbt": "+++\ntypo = 1\n[fields]\nx = \"term.cols\"\nand = \"term.rows\"\n+++\n{x}\n{>_f}", "_f.bbt": "{x}"},
			[]string{`screens/s.bbt:2:1: unknown key "typo"`}},
		// A screen with a front matter error reports the errors of its aliases in a fragment and an art.
		{map[string]string{
			"a.bbt":  "+++\nwidth = [1, 0]\n[fields]\nx = 'bbs.time live'\n+++\n{>_f}\n{>a.ans}",
			"b.bbt":  "+++\n[fields]\nx = \"term.type\"\n+++\n{>_f}\n{>a.ans}",
			"_f.bbt": "{#if x == \"a\"}{/if}",
			"a.ans":  sauceArt("{#if x == \"a\"}{/if}", 80),
		}, []string{
			`screens/_f.bbt:1:1: screen "a": alias "x" has the live mark: a condition takes the values of the show`,
			`screens/a.ans:1:1: screen "a": alias "x" has the live mark: a condition takes the values of the show`,
			`screens/a.bbt:2:9: width: width [1, 0] is not in 1 <= min <= max <= 255`,
		}},
		// A fragment that only screens with the alias use has no error.
		{map[string]string{
			"s.bbt":  "+++\n[fields]\nx = \"term.type\"\n+++\n{>_f}",
			"t.bbt":  "+++\n[fields]\nx = \"term.name\"\n+++\n{>_f}",
			"_f.bbt": "{x}",
		}, nil},
	}
	for _, tt := range tests {
		_, msgs := load(tt.files)
		if !slices.Equal(msgs, tt.want) {
			t.Errorf("%q:\n got %q\nwant %q", tt.files, msgs, tt.want)
		}
	}
}
