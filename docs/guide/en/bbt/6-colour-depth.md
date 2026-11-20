# Task 6. The colour depth

## The depth of a terminal

A terminal shows 16, 256 or 16777216 colours: its "colour depth".
At the connect, the BBS asks the terminal and reads the terminal type that the client sends.
The first row of the table that matches gives the depth:

| Terminal | Depth |
|---|---|
| Shows an RGB colour, for example Windows Terminal, iTerm2, Ghostty, xterm with direct colour | 16777216 |
| PuTTY or tmux | 16777216 |
| Type ends with `-direct`, or has `truecolor` or `24bit` | 16777216 |
| Type has `256color` | 256 |
| Type ends with `-16color` or `-8color`, or is `ansi`, `dumb`, `vt52`, `vt100`, `vt102`, `vt220`, `vt320`, `rxvt`, `xterm-color`, `eterm-color` or `cygwin` | 16 |
| Each other type, for example `xterm` | 256 |

- A client without a type gets `ansi`. If its terminal does not answer, it gets 16 colours.
- Tera Term sends `xterm` and shows no RGB colour, so 256 colours.

The BBS changes each colour to the nearest colour of the depth.
At 256 colours, the cream text `#e9dccb` becomes the neutral grey 253, and the screen loses its warmth.
So choose the colours for each depth by hand.
The indexes 16–255 are the colours of the xterm 256-colour palette.
The indexes 0–15 are the VGA colours of [telnet-term](../../../specs/telnet-term.md), section 7.3.

Full rules: [telnet-term](../../../specs/telnet-term.md), sections 6 and 7.3.

## Depth themes

```toml
[theme.256]
normal = "223/235"
title  = "230 bold"

[theme.16]
normal = "7/0"
title  = "15 bold"
```

- A "depth theme" gives other styles for one depth.
  `[theme.256]` is for 256 colours. `[theme.16]` is for 16 colours.
- A style of a depth theme replaces the style with the same name of `[theme.default]`.
  The other styles stay.
- Each name of a depth theme must also be in `[theme.default]`, or be `normal`, `more` or an `md_*` style (task 7).
- At 16 colours, use the indexes 0–15. At 256 colours, use the indexes 16–255.

Full rules: [screens-toml](../../../specs/screens-toml.md), section 3.2.

## The More row

A screen with more rows than the terminal comes in pages. Each page except the last ends
with the "More row". The user presses any key for the next page.
Items do not work on the More row.

```toml
[screen]
more = "    more ∙ any key"
```

`more` gives the text of the More row. Without `more`, the text is `-- More --`.
The style `more` gives its colours.
With the file `screens/screens.toml` below, at 12 rows, the screen `about` has two pages:

```
velourbbs preview about --size 80x12
```

```
page 1 80x12 fill spaces
...
|    more ∙ any key                                                              |
styles
...
page 2 80x12 fill spaces
...
```

The screen has 14 rows without the spacer line. So the spacer line gives no rows.
Page 1 has rows 1–11 of the screen and the More row. Page 2 has rows 12–14.

Full rules: [bbt-format](../../../specs/bbt-format.md), section 6.4,
and [menu-nav](../../../specs/menu-nav.md), section 5.

## The file `screens.toml`

Write `screens/screens.toml` again with this text:

`examples/lamp/en/screens/screens.toml`:
```
[screen]
width = [80, 80]
more = "    more ∙ any key"
fill = "spaces"

[theme.default]
normal = "#e9dccb/#2a201d"
dim    = "#9e8a7c"
line   = "#8a6a5c"
title  = "#f7c7a3 bold"
key    = "#ffb088 bold"
orn    = "#f2a27e"
more   = "#9e8a7c"
md_h1     = "#f7c7a3 bold"
md_h2     = "#f2a27e"
md_h3     = "#d9b99b"
md_strong = "#f7c7a3 bold"
md_em     = "#d9b99b"
md_code   = "#c8b88a"
md_link   = "#d9b99b underline"
md_mark   = "#f2a27e"
md_frame  = "#8a6a5c"
md_dim    = "#9e8a7c"

[theme.256]
normal = "223/235"
dim    = "137"
line   = "95"
title  = "230 bold"
key    = "216 bold"
orn    = "216"
more   = "137"
md_h1     = "230 bold"
md_h2     = "216"
md_h3     = "180"
md_strong = "230 bold"
md_em     = "180"
md_code   = "144"
md_link   = "180 underline"
md_mark   = "216"
md_frame  = "95"
md_dim    = "137"

[theme.16]
normal = "7/0"
dim    = "8"
line   = "8"
title  = "15 bold"
key    = "15 bold"
orn    = "7"
more   = "8"
md_h1     = "15 bold"
md_h2     = "15"
md_h3     = "15"
md_strong = "15 bold"
md_em     = "7"
md_code   = "15"
md_link   = "15 underline"
md_mark   = "7"
md_frame  = "8"
md_dim    = "8"
```

