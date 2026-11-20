# Task 10. Errors

`velourbbs check` finds the errors before the BBS starts.
Each line gives the file, the line number, the column and the message (task 1).
In art, the line and the column are the row and the column of the art.
This task gives frequent errors and their fixes.

## `unknown style`

```
screens/about.bbt:10:8: unknown style "titel"
```

The style token `{!titel}` names a style that is not in the theme.
Correct the name, or add the style to `[theme.default]`.

## `token has no "}"`

```
screens/about.bbt:10:35: token has no "}"
```

A `{` starts a token (task 5). Close the token with `}`, or write `{{` for a plain `{`.

## `unknown field`

```
screens/about.bbt:11:13: unknown field "term.colours"
screens/about.bbt:14:10: unknown field "note"
screens/term.bbt:14:67: unknown field "item.S"
```

- The name is not a field or an alias (task 5). Correct the name.
- For text in braces, write `{{note}`. In art, `{{` does not help:
  change the text in the art editor, then convert the art again.
- `item.S` needs an item with the key `S` in the same screen.

## `{#if} has no {/if}`

```
screens/about.bbt:12:8: {#if} has no {/if}
```

Add `{/if}`. A block ends in the file where it starts.

## `control character U+0009`

```
screens/about.bbt:13:1: control character U+0009
```

The line holds a TAB. Replace it with spaces. Set your editor to write spaces for the Tab key.

## A condition

```
screens/term.bbt:14:8: expected an integer or a string, found "xterm"
screens/term.bbt:14:38: integer field "term.colors" needs an operator
```

- `{#if term.type == xterm}`: write text in double quotes: `term.type == "xterm"`.
- `{#if term.colors = 256}`: compare with `==`.

## `item: must be an array of tables`

```
screens/main.bbt:2:1: item: must be an array of tables
screens/main.bbt:5:2: table item already exists
screens/main.bbt:5:2: table item already exists as an array of tables
screens/main.bbt:5:3: key item already exists as a table, but should be an array table
```

The front matter has `[item]`. Write `[[item]]`.
The first message comes from a screen with one item. The other messages give the line of a table.

## `unknown screen`

```
screens/main.bbt:2:1: item "T": do: unknown screen "term"
```

`gosub term` needs the file `screens/term.bbt`. Create the file, or correct the name.

## `screen has no items`

```
screens/x.bbt:1:1: screen has no items
```

Each screen except `goodbye` and `shutdown` needs at least one item. Add an item.

## `a user cannot reach this screen`

```
screens/old.bbt:1:1: warning: a user cannot reach this screen
```

`velourbbs check` follows each `goto`, `gosub`, `select` and `lang` from the first screen `main`.
No path goes to the screen `old`.
Add an item that opens the screen, for example `do = "gosub old"` in `main`.
Or delete the file and each item that opens it.

- Screens that only open each other also get the warning.
- An item with `when` is a path.
- `goodbye` and `shutdown` get no warning.
- In a BBS in two languages (task 7), the check follows each language.
  `screens/en/menu.bbt` has an item `gosub news`, and `screens/ru/menu.bbt` does not.
  Then the common file `screens/news.bbt` gets the warning for `ru`:
  `screens/news.bbt:1:1: warning: a user cannot reach this screen (language "ru")`.
  Add the item `gosub news` to `screens/ru/menu.bbt`.
- A warning does not stop the BBS.

## `content file ... does not exist`

```
screens/code.bbt:2:1: file: content file "src/lamp.c" does not exist
screens/texts.bbt:2:1: item "1": args: file: content file "lmap.md" does not exist
```

- The screen `code` holds the snippet of task 7 ("A source file").
  Its key `file` names `src/lamp.c`, and `content/src/lamp.c` does not exist.
  Put the file at this path in `content/`, or correct the path.
- The item `1` gives `gosub reader` the file `lmap.md` in `args`, and the file is not in `content/`.
  Correct the name in `args`.

## `directory ... does not exist`

```
screens/log.bbt:2:1: data: args: dir: directory "lgo" does not exist
```

The key `data` of the list screen `log` names the directory `lgo` (task 8),
and `content/lgo/` does not exist. Correct the name, or create the directory.

## `front matter: no end line`

```
content/lamp.md:1:1: front matter: no end line "+++"
```

The first line of the file is `+++`, and no other line is `+++`.
Add the line `+++` after the last key of the front matter (task 7).

## `not in the flat form`

```
screens/logo.ans:1:1: not in the flat form, convert it with velourbbs art import
```

The art is the file of an art editor. Convert it (task 9).

## A field in art changes the colour

```
screens/size.ans:1:7: token "{w}" changes the foreground: 11 at col 7, 14 at col 9
```

The characters of the token have two colours: 11 in column 7 and 14 in column 9.
In the art editor, draw the whole token in one colour. Then convert the art again.

## Layout at a terminal size

`velourbbs check --size <cols>x<rows>` renders each screen at this size and warns about layout mistakes.
Give `--size` for each terminal size that you expect:

```
velourbbs check --size 80x24 --size 132x43
```

Lamp gives no warning at these sizes. Its screens have `width = [80, 80]`,
so a terminal of 40 columns cuts their rows:

```
screens/about.bbt:10:1: warning: at 40x24: the row is 43 cells wide, the screen shows 40
screens/_head.bbt:2:1: warning: screen "main": at 40x24: the row is 78 cells wide, the screen shows 40
```

The warnings and their fixes:

| Warning | Problem | Fix |
|---|---|---|
| `the line wraps to 2 rows` | The text line is wider than the screen. | Make the line shorter. |
| `31 rows, the screen gets a More row` | The screen has more rows than the terminal. | Remove rows, or split the screen. |
| `25 rows, the rows after 24 are cut` | A file screen or a list screen does not fit the terminal. | Remove rows above or below the file area or the list. |
| `the row is 43 cells wide, the screen shows 40` | The BBS cuts the row. | Make the row shorter, or make `min` of `width` smaller. |
| `art "logo.ans" is 100 cells wide, the screen shows 80` | The BBS cuts the art. | Show smaller art on a small terminal (task 9). |

- A warning in a fragment starts with the screen that includes it: `screen "main": `.
- A warning does not stop the BBS.

## Result

1. In `screens/about.bbt` of Lamp, line 10, put `{!titel}` before `Lamp`.
2. Run `velourbbs check`. It gives:

```
screens/about.bbt:10:8: unknown style "titel"
```

3. Remove `{!titel}`. `velourbbs check` gives no output.

Full rules: [screen-render](../../../specs/screen-render.md), sections 3.1 and 3.6,
[menu-nav](../../../specs/menu-nav.md), section 6,
[file-view](../../../specs/file-view.md), section 7,
and [list-view](../../../specs/list-view.md), section 10.
