# BBT guide

This guide teaches you to write the screens of a Velour BBS.
It is for the artist and the screen author.
You need a text editor, a terminal and a telnet client.

## Introduction

### The BBS directory

A "BBS directory" holds the files of one BBS:

| File | Content |
|---|---|
| `bbs.toml` | The settings of the host. In this guide, it is empty. |
| `screens/screens.toml` | The settings of the screens: the width, the styles and other values. |
| `screens/<name>.bbt` | A "screen": a text file that the BBS shows. |
| `screens/_<name>.bbt` | A "fragment": a part of a screen. Its name starts with `_`. |

Run each `velourbbs` command in the BBS directory.

### The program `velourbbs`

Build the program with Go. In the source repository of Velour BBS, run:

```
go build ./cmd/velourbbs
```

The command writes the file `velourbbs` (`velourbbs.exe` on Windows).
Put it in a directory of your `PATH`. Or ask the sysop for the program.

### The example

The tasks build the example BBS "Lamp". Its files are in
[`examples/lamp/en/`](../../../examples/lamp/en/).

- A line with a path and a colon shows a whole file of the example.
  Example: `` `examples/lamp/en/screens/main.bbt`: ``.
  Write this file in your BBS directory without `examples/lamp/en/`: `screens/main.bbt`.
- Each other block shows a part of a file, a command or an output.

Each task refers to the specifications for the full rules.

## Task 1. The BBS directory

1. Create a directory, for example `mybbs`. Create the directory `screens` in it.
2. Create an empty file `bbs.toml`.
3. Create `screens/main.bbt` with this text:

```
+++
[[item]]
key = "G"
do = "logoff"
+++
Hello! This is my BBS.

G  goodbye
```

The lines between the two `+++` lines are the "front matter". It gives the menu of the screen.
Here the menu has one item: the key `G` ends the session (`logoff`).
A screen must have at least one item. Only the screen `goodbye` has no items (task 4).
The other lines are the "body": the text of the screen.
The BBS shows the screen `main` first.

### Check the files

Run:

```
velourbbs check
```

When the files are correct, the command gives no output.
Else it gives one line for each error or warning:

```
screens/_x.bbt:1:1: warning: no screen uses this file
screens/main.bbt:2:1: item "G": do: unknown action "logof"
```

The line gives the file, the line number, the column and the message.
An error in an item gives line 2: the first line of the front matter.
A warning has `warning:` before the message. A warning does not stop the BBS. An error stops it.
Full rules: [screen-render](../../specs/screen-render.md), section 3.1.

### Look at the screen

Run:

```
velourbbs preview main
```

The command gives a "cell dump" of the screen at 80 columns and 24 rows:

```
page 1 80x24 fill none
|Hello! This is my BBS.                                                          |
|                                                                                |
|G  goodbye                                                                      |
styles
1: 1-80 7/0
2: 1-80 7/0
3: 1-80 7/0
```

- The first line gives the page, the terminal size and the fill of the rows below the screen.
- Each row of the screen is between two `|`.
- Each line after `styles` gives a row number and the "style runs" of the row.
  A run gives the first column, the last column and the style. `7/0` is light grey on black.

Use `--size` for other sizes, for example `velourbbs preview main --size 132x43`.
Full rules: [screen-render](../../specs/screen-render.md), sections 3.2 and 3.3.

### See the screen in a telnet client

1. Run `velourbbs serve`. The server listens on port 2323.
2. Connect a telnet client to `localhost`, port 2323. Use PuTTY or Tera Term:
   they show UTF-8 and 256 colours. SyncTERM and NetRunner do not show UTF-8.
   In PuTTY, select the connection type "Telnet".
   Set "Close window on exit" to "Never": else PuTTY closes the window at the end of the session.
3. Press `G` to end the session.

The server reads the files only when it starts. After each change of a file,
stop the server with Ctrl+C and start it again.
Full rules: [telnet-term](../../specs/telnet-term.md), sections 3 and 6.

### Result

- `velourbbs check` gives no output.
- The telnet client shows the two lines of text. `G` ends the session.

## Task 2. The first screen

In this task, you write the first screen of Lamp. The menu stays: one item `G`. Task 4 adds the other items.

### The width and the styles

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

Full rules: [screens-toml](../../specs/screens-toml.md), sections 3.1 and 3.2,
and [bbt-format](../../specs/bbt-format.md), sections 6.1–6.3, 6.5 and 7.

### Text and style tokens

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

### The screen

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

PuTTY and Tera Term show 256 colours. There the ground is dark grey.
Task 6 gives the styles for 256 colours.

### Result

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

## Task 3. The header

Each screen of Lamp starts with the same 5 rows: the "header".
Write them once in a fragment, and include the fragment in each screen.

- A fragment is `screens/_<name>.bbt`. It has no front matter.
- The line `{>_<name>}` includes the fragment: the BBS puts the lines of the fragment in its place.
  This line must have no other text.
- After the include, the style is the style before the include.

Full rules: [bbt-format](../../specs/bbt-format.md), sections 3.3 and 3.4.

In `screens/main.bbt`, replace rows 1–5 with one line:

```
+++
[[item]]
key = "G"
do = "logoff"
+++
{>_head}
       Good evening. It's quiet and warm here.
```

Move the 5 rows to the fragment:

