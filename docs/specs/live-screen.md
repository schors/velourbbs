# Live screen

This specification defines the live fields, the live page, its ticks and the effects.
Related parts: [bbt-format](bbt-format.md), [art-format](art-format.md), [screens-toml](screens-toml.md),
[menu-nav](menu-nav.md), [session-fields](session-fields.md), [screen-render](screen-render.md)
and [telnet-term](telnet-term.md).

## 1. Goal

A screen lives while it waits for a key, with no script.
A live field ticks in its place. Effects animate places of the page: twinkling stars, stars that appear
and go, flickering candles, glints, a sweep along a row and lightning.

## 2. Concepts

- The "shown page" is the page that the host shows and that waits for keys
  ([menu-nav](menu-nav.md), section 5): the last page of a screen, the current page of a file screen
  or of a list screen. A page before a More row is not a shown page.
- The "show" is the output of the shown page with a clear ([telnet-term](telnet-term.md), section 7.1).
  The "show time" is the time of the show.
- A "live field" is a field token with the live mark (section 3).
- A "slot" is the run of cells that the render gives a live field on the shown page.
- A "live page" is a shown page with a slot or with an effect.
- A "tick" writes the new values into the slots of a live page at a whole second (section 4).
- The "base" is the render of the show or of the last cursor move in a list screen
  ([menu-nav](menu-nav.md), section 5.3), with the slot values of the last tick.
- An "effect" is one `[[effect]]` table of the front matter (section 6).
  Its "place" is the cells where it acts (section 6.2).
  The "layer" is the cells that the effects change on the base at a time.
- The "grid" is the base with the layer: the cells of all terminal rows. A row of the page has
  the cells outside the screen area ([bbt-format](bbt-format.md), section 6.2). A cell of a fill row
  ([bbt-format](bbt-format.md), section 6.5) is a space: in the colour `default` without attributes
  with `fill = "none"`, in `normal` with `"spaces"` and `"erase"`.
- A "frame" is the output of the changed cells of the grid ([telnet-term](telnet-term.md), section 7.5).
- An "empty cell" is a cell of a fill row, or a space with one of these backgrounds:
  the background of `normal`, `default` or black (an index 0, the VGA colour 0 or `#000000`).
  So the black sky of an art is empty.

## 3. Live fields

The live mark follows the path of a field token ([bbt-format](bbt-format.md), section 3.3):

```
field  = [ spaces ] path [ " live" [ " " layout ] ] [ spaces ]
layout = '"' { any character except '"' and '}' } '"'
```

- Examples: `{bbs.time live}`, `{bbs.time live "15:04:05"}`, `{session.online live}`,
  `{bbs.uptime live    }`, `{    bbs.sessions live}`.
- The spaces at the start and at the end have the rules of a field token: the natural width,
  a fixed width aligned left or a fixed width aligned right. The width of a fixed-width field includes
  the whole token.
- `layout` is a layout of the Go package `time`. It is valid only with `live` on `bbs.time`.
  Without it, `bbs.time` has the layout `[screen] time` ([screens-toml](screens-toml.md), section 3.1).
- An alias can hold the mark: `clock = 'bbs.time live "15:04:05"'` ([screens-toml](screens-toml.md), section 3.3).
  Then `{clock}` and `{clock   }` are live fields. The mark goes only after a path of the registry:
  a token with an alias takes no mark.
- In art, a field token has the same form ([art-format](art-format.md), section 4).
  A layout makes the token long, so an alias is the short form in art.
- These fields change while a page is shown: `bbs.time`, `session.online`, `bbs.uptime` and `bbs.sessions`.
  The mark on each other field is valid and gives no slot: the field shows its value of the show.
- The mark is not a part of the path. A condition ([bbt-format](bbt-format.md), section 5) cannot have it.
  `{#if}` and `when` use the values of the show.

`velourbbs check` ([screen-render](screen-render.md), section 3.1) reports these errors:

