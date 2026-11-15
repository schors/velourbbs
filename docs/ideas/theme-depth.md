# Idea: theme per colour depth

Backlog. It extends [bbs-toml](../specs/bbs-toml.md) and [bbt-format](../specs/bbt-format.md).
[example-style](example-style.md) needs it.

## Goal

One template has the look that its author selects in each colour depth.
The sysop sets other styles for 16 colours and for 256 colours.

## Decisions

- `bbs.toml` gets a theme for a colour depth beside `[theme]`, for example `[theme.16]`.
- A style of a depth theme replaces the style with the same name of `[theme]`
  when `term.colors` is this depth.
- Screens do not change. A style token names the style as before.
- A style that the depth theme does not have comes from `[theme]`.
  The host reduces its colours ([telnet-term](../specs/telnet-term.md), section 7.3).

Example:

```toml
[theme]
title = "#60a0d0 bold"

[theme.16]
title = "15 bold"
```

## Open questions

1. The syntax and the set of depths.
2. A name in a depth theme that `[theme]` does not have: an error or not.
