# Texts and art

## What it does

Velour shows texts in Markdown and ANSI art. It reads the texts as they are.
The SysOp converts art once to the form of the server with the command `velourbbs art import`.

## How: texts

The library `goldmark` reads Markdown: CommonMark 0.31.2
([specification](https://spec.commonmark.org/0.31.2/)) and the GFM tables
([GFM tables](https://github.github.com/gfm/#tables-extension-)).
It gives a document tree without HTML, and Velour lays it out in cells:

- headings, paragraphs, lists, quotes, code and tables, each in its own theme style;
- a long paragraph wraps by words to the width of the column, a code row is cut with the mark `...`;
- a link with an address gets a note number `[1]`, and the addresses are a list at the end of the text.

The text splits into pages by the height of the terminal: `Space` goes forward, `B` back.
The reading time is the number of words divided by 200.

The guides are in `docs/guide/` of the repository. The station sees them through a symlink
and does not copy them. So a guide on the station and on GitHub is one and the same file.

## How: art

ANSI art is drawn in Moebius 1.0.29, icy_draw 0.6.2 and PabloDraw 3.3.14-beta.
Usually they write an `.ans` file in CP437 with a SAUCE record at the end: 128 bytes
with the width, the author and the flags ([SAUCE](https://www.acid.org/info/sauce/sauce.htm)).

`velourbbs art import` plays the file into a grid of cells, as a terminal would do:
cursor moves, wraps, erases. It takes the width from `TInfo1` of the SAUCE record,
without it 80. Bit 0 of `TFlags` turns on iCE colours: a bright background in place of blink.
Then the command writes the grid in the flat form: a BOM, rows in UTF-8 and between them only SGR styles.

The server reads only flat art in UTF-8. It never plays cursor moves
and never reads CP437. A screen inserts art among the text: `{>logo.ans}`.

## What it gets

- A text that looks the same at 80 and at 132 columns and in each colour depth.
- Art that the server outputs as usual cells: it can stand next to text,
  and effects can animate its stars and candles.
- Velour sends the VGA colours of art as RGB when the palette of the terminal is not VGA.
  So art keeps the colours of its artist.

## Limits

- Links to other Markdown files do not open yet.
- The front matter of a text is only TOML. YAML is an error.
- The editors write the CP437 glyphs at the places of LF, CR, 0x1A and ESC as controls.
  The import cannot restore them.
- "Export As UTF-8" of Moebius and the custom fonts of icy_draw are not supported.
- There is no animation: the grid is the last frame. The import warns at the first overwrite of a cell.
- The play writes at most 4080000 cells: 16 grids of 255 columns and 1000 rows.
