# Task 10. Effects

An "effect" changes cells of a screen while the screen waits for a key.
In this task, you add effects to the lamp of task 9: the light of the lamp shimmers,
stars twinkle around it, small sparks appear in the dark and go, two small flames burn
beside the key, the shade glints, and a bright band runs over the lamp once.
Use the directory `myart` of task 9.

## The table `[[effect]]`

Each effect is one table `[[effect]]` in the front matter:

```
[[effect]]
do = "twinkle"
at = [37, 3]
chars = "*+·"
```

- `do` names the effect: `twinkle`, `sparkle`, `candle`, `glint`, `sweep` or `lightning`.
- `at` is the "place": the cells where the effect acts (below).
- A screen has at most 16 effects. The effects apply in the order of the file:
  an effect acts on the cells after the effects above it.
- The screens `goodbye` and `shutdown` do not wait for a key, so they cannot have effects.
- Effects do not act on a page with a More row and while the user types a line (task 8).
- Each time the BBS shows the screen, the effects start again.

Full rules: [live-screen](../../../specs/live-screen.md), sections 5 and 6.1.

## The place

`at` has one of three forms:

| Form | Example | Cells |
|---|---|---|
| A point `[col, row]` | `[61, 3]` | One cell. |
| A rectangle `[col, row, width, height]` | `[1, 1, 80, 6]` | The rectangle from the point, `width` columns to the right and `height` rows down. |
| An array of points | `[[61, 3], [70, 5]]` | Each point is a place with its own timing. |

- With `in = "<name>.ans"`, the coordinates are in the art: column 1 and row 1 are
  the first cell of the art. The place moves with the art. When the text above the art wraps
  into more rows, the effect stays on the same cells of the art.
- Without `in`, the coordinates are in the page. Column 1 is the left edge of the screen,
  row 1 is the first row of the page. A text that wraps moves the rows below it, so use
  this form only for rows that do not wrap: a header, a corner or the key row.
- A negative coordinate counts from the end: `-1` is the last column or the last row
  of the art, or of the screen and the page. Example: `[-1, -1]` is the cell in the lower right corner.
- A live field (task 5) and a wide character, for example `語`, are not in a place.
  An effect never changes them.
- When the screen does not show a cell, the cell is not in the place: for example,
  a cell of art that is cut at the edge of a narrow terminal.

`velourbbs check` gives an error when the screen does not insert the art of `in`,
or when a cell of the place is outside the art.

Full rules: [live-screen](../../../specs/live-screen.md), section 6.2.

## The timing

| Key | Default | Value |
|---|---|---|
| `delay` | `0` | The seconds before the timing starts: 0–3600. |
| `every` | Of the effect. | The seconds between two changes: 0.05–3600. |
| `spread` | Half of `every`. | A random part of `every`: each time is `every` plus or minus up to `spread`. |

- A value is a number of seconds, for example `0.5`.
- The first change comes `delay` plus one interval after the show.
  Example: `delay = 1` and `every = 1` with `spread = 0` give the first change after 2 seconds.
- With `spread = 0`, each change comes exactly `every` seconds after the change before.
- `every - spread` must be at least 0.05.
- The changes are random. Each session has its own random values.

Full rules: [live-screen](../../../specs/live-screen.md), section 6.3.

## The effect `twinkle`

A star that is always visible and changes its character.

| Key | Default | Value |
|---|---|---|
| `at` | Required. | A point, a rectangle or an array of points. |
| `chars` | Required. | 1–32 characters of one cell, with 2 different characters or more. |
| `colors` | None. | 1–16 colours: a style, for example `"215"` or `"#ffcc00"`, or the name of a theme style. |
| `every` | `1` | |
| `share` | `0.1` | The part of the stars of a rectangle that change at each time: more than 0, at most 1. |

- A point shows a star from the start: its own character when it is in `chars`, else the first
  character of `chars`. At each change, the star takes another character of `chars`.
