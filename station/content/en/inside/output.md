# Output and input

## What it does

Velour draws the screen in memory as a grid of cells and sends the terminal the bytes for this grid.
In the other direction, it turns the bytes from the terminal into keys.

## How

### Cells

A cell holds one grapheme and a style. A grapheme is what a person sees as one sign:
a letter with an accent, a flag, an emoji of several code points
([UAX #29](https://www.unicode.org/reports/tr29/)). The width of a grapheme is 0, 1 or 2 cells,
by [Unicode 17](https://www.unicode.org/versions/Unicode17.0.0/).
The characters of width Ambiguous (Cyrillic and lines) take one cell.

### Page

Velour sends each page in one write:

1. `ESC[?2026h`: the start of synchronized output
   ([mode 2026](https://contour-terminal.org/vt-extensions/synchronized-output/)).
   The terminal collects the output and shows the page at once, not row by row.
   A terminal without this mode ignores it.
2. `ESC[0m ESC[2J`: reset the style and clear the screen. With `fill = "erase"`, the style
   `normal` comes in place of `ESC[0m`: the terminal clears the screen in the background of the theme.
3. For each row: `ESC[<row>;1H`, then the cells of the row.
   There are no line ends, so the terminal does not scroll.
   The exception is in the part "Complex graphemes".
4. `ESC[0m ESC[?2026l`: reset the style and end synchronized output.

The cursor position, the screen clear and the styles are commands of ECMA-48
([ECMA-48](https://ecma-international.org/publications-and-standards/standards/ecma-48/)).

### Styles

Before a cell with a new style, Velour sends SGR: `ESC[0;<parameters>m`.
The parameters: `1` bold, `4` underline, `5` blink; `30`-`37` and `90`-`97` for 16 colours,
`38;5;n` for 256, `38;2;r;g;b` for RGB. The background has the same numbers plus 10: `40`, `48;5;n`.

Velour replaces a colour that the terminal does not have with the nearest one: the smallest
squared RGB distance. The 256 colours are the xterm palette, the 16 colours are the VGA colours.
So one screen looks right in truecolor, in 256 and in 16 colours.

### Complex graphemes

Terminals count the width of a grapheme of several code points in different ways.
For example, `❤️` with VS16 takes one cell in PuTTY, xterm, Tera Term and VTE.
So after such a grapheme, Velour prints a space and puts the cursor in its place:
`ESC[<row>;<column>H`. The next cells go to their columns at each width.
But a grapheme that the terminal draws wider can pass the right edge and wrap.
On the last row, such a wrap scrolls the page.

### Update

A live field changes each second, and an effect changes cells while it runs. Then Velour
does not draw the page again, but sends only the changed runs:

1. `ESC[?2026h ESC[?25l`: synchronized output and a hidden cursor.
2. For each run: `ESC[<row>;<column>H` and its cells.
   Two runs with 3 or fewer equal cells between them are one run.
3. `ESC[0m`, the cursor back to its place, `ESC[?25h ESC[?2026l`.

Frames come at most once in 50 ms. When nothing changed, Velour sends nothing.

### Input

The bytes from the client go through four steps:

1. The telnet parser.
2. Enter: CR LF, CR NUL, CR and LF give one key `ENTER`.
3. The UTF-8 decoder ([RFC 3629](https://www.rfc-editor.org/rfc/rfc3629)).
   It drops invalid bytes.
4. The key decoder.

The key decoder knows the sequences of the arrows, Home, End, PgUp, PgDn, Del and F1-F12
in the forms of xterm, VT and the Linux console. After ESC and after the byte that starts
a sequence (`[`, `O`, `]`, `P`, `_`), it waits 150 ms for the next byte. Then it waits
up to 1 second for each byte. When no byte comes after ESC, this is the key Esc. After a longer pause,
the started sequence goes away: so a reply that arrives in parts does not become keys.
It drops OSC, DCS and APC strings as a whole: a late probe reply does not become keys either.

In the line input, Velour turns on bracketed paste: `ESC[?2004h`
([xterm](https://invisible-island.net/xterm/ctlseqs/ctlseqs.html)).
The terminal puts the marks `ESC[200~` and `ESC[201~` around a paste.
So a paste does not press menu items: outside the line input, Velour drops it.

## Limits

- PuTTY 0.83, Tera Term, xterm and tmux before 3.7 do not know mode 2026.
  They show the cells as they arrive. The cursor is hidden and does not run over the screen.
- The fill `fill = "erase"` needs bce. tmux and GNU screen do not have it,
  and the empty rows keep the colour of the terminal.
- Velour drops keys with modifiers, for example Ctrl+arrow (`ESC[1;5A`).
  Alt+letter gives Esc and the letter, except Alt+P.
- Velour drops a sequence longer than 128 bytes as a whole.
- Terminals show blink in different ways, and PuTTY and iTerm2 turn it off by default.
  So Velour has effects in place of blink.
- There is no mouse.
