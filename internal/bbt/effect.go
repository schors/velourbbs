package bbt

import (
	"fmt"
	"math"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/schors/velourbbs/internal/screen"
	"github.com/schors/velourbbs/internal/tomlcheck"
)

// MaxEffects is the most effects of a screen (live-screen, section 6.1).
const MaxEffects = 16

// effectDoc is the plain decode of one [[effect]] table: the keys of all effects.
// The decode of the document checks their types. A nil pointer is a key that is not set.
type effectDoc struct {
	Do          *string   `toml:"do"`
	In          *string   `toml:"in"`
	At          any       `toml:"at"`
	Delay       *float64  `toml:"delay"`
	DelaySpread *float64  `toml:"delay_spread"`
	Every       *float64  `toml:"every"`
	Spread      *float64  `toml:"spread"`
	Times       *int64    `toml:"times"`
	Chars       *string   `toml:"chars"`
	Colors      *[]string `toml:"colors"`
	Peak        *string   `toml:"peak"`
	Share       *float64  `toml:"share"`
	Over        *bool     `toml:"over"`
	Count       *int64    `toml:"count"`
	Time        *float64  `toml:"time"`
	Width       *int64    `toml:"width"`
	Speed       *float64  `toml:"speed"`
	Shape       *string   `toml:"shape"`
}

// keys gives the keys that the table sets, in the order of effectDoc.
func (d *effectDoc) keys() []string {
	var out []string
	v := reflect.ValueOf(d).Elem()
	for i := range v.NumField() {
		if !v.Field(i).IsNil() {
			out = append(out, v.Type().Field(i).Tag.Get("toml"))
		}
	}
	return out
}

// effectKinds maps a value of do to its effect.
var effectKinds = map[string]screen.EffectKind{
	"twinkle": screen.Twinkle, "sparkle": screen.Sparkle, "candle": screen.Candle,
	"glint": screen.Glint, "sweep": screen.Sweep, "lightning": screen.Lightning,
}

// effectKeys gives the keys of each effect after the keys of all effects: do, in, at, delay,
// delay_spread, every, spread and times (live-screen, sections 6.4–6.9).
var effectKeys = map[screen.EffectKind][]string{
	screen.Twinkle:   {"chars", "colors", "share"},
	screen.Sparkle:   {"chars", "colors", "over", "count", "time"},
	screen.Candle:    {"chars", "colors", "share"},
	screen.Glint:     {"peak", "time"},
	screen.Sweep:     {"peak", "time", "width", "speed"},
	screen.Lightning: {"shape", "chars", "peak", "over", "time"},
}

// commonKeys is the keys of each effect.
var commonKeys = []string{"do", "in", "at", "delay", "delay_spread", "every", "spread", "times"}

// The limits of the keys of live-screen, section 6.1.
const (
	maxPoints = 64
	maxColors = 16
	maxChars  = 32
)

// effectParser parses one [[effect]] table d at the place at.
type effectParser struct {
	d      *effectDoc
	at     tomlcheck.Place
	label  string // "effect <n>", the 1-based number of the effect
	theme  map[string]screen.Style
	errorf errorFunc
	ok     bool // no error so far
}

// parseEffects parses the array of tables effect of the front matter (live-screen, section 6.1).
// root is the place of the root table, theme the base theme.
func parseEffects(docs []effectDoc, root tomlcheck.Place, theme map[string]screen.Style,
	errorf errorFunc) []screen.Effect {
	if len(docs) > MaxEffects {
		errorf(root.At("effect").Elem(MaxEffects).Value(), "effect %d: more than %d effects", MaxEffects+1, MaxEffects)
	}
	var out []screen.Effect
	for i := range docs {
		p := &effectParser{d: &docs[i], at: root.At("effect").Elem(i), label: fmt.Sprintf("effect %d", i+1),
			theme: theme, errorf: errorf, ok: true}
		if e := p.parse(); p.ok {
			out = append(out, e)
		}
	}
	return out
}

// fail reports an error at the value of the key key, or at the table when key is "".
func (p *effectParser) fail(key string, format string, args ...any) {
	p.ok = false
	msg := fmt.Sprintf(format, args...)
	if key == "" {
		p.errorf(p.at.Value(), "%s: %s", p.label, msg)
		return
	}
	p.errorf(p.at.At(key).Value(), "%s: %s: %s", p.label, key, msg)
}