- A token content after the path that is not the mark of the grammar above.
- The mark on a token with an alias.
- A layout without `live`, on a field other than `bbs.time`, empty or with a control character.
- An alias with the mark in a condition.

### 3.1. Slot

- A live field gives one slot for each row where the render shows it: a row of the body,
  of a fragment or of an art, a frame row of a file screen, and each row of a list line or of a list row
  ([list-view](list-view.md), sections 5 and 9).
- The width of the slot is the width of the token for a fixed-width field and for a field in art.
  Else it is the "largest width" of the value:
  - `bbs.time`: the sum of the largest widths of the elements of the layout. An element of the Go package `time`
    has its largest width: `January` and `Monday` 9; `1`, `2`, `3`, `4`, `5` 2; `.9…` the count of its digits plus 1;
    a `Z07…` form the width of its numeric form (`Z07:00` gives 6); `MST` the widest abbreviation
    of `[screen] zone` from 2000 to 2100 (the Go method `time.Time.ZoneBounds` gives the periods of the zone).
    Each other element and each other character has the width of its value;
  - `session.online` and `bbs.uptime`: 14 cells, the width of `9999d 23:59:59`;
  - `bbs.sessions`: the digits of `max_conns` ([bbs-toml](bbs-toml.md), section 3.3).
- A slot of the natural width aligns left. A value pads with spaces or cuts at the end
  by the rule of a fixed-width field ([bbt-format](bbt-format.md), section 4.3).
- A value wider than its slot, for example a value of `--set` in `velourbbs preview`, is cut.
- A slot is one word ([bbt-format](bbt-format.md), section 6.3). A slot wider than W has the width W.
- The cells of a slot have the style of the field.
- A cut of the row at W or at `term.cols` cuts the slot too. A tick writes only the cells of the page.

## 4. Live page

- The host ticks a live page with a slot at each whole second of the clock of the server.
- A tick takes the values of the session fields again ([session-fields](session-fields.md), section 2).
  It fits the value of each slot to the slot and writes its graphemes into the base.
  It does not evaluate `{#if}` or `when` and does not render the page. The cells outside the slots,
  the rows and the enabled items stay.
- A cursor move in a list screen renders with the values of the show. Its render with the slot values
  of the last tick is the new base. The update sends it with the layer.
- A tick is output, not a key. It does not reset the idle timeout ([telnet-term](telnet-term.md), section 8.3).
- A More row waits for a key with no ticks and no effects.
- The line input of `ask` ([menu-nav](menu-nav.md), section 5.4) has no ticks and no frames.
  The page with the ask row has no layer. After the input, the page shows again (section 5).
- `goodbye` and `shutdown` do not wait for a key: their slots show the values of the show.
- The stop ([telnet-term](telnet-term.md), section 10.2) ends the wait of a live page as each wait.

## 5. Frames

- The layer is a function of the time from the show time and of the random source of the session (section 6).
- Directly after the show, the host makes a frame at the show time: it sends the layer of the time 0.
  A cell dump of the show in `velourbbs preview` is the base (section 7).
- The host makes a frame at each tick and at each time when the layer changes: the start of an event
  of an effect, each 50 ms while an event runs, and the end of an event (section 6).
- Two frames are at least 50 ms apart. A frame that comes earlier waits until 50 ms after the frame before.
  A frame takes the layer and the base at its own time. So after a slow write, the next frame
  shows the current state, and the host does not send the missed frames.
- A frame without changed cells sends nothing.
- The wait for a key ends at the time of the next frame. Keys that arrive during a frame stay in the queue.
- The effects keep their time and their state from the show time until the next show.
  Each show starts them again. Examples: a new screen, `back`, `theme`, a `page` action,
  and the page again after `ask` ([menu-nav](menu-nav.md), section 5.4, step 5).

## 6. Effects

A screen file has the array of tables `[[effect]]` in its front matter ([bbt-format](bbt-format.md), section 3.2).
An array of inline tables is the same. The effects apply in source order.

