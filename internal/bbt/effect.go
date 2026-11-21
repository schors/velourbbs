package bbt

import (
	"fmt"
	"maps"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/schors/velourbbs/internal/screen"
	"github.com/schors/velourbbs/internal/tomlcheck"
)

// MaxEffects is the most effects of a screen (live-screen, section 6.1).
const MaxEffects = 16

// effectKinds maps a value of do to its effect.
var effectKinds = map[string]screen.EffectKind{
	"twinkle": screen.Twinkle, "sparkle": screen.Sparkle, "candle": screen.Candle,
	"glint": screen.Glint, "sweep": screen.Sweep, "lightning": screen.Lightning,
}

// effectKeys gives the keys of each effect after the keys of all effects: do, in, at, delay, every
// and spread (live-screen, sections 6.4–6.9).
var effectKeys = map[screen.EffectKind][]string{
	screen.Twinkle:   {"chars", "colors", "share"},
	screen.Sparkle:   {"chars", "colors", "over", "count", "time"},
	screen.Candle:    {"chars", "colors", "share"},
	screen.Glint:     {"peak", "time"},
	screen.Sweep:     {"peak", "time", "width", "speed"},
	screen.Lightning: {"chars", "peak", "over", "time"},
}

// commonKeys is the keys of each effect.
var commonKeys = []string{"do", "in", "at", "delay", "every", "spread"}

// The limits of the keys of live-screen, section 6.1.
const (
	maxPoints = 64
	maxColors = 16
	maxChars  = 32
)

// effectParser parses one [[effect]] table: the 1-based effect n at the place at.
type effectParser struct {
	t      map[string]any
	at     tomlcheck.Place
	label  string
	theme  map[string]screen.Style
	errorf errorFunc
}

// parseEffects parses the array of tables effect of the front matter (live-screen, section 6.1).
// v is its plain decode, root the place of the root table, theme the base theme.
func parseEffects(v []map[string]any, root tomlcheck.Place, theme map[string]screen.Style,
	errorf errorFunc) []screen.Effect {
	if len(v) > MaxEffects {
		errorf(root.At("effect").Elem(MaxEffects).Value(), "effect: more than %d effects", MaxEffects)
	}
	var out []screen.Effect
	for i, t := range v {
		p := &effectParser{t: t, at: root.At("effect").Elem(i), label: fmt.Sprintf("effect %d", i+1),
			theme: theme, errorf: errorf}
		if e, ok := p.parse(); ok {
			out = append(out, e)
		}
	}
	return out
}

// fail reports an error of the key key, or of the table when key is "".
func (p *effectParser) fail(key string, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if key == "" {
		p.errorf(p.at.Value(), "%s: %s", p.label, msg)
		return
	}
	p.errorf(p.at.At(key).Value(), "%s: %s: %s", p.label, key, msg)
}

// parse gives the effect of the table. ok is false after an error.
func (p *effectParser) parse() (e screen.Effect, ok bool) {
	ok = true
	fail := func(key, format string, args ...any) {
		p.fail(key, format, args...)
		ok = false
	}
	do, has, valid := p.str("do")
	if !has {
		fail("", "do is missing")
		return e, false
	}
	if !valid {
		return e, false
	}
	e.Do = effectKinds[do]
	if e.Do == 0 {
		fail("do", "unknown effect %q", do)
		return e, false
	}
	keys := append(slices.Clone(commonKeys), effectKeys[e.Do]...)
	for _, k := range slices.Sorted(maps.Keys(p.t)) {
		if !slices.Contains(keys, k) {
			fail(k, "%s has no key %s", do, k)
		}
	}
	if in, has, valid := p.str("in"); valid {
		if !screen.IsName(strings.TrimSuffix(in, ".ans")) || !strings.HasSuffix(in, ".ans") {
			fail("in", "%q is not the name of an art insert", in)
		}
		e.In = in
	} else if has {
		ok = false
	}
	if !p.parseAt(&e) {
		ok = false
	}
	if !p.parseTimes(&e) {
		ok = false
	}
	if !p.parseKeys(&e) {
		ok = false
	}
	return e, ok
}

// takes gives the kinds of at of each effect and whether it needs at (live-screen, sections 6.4–6.9).
func takes(k screen.EffectKind) (kinds []screen.AtKind, required bool) {
	switch k {
	case screen.Twinkle, screen.Candle:
		return []screen.AtKind{screen.AtPoint, screen.AtRect, screen.AtPoints}, true
	case screen.Sweep, screen.Lightning:
		return []screen.AtKind{screen.AtAll, screen.AtRect}, false
	}
	return []screen.AtKind{screen.AtAll, screen.AtPoint, screen.AtRect, screen.AtPoints}, false
}

