# 3. `bbs.toml`

`bbs.toml` holds four tables: `[bbs]` has the name of the BBS,
`[telnet]` has the settings of the telnet server, `[limits]` has the connection limits and the timeouts,
and `[store]` has the data file.
Each key is optional. A key that is not set has its default. An empty file gives the defaults.

## The name

`name` of `[bbs]` is the name of the BBS. A screen shows it in the field `bbs.name`.
The default is `"Kind Brothers BBS"`.

## The address

`listen` of `[telnet]` is the TCP address of the server: `<host>:<port>`. The default is `":2323"`.

- An empty host: the server listens on all addresses of the host: `":2323"`.
- One address: the server listens only on this address, for example `"192.0.2.1:2323"`.
  Use it when the host has more than one address and the BBS must use only one.

`velourbbs check` checks only the form of `listen` and the port.
`velourbbs serve` checks the address when it starts. If the host does not have
the address, `velourbbs serve` stops.

## The connection limits

These keys are in `[limits]`.

| Key | Default | What it does | When to change it |
|---|---|---|---|
| `max_conns` | 64 | The maximum number of connections. | More users connect at the same time. Then also change the limits of section 5. |
| `max_conns_per_addr` | 4 | The maximum number of connections from one address. | Many users share one address, for example behind NAT. |
| `max_connects_per_min` | 10 | The maximum number of connects from one address in 60 s. | A client connects again many times. |
| `allow` | `[]` | Addresses and prefixes without the address limits. | One address must have more connections, for example a known NAT. |

- "One address" is one IPv4 address, or one IPv6 /64 prefix.
- `max_conns` applies also to the addresses in `allow`.
- A connection over a limit gets the line `Too many connections. Try again later.`
  Then the server closes it.

Example:

```toml
[limits]
allow = ["192.0.2.10", "2001:db8::/48"]
```

## The PROXY protocol

`proxy` of `[telnet]` holds the addresses and prefixes of the proxies that send the PROXY protocol header.
The default is `[]`: the server reads no header.
Set it when a proxy, for example HAProxy, sends the connections to the BBS.
Then the log and the limits use the address of the client, not the address of the proxy.

- The server reads the header only from the addresses in `proxy`.
- A connection from an address in `proxy` without a valid header closes without output.
  So, if `proxy` holds `127.0.0.1`, a telnet client on the same host cannot connect
  to `127.0.0.1`.
- The proxy must send the header at the connect. AWS NLB waits for the data of the client
  by default. Then it does not work.

Example:

```toml
[telnet]
proxy = ["127.0.0.1"]
```

## The timeouts

`probe_timeout` is in `[telnet]`. `idle_timeout` and `output_timeout` are in `[limits]`.

| Key | Default | What it does | When to change it |
|---|---|---|---|
| `idle_timeout` | 600 | Seconds without a key until the server disconnects. | Users read long screens. |
| `probe_timeout` | 5 | Seconds from the connect until the server knows the terminal. | Clients have a slow link. |
| `output_timeout` | 60 | Seconds until a write must complete. Else the server disconnects. | Clients have a slow link. |

## The data file

`path` of `[store]` is the data file. The BBS keeps its data in this file between sessions.
The default is `"velourbbs.jsonl"`. A relative path is relative to the BBS directory.

Set it to `"/var/lib/velourbbs/velourbbs.jsonl"`. The user `velourbbs` can write
only to this directory (section 5). `velourbbs serve` creates the file.
The directory must exist (section 2).

## An example file

This file has the defaults, except `path`. It sets each key except `allow` and `proxy`:

```toml
[bbs]
name = "Kind Brothers BBS"

[telnet]
listen = ":2323"
probe_timeout = 5

[limits]
max_conns = 64
max_conns_per_addr = 4
max_connects_per_min = 10
idle_timeout = 600
output_timeout = 60

[store]
path = "/var/lib/velourbbs/velourbbs.jsonl"
```

Write it to `/srv/velourbbs/bbs.toml`. Then change the values that you need.
After each change, run `sudo chmod -R a+rX /srv/velourbbs` and `sudo -u velourbbs velourbbs check`.

Full rules: [bbs-toml](../../../specs/bbs-toml.md), section 3,
and [data-store](../../../specs/data-store.md).

## Result

`sudo -u velourbbs velourbbs check` gives no output.
