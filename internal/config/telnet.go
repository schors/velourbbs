package config

import (
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"time"
)

// Telnet is the table [telnet] (telnet-term, section 9.1).
type Telnet struct {
	// Listen is the TCP address of the server: <host>:<port>.
	Listen string
	// MaxConns is the maximum number of connections.
	MaxConns int
	// MaxConnsPerAddr is the maximum number of connections from one address.
	MaxConnsPerAddr int
	// MaxConnectsPerMin is the maximum number of connects from one address in 60 s.
	MaxConnectsPerMin int
	// Allow is the prefixes without the address limits.
	Allow []netip.Prefix
	// Proxy is the prefixes of the PROXY protocol senders.
	Proxy []netip.Prefix
	// IdleTimeout, ProbeTimeout and OutputTimeout are the timeouts of the session.
	IdleTimeout   time.Duration
	ProbeTimeout  time.Duration
	OutputTimeout time.Duration
}

// The defaults of section 9.1.
var defaultTelnet = Telnet{
	Listen:            ":2323",
	MaxConns:          64,
	MaxConnsPerAddr:   4,
	MaxConnectsPerMin: 10,
	IdleTimeout:       600 * time.Second,
	ProbeTimeout:      5 * time.Second,
	OutputTimeout:     60 * time.Second,
}

// maxInt is the largest integer of [telnet]. It keeps a timeout in a time.Duration.
const maxInt = 1000000

// telnetDoc is the plain decode of [telnet]. A nil pointer is a key that is not set.
type telnetDoc struct {
	Listen            *string   `toml:"listen"`
	MaxConns          *int64    `toml:"max_conns"`
	MaxConnsPerAddr   *int64    `toml:"max_conns_per_addr"`
	MaxConnectsPerMin *int64    `toml:"max_connects_per_min"`
	Allow             *[]string `toml:"allow"`
	Proxy             *[]string `toml:"proxy"`
	IdleTimeout       *int64    `toml:"idle_timeout"`
	ProbeTimeout      *int64    `toml:"probe_timeout"`
	OutputTimeout     *int64    `toml:"output_timeout"`
}

// telnetValues is the decode of the values of [telnet] with a check.
type telnetValues struct {
	Listen address  `toml:"listen"`
	Allow  []prefix `toml:"allow"`
	Proxy  []prefix `toml:"proxy"`
}

type address struct{}

func (address) UnmarshalText(text []byte) error { return checkListen(string(text)) }

type prefix struct{ p netip.Prefix }

func (p *prefix) UnmarshalText(text []byte) error {
	var err error
	if p.p, err = parsePrefix(string(text)); err != nil {
		return fmt.Errorf("invalid address or prefix %q", text)
	}
	return nil
}

// telnet gives the settings of [telnet]. An integer out of its range keeps its default.
func telnet(d telnetDoc, v telnetValues, errorf func(format string, args ...any)) Telnet {
	c := defaultTelnet
	if d.Listen != nil {
		c.Listen = *d.Listen
	}
	if d.Allow != nil {
		c.Allow = prefixes(v.Allow)
	}
	if d.Proxy != nil {
		c.Proxy = prefixes(v.Proxy)
	}
	ints := []struct {
		key string
		src *int64
		dst *int
	}{
		{"max_conns", d.MaxConns, &c.MaxConns},
		{"max_conns_per_addr", d.MaxConnsPerAddr, &c.MaxConnsPerAddr},
		{"max_connects_per_min", d.MaxConnectsPerMin, &c.MaxConnectsPerMin},
	}
	timeouts := []struct {
		key string
		src *int64
		dst *time.Duration
	}{
		{"idle_timeout", d.IdleTimeout, &c.IdleTimeout},
		{"probe_timeout", d.ProbeTimeout, &c.ProbeTimeout},
		{"output_timeout", d.OutputTimeout, &c.OutputTimeout},
	}
	check := func(key string, n int64) bool {
		if n < 1 || n > maxInt {
			errorf("telnet.%s: %d is not from 1 to %d", key, n, maxInt)
			return false
		}
		return true
	}
	for _, e := range ints {
		if e.src != nil && check(e.key, *e.src) {
			*e.dst = int(*e.src)
		}
	}
	for _, e := range timeouts {
		if e.src != nil && check(e.key, *e.src) {
			*e.dst = time.Duration(*e.src) * time.Second
		}
	}
	return c
}

func prefixes(a []prefix) []netip.Prefix {
	ps := make([]netip.Prefix, len(a))
	for i, p := range a {
		ps[i] = p.p
	}
	return ps
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
// becomes IPv4, as the address of a client (section 9.2).
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
