package bbt_test

import (
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/schors/velourbbs/internal/screen"
)

// front gives a screen file with the front matter lines and the body body.
func front(body string, lines ...string) string {
	return "+++\n" + strings.Join(lines, "\n") + "\n+++\n" + body
}

// The keys of an effect and their defaults (live-screen, sections 6.1 and 6.4–6.9).
func TestEffects(t *testing.T) {
	screens := mustLoad(t, map[string]string{
		"s.bbt": front("{>a.ans}",
			`[[effect]]`, `do = "twinkle"`, `in = "a.ans"`, `at = [[1, 1], [-1, -2]]`, `chars = "*+"`, `colors = ["14", "title"]`,
			`[[effect]]`, `do = "candle"`, `at = [1, -1, 3, 1]`, `chars = "^"`, `colors = ["#ff8800"]`, `delay = 1`, `every = 0.5`, `spread = 0.1`, `share = 1`,
			`[[effect]]`, `do = "sweep"`, `every = 10`,
			`[[effect]]`, `do = "lightning"`, `at = [1, 1, 10, 5]`, `over = true`,
			`[[effect]]`, `do = "sparkle"`, `chars = "."`, `count = 3`, `time = 1`,
			`[[effect]]`, `do = "glint"`, `peak = "prompt"`),
		"a.ans": sauceArt("abc\nd", 10),
	})
	want := []screen.Effect{
		{Do: screen.Twinkle, In: "a.ans", At: screen.At{Kind: screen.AtPoints, Points: []screen.Point{{Col: 1, Row: 1}, {Col: -1, Row: -2}}},
			Every: time.Second, Spread: time.Second / 2, Chars: []string{"*", "+"}, Colors: []string{"14", "title"}, Share: 0.1},
		{Do: screen.Candle, At: screen.At{Kind: screen.AtRect, Points: []screen.Point{{Col: 1, Row: -1}}, Width: 3, Height: 1},
			Delay: time.Second, Every: time.Second / 2, Spread: 100 * time.Millisecond, Chars: []string{"^"}, Colors: []string{"#ff8800"}, Share: 1},
		{Do: screen.Sweep, Delay: time.Second / 2, Every: 10 * time.Second, Peak: "#ffffff", Time: 600 * time.Millisecond, Width: 3, Speed: 40},
		{Do: screen.Lightning, At: screen.At{Kind: screen.AtRect, Points: []screen.Point{{Col: 1, Row: 1}}, Width: 10, Height: 5},
			Every: 30 * time.Second, Spread: 15 * time.Second, Chars: []string{"╲", "│", "╱"}, Peak: "#ffffff", Over: true, Time: 300 * time.Millisecond},
		{Do: screen.Sparkle, Every: time.Second, Spread: time.Second / 2, Chars: []string{"."}, Count: 3, Time: time.Second},
		{Do: screen.Glint, Every: 3 * time.Second, Spread: 1500 * time.Millisecond, Peak: "prompt", Time: 600 * time.Millisecond},
	}
	if got := screens["s"].Effects; !reflect.DeepEqual(got, want) {
		t.Errorf("effects:\n got %+v\nwant %+v", got, want)
	}
	// A sweep without every has one event.
	screens = mustLoad(t, map[string]string{"s.bbt": front("", `effect = [{ do = "sweep", at = [1, 1, 2, 2] }]`)})
	if e := screens["s"].Effects[0]; e.Every != 0 || e.Spread != 0 {
		t.Errorf("sweep: every %v, spread %v; want 0, 0", e.Every, e.Spread)
	}
}

