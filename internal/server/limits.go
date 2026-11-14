package server

import (
	"container/list"
	"net/netip"
	"sync"
	"time"

	"github.com/schors/velourbbs/internal/config"
)

// The connect counter of section 9.2: its address limit and its time window in seconds.
const (
	maxAddrs   = 10000
	windowSecs = 60
)

// Limits are the connection limits of section 9.2. They are safe for concurrent use.
type Limits struct {
	maxConns, perAddr, perMin int
	allow                     []netip.Prefix

	mu    sync.Mutex
	conns int
	// open is the count of open connections for each key. max_conns bounds its size.
	open map[netip.Addr]int
	// times is the connect counts for each key. order holds its values,
	// ordered by the last connect: the front is the oldest.
	times map[netip.Addr]*list.Element
	order list.List
}

// connects is the connect counts of one key in the last 60 s, in steps of 1 s.
type connects struct {
	key netip.Addr
	// sec is the second of the last connect. counts[i] is the count of the second
	// that is i modulo 60. sum is the sum of counts.
	sec    int64
	counts [windowSecs]uint32
	sum    int
}

// NewLimits makes the limits of c.
func NewLimits(c config.Telnet) *Limits {
	return &Limits{
		maxConns: c.MaxConns, perAddr: c.MaxConnsPerAddr, perMin: c.MaxConnectsPerMin,
		allow: c.Allow,
		open:  map[netip.Addr]int{}, times: map[netip.Addr]*list.Element{},
	}
}

// Open counts a connection for max_conns. It returns false when the limit refuses it.
func (l *Limits) Open() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.conns >= l.maxConns {
		return false
	}
	l.conns++
	return true
}

// Close ends a connection that Open counted.
func (l *Limits) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.conns--
}

// OpenAddr counts a connect from the client address a at the time now
// for the address limits. It returns false when a limit refuses it.
// A refused connect also counts for max_connects_per_min.
// Call it only for a connection that Open counted.
func (l *Limits) OpenAddr(a netip.Addr, now time.Time) bool {
	a = a.Unmap().WithZone("")
	if l.allowed(a) {
		return true
	}
	k := key(a)
	l.mu.Lock()
	defer l.mu.Unlock()
	ok := l.connect(k, now)
	if ok && l.open[k] >= l.perAddr {
		ok = false
	}
	if ok {
		l.open[k]++
	}
	return ok
}

// CloseAddr ends a connection that OpenAddr counted.
func (l *Limits) CloseAddr(a netip.Addr) {
	a = a.Unmap().WithZone("")
	if l.allowed(a) {
		return
	}
	k := key(a)
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.open[k]--; l.open[k] <= 0 {
		delete(l.open, k)
	}
}

func (l *Limits) allowed(a netip.Addr) bool {
	for _, p := range l.allow {
		if p.Contains(a) {
			return true
		}
	}
	return false
}

// key is "one address" of section 9.2: an IPv4 address or an IPv6 /64 prefix.
// a is not IPv4-mapped and has no zone.
func key(a netip.Addr) netip.Addr {
	if a.Is4() {
		return a
	}
	p, _ := a.Prefix(64)
	return p.Addr()
}

// connect counts a connect of the key k at the time now. It returns false when k has
// max_connects_per_min connects in the 60 s before: the second of now and the 59 seconds
// before it.
func (l *Limits) connect(k netip.Addr, now time.Time) bool {
	e, ok := l.times[k]
	if ok {
		l.order.MoveToBack(e)
	} else {
		if len(l.times) >= maxAddrs {
			old := l.order.Front()
			delete(l.times, l.order.Remove(old).(*connects).key)
		}
		e = l.order.PushBack(&connects{key: k, sec: now.Unix()})
		l.times[k] = e
	}
	c := e.Value.(*connects)
	// A clock that goes back counts in the second of the last connect.
	sec := max(now.Unix(), c.sec)
	if sec-c.sec >= windowSecs {
		c.counts, c.sum = [windowSecs]uint32{}, 0
	} else {
		for s := c.sec + 1; s <= sec; s++ {
			i := s % windowSecs
			c.sum -= int(c.counts[i])
			c.counts[i] = 0
		}
	}
	c.sec = sec
	ok = c.sum < l.perMin
	c.counts[sec%windowSecs]++
	c.sum++
	return ok
}
