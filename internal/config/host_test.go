package config

import (
	"fmt"
	"net/netip"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

// host parses bbs.toml data over the defaults.
func host(data string) (Config, []string) {
	cfg := Config{Telnet: defaultTelnet, Limits: defaultLimits, Store: defaultStore, Name: defaultName}
	ds := parseHost([]byte(data), &cfg)
	return cfg, messages(ds)
}

func TestHostDefaults(t *testing.T) {
	cfg, ds := host("[telnet]\n[limits]\n")
	if len(ds) != 0 {
		t.Fatalf("diags: %q", ds)
	}
	wantTelnet := Telnet{Listen: ":2323", ProbeTimeout: 5 * time.Second}
	wantLimits := Limits{
		MaxConns: 64, MaxConnsPerAddr: 4, MaxConnectsPerMin: 10,
		IdleTimeout: 600 * time.Second, OutputTimeout: 60 * time.Second,
	}
	if !reflect.DeepEqual(cfg.Telnet, wantTelnet) || !reflect.DeepEqual(cfg.Limits, wantLimits) {
		t.Errorf("Telnet = %+v, Limits = %+v", cfg.Telnet, cfg.Limits)
	}
	if cfg.Store.Path != "velourbbs.jsonl" {
		t.Errorf("Store = %+v, want the path velourbbs.jsonl", cfg.Store)
	}
}

func TestStorePath(t *testing.T) {
	for _, p := range []string{"data/velourbbs.jsonl", "/var/lib/velourbbs/velourbbs.jsonl"} {
		cfg, ds := host("[store]\npath = \"" + p + "\"\n")
		if len(ds) != 0 || cfg.Store.Path != p {
			t.Errorf("path %q: Store = %+v, diags %q", p, cfg.Store, ds)
		}
	}
}

func TestHost(t *testing.T) {
	// The example of bbs-toml, section 4, with other numbers, and the entry forms of allow and proxy.
	data := `
[telnet]
listen = "127.0.0.1:23"
proxy = ["127.0.0.1"]
probe_timeout = 4

[limits]
max_conns = 1
max_conns_per_addr = 2
max_connects_per_min = 1000000
allow = ["192.0.2.10", "2001:db8::/48", "198.51.100.7/24", "::ffff:203.0.113.5", "::ffff:203.0.113.0/120", "fe80::1%eth0"]
idle_timeout = 3
output_timeout = 5
`
	cfg, ds := host(data)
	if len(ds) != 0 {
		t.Fatalf("diags: %q", ds)
	}
	pp := func(s ...string) []netip.Prefix {
		var ps []netip.Prefix
		for _, e := range s {
			ps = append(ps, netip.MustParsePrefix(e))
		}
		return ps
	}
	wantTelnet := Telnet{Listen: "127.0.0.1:23", Proxy: pp("127.0.0.1/32"), ProbeTimeout: 4 * time.Second}
	wantLimits := Limits{
		MaxConns: 1, MaxConnsPerAddr: 2, MaxConnectsPerMin: 1000000,
		Allow:       pp("192.0.2.10/32", "2001:db8::/48", "198.51.100.0/24", "203.0.113.5/32", "203.0.113.0/24", "fe80::1/128"),
		IdleTimeout: 3 * time.Second, OutputTimeout: 5 * time.Second,
	}
	if !reflect.DeepEqual(cfg.Telnet, wantTelnet) {
		t.Errorf("Telnet =\n%+v\nwant\n%+v", cfg.Telnet, wantTelnet)
	}
	if !reflect.DeepEqual(cfg.Limits, wantLimits) {
		t.Errorf("Limits =\n%+v\nwant\n%+v", cfg.Limits, wantLimits)
	}
}

func TestTelnetListen(t *testing.T) {
	for _, s := range []string{":1", ":65535", "[::1]:2323", "localhost:2323"} {
		cfg, ds := host(`telnet.listen = "` + s + `"`)
		if len(ds) != 0 || cfg.Telnet.Listen != s {
			t.Errorf("listen %q: Listen %q, diags %q", s, cfg.Telnet.Listen, ds)
		}
	}
}

func TestHostValueErrors(t *testing.T) {
	tests := []struct {
		data string
		want string
	}{
		{"telnet = 1", `1:10: telnet: must be a table`},
		{"limits = 1", `1:10: limits: must be a table`},
		{"[telnet]\nport = 23", `2:1: unknown key "telnet.port"`},
		{"[telnet]\nmax_conns = 1", `2:1: unknown key "telnet.max_conns"`},
		{"[limits]\nlisten = \":1\"", `2:1: unknown key "limits.listen"`},
		{"[telnet]\nlisten = 2323", `2:10: telnet.listen: must be a string`},
		{"[telnet]\nlisten = \"2323\"", `2:10: telnet.listen: invalid address "2323": must be <host>:<port>`},
		{"[telnet]\nlisten = \":0\"", `2:10: telnet.listen: invalid port "0": must be from 1 to 65535`},
		{"[telnet]\nlisten = \":65536\"", `2:10: telnet.listen: invalid port "65536": must be from 1 to 65535`},
		{"[telnet]\nlisten = \":telnet\"", `2:10: telnet.listen: invalid port "telnet": must be from 1 to 65535`},
		{"[telnet]\nlisten = \"::1:23\"", `2:10: telnet.listen: invalid address "::1:23": must be <host>:<port>`},
		{"[limits]\nallow = \"192.0.2.1\"", `2:9: limits.allow: must be an array of strings`},
		{"[limits]\nallow = [1]", `2:10: limits.allow: must be an array of strings`},
		{"[limits]\nallow = [\"192.0.2.300\"]", `2:9: limits.allow: invalid address or prefix "192.0.2.300"`},
		{"[telnet]\nproxy = [\"192.0.2.0/33\"]", `2:9: telnet.proxy: invalid address or prefix "192.0.2.0/33"`},
		{"[telnet]\nproxy = [\"host\"]", `2:9: telnet.proxy: invalid address or prefix "host"`},
		{"[store]\npath = \"\"", `2:8: store.path: must not be empty`},
		{"[store]\npath = 1", `2:8: store.path: must be a string`},
		{"[store]\nPath = \"x\"", `2:1: unknown key "store.Path"`},
	}
	keys := []string{"telnet.probe_timeout", "limits.max_conns", "limits.max_conns_per_addr",
		"limits.max_connects_per_min", "limits.idle_timeout", "limits.output_timeout"}
	for _, full := range keys {
		table, key, _ := strings.Cut(full, ".")
		// Each error has the position of the value.
		col := fmt.Sprintf("2:%d: ", len(key)+4)
		head := "[" + table + "]\n" + key
		tests = append(tests,
			struct{ data, want string }{head + " = 0", col + full + ": 0 is not from 1 to 1000000"},
			struct{ data, want string }{head + " = 1000001", col + full + ": 1000001 is not from 1 to 1000000"},
			struct{ data, want string }{head + " = 1.5", col + full + ": must be an integer"},
			struct{ data, want string }{head + " = \"5\"", col + full + ": must be an integer"},
		)
	}
	for _, tt := range tests {
		cfg, ds := host(tt.data)
		want := "bbs.toml:" + tt.want
		if !slices.Equal(ds, []string{want}) {
			t.Errorf("parseHost(%q) = %q, want %q", tt.data, ds, want)
		}
		// An invalid value keeps the default.
		if !reflect.DeepEqual(cfg.Telnet, defaultTelnet) || !reflect.DeepEqual(cfg.Limits, defaultLimits) || cfg.Store != defaultStore {
			t.Errorf("parseHost(%q): Telnet = %+v, Limits = %+v, Store = %+v", tt.data, cfg.Telnet, cfg.Limits, cfg.Store)
		}
	}
}