- At 256 colours, the ground is the dark grey 235, and the text is the cream 223, not the grey 253.
- At 16 colours, the screen is grey: the black ground, the light grey text,
  the dark grey `dim` and `line`, the white `title` and `key`.
- The styles `md_*` are the styles of the texts. Task 7 uses them.

## Look at each depth

`velourbbs preview` uses the type `xterm`, so 256 colours. Set another depth with `--set`:

```
velourbbs preview main --set term.colors=16
velourbbs preview main --set term.type=vt100
```

Both commands show the screen with the styles of `[theme.16]`.
The dump shows the colours of the theme as you wrote them.
Only the telnet client shows the nearest colours of its depth.

## The light theme

Some terminals have a light background. A dark ground on a light terminal looks wrong.
The "light theme" gives other styles for these terminals.

```toml
[screen]
theme = "auto"

[theme.light.default]
normal = "#3a2a24/#f4ece0"
title  = "#8a3a1c bold"

[theme.light.16]
normal = "0/15"
title  = "4 bold"
```

- `[screen] theme` selects the theme at the start of a session: `dark`, `light` or `auto`.
  Without `theme`, the session uses the dark theme.
- With `auto`, the field `term.background` selects the theme (task 5).
  `light` gives the light theme. A terminal without a reply gets the dark theme.
- The "light tables" `[theme.light.default]`, `[theme.light.16]` and `[theme.light.256]`
  give other styles in the light theme.
- The light theme of a depth is the theme of this depth with the styles of `[theme.light.default]`,
  then with the styles of the light table of this depth.
- Each name of a light table must be a name of the theme, as in a depth theme.
- In the light theme, `normal` takes a missing colour from `0/15`: black text on a white ground.

Lamp keeps the dark theme. To try the light theme, add the key `theme` to the table `[screen]`
of `screens/screens.toml` and the light tables to the end of the file.
Then set the background of the terminal:

```
velourbbs preview main --set term.background=light
velourbbs preview main --set term.background=light --set term.colors=16
```

The first command shows the styles of `[theme.light.default]`.
The second command also shows the styles of `[theme.light.16]`.

### The theme of the user

Items with the action `theme` let the user select the theme.
To try them, add these items to the front matter of `screens/main.bbt`
and a key row to its text:

```toml
[[item]]
key = "L"
do = "theme light"
[[item]]
key = "D"
do = "theme dark"
[[item]]
key = "B"
do = "theme auto"
```

```
[L]ight  [D]ark  [B]y terminal: {session.theme}
```

- `theme light` and `theme dark` select the theme. `theme auto` gives the selection back to `term.background`.
- The BBS shows the current screen again with the new theme. In a file screen, it shows the same page.
- The choice lasts until the end of the session.
- The field `session.theme` gives the choice: `dark`, `light` or `auto`.
  At the start of a session, it is `[screen] theme`.

Run the items with `--keys`. `--set session.theme` sets the choice at the start:

```
velourbbs preview main --keys D,L,B --set term.background=light
velourbbs preview main --set session.theme=light
```

The first command shows `main` with the dark theme after `D` and with the light theme after `L`.
After `B`, it shows the light theme again, because `term.background` is `light`.

Full rules: [screens-toml](../../../specs/screens-toml.md), sections 3.1 and 3.2,
and [menu-nav](../../../specs/menu-nav.md), section 4.

## Result

- `velourbbs check` gives no output.
- `velourbbs preview main` gives row 4 with the styles of `[theme.256]`:

```
4: 1-7 223/235, 8-11 230/235 bold, 12-13 223/235, 14-15 216/235, 16-33 137/235, 34-80 223/235
```

- `velourbbs preview main --set term.colors=16` gives row 4 with the styles of `[theme.16]`:

```
4: 1-7 7/0, 8-11 15/0 bold, 12-15 7/0, 16-33 8/0, 34-80 7/0
```

- PuTTY shows the dark grey ground and the cream text, not grey text.
  In PuTTY, set Connection > Data > "Terminal-type string" to `vt100` and connect again:
  PuTTY shows the grey screen of 16 colours.
  Then set the "Terminal-type string" back to `xterm`.