Example: two stars twinkle in the art, new stars appear in its sky, and a sweep runs over the logo once.

```
+++
[[effect]]
do = "twinkle"
in = "main.ans"
at = [[61, 3], [70, 5]]
chars = "*+·"
[[effect]]
do = "sparkle"
in = "main.ans"
at = [1, 1, 80, 6]
chars = ".·*"
count = 2
[[effect]]
do = "sweep"
in = "logo.ans"
[[item]]
key = "Q"
do = "logoff"
+++
```

### 6.1. Keys and checks

Each effect has these keys:

| Key | Type | Default | Range |
|---|---|---|---|
| `do` | string | Required. | `twinkle`, `sparkle`, `candle`, `glint`, `sweep` or `lightning`. |
| `in` | string | None: the page. | The name of an art insert: `<name>.ans` (section 6.2). |
| `at` | array | None: the whole art or grid. | A point, a rectangle or an array of points (section 6.2). |
| `delay` | number, seconds | `0` | 0–3600. |
| `every` | number, seconds | Of the effect (sections 6.4–6.9). | 0.05–3600. |
| `spread` | number, seconds | Half of `every`, except `sweep`. | 0 or more. |

The tables of sections 6.4–6.9 give the other keys of each effect.

- A "grapheme string" (`chars`) has 1–32 graphemes of width 1 and no control character.
- A "colour value" (`colors`, `peak`) is a value of a theme style ([screens-toml](screens-toml.md), section 3.2):
  a style or the name of a theme style. The host takes its foreground in the session theme.
  So a light table can give another colour.

`velourbbs check` ([screen-render](screen-render.md), section 3.1) reports these errors:

- `[[effect]]` in `goodbye` or `shutdown`: these screens do not wait for a key.
- More than 16 effects.
- A missing or unknown `do`, or a key that the effect does not have.
- A value outside its range, an invalid grapheme string or an invalid colour value.
- A written `spread` with `every - spread < 0.05`, and `spread` of a `sweep` without `every`.
- An `at` that is not a point, a rectangle or an array of 1–64 points; a coordinate 0;
  a width or a height less than 1. The coordinates are integers.
- `at` of the kind that the effect does not take (sections 6.4–6.9).
- `in` with an art that the screen and its fragments do not insert, in a screen set
  ([session-language](session-language.md), section 6).
- With `in`, a cell of the place that is not in the art grid, in a screen set.

The errors have the positions of [screen-render](screen-render.md), section 3.1.
The message starts with `effect <n>: `, the 1-based number of the effect.

### 6.2. Place

- A "point" is `[col, row]`. A "rectangle" is `[col, row, width, height]`: its first cell is the point,
  then `width` columns to the right and `height` rows down. An array of points is `[[col, row], ...]`.
- With `in`, the coordinates are in the art grid ([art-format](art-format.md), section 3) of the first insert
  of that art on the shown page. The place moves with the art.
- Without `in`, the coordinates are in the shown page after the layout. Column 1 is the left edge
  of the screen area ([bbt-format](bbt-format.md), section 6.1). Row 1 is the first row of the page.
  A place can reach the margins and the fill rows: the host cuts it to the cells of the grid.
  These coordinates do not follow a text that wraps at another width. Use them on rows that do not wrap:
  a header, a corner or the key row.
- A negative coordinate counts from the end: -1 is the last column or row of the art grid,
  or of the screen area and of the page.
- Without `at`, the place is the whole art of `in`, else all cells of the grid.
- The "rectangle of a place" is the rectangle of `at`, the art grid or the grid, cut to the cells of the grid.
- A cell of a place that the page does not show is not in the place: a cut at W or at `term.cols`,
  an art in a false block or on a page before a More row. A place without cells gives no events.
- Each point of an array of points is a place with its own events (section 6.3).
- An effect never changes a slot cell or a cell of a grapheme of width 2. These cells are not in a place.

### 6.3. Events

