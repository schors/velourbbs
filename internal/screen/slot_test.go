package screen

import (
	"math/rand/v2"
	"testing"
	"time"
)

// The largest width of a layout is the sum of the largest widths of its elements (live-screen, section 3.1).
func TestLayoutWidth(t *testing.T) {
	berlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		layout string
		loc    *time.Location
		want   int
	}{
		{"2006-01-02 15:04:05", time.UTC, 19},
		{"15:04:05", time.UTC, 8},
		{"15:04:05 MST", time.UTC, 12},
		{"15:04:05 MST", berlin, 13}, // CEST
		{"Monday, January 2", time.UTC, 23},
		{"Mon Jan _2", time.UTC, 10},
		{"3:04PM", time.UTC, 7},
		{"03:04 pm", time.UTC, 8},
		{"1/2/06", time.UTC, 8},
		{"15:04:05.000", time.UTC, 12},
		{"15:04:05.999999", time.UTC, 15},
		{"15:04:05,9", time.UTC, 10},
		{"05.0001", time.UTC, 7}, // "." and "00" are literal: a digit follows the run; "01" is the month
		{"2006-01-02T15:04:05Z07:00", time.UTC, 25},
		{"Z0700 Z07 Z070000 Z07:00:00", time.UTC, 27},
		{"-0700 -07 -070000 -07:00:00 -07:00", time.UTC, 34},
		{"_2006", time.UTC, 5},
		{"002 __2", time.UTC, 7},
		{"Month Janet", time.UTC, 11}, // a lower-case letter after Mon and Jan: literal text
		{"1.5", time.UTC, 5},
		{"年2006", time.UTC, 6},
		{"MST MST", berlin, 9},
		{"", time.UTC, 0},
	}
	for _, tt := range tests {
		if got := layoutWidth(tt.layout, tt.loc); got != tt.want {
			t.Errorf("%q in %s: got %d, want %d", tt.layout, tt.loc, got, tt.want)
		}
	}
}

// Each formatted time from 2000 to 2100 fits in the largest width of its layout.
func TestLayoutWidthBound(t *testing.T) {
	layouts := []string{
		"2006-01-02 15:04:05", "Monday, January 2 3:04:05.999 PM MST", "Mon Jan _2 15:04:05 2006",
		"02 Jan 06 15:04 -0700", "2006-01-02T15:04:05.000000Z07:00", "1/2 3:4:5 __2 002 pm Z07",
	}
	var locs []*time.Location
	for _, name := range []string{"UTC", "Europe/Berlin", "America/St_Johns", "Australia/Lord_Howe", "Asia/Kolkata"} {
		loc, err := time.LoadLocation(name)
		if err != nil {
			t.Fatal(err)
		}
		locs = append(locs, loc)
	}
	r := rand.New(rand.NewPCG(1, 2))
	start := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	end := time.Date(2101, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	for range 2000 {
		at := time.Unix(start+r.Int64N(end-start), r.Int64N(1e9))
		for _, loc := range locs {
			for _, l := range layouts {
				s := at.In(loc).Format(l)
				if n := len([]rune(s)); n > layoutWidth(l, loc) {
					t.Fatalf("%q in %s: %q has %d cells, more than %d", l, loc, s, n, layoutWidth(l, loc))
				}
			}
		}
	}
}

// The width of MST is the widest abbreviation of the zone from 2000 to 2100 (live-screen, section 3.1).
func TestZoneWidth(t *testing.T) {
	tests := []struct {
		zone string
		want int
	}{
		{"UTC", 3},
		{"Europe/Berlin", 4},       // CET, CEST
		{"America/New_York", 3},    // EST, EDT
		{"Asia/Dubai", 3},          // +04
		{"Australia/Lord_Howe", 5}, // +1030, +11
	}
	for _, tt := range tests {
		loc, err := time.LoadLocation(tt.zone)
		if err != nil {
			t.Fatal(err)
		}
		if got := zoneWidth(loc); got != tt.want {
			t.Errorf("%s: got %d, want %d", tt.zone, got, tt.want)
		}
	}
	if got := zoneWidth(time.FixedZone("", 3600)); got != 5 { // the form -0700
		t.Errorf("a zone without an abbreviation: got %d, want 5", got)
	}
}

// The largest widths of session.online, bbs.uptime and bbs.sessions (live-screen, section 3.1).
func TestSlotWidth(t *testing.T) {
	env := &Env{Session: SessionValues{Layout: "15:04", MaxConns: 1000}}
	tests := []struct {
		path, layout string
		want         int
	}{
		{"session.online", "", 14},
		{"bbs.uptime", "", 14},
		{"bbs.sessions", "", 4},
		{"bbs.time", "", 5},
		{"bbs.time", "2006", 4},
	}
	for _, tt := range tests {
		r, err := resolve(tt.path, nil)
		if err != nil {
			t.Fatal(err)
		}
		if got := env.slotWidth(r, tt.layout); got != tt.want {
			t.Errorf("%s %q: got %d, want %d", tt.path, tt.layout, got, tt.want)
		}
	}
	if got := len(FormatDuration(9999*24*time.Hour + 86399*time.Second)); got != durationWidth {
		t.Errorf("the duration form of 9999d 23:59:59 has %d cells, want %d", got, durationWidth)
	}
}
