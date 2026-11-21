# Live screen

This specification defines the live page, its ticks and the effects.
Related parts: [bbt-format](bbt-format.md), [screens-toml](screens-toml.md), [menu-nav](menu-nav.md),
[session-fields](session-fields.md), [screen-render](screen-render.md) and [telnet-term](telnet-term.md).

## 1. Goal

A screen lives while it waits for a key, with no script.
The time fields tick. Effects animate the page: twinkling stars, flickering candles, glints,
a sweep along a row and lightning.

## 2. Concepts

- The "shown page" is the page that the host shows and that waits for keys
  ([menu-nav](menu-nav.md), section 5): the last page of a screen, the current page of a file screen
  or of a list screen. A page before a More row is not a shown page.
- The "show" is the output of the shown page with a clear ([telnet-term](telnet-term.md), section 7.1).
  The "show time" is the time of the show.
- A "live page" is a shown page of a screen with `live = true` (section 3).
- A "tick" is a new render of a live page at a whole second (section 4).
- The "base" is the last render of the shown page: the render of the show or of the last tick.
- An "effect" is one `[[effect]]` table of the front matter (section 6).
  The "layer" is the cells that the effects change on the base at a time.
- The "grid" is the cells of all terminal rows: the rows of the page with the cells outside the screen area,
  and the fill rows ([bbt-format](bbt-format.md), section 6.5). A fill row with `fill = "none"`
  has spaces in the colour `default` without attributes.
- A "frame" is the output of the changed cells of the grid ([telnet-term](telnet-term.md), section 7.5).
- A "field cell" is a cell of the value of a field token.

## 3. Front matter

A screen file has these keys in its front matter ([bbt-format](bbt-format.md), section 3.2):

| Key | Type | Meaning |
|---|---|---|
| `live` | boolean | `true`: the shown page ticks (section 4). The default is `false`. |
| `[[effect]]` | array of tables | The effects of the shown page, in source order (section 6). They run also without `live`. An array of inline tables is the same. |

Example: the clock ticks, the stars twinkle and a sweep runs over the logo once.

```
+++
live = true
[[effect]]
do = "twinkle"
on = "*+·"
[[effect]]
do = "sweep"
style = ["g1", "g3", "g5", "g7"]
[[item]]
key = "Q"
do = "logoff"
+++
```

`velourbbs check` ([screen-render](screen-render.md), section 3.1) reports these errors:

- `live` or `[[effect]]` in `goodbye` or `shutdown`: these screens do not wait for a key.
- More than 16 effects.
- A missing or unknown `do`, or a key that the table of the effect in section 6 does not have.
- A value outside its range in the tables of section 6.
- `on` with a control character, a grapheme of width other than 1, or more than 32 graphemes.
- `style` with a name that is not a name of the theme ([bbt-format](bbt-format.md), section 7.3).
- A style value of `peak` or `colors` that is not a valid value of a theme style
  ([screens-toml](screens-toml.md), section 3.2): a style or the name of a theme style.
- `spread` with `every - spread < 0.05`.

The errors have the positions of [screen-render](screen-render.md), section 3.1.
The message starts with `effect <n>: `, the 1-based number of the effect.

## 4. Live page

- The host ticks a live page at each whole second of the clock of the server.
- A tick takes the values of the session fields again ([session-fields](session-fields.md), section 2).
  The terminal values, the arguments, the theme and the session language stay the values of the show.
  So a resize takes effect at the next screen, as before.
- A tick evaluates the `when` of each item and renders the shown page again
  ([menu-nav](menu-nav.md), section 5, steps 2 and 3):
  - a screen with More rows: the last page of the new render;
  - a file screen: the current file page. The file pages stay ([file-view](file-view.md), section 3.2);
  - a list screen: the current list page and the cursor. The page size stays.
- The new render is the base. Its enabled items match the next keys ([menu-nav](menu-nav.md), section 3.3).
  Example: an item with `when = "bbs.sessions > 1"` becomes enabled at the tick after a second session opens.