- An effect has "events" in each place. An event starts at a time and can run for a time.
- The first event starts at the show time plus `delay` plus d. Each next event starts at the start
  of the event before plus d. d is `every` plus a random value from `-spread` to `spread`,
  with the same chance for each value. An effect without `every` (a `sweep`) has one event
  at the show time plus `delay`.
- The events of one place can overlap. On a cell of two events, the event that started later wins.
  At the end of an event, its cells show the other events or the cell before the effect.
- An event that finds no cell changes nothing and takes no random values.
  The next events start as before.
- Each effect takes the cell after the earlier effects: its grapheme, and its foreground for the mix.
  An event takes its cells at its start, from the grid after the earlier effects.
  The stars of a `twinkle` rectangle are the exception: they come from the base at the show.
- A new base (a cursor move) does not start the effects again. The layer keeps its state.
- "Rounded" is rounded half up. The host takes the random cells of one event as different cells.
- Each session has its own random source: the generator PCG of the Go package `math/rand/v2`
  with a random seed. `velourbbs preview` takes a fixed seed (section 7).

The "mix" of a cell with a colour P at an amount a, from 0 to 1, gives the foreground
`C + (P - C) × a` for each RGB channel, rounded to the nearest integer. C is the foreground of the cell:

- an index 0–15 or a VGA colour: the VGA values of [telnet-term](telnet-term.md), section 7.3;
- an index 16–255: the xterm palette;
- `default`: the foreground of `normal`, or `aaaaaa` when it is `default` too.

P has the same rules. The background and the attributes of the cell stay. The output reduces the colour
to `term.colors` ([telnet-term](telnet-term.md), section 7.3). So on 16 colours a glint has few steps.

A drawn grapheme keeps the background and the attributes of the cell. Its foreground is a random value
of `colors` when the effect has `colors`, else the foreground of the cell.

### 6.4. `twinkle`

A star that is always visible and changes its grapheme.

- A point: from the show, the cell shows a star: its grapheme when it is a grapheme of `chars`,
  else the first grapheme of `chars`. At each event, the star changes to another grapheme of `chars`.
- A rectangle: the stars are the cells of the rectangle with a grapheme of `chars`.
  They show their grapheme until their first change. At each event, `share` of the stars, rounded,
  but at least 1, change to another grapheme of `chars`.
- The host takes the stars and the graphemes at random. A change stays until the next change of the star.

| Key | Type | Default | Range |
|---|---|---|---|
| `at` | — | Required. | A point, a rectangle or an array of points. |
| `chars` | string | Required. | A grapheme string with 2 different graphemes or more. |
| `colors` | array of strings | None. | 1–16 colour values. |
| `every` | number, seconds | `1` | Section 6.1. |
| `share` | number | `0.1` | More than 0, at most 1. |

### 6.5. `sparkle`

A star that appears and goes.

- A point: at each event, the cell shows a random grapheme of `chars` for `time`.
- A rectangle, or the whole art or grid without `at`: at each event, `count` random cells of the place
  show a random grapheme of `chars` for `time` each. When fewer cells can take a star, all of them take one.
- With `over = false`, only an empty cell can take a star. With `over = true`, each cell of the place can.
  A cell with a star of this effect cannot take another.

| Key | Type | Default | Range |
|---|---|---|---|
| `at` | — | None. | A point, a rectangle or an array of points. |
| `chars` | string | Required. | A grapheme string. |
| `colors` | array of strings | None. | 1–16 colour values. |
| `over` | boolean | `false` | — |
| `count` | integer | `1` | 1–100. Not with an array of points or a point. |
| `time` | number, seconds | `2` | 0.05–60. |
| `every` | number, seconds | `1` | Section 6.1. |

### 6.6. `candle`

A flame that flickers.

- From the show, each cell of the place shows the first grapheme of `chars` in the first value of `colors`.
- A point: at each event, the cell takes a random grapheme of `chars` and a random value of `colors`.
- A rectangle: at each event, `share` of its cells, rounded, but at least 1, do so.
  The host takes the cells at random.

