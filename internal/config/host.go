package config

import (
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"time"
)

// Telnet is the table [telnet] (bbs-toml, section 3.1).
type Telnet struct {
	// Listen is the TCP address of the server: <host>:<port>.
	Listen string
	// Proxy is the prefixes of the PROXY protocol senders.
	Proxy []netip.Prefix
	// ProbeTimeout is the time from accept until the probe ends.
	ProbeTimeout time.Duration
}

// Limits is the table [limits] (bbs-toml, section 3.3).
type Limits struct {
	// MaxConns is the maximum number of connections.
	MaxConns int
	// MaxConnsPerAddr is the maximum number of connections from one address.
	MaxConnsPerAddr int
	// MaxConnectsPerMin is the maximum number of connects from one address in 60 s.
	MaxConnectsPerMin int
	// Allow is the prefixes without the address limits.
	Allow []netip.Prefix
	// IdleTimeout and OutputTimeout are the timeouts of the session.
	IdleTimeout   time.Duration
	OutputTimeout time.Duration
}

// The defaults of bbs-toml, sections 3.1 and 3.3.
var (
	defaultTelnet = Telnet{Listen: ":2323", ProbeTimeout: 5 * time.Second}
	defaultLimits = Limits{
		MaxConns:          64,
		MaxConnsPerAddr:   4,
		MaxConnectsPerMin: 10,
		IdleTimeout:       600 * time.Second,
		OutputTimeout:     60 * time.Second,
	}
)

// maxInt is the largest integer of bbs.toml. It keeps a timeout in a time.Duration.
const maxInt = 1000000

// telnetDoc is the plain decode of [telnet]. A nil pointer is a key that is not set.
type telnetDoc struct {
	Listen       *string   `toml:"listen"`
	Proxy        *[]string `toml:"proxy"`
	ProbeTimeout *int64    `toml:"probe_timeout"`
}

// limitsDoc is the plain decode of [limits]. A nil pointer is a key that is not set.
type limitsDoc struct {
	MaxConns          *int64    `toml:"max_conns"`
	MaxConnsPerAddr   *int64    `toml:"max_conns_per_addr"`
	MaxConnectsPerMin *int64    `toml:"max_connects_per_min"`
	Allow             *[]string `toml:"allow"`
	IdleTimeout       *int64    `toml:"idle_timeout"`
	OutputTimeout     *int64    `toml:"output_timeout"`
}

// errorFunc reports an error of a value.
type errorFunc func(format string, args ...any)

// telnet gives the settings of [telnet]. An invalid value keeps its default.
func telnet(d telnetDoc, errorf errorFunc) Telnet {
	c := defaultTelnet
	if d.Listen != nil {
		if err := checkListen(*d.Listen); err != nil {
			errorf("telnet.listen: %v", err)
		} else {
			c.Listen = *d.Listen
		}
	}
	setPrefixes("telnet.proxy", d.Proxy, &c.Proxy, errorf)
	setSeconds("telnet.probe_timeout", d.ProbeTimeout, &c.ProbeTimeout, errorf)
	return c
}

// limits gives the settings of [limits]. An invalid value keeps its default.
func limits(d limitsDoc, errorf errorFunc) Limits {
	c := defaultLimits
	setInt("limits.max_conns", d.MaxConns, &c.MaxConns, errorf)
	setInt("limits.max_conns_per_addr", d.MaxConnsPerAddr, &c.MaxConnsPerAddr, errorf)
	setInt("limits.max_connects_per_min", d.MaxConnectsPerMin, &c.MaxConnectsPerMin, errorf)
	setPrefixes("limits.allow", d.Allow, &c.Allow, errorf)
	setSeconds("limits.idle_timeout", d.IdleTimeout, &c.IdleTimeout, errorf)
	setSeconds("limits.output_timeout", d.OutputTimeout, &c.OutputTimeout, errorf)
	return c
}

// checkInt reports whether the integer n of key is from 1 to maxInt.
func checkInt(key string, n int64, errorf errorFunc) bool {
	if n < 1 || n > maxInt {
		errorf("%s: %d is not from 1 to %d", key, n, maxInt)
		return false
	}
	return true
}

// setInt sets dst to src when src is set and valid.
func setInt(key string, src *int64, dst *int, errorf errorFunc) {
	if src != nil && checkInt(key, *src, errorf) {
		*dst = int(*src)
	}
}

// setSeconds sets dst to src seconds when src is set and valid.
func setSeconds(key string, src *int64, dst *time.Duration, errorf errorFunc) {
	if src != nil && checkInt(key, *src, errorf) {
		*dst = time.Duration(*src) * time.Second
	}
}

// setPrefixes sets dst to the prefixes of src when src is set and each entry is valid.
func setPrefixes(key string, src *[]string, dst *[]netip.Prefix, errorf errorFunc) {
	if src != nil {
		if ps, ok := prefixes(key, *src, errorf); ok {
			*dst = ps
		}
	}
}

// prefixes parses the entries of the list key. ok is false after an error.
func prefixes(key string, entries []string, errorf errorFunc) (ps []netip.Prefix, ok bool) {
	ok = true
	ps = make([]netip.Prefix, 0, len(entries))
	for _, s := range entries {
		p, err := parsePrefix(s)
		if err != nil {
			errorf("%s: invalid address or prefix %q", key, s)
			ok = false
			continue
		}
		ps = append(ps, p)
	}
	return ps, ok
}

// checkListen checks <host>:<port>. The host can be empty; listen checks it.
func checkListen(s string) error {
	_, port, err := net.SplitHostPort(s)
	if err != nil {
		return fmt.Errorf("invalid address %q: must be <host>:<port>", s)
	}
	n, err := strconv.ParseUint(port, 10, 16)
	if err != nil || n == 0 {
		return fmt.Errorf("invalid port %q: must be from 1 to 65535", port)
	}
	return nil
}

// parsePrefix parses an address or a prefix.
// An address is a prefix of its full length. An IPv4-mapped IPv6 entry of at least /96
// becomes IPv4, as the address of a client (telnet-term, section 9.2).
func parsePrefix(s string) (netip.Prefix, error) {
	if !strings.Contains(s, "/") {
		a, err := netip.ParseAddr(s)
		if err != nil {
			return netip.Prefix{}, err
		}
		a = a.Unmap().WithZone("")
		return netip.PrefixFrom(a, a.BitLen()), nil
	}
	p, err := netip.ParsePrefix(s)
	if err != nil {
		return netip.Prefix{}, err
	}
	a, bits := p.Addr(), p.Bits()
	if a.Is4In6() && bits >= 96 {
		a, bits = a.Unmap(), bits-96
	}
	return netip.PrefixFrom(a, bits).Masked(), nil
}