`examples/lamp/en/screens/_head.bbt`:
```

  {!line}────────────────────────────────────────────────────────────────────────────{!}

       {!title}Lamp{!}  {!orn}☼ {!dim}a BBS by lamplight{!}

```

### Result

- `velourbbs check` gives no output.
- `velourbbs preview main` gives the same output as in task 2. To compare, run:

```
velourbbs preview main > task3.txt
diff task2.txt task3.txt
```

`diff` gives no output. On Windows, use `fc task2.txt task3.txt`.

## Task 4. A menu with hotkeys

### Items

Each `[[item]]` in the front matter is one item of the menu:

```
[[item]]
key = "A"
do = "goto about"
```

- `key` is the key: a Latin letter, a digit or a named key, for example `ESC` or `ENTER`.
  Case does not matter. A Russian letter works as the Latin letter on the same key: `ф` is `A`.
- `do` is the action: `goto <screen>` shows the screen, `logoff` ends the session.
- The BBS does not show the items. Write the text of the items and the key row in the body.
  The key row shows all keys of the screen.

Full rules: [bbt-format](../../specs/bbt-format.md), section 8,
and [menu-nav](../../specs/menu-nav.md), sections 3–6.

### The screen `main`

The menu has three items: `A` goes to the screen `about`, `T` to the screen `term`,
`G` ends the session.

`examples/lamp/en/screens/main.bbt`:
```
+++
[[item]]
key = "A"
do = "goto about"
[[item]]
key = "T"
do = "goto term"
[[item]]
key = "G"
do = "logoff"
+++
{>_head}
       Good evening. It's quiet and warm here.
       Pull up a chair by the lamp.


         {!key}A{!}   About                                   {!dim}who we are{!}

         {!key}T{!}   Your terminal                           {!dim}size and colours{!}

         {!key}G{!}   Goodbye                                 {!dim}the lamp will wait{!}




       {!dim}The light stays on till morning.{!}

  {!line}────────────────────────────────────────────────────────────────────────────{!}

    {!key}A T G{!}  {!dim}choose{!}
```

The screen:

```text

  ────────────────────────────────────────────────────────────────────────────

       Lamp  ☼ a BBS by lamplight

       Good evening. It's quiet and warm here.
       Pull up a chair by the lamp.


         A   About                                   who we are

         T   Your terminal                           size and colours

         G   Goodbye                                 the lamp will wait




       The light stays on till morning.

  ────────────────────────────────────────────────────────────────────────────

    A T G  choose
```

The keys are in the orange `key`. The notes on the right and the word "choose" are in the grey-brown `dim`.

### The screen `about`

The keys `Esc` and `Enter` go back to `main`:

```
[[item]]
key = "ESC"
do = "goto main"
```

`examples/lamp/en/screens/about.bbt`:
```
+++
[[item]]
key = "ESC"
do = "goto main"
[[item]]
key = "ENTER"
do = "goto main"
+++
{>_head}
       Lamp is a small station for friends.
       Come in, look around, stay a while.

       The station runs on Velour BBS.
       Its screens are plain text files.










  {!line}────────────────────────────────────────────────────────────────────────────{!}

    {!key}Esc Enter{!}  {!dim}back{!}
```

The screen:

```text

  ────────────────────────────────────────────────────────────────────────────

       Lamp  ☼ a BBS by lamplight

       Lamp is a small station for friends.
       Come in, look around, stay a while.

       The station runs on Velour BBS.
       Its screens are plain text files.










  ────────────────────────────────────────────────────────────────────────────

    Esc Enter  back
```

The keys `Esc Enter` are in the orange `key`. The word "back" is in the grey-brown `dim`.

The screen `term` comes in task 5. Until then, copy `screens/about.bbt` to `screens/term.bbt`.
Else `velourbbs check` gives the error `item "T": do: unknown screen "term"`.

### The screen `goodbye`

After `logoff`, the BBS shows the screen `goodbye` and ends the session.
`goodbye` has no front matter and no items. It has no key row.

```
       {!title}See you soon.{!}
       The lamp stays on for you.
```

`examples/lamp/en/screens/goodbye.bbt`:
```
{>_head}
       {!title}See you soon.{!}
       The lamp stays on for you.













  {!line}────────────────────────────────────────────────────────────────────────────{!}
```

The screen has 21 rows:

```text

  ────────────────────────────────────────────────────────────────────────────

       Lamp  ☼ a BBS by lamplight

       See you soon.
       The lamp stays on for you.













  ────────────────────────────────────────────────────────────────────────────
```

"See you soon." is in the light peach `title` in bold.

### Test the keys

`velourbbs preview <screen> --keys <keys>` presses the keys and shows each screen.
Separate the keys with commas:

```
velourbbs preview main --keys A,ESC,G
```

The output has an event line before each cell dump:

```
screen main
page 1 80x24 fill spaces
...
key A
screen about
page 1 80x24 fill spaces
...
key ESC
screen main
page 1 80x24 fill spaces
...
key G
logoff
screen goodbye
page 1 80x24 fill spaces
...
```

Full rules: [menu-nav](../../specs/menu-nav.md), section 7.

### Result

- `velourbbs check` gives no output.
- `velourbbs preview main --keys A,ESC,G` gives the event lines above.
- `velourbbs preview goodbye` gives the 21 rows above.
- In the telnet client, `A` shows `about`, `Esc` goes back, `G` shows `goodbye` and ends the session.
