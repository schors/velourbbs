package bbt_test

import "testing"

// TestLiveTokens covers the live mark of a field token and its errors (live-screen, section 3).
func TestLiveTokens(t *testing.T) {
	tests := []struct {
		body string
		want string // the body, or the first diagnostic
	}{
		{"{bbs.time live}", `{bbs.time live} /`},
		{`{bbs.time live "15:04:05"}`, `{bbs.time live "15:04:05"} /`},
		{"{session.online live}", `{session.online live} /`},
		{"{bbs.uptime live    }", `{bbs.uptime:21 live} /`},
		{"{    bbs.sessions live}", `{bbs.sessions:23R live} /`},
		{`{bbs.time live "Jan 2 15:04" }`, `{bbs.time:30 live "Jan 2 15:04"} /`},
		{`{bbs.time live "年" }`, `{bbs.time:21 live "年"} /`}, // the width of the token in cells
		// An alias holds the mark. The mark on another field gives no slot.
		{"{clock}", `{bbs.time live "15:04:05"} /`},
		{"{clock   }", `{bbs.time:10 live "15:04:05"} /`},
		{"{up}", `{bbs.uptime live} /`},
		{"{bbs.name live}", `{bbs.name} /`},
		{"{term.cols live}", `{term.cols} /`},
		{"{name}", `{bbs.name} /`},
		// The errors of live-screen, section 3.
		{"{bbs.time foo}", `screens/s.bbt:1:1: "foo" after the path "bbs.time" is not the live mark`},
		{"{bbs.time  live}", `screens/s.bbt:1:1: " live" after the path "bbs.time" is not the live mark`},
		{"{bbs.time lived}", `screens/s.bbt:1:1: "lived" after the path "bbs.time" is not the live mark`},
		{`{bbs.time live 15:04}`, `screens/s.bbt:1:1: "live 15:04" after the path "bbs.time" is not the live mark`},
		{`{bbs.time live "15" "04"}`, `screens/s.bbt:1:1: "live \"15\" \"04\"" after the path "bbs.time" is not the live mark`},
		{`{bbs.time live  "15"}`, `screens/s.bbt:1:1: "live  \"15\"" after the path "bbs.time" is not the live mark`},
		{"{clock live}", `screens/s.bbt:1:1: the live mark on the alias "clock": an alias takes no mark`},
		{`{bbs.time "15:04"}`, `screens/s.bbt:1:1: layout "15:04" without "live"`},
		{`{bbs.uptime live "15:04"}`, `screens/s.bbt:1:1: layout "15:04" on "bbs.uptime": only bbs.time takes a layout`},
		{`{bbs.time live ""}`, `screens/s.bbt:1:1: empty layout`},
		{"{bbs.time live \"15\u008504\"}", `screens/s.bbt:1:1: layout "15\u008504" has a control character`},
		{"{#if clock == \"x\"}a{/if}", `screens/s.bbt:1:1: alias "clock" has the live mark: a condition takes the values of the show`},
		{"{#if bbs.sessions live > 1}a{/if}", `screens/s.bbt:1:1: integer field "bbs.sessions" needs an operator`},
	}
	for _, tt := range tests {
		screens, msgs := load(map[string]string{"s.bbt": tt.body})
		got := ""
		if screens != nil {
			got = show(screens["s"].Body)
		} else if len(msgs) > 0 {
			got = msgs[0]
		}
		if got != tt.want {
			t.Errorf("%q:\n got %s\nwant %s\nall %q", tt.body, got, tt.want, msgs)
		}
	}
}

// TestLiveWhen covers an alias with the live mark in when (live-screen, section 3).
func TestLiveWhen(t *testing.T) {
	_, msgs := load(map[string]string{"s.bbt": "+++\n[[item]]\nkey = \"A\"\ndo = \"logoff\"\nwhen = 'up == \"x\"'\n+++\n"})
	want := `screens/s.bbt:5:8: item "A": when: alias "up" has the live mark: a condition takes the values of the show`
	if len(msgs) != 1 || msgs[0] != want {
		t.Errorf("got %q\nwant %q", msgs, want)
	}
}
