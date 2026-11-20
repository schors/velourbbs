# Task 9. Art

"ANSI art" is a picture of characters and colours.
In this task, you show art on a screen. Use a new BBS directory, for example `myart`.
The files are in [`examples/art/`](../../../../examples/art/). Write each file without `examples/art/`.

1. Create the directory `myart` and the directory `screens` in it.
2. Copy the file `examples/art/screens/lamp.ans` from the source repository
   to `myart/screens/lamp.ans`. Or ask the sysop for the file.
   This guide does not show the file: it holds escape sequences.

## The art insert

```
{>lamp.ans}
```

- `{><name>.ans}` shows the art `screens/<name>.ans`. The line must have no other text.
- Each row of the art starts at the left edge of the screen.
- The art does not change the style of the text after it.
- The BBS cuts art that is wider than the screen.
  `velourbbs check` warns when art is wider than the largest width of the screen.

Full rules: [bbt-format](../../../specs/bbt-format.md), sections 3.4 and 6.2.

## The files

`examples/art/bbs.toml`:
```
```

The file is empty: the BBS has the default settings.

`examples/art/screens/screens.toml`:
```
[theme.default]
normal = "223/235"
key    = "215 bold"
```

The style `key` is the 256-colour index 215. The orange 215 stays warm at 16 colours too: the BBS changes it to the nearest index 0–15.

`examples/art/screens/main.bbt`:
```
+++
width = [40, 40]
fill = "spaces"
[[item]]
key = "G"
do = "logoff"
+++
{>lamp.ans}

               {!key}G{!}  goodbye
```

- The art is 40 columns wide. `width = [40, 40]` in the front matter makes this screen
  40 columns wide.
- The keys `width`, `align`, `valign`, `fill` and `gutter` in the front matter set the layout of one screen.
  In `[screen]` of `screens/screens.toml`, they set the layout of all screens.
- Here `screens/screens.toml` has no `[screen]`. So the screen has the default `align = "center"`:
  on a terminal of 80 columns, the BBS puts the screen in the centre.

The screen:

```text

                ▄▄▄▄▄▄▄▄
             ▄████████████▄
          ▄██████████████████▄
       ░▒▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▒░
          ░░▒▒▒▒▒▒ ██ ▒▒▒▒▒▒░░
            ░░░░   ██   ░░░░
                   ██
               ▄▄▄▄██▄▄▄▄
            ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀

             lit for xterm

               G  goodbye
```

- The ground is dark grey. The shade of the lamp is orange with a cream top.
  The light under it is yellow and dark orange.
- The stem is grey-brown. The base is brown.
- The words "lit for" are grey-brown. The value of the field, "xterm", is cream.
  The key `G` is in the orange `key` in bold.

## The position of a screen

```
+++
width = [40, 40]
align = "left"
valign = "middle"
+++
```

- `align` is the position on a wider terminal: `"left"`, `"center"` (the default) or `"right"`.
- `valign` is the position of a screen with fewer rows than the terminal:
  `"top"` (the default), `"middle"` or `"bottom"`. A screen with a spacer line (task 2)
  fills the terminal, so `valign` has no effect on it.

A screen for terminals of 80 and 132 columns can grow with the terminal:

```
+++
width = [68, 108]
gutter = 6
+++
```

- `gutter` is the cells that the screen leaves on each side of the terminal: from 0 (the default) to 100.
  The width of the screen is the terminal width minus 2 gutters, in the limits of `width`.
- At 80 columns, the screen is 68 columns wide, with 6 empty columns on each side.
  At 132 columns, the screen is 108 columns wide, in the centre: 12 empty columns on each side.
- A stretch token (task 2) fills the width of the screen at each size.
  So `{~─}` gives a hairline of 68 cells at 80 columns and of 108 cells at 132 columns.

Full rules: [bbt-format](../../../specs/bbt-format.md), sections 6.1, 6.5 and 6.6.

## A field in art

Row 12 of `lamp.ans` holds the text `lit for {term.type    }`.
The BBS finds the field in the art and puts the value in its place.

- The field takes the cells of its token: here 15 cells. The BBS cuts a longer value.
- Spaces before the name align the value right. Else the value aligns left.
- All cells of the token must have one background colour.
  All its characters except spaces must have one foreground colour.
  Else `velourbbs check` gives an error (task 10).
- The value takes the colours of the first character of the token.
- Art can hold only fields and boxes (below): no style tokens and no includes.
  A short alias (task 5) gives a short field, for example `{w}`.

Full rules: [art-format](../../../specs/art-format.md), section 4.

## A list row in art

A list screen (task 8) can show its rows in art. Example: the screen `shelf`
shows the notes of `content/log/` in a frame.

The art `screens/shelf.ans`, 40 columns wide:

```text
╔══════════════════════════════════════╗
║ The keeper's log      page { n}/{c } ║
║ { row.num}  {row.title            }  ║
║                                      ║
║                                      ║
╚═[Enter]═[Space]═[Esc]════════════════╝
```

The aliases of `[fields]` in `screens/screens.toml` (task 5):

```
[fields]
n = "list.page"
c = "list.pages"
```

The screen `screens/shelf.bbt`:

