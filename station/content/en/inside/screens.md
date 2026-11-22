# Screens

## What it does

A screen is a BBT text file that the SysOp writes. At each show, Velour builds a page from it
for the terminal of the guest: the width, the height, the colours and the language.

## How

### File

A BBT file is UTF-8. At the top, between the lines `+++`, are the settings in TOML 1.1
([TOML](https://toml.io/en/v1.1.0)): the width, the menu items, the effects.
Below is the text with tokens:

- `{term.cols}`: a field; `{term.type      }`: a field of fixed width;
- `{!gold}`: a style by its name in the theme;
- `{#if term.cols >= 100}`, `{#else}`, `{/if}`: a condition;
- `{>_foot}`: an include of a fragment; `{>logo.ans}`: an art insert;
- `{~}`, `{~.}`, `{~─}`: a stretch.

### Width and height

A screen gives a range of widths. This station has `width = [68, 108]`, `align = "center"`
and `gutter = 6`: the text column is 64 cells at 80 columns and 104 at 132.
The stretches share the free cells of a row: `{~.}` draws a dot leader, `{~─}` a rule.
A line `{~}` shares the free rows of the screen. So the key row stays at the bottom
at 24 and at 43 rows.

### Conditions

The grammar of the conditions has 7 rules: `or`, `and`, `not` and the comparisons `==`, `!=`, `<`, `<=`, `>`, `>=`.
Numbers take each operator, strings only `==` and `!=`.
There are no parentheses and no arithmetic. The types of the fields are known in advance, so `velourbbs check`
finds a type error before the start.

### Theme

`screens.toml` maps the role names of the brandbook to styles. The tables `[theme.16]` and `[theme.256]`
give other colours for terminals with 16 and 256 colours, `[theme.light]` for a light background.
With `theme = "auto"`, Velour selects the light theme by the background colour from the terminal probe.

### Live screen

A screen lives while it waits for a key, and it needs no script for this.

- A live field, for example `{bbs.time live}`, changes each second in its place.
  So the clock at the bottom of the sections runs. On the screen "About the station",
  the time online, the number of open sessions and the uptime of the station change the same way.
- There are six effects: `twinkle` (stars twinkle), `sparkle` (stars appear and go),
  `candle` (a candle flickers), `glint` (a letter flashes), `sweep` (a light runs along a row)
  and `lightning` (lightning strikes).
- This station uses `sweep`: over the logotype at the entry, over the header with its line
  and over the line above the keys. The light on the header runs about once in 45 seconds.

### Three languages

One BBS directory holds Russian, English and Lithuanian. The screens of a language are
in `screens/<language>/`, the texts in `content/<language>/`. A file without text, for example the logotype,
is one for all languages. The entry screen asks for the language and opens the menu in it.
A test checks that the screens of all languages have the same items.

### Check

`velourbbs check` reads all screens before the start. An error comes with the file,
the line and the column. An unknown TOML key is also an error with its place.

## What it gets

One page for one terminal. An example is the screen "About the station":
it shows the size, the type, the program name, the colours, the background and the width of Cyrillic
that Velour found at connection.

## Limits

- The layout does not change with `term.ambiguous` 2: the columns move, frames and art break.
- A grapheme of width 0 or longer than 32 bytes in a file is an error.
- There are no scripts yet. Lua mods are planned.
