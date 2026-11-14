# Idea: bbs-toml-spec

Backlog. It moves the `bbs.toml` keys out of [bbt-format](../specs/bbt-format.md)
and [telnet-term](../specs/telnet-term.md) into their own specification.

## Goal

One specification defines each key of `bbs.toml`.
The sysop and the setup guide read the configuration in one place.

## Decisions

- The specification holds each key: the tables `[screen]`, `[theme]` and `[fields]`
  of bbt-format, `[menu]` of menu-nav §5 and the `[telnet]` keys of telnet-term §9.1.
- The design specifications refer to it and do not repeat it.

## Open questions

- The place of the checks of `velourbbs check` for `bbs.toml` (screen-render §3.1).
