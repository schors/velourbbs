# Art import

This specification defines the command `velourbbs art import`.
It converts ANSI art to the flat form ([art-format](art-format.md)).

The [BBT guide](bbt-guide.md) teaches this specification. Keep the guide in sync:
[bbt-guide](bbt-guide.md), section 3.

## 1. Goal

The sysop draws art in an art editor and saves it in the form of the editor.
`velourbbs art import` plays the file into a grid of cells once and writes the grid in the flat form.
The host never plays cursor moves and never reads CP437.

## 2. Command

`velourbbs art import <input> <output>` reads the file `<input>` and writes the file `<output>`.

- Each other count of arguments is a usage error ([screen-render](screen-render.md), section 3).
- `<input>` and `<output>` that name the same file are a usage error.
- The diagnostics have the form of [screen-render](screen-render.md), section 3.1, with the file `<input>`.
- After an error, the import writes no file and exits with status 1.
  An existing `<output>` stays unchanged.
- Without an error, the import writes `<output>`. It replaces an existing file.
  It writes the warnings to the standard error.
- When `<output>` is a symlink, the import writes the file that the link names.
  A symlink to a missing file is an error.
- An existing `<output>` keeps its mode. A new `<output>` gets the mode 0644.
- The source file is the file that the editor can edit. Moebius and PabloDraw cannot read
  the flat form. Edit the source file, then import it again.
  Keep the source file outside `screens/` ([bbt-format](bbt-format.md), section 2).

## 3. Target implementations

The import reads each form that these editors write with their default options:

| Editor | Forms |
|---|---|
| Moebius 1.0.29 (release, commit `5df8344`) | Save (`.ans`, CP437, SAUCE). |
| icy_draw 0.6.2 | Export to `.ans` on each compatibility level: ANSI.SYS, VT-100, IcyTerm, UTF-8. |
| PabloDraw 3.3.14-beta | Save as ANSI (`.ans`, CP437). |

Where they differ from ECMA-48, the import follows them.

These are not targets:

- "Export As UTF-8" of Moebius. A file with the extension `.utf8ans` is an error.
- The icy_draw export options other than the defaults: a line length limit,
  the line break "GotoXY", the control characters other than "Write unchanged".
- icy_draw documents with custom fonts, sixel images or the attributes faint, italic,
  concealed, crossed out or double underline (SGR `2`, `3`, `8`, `9`, `21`).

The editors write the CP437 glyphs of LF, CR, 0x1A and ESC as these controls
or as other glyphs. The import cannot restore these glyphs.
icy_draw at the level UTF-8 also writes the glyphs of 0x08, 0x09 and 0x7F as these controls,
and each is an error. Export such art at the level IcyTerm.

## 4. Reading

1. The SAUCE record and the COMNT block have the layout of [art-format](art-format.md), section 2, step 1.
   The input can have no SAUCE record. When `Comments` is > 0, the COMNT block must be
   at its place. The import does not check the SAUCE `FileSize`.
   Moebius pads each comment line to 64 characters (UTF-16 code units) and then writes
   UTF-8, so a non-ASCII comment moves the block. When the block is not at its place
   and the bytes after the first 0x1A start with `COMNT`, these bytes are the block:
   its text in lines of 64 characters. A block without lines (Moebius writes it
   for a comment of only spaces) is valid.
2. The data is the bytes before the COMNT block or the SAUCE record.
   The first 0x1A ends the data. A byte after it in the data is an error.
3. The input is UTF-8 when the data starts with a BOM. The import removes the BOM.
   Else the input is CP437.
4. UTF-8 input must be valid UTF-8. Invalid UTF-8 is one error, and the import stops.
   The play finds it, so an error that stops the play before it comes in its place.
   LF, CR and ESC are controls. Each other control character
   ([bbt-format](bbt-format.md), section 2.1) is an error.
5. In CP437 input, LF, CR and ESC are controls. Each other byte is a character
   by the table of section 7.
6. A "character SAUCE" has `DataType` 1 and `FileType` 1. The import uses `TInfo1`,
   `TInfo2` and `TFlags` only from a character SAUCE.
   The art width is `TInfo1` when it is > 0. Else the width is 80.
   A width over 255 or a `TInfo2` over 1000 is an error.
7. iCE colours are on when bit 0 of `TFlags` is 1.

The import checks each grapheme by [bbt-format](bbt-format.md), section 2.1.

## 5. Play

The import plays the data into a grid of cells. The grid width is the art width.

- The cursor starts at row 1, column 1. The style starts as `7/0`.
  Each cell that the play does not write is a space in `7/0`.