| Key | Type | Default | Range |
|---|---|---|---|
| `at` | — | Required. | A point, a rectangle or an array of points. |
| `chars` | string | Required. | A grapheme string. |
| `colors` | array of strings | Required. | 1–16 colour values. |
| `every` | number, seconds | `0.2` | Section 6.1. |
| `share` | number | `0.3` | More than 0, at most 1. |

### 6.7. `glint`

A cell becomes bright and fades.

- The cells of the place that are not spaces can glint.
- Each event takes one random cell of them. The event runs for `time`.
  At the time t of the event, the cell is the mix with `peak` at the amount `1 - t / time`.
- A point is one fixed glint.

| Key | Type | Default | Range |
|---|---|---|---|
| `at` | — | None. | A point, a rectangle or an array of points. |
| `peak` | string | `#ffffff` | A colour value. |
| `time` | number, seconds | `0.6` | 0.05–10. |
| `every` | number, seconds | `3` | Section 6.1. |

### 6.8. `sweep`

A bright spot runs from left to right over the cells of the place that are not spaces: the "lit cells".

- The spot runs over each terminal row with lit cells at the same time. So over several rows,
  it is a band. Its centre starts at the column L - `width` and stops at the column R + `width`.
  L is the first column and R the last column with a lit cell, over all rows. It moves `speed` columns each second.
- A lit cell at the distance x from the centre, with x < `width`, is the mix with `peak`
  at the amount `1 - x / width`.
- When the centre is at the column of the last lit cell of a row, that cell starts a glint (section 6.7)
  with `peak` and `time`. The host takes this time from `speed`, not from a frame.
  The amount of the cell is the larger amount of the band and of the glint.
- The event runs until the centre stops and the last glint ends.
- Without `every`, the effect has one event.

| Key | Type | Default | Range |
|---|---|---|---|
| `at` | — | None. | A rectangle. |
| `delay` | number, seconds | `0.5` | Section 6.1. |
| `every` | number, seconds | None: one event. | Section 6.1. |
| `spread` | number, seconds | `0`. Only with `every`: else an error. | Section 6.1. |
| `peak`, `time` | — | As `glint`. | As `glint`. |
| `width` | integer, cells | `3` | 1–20. |
| `speed` | number, columns each second | `40` | 1–1000. |

### 6.9. `lightning`

Each event draws a bolt in the place for `time`.

- The bolt runs in the rectangle of the place (section 6.2). A cell of the bolt that is not in the place
  is not drawn.
- The main bolt starts at a random column of the first row of the rectangle. It goes down one row
  at each step, with a random column step of -1, 0 or +1, for a random count of rows from h / 2,
  rounded down, but at least 1, to h. h is the height of the rectangle. A step out of the rectangle is a step 0.
- The bolt has 1–3 branches, a random count. A branch starts at a random cell of the main bolt
  below its first row. Its first cell is in the next row. It goes down with one column step, -1 or +1 at random,
  for 2–6 rows, a random count. A step out of the rectangle ends the branch.
  A main bolt of one row, or a branch without a cell in the next row, gives no branch.
- Where two parts cross, the later part wins: the branches come after the main bolt, in their order.
- The grapheme of a cell shows the step to the next row: `chars` has the graphemes
  of the steps +1, 0 and -1. The last cell of a part has the grapheme of the step 0.
- With `over = false`, the bolt draws only on empty cells: the art, the text and the fields cover it.
  With `over = true`, it draws on each cell of the place.
- A drawn cell has the grapheme of the step, the foreground of `peak`, the background of the cell
  and no attributes.

| Key | Type | Default | Range |
|---|---|---|---|
| `at` | — | None. | A rectangle. |
| `chars` | string | `╲│╱` | A grapheme string of 3 graphemes. |
| `peak` | string | `#ffffff` | A colour value. |
| `over` | boolean | `false` | — |
| `time` | number, seconds | `0.3` | 0.05–10. |
| `every` | number, seconds | `30` | Section 6.1. |