// atKindNames is the name of each kind of at for the errors.
var atKindNames = map[screen.AtKind]string{
	screen.AtPoint: "a point", screen.AtRect: "a rectangle", screen.AtPoints: "an array of points",
}

// parseAt parses the key at (live-screen, section 6.2). It reports false after an error.
func (p *effectParser) parseAt(e *screen.Effect) bool {
	kinds, required := takes(e.Do)
	v, has := p.t["at"]
	if !has {
		if required {
			p.fail("", "at is missing")
			return false
		}
		return true
	}
	at, err := parseAtValue(v)
	if err != "" {
		p.fail("at", "%s", err)
		return false
	}
	if !slices.Contains(kinds, at.Kind) {
		p.fail("at", "%s cannot take %s", effectName(e.Do), atKindNames[at.Kind])
		return false
	}
	e.At = at
	return true
}

// effectName gives the value of do of the effect k.
func effectName(k screen.EffectKind) string {
	for name, v := range effectKinds {
		if v == k {
			return name
		}
	}
	return ""
}

// parseAtValue parses the value v of at. It gives a message after an error.
func parseAtValue(v any) (screen.At, string) {
	const kinds = "must be a point, a rectangle or an array of points"
	a, ok := v.([]any)
	if !ok || len(a) == 0 {
		return screen.At{}, kinds
	}
	if _, nested := a[0].([]any); nested {
		if len(a) > maxPoints {
			return screen.At{}, fmt.Sprintf("has %d points, at most %d", len(a), maxPoints)
		}
		at := screen.At{Kind: screen.AtPoints}
		for _, el := range a {
			el, ok := el.([]any)
			if !ok || len(el) != 2 {
				return screen.At{}, kinds
			}
			c, msg := coords(el)
			if msg != "" {
				return screen.At{}, msg
			}
			at.Points = append(at.Points, screen.Point{Col: c[0], Row: c[1]})
		}
		return at, ""
	}
	if len(a) != 2 && len(a) != 4 {
		return screen.At{}, kinds
	}
	c, msg := coords(a)
	if msg != "" {
		return screen.At{}, msg
	}
	if len(c) == 2 {
		return screen.At{Kind: screen.AtPoint, Points: []screen.Point{{Col: c[0], Row: c[1]}}}, ""
	}
	if c[2] < 1 || c[3] < 1 {
		return screen.At{}, "the width and the height of a rectangle must be at least 1"
	}
	return screen.At{Kind: screen.AtRect, Points: []screen.Point{{Col: c[0], Row: c[1]}}, Width: c[2], Height: c[3]}, ""
}

// maxCoord is the largest coordinate and size of a place: no grid or art is larger.
const maxCoord = 1 << 16

// coords gives the integers of a: the coordinates of a point and the size of a rectangle.
func coords(a []any) ([]int, string) {
	out := make([]int, len(a))
	for i, el := range a {
		n, ok := el.(int64)
		switch {
		case !ok:
			return nil, "the coordinates must be integers"
		case i < 2 && n == 0:
			return nil, "a coordinate cannot be 0"
		}
		out[i] = int(max(min(n, maxCoord), -maxCoord))
	}
	return out, ""
}

// The ranges of the times, in seconds (live-screen, section 6.1).
const (
	maxDelay = 3600
	minEvery = 0.05
	maxEvery = 3600
)

// The default every of each effect, in seconds (live-screen, sections 6.4–6.9). A sweep has none.
var defaultEvery = map[screen.EffectKind]float64{
	screen.Twinkle: 1, screen.Sparkle: 1, screen.Candle: 0.2, screen.Glint: 3, screen.Lightning: 30,
}

// seconds gives the duration of s seconds.
func seconds(s float64) time.Duration { return time.Duration(math.Round(s * float64(time.Second))) }

