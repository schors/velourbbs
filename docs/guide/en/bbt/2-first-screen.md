# Task 2. The first screen

In this task, you write the first screen of Lamp. The menu stays: one item `G`. Task 4 adds the other items.

## The width and the styles

Create `screens/screens.toml` with this text:

```toml
[screen]
width = [80, 80]
fill = "spaces"

[theme.default]
normal = "#e9dccb/#2a201d"
dim    = "#9e8a7c"
line   = "#8a6a5c"
title  = "#f7c7a3 bold"
key    = "#ffb088 bold"
orn    = "#f2a27e"
```

- `width = [80, 80]` gives the smallest and the largest width of a screen.
  So the screen is always 80 columns wide. On a wider terminal, the BBS puts it in the centre.
  On a narrower terminal, it cuts the rows.
- `fill = "spaces"` fills the terminal rows below the screen with the style `normal`.
- `[theme.default]` is the "theme": it gives a name to each style.
  A style is a text colour, a background colour after `/` and the attributes, for example `bold`.
  A colour is `#rrggbb`, or a number from 0 to 255.
- `normal` is the style of the screen ground and of the plain text.
  A style without a background colour takes the background of `normal`.

Full rules: [screens-toml](../../../specs/screens-toml.md), sections 3.1 and 3.2,
and [bbt-format](../../../specs/bbt-format.md), sections 6.1–6.3, 6.5 and 7.

## Text and style tokens

- Each line of the body is one row of the screen. An empty line is an empty row.
  A line wider than the screen breaks at a space into more rows.
- The spaces at the start of a line stay.
- `{!name}` starts the style `name` of the theme. `{!}` starts the style `normal` again.
  A style continues on the next lines until the next style token.

Example:

```
       {!title}Lamp{!}  {!orn}☼ {!dim}a BBS by lamplight{!}
```

This line shows "Lamp" in the light peach `title` in bold, the sun `☼` in the warm orange `orn`
and the tagline in the grey-brown `dim`.

## The screen

Write `screens/main.bbt` again with this text:

```
+++
[[item]]
key = "G"
do = "logoff"
+++

  {!line}────────────────────────────────────────────────────────────────────────────{!}

       {!title}Lamp{!}  {!orn}☼ {!dim}a BBS by lamplight{!}

       Good evening. It's quiet and warm here.
       Pull up a chair by the lamp.











       {!dim}The light stays on till morning.{!}

  {!line}────────────────────────────────────────────────────────────────────────────{!}

    {!key}G{!}  {!dim}goodbye{!}
```

The screen has 23 rows:

```text

  ────────────────────────────────────────────────────────────────────────────

       Lamp  ☼ a BBS by lamplight

       Good evening. It's quiet and warm here.
       Pull up a chair by the lamp.











       The light stays on till morning.

  ────────────────────────────────────────────────────────────────────────────

    G  goodbye
```

- The ground is dark brown. The text is cream.
- The two lines are in the brown `line`. They take columns 3–78.
- The closing line and the word "goodbye" are in the grey-brown `dim`.
- The key `G` is in the orange `key` in bold.

Tera Term shows 256 colours. There the ground is dark grey.
Task 6 gives the styles for 256 colours.

## Result

- `velourbbs check` gives no output.
- `velourbbs preview main` gives 23 rows. Rows 2 and 4 have these style runs:

```
2: 1-2 #e9dccb/#2a201d, 3-78 #8a6a5c/#2a201d, 79-80 #e9dccb/#2a201d
4: 1-7 #e9dccb/#2a201d, 8-11 #f7c7a3/#2a201d bold, 12-13 #e9dccb/#2a201d, 14-15 #f2a27e/#2a201d, 16-33 #9e8a7c/#2a201d, 34-80 #e9dccb/#2a201d
```

- The telnet client shows the screen with the dark ground on all rows.

Save the output for task 3:

```
velourbbs preview main > task2.txt
```