- The new render can have another count of rows or another first row. The frame sends
  the changed cells of the grid (section 5), so the host never clears the screen at a tick.
- A tick is output, not a key. It does not reset the idle timeout ([telnet-term](telnet-term.md), section 8.3).
- A More row waits for a key with no ticks and no effects.
- The line input of `ask` ([menu-nav](menu-nav.md), section 5.4) has no ticks and no frames.
  The page with the ask row has no layer. After the input, the frames go on.
- The stop ([telnet-term](telnet-term.md), section 10.2) ends the wait of a live page as each wait.

## 5. Frames

- The layer is a function of the time from the show time and of the random source of the session (section 6).
- The host makes a frame at each tick and at each time when the layer changes: the start of an event
  of an effect, and each 50 ms while an event runs (section 6).
- Two frames are at least 50 ms apart. A frame that comes earlier waits until 50 ms after the frame before.
  A frame takes the layer and the base at its own time. So after a slow write, the next frame
  shows the current state, and the host does not send the missed frames.
- A frame without changed cells sends nothing.
- The wait for a key ends at the time of the next frame. Keys that arrive during a frame stay in the queue.
- The effects keep their time and their state from the show time until the next show.
  Each show starts them again. Examples: a new screen, `back`, `theme`, a `page` action,
  and the page again after `ask` ([menu-nav](menu-nav.md), section 5.4, step 5).

## 6. Effects

### 6.1. Common rules

- An effect has "events". An event starts at a time and can run for a time.
- The first event starts at the show time plus d. Each next event starts at the start of the event
  before plus d. d is `every` plus a random value from `-spread` to `spread`. The events of one effect can overlap.
- The "selected cells" of an effect are the cells of the rows of the base in the screen area that:
  - have a grapheme of `on`, when the effect has `on`;
  - have a theme style of `style`, when the effect has `style`.
    The theme style of a cell is the name of the last style token `{!<name>}` before it,
    when no colour token (`{!14}`, `{!/1}`) comes after that token. `{!}` gives `normal`.
    A cell of art has no theme style;
  - are not field cells and have a grapheme of width 1.
- Effects never change a field cell or a cell of a grapheme of width 2.
- The effects apply in source order. A later effect replaces the cell of an earlier effect.
- Each session has its own random source: the generator PCG of the Go package `math/rand/v2`
  with a random seed. `velourbbs preview` takes a fixed seed (section 7).
- A tick can move or remove the selected cells. The layer keeps the state of a position
  only while its base cell is selected.

The "mix" of a cell with a colour P at an amount a, from 0 to 1, gives the foreground
`C + (P - C) × a` for each RGB channel, rounded to the nearest integer. C is the foreground of the cell:

- an index 0–15 or a VGA colour: the VGA values of [telnet-term](telnet-term.md), section 7.3;
- an index 16–255: the xterm palette;
- `default`: the foreground of `normal`, or `aaaaaa` when it is `default` too.

The background and the attributes of the cell stay. The output reduces the colour to `term.colors`
([telnet-term](telnet-term.md), section 7.3). So on 16 colours a glint has few steps.

`peak` and each value of `colors` are values of a theme style. The host takes their foreground in the session theme.
So a light table can give another peak ([screens-toml](screens-toml.md), section 3.2).

### 6.2. `twinkle`

On each event, `share` of the selected cells, rounded, but at least 1, change to another grapheme of `on`.
The host takes the cells and the graphemes at random. The style stays. The change stays until the next event of the cell.

| Key | Type | Default | Range |
|---|---|---|---|
| `on` | string | Required. | 2 graphemes or more. |
| `style` | array of strings | None. | Names of theme styles. |
| `every` | number, seconds | `1` | 0.05–3600. |
| `spread` | number, seconds | `0.5` | 0 or more. |
| `share` | number | `0.1` | More than 0, at most 1. |

