package screen

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// StyleSpec is a parsed style of section 7.2. A colour that the text does not name
// has its Has flag false.
type StyleSpec struct {
	Fg, Bg       Colour
	HasFg, HasBg bool
	Attr         Attr
}

// Over returns base with the colours that s names and with the attributes of s added.
// A style token applies to the current style. A theme style applies to the colours
// of normal without attributes.
func (s StyleSpec) Over(base Style) Style {
	if s.HasFg {
		base.Fg = s.Fg
	}
	if s.HasBg {
		base.Bg = s.Bg
	}
	base.Attr |= s.Attr
	return base
}

// ParseStyle parses the style syntax of section 7.2.
// Examples: "14", "14/1", "/1", "#ff8800/#000020", "15/4 bold", "-/-".
func ParseStyle(text string) (StyleSpec, error) {
	var s StyleSpec
	colours, attrs, hasAttrs := strings.Cut(text, " ")
	fg, bg, hasBg := strings.Cut(colours, "/")
	var err error
	if fg != "" || !hasBg {
		if s.Fg, err = parseColour(fg); err != nil {
			return s, err
		}
		s.HasFg = true
	}
	if hasBg {
		if s.Bg, err = parseColour(bg); err != nil {
			return s, err
		}
		s.HasBg = true
	}
	if hasAttrs {
		s.Attr, err = ParseAttrs(attrs)
	}
	return s, err
}

// ParseAttrs parses the attributes of a style: the attr names of section 7.2 with one space
// between them. Example: "bold underline".
func ParseAttrs(text string) (Attr, error) {
	var attr Attr
	for name := range strings.SplitSeq(text, " ") {
		a, err := parseAttr(name)
		if err != nil {
			return 0, err
		}
		attr |= a
	}
	return attr, nil
}

// parseColour parses one colour. With an explicit base, strconv.ParseUint
// accepts no sign, prefix or underscore, as the syntax requires.
func parseColour(text string) (Colour, error) {
	switch {
	case text == "-":
		return Colour{}, nil
	case strings.HasPrefix(text, "#"):
		v, err := strconv.ParseUint(text[1:], 16, 32)
		if len(text) != 7 || err != nil {
			return Colour{}, fmt.Errorf("invalid colour %q", text)
		}
		return Colour{Kind: RGB, Value: uint32(v)}, nil
	default:
		v, err := strconv.ParseUint(text, 10, 8)
		if err != nil {
			return Colour{}, fmt.Errorf("invalid colour %q", text)
		}
		return IndexColour(uint8(v)), nil
	}
}

func parseAttr(name string) (Attr, error) {
	for _, a := range attrNames {
		if a.name == name {
			return a.attr, nil
		}
	}
	if name == "" {
		return 0, errors.New("empty attribute")
	}
	return 0, fmt.Errorf("unknown attribute %q", name)
}
