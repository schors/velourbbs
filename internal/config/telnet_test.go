package config

import (
	"fmt"
	"net/netip"
	"reflect"
	"slices"
	"testing"
	"time"
)

func TestTelnetDefaults(t *testing.T) {
	tn, _, ds := parseHost([]byte("[telnet]\n"))
	if len(ds) != 0 {
		t.Fatalf("diags: %q", messages(ds))
	}
	want := Telnet{
		Listen: ":2323", MaxConns: 64, MaxConnsPerAddr: 4, MaxConnectsPerMin: 10,
		IdleTimeout: 600 * time.Second, ProbeTimeout: 5 * time.Second, OutputTimeout: 60 * time.Second,
	}
	if !reflect.DeepEqual(tn, want) {
		t.Errorf("Telnet = %+v, want %+v", tn, want)
	}
}

func TestTelnet(t *testing.T) {
	// The [telnet] example of bbs-toml, section 4, with other numbers, and the entry forms of allow and proxy.
	data := `
[telnet]
listen = "127.0.0.1:23"
max_conns = 1
max_conns_per_addr = 2
max_connects_per_min = 1000000
allow = ["192.0.2.10", "2001:db8::/48", "198.51.100.7/24", "::ffff:203.0.113.5", "::ffff:203.0.113.0/120", "fe80::1%eth0"]
proxy = ["127.0.0.1"]
idle_timeout = 3
probe_timeout = 4
output_timeout = 5
`
	tn, _, ds := parseHost([]byte(data))
	if len(ds) != 0 {
		t.Fatalf("diags: %q", messages(ds))
	}
	pp := func(s ...string) []netip.Prefix {
		var ps []netip.Prefix
		for _, e := range s {
			ps = append(ps, netip.MustParsePrefix(e))
		}
		return ps
	}
	want := Telnet{
		Listen: "127.0.0.1:23", MaxConns: 1, MaxConnsPerAddr: 2, MaxConnectsPerMin: 1000000,
		Allow:       pp("192.0.2.10/32", "2001:db8::/48", "198.51.100.0/24", "203.0.113.5/32", "203.0.113.0/24", "fe80::1/128"),
		Proxy:       pp("127.0.0.1/32"),
		IdleTimeout: 3 * time.Second, ProbeTimeout: 4 * time.Second, OutputTimeout: 5 * time.Second,
	}
	if !reflect.DeepEqual(tn, want) {
		t.Errorf("Telnet =\n%+v\nwant\n%+v", tn, want)
	}
}

func TestTelnetListen(t *testing.T) {
	for _, s := range []string{":1", ":65535", "[::1]:2323", "localhost:2323"} {
		tn, _, ds := parseHost([]byte(`telnet.listen = "` + s + `"`))
		if len(ds) != 0 || tn.Listen != s {
			t.Errorf("listen %q: Listen %q, diags %q", s, tn.Listen, messages(ds))
		}
	}
}

func TestTelnetErrors(t *testing.T) {
	tests := []struct {
		data string
		want string
	}{
		{"telnet = 1", `1:10: telnet: must be a table`},
		{"[telnet]\nport = 23", `2:1: unknown key "telnet.port"`},
		{"[telnet]\nlisten = 2323", `2:10: telnet.listen: must be a string`},
		{"[telnet]\nlisten = \"2323\"", `1:1: telnet.listen: invalid address "2323": must be <host>:<port>`},
		{"[telnet]\nlisten = \":0\"", `1:1: telnet.listen: invalid port "0": must be from 1 to 65535`},
		{"[telnet]\nlisten = \":65536\"", `1:1: telnet.listen: invalid port "65536": must be from 1 to 65535`},
		{"[telnet]\nlisten = \":telnet\"", `1:1: telnet.listen: invalid port "telnet": must be from 1 to 65535`},
		{"[telnet]\nlisten = \"::1:23\"", `1:1: telnet.listen: invalid address "::1:23": must be <host>:<port>`},
		{"[telnet]\nallow = \"192.0.2.1\"", `2:9: telnet.allow: must be an array of strings`},
		{"[telnet]\nallow = [1]", `2:10: telnet.allow: must be an array of strings`},
		{"[telnet]\nallow = [\"192.0.2.300\"]", `1:1: telnet.allow: invalid address or prefix "192.0.2.300"`},
		{"[telnet]\nproxy = [\"192.0.2.0/33\"]", `1:1: telnet.proxy: invalid address or prefix "192.0.2.0/33"`},
		{"[telnet]\nproxy = [\"host\"]", `1:1: telnet.proxy: invalid address or prefix "host"`},
	}
	for _, key := range []string{"max_conns", "max_conns_per_addr", "max_connects_per_min",
		"idle_timeout", "probe_timeout", "output_timeout"} {
		// An integer out of its range is an other error at 1:1; a type error has the position of the value.
		col := fmt.Sprintf("2:%d: ", len(key)+4)
		tests = append(tests,
			struct{ data, want string }{"[telnet]\n" + key + " = 0", "1:1: telnet." + key + ": 0 is not from 1 to 1000000"},
			struct{ data, want string }{"[telnet]\n" + key + " = 1000001", "1:1: telnet." + key + ": 1000001 is not from 1 to 1000000"},
			struct{ data, want string }{"[telnet]\n" + key + " = 1.5", col + "telnet." + key + ": must be an integer"},
			struct{ data, want string }{"[telnet]\n" + key + " = \"5\"", col + "telnet." + key + ": must be an integer"},
		)
	}
	for _, tt := range tests {
		tn, _, ds := parseHost([]byte(tt.data))
		want := "bbs.toml:" + tt.want
		if got := messages(ds); !slices.Equal(got, []string{want}) {
			t.Errorf("parseHost(%q) = %q, want %q", tt.data, got, want)
		}
		// An invalid value keeps the default.
		if !reflect.DeepEqual(tn, defaultTelnet) {
			t.Errorf("parseHost(%q): Telnet = %+v", tt.data, tn)
		}
	}
}
