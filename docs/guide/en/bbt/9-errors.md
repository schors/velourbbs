# Task 9. Errors

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

The art is the file of an art editor. Convert it (task 8).

## A field in art changes the colour

```
screens/size.ans:1:7: token "{w}" changes the foreground: 11 at col 7, 14 at col 9
```

The characters of the token have two colours: 11 in column 7 and 14 in column 9.
In the art editor, draw the whole token in one colour. Then convert the art again.

## Result

1. In `screens/about.bbt` of Lamp, line 10, put `{!titel}` before `Lamp`.
2. Run `velourbbs check`. It gives:

```
screens/about.bbt:10:8: unknown style "titel"
```

3. Remove `{!titel}`. `velourbbs check` gives no output.

Full rules: [screen-render](../../../specs/screen-render.md), section 3.1,
[menu-nav](../../../specs/menu-nav.md), section 6,
and [file-view](../../../specs/file-view.md), section 7.