- In a rectangle, the stars are the cells with a character of `chars`.
  At each change, `share` of them take another character of `chars`.
- With `colors`, each new character takes a random colour of `colors`.
  Else it keeps the colour of the cell. The background stays.

## The effect `candle`

A flame that flickers.

| Key | Default | Value |
|---|---|---|
| `at` | Required. | A point, a rectangle or an array of points. |
| `chars` | Required. | 1–32 characters of one cell. |
| `colors` | Required. | 1–16 colours, as in `twinkle`. |
| `every` | `0.2` | |
| `share` | `0.3` | The part of the cells of a rectangle that change at each time. |

- From the start, each cell of the place shows the first character of `chars`
  in the first colour of `colors`.
- At each change, a point takes a random character of `chars` and a random colour of `colors`.
  In a rectangle, `share` of the cells do so.

Full rules: [live-screen](../../../specs/live-screen.md), sections 6.4 and 6.6.

## The effect `sparkle`

A star that appears and goes.

| Key | Default | Value |
|---|---|---|
| `at` | None: the whole art or page. | A point, a rectangle or an array of points. |
| `chars` | Required. | 1–32 characters of one cell. |
| `colors` | None. | 1–16 colours, as in `twinkle`. |
| `over` | `false` | `true`: a star can cover each cell. |
| `count` | `1` | The stars of a rectangle at each time: 1–100. Not with a point. |
| `time` | `2` | The seconds that a star shows: 0.05–60. |
| `every` | `1` | |

- At each time, a point shows a random character of `chars` for `time` seconds.
  In a rectangle, `count` random cells do so.
- With `over = false`, a star appears only on an empty cell: a space.
  So the art and the text stay. With `over = true`, a star can cover each cell of the place.
- A cell with a star does not take a second star of the same effect.
- With `colors`, each star takes a random colour of `colors`. Else it keeps the colour of the cell.
  The background stays.

Full rules: [live-screen](../../../specs/live-screen.md), section 6.5.

## The brightness

`glint` and `sweep` make a cell bright and keep its character.
They move the colour of the character to the colour `peak` and back.
The background stays.

- A `peak` of `"#ffffff"` gives white. A `peak` of `"230"` gives cream. The name of a theme style
  gives the colour of the style.
- A cell in the colour `default` takes the colour of `normal`.
- On a terminal with 16 colours, the steps between the colours are few.
  On 256 colours and on true colour, they are smooth.

Full rules: [live-screen](../../../specs/live-screen.md), section 6.3.

## The effect `glint`

A cell becomes bright and fades.

| Key | Default | Value |
|---|---|---|
| `at` | None: the whole art or page. | A point, a rectangle or an array of points. |
| `peak` | `#ffffff` | The colour of the brightest moment: a style or the name of a theme style. |
| `time` | `0.6` | The seconds from the brightest moment to the end: 0.05–10. |
| `every` | `3` | |

- At each time, one random cell of the place becomes bright. It fades in `time` seconds.
- A space does not glint. Only a cell with a character can glint.
- A point glints always at the same cell.

## The effect `sweep`

A bright spot runs from left to right over the characters of the place.

| Key | Default | Value |
|---|---|---|
| `at` | None: the whole art or page. | A rectangle. |
| `delay` | `0.5` | The seconds before the timing starts, as in the other effects. |
| `every` | None: one run. | The seconds between two runs. |
| `spread` | `0` | Only with `every`. |
| `peak` | `#ffffff` | As in `glint`. |
| `time` | `0.6` | The seconds of the glint at the end of each row. |
| `width` | `3` | The half width of the spot in cells: 1–20. |
| `speed` | `40` | The columns each second: 1–1000. |

- The spot runs over all rows of the place at the same time. Over several rows,
  it is a band.
- A space does not become bright. So the band shows the shape of the characters.
- When the spot comes to the last character of a row, that character glints for `time` seconds.
- Without `every`, the band runs once, `delay` seconds after the show.
- With `every`, the timing is as in the other effects: the first run comes `delay` plus one interval
  after the show. Example: `delay = 0.5` and `every = 2` give the first run after 2.5 seconds.