// parseTimes parses the keys delay, every and spread (live-screen, sections 6.1 and 6.3).
// It reports false after an error.
func (p *effectParser) parseTimes(e *screen.Effect) bool {
	ok := true
	delay := 0.0
	if e.Do == screen.Sweep {
		delay = 0.5
	}
	if !p.number("delay", &delay, 0, maxDelay) {
		ok = false
	}
	e.Delay = seconds(delay)
	every, hasEvery := defaultEvery[e.Do]
	if _, set := p.t["every"]; set {
		hasEvery = true
	}
	if !p.number("every", &every, minEvery, maxEvery) {
		ok = false
	}
	if hasEvery {
		e.Every = seconds(every)
	}
	spread := every / 2
	if e.Do == screen.Sweep {
		spread = 0
	}
	if _, set := p.t["spread"]; set {
		if !p.number("spread", &spread, 0, math.Inf(1)) {
			return false
		}
		switch {
		case !hasEvery:
			p.fail("spread", "a sweep without every has no spread")
			return false
		case ok && seconds(every)-seconds(spread) < seconds(minEvery):
			p.fail("spread", "every - spread is %g, must be at least %g", (seconds(every) - seconds(spread)).Seconds(), minEvery)
			return false
		}
	}
	e.Spread = seconds(spread)
	return ok
}

// The ranges of the other keys (live-screen, sections 6.4–6.9).
const (
	maxCount     = 100
	minTime      = 0.05
	maxTime      = 10
	maxSparkTime = 60
	maxWidth     = 20
	minSpeed     = 1
	maxSpeed     = 1000
)

// parseKeys parses the keys of each effect (live-screen, sections 6.4–6.9).
// It reports false after an error.
func (p *effectParser) parseKeys(e *screen.Effect) bool {
	ok := true
	check := func(valid bool) {
		if !valid {
			ok = false
		}
	}
	switch e.Do {
	case screen.Twinkle, screen.Sparkle, screen.Candle:
		need := 1
		if e.Do == screen.Twinkle {
			need = 2
		}
		check(p.chars(e, "", need))
		check(p.colors(e, e.Do == screen.Candle))
	case screen.Lightning:
		check(p.chars(e, "╲│╱", 3))
	}
	switch e.Do {
	case screen.Twinkle, screen.Candle:
		e.Share = 0.1
		if e.Do == screen.Candle {
			e.Share = 0.3
		}
		if v, has := p.t["share"]; has {
			if s, isNum := asNumber(v); !isNum {
				p.fail("share", "must be a number")
				ok = false
			} else if s <= 0 || s > 1 {
				p.fail("share", "%g, must be more than 0 and at most 1", s)
				ok = false
			} else {
				e.Share = s
			}
		}
	}
	switch e.Do {
	case screen.Sparkle, screen.Lightning:
		check(p.boolean("over", &e.Over))
	}
	if e.Do == screen.Sparkle {
		e.Count = 1
		if _, has := p.t["count"]; has && (e.At.Kind == screen.AtPoint || e.At.Kind == screen.AtPoints) {
			p.fail("count", "a point or an array of points has no count")
			ok = false
		}
		check(p.integer("count", &e.Count, 1, maxCount))
	}
	t, maxT := 0.0, float64(maxTime)
	switch e.Do {
	case screen.Sparkle:
		t, maxT = 2, maxSparkTime
	case screen.Glint, screen.Sweep:
		t = 0.6
	case screen.Lightning:
		t = 0.3
	}
	if t > 0 {
		check(p.number("time", &t, minTime, maxT))
		e.Time = seconds(t)
	}
	switch e.Do {
	case screen.Glint, screen.Sweep, screen.Lightning:
		e.Peak = "#ffffff"
		if v, has, valid := p.str("peak"); valid {
			if msg := p.colourValue(v); msg != "" {
				p.fail("peak", "%s", msg)
				ok = false
			}
			e.Peak = v
		} else if has {
			ok = false
		}
	}
	if e.Do == screen.Sweep {
		e.Width, e.Speed = 3, 40
		check(p.integer("width", &e.Width, 1, maxWidth))
		check(p.number("speed", &e.Speed, minSpeed, maxSpeed))
	}
	return ok
}