- When a grapheme fills the last column, a "wrap is pending" and the cursor stays.
  - LF clears the pending wrap and moves the cursor to column 1 of the next row.
  - A grapheme, CR or a cursor sequence first does the wrap: the cursor goes
    to column 1 of the next row.
  - The cursor sequences are `CSI` `A`–`D`, `H`, `f`, `s`, `u`, `ESC 7` and `ESC 8`.
    `CSI 2 J` clears the pending wrap. Each other sequence keeps it.
- CR moves the cursor to column 1.
- A wide grapheme that does not fit before the right edge does the wrap first.
- When wrap is off, a grapheme at the right edge overwrites the last cell.
  A wide grapheme overwrites the last two cells.
- A write to one half of a wide grapheme turns the other half into a space in its style.
- `CSI A`, `B`, `H` and `f` keep the cursor inside the grid width and the row limit
  ([art-format](art-format.md), section 6). `CSI C` and `D` change only the column, inside the grid width.
  A restore (`CSI u`, `ESC 8`) goes back to the saved row and column.
- In UTF-8 input, a wrap also resets the style to `7/0` without attributes.
  icy_draw at the level UTF-8 starts each row in this style and writes no SGR
  after a full row.
- The grid height is the last row with a written cell after the last `CSI 2 J`,
  or `TInfo2` of SAUCE if it is larger.

The import parses a CSI sequence by ECMA-48, section 5.4: parameter bytes, intermediate bytes,
a final byte. A missing or 0 parameter is 1 for `A`–`D`, `H`, `f` and `b`.
A parameter over 9999 counts as 9999.

| Sequence | Effect |
|---|---|
| `CSI n A`, `B`, `C`, `D` | Cursor up, down, right, left. |
| `CSI r;c H`, `CSI r;c f` | Cursor position. |
| `CSI 2 J` | Clear the grid. Cursor to row 1, column 1. |
| `CSI n K` | Erase in the row: 0 to the end, 1 to the start, 2 the whole row. |
| `CSI s`, `CSI u`, `ESC 7`, `ESC 8` | Save and restore the cursor. |
| `CSI n b` | Repeat the last grapheme n times. It stops at the row limit. |
| `CSI ? 33 h`, `CSI ? 33 l` | iCE colours on, off. |
| `CSI ? 7 h`, `CSI ? 7 l` | Wrap on, off. |
| `CSI 0 ; n SP D` | Font page n (icy_draw). No effect. |
| `CSI 0 ; r ; g ; b t`, `CSI 1 ; r ; g ; b t` | Background, foreground: RGB (PabloDraw). |
| `CSI ... m` | SGR, see below. |

- `CSI n K` and `CSI 2 J` write spaces with the foreground 7, the background of a written cell
  and no attributes.
- These are errors with the position of the cursor. The play goes on after them:
  - a CSI sequence that is not in the table;
  - an SGR parameter that is not in the table below;
  - an invalid colour in `38`, `48` or `t`;
  - an incomplete sequence: a byte outside the ranges of ECMA-48 ends it, and the byte plays;
  - ESC followed by a byte other than `[`, `7` or `8`. The byte plays.

| SGR | Effect |
|---|---|
| `0` | Reset to `7/0` without attributes. |
| `1`, `22` | Bold on, off. |
| `4`, `24` | Underline on, off. |
| `5`, `25` | Blink on, off. |
| `7`, `27` | Reverse on, off. |
| `30`–`37`, `90`–`97`, `39` | Foreground: index 0–7, index 8–15, index 7. |
| `40`–`47`, `100`–`107`, `49` | Background: index 0–7, index 8–15, index 0. |
| `38;5;n`, `48;5;n` | Foreground, background: index n. They take 3 parameters. |
| `38;2;r;g;b`, `48;2;r;g;b` | Foreground, background: RGB. They take 5 parameters. |

After an invalid `38` or `48`, the rest of the sequence does nothing.

The style of a written cell, in this order:

1. Bold with a foreground index 0–7 adds 8 to the index. Cells have no `bold` attribute.
2. Blink with iCE colours and a background index 0–7 adds 8 to the index.
   Blink without iCE colours sets the `blink` attribute.
3. Reverse swaps the foreground and the background.

## 6. Animation

A grapheme, `CSI n K` or `CSI 2 J` that writes a cell that a grapheme wrote before
is an "overwrite". The cell keeps the last write: the grid is the last frame.
The first overwrite gives one warning with the position of the cell.

## 7. CP437 table

- 0x00 is a space.
- 0x01–0x1F, except LF, CR, ESC and 0x1A, are the glyphs that the art editors show:

| Byte | Char | Byte | Char | Byte | Char | Byte | Char |
|---|---|---|---|---|---|---|---|
| 01 | U+263A ☺ | 09 | U+25CB ○ | 11 | U+25C4 ◄ | 19 | U+2193 ↓ |
| 02 | U+263B ☻ | 0B | U+2642 ♂ | 12 | U+2195 ↕ | 1C | U+221F ∟ |
| 03 | U+2665 ♥ | 0C | U+2640 ♀ | 13 | U+203C ‼ | 1D | U+2194 ↔ |
| 04 | U+2666 ♦ | 0E | U+266B ♫ | 14 | U+00B6 ¶ | 1E | U+25B2 ▲ |
| 05 | U+2663 ♣ | 0F | U+263C ☼ | 15 | U+00A7 § | 1F | U+25BC ▼ |
| 06 | U+2660 ♠ | 10 | U+25BA ► | 16 | U+25AC ▬ | | |
| 07 | U+2022 • | | | 17 | U+21A8 ↨ | | |
| 08 | U+25D8 ◘ | | | 18 | U+2191 ↑ | | |

- 0x20–0x7E are ASCII. 0x7F is U+2302 ⌂.
- 0x80–0xFF are the characters of code page 437 (`golang.org/x/text/encoding/charmap`
  `CodePage437` has the same table). 0xFF is U+00A0.

The import has its own table of 256 entries. `charmap` decodes 0x00–0x1F and 0x7F as controls.
The table is fixed data, so a dependency gives no gain.

## 8. Limits

These limits apply to the input: the file and the art grid ([art-format](art-format.md), section 6),
the grapheme ([bbt-format](bbt-format.md), section 9) and the diagnostics
([screen-render](screen-render.md), section 3.1). The play has one more limit:

- The play writes and erases at most 4080000 cells: 16 grids of 255 columns and 1000 rows.
  Then one error at the cursor, and the import stops.
- A grapheme counts its cells. `CSI n K` counts the erased cells. `CSI 2 J` counts the rows
  from row 1 to the last written row, times the art width.

## 9. Output

The import writes the grid in the flat form ([art-format](art-format.md)):

1. A BOM.
2. Each row of the grid, then LF. A row has no trailing cells that are a space in `7/0`.
   A wide grapheme is written once.
3. Before a cell with a style other than the current style, the import writes
   `ESC [ 0 ; <params> m` with the parameters of the style:
   - attributes: `4` underline, `5` blink;
   - the foreground: `30`–`37` for index 0–7, `90`–`97` for index 8–15,
     `38;5;n` for index 16–255, `38;2;r;g;b` for RGB; none for index 7;
   - the background: `40`–`47`, `100`–`107`, `48;5;n`, `48;2;r;g;b`; none for index 0.

   The style `7/0` without attributes is `ESC [ 0 m`. The current style starts as `7/0`
   and does not change at LF.
4. 0x1A and the SAUCE record (SAUCE 00) with a COMNT block when there are comments.

After the write, the import reads its output by [art-format](art-format.md), sections 2 and 3.
The fields of art-format, section 4, need `screens/screens.toml`, so `velourbbs check` checks them.
When that grid differs from the grid of the play, it is an error at the first different cell.
Example: two regional indicators in two cells join into one flag grapheme.
An output over the file limit ([art-format](art-format.md), section 6) is an error.

SAUCE fields of the output:

| Field | Value |
|---|---|
| `Title`, `Author`, `Group`, `Date`, comments | From the input SAUCE. Else empty. |
| `FileSize` | The bytes of the output before the 0x1A, the BOM included. |
| `DataType`, `FileType` | 1, 1. |
| `TInfo1`, `TInfo2` | The art width and the grid height. |
| `TFlags` | Bits 1–4 from a character SAUCE, else 0. Bit 0 is 0: the output has no iCE colours. |
| `TInfoS` | From a character SAUCE, else empty. |

- The import reads a text field of the input SAUCE as UTF-8 when it is valid UTF-8.
  Else it reads it by the table of section 7. Moebius and icy_draw write UTF-8,
  PabloDraw writes ASCII. A CP437 text that is valid UTF-8 reads wrong.
  The import removes trailing spaces and NUL bytes.
- The output text fields are UTF-8. `TInfoS` is padded with NUL bytes (a SAUCE ZString),
  the other fields with spaces. A field that is longer than its bytes is cut after the last
  whole grapheme that fits. The cut gives a warning. The cut comment lines give one warning.
- The output has at most 255 comment lines. The import keeps the first 255 lines
  and gives a warning.
