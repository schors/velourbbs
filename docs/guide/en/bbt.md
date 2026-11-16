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

## Task 5. Fields and conditions

In this task, you write the screen `term`. It shows what the BBS knows about the terminal of the user.

### Fields

A "field" is a token that the BBS replaces with a value:

```
       {!dim}Size{!}     {term.cols}×{term.rows}
```

On a terminal of 80 columns and 24 rows, this line shows `Size     80×24`.

| Field | Value |
|---|---|
| `term.cols` | The columns of the terminal, from 20 to 255. |
| `term.rows` | The rows of the terminal, from 10 to 100. |
| `term.colors` | The colour depth: `16`, `256` or `16777216` (task 6). |
| `term.type` | The terminal type that the client sends, in lower case. Without a type, `ansi`. |
| `term.charset` | Always `utf-8`. |

- A field without spaces in its braces takes the width of its value.
- A field with spaces in its braces has a "fixed width": the width of the token with the braces.
  The BBS pads a short value with spaces and cuts the end of a long value.
  Spaces after the name align the value left. Spaces before the name align it right.

```
[{term.type      }]
[{      term.type}]
```

Each token is 17 cells wide. The BBS shows:

```text
[xterm            ]
[            xterm]
```

- `{` starts a token when a Latin letter, a space, `#`, `/`, `!`, `>` or `@` follows it.
  `{{` gives a plain `{`. Example: `Write {{name} here.` shows `Write {name} here.`

### Aliases

`[fields]` in `screens/screens.toml` gives a short name to a field, an "alias":

```toml
[fields]
w = "term.cols"
```

Then `{w}` shows the columns of the terminal. A short alias helps in art (task 7).

Full rules: [bbt-format](../../specs/bbt-format.md), sections 3.3 and 4,
and [screens-toml](../../specs/screens-toml.md), section 3.3.

### Conditions

```
{#if term.colors == 16777216}16 million{#else}{term.colors}{/if}
```

This text shows `16 million` when the terminal shows 16777216 colours. Else it shows the number.

- `{#if <condition>}` shows the text up to `{#else}` or `{/if}` when the condition is true.
- `{#else}` shows the text up to `{/if}` when the condition is false. `{#else}` is optional.
- A block can span lines and can hold other blocks. Each `{#if}` needs its `{/if}` in the same file.
- A line with `{#if}`, `{#else}` or `{/if}` gives no row when only spaces stay on it.
- The operators are `==`, `!=`, `<`, `<=`, `>` and `>=`.
  A number field takes a number and each operator.
  A text field takes text in double quotes and only `==` or `!=`: `term.type == "xterm"`.
- `and` and `or` join conditions. `and` binds tighter. There are no parentheses.

Full rules: [bbt-format](../../specs/bbt-format.md), sections 5 and 6.2.

### Items with conditions

`when` gives a condition to an item. When the condition is false, the item does not work.
`{#if item.<KEY>}` is true when the item `<KEY>` works. So the text of an item and its key agree:

```
+++
[[item]]
key = "W"
do = "goto wide"
when = "term.cols >= 132"
[[item]]
key = "G"
do = "logoff"
+++
{#if item.W}W  wide screen{/if}
G  goodbye
```

On a terminal of 132 columns or more, the screen shows `W  wide screen`, and `W` works.
On a narrower terminal, the line gives no row, and the BBS ignores `W`.
`not item.W` is true when the item does not work.

Full rules: [bbt-format](../../specs/bbt-format.md), section 8.2.

### The screen `term`

Write `screens/term.bbt` again with this text. It replaces the copy of `about.bbt` from task 4.

