package screen

import "testing"

func TestParseStyle(t *testing.T) {
	base := Style{Fg: IndexColour(3), Bg: IndexColour(5), Attr: Underline}
	tests := []struct {
		text string
		want string // the result of Over(base)
	}{
		{"14", "14/5 underline"},
		{"14/1", "14/1 underline"},
		{"/1", "3/1 underline"},
		{"#ff8800/#000020", "#ff8800/#000020 underline"},
		{"15/4 bold", "15/4 bold underline"},
		{"-/-", "-/- underline"},
		{"#FF8800 blink bold", "#ff8800/5 bold underline blink"},
		{"255/0", "255/0 underline"},
	}
	for _, tt := range tests {
		spec, err := ParseStyle(tt.text)
		if err != nil {
			t.Errorf("ParseStyle(%q): %v", tt.text, err)
			continue
		}
		if got := spec.Over(base).String(); got != tt.want {
			t.Errorf("ParseStyle(%q).Over = %q, want %q", tt.text, got, tt.want)
		}
	}
}

func TestParseStyleInvalid(t *testing.T) {
	for _, text := range []string{
		"", "/", "14/", "1/2/3", "256", "-1", "+1", "1_0", "#ff880", "#ff88001", "#gg0000",
		"#0xff00", "bold", " 14", "14 ", "14  bold", "14 bolt", "14/1 Bold", "14\tbold",
	} {
		if spec, err := ParseStyle(text); err == nil {
			t.Errorf("ParseStyle(%q) = %+v, want an error", text, spec)
		}
	}
}

func TestStyleString(t *testing.T) {
	s := Style{Fg: RGBColour(1, 2, 3), Bg: Colour{}, Attr: Blink | Bold | Underline}
	if got, want := s.String(), "#010203/- bold underline blink"; got != want {
		t.Errorf("String = %q, want %q", got, want)
	}
}

func TestSetStyleApply(t *testing.T) {
	cur := Style{Fg: IndexColour(3), Bg: IndexColour(5), Attr: Underline}
	title, _ := ParseStyle("15/4 bold")
	theme := map[string]Style{"title": title.Over(Style{})}
	colour, _ := ParseStyle("/1")
	if got := (SetStyle{Name: "title"}).Apply(cur, theme).String(); got != "15/4 bold" {
		t.Errorf("theme style: %s", got)
	}
	if got := (SetStyle{Spec: colour}).Apply(cur, theme).String(); got != "3/1 underline" {
		t.Errorf("colour change: %s", got)
	}
}