### 6.3. `candle`

On each event, `share` of the selected cells, rounded, but at least 1, take a random grapheme of `on`
and a random foreground of `colors`. The change stays until the next event of the cell.

| Key | Type | Default | Range |
|---|---|---|---|
| `on` | string | Required. | 1 grapheme or more. |
| `colors` | array of strings | Required. | 1–16 values. |
| `style`, `every`, `spread`, `share` | — | `every = 0.2`, `spread = 0.1`, `share = 0.3`. | As `twinkle`. |

### 6.4. `glint`

Each event takes one random selected cell. The event runs for `time`.
At the time t of the event, the cell is the mix with `peak` at the amount `1 - t / time`.

| Key | Type | Default | Range |
|---|---|---|---|
| `on`, `style` | — | One of them is required. | As `twinkle`, with 1 grapheme or more. |
| `every` | number, seconds | `3` | 0.05–3600. |
| `spread` | number, seconds | `1` | 0 or more. |
| `peak` | string | `#ffffff` | A value of a theme style. |
| `time` | number, seconds | `0.6` | 0.05–10. |

### 6.5. `sweep`

A bright spot runs from left to right over the selected cells.

- The spot runs over each terminal row with selected cells at the same time. So over several rows,
  it is a band. Its centre starts at the column L - `width` and stops at the column R + `width`.
  L is the first column and R the last column with a selected cell, over all rows. It moves `speed` columns each second.
- A selected cell at the distance x from the centre, with x < `width`, is the mix with `peak`
  at the amount `1 - x / width`.
- When the centre passes the last selected cell of a row, that cell starts a glint (section 6.4) with `peak` and `time`.
- The event runs until the centre stops and the last glint ends.
- Without `every`, one event starts 0.5 s after the show time.

| Key | Type | Default | Range |
|---|---|---|---|
| `on`, `style` | — | One of them is required. | As `glint`. |
| `every` | number, seconds | None: one event. | 0.05–3600. |
| `spread` | number, seconds | `0`. Only with `every`. | 0 or more. |
| `peak`, `time` | — | As `glint`. | As `glint`. |
| `width` | integer, cells | `3` | 1–20. |
| `speed` | number, columns each second | `40` | 1–1000. |

### 6.6. `lightning`

Each event draws a bolt on the grid for `time`.

- The main bolt starts at a random column of the terminal row 1. It goes down one terminal row
  at each step, with a random column step of -1, 0 or +1, for a random count of rows from `term.rows / 2`,
  rounded down, to `term.rows`. A step out of the grid ends the part.
- The bolt has 1–3 branches, a random count. A branch starts at a random cell of the main bolt
  below row 1. It goes down with one column step, -1 or +1 at random, for 2–6 rows, a random count.
- The grapheme of a cell shows the step to the next row: `chars` has the graphemes
  of the steps +1, 0 and -1. The last cell of a part has the grapheme of the step 0.
- For the main bolt and for each branch, the host takes at random: behind or in front.
  - A part behind draws only on empty cells. An "empty cell" is a space in the background of `normal`,
    or a cell of a fill row. So the art, the menu text and the fields cover it.
  - A part in front draws on each cell except the cells of section 6.1 that effects never change.
- A drawn cell has the grapheme of the step, the foreground of `peak`, the background of the cell
  and no attributes.

| Key | Type | Default | Range |
|---|---|---|---|
| `every` | number, seconds | `30` | 0.05–3600. |
| `spread` | number, seconds | `15` | 0 or more. |
| `peak` | string | `#ffffff` | A value of a theme style. |
| `time` | number, seconds | `0.3` | 0.05–10. |
| `chars` | string | `╲│╱` | 3 graphemes, as `on`. |

- `╲`, `│` and `╱` are East Asian Ambiguous. The layout already uses 1 cell for them
  ([telnet-term](telnet-term.md), section 6). `chars = "\\|/"` gives ASCII.

