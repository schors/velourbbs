package main

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/schors/velourbbs/internal/screen"
)

// The live fields of live-screen, section 3: velourbbs check gives no output, and the preview
// equals its golden dump. Each slot has its largest width: bbs.time in the layout [screen] time
// and in a layout with MST of Europe/Berlin (CEST: 4 cells), session.online and bbs.uptime 14 cells,
// bbs.sessions the 3 digits of max_conns = 250. The dumps show a slot in a text row,
// in a row with stretch tokens, in an art and in a list line. At 40 columns a slot is at most W.
// With --set, bbs.time shows the value in each layout, cut to its slot.
func TestLive(t *testing.T) {
	out, errOut, status := runIn(t, "live", "check")
	if out != "" || errOut != "" || status != 0 {
		t.Errorf("check: stdout %q, stderr %q, status %d", out, errOut, status)
	}
	tests := []struct {
		file string
		args []string
	}{
		{"main-80x24.dump", []string{"main"}},
		{"main-40x24.dump", []string{"main", "--size", "40x24"}},
		{"main-set-80x24.dump", []string{"main", "--set", "bbs.time=Thursday, 1 January 2026, 12:00 noon"}},
		{"list-80x24.dump", []string{"list"}},
	}
	for _, tt := range tests {
		out, errOut, status := runIn(t, "live", append([]string{"preview"}, tt.args...)...)
		if errOut != "" || status != 0 {
			t.Errorf("%s: stderr %q, status %d", tt.file, errOut, status)
		}
		golden(t, filepath.Join("testdata", "live", tt.file), out)
	}
}

// The ticks and the frames of live-screen, sections 4, 5 and 7, in velourbbs preview --keys.
// main: a wait of 250 ms gives no change, then two ticks, each with its grid dump.
// more: a wait on a More row and in the line input of ask moves the clock with no frame;
// after the input, the page shows again with the values of the last tick and ticks on.
// With --view, the grid is in the line form.
func TestLiveKeys(t *testing.T) {
	tests := []struct {
		file string
		args []string
	}{
		{"keys-main-80x24.dump", []string{"main", "--keys", "+250ms,+1s,+1s"}},
		{"keys-more-40x10.dump", []string{"more", "--size", "40x10", "--keys", "+1s,x,+1s,J,+2s,ESC,+1s"}},
		{"keys-view-40x10.dump", []string{"more", "--size", "40x10", "--view", "--keys", "x,+1s"}},
	}
	for _, tt := range tests {
		out, errOut, status := runIn(t, "live", append([]string{"preview"}, tt.args...)...)
		if errOut != "" || status != 0 {
			t.Errorf("%s: stderr %q, status %d", tt.file, errOut, status)
		}
		golden(t, filepath.Join("testdata", "live", tt.file), out)
	}
}

// A wait has a duration from 1 ms to 1 h. A wait after logoff is an error, as a key is
// (live-screen, section 7).
func TestLiveKeysErrors(t *testing.T) {
	tests := []struct {
		keys, want string
		status     int
	}{
		{"+0s", "invalid value \"+0s\" for flag -keys: invalid wait \"+0s\": must be a duration from 1ms to 1h\n", exitUsage},
		{"+2h", "invalid value \"+2h\" for flag -keys: invalid wait \"+2h\": must be a duration from 1ms to 1h\n", exitUsage},
		{"+x", "invalid value \"+x\" for flag -keys: invalid wait \"+x\": must be a duration from 1ms to 1h\n", exitUsage},
		{"G,+1s", "velourbbs preview: wait \"+1s\" after logoff\n", exitError},
	}
	for _, tt := range tests {
		_, errOut, status := runIn(t, "live", "preview", "main", "--keys", tt.keys)
		if !strings.HasPrefix(errOut, tt.want) || status != tt.status {
			t.Errorf("--keys %s: stderr %q, status %d; want %q, %d", tt.keys, errOut, status, tt.want, tt.status)
		}
	}
}

// The fixed clock moves bbs.time, session.online and bbs.uptime; a value of --set stays.
func TestClockValues(t *testing.T) {
	start := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
	sv := screen.SessionValues{Time: "x", Online: "o", Uptime: "u", Layout: "15:04:05"}
	got := clockValues(sv, map[string]bool{"session.online": true}, start, start.Add(61*time.Second))
	if got.Time != "15:05:06" || !got.Clock.Equal(start.Add(61*time.Second)) || got.Online != "o" || got.Uptime != "00:01:01" {
		t.Errorf("clockValues = %+v", got)
	}
}

// The effects of live-screen, sections 6.4–6.9, at three times with the fixed seed:
// with in and without in, with a point, an array of points and a rectangle where the effect takes them.
// twinkle: a point on the page; an array of points in the art with colors, one of them
// at the last cell of the art; a rectangle of the art with share 0.5.
// candle: a point at the last column of the page; an array of points in the art; a rectangle
// on the last row of the page with delay 1, which shows from the show.
// glint: a point on the page with an RGB peak; an array of points in the art with a theme style as peak;
// a rectangle of the art. The dump shows the grid of the last frame before the end of a wait.
// sweep: the whole art once, from 0.5 s; a rectangle on the last row of the page with every, from its delay.
// sparkle: a point on a space of the page; an array of points in the art with over and colors;
// a rectangle of the art with count 3.
// lightning: a slant bolt on the whole grid with ASCII chars; a square bolt in the art with a theme style as peak,
// on empty cells only, with a gap between two flashes.
// fragment: in of a fragment below a text row: a twinkle at the first cell and at the last cell of the fragment rows,
// then a sweep over all fragment rows once.
func TestEffects(t *testing.T) {
	out, errOut, status := runIn(t, "effects", "check")
	if out != "" || errOut != "" || status != 0 {
		t.Errorf("check: stdout %q, stderr %q, status %d", out, errOut, status)
	}
	for _, tt := range []struct{ name, keys string }{
		{"twinkle", "+1s,+1s,+1s"},
		{"candle", "+1s,+1s,+1s"},
		{"glint", "+1250ms,+1s,+1s"},
		{"sweep", "+1s,+500ms,+1s"},
		{"sparkle", "+1s,+1s,+1s"},
		{"lightning", "+1050ms,+500ms,+100ms,+100ms"},
		{"fragment", "+1s,+500ms,+1s"},
	} {
		out, errOut, status := runIn(t, "effects", "preview", tt.name, "--size", "40x10", "--keys", tt.keys)
		if errOut != "" || status != 0 {
			t.Errorf("%s: stderr %q, status %d", tt.name, errOut, status)
		}
		golden(t, filepath.Join("testdata", "effects", tt.name+"-40x10.dump"), out)
	}
}