// Each error of live-screen, section 6.1.
func TestEffectErrors(t *testing.T) {
	tw := func(lines ...string) string {
		return front("", append([]string{`[[effect]]`, `do = "twinkle"`, `chars = "*+"`}, lines...)...)
	}
	tests := []struct {
		files map[string]string
		want  []string
	}{
		{map[string]string{"s.bbt": front("", `[effect]`, `do = "glint"`)},
			[]string{`screens/s.bbt:2:2: effect: must be an array of tables`}},
		{map[string]string{"s.bbt": front("", `effect = 5`)}, []string{`screens/s.bbt:2:10: effect: must be an array of tables`}},
		{map[string]string{"s.bbt": front("", `effect = [1]`)}, []string{`screens/s.bbt:2:11: effect: must be an array of tables`}},
		{map[string]string{"goodbye.bbt": tw(`at = [1, 1]`)},
			[]string{`screens/goodbye.bbt:2:1: effect: the goodbye screen cannot have effects`}},
		{map[string]string{"shutdown.bbt": tw(`at = [1, 1]`)},
			[]string{`screens/shutdown.bbt:2:1: effect: the shutdown screen cannot have effects`}},
		{map[string]string{"s.bbt": front("", slices.Repeat([]string{`[[effect]]`, `do = "glint"`}, 17)...)},
			[]string{`screens/s.bbt:34:1: effect: more than 16 effects`}},
		{map[string]string{"s.bbt": front("", `[[effect]]`, `at = [1, 1]`, `[[effect]]`, `do = "blink"`, `[[effect]]`, `do = 1`)}, []string{
			`screens/s.bbt:2:1: effect 1: do is missing`,
			`screens/s.bbt:5:6: effect 2: do: unknown effect "blink"`,
			`screens/s.bbt:7:6: effect 3: do: must be a string`,
		}},
		{map[string]string{"s.bbt": tw(`at = [1, 1]`, `peak = "1"`, `width = 2`)}, []string{
			`screens/s.bbt:6:8: effect 1: peak: twinkle has no key peak`,
			`screens/s.bbt:7:9: effect 1: width: twinkle has no key width`,
		}},
		// A value outside its range.
		{map[string]string{"s.bbt": tw(`at = [1, 1]`, `delay = 3601`, `every = 0.01`, `share = 0`)}, []string{
			`screens/s.bbt:6:9: effect 1: delay: 3601, must be 0–3600`,
			`screens/s.bbt:7:9: effect 1: every: 0.01, must be 0.05–3600`,
			`screens/s.bbt:8:9: effect 1: share: 0, must be more than 0 and at most 1`,
		}},
		{map[string]string{"s.bbt": front("", `[[effect]]`, `do = "sparkle"`, `chars = "."`, `count = 101`, `time = 61`, `over = 1`)}, []string{
			`screens/s.bbt:5:9: effect 1: count: 101, must be 1–100`,
			`screens/s.bbt:6:8: effect 1: time: 61, must be 0.05–60`,
			`screens/s.bbt:7:8: effect 1: over: must be a boolean`,
		}},
		{map[string]string{"s.bbt": front("", `[[effect]]`, `do = "sweep"`, `width = 21`, `speed = 0.5`, `time = 11`, `delay = -1`)}, []string{
			`screens/s.bbt:4:9: effect 1: width: 21, must be 1–20`,
			`screens/s.bbt:5:9: effect 1: speed: 0.5, must be 1–1000`,
			`screens/s.bbt:6:8: effect 1: time: 11, must be 0.05–10`,
			`screens/s.bbt:7:9: effect 1: delay: -1, must be 0–3600`,
		}},
		// An invalid grapheme string.
		{map[string]string{"s.bbt": front("", `[[effect]]`, `do = "twinkle"`, `at = [1, 1]`, `chars = "**"`)},
			[]string{`screens/s.bbt:5:9: effect 1: chars: must have 2 different graphemes or more`}},
		{map[string]string{"s.bbt": front("", `[[effect]]`, `do = "candle"`, `at = [1, 1]`, `colors = ["1"]`, `chars = "a\tb"`)},
			[]string{`screens/s.bbt:6:9: effect 1: chars: control character U+0009`}},
		{map[string]string{"s.bbt": front("", `[[effect]]`, `do = "candle"`, `at = [1, 1]`, `colors = ["1"]`, `chars = "я語"`)},
			[]string{`screens/s.bbt:6:9: effect 1: chars: grapheme "語" is 2 cells wide, must be 1`}},
		{map[string]string{"s.bbt": front("", `[[effect]]`, `do = "sparkle"`, `chars = "`+strings.Repeat(".", 33)+`"`)},
			[]string{`screens/s.bbt:4:9: effect 1: chars: has 33 graphemes, must have 1–32`}},
		{map[string]string{"s.bbt": front("", `[[effect]]`, `do = "sparkle"`, `chars = ""`)},
			[]string{`screens/s.bbt:4:9: effect 1: chars: has 0 graphemes, must have 1–32`}},
		{map[string]string{"s.bbt": front("", `[[effect]]`, `do = "lightning"`, `chars = "|/"`)},
			[]string{`screens/s.bbt:4:9: effect 1: chars: has 2 graphemes, must have 3`}},
		{map[string]string{"s.bbt": front("", `[[effect]]`, `do = "sparkle"`)},
			[]string{`screens/s.bbt:2:1: effect 1: chars is missing`}},
		// An invalid colour value.
		{map[string]string{"s.bbt": tw(`at = [1, 1]`, `colors = ["head"]`)},
			[]string{`screens/s.bbt:6:10: effect 1: colors: unknown style "head"`}},
		{map[string]string{"s.bbt": tw(`at = [1, 1]`, `colors = ["/1"]`)},
			[]string{`screens/s.bbt:6:10: effect 1: colors: invalid colour value "/1": no foreground`}},
		{map[string]string{"s.bbt": tw(`at = [1, 1]`, `colors = []`)},
			[]string{`screens/s.bbt:6:10: effect 1: colors: has 0 values, must have 1–16`}},
		{map[string]string{"s.bbt": front("", `[[effect]]`, `do = "glint"`, `peak = "#12"`)},
			[]string{`screens/s.bbt:4:8: effect 1: peak: invalid colour value "#12": invalid colour "#12"`}},
		{map[string]string{"s.bbt": front("", `[[effect]]`, `do = "candle"`, `at = [1, 1]`, `chars = "^"`)},
			[]string{`screens/s.bbt:2:1: effect 1: colors is missing`}},
		// spread.
		{map[string]string{"s.bbt": tw(`at = [1, 1]`, `every = 0.5`, `spread = 0.46`)},
			[]string{`screens/s.bbt:7:10: effect 1: spread: every - spread is 0.04, must be at least 0.05`}},
		{map[string]string{"s.bbt": front("", `[[effect]]`, `do = "sweep"`, `spread = 0`)},
			[]string{`screens/s.bbt:4:10: effect 1: spread: a sweep without every has no spread`}},
		// at.
		{map[string]string{"s.bbt": tw()}, []string{`screens/s.bbt:2:1: effect 1: at is missing`}},
		{map[string]string{"s.bbt": tw(`at = [1, 2, 3]`)},
			[]string{`screens/s.bbt:5:6: effect 1: at: must be a point, a rectangle or an array of points`}},
		{map[string]string{"s.bbt": tw(`at = [[1, 1], [2]]`)},
			[]string{`screens/s.bbt:5:6: effect 1: at: must be a point, a rectangle or an array of points`}},
		{map[string]string{"s.bbt": tw(`at = "1, 1"`)},
			[]string{`screens/s.bbt:5:6: effect 1: at: must be a point, a rectangle or an array of points`}},
		{map[string]string{"s.bbt": tw(`at = [` + strings.Repeat(`[1, 1], `, 65) + `]`)},
			[]string{`screens/s.bbt:5:6: effect 1: at: has 65 points, at most 64`}},
		{map[string]string{"s.bbt": tw(`at = [0, 1]`)}, []string{`screens/s.bbt:5:6: effect 1: at: a coordinate cannot be 0`}},
		{map[string]string{"s.bbt": tw(`at = [1, 1.5]`)}, []string{`screens/s.bbt:5:6: effect 1: at: the coordinates must be integers`}},
		{map[string]string{"s.bbt": tw(`at = [1, 1, 0, 1]`)},
			[]string{`screens/s.bbt:5:6: effect 1: at: the width and the height of a rectangle must be at least 1`}},
		// at of a kind that the effect does not take.
		{map[string]string{"s.bbt": front("", `[[effect]]`, `do = "sweep"`, `at = [1, 1]`)},
			[]string{`screens/s.bbt:4:6: effect 1: at: sweep cannot take a point`}},
		{map[string]string{"s.bbt": front("", `[[effect]]`, `do = "lightning"`, `at = [[1, 1]]`)},
			[]string{`screens/s.bbt:4:6: effect 1: at: lightning cannot take an array of points`}},
		{map[string]string{"s.bbt": front("", `[[effect]]`, `do = "sparkle"`, `chars = "."`, `at = [1, 1]`, `count = 2`)},
			[]string{`screens/s.bbt:6:9: effect 1: count: a point or an array of points has no count`}},
		// in: an art of the screen set, and a place in its grid.
		{map[string]string{"s.bbt": tw(`at = [1, 1]`, `in = "a"`)},
			[]string{`screens/s.bbt:6:6: effect 1: in: "a" is not the name of an art insert`}},
		{map[string]string{"s.bbt": tw(`at = [1, 1]`, `in = "a.ans"`) + "{>_f}", "_f.bbt": "{>b.ans}", "a.ans": sauceArt("a", 10), "b.ans": sauceArt("b", 10)},
			[]string{`screens/a.ans:1:1: warning: no screen uses this file`, `screens/s.bbt:6:6: effect 1: in: the screen does not insert art "a.ans"`}},
		{map[string]string{"s.bbt": tw(`at = [[1, 1], [-10, -1]]`, `in = "a.ans"`) + "{>_f}", "_f.bbt": "{#if term.cols > 1}\n{>a.ans}\n{/if}", "a.ans": sauceArt("a", 10)}, nil},
		{map[string]string{"s.bbt": tw(`at = [[1, 1], [-11, 1]]`, `in = "a.ans"`) + "{>a.ans}", "a.ans": sauceArt("a", 10)},
			[]string{`screens/s.bbt:5:6: effect 1: at: a cell is not in the art grid of "a.ans", 10x1`}},
		{map[string]string{"s.bbt": tw(`at = [2, 1, 10, 2]`, `in = "a.ans"`) + "{>a.ans}", "a.ans": sauceArt("a\nb", 10)},
			[]string{`screens/s.bbt:5:6: effect 1: at: a cell is not in the art grid of "a.ans", 10x2`}},
	}
	for _, tt := range tests {
		_, msgs := load(tt.files)
		if !slices.Equal(msgs, tt.want) {
			t.Errorf("%q:\n got %q\nwant %q", tt.files, msgs, tt.want)
		}
	}
}
