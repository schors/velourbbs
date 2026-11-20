package bbs

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/schors/velourbbs/internal/bbt"
	"github.com/schors/velourbbs/internal/config"
	"github.com/schors/velourbbs/internal/screen"
	"github.com/schors/velourbbs/internal/tomlcheck"
)

// screens makes screens from specs. A spec is a list of items "<KEY> <do>".
// Each screen has the front matter of withFront.
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
			if target, ok := strings.CutPrefix(do, "gosub "); ok {
				act = screen.Action{Op: screen.Gosub, Screen: target}
			}
			s.Items = append(s.Items, screen.Item{Key: k, Do: act})
		}
		withFront(s)
	}
	return ss
}

// withFront gives s a front matter with its items. In the file, the n-th item has
// its [[item]] header at (4n-2):1, its do at 4n:6 and its args at (4n+1):8.
func withFront(s *screen.Screen) *screen.Screen {
	var b strings.Builder
	for _, it := range s.Items {
		fmt.Fprintf(&b, "[[item]]\nkey = %q\ndo = \"x\"\n", it.Key)
		if it.Do.Args != nil {
			var args []string
			for _, k := range slices.Sorted(maps.Keys(it.Do.Args)) {
				args = append(args, fmt.Sprintf("%q = %q", k, it.Do.Args[k]))
			}
			fmt.Fprintf(&b, "args = { %s }", strings.Join(args, ", "))
		}
		b.WriteString("\n")
	}
	var doc struct {
		Item []struct {
			Key  string            `toml:"key"`
			Do   string            `toml:"do"`
			Args map[string]string `toml:"args"`
		} `toml:"item"`
	}
	root, errs := tomlcheck.DecodeFront([]byte(b.String()), &doc)
	if errs != nil {
		panic(fmt.Sprint(errs))
	}
	s.Front = root
	return s
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
	"main":     {"S gosub sysinfo", "G logoff"},
	"sysinfo":  {"ANY goto main"},
	"goodbye":  {},
	"shutdown": {},
}

