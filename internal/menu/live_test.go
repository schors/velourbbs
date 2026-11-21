package menu

import (
	"errors"
	"io"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/schors/velourbbs/internal/config"
	"github.com/schors/velourbbs/internal/screen"
)

// clockValues gives the values of a session on f: bbs.time with the clock of f (live-screen, section 4).
func clockValues(f *fakeTerm) Values {
	return func(*Session) screen.SessionValues {
		now := f.Now()
		return screen.SessionValues{Time: now.Format("15:04:05"), Clock: now, Layout: "15:04:05", MaxConns: 64}
	}
}

// liveMain is a main screen with a live clock and an item with ask.
const liveMain = `+++
[[item]]
key = "J"
do = "gosub msg"
args = { n = "{in}" }
ask = "N: "
[[item]]
key = "G"
do = "logoff"
+++
T {bbs.time live}
`

// A live page ticks at each whole second while it waits for a key. A frame comes at least 50 ms
// after the frame before. A More row and the line input of ask have no ticks; after the input,
// the page shows again and ticks (live-screen, sections 4 and 5).
func TestLiveTicks(t *testing.T) {
	tests := []struct {
		name  string
		start time.Duration // the show time after fakeStart
		rows  int
		body  string
		keys  []string
		want  []string
	}{
		{"two ticks", 0, 10, liveMain, []string{"+250ms", "+2s", "G"}, []string{
			"page: T 15:04:05", "wait +250ms", "frame 06.000: T 15:04:06", "frame 07.000: T 15:04:07", "wait +2s",
			"page: Bye",
		}},
		{"50 ms after the show", 980 * time.Millisecond, 10, liveMain, []string{"+100ms"}, []string{
			"page: T 15:04:05", "frame 06.030: T 15:04:06", "wait +100ms",
		}},
		{"more", 0, 2, strings.Replace(liveMain, "T {bbs.time live}", "A {bbs.time live}\nB\nC", 1),
			[]string{"+2s", "x", "+1s"}, []string{
				"page: A 15:04:05", "wait +2s no frame", "page: B", "wait +1s",
			}},
		{"ask", 0, 10, liveMain, []string{"J", "+2s", "ESC", "+1s"}, []string{
			"page: T 15:04:05", "page: N:", "line 1:4 76", "wait +2s no frame", "page: T 15:04:05",
			"frame 08.000: T 15:04:08", "wait +1s",
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ss := load(t, map[string]string{"main": tc.body, "info": "", "long": "", "msg": "+++\n[[item]]\nkey = \"B\"\ndo = \"back\"\n+++\n{arg.n}\n"})
			ft := &fakeTerm{cols: []int{80}, rows: []int{tc.rows}, keys: tc.keys, now: fakeStart.Add(tc.start)}
			_, err := runSession(ft, config.Config{}, ss, "main", clockValues(ft))
			if err != nil && !errors.Is(err, io.EOF) {
				t.Fatal(err)
			}
			if !slices.Equal(ft.log, tc.want) {
				t.Errorf("log:\n%s\nwant:\n%s", strings.Join(ft.log, "\n"), strings.Join(tc.want, "\n"))
			}
		})
	}
}

// A cursor move in a live list screen renders with the values of the show; the slots of its base
// have the values of the last tick. The ticks go on (live-screen, section 4).
func TestLiveListMove(t *testing.T) {
	ss := loadList(t, "{bbs.time live}\n")
	ft := &fakeTerm{cols: []int{80}, keys: []string{"L", "+1s", "DOWN", "+1s"}}
	_, err := runSession(ft, config.Config{Start: "main"}, ss, "main", clockValues(ft))
	want := []string{
		"page: Main", "page: 15:04:05", "frame 06.000: 15:04:06|>a.txt| b.txt", "wait +1s",
		"update: 15:04:06| a.txt|>b.txt", "frame 07.000: 15:04:07| a.txt|>b.txt", "wait +1s",
	}
	if !errors.Is(err, io.EOF) || !slices.Equal(ft.log, want) {
		t.Errorf("err %v, log:\n%s\nwant:\n%s", err, strings.Join(ft.log, "\n"), strings.Join(want, "\n"))
	}
}
