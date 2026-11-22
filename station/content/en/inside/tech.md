# Technology

Velour is written in Go. It is one program: it takes the guests by telnet itself,
draws the screens and keeps the data. The source is open:
[github.com/schors/velourbbs](https://github.com/schors/velourbbs).

## Libraries

Apart from the Go standard library, Velour takes only three:

- `go-toml` reads the settings and the screens. It gives the line and the column
  of an unknown key, so `velourbbs check` shows the place of the error.
- `displaywidth` splits the text into graphemes and knows the width of each. It has Unicode 17.
- `goldmark` reads Markdown: CommonMark and the GFM tables. It has no dependencies of its own.

## Own code

Velour does the rest itself, and each decision has a reason:

- Telnet. The Go libraries do not give a server NAWS and TTYPE with bounded buffers.
  The own parser has about 200 lines.
- PROXY protocol. The library has API v0.x, and its 2026 releases
  changed the default behaviour. The own receiver of versions 1 and 2 has about 200 lines.
- Terminal output. `tcell` draws the full screen at once, but Velour shows pages
  with a More row. Also, `tcell` does not detect the terminal at connect.
- The ANSI art parser. The program needs only the colours, and the import about 15 sequences.
  A full terminal parser is a large dependency.
- The conditions of the screens. Their grammar has 7 rules. An expression language
  would add arithmetic and functions that the format excludes.

## Data

The data of the station is in one JSON file. The sysop reads and edits it in a text editor.
Velour does not take SQLite: its file needs a separate program to read,
and SQLite adds 6 MB to the program.

## Live screen

A screen lives while it waits for a key, with no script.
A live field changes in its place each second. Velour does not draw the rest of the screen again.
So the clock ticks in the bottom row of the menu and the sections. On the screen "About",
the time online, the open sessions and the uptime of the station change in the same way.

Effects animate parts of the screen. There are six: stars twinkle, stars appear and go,
a candle flickers, a letter glints, a light runs along a row, lightning strikes.
This station uses the running light, `sweep`: it runs over the logotype at the entry
and over the line under the header of the menu.
Velour sends only the changed cells to the terminal, and the frames are at least 50 ms apart.

## Art

ANSI art is drawn in Moebius, icy_draw and PabloDraw. These editors usually write CP437 files.
`velourbbs art import` converts an editor file to UTF-8 once.
The server reads only UTF-8 and never reads CP437.
A screen puts the art among the text.
