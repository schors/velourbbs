package screen

import (
	"fmt"
	"strings"
)

// Layout is the position and the fill of a screen: the keys width, align, valign and fill
// of the front matter and of [screen] (section 3.2 and screens-toml, section 3.1).
// The zero values of Align, VAlign and Fill are the defaults.
type Layout struct {
	Width  WidthLimits
	Align  Align
	VAlign VAlign
	Fill   Fill
}

// LayoutKeys is the keys of a layout that a TOML table sets. A nil pointer is a key that is not set.
type LayoutKeys struct {
	Width  *WidthLimits
	Align  *Align
	VAlign *VAlign
	Fill   *Fill
}

// ParseLayoutKeys parses the TOML values of the keys width, align, valign and fill.
// A nil pointer is a key that is not set. errorf gets the key and the error of each invalid value.
func ParseLayoutKeys(width *[]int64, align, valign, fill *string, errorf func(key string, err error)) LayoutKeys {
	var k LayoutKeys
	if width != nil {
		if w, err := NewWidth(*width); err != nil {
			errorf("width", err)
		} else {
			k.Width = &w
		}
	}
	// The values in the order of section 3.2.
	k.Align = parseKey(align, "align", []Align{AlignLeft, AlignCenter, AlignRight}, errorf)
	k.VAlign = parseKey(valign, "valign", []VAlign{VAlignTop, VAlignMiddle, VAlignBottom}, errorf)
	k.Fill = parseKey(fill, "fill", []Fill{FillNone, FillSpaces, FillErase}, errorf)
	return k
}

// Over gives l with each key that k sets.
func (k LayoutKeys) Over(l Layout) Layout {
	if k.Width != nil {
		l.Width = *k.Width
	}
	if k.Align != nil {
		l.Align = *k.Align
	}
	if k.VAlign != nil {
		l.VAlign = *k.VAlign
	}
	if k.Fill != nil {
		l.Fill = *k.Fill
	}
	return l
}

// Align is the horizontal position of a screen (section 6.1).
type Align uint8

// The values of align.
const (
	AlignCenter Align = iota
	AlignLeft
	AlignRight
)

// VAlign is the vertical position of a screen (section 6.5).
type VAlign uint8

// The values of valign.
const (
	VAlignTop VAlign = iota
	VAlignMiddle
	VAlignBottom
)

// Fill gives the cells of the fill rows (section 6.5).
type Fill uint8

// The values of fill.
const (
	FillNone Fill = iota
	FillSpaces
	FillErase
)

// The names of the values, in the order of the constants.
var (
	alignNames  = []string{"center", "left", "right"}
	valignNames = []string{"top", "middle", "bottom"}
	fillNames   = []string{"none", "spaces", "erase"}
)

func (a Align) String() string  { return alignNames[a] }
func (v VAlign) String() string { return valignNames[v] }
func (f Fill) String() string   { return fillNames[f] }

// parseKey gives the value of values whose name is the value p of key,
// or nil when p is nil or invalid. The error lists the names in the order of values.
func parseKey[T fmt.Stringer](p *string, key string, values []T, errorf func(key string, err error)) *T {
	if p == nil {
		return nil
	}
	for _, v := range values {
		if v.String() == *p {
			return &v
		}
	}
	q := make([]string, len(values))
	for i, v := range values {
		q[i] = fmt.Sprintf("%q", v.String())
	}
	errorf(key, fmt.Errorf("invalid value %q, want %s or %s", *p, strings.Join(q[:len(q)-1], ", "), q[len(q)-1]))
	return nil
}
