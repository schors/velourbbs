package screen

import (
	"slices"
	"testing"
)

func TestDiags(t *testing.T) {
	var ds Diags
	ds.Warnf("screens/about.bbt", 3, 1, "art %q is %d cells wide", "logo.ans", 132)
	ds.Errorf("screens/main.bbt", 9, 1, "unknown field %q", "term.colours")
	ds.Errorf("screens/about.bbt", 2, 5, "second")
	if !ds.HasErrors() {
		t.Error("HasErrors = false")
	}
	ds.Sort()
	var got []string
	for _, d := range ds {
		got = append(got, d.String())
	}
	want := []string{
		"screens/about.bbt:2:5: second",
		`screens/about.bbt:3:1: warning: art "logo.ans" is 132 cells wide`,
		`screens/main.bbt:9:1: unknown field "term.colours"`,
	}
	if !slices.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
	if (Diags{{Warning: true}}).HasErrors() {
		t.Error("HasErrors = true for a warning")
	}
}
