# Session

This specification defines the load of the BBS, the session object, the list of open sessions
and the cancel path of a session. It changes no behaviour for the user.
Related parts: [menu-nav](menu-nav.md), [telnet-term](telnet-term.md)
and [session-fields](session-fields.md).

## 1. Goal

Later features get one place for the state of a session and one way to load the BBS:
who is online ([#81](https://github.com/schors/velourbbs/issues/81)),
the session language ([#83](https://github.com/schors/velourbbs/issues/83))
and the reload of screens ([#14](https://github.com/schors/velourbbs/issues/14)).

## 2. Concepts

- The "model" is the result of a load without errors: the settings of `bbs.toml`
  and `screens/screens.toml`, and the screens. The model does not change after the load.
  Sessions read it at the same time without locks.
- A session is "open" as [session-fields](session-fields.md), section 3.2, defines.
- The "stop" is the stop of [telnet-term](telnet-term.md), section 10.2.

## 3. Load

- One function loads the BBS directory: `bbs.toml`, `screens/screens.toml`, the screens
  and the checks between files ([menu-nav](menu-nav.md), section 6).
  It gives the model and the sorted diagnostics. After an error, it gives no model.
- `velourbbs check`, `velourbbs preview` and `velourbbs serve` use this function.
  Their diagnostics do not change.
- A new part of the BBS directory adds its load and its checks to this function.

## 4. Session object

- The session object owns the state of one session: the client address, the accept time,
  the model, the terminal, the theme choice and the navigation stack
  ([menu-nav](menu-nav.md), section 4.1).
  A new state of a session is a field of the session object.
- The server makes the session object when the session opens, after the probe.
  `velourbbs preview --keys` makes it without a server and with no client address.
- The goroutine of the session changes the state. A field that another goroutine reads
  needs a lock.
- The session object runs the session of [menu-nav](menu-nav.md), section 5, until `logoff`
  or an error. After the stop, it shows the screen `shutdown` with its theme choice
  ([menu-nav](menu-nav.md), section 5.2).

## 5. Open sessions

- The server keeps the list of open sessions. A session is in the list while it is open.
- `bbs.sessions` ([session-fields](session-fields.md), section 2) is the length of the list.

## 6. Cancel path

- The server has one context for its sessions. The stop cancels it.
- The probe ([telnet-term](telnet-term.md), section 5) and the session take this context.
  When the context ends:
  - a probe ends at once. The server closes the connection with no output;
  - a wait for a key ends. A page that the host writes completes.
    Then the session shows `shutdown`.
- Each later wait of a session takes this context, for example a line input
  ([line-input](line-input.md)).
- A connection before the probe, for example in the read of a PROXY protocol header
  ([telnet-term](telnet-term.md), section 9.3), has no context. The stop closes it.

## 7. Code

| Package | Content |
|---|---|
| `internal/bbs` | The load of section 3, the model and the checks of [menu-nav](menu-nav.md), section 6. |
| `internal/menu` | The session object of section 4 with its run and its `shutdown`. |
| `internal/term` | The probe and the wait for a key with the context of section 6. |
| `internal/server` | The list of section 5 and the context of section 6. |

## 8. Acceptance

- The tests of [menu-nav](menu-nav.md), [telnet-term](telnet-term.md)
  and [session-fields](session-fields.md) pass. Their expected output does not change.
- A test covers the context in the probe and in a wait for a key: each ends with the stop error.
- A server test shows `bbs.sessions` from the list of open sessions:
  two open sessions, and not a connection in the probe.
- `go vet ./...`, `go test ./...` and `golangci-lint run ./...` pass.

## 9. Not in this specification

- The change of a session to a new model: [reload](https://github.com/schors/velourbbs/issues/14).
- Who is online and the last callers: [#81](https://github.com/schors/velourbbs/issues/81).
- The session language: [#83](https://github.com/schors/velourbbs/issues/83).
- The data store: [#80](https://github.com/schors/velourbbs/issues/80).
- The end of one session by the sysop.