// parse gives the effect of the table. p.ok is false after an error.
func (p *effectParser) parse() (e screen.Effect) {
	d := p.d
	if d.Do == nil {
		p.fail("", "do is missing")
		return e
	}
	e.Do = effectKinds[*d.Do]
	if e.Do == 0 {
		p.fail("do", "unknown effect %q", *d.Do)
		return e
	}
	keys := append(slices.Clone(commonKeys), effectKeys[e.Do]...)
	for _, k := range d.keys() {
		if !slices.Contains(keys, k) { // an error of the name of a key: at the key
			p.ok = false
			p.errorf(p.at.At(k).Key(), "%s: %s: %s has no key %s", p.label, k, *d.Do, k)
		}
	}
	if d.In != nil {
		art, isArt := strings.CutSuffix(*d.In, ".ans")
		fragment, isFragment := strings.CutPrefix(*d.In, "_")
		if valid := isArt && screen.IsName(art) || isFragment && screen.IsName(fragment); !valid {
			p.fail("in", "%q is not the name of an art insert or a fragment", *d.In)
		}
		e.In = *d.In
	}
	p.parseAt(&e)
	p.parseTimes(&e)
	p.parseKeys(&e)
	return e
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

// parseAt parses the key at (live-screen, section 6.2).
func (p *effectParser) parseAt(e *screen.Effect) {
	kinds, required := takes(e.Do)
	if p.d.At == nil {
		if required {
			p.fail("", "at is missing")
		}
		return
	}
	at, err := parseAtValue(p.d.At)
	if err != "" {
		p.fail("at", "%s", err)
		return
	}
	if !slices.Contains(kinds, at.Kind) {
		p.fail("at", "%s cannot take %s", *p.d.Do, atKindNames[at.Kind])
		return
	}
	e.At = at
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

// maxCoord bounds the coordinates and the sizes of a place, so that their sums do not overflow.
// It is far larger than each grid and art, so the cut of a place to the grid stays the same.
const maxCoord = 1 << 60

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
	maxTimes = 10000
)

// The default every of each effect, in seconds (live-screen, sections 6.4–6.9). A sweep has none.
var defaultEvery = map[screen.EffectKind]float64{
	screen.Twinkle: 1, screen.Sparkle: 1, screen.Candle: 0.2, screen.Glint: 3, screen.Lightning: 30,
}

// seconds gives the duration of s seconds.
func seconds(s float64) time.Duration { return time.Duration(math.Round(s * float64(time.Second))) }

// difference gives a - b in seconds, exact to the nanosecond as the durations of the times.
// b can be too large for a time.Duration: then it gives the float difference.
func difference(a, b float64) float64 {
	if b > maxEvery {
		return a - b
	}
	return (seconds(a) - seconds(b)).Seconds()
}

// parseTimes parses the keys delay, delay_spread, every, spread and times (live-screen, sections 6.1 and 6.3).
func (p *effectParser) parseTimes(e *screen.Effect) {
	d := p.d
	delay := 0.0
	if e.Do == screen.Sweep {
		delay = 0.5
	}
	delayOK := p.number("delay", d.Delay, &delay, 0, maxDelay)
	e.Delay = seconds(delay)
	if d.DelaySpread != nil {
		var ds float64
		if p.number("delay_spread", d.DelaySpread, &ds, 0, math.Inf(1)) {
			if delayOK && ds > delay {
				p.fail("delay_spread", "delay - delay_spread is %g, must be at least 0", difference(delay, ds))
			}
			e.DelaySpread = seconds(ds)
		}
	}
	every, hasEvery := defaultEvery[e.Do]
	everyOK := true
	if d.Every != nil {
		hasEvery = true
		everyOK = p.number("every", d.Every, &every, minEvery, maxEvery)
	}
	if hasEvery {
		e.Every = seconds(every)
	} else {
		e.Times = 1 // a sweep without every has one event
	}
	if d.Times != nil {
		if !hasEvery {
			p.fail("times", "a sweep without every has one event")
		} else {
			p.integer("times", d.Times, &e.Times, 1, maxTimes)
		}
	}
	spread := every / 2
	if e.Do == screen.Sweep {
		spread = 0
	}
	if d.Spread != nil {
		if !p.number("spread", d.Spread, &spread, 0, math.Inf(1)) {
			return
		}
		switch {
		case !hasEvery:
			p.fail("spread", "a sweep without every has no spread")
			return
		case everyOK && difference(every, spread) < minEvery:
			p.fail("spread", "every - spread is %g, must be at least %g", difference(every, spread), minEvery)
			return
		}
	}
	e.Spread = seconds(spread)
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
func (p *effectParser) parseKeys(e *screen.Effect) {
	d := p.d
	switch e.Do {
	case screen.Twinkle, screen.Sparkle, screen.Candle:
		p.chars(e, "")
		p.colors(e, e.Do == screen.Candle)
	case screen.Lightning:
		p.shape(e)
		switch {
		case e.Shape != screen.ShapeSquare:
			p.chars(e, "╲│╱")
		case d.Chars != nil:
			p.fail("chars", "a square bolt has no chars")
		}
	}
	switch e.Do {
	case screen.Twinkle, screen.Candle:
		e.Share = 0.1
		if e.Do == screen.Candle {
			e.Share = 0.3
		}
		if s := d.Share; s != nil {
			if math.IsNaN(*s) || *s <= 0 || *s > 1 {
				p.fail("share", "%g, must be more than 0 and at most 1", *s)
			} else {
				e.Share = *s
			}
		}
	}
	if d.Over != nil {
		e.Over = *d.Over
	}
	if e.Do == screen.Sparkle {
		e.Count = 1
		if d.Count != nil && (e.At.Kind == screen.AtPoint || e.At.Kind == screen.AtPoints) {
			p.fail("count", "a point or an array of points has no count")
		}
		p.integer("count", d.Count, &e.Count, 1, maxCount)
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
		p.number("time", d.Time, &t, minTime, maxT)
		e.Time = seconds(t)
	}
	switch e.Do {
	case screen.Glint, screen.Sweep, screen.Lightning:
		e.Peak = "#ffffff"
		if d.Peak != nil {
			if msg := p.colourValue(*d.Peak); msg != "" {
				p.fail("peak", "%s", msg)
			}
			e.Peak = *d.Peak
		}
	}
	if e.Do == screen.Sweep {
		e.Width, e.Speed = 3, 80
		p.integer("width", d.Width, &e.Width, 1, maxWidth)
		p.number("speed", d.Speed, &e.Speed, minSpeed, maxSpeed)
	}
}

// shapes maps a value of the key shape of a lightning to its shape.
var shapes = map[string]screen.Shape{"both": screen.ShapeBoth, "slant": screen.ShapeSlant, "square": screen.ShapeSquare}

// shape parses the key shape of a lightning (live-screen, section 6.9). The default is both.
func (p *effectParser) shape(e *screen.Effect) {
	if p.d.Shape == nil {
		return
	}
	s, ok := shapes[*p.d.Shape]
	if !ok {
		p.fail("shape", "%q, must be slant, square or both", *p.d.Shape)
	}
	e.Shape = s
}

// chars parses the key chars: a grapheme string (live-screen, section 6.1). def is the default;
// without a default, chars is required. A twinkle needs 2 different graphemes, a lightning 3 graphemes.
func (p *effectParser) chars(e *screen.Effect, def string) {
	v := def
	switch {
	case p.d.Chars != nil:
		v = *p.d.Chars
	case def == "":
		p.fail("", "chars is missing")
		return
	}
	var gs []string
	for g, w := range screen.Graphemes(v) {
		if err := screen.CheckGrapheme(g, w); err != nil {
			p.fail("chars", "%v", err)
			return
		}
		if w != 1 {
			p.fail("chars", "grapheme %q is %d cells wide, must be 1", g, w)
			return
		}
		gs = append(gs, g)
	}
	switch {
	case len(gs) == 0 || len(gs) > maxChars:
		p.fail("chars", "has %d graphemes, must have 1–%d", len(gs), maxChars)
	case e.Do == screen.Lightning && len(gs) != 3:
		p.fail("chars", "has %d graphemes, must have 3", len(gs))
	case e.Do == screen.Twinkle && len(slices.Compact(slices.Sorted(slices.Values(gs)))) < 2:
		p.fail("chars", "must have 2 different graphemes or more")
	default:
		e.Chars = gs
	}
}

// colors parses the key colors: 1–16 colour values. required is true for a candle.
func (p *effectParser) colors(e *screen.Effect, required bool) {
	if p.d.Colors == nil {
		if required {
			p.fail("", "colors is missing")
		}
		return
	}
	a := *p.d.Colors
	if len(a) == 0 || len(a) > maxColors {
		p.fail("colors", "has %d values, must have 1–%d", len(a), maxColors)
		return
	}
	for _, s := range a {
		if msg := p.colourValue(s); msg != "" {
			p.fail("colors", "%s", msg)
			return
		}
	}
	e.Colors = a
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

// number sets *v to the value n of the key k when it is set and in [lo, hi].
// It reports false after an error.
func (p *effectParser) number(k string, n *float64, v *float64, lo, hi float64) bool {
	if n == nil {
		return true
	}
	if math.IsNaN(*n) || *n < lo || *n > hi {
		if math.IsInf(hi, 1) {
			p.fail(k, "%g, must be at least %g", *n, lo)
		} else {
			p.fail(k, "%g, must be %g–%g", *n, lo, hi)
		}
		return false
	}
	*v = *n
	return true
}

// integer sets *v to the value n of the key k when it is set and in [lo, hi].
func (p *effectParser) integer(k string, n *int64, v *int, lo, hi int) {
	if n == nil {
		return
	}
	if *n < int64(lo) || *n > int64(hi) {
		p.fail(k, "%d, must be %d–%d", *n, lo, hi)
		return
	}
	*v = int(*n)
}