// chars parses the key chars: a grapheme string with at least need graphemes, or exactly 3 for a lightning.
// def is the default; without a default, chars is required. A twinkle needs 2 different graphemes.
func (p *effectParser) chars(e *screen.Effect, def string, need int) bool {
	v, has, valid := p.str("chars")
	switch {
	case !has && def == "":
		p.fail("", "chars is missing")
		return false
	case !has:
		v = def
	case !valid:
		return false
	}
	var gs []string
	for g, w := range screen.Graphemes(v) {
		if err := screen.CheckGrapheme(g, w); err != nil {
			p.fail("chars", "%v", err)
			return false
		}
		if w != 1 {
			p.fail("chars", "grapheme %q is %d cells wide, must be 1", g, w)
			return false
		}
		gs = append(gs, g)
	}
	switch {
	case len(gs) == 0 || len(gs) > maxChars:
		p.fail("chars", "has %d graphemes, must have 1–%d", len(gs), maxChars)
		return false
	case e.Do == screen.Lightning && len(gs) != 3:
		p.fail("chars", "has %d graphemes, must have 3", len(gs))
		return false
	case e.Do == screen.Twinkle && len(slices.Compact(slices.Sorted(slices.Values(gs)))) < need:
		p.fail("chars", "must have %d different graphemes or more", need)
		return false
	}
	e.Chars = gs
	return true
}

// colors parses the key colors: 1–16 colour values. required is true for a candle.
func (p *effectParser) colors(e *screen.Effect, required bool) bool {
	v, has := p.t["colors"]
	if !has {
		if required {
			p.fail("", "colors is missing")
			return false
		}
		return true
	}
	a, ok := v.([]any)
	if !ok {
		p.fail("colors", "must be an array of strings")
		return false
	}
	if len(a) == 0 || len(a) > maxColors {
		p.fail("colors", "has %d values, must have 1–%d", len(a), maxColors)
		return false
	}
	for _, el := range a {
		s, ok := el.(string)
		if !ok {
			p.fail("colors", "must be an array of strings")
			return false
		}
		if msg := p.colourValue(s); msg != "" {
			p.fail("colors", "%s", msg)
			return false
		}
		e.Colors = append(e.Colors, s)
	}
	return true
}

// colourValue checks the colour value v: a style with a foreground or the name of a theme style
// (live-screen, section 6.1). It gives a message after an error.
func (p *effectParser) colourValue(v string) string {
	if screen.IsIdent(v) {
		if _, ok := p.theme[v]; !ok {
			return fmt.Sprintf("unknown style %q", v)
		}
		return ""
	}
	s, err := screen.ParseStyle(v)
	switch {
	case err != nil:
		return fmt.Sprintf("invalid colour value %q: %v", v, err)
	case !s.HasFg:
		return fmt.Sprintf("invalid colour value %q: no foreground", v)
	}
	return ""
}

// str gives the string of the key k. has is false when the table does not set k;
// valid is false when k is not a string: str reports it.
func (p *effectParser) str(k string) (v string, has, valid bool) {
	raw, has := p.t[k]
	if !has {
		return "", false, false
	}
	v, valid = raw.(string)
	if !valid {
		p.fail(k, "must be a string")
	}
	return v, true, valid
}

// asNumber gives the value of a TOML integer or float.
func asNumber(v any) (float64, bool) {
	switch v := v.(type) {
	case int64:
		return float64(v), true
	case float64:
		return v, true
	}
	return 0, false
}

// number sets *v to the number of the key k in [lo, hi] when the table sets k.
// It reports false after an error.
func (p *effectParser) number(k string, v *float64, lo, hi float64) bool {
	raw, has := p.t[k]
	if !has {
		return true
	}
	n, ok := asNumber(raw)
	switch {
	case !ok || math.IsNaN(n):
		p.fail(k, "must be a number")
		return false
	case n < lo || n > hi:
		if math.IsInf(hi, 1) {
			p.fail(k, "%g, must be at least %g", n, lo)
		} else {
			p.fail(k, "%g, must be %g–%g", n, lo, hi)
		}
		return false
	}
	*v = n
	return true
}

// integer sets *v to the integer of the key k in [lo, hi] when the table sets k.
func (p *effectParser) integer(k string, v *int, lo, hi int) bool {
	raw, has := p.t[k]
	if !has {
		return true
	}
	n, ok := raw.(int64)
	switch {
	case !ok:
		p.fail(k, "must be an integer")
		return false
	case n < int64(lo) || n > int64(hi):
		p.fail(k, "%d, must be %d–%d", n, lo, hi)
		return false
	}
	*v = int(n)
	return true
}

// boolean sets *v to the boolean of the key k when the table sets k.
func (p *effectParser) boolean(k string, v *bool) bool {
	raw, has := p.t[k]
	if !has {
		return true
	}
	b, ok := raw.(bool)
	if !ok {
		p.fail(k, "must be a boolean")
		return false
	}
	*v = b
	return true
}
