# Session fields

This specification defines the fields of the session and of the BBS: their values
in `velourbbs serve` and in `velourbbs preview`.
Related parts: [bbt-format](bbt-format.md), [bbs-toml](bbs-toml.md),
[screens-toml](screens-toml.md), [screen-render](screen-render.md), [menu-nav](menu-nav.md),
[file-view](file-view.md) and [telnet-term](telnet-term.md).

## 1. Goal

A screen shows the user the data of the session and of the BBS: the client address,
the online time, the BBS name, the version, the open sessions, the server time and the uptime.

## 2. Fields

| Path | Type | Value | Example |
|---|---|---|---|
| `session.ip` | string | The client address (section 3.1). | `203.0.113.7` |
| `session.theme` | string | The theme choice of the session: `dark`, `light` or `auto` ([screens-toml](screens-toml.md), section 3.2). | `auto` |
| `session.lang` | string | The session language: a language code, or empty for no language ([session-language](session-language.md), section 5). | `ru` |
| `session.online` | string | The time from the accept of the connection, in the duration form (section 3.3). | `01:25:07` |
| `bbs.name` | string | `[bbs] name` of [bbs-toml](bbs-toml.md), section 3.2. | `Kind Brothers BBS` |
| `bbs.version` | string | The version of the `velourbbs` binary (section 3.4). | `v1.2.0` |
| `bbs.sessions` | integer | The open sessions, with the current session (section 3.2). | `3` |
| `bbs.time` | string | The time of the server in its time zone (section 3.5), in the layout `[screen] time` of [screens-toml](screens-toml.md), section 3.1. | `2026-11-16 15:04:05` |
| `bbs.uptime` | string | The time from the start of the listen of `velourbbs serve`, in the duration form (section 3.3). | `3d 04:17:09` |

- The host takes the values once for each screen, before the `when` of the items
  ([menu-nav](menu-nav.md), section 5, step 2). The `when` and the render use these values.
- A file screen takes the values at step 1 of [menu-nav](menu-nav.md), section 5.1,
  for the area height ([file-view](file-view.md), section 3.2). It takes them again for each page
  at step 2.

## 3. Values

### 3.1. Client address

- `session.ip` is the client address of [telnet-term](telnet-term.md), section 9.3, without its zone.
  An IPv4-mapped address gives the IPv4 address.
- The form is the text form of the Go package `net/netip`. Examples: `192.0.2.1`, `2001:db8::1`.

### 3.2. Open sessions

- A session is open from the decision of the terminal values
  ([telnet-term](telnet-term.md), section 5.1, step 5), before the render of the first screen.
- It is open until the start of the close of its connection ([telnet-term](telnet-term.md), section 10).

### 3.3. Duration form

- The host rounds a duration down to whole seconds.
- Less than 24 hours gives `hh:mm:ss`. Example: `04:17:09`.
- 24 hours or more gives `<d>d hh:mm:ss`, where `<d>` is the number of whole days in decimal.
  Example: `3d 04:17:09`.
- `hh`, `mm` and `ss` have 2 digits.

### 3.4. Version

- The version comes from the build info of the binary: the field `Main.Version`
  of the Go package `runtime/debug`.
- The value is the version when it is a tag. Example: `v1.2.0`.
- Else the value is `dev`. These versions are not a tag:
  an empty version, `(devel)`, a pseudo-version and a version with build metadata (`+dirty`).
- A pseudo-version ends with `<14 digits>-<12 hexadecimal digits>`.
  Example: `v0.0.0-20261116150408-c31024543e47`.
- The Go command writes a tag only for a build in a git checkout of the tagged commit.
  The tag must be a semantic version `vX.Y.Z` with an optional pre-release, and with
  the major version 0 or 1.
- A changed file, or a file that git does not track and `.gitignore` does not name,
  gives `+dirty`. `go run`, `go test` and `-buildvcs=false` give `(devel)`.

### 3.5. Server time

- The time zone is `[screen] zone` of [screens-toml](screens-toml.md), section 3.1.
- `Local` is `time.Local` of Go: `TZ`, else `/etc/localtime`.

## 4. `velourbbs preview`

`velourbbs preview` has no connection. It takes these values:

| Field | Value |
|---|---|
| `session.ip` | `127.0.0.1` |
| `session.online` | `00:00:00` |
| `bbs.sessions` | `1` |
| `bbs.time` | 2006-01-02 15:04:05 in `[screen] zone`, in the layout `[screen] time`. `Local` gives UTC. |
| `bbs.uptime` | `00:00:00` |
| `session.theme` | `[screen] theme`. |
| `session.lang` | `[screen] lang`, else empty. |

- `--set <path>=<value>` replaces the value ([screen-render](screen-render.md), section 3.2).
- A string field takes the value as it is. The value has the rules of an argument value
  ([bbt-format](bbt-format.md), section 8.5). `bbs.sessions` takes a decimal integer.
  `session.theme` takes `dark`, `light` or `auto`.
  In `--keys` ([menu-nav](menu-nav.md), section 7), the action `theme` changes it.
  `session.lang` takes the rules of [session-language](session-language.md), section 8.
  An invalid value is an invalid argument.
- `bbs.name` and `bbs.version` have the values of section 2.
- The default time has the date and the clock of the reference time of the Go package `time`.
  UTC in place of `Local` makes the dump the same on each machine.

## 5. Code

| Package | Content |
|---|---|
| `internal/screen` | The fields of section 2 in the registry and the duration form. |
| `internal/config` | `[bbs] name`, `[screen] time`, `[screen] zone` and `[screen] theme`. |
| `internal/menu` | The values of section 2 for each screen and each page of a file screen, and the theme choice. |
| `internal/server` | The client address, the open sessions and the start of the listen. |
| `cmd/velourbbs` | The version and the values of section 4. |

- `.gitignore` names `/velourbbs`, so the binary of `go build ./cmd/velourbbs`
  in the checkout does not give `+dirty`.

## 6. Acceptance

- A test covers the duration form: 0 seconds, 59 seconds, 23:59:59, 24:00:00 and several days.
- A test covers `[screen] zone`: `Local`, `UTC`, a name of the database and an unknown name.
  A server test and a test of `velourbbs preview` show `bbs.time` in a zone.
- A test covers the version: a tag, an empty version, `(devel)`, a pseudo-version and `+dirty`.
- A server test shows a screen with each field of section 2 through a PROXY protocol
  header of version 1. `session.ip` is the address of the header.
  `bbs.sessions` counts two open sessions and not a connection in the probe.
- Golden cell dumps of `velourbbs preview` show each field with the values of section 4,
  without `--set` and with `--set` for each field.
- A test covers an invalid `--set` value of a string field, of `bbs.sessions` and of `session.theme`.
- A golden output of `velourbbs preview --keys` shows an item with the action `theme`
  that changes `session.theme` and the theme, in a screen and in a file screen.

## 7. Not in this specification

- The update of a field on an open screen: [live-fields](https://github.com/schors/velourbbs/issues/10).
- A node number.
- The names of months and days in other languages.