```
+++
width = [40, 40]
data = { from = "files", args = { dir = "log" } }
rows = 3
select = { do = "gosub reader", args = { file = "{row.path}" } }
[[item]]
key = "SPACE"
do = "page next"
when = "not list.last"
[[item]]
key = "ESC"
do = "back"
+++
{>shelf.ans}
```

The screen:

```text
╔══════════════════════════════════════╗
║ The keeper's log      page    1/3    ║
║          1  The first night          ║
║          2  Rain on the roof         ║
║          3  A caller from the north  ║
╚═[Enter]═[Space]═[Esc]════════════════╝
```

Row 3 is the selected row: it is black on grey, the style `select`.
The other rows have the colours of the art.

- The art row with the `row.*` fields is the "list row". Art has at most one list row.
- The "list rectangle" starts at the first cell of the first `row.*` token
  and ends at the last cell of the last `row.*` token. It has `rows` art rows from the list row down.
  Here it has the columns 3–37 of the rows 3–5.
- Each rectangle row shows one row of the list page: the cells of the list row
  with the values of that row.
- Each cell of the rectangle below the list row must be a space. Else `velourbbs check`
  gives an error. A rectangle that ends below the art is an error too.
- The cells outside the rectangle stay: here the frame.
- A rectangle row without a row of the page is spaces in the colours of the list row.
- The selected row takes the theme style `select`. The default is `0/7`: black on grey.
  Set it in `[theme.default]`, for example `select = "235/215"`.
- A box cannot be in the list row, so `row.selected` does not work in art.
- A list screen needs a list line (task 8) or a list row in art.

Full rules: [list-view](../../../specs/list-view.md), section 9.

## A box in art

A "box" shows a part of an art row only when its condition is true.
Example: change the bottom row of `shelf.ans` to this row. It shows `[Space]`
only when the item `SPACE` works:

```text
╚═[Esc]═{#if item.SPACE}[Space]{/if}═══╝
```

The bottom row on pages 1 and 2, and on the last page 3:

```text
╚═[Esc]═[Space]                     ═══╝
╚═[Esc]═                            ═══╝
```

- A box goes from `{#if <condition>}` to `{/if}` on one art row.
  The condition is the condition of `{#if}` (task 5).
- When the condition is true, the box shows the text between the two tokens from the first cell
  of the box. The rest of the box is spaces.
- When the condition is false, the box is spaces.
- The spaces have the colours of the first cell of the box. The cells after the box do not move.
- The text of a box can hold fields.
- Boxes do not nest and have no `{#else}`. A box cannot be in the list row.
- Each other token that starts with `{#`, `{/`, `{!` or `{>` is an error in art,
  for example `{!title}`.

Full rules: [art-format](../../../specs/art-format.md), section 5.

## The colours of art

The colours 0–15 of art are the VGA colours: the colours of the art editors.
A terminal can show the colours 0–15 with other values. For example, Windows Terminal
shows the colour 1 as `c50f1f`, not as the VGA `aa0000`.

- At the connect, the BBS asks the terminal for its colours 0–15.
- When they differ from VGA, the BBS sends the VGA values of art as RGB colours.
  To a terminal with 256 colours, the BBS sends the nearest index 16–255.
- To a terminal with 16 colours, and when the terminal does not reply, the BBS sends
  the indexes 0–15. The terminal shows them in its own colours.
- The theme and the style tokens, for example `{!14}`, always give indexes.
  The terminal shows them in its own colours.

Full rules: [telnet-term](../../../specs/telnet-term.md), section 7.3.

## Art for small terminals

A screen can show other content on a narrow terminal. Example with art of 80 columns:

```
+++
width = [40, 80]
[[item]]
key = "G"
do = "logoff"
+++
{#if term.cols < 80}
{>_main_small}
{#else}
{>main.ans}
{/if}
```

On a terminal of fewer than 80 columns, the screen shows the fragment `screens/_main_small.bbt`:
text or smaller art.

When `[screen]` in `screens/screens.toml` has a `gutter`, add `gutter = 0` to the front matter.
So the art gets the full width of the terminal.

Full rules: [bbt-format](../../../specs/bbt-format.md), section 6.1.

## New art

The BBS reads art only in the "flat form". `velourbbs art import` converts the file
of an art editor to the flat form.

1. Draw the art in one of these editors. Save it with the default options:
   - Moebius 1.0.29: Save (`.ans`);
   - icy_draw 0.6.2: Export to `.ans`;
   - PabloDraw 3.3.14-beta: Save as ANSI (`.ans`).
2. Keep this "source file" outside `screens/`, for example in `art/logo.ans`.
3. Convert it:

```
velourbbs art import art/logo.ans screens/logo.ans
```

4. Insert the art with `{>logo.ans}`.

Moebius and PabloDraw cannot open the flat form. To change the art, edit the source file
and convert it again.

Full rules: [art-import](../../../specs/art-import.md), sections 2 and 3.

## Result

- `velourbbs check` gives no output.
- `velourbbs preview main` gives the screen above, with 20 columns of ground on each side.
- `velourbbs preview main --set term.type=ansi` gives the caption `lit for ansi`.
- Stop the server of Lamp. Run `velourbbs serve` in `myart`.
  PuTTY shows the lamp in the centre of the window.
