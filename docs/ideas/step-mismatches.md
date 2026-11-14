# Idea: step-mismatches

Backlog. It fixes small mismatches between the steps of a session and the tools.
It changes [telnet-term](../specs/telnet-term.md), `cmd/bbs`, `internal/term` and the tests.
Source: the second MVP audit (issue `jgqjh-mvp-reaudit`).

## Goal

Each step accepts and gives the same variants. The specs promise only what the code does.

## Items

1. Line mode. A client without SGA sends `S` LF. The `S` runs the item,
   and the LF runs `ANY` or `ENTER` on the next screen.
   telnet-term §4.1 says that such a client works.
   Cut the promise, or drop the line end after a hotkey.
2. Default size. The default is 80x25 (`internal/term/values.go:16`). 25 rows are for SyncTERM,
   and the host refuses SyncTERM (telnet-term §5.5). Without NAWS, a terminal with 24 rows
   loses row 24 under the More row. Make the default 80x24.
3. `bbs preview --set term.type` (`cmd/bbs/preview.go:114`) keeps the name as it is.
   A session cleans the name, makes it lower case and derives `term.colors` from it.
   `VT100` gives `VT100` and 256 colours in the preview, and `vt100` and 16 colours in a session.
   Use one function of `internal/term` in both.
4. telnet-term §7.4 says "No other part changes" for 8-bit charsets. The key decoder,
   the probe and `term.charset` also change. Remove the sentence.
5. Key names have two definitions: `internal/key/key.go` and the string literals
   in `internal/term/key.go`. No test links them. A typo makes a key that never matches.
   Add a test that each decoder key passes `key.Parse`, or use the constants of `key`.
6. `FuzzLoad` does not render. The panic paths in `internal/screen/field.go:229`
   and `internal/screen/cond.go:129` are not fuzzed. Render and drain the pages in the fuzz test.

## Open questions

- Item 1: cut the promise or drop the line end.
