package menu

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"newbbs/internal/config"
	"newbbs/internal/screen"
)

// screens makes screens from specs. A spec is a list of items "<KEY> <do>".
// The do of the n-th item is on line n+3.
func screens(specs map[string][]string) map[string]*screen.Screen {
	ss := map[string]*screen.Screen{}
	for name := range specs {
		ss[name] = &screen.Screen{Name: name, File: "screens/" + name + ".bbt"}
	}
	for name, spec := range specs {
		s := ss[name]
		for _, line := range spec {
			k, do, _ := strings.Cut(line, " ")
			act := screen.Action{Op: screen.Logoff}
			if target, ok := strings.CutPrefix(do, "goto "); ok {
				act = screen.Action{Op: screen.Goto, Screen: target}
			}
			s.Items = append(s.Items, screen.Item{Key: k, Do: act})
		}
	}
	return ss
}

func messages(ds screen.Diags) []string {
	var s []string
	for _, d := range ds {
		s = append(s, d.String())
	}
	return s
}

// valid is a valid set of screens. Each case changes it.
var valid = map[string][]string{
	"main":    {"S goto sysinfo", "G logoff"},
	"sysinfo": {"ANY goto main"},
	"goodbye": {},
}

func TestCheck(t *testing.T) {
	start := config.Config{Start: "main"}
	tests := []struct {
		name   string
		change map[string][]string // the screens to add or replace; nil deletes a screen
		start  string
		want   []string
	}{
		{name: "valid"},
		{name: "unknown screen", change: map[string][]string{"main": {"S goto sysinf", "G logoff"}}, want: []string{
			`screens/main.bbt:2:1: item "S": do: unknown screen "sysinf"`,
		}},
		{name: "goto goodbye", change: map[string][]string{"sysinfo": {"ESC goto main", "ANY goto goodbye"}}, want: []string{
			`screens/sysinfo.bbt:2:1: item "ANY": do: screen "goodbye" has no items`,
		}},
		{name: "no items", change: map[string][]string{"about": {}}, want: []string{
			`screens/about.bbt:1:1: screen has no items`,
		}},
		{name: "no main", change: map[string][]string{"main": nil}, want: []string{
			`bbs.toml:1:1: menu: start: unknown screen "main"`,
			`screens/sysinfo.bbt:2:1: item "ANY": do: unknown screen "main"`,
		}},
		{name: "start unknown", start: "mian", want: []string{
			`bbs.toml:1:1: menu: start: unknown screen "mian"`,
		}},
		{name: "start without items", start: "goodbye", want: []string{
			`bbs.toml:1:1: menu: start: screen "goodbye" has no items`,
		}},
		{name: "start valid", start: "sysinfo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			specs := map[string][]string{}
			for name, spec := range valid {
				specs[name] = spec
			}
			for name, spec := range tt.change {
				if spec == nil {
					delete(specs, name)
				} else {
					specs[name] = spec
				}
			}
			cfg := start
			if tt.start != "" {
				cfg.Start = tt.start
			}
			got := messages(Check(cfg, screens(specs), nil))
			if !slices.Equal(got, tt.want) {
				t.Errorf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(tt.want, "\n"))
			}
		})
	}
}

// Check keeps the diagnostics of the load and adds its errors up to the limit of a file.
func TestCheckLimit(t *testing.T) {
	var items []string
	for i := range screen.MaxFileDiags {
		items = append(items, fmt.Sprintf("K%d goto none", i))
	}
	ss := screens(map[string][]string{"main": items})
	var ds screen.Diags
	ds.Warnf("screens/main.bbt", 1, 1, "w")
	ds.Warnf("screens/x.ans", 1, 1, "w")
	got := messages(Check(config.Config{Start: "main"}, ss, ds))
	if len(got) != screen.MaxFileDiags+2 {
		t.Fatalf("%d diagnostics, want %d", len(got), screen.MaxFileDiags+2)
	}
	if got[0] != "screens/main.bbt:1:1: warning: w" || got[len(got)-1] != "screens/x.ans:1:1: warning: w" ||
		!slices.Contains(got, "screens/main.bbt:2:1: too many diagnostics") {
		t.Errorf("first %q, last %q, no limit error", got[0], got[len(got)-1])
	}
}
