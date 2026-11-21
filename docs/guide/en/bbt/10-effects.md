# Task 10. Effects

An "effect" changes cells of a screen while the screen waits for a key.
In this task, you add effects to the lamp of task 9: the light of the lamp shimmers,
stars twinkle around it, and two small flames burn beside the key.
Use the directory `myart` of task 9.

## The table `[[effect]]`

Each effect is one table `[[effect]]` in the front matter:

```
[[effect]]
do = "twinkle"
at = [37, 3]
chars = "*+·"
```

- `do` names the effect. This task shows `twinkle` and `candle`.
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
| `delay` | `0` | The seconds before the first change: 0–3600. |
| `every` | Of the effect. | The seconds between two changes: 0.05–3600. |
| `spread` | Half of `every`. | A random part of `every`: each time is `every` plus or minus up to `spread`. |

- A value is a number of seconds, for example `0.5`.
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
do = "candle"
at = [[13, -1], [28, -1]]
chars = "▲▴"
colors = ["208", "214", "220"]
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
- The `candle` gives two flames on the last row of the page, beside the key `G`, in orange and yellow.
  The page coordinates work here: the screen is 40 columns wide, so its rows do not wrap.

## Result

- `velourbbs check` gives no output.
- `velourbbs preview main` gives the screen of task 9: the preview shows the screen
  before the effects.
- `velourbbs preview main --keys +1s` shows the screen after one second: the stars and the flames.
  `+1s,+1s` shows it after one second and after two seconds.
  The random values of the preview are the same at each run.
- Run `velourbbs serve`. PuTTY shows the light that shimmers, the stars and the flames.
