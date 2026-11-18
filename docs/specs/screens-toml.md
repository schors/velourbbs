# screens.toml

This specification defines the screens settings file `screens/screens.toml`:
each table and each key.
Related parts: [bbt-format](bbt-format.md), [menu-nav](menu-nav.md),
[telnet-term](telnet-term.md), [screen-render](screen-render.md), [bbs-toml](bbs-toml.md),
[file-view](file-view.md) and [session-fields](session-fields.md).

The [BBT guide](bbt-guide.md) teaches this specification. Keep the guide in sync:
[bbt-guide](bbt-guide.md), section 3.

## 1. Goal

The author of the screens reads each key of `screens/screens.toml` in this specification.
The settings go with the screens: a copy of `screens/` keeps its look and its first screen.
The other specifications define the effect of a key and refer to this specification.

## 2. File

- The file is `screens/screens.toml` in the BBS directory ([bbs-toml](bbs-toml.md), section 2).
- The file is optional. A missing file gives the defaults.
  A path that exists and is not a regular file is an error.
- The rules of `bbs.toml` for the format, the size, the defaults, the unknown keys
  and the form of the errors apply: [bbs-toml](bbs-toml.md), section 2.
- A value that breaks a rule of section 3 is an error.

## 3. Tables

### 3.1. `[screen]`

| Key | Type | Default | Meaning |
|---|---|---|---|
| `width` | array of 2 integers | `[40, 80]` | The width limits `[min, max]` of a screen without `width` in its front matter: [bbt-format](bbt-format.md), section 6.1. |
| `more` | string | `"-- More --"` | The text of the More row: [bbt-format](bbt-format.md), section 6.4. |
| `align` | string | `"center"` | The horizontal position of a screen: [bbt-format](bbt-format.md), section 6.1. |
| `valign` | string | `"top"` | The vertical position of a screen: [bbt-format](bbt-format.md), section 6.5. |
| `fill` | string | `"none"` | The fill rows: [bbt-format](bbt-format.md), section 6.5. |
| `time` | string | `"2006-01-02 15:04:05"` | The layout of the Go package `time` for the field `bbs.time`: [session-fields](session-fields.md), section 3.5. |
| `zone` | string | `"Local"` | The time zone of the field `bbs.time`: [session-fields](session-fields.md), section 3.5. |
| `theme` | string | `"dark"` | The theme choice at the start of a session: `dark`, `light` or `auto` (section 3.2). |

- `width`, `align`, `valign` and `fill` have the rules of the front matter keys
  ([bbt-format](bbt-format.md), section 3.2).
- `zone` is `"Local"`, `"UTC"` or a zone name that the Go function `time.LoadLocation` loads.
  Example: `"Europe/Berlin"`. The function reads the zone files of the host, else the IANA
  time zone database in the binary. So each name of that database works on each host.
  A name only in the zone files of a host, for example `"localtime"`, works only on that host.
  Another value is an error.
- `more` cannot have a control character, a grapheme of width 0 or a grapheme
  of more than 32 bytes ([bbt-format](bbt-format.md), section 2.1).

### 3.2. `[theme]`

The theme maps a style name to a style in the syntax of [bbt-format](bbt-format.md),
section 7.2. The style tokens use the theme: [bbt-format](bbt-format.md), section 7.3.

- `[theme]` has only the tables `default`, `16`, `256` and `light`.
  `[theme.16]` and `[theme.256]` are the "depth themes". A depth theme gives other styles
  for one colour depth `term.colors` ([telnet-term](telnet-term.md), section 6).
- `[theme.light]` has only the tables `default`, `16` and `256`: the "light tables".
  They give other styles in the light theme (below).
- A style name matches `[a-z][a-z0-9_]*`. Its value is a string.
- The "base theme" is `normal = "7/0"`, `more = "15/0"`, the `md_*` styles of
  [file-view](file-view.md), section 4, and the styles of `[theme.default]`.
  A style of `[theme.default]` replaces the style with the same name.
  A "name of the theme" is a name of the base theme.
  The effect of `normal` and `more`: [bbt-format](bbt-format.md), sections 6.2, 6.4 and 7.3.
- Each name of a depth theme and of a light table must be a name of the theme.
- The "dark theme" of one `term.colors` is the base theme with each style of the depth theme
  of `term.colors` in place of the style with the same name.
  When `term.colors` has no depth theme, the dark theme is the base theme.
- The "light theme" of one `term.colors` is the dark theme with each style
  of `[theme.light.default]`, then each style of the light table of `term.colors`,
  in place of the style with the same name.
  When `term.colors` has no light table, only `[theme.light.default]` applies.
- The "theme choice" of a session is `dark`, `light` or `auto`. It starts as `[screen] theme`.
  The action `theme` changes it ([menu-nav](menu-nav.md), section 4).
  The field `session.theme` gives it ([session-fields](session-fields.md), section 2).
- The "session theme" is the light theme of `term.colors` when the theme choice is `light`,
  or when it is `auto` and `term.background` is `light` ([telnet-term](telnet-term.md), section 6).
  Else it is the dark theme of `term.colors`.
- A style of the session theme without a foreground or a background takes it
  from `normal` of the session theme. `normal` takes it from `7/0` in the dark theme
  and from `0/15` in the light theme.

Example:

```toml
[theme.default]
normal = "#c0c0c0/#101020"
title  = "#60a0d0 bold"

[theme.16]
normal = "7/0"
title  = "15 bold"

[theme.light.default]
normal = "#202020/#f0f0e8"
title  = "#004080 bold"

[theme.light.16]
normal = "0/7"
```

With 16 colours, `title` is `15/0 bold` in the dark theme. With 256 colours,
it is `#60a0d0/#101020 bold`, and the host reduces it ([telnet-term](telnet-term.md), section 7.3).
In the light theme with 16 colours, `title` is `#004080/7 bold`, and the host reduces
the foreground to an index from 0 to 15.

### 3.3. `[fields]`

`[fields]` maps an alias to a registry path: [bbt-format](bbt-format.md), sections 4.1 and 4.2.

- An alias matches `[a-z][a-z0-9_]*`. Each value is a string. The default `[fields]` is empty.
- An alias cannot be `term`, `item`, `arg`, `file`, `meta`, `session`, `bbs`, `and`, `or` or `not`.
- A field token with an alias is a field token with its path. So an alias to `arg.<name>`
  gives an argument of the screen ([bbt-format](bbt-format.md), section 8.5), and an alias to a field
  of a file screen is valid only in a file screen ([file-view](file-view.md), section 3.4).
- The path is a path of the registry with a dot. It cannot be `item.<KEY>`.

### 3.4. `[menu]`

| Key | Type | Default | Meaning |
|---|---|---|---|
| `start` | string | `"main"` | The name of the first screen of a session: [menu-nav](menu-nav.md), section 5. |

- `start` matches the screen name of [bbt-format](bbt-format.md), section 2.
- [menu-nav](menu-nav.md), section 6, has the checks of the start screen.

## 4. Example

```toml
[screen]
width = [40, 100]
more = "-- More --"
valign = "middle"
fill = "spaces"
theme = "auto"

[theme.default]
normal = "7/0"
title  = "15/4 bold"
prompt = "14"

[theme.16]
title = "15/1 bold"

[theme.light.default]
normal = "0/15"
title  = "15/4 bold"
prompt = "4"

[fields]
w = "term.cols"

[menu]
start = "welcome"
```

## 5. Code

| Package | Content |
|---|---|
| `internal/config` | The parse of `screens/screens.toml`, the checks of section 3, and the dark theme and the light theme of each `term.colors`. |
