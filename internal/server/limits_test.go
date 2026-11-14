package server_test

import (
	"net/netip"
	"slices"
	"testing"
	"time"

	"github.com/schors/velourbbs/internal/config"
	"github.com/schors/velourbbs/internal/server"
)

var t0 = time.Date(2026, 11, 13, 12, 0, 0, 0, time.UTC)

func limits(maxConns, perAddr, perMin int, allow ...string) *server.Limits {
	c := config.Telnet{MaxConns: maxConns, MaxConnsPerAddr: perAddr, MaxConnectsPerMin: perMin}
	for _, s := range allow {
		c.Allow = append(c.Allow, netip.MustParsePrefix(s))
	}
	return server.NewLimits(c)
}

func ip(s string) netip.Addr { return netip.MustParseAddr(s) }

// openN calls OpenAddr n times at now and gives the count of true results.
func openN(l *server.Limits, a string, n int, now time.Time) int {
	ok := 0
	for range n {
		if l.OpenAddr(ip(a), now) {
			ok++
		}
	}
	return ok
}

func TestMaxConns(t *testing.T) {
	l := limits(2, 1, 1)
	got := []bool{l.Open(), l.Open(), l.Open()}
	if !slices.Equal(got, []bool{true, true, false}) {
		t.Fatalf("Open = %v, want 2 connections", got)
	}
	l.Close()
	if !l.Open() {
		t.Error("Open after Close = false")
	}
}

func TestPerAddr(t *testing.T) {
	l := limits(100, 2, 100)
	if got := openN(l, "192.0.2.1", 3, t0); got != 2 {
		t.Fatalf("%d connections, want 2", got)
	}
	if !l.OpenAddr(ip("192.0.2.2"), t0) {
		t.Error("another address is refused")
	}
	l.CloseAddr(ip("192.0.2.1"))
	if !l.OpenAddr(ip("192.0.2.1"), t0) {
		t.Error("OpenAddr after CloseAddr = false")
	}
	l.CloseAddr(ip("192.0.2.1"))
	l.CloseAddr(ip("192.0.2.1"))
	if got := openN(l, "192.0.2.1", 3, t0); got != 2 {
		t.Errorf("%d connections after all CloseAddr, want 2", got)
	}
}

// One address is an IPv4 address, an IPv6 /64, and an IPv4-mapped address is IPv4.
func TestKey(t *testing.T) {
	tests := []struct{ a, b string }{
		{"2001:db8:1:2::1", "2001:db8:1:2:ffff:ffff:ffff:ffff"},
		{"::ffff:192.0.2.1", "192.0.2.1"},
		{"fe80::1%eth0", "fe80::2"},
	}
	for _, tt := range tests {
		l := limits(100, 1, 100)
		if !l.OpenAddr(ip(tt.a), t0) || l.OpenAddr(ip(tt.b), t0) {
			t.Errorf("%s and %s are not one address", tt.a, tt.b)
		}
		l.CloseAddr(ip(tt.b))
		if !l.OpenAddr(ip(tt.a), t0) {
			t.Errorf("%s: CloseAddr(%s) did not close", tt.a, tt.b)
		}
	}
	for _, other := range []string{"2001:db8:1:3::1", "192.0.2.2", "::ffff:192.0.2.2", "::c000:201"} {
		l := limits(100, 1, 100)
		if !l.OpenAddr(ip("192.0.2.1"), t0) || !l.OpenAddr(ip("2001:db8:1:2::1"), t0) || !l.OpenAddr(ip(other), t0) {
			t.Errorf("%s is refused", other)
		}
	}
}

func TestAllow(t *testing.T) {
	l := limits(100, 1, 1, "192.0.2.0/24", "2001:db8::1/128")
	for _, a := range []string{"192.0.2.7", "::ffff:192.0.2.7", "2001:db8::1"} {
		if got := openN(l, a, 5, t0); got != 5 {
			t.Errorf("%s: %d of 5 connections", a, got)
		}
		l.CloseAddr(ip(a))
	}
	if got := openN(l, "2001:db8::2", 2, t0); got != 1 {
		t.Errorf("2001:db8::2: %d connections, want 1", got)
	}
}

func TestPerMin(t *testing.T) {
	l := limits(100, 100, 3)
	at := func(d time.Duration) bool { return l.OpenAddr(ip("192.0.2.1"), t0.Add(d)) }
	if got := openN(l, "192.0.2.1", 4, t0); got != 3 {
		t.Fatalf("%d connects, want 3", got)
	}
	// The window is the second of the connect and the 59 seconds before it.
	if at(59*time.Second + 999*time.Millisecond) {
		t.Error("a connect in 60 s is not refused")
	}
	if !at(60 * time.Second) {
		t.Error("a connect after 60 s is refused")
	}
	// The refused connect at 59 s counts: 59 s, 60 s and 61 s are 3 connects.
	if !at(61*time.Second) || at(62*time.Second) {
		t.Error("the refused connect does not count")
	}
	// After 60 s without a connect, all counts are gone.
	if got := openN(l, "192.0.2.1", 4, t0.Add(10*time.Minute)); got != 3 {
		t.Errorf("%d connects after 10 min, want 3", got)
	}
}

// A clock that goes back counts in the second of the last connect.
func TestPerMinClockBack(t *testing.T) {
	l := limits(100, 100, 2)
	a := ip("192.0.2.1")
	if !l.OpenAddr(a, t0.Add(time.Hour)) || !l.OpenAddr(a, t0) || l.OpenAddr(a, t0.Add(time.Hour)) {
		t.Error("a connect with an older time does not count")
	}
}

// A connect that max_conns_per_addr refuses counts for max_connects_per_min.
func TestPerMinRefused(t *testing.T) {
	l := limits(100, 1, 2)
	a := ip("192.0.2.1")
	if !l.OpenAddr(a, t0) || l.OpenAddr(a, t0.Add(time.Second)) {
		t.Fatal("want 1 connection")
	}
	l.CloseAddr(a)
	if l.OpenAddr(a, t0.Add(2*time.Second)) {
		t.Error("a third connect in 60 s is not refused")
	}
}

// The counter keeps 10000 addresses. It removes the address with the oldest connect.
func TestMaxAddrs(t *testing.T) {
	const maxAddrs = 10000
	l := limits(100, 100, 1)
	a, b := ip("2001:db8::a"), ip("2001:db8:0:1::b")
	l.OpenAddr(a, t0)
	l.OpenAddr(b, t0)
	if l.OpenAddr(a, t0.Add(time.Second)) {
		t.Fatal("a second connect is not refused")
	}
	other := ip("10.0.0.0")
	for range maxAddrs - 1 {
		l.OpenAddr(other, t0.Add(2*time.Second))
		other = other.Next()
	}
	// The loop fills the counter and removes b: its last connect is the oldest.
	// a stays and is refused. b comes back as a new key and is accepted.
	if l.OpenAddr(a, t0.Add(3*time.Second)) {
		t.Error("an address with a newer connect is removed")
	}
	if !l.OpenAddr(b, t0.Add(3*time.Second)) {
		t.Error("the oldest address is not removed")
	}
}
