package config

import (
	"errors"
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

// telnet parses [telnet]. An invalid value keeps its default.
func (p *parser) telnet(v any) {
	t, ok := p.table(v, "telnet")
	if !ok {
		return
	}
	c := &p.cfg.Telnet
	ints := map[string]*int{
		"max_conns":            &c.MaxConns,
		"max_conns_per_addr":   &c.MaxConnsPerAddr,
		"max_connects_per_min": &c.MaxConnectsPerMin,
	}
	timeouts := map[string]*time.Duration{
		"idle_timeout":   &c.IdleTimeout,
		"probe_timeout":  &c.ProbeTimeout,
		"output_timeout": &c.OutputTimeout,
	}
	prefixes := map[string]*[]netip.Prefix{"allow": &c.Allow, "proxy": &c.Proxy}
	for key, v := range t {
		path := []string{"telnet", key}
		var err error
		if dst, ok := ints[key]; ok {
			var n int
			if n, err = parseInt(v); err == nil {
				*dst = n
			}
		} else if dst, ok := timeouts[key]; ok {
			var n int
			if n, err = parseInt(v); err == nil {
				*dst = time.Duration(n) * time.Second
			}
		} else if dst, ok := prefixes[key]; ok {
			var ps []netip.Prefix
			if ps, err = parsePrefixes(v); err == nil {
				*dst = ps
			}
		} else if key == "listen" {
			if err = checkListen(v); err == nil {
				c.Listen = v.(string)
			}
		} else {
			p.errorf(path, "unknown key \"telnet.%s\"", key)
			continue
		}
		if err != nil {
			p.errorf(path, "telnet.%s: %v", key, err)
		}
	}
}

func parseInt(v any) (int, error) {
	n, ok := v.(int64)
	if !ok {
		return 0, errors.New("must be an integer")
	}
	if n < 1 || n > maxInt {
		return 0, fmt.Errorf("%d is not from 1 to %d", n, maxInt)
	}
	return int(n), nil
}

// checkListen checks <host>:<port>. The host can be empty; listen checks it.
func checkListen(v any) error {
	s, ok := v.(string)
	if !ok {
		return errors.New("must be a string")
	}
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

// parsePrefixes parses an array of addresses and prefixes.
// An address is a prefix of its full length. An IPv4-mapped IPv6 entry of at least /96
// becomes IPv4, as the address of a client (section 9.2).
func parsePrefixes(v any) ([]netip.Prefix, error) {
	a, ok := v.([]any)
	if !ok {
		return nil, errors.New("must be an array of strings")
	}
	ps := make([]netip.Prefix, 0, len(a))
	for _, e := range a {
		s, ok := e.(string)
		if !ok {
			return nil, errors.New("must be an array of strings")
		}
		p, err := parsePrefix(s)
		if err != nil {
			return nil, fmt.Errorf("invalid address or prefix %q", s)
		}
		ps = append(ps, p)
	}
	return ps, nil
}

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