- `╲`, `│` and `╱` are East Asian Ambiguous. The layout already uses 1 cell for them
  ([telnet-term](telnet-term.md), section 6). `chars = "\\|/"` gives ASCII.

## 7. `velourbbs preview`

- In `--keys` ([menu-nav](menu-nav.md), section 7), an element `+<duration>` waits for the time
  `<duration>`: a duration of the Go function `time.ParseDuration` from 1 ms to 1 h. Examples: `+1s`, `+250ms`.
  A wait is not a key. It gives no `key` line.
- The preview has a fixed clock. It starts at the time of [session-fields](session-fields.md), section 4.
  Only a wait moves it. A wait moves `bbs.time`, `session.online` and `bbs.uptime`.
  A value of `--set` stays fixed.
- The random source has the fixed seed `(1, 2)`. So the output is the same at each run.
- During a wait, the session makes the ticks and the frames of sections 4 and 5 in order, with the fixed clock.
- The output of a wait is the line `wait <duration>`, with the duration as the element gives it,
  then the "grid dump" of the grid at the end of the wait.
- A tick or an event at the end of a wait is in the wait.
- In the line input of `ask` and on a More row, a wait moves the clock with no frame.
  Its output is only the line `wait <duration>`.
- A wait after the end of the session at `logoff` is an error, as a key is.
- The grid dump is the cell dump of [screen-render](screen-render.md), section 3.3, for all terminal rows:
  the first line is `grid <cols>x<rows>`, then one row between two `|` for each terminal row,
  then the `styles`. A cell of a fill row with `fill = "none"` has the style `-/-`.
- With `--view`, the grid is in the line form ([screen-render](screen-render.md), section 3.7):
  each terminal row as a row of the page.
- Without `--keys`, the preview shows the pages at the show time, with no layer.

Example: `{clock}` with `clock = 'bbs.time live "15:04:05"'`.

```
screen main
page 1 80x24 fill none
|15:04:05 ...
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
| `internal/bbt`, `internal/art` | The live mark in a field token and in an alias, the front matter `[fields]` and `[[effect]]`, and their checks. |
| `internal/screen` | The live field in the model, the slot width and the slots of a page. The render marks the cells of each slot. |
| `internal/effect` | The places, the events, the layer of each effect and the mix. It has no I/O and takes the time and the random source as values. |
| `internal/menu` | The ticks of section 4 and the frames of section 5. |
| `internal/term` | The wait for a key with a deadline and the frame output of [telnet-term](telnet-term.md), section 7.5. |
| `cmd/velourbbs` | The waits, the fixed clock and the grid dump of section 7. |

## 10. Acceptance

- A test covers each error of sections 3 and 6.1.
- A golden cell dump shows a slot in a text row with a stretch token, in an art, in a list line
  and a slot of each field of section 3 with its largest width.
- A golden output of `velourbbs preview --keys` shows a live clock over two ticks.
- A golden output shows each effect at three times at least, with the fixed seed: with `in` and without `in`,
  with a point, an array of points and a rectangle where the effect takes them.
- A test covers a place in art that moves with the art, a negative coordinate, `over`, a slot cell
  and a wide grapheme in a place.
- A test covers the mix: an index 0–15, a VGA colour, an index 16–255, RGB and `default`.
- A test covers the frame bytes of [telnet-term](telnet-term.md), section 7.5, and the 50 ms rule.
- A test covers the line input on a live page: no frame during the input.
- A test covers a stop during the wait of a live page.

## 11. Not in this specification

- Anchors in the text for the place of an effect: [marks](https://github.com/schors/velourbbs/issues/67).
- A screen script in Lua: [Lua mods](https://github.com/schors/velourbbs/issues/12).
- Raw ANSI animation: [raw art](https://github.com/schors/velourbbs/issues/73).
- New timer fields.
- Ticks during the line input.