## 7. `velourbbs preview`

- In `--keys` ([menu-nav](menu-nav.md), section 7), an element `+<duration>` waits for the time
  `<duration>`: a duration of the Go function `time.ParseDuration` from 1 ms to 1 h. Examples: `+1s`, `+250ms`.
- The preview has a fixed clock. It starts at the time of [session-fields](session-fields.md), section 4.
  Only a wait moves it. A wait moves `bbs.time`, `session.online` and `bbs.uptime`.
  A value of `--set` stays fixed.
- The random source has the fixed seed `(1, 2)`. So the output is the same at each run.
- During a wait, the session makes the ticks and the frames of sections 4 and 5 in order, with the fixed clock.
- The output of a wait is the line `wait <duration>`, then the "grid dump" of the grid at the end of the wait.
- The grid dump is the cell dump of [screen-render](screen-render.md), section 3.3, for all terminal rows:
  the first line is `grid <cols>x<rows>`, then one row between two `|` for each terminal row,
  then the `styles`. A cell of a fill row with `fill = "none"` has the style `-/-`.
- With `--view`, the grid is in the line form ([screen-render](screen-render.md), section 3.7):
  each terminal row as a row of the page.
- Without `--keys`, the preview shows the pages at the show time, with no layer.

Example: the clock `{bbs.time}` with `[screen] time = "15:04:05"` and `live = true`.

```
screen main
page 1 80x24 fill none
|15:04:05 ...
key +1s
wait 1s
grid 80x24
|15:04:06 ...
```

## 8. Blink

The attribute `blink` of a style ([bbt-format](bbt-format.md), section 7.2) is not an effect.
The terminals show it in different ways:

| Terminal | `blink` |
|---|---|
| PuTTY | Off by default. It shows a bright background in place of the blink. |
| Tera Term | No blink. It shows a blink colour. |
| Windows Terminal | Blinks. All cells blink at the same time. |
| xterm | Blinks when the build has blink. Else bold. |
| iTerm2 | Off by default. |
| Ghostty | No blink. |
| tmux | It sends the attribute to the outer terminal. |

- The BBT guide advises effects in place of `blink`.

## 9. Code

| Package | Content |
|---|---|
| `internal/screen` | The keys of section 3 in the model and their checks. The render marks each cell with its theme style and as a field cell. |
| `internal/effect` | The events, the selected cells, the layer of each effect and the mix. It has no I/O and takes the time and the random source as values. |
| `internal/menu` | The ticks of section 4 and the frames of section 5. |
| `internal/term` | The wait for a key with a deadline and the frame output of [telnet-term](telnet-term.md), section 7.5. |
| `cmd/velourbbs` | The waits, the fixed clock and the grid dump of section 7. |

## 10. Acceptance

- A test covers each error of section 3.
- A golden output of `velourbbs preview --keys` shows a live clock over two ticks,
  and an item with `when` on `session.online` that becomes enabled at a tick.
- A golden output shows each effect at three times at least, with the fixed seed.
- A test covers the selection of section 6.1: `on`, `style`, both, a colour token after a style token,
  a field cell, art and a wide grapheme.
- A test covers the mix: an index 0–15, a VGA colour, an index 16–255, RGB and `default`.
- A test covers the frame bytes of [telnet-term](telnet-term.md), section 7.5, and the 50 ms rule.
- A test covers a tick that changes the count of rows of the page.
- A test covers the line input on a live page: no frame during the input.
- A test covers a stop during the wait of a live page.

## 11. Not in this specification

- A screen script in Lua: [Lua mods](https://github.com/schors/velourbbs/issues/12).
- Raw ANSI animation: [raw art](https://github.com/schors/velourbbs/issues/73).
- Layers or marks in the art: [marks](https://github.com/schors/velourbbs/issues/67).
- New timer fields.
- Ticks during the line input.