Full rules: [live-screen](../../../specs/live-screen.md), sections 6.7 and 6.8.

## The effect `lightning`

A bolt that flashes for a short time.

| Key | Default | Value |
|---|---|---|
| `at` | None: the whole art or page. | A rectangle. |
| `chars` | `╲│╱` | 3 characters: the steps down to the right, straight down and down to the left. |
| `peak` | `#ffffff` | The colour of the bolt: a style or the name of a theme style. |
| `over` | `false` | `true`: the bolt can cover each cell. |
| `time` | `0.3` | The seconds that a bolt shows: 0.05–10. |
| `every` | `30` | |

- The bolt starts at a random column of the top row of the rectangle. It goes down
  to the bottom half of the rectangle or further, with random steps to the left and to the right.
  It has 1–3 branches.
- With `over = false`, the bolt draws only on empty cells: the art and the text cover it.
  With `over = true`, it covers each cell of the place.
- A bolt keeps the background of each cell. It has no bold and no other attributes.
- For ASCII characters, write `chars = "\\|/"`: in TOML, `\\` is one `\`.

Example: a bolt in the sky above the text, about each 20 seconds, on a page with 80 columns:

```
[[effect]]
do = "lightning"
at = [1, 1, 80, 8]
every = 20
```

Full rules: [live-screen](../../../specs/live-screen.md), section 6.9.

## The file

Change `screens/main.bbt`:

`examples/art/screens/main.bbt`:
```
+++
width = [40, 40]
fill = "spaces"
[[effect]]
do = "twinkle"
in = "lamp.ans"
at = [8, 5, 26, 3]
chars = "░▒"
share = 0.2
[[effect]]
do = "twinkle"
in = "lamp.ans"
at = [[4, 2], [37, 3], [2, 9], [39, 10]]
chars = "·+*"
colors = ["223", "215"]
every = 1.5
[[effect]]
do = "sparkle"
in = "lamp.ans"
chars = "·."
colors = ["137", "223"]
count = 2
time = 1.5
[[effect]]
do = "candle"
at = [[13, -1], [28, -1]]
chars = "▲▴"
colors = ["208", "214", "220"]
[[effect]]
do = "glint"
in = "lamp.ans"
at = [11, 2, 20, 3]
every = 2
[[effect]]
do = "sweep"
in = "lamp.ans"
[[item]]
key = "G"
do = "logoff"
+++
{>lamp.ans}

               {!key}G{!}  goodbye
```

- The first `twinkle` is the light under the shade: a rectangle of the art, the rows 5–7.
  Its stars are the cells `░` and `▒`. About each second, a fifth of them change.
- The second `twinkle` gives four stars in the dark ground of the art, in cream and orange.
- The `sparkle` acts on the whole art. About each second, two sparks appear for 1.5 seconds.
  They appear only on empty cells, so they stay in the dark ground and do not cover the lamp.
- The `candle` gives two flames on the last row of the page, beside the key `G`, in orange and yellow.
  The page coordinates work here: the screen is 40 columns wide, so its rows do not wrap.
- The `glint` acts on the shade: a rectangle of the art, the rows 2–4.
  About each 2 seconds, one cell of the shade becomes white and fades.
- The `sweep` runs once over the whole lamp, half a second after the show.
  It comes last, so it makes the stars and the light of the effects above it bright too.

## Result

- `velourbbs check` gives no output.
- `velourbbs preview main` gives the screen of task 9: the preview shows the screen
  before the effects.
- `velourbbs preview main --keys +1s` shows the screen after one second: the stars, the sparks,
  the flames and the band of the sweep in the middle of the lamp.
  `+1s,+1s` shows it after one second and after two seconds.
  The random values of the preview are the same at each run.
- Run `velourbbs serve`. PuTTY shows the light that shimmers, the stars, the sparks, the flames,
  the glints and the band.
