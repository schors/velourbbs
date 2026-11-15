# bbs.toml

This specification defines the host settings file `bbs.toml`: each table and each key.
Related parts: [bbt-format](bbt-format.md), [menu-nav](menu-nav.md),
[telnet-term](telnet-term.md) and [screen-render](screen-render.md).

## 1. Goal

The sysop reads each key of `bbs.toml` in this specification.
The other specifications define the effect of a key and refer to this specification.

## 2. File

- The BBS directory is the current directory of `velourbbs`. It holds `bbs.toml` and `screens/`.
- `bbs.toml` must exist and be a regular file.
- The file is TOML 1.1 in UTF-8. The loader removes a BOM.
- The file is at most 64 KB without the BOM.
- Each table and each key is optional. A key that is not set has its default.
  An empty file gives the defaults.
- An unknown table or key is an error.
- A value that breaks a rule of section 3 is an error.
- [screen-render](screen-render.md), section 3.1, gives the form of the errors.

## 3. Tables

### 3.1. `[screen]`

| Key | Type | Default | Meaning |
|---|---|---|---|
| `width` | array of 2 integers | `[40, 80]` | The width limits `[min, max]` of a screen without `width` in its front matter: [bbt-format](bbt-format.md), section 6.1. |
| `more` | string | `"-- More --"` | The text of the More row: [bbt-format](bbt-format.md), section 6.4. |

- `width` has the rules of `width` in the front matter ([bbt-format](bbt-format.md), section 3.2).
- `more` cannot have a control character, a grapheme of width 0 or a grapheme
  of more than 32 bytes ([bbt-format](bbt-format.md), section 2.1).

### 3.2. `[theme]`

`[theme]` maps a style name to a style in the syntax of [bbt-format](bbt-format.md),
section 7.2. The style tokens use the theme: [bbt-format](bbt-format.md), section 7.3.

- A name matches `[a-z][a-z0-9_]*`. Each value is a string.
- The theme always has `normal` and `more`. Their defaults are `7/0` and `15/0`.
- A theme style without a foreground or a background takes it from `normal`.
  `normal` takes it from `7/0`.
- The effect of `normal` and `more`: [bbt-format](bbt-format.md), sections 6.2, 6.4 and 7.3.

### 3.3. `[fields]`

`[fields]` maps an alias to a registry path: [bbt-format](bbt-format.md), sections 4.1 and 4.2.

- An alias matches `[a-z][a-z0-9_]*`. Each value is a string. The default `[fields]` is empty.
- An alias cannot be `term`, `item`, `and`, `or` or `not`.
- The path is a path of the registry with a dot. It cannot be `item.<KEY>`.

### 3.4. `[menu]`

| Key | Type | Default | Meaning |
|---|---|---|---|
| `start` | string | `"main"` | The name of the first screen of a session: [menu-nav](menu-nav.md), section 5. |

- `start` matches the screen name of [bbt-format](bbt-format.md), section 2.
- [menu-nav](menu-nav.md), section 6, has the checks of the start screen.

### 3.5. `[telnet]`

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

## 4. Example

```toml
[screen]
width = [40, 100]
more = "-- More --"

[theme]
normal = "7/0"
title  = "15/4 bold"
prompt = "14"

[fields]
w = "term.cols"

[menu]
start = "welcome"

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