`examples/lamp/en/screens/term.bbt`:
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
       {!dim}Size{!}     {term.cols}×{term.rows}
       {!dim}Colours{!}  {#if term.colors == 16777216}16 million{#else}{term.colors}{/if}
       {!dim}Type{!}     {term.type}

       This is how the station sees you.










  {!line}────────────────────────────────────────────────────────────────────────────{!}

    {!key}Esc Enter{!}  {!dim}back{!}
```

The screen at 80 columns and 24 rows:

```text

  ────────────────────────────────────────────────────────────────────────────

       Lamp  ☼ a BBS by lamplight

       Size     80×24
       Colours  256
       Type     xterm

       This is how the station sees you.










  ────────────────────────────────────────────────────────────────────────────

    Esc Enter  back
```

The labels "Size", "Colours" and "Type" are in the grey-brown `dim`. The values are in the cream `normal`.

### Result

- `velourbbs check` gives no output.
- `velourbbs preview term` gives the rows `Size     80×24`, `Colours  256` and `Type     xterm`.
- `velourbbs preview term --set term.type=ansi` gives `Colours  16` and `Type     ansi`.
- `velourbbs preview term --set term.colors=16777216` gives `Colours  16 million`.
- In the telnet client, `T` shows the size of the window. Change the width of the window,
  keep at least 24 rows, and press `Esc` and `T` again: the screen shows the new size.
  With fewer rows, the screen comes in pages (task 6), and a key only shows the next page.

## Task 6. The colour depth

### The depth of a terminal

A terminal shows 16, 256 or 16777216 colours: its "colour depth".
The BBS finds the depth from the terminal type that the client sends.
The first row of the table that matches the type gives the depth:

| Terminal type | Depth |
|---|---|
| Ends with `-direct`, or has `truecolor` or `24bit` | 16777216 |
| Has `256color` | 256 |
| Ends with `-16color` or `-8color`, or is `ansi`, `dumb`, `vt52`, `vt100`, `vt102`, `vt220`, `vt320`, `rxvt`, `xterm-color`, `eterm-color` or `cygwin` | 16 |
| Each other type, for example `xterm` | 256 |

- A client without a type gets `ansi`, so 16 colours.
- PuTTY and Tera Term send `xterm`, so 256 colours.

The BBS changes each colour to the nearest colour of the depth.
At 256 colours, the cream text `#e9dccb` becomes the neutral grey 253, and the screen loses its warmth.
So choose the colours for each depth by hand.
The indexes 16–255 are the colours of the xterm 256-colour palette.
The indexes 0–15 are the VGA colours of [telnet-term](../../specs/telnet-term.md), section 7.3.

Full rules: [telnet-term](../../specs/telnet-term.md), sections 6 and 7.3.

### Depth themes

```toml
[theme.256]
normal = "223/235"
title  = "230 bold"

[theme.16]
normal = "7/0"
title  = "15 bold"
```

- A "depth theme" gives other styles for one depth.
  `[theme.256]` is for 256 colours. `[theme.16]` is for 16 colours.
- A style of a depth theme replaces the style with the same name of `[theme.default]`.
  The other styles stay.
- Each name of a depth theme must also be in `[theme.default]`, or be `normal` or `more`.
- At 16 colours, use the indexes 0–15. At 256 colours, use the indexes 16–255.

Full rules: [screens-toml](../../specs/screens-toml.md), section 3.2.

### The More row

A screen with more rows than the terminal comes in pages. Each page except the last ends
with the "More row". The user presses any key for the next page.
Items do not work on the More row.

```toml
[screen]
more = "    more ∙ any key"
```

`more` gives the text of the More row. Without `more`, the text is `-- More --`.
The style `more` gives its colours.
With the file `screens/screens.toml` below, at 20 rows, the screen `about` has two pages:

```
velourbbs preview about --size 80x20
```

```
page 1 80x20 fill spaces
...
|    more ∙ any key                                                              |
styles
...
page 2 80x20 fill spaces
...
```

Page 1 has rows 1–19 of the screen and the More row. Page 2 has rows 20–23.

Full rules: [bbt-format](../../specs/bbt-format.md), section 6.4,
and [menu-nav](../../specs/menu-nav.md), section 5.

### The file `screens.toml`

Write `screens/screens.toml` again with this text:

`examples/lamp/en/screens/screens.toml`:
```
[screen]
width = [80, 80]
more = "    more ∙ any key"
fill = "spaces"

[theme.default]
normal = "#e9dccb/#2a201d"
dim    = "#9e8a7c"
line   = "#8a6a5c"
title  = "#f7c7a3 bold"
key    = "#ffb088 bold"
orn    = "#f2a27e"
more   = "#9e8a7c"

[theme.256]
normal = "223/235"
dim    = "137"
line   = "95"
title  = "230 bold"
key    = "216 bold"
orn    = "216"
more   = "137"

[theme.16]
normal = "7/0"
dim    = "8"
line   = "8"
title  = "15 bold"
key    = "15 bold"
orn    = "7"
more   = "8"
```

- At 256 colours, the ground is the dark grey 235, and the text is the cream 223, not the grey 253.
- At 16 colours, the screen is grey: the black ground, the light grey text,
  the dark grey `dim` and `line`, the white `title` and `key`.

### Look at each depth

`velourbbs preview` uses the type `xterm`, so 256 colours. Set another depth with `--set`:

```
velourbbs preview main --set term.colors=16
velourbbs preview main --set term.type=vt100
```

Both commands show the screen with the styles of `[theme.16]`.
The dump shows the colours of the theme as you wrote them.
Only the telnet client shows the nearest colours of its depth.

### Result

- `velourbbs check` gives no output.
- `velourbbs preview main` gives row 4 with the styles of `[theme.256]`:

```
4: 1-7 223/235, 8-11 230/235 bold, 12-13 223/235, 14-15 216/235, 16-33 137/235, 34-80 223/235
```

- `velourbbs preview main --set term.colors=16` gives row 4 with the styles of `[theme.16]`:

```
4: 1-7 7/0, 8-11 15/0 bold, 12-15 7/0, 16-33 8/0, 34-80 7/0
```

- PuTTY shows the dark grey ground and the cream text, not grey text.
  In PuTTY, set Connection > Data > "Terminal-type string" to `vt100` and connect again:
  PuTTY shows the grey screen of 16 colours.
  Then set the "Terminal-type string" back to `xterm`.

## Task 7. Art

"ANSI art" is a picture of characters and colours.
In this task, you show art on a screen. Use a new BBS directory, for example `myart`.
The files are in [`examples/art/`](../../../examples/art/). Write each file without `examples/art/`.

1. Create the directory `myart` and the directory `screens` in it.
2. Copy the file `examples/art/screens/lamp.ans` from the source repository
   to `myart/screens/lamp.ans`. Or ask the sysop for the file.
   This guide does not show the file: it holds escape sequences.

### The art insert

```
{>lamp.ans}
```

- `{><name>.ans}` shows the art `screens/<name>.ans`. The line must have no other text.
- Each row of the art starts at the left edge of the screen.
- The art does not change the style of the text after it.
- The BBS cuts art that is wider than the screen.
  `velourbbs check` warns when art is wider than the largest width of the screen.

Full rules: [bbt-format](../../specs/bbt-format.md), sections 3.4 and 6.2.

### The files

`examples/art/bbs.toml`:
```
```

The style `key` is a 256-colour index:

```toml
key    = "215 bold"
```

The orange 215 stays warm at 16 colours too: the BBS changes it to the nearest index 0–15.

`examples/art/screens/screens.toml`:
```
[theme.default]
normal = "223/235"
key    = "215 bold"
```

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

### The position of a screen

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

Full rules: [bbt-format](../../specs/bbt-format.md), sections 6.1 and 6.5.

### A field in art

Row 12 of `lamp.ans` holds the text `lit for {term.type    }`.
The BBS finds the field in the art and puts the value in its place.

- The field takes the cells of its token: here 15 cells. The BBS cuts a longer value.
- Spaces before the name align the value right. Else the value aligns left.
- All cells of the token must have one background colour.
  All its characters except spaces must have one foreground colour.
  Else `velourbbs check` gives an error (task 8).
- The value takes the colours of the first character of the token.
- Art can hold only fields: no style tokens, no conditions and no includes.
  A short alias (task 5) gives a short field, for example `{w}`.

Full rules: [art-format](../../specs/art-format.md), section 4.

### Art for small terminals

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

Full rules: [bbt-format](../../specs/bbt-format.md), section 6.1.

### New art

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

Full rules: [art-import](../../specs/art-import.md), sections 2 and 3.

### Result

- `velourbbs check` gives no output.
- `velourbbs preview main` gives the screen above, with 20 columns of ground on each side.
- `velourbbs preview main --set term.type=ansi` gives the caption `lit for ansi`.
- Stop the server of Lamp. Run `velourbbs serve` in `myart`.
  PuTTY shows the lamp in the centre of the window.

## Task 8. Errors

`velourbbs check` finds the errors before the BBS starts.
Each line gives the file, the line number, the column and the message (task 1).
In art, the line and the column are the row and the column of the art.
This task gives frequent errors and their fixes.

### `unknown style`

```
screens/about.bbt:10:8: unknown style "titel"
```

The style token `{!titel}` names a style that is not in the theme.
Correct the name, or add the style to `[theme.default]`.

### `token has no "}"`

```
screens/about.bbt:10:35: token has no "}"
```

A `{` starts a token (task 5). Close the token with `}`, or write `{{` for a plain `{`.

### `unknown field`

```
screens/about.bbt:11:13: unknown field "term.colours"
screens/about.bbt:14:10: unknown field "note"
screens/term.bbt:14:67: unknown field "item.S"
```

- The name is not a field or an alias (task 5). Correct the name.
- For text in braces, write `{{note}`. In art, `{{` does not help:
  change the text in the art editor, then convert the art again.
- `item.S` needs an item with the key `S` in the same screen.

### `{#if} has no {/if}`

```
screens/about.bbt:12:8: {#if} has no {/if}
```

Add `{/if}`. A block ends in the file where it starts.

### `control character U+0009`

```
screens/about.bbt:13:1: control character U+0009
```

The line holds a TAB. Replace it with spaces. Set your editor to write spaces for the Tab key.

### A condition

```
screens/term.bbt:14:8: expected an integer or a string, found "xterm"
screens/term.bbt:14:38: integer field "term.colors" needs an operator
```

- `{#if term.type == xterm}`: write text in double quotes: `term.type == "xterm"`.
- `{#if term.colors = 256}`: compare with `==`.

### `item: must be an array of tables`

```
screens/main.bbt:2:1: item: must be an array of tables
screens/main.bbt:5:2: table item already exists
screens/main.bbt:5:2: table item already exists as an array of tables
screens/main.bbt:5:3: key item already exists as a table, but should be an array table
```

The front matter has `[item]`. Write `[[item]]`.
The first message comes from a screen with one item. The other messages give the line of a table.

### `unknown screen`

```
screens/main.bbt:2:1: item "T": do: unknown screen "term"
```

`goto term` needs the file `screens/term.bbt`. Create the file, or correct the name.

### `screen has no items`

```
screens/x.bbt:1:1: screen has no items
```

Each screen except `goodbye` needs at least one item. Add an item.

### `not in the flat form`

```
screens/logo.ans:1:1: not in the flat form, convert it with velourbbs art import
```

The art is the file of an art editor. Convert it (task 7).

### A field in art changes the colour

```
screens/size.ans:1:7: token "{w}" changes the foreground: 11 at col 7, 14 at col 9
```

The characters of the token have two colours: 11 in column 7 and 14 in column 9.
In the art editor, draw the whole token in one colour. Then convert the art again.

### Result

1. In `screens/about.bbt` of Lamp, line 10, put `{!titel}` before `Lamp`.
2. Run `velourbbs check`. It gives:

```
screens/about.bbt:10:8: unknown style "titel"
```

3. Remove `{!titel}`. `velourbbs check` gives no output.

Full rules: [screen-render](../../specs/screen-render.md), section 3.1,
and [menu-nav](../../specs/menu-nav.md), section 6.

## Not in the MVP

The BBS does not have these features yet:

- lists and loops (`#each`), list rows in art, data providers and conditions in art.
  In the MVP, only the `.bbt` files have conditions;
- marks (`{@<id>}`) and the choice of an item with the arrow keys.
  The user chooses an item only with its key;
- access checks (`require`). They need users;
- the arguments of an action (`args`, `ask`, `{arg.*}`);
- punctuation keys, for example `?`;
- a label of an item. The BBS does not print the text of an item;
- the display codes of other BBS programs: pipe codes, `@` codes and MCI codes.
  There is no converter.

Full rules: [bbt-format](../../specs/bbt-format.md), section 10.
