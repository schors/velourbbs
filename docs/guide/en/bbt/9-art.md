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
- The keys `width`, `align`, `valign` and `fill` in the front matter set the layout of one screen.
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
  `"top"` (the default), `"middle"` or `"bottom"`.

Full rules: [bbt-format](../../../specs/bbt-format.md), sections 6.1 and 6.5.

## A field in art

Row 12 of `lamp.ans` holds the text `lit for {term.type    }`.
The BBS finds the field in the art and puts the value in its place.

- The field takes the cells of its token: here 15 cells. The BBS cuts a longer value.
- Spaces before the name align the value right. Else the value aligns left.
- All cells of the token must have one background colour.
  All its characters except spaces must have one foreground colour.
  Else `velourbbs check` gives an error (task 10).
- The value takes the colours of the first character of the token.
- Art can hold only fields: no style tokens, no conditions and no includes.
  A short alias (task 5) gives a short field, for example `{w}`.

Full rules: [art-format](../../../specs/art-format.md), section 4.

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