func TestCheck(t *testing.T) {
	// The start screen error has the position of menu.start (menu-nav, section 6).
	start := config.Config{Start: "main", StartPos: tomlcheck.Pos{Line: 12, Col: 9}, Lang: "en", Langs: []string{"en"}}
	tests := []struct {
		name   string
		change map[string][]string // the screens to add or replace; nil deletes a screen
		start  string
		want   []string
	}{
		{name: "valid"},
		// An item error has the position of the do of the item.
		{name: "unknown screen", change: map[string][]string{"main": {"S goto sysinf", "G logoff"}}, want: []string{
			`screens/main.bbt:4:6: item "S": do: unknown screen "sysinf"`,
		}},
		{name: "gosub unknown screen", change: map[string][]string{"main": {"S gosub sysinf", "G logoff"}}, want: []string{
			`screens/main.bbt:4:6: item "S": do: unknown screen "sysinf"`,
		}},
		{name: "gosub goodbye", change: map[string][]string{"main": {"S gosub goodbye", "G logoff"}}, want: []string{
			`screens/main.bbt:4:6: item "S": do: screen "goodbye" has no items`,
		}},
		{name: "goto goodbye", change: map[string][]string{"sysinfo": {"ESC goto main", "ANY goto goodbye"}}, want: []string{
			`screens/sysinfo.bbt:8:6: item "ANY": do: screen "goodbye" has no items`,
		}},
		{name: "goto shutdown", change: map[string][]string{"sysinfo": {"ESC goto main", "ANY goto shutdown"}}, want: []string{
			`screens/sysinfo.bbt:8:6: item "ANY": do: screen "shutdown" has no items`,
		}},
		{name: "no items", change: map[string][]string{"about": {}}, want: []string{
			`screens/about.bbt:1:1: screen has no items`,
		}},
		{name: "no main", change: map[string][]string{"main": nil}, want: []string{
			`screens/screens.toml:12:9: menu: start: unknown screen "main"`,
			`screens/sysinfo.bbt:4:6: item "ANY": do: unknown screen "main"`,
		}},
		{name: "start unknown", start: "mian", want: []string{
			`screens/screens.toml:12:9: menu: start: unknown screen "mian"`,
		}},
		{name: "start without items", start: "goodbye", want: []string{
			`screens/screens.toml:12:9: menu: start: screen "goodbye" has no items`,
		}},
		{name: "start shutdown", start: "shutdown", want: []string{
			`screens/screens.toml:12:9: menu: start: screen "shutdown" has no items`,
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
			got := messages(Check(cfg, bbt.Sets{"en": screens(specs)}, nil))
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
	got := messages(Check(config.Config{Start: "main", Lang: "en", Langs: []string{"en"}}, bbt.Sets{"en": ss}, ds))
	if len(got) != screen.MaxFileDiags+2 {
		t.Fatalf("%d diagnostics, want %d", len(got), screen.MaxFileDiags+2)
	}
	if got[0] != "screens/main.bbt:1:1: warning: w" || got[len(got)-1] != "screens/x.ans:1:1: warning: w" ||
		!slices.Contains(got, "screens/main.bbt:400:6: too many diagnostics") {
		t.Errorf("first %q, last %q, no limit error", got[0], got[len(got)-1])
	}
}

// The args of a goto or a gosub must equal the arguments of the target. The start screen
// and goodbye and shutdown cannot have arguments (menu-nav, section 6).
func TestCheckArgs(t *testing.T) {
	goto_ := func(key, target string, args map[string]string) screen.Item {
		return screen.Item{Key: key, Do: screen.Action{Op: screen.Goto, Screen: target, Args: args}}
	}
	gosub := func(key, target string, args map[string]string) screen.Item {
		return screen.Item{Key: key, Do: screen.Action{Op: screen.Gosub, Screen: target, Args: args}}
	}
	ss := map[string]*screen.Screen{
		"main": withFront(&screen.Screen{Name: "main", File: "screens/main.bbt", Items: []screen.Item{
			goto_("1", "view", map[string]string{"f": "a", "n": "1"}),
			goto_("2", "view", map[string]string{"f": "a"}),
			goto_("3", "view", nil),
			goto_("4", "view", map[string]string{"f": "a", "n": "1", "x": "y"}),
			goto_("5", "main", map[string]string{"x": "y"}),
			goto_("6", "none", map[string]string{"x": "y"}),
			gosub("7", "view", map[string]string{"f": "a", "n": "1"}),
			gosub("8", "view", map[string]string{"f": "a", "x": "y"}),
		}}),
		"view": withFront(&screen.Screen{Name: "view", File: "screens/view.bbt", Args: []string{"f", "n"},
			Items: []screen.Item{goto_("Q", "main", map[string]string{})}}),
		"goodbye":  {Name: "goodbye", File: "screens/goodbye.bbt", Args: []string{"x"}},
		"shutdown": {Name: "shutdown", File: "screens/shutdown.bbt", Args: []string{"x"}},
	}
	want := []string{
		`screens/goodbye.bbt:1:1: the goodbye screen cannot have arguments`,
		// An args error at the value of args, else at the [[item]] header.
		`screens/main.bbt:9:8: item "2": args: missing argument "n" of screen "view"`,
		`screens/main.bbt:10:1: item "3": args: missing argument "f" of screen "view"`,
		`screens/main.bbt:10:1: item "3": args: missing argument "n" of screen "view"`,
		`screens/main.bbt:17:8: item "4": args: screen "view" has no argument "x"`,
		`screens/main.bbt:21:8: item "5": args: screen "main" has no argument "x"`,
		`screens/main.bbt:24:6: item "6": do: unknown screen "none"`,
		`screens/main.bbt:33:8: item "8": args: missing argument "n" of screen "view"`,
		`screens/main.bbt:33:8: item "8": args: screen "view" has no argument "x"`,
		`screens/shutdown.bbt:1:1: the shutdown screen cannot have arguments`,
	}
	got := messages(Check(config.Config{Start: "main", Lang: "en", Langs: []string{"en"}}, bbt.Sets{"en": ss}, nil))
	if !slices.Equal(got, want) {
		t.Errorf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}

	got = messages(Check(config.Config{Start: "view", StartPos: tomlcheck.Pos{Line: 12, Col: 9}, Lang: "en", Langs: []string{"en"}},
		bbt.Sets{"en": ss}, nil))
	if !slices.Contains(got, `screens/screens.toml:12:9: menu: start: screen "view" has arguments`) {
		t.Errorf("start view: got\n%s", strings.Join(got, "\n"))
	}
}
