package config

import (
	"slices"
	"strings"
	"testing"
)

// An alias of [fields] can hold the live mark; an invalid mark is an error at its value
// (screens-toml, section 3.3, and live-screen, section 3).
func TestLiveAliases(t *testing.T) {
	data := strings.Join([]string{
		`[fields]`,
		`clock = 'bbs.time live "15:04:05"'`,
		`up = "bbs.uptime live"`,
		`a = "bbs.time foo"`,
		`b = 'bbs.uptime live "15:04"'`,
		`c = 'bbs.time live ""'`,
		`d = 'bbs.time "15:04"'`,
	}, "\n")
	cfg, ds := parseScreens([]byte(data))
	want := []string{
		`screens/screens.toml:4:5: fields: alias "a": "foo" after the path "bbs.time" is not the live mark`,
		`screens/screens.toml:5:5: fields: alias "b": layout "15:04" on "bbs.uptime": only bbs.time takes a layout`,
		`screens/screens.toml:6:5: fields: alias "c": empty layout`,
		`screens/screens.toml:7:5: fields: alias "d": layout "15:04" without "live"`,
	}
	if got := messages(ds); !slices.Equal(got, want) {
		t.Errorf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	clock, up := cfg.Fields["clock"], cfg.Fields["up"]
	if clock.Ref.String() != "bbs.time" || !clock.Live || clock.Layout != "15:04:05" {
		t.Errorf("clock = %+v", clock)
	}
	if up.Ref.String() != "bbs.uptime" || !up.Live || up.Layout != "" {
		t.Errorf("up = %+v", up)
	}
}
