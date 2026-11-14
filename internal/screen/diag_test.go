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

func TestFileDiags(t *testing.T) {
	f := FileDiags{File: "a.ans"}
	for i := range MaxFileDiags + 5 {
		if f.full() {
			t.Fatalf("full after %d diagnostics", i)
		}
		if i%2 == 0 {
			f.Errorf(1, i+1, "error %d", i)
		} else {
			f.Warnf(1, i+1, "warning %d", i)
		}
		if i == MaxFileDiags {
			break
		}
	}
	if !f.full() || len(f.Diags) != MaxFileDiags+1 {
		t.Fatalf("full %t, %d diagnostics", f.full(), len(f.Diags))
	}
	if got := f.Diags[MaxFileDiags].String(); got != "a.ans:1:101: too many diagnostics" {
		t.Errorf("last %q", got)
	}
	f.Errorf(2, 1, "dropped")
	if len(f.Diags) != MaxFileDiags+1 {
		t.Errorf("%d diagnostics after full", len(f.Diags))
	}
}

func TestDiagEscape(t *testing.T) {
	d := Diag{File: "screens/\x1b]0;x\a\x1b[2J.ans", Row: 1, Col: 1, Warning: true, Msg: "a\nb\xffc é"}
	want := `screens/\x1b]0;x\a\x1b[2J.ans:1:1: warning: a\nb\xffc é`
	if got := d.String(); got != want {
		t.Errorf("String() = %s, want %s", got, want)
	}
}
