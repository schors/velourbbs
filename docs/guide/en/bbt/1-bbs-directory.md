# Task 1. The BBS directory

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
A screen must have at least one item. Only the screens `goodbye` and `shutdown` have no items (task 4).
The other lines are the "body": the text of the screen.
The BBS shows the screen `main` first.

## Check the files

Run:

```
velourbbs check
```

When the files are correct, the command gives no output.
Else it gives one line for each error or warning:

```
screens/_x.bbt:1:1: warning: no screen uses this file
screens/main.bbt:4:6: item "G": do: unknown action "logof"
```

The line gives the file, the line number, the column and the message.
An error in an item gives the position of the wrong value: here `"logof"` on line 4.
A warning has `warning:` before the message. A warning does not stop the BBS. An error stops it.
Full rules: [screen-render](../../../specs/screen-render.md), section 3.1.

## Look at the screen

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
Full rules: [screen-render](../../../specs/screen-render.md), sections 3.2 and 3.3.

## See the screen in a telnet client

1. Run `velourbbs serve`. The server listens on port 2323.
2. Connect a telnet client to `localhost`, port 2323. Use PuTTY or Tera Term:
   they show UTF-8. PuTTY shows 16777216 colours, Tera Term shows 256 colours.
   SyncTERM and NetRunner do not show UTF-8.
   In PuTTY, select the connection type "Telnet".
   Set "Close window on exit" to "Never": else PuTTY closes the window at the end of the session.
3. Press `G` to end the session.

The server reads the files only when it starts. After each change of a file,
stop the server with Ctrl+C and start it again.
Full rules: [telnet-term](../../../specs/telnet-term.md), sections 3 and 6.

## Result

- `velourbbs check` gives no output.
- The telnet client shows the two lines of text. `G` ends the session.
