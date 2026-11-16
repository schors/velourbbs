# bbs.toml

This specification defines the host settings file `bbs.toml`: each table and each key.
Related parts: [screens-toml](screens-toml.md), [telnet-term](telnet-term.md),
[screen-render](screen-render.md) and [session-fields](session-fields.md).

## 1. Goal

The sysop reads each key of `bbs.toml` in this specification.
The other specifications define the effect of a key and refer to this specification.

## 2. File

- The BBS directory is the current directory of `velourbbs`. It holds `bbs.toml` and `screens/`.
- The settings of the screens are in `screens/screens.toml`: [screens-toml](screens-toml.md).
- `bbs.toml` must exist and be a regular file.
- The file is TOML 1.1 in UTF-8. The loader removes a BOM.
- The file is at most 64 KB without the BOM.
- Each table and each key is optional. A key that is not set has its default.
  An empty file gives the defaults.
- An unknown table or key is an error.
- A value that breaks a rule of section 3 is an error.
- [screen-render](screen-render.md), section 3.1, gives the form of the errors.

## 3. Tables

### 3.1. `[telnet]`

| Key | Type | Default | Meaning |
|---|---|---|---|
| `listen` | string | `":2323"` | The TCP address of the server: `<host>:<port>`. [telnet-term](telnet-term.md), section 10. |
| `max_conns` | integer | 64 | The maximum number of connections. [telnet-term](telnet-term.md), section 9.2. |
| `max_conns_per_addr` | integer | 4 | The maximum number of connections from one address. [telnet-term](telnet-term.md), section 9.2. |
| `max_connects_per_min` | integer | 10 | The maximum number of connects from one address in 60 s. [telnet-term](telnet-term.md), section 9.2. |
| `allow` | array of strings | `[]` | Addresses and prefixes without the address limits. [telnet-term](telnet-term.md), section 9.2. |
| `proxy` | array of strings | `[]` | Addresses and prefixes of the PROXY protocol senders. [telnet-term](telnet-term.md), section 9.3. |
| `idle_timeout` | integer | 600 | Seconds without a key until the host disconnects. [telnet-term](telnet-term.md), section 8.3. |
| `probe_timeout` | integer | 5 | Seconds from accept until the probe ends. [telnet-term](telnet-term.md), section 5.1. |
| `output_timeout` | integer | 60 | Seconds until a write must complete. Else the host disconnects. [telnet-term](telnet-term.md), sections 4.3 and 9.4. |

- Each integer is from 1 to 1000000. The port of `listen` is from 1 to 65535.
- The host of `listen` is empty, an IP address or a host name. `velourbbs check` checks
  only the form `<host>:<port>` and the port. `velourbbs serve` checks the host when it listens
  ([telnet-term](telnet-term.md), section 10).
- An address is an IPv4 or IPv6 address. A prefix is an address with `/<bits>`.
- An IPv6 address can have a zone `%<zone>`. The host ignores the zone. A prefix cannot have a zone.

### 3.2. `[bbs]`

| Key | Type | Default | Meaning |
|---|---|---|---|
| `name` | string | `"Kind Brothers BBS"` | The name of the BBS: the field `bbs.name` of [session-fields](session-fields.md), section 2. |

## 4. Example

```toml
[bbs]
name = "Kind Brothers BBS"

[telnet]
listen = ":2323"
max_conns = 64
max_conns_per_addr = 4
max_connects_per_min = 10
allow = ["192.0.2.10", "2001:db8::/48"]
proxy = ["127.0.0.1"]
idle_timeout = 600
probe_timeout = 5
output_timeout = 60
```

## 5. Code

| Package | Content |
|---|---|
| `internal/config` | The parse of `bbs.toml` and the checks of section 3. |
