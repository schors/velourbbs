# Art format

This specification defines the art files `screens/<name>.ans` in the flat form.
Related parts: [bbt-format](bbt-format.md) and [art-import](art-import.md).

The [BBT guide](bbt-guide.md) teaches this specification. Keep the guide in sync:
[bbt-guide](bbt-guide.md), section 3.

## 1. Flat form

Art files are in the "flat form": UTF-8 graphemes, SGR, LF and SAUCE.
[art-import](art-import.md) converts the files of the art editors to the flat form.
The host reads only the flat form. Each other input is an error.

## 2. Reading

1. A SAUCE record (SAUCE 00) is the last 128 bytes of the file, when they start with `SAUCE00`.
   The COMNT block of `Comments` lines comes before it, when it starts with `COMNT`.
   The file must have a SAUCE record.
2. The SAUCE must have `DataType` 1 and `FileType` 1. The art width is `TInfo1`, 1–255.
   `TInfo2` is the least height, 0–1000. Bit 0 of `TFlags` (iCE colours) must be 0.
   The loader does not check the other fields.
3. The data is the bytes before the COMNT block or the SAUCE record. The loader removes
   one 0x1A at the end of the data.
4. The data must start with a BOM. The loader removes it.
5. The data must be valid UTF-8. Invalid UTF-8 is one error, and the load stops.
6. LF and ESC are controls. Each other control character
   ([bbt-format](bbt-format.md), section 2.1) is an error.

An error of steps 1, 2 or 4 is one error `not in the flat form, convert it with velourbbs art import`,
and the load stops.

## 3. Grid

The loader reads the data into a grid of cells. The grid width is the art width.
Cells, colours and styles are those of [bbt-format](bbt-format.md), section 7.1.
The notation `7/0` is the style syntax of section 7.2 of bbt-format.

- Each LF ends a row. The last row can have no LF.
- A row writes its graphemes from column 1. A row over the art width is an error.
- A grapheme follows the rules of [bbt-format](bbt-format.md), section 2.1.
- The grid height is the count of rows, or `TInfo2` if it is larger.
- Each cell that the data does not write is a space in `7/0`. Art has no colour `default`.
- The style starts as `7/0`. It does not change at LF.

The only sequence is SGR: `ESC [`, parameters, `m`. A parameter is digits;
`;` separates the parameters. A missing parameter is 0.
Each other sequence, an incomplete sequence and ESC without `[` are errors.

| SGR | Effect |
|---|---|
| `0` | Reset to `7/0` without attributes. |
| `4`, `24` | Underline on, off. |
| `5`, `25` | Blink on, off. |
| `30`–`37`, `90`–`97`, `39` | Foreground: index 0–7, index 8–15, index 7. |
| `40`–`47`, `100`–`107`, `49` | Background: index 0–7, index 8–15, index 0. |
| `38;5;n`, `48;5;n` | Foreground, background: index n. They take 3 parameters. |
| `38;2;r;g;b`, `48;2;r;g;b` | Foreground, background: RGB. They take 5 parameters. |

Each other SGR parameter and an invalid `38` or `48` are errors.
A cell takes the style of the SGR state. Art cells have no `bold` attribute.

## 4. Fields in art

- After the read, the loader finds fields in each grid row. A field token in art is `{`,
  optional spaces, a path ([bbt-format](bbt-format.md), section 4), optional spaces and `}`
  on one row. Each other `{` is a plain character.
- All cells of a token must have one background. All its non-space cells must have
  one foreground. An error gives both colours and their columns.
- A field takes the cells of its token and the style of its first non-space cell.
  It aligns right when the path follows spaces, else left.

## 5. Limits

| Content | Limit |
|---|---|
| `.ans` file | 1 MB. |
| Art grid | 255 columns, 1000 rows. |

A file or a part of a file over a limit is an error.

## 6. Not in the MVP

- Raw art and ANSI animation: [content-trust](https://github.com/schors/velourbbs/issues/5).
