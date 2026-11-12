# Screen format

This specification defines the screen files, the screen model, the layout
and the commands `bbs check` and `bbs preview`.
Related parts: [menu-nav](../ideas/menu-nav.md) and [telnet-term](../ideas/telnet-term.md).

## 1. Goal

The host loads the screen files into an in-memory screen model at start.
It renders a screen for one terminal of any size. 80x25 is not a limit.
`bbs check` finds each error in the files. `bbs preview` renders a screen to a "cell dump".

The host parses the screen files at load. There is no compile step and no cache.
Screen files are untrusted data: a sysop can install them from other authors without a review.
The limits of section 10 bound the memory and the time for each file and each render.

## 2. Files

| Path | Content |
|---|---|
| `screens/<name>.bbt` | A "screen": a "BBS template" that the host shows. Section 3. |
| `screens/_<name>.bbt` | A "fragment": a BBS template that an include inserts. Section 3.4. |
| `screens/<name>.ans` | ANSI art. Section 8. |
| `bbs.toml` | Host settings. This specification defines its tables `[screen]` and `[theme]`. |

- `<name>` matches `[a-z0-9][a-z0-9_-]*`. Example: `main_wide`.
- A "narrow screen" is a screen that the `narrow` key of another screen names (section 6.1).
- An `.ans` file is art only. Only an art insert in a BBS template shows it.

### 2.1. Characters

- A "grapheme" is an extended grapheme cluster (Unicode UAX #29).
  A grapheme of more than 32 bytes is an error in a file.
- The width of a grapheme is 0, 1 or 2 cells, by Unicode 17. East Asian Ambiguous characters
  have the width 1. Terminals that show them as wide are not in the MVP (telnet-term).
- A grapheme of width 0 is an error in a file. Examples: U+200B, a combining mark
  at the start of a line.
- The "control characters" are U+0000–U+001F, U+007F–U+009F, U+202A–U+202E
  and U+2066–U+2069.

## 3. BBS template

### 3.1. Structure

- The file is UTF-8. The loader removes a BOM. Invalid UTF-8 is an error.
- A line ends with LF or CR LF.
- If the first line is `+++`, the lines up to the next `+++` line are the "front matter".
  Else the file has no front matter. A fragment must have no front matter.
- The lines after the front matter are the "body". A control character in the body is an error.
  TAB is a control character.

### 3.2. Front matter

The front matter is TOML 1.1. An unknown key is an error.

| Key | Type | Meaning |
|---|---|---|
| `width` | array of 2 integers | `[min, max]`, with 1 <= `min` <= `max` <= 255. Section 6.1. |
| `narrow` | string | The narrow screen. Section 6.1. |
| `[fields]` | table of strings | Aliases. Section 4.2. |
| `[[item]]` | array of tables | Menu items with `key`, `do` and an optional `when`. |

Example:

```
+++
width = [64, 100]
narrow = "main_small"
[fields]
w = "term.cols"
[[item]]
key = "A"
do = "goto about"
+++
{#if term.cols >= 100}
{>main_wide.ans}
{#else}
{>main.ans}
{/if}
{!title}Welcome! Your terminal has {w} columns.
{!prompt}[A]bout [G]oodbye
```

[menu-nav](../ideas/menu-nav.md) defines `key` and `do`, and the checks for them.
This specification parses `when` as a condition (section 5) and keeps `key` and `do` as strings.

### 3.3. Tokens

`{` starts a token when a letter, a space, `#`, `/`, `!`, `>` or `@` follows it.
Each other `{` is a plain character. In the body, `{{` gives a plain `{`.
A token ends at the first `}` on the same line. A token without `}` is an error.
Section 8.4 has the rule for art.

| Token | Meaning |
|---|---|
| `{<path>}` | Field with its natural width. Section 4. |
| `{<path>      }` | Field with the width of the token, aligned left. |
| `{      <path>}` | Field with the width of the token, aligned right. |
| `{#if <cond>}` `{#else}` `{/if}` | Conditional block. Section 5. |
| `{!<style>}` | Style change. Section 7.3. |
| `{>_<name>}` | Include of `screens/_<name>.bbt`. Section 3.4. |
| `{><name>.ans}` | Art insert of `screens/<name>.ans`. Section 6.2. |
| `{@<id>}` `{/@}` | Reserved for marks. An error in the MVP. |

- The width of a fixed-width field includes the braces. Example: `{term.type      }` is 17 cells.
- A field token with spaces on both sides of the path is an error.
- "Block tokens" are `{#if}`, `{#else}` and `{/if}`. Blocks can nest and can span lines.
  `{#else}` is optional. The blocks of each file must balance.

### 3.4. Includes and art inserts

- An include or an art insert must be the only content of its source line.
- An include inserts the body lines of the fragment in place of its line.
- An include of a name without `_` is an error. A missing file is an error.
- Includes can nest. An include cycle is an error.
- After an include, the style is the style before the include.

## 4. Fields

### 4.1. Registry

| Path | Type | Value |
|---|---|---|
| `term.type` | string | Terminal type. |
| `term.cols` | integer | Terminal columns. |
| `term.rows` | integer | Terminal rows. |
| `term.charset` | string | Output charset, `utf-8` in the MVP. |
| `term.colors` | integer | Colour depth: `16`, `256` or `16777216`. |
| `item.<KEY>` | boolean | True when the item with the key `<KEY>` exists and is enabled. |

- A path matches `[a-z][a-z0-9_]*(\.[A-Za-z0-9_]+)?`. A path without a dot is an alias.
- [telnet-term](../ideas/telnet-term.md) gives the `term.*` values.
- `item.<KEY>` is valid only in `{#if}`. [menu-nav](../ideas/menu-nav.md) defines `<KEY>`.
- An unknown path is an error.
- A field token shows an integer in decimal and a string as it is.

### 4.2. Aliases

- `[fields]` maps an alias to a registry path. Example: `w = "term.cols"`.
- An alias cannot be `term`, `item`, `and`, `or` or `not`.
- The path of an alias cannot be `item.<KEY>`.
- An alias is valid in the screen, in its fragments and in its art.
  Short aliases let art have short fields.

### 4.3. Values

- The host inserts a value after the parse, so a value cannot make a token.
- The host removes the control characters and the graphemes of width 0 from a value.
  It replaces a grapheme of more than 32 bytes with U+FFFD.
- A fixed-width field pads its value with spaces or cuts it at a grapheme boundary.
  A wide grapheme that does not fit becomes a space.

## 5. Conditions

`{#if}` and `when` use one grammar:

```
cond    = and { "or" and } ;
and     = unary { "and" unary } ;
unary   = "not" path | path [ op literal ] ;
op      = "==" | "!=" | "<" | "<=" | ">" | ">=" ;
literal = integer | string ;
integer = [ "-" ] digit { digit } ;
string  = '"' { any character except '"' and '}' } '"' ;
```

- Spaces between the parts are optional. Example: `term.cols>=100`.
- `and` binds tighter than `or`. There are no parentheses.
- A path without an operator and a path after `not` must be boolean.
- An integer field takes an integer literal and each operator.
- A string field takes a string literal and only `==` or `!=`.
- Each other combination is an error.
- The host evaluates the conditions of a screen at each render.

## 6. Layout

### 6.1. Width

- The "render width" W is `term.cols`, clamped to the "width limits" `[min, max]`.
- The width limits come from the first source of this list:
  1. `width` in the front matter;
  2. `[a, a]`, when each art that the screen can insert has the width `a`;
  3. `[screen] width` in `bbs.toml`. Its default is `[40, 80]`.
- A screen with art of different widths and without `width` is an error.
- When `term.cols` > W, the "screen area" has a left margin of `(term.cols - W) / 2` cells,
  rounded down.
- When `term.cols` < `min` and `narrow` is set, the host shows the narrow screen instead.
  It renders the narrow screen with its own width limits.
- When `term.cols` < `min` and `narrow` is not set, the host renders at W = `min`
  and cuts each row at `term.cols`.
- A narrow screen with items or with `narrow` is an error. `bbs check` checks
  its `item.<KEY>` paths against the items of each screen that names it.

### 6.2. Rows

The body gives rows in source order:

- The content of a false block gives nothing: no text, no art and no include.
- A source line with a block token gives no row when it has only spaces after the evaluation.
- An art insert gives the rows of the art. Each row starts at the left edge of the screen area.
  The host cuts the art at W. `bbs check` warns when the art is wider than `max`.
- Each other source line gives one or more rows (section 6.3). An empty line gives an empty row.
- Each cell that the screen does not draw is a space in the theme style `normal`.

### 6.3. Text wrap

- A row has W cells. A text line breaks at a space. The spaces at a break are removed.
  The spaces at the start of the source line stay.
- A word that is wider than W breaks at a grapheme boundary.
- A wide grapheme that does not fit at the end of a row goes to the next row.
- A fixed-width field is one word. If it is wider than W, the host cuts it at W.
- The style continues on the next row.

### 6.4. Height

- A screen has no height limit.
- When a screen has more rows than `term.rows`, the host shows it in pages of `term.rows - 1` rows.
  After each page except the last, the host shows the "More row" and waits for any key.
- The More row is `[screen] more` in `bbs.toml`, in the theme style `more`,
  at the left edge of the screen area. The default text is `-- More --`.
- The host sends no line end after the last row of a page, so the terminal does not scroll.
- The model keeps the file, the line and the column of each token. `bbs check` uses them.

## 7. Colours and styles

### 7.1. Colour model

- A cell holds one grapheme and a style, never encoded bytes. A wide grapheme takes 2 cells.
- A "colour" is `default` (the colour of the terminal), an index from 0 to 255, or RGB.
  Indexes 0–7 are the standard colours, 8–15 the bright colours,
  16–255 the xterm 256-colour palette.
- A "style" is a foreground colour, a background colour and the attributes
  `bold`, `underline` and `blink`.
- [telnet-term](../ideas/telnet-term.md) reduces each colour to the colour depth of the terminal.

### 7.2. Style syntax

```
style  = colour [ "/" colour ] { " " attr } | "/" colour { " " attr } ;
colour = integer | "#" hex hex hex hex hex hex | "-" ;
attr   = "bold" | "underline" | "blink" ;
```

- An integer colour is an index from 0 to 255. `-` is the colour `default`.
- Examples: `14`, `14/1`, `/1`, `#ff8800/#000020`, `15/4 bold`, `-/-`.

### 7.3. Theme and style tokens

- `[theme]` in `bbs.toml` maps a style name to a style. A name matches `[a-z][a-z0-9_]*`.
- The theme always has `normal` and `more`. Their defaults are `7/0` and `15/0`.
- A theme style without a foreground or a background takes it from `normal`.
- `{!<name>}` sets the whole theme style `<name>`. An unknown name is an error.
- `{!14}`, `{!14/1}` and `{!/1}` set only the colours that they name.
  The other colour and the attributes stay. A style token cannot have attributes.
- `{!}` sets the theme style `normal`.
- Each screen starts with `normal`. The style continues across lines and blocks.
  A style token in a false block has no effect.
- An art insert does not change the style of the text after it.

Example `bbs.toml`:

```toml
[theme]
normal = "7/0"
title  = "15/4 bold"
prompt = "14"
```

## 8. Art

### 8.1. Target implementations

Art files are UTF-8 only. The target implementations are these art editors
and their UTF-8 save forms:

| Editor | Save form |
|---|---|
| Moebius 1.0.29 | "Export As UTF-8": `38;5;n` colours, an LF before each row, no BOM and no SAUCE. |
| icy_draw 0.6.2 | Level UTF-8: a BOM and `38;2` colours. |

The loader accepts each form that they write. Where they differ from ECMA-48,
the loader follows them. A CP437 file is an error.
[art-convert](../ideas/art-convert.md) converts CP437 art to UTF-8.

### 8.2. Reading

1. A "SAUCE record" (SAUCE 00) is the last 128 bytes of the file, when they start with `SAUCE00`.
   The COMNT block of `Comments` lines comes before it, when it starts with `COMNT`.
2. The data is the bytes before the COMNT block or the SAUCE record. The loader removes
   one 0x1A at the end of the data.
3. The data must be valid UTF-8. The loader removes a BOM.
4. LF, CR and ESC are controls. Each other control character is an error.
5. The art width is `TInfo1` when `DataType` is 1, `FileType` is 1 and `TInfo1` > 0.
   Else the width is 80. A width over 255 is an error.
6. The Moebius form has no BOM and no SAUCE. For a file without both, an LF before
   the first grapheme does nothing. Its width is the widest row without wrap, at least 80.
7. iCE colours are on when bit 0 of `TFlags` is 1.

### 8.3. Grid

The loader plays the data into a grid of cells. The grid width is the art width.
The play time is linear in the file size plus the grid cells.

- The cursor starts at row 1, column 1. The style starts as `7/0`.
  Each cell that the play does not write is a space in `7/0`. Art has no colour `default`.
- When a grapheme fills the last column, a "wrap is pending" and the cursor stays.
  - LF clears the pending wrap and moves the cursor to column 1 of the next row.
  - A grapheme, CR or a cursor sequence first does the wrap: the cursor goes
    to column 1 of the next row.
  - SGR and mode sequences keep the wrap pending.
- CR moves the cursor to column 1.
- A wide grapheme that does not fit before the right edge does the wrap first.
- When wrap is off, a grapheme at the right edge overwrites the last cell.
- A cursor move stays inside the grid width and the row limit (section 10).
- The grid height is the last row with a written cell, or `TInfo2` of SAUCE if it is larger.

The loader parses a CSI sequence by ECMA-48, section 5.4: parameter bytes, intermediate bytes,
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
| `CSI ... m` | SGR, see below. |

- Each other CSI sequence does nothing. Example: `CSI 0;1 SP D` (icy_draw font page).
  `bbs check` warns with its position.
- ESC followed by a byte other than `[`, `7` or `8`: the loader drops the ESC
  and plays the byte, as Moebius does.

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

Each other SGR parameter does nothing. `bbs check` warns with its position.

The style of a written cell, in this order:

1. Bold with a foreground index 0–7 adds 8 to the index. Art cells have no `bold` attribute.
2. Blink with iCE colours and a background index 0–7 adds 8 to the index.
   Blink without iCE colours sets the `blink` attribute.
3. Reverse swaps the foreground and the background.

### 8.4. Fields in art

- After the play, the loader finds fields in each grid row. A field token in art is `{`,
  optional spaces, a path, optional spaces and `}` on one row. Each other `{` is a plain character.
- All cells of a token must have one background. All its non-space cells must have
  one foreground. An error gives both colours and their columns.
- A field takes the cells of its token and the style of its first non-space cell.
  It aligns right when the path follows spaces, else left.

## 9. Commands

### 9.1. `bbs check`

- `bbs check` reads `bbs.toml` and each file in `screens/`.
  It checks each screen with the fragments and the art that the screen uses.
- It warns about a fragment or an art file that no screen uses.
- It reports all errors and warnings, not only the first. It exits with status 1 after an error.
- An error has the form `<file>:<row>:<col>: <message>`, with 1-based numbers.
  For art, the row and the column are grid positions, as the art editor shows them.
  A warning has `warning: ` before the message.
- The host runs the same checks at start. An error stops the start.

Examples:

```
screens/main.bbt:9:1: item "S": when: unknown field "term.colours"
screens/main.ans:4:15: token "{w}" changes the foreground: 11 at col 15, 15 at col 16
screens/about.bbt:3:1: warning: art "logo.ans" is 132 cells wide, max width is 100
```

### 9.2. `bbs preview`

`bbs preview <name> [--size <cols>x<rows>] [--set <path>=<value>]...` renders a screen
and writes its cell dump to the standard output.

- The default size is 80x25. `--size` is the only way to set the size.
- `--set` sets `term.type` or `term.colors`. Example: `--set term.colors=256`.
  The defaults are `ansi` and 16. `term.charset` is `utf-8`.
- `item.<KEY>` comes from the items of the screen and their `when`.

### 9.3. Cell dump

The cell dump is the golden file for the tests. It has one part for each page.
This dump shows the screen `+++` `[[item]]` `key = "G"` `do = "logoff"` `+++`
with the body `{!title}Welcome!{!}` and `{!14}[G]oodbye`, and the theme of section 7.3:

```
page 1/1 80x25
|Welcome!                                                                        |
|[G]oodbye                                                                       |
styles
1: 1-8 15/4 bold, 9-80 7/0
2: 1-9 14/0, 10-80 7/0
```

- The first line gives the page number, the page count and the terminal size.
- Each rendered row of the page is between two `|`. A wide grapheme appears once
  and takes 2 columns.
- Each line of `styles` gives the row number and the runs of equal styles,
  in the syntax of section 7.2 with 1-based columns.
  A run has the form `<first>-<last> <style>`, also for one column. Example: `5-5 7/0`.
- The dump has the colours of the model, before the reduction to `term.colors`.
- A page with a More row includes it as its last row.

## 10. Limits

| Item | Limit |
|---|---|
| `.bbt` or `.ans` file | 1 MB. |
| Art grid | 255 columns, 1000 rows. |
| Body of a screen after includes | 10000 lines. An art insert counts its rows. |
| Rendered screen | 10000 rows. |
| Grapheme | 32 bytes. |

- An item of a file over a limit is an error. The loader counts the lines while it expands
  the includes, and stops at the limit.
- The render stops at the row limit.

## 11. Code

| Package | Content |
|---|---|
| `cmd/bbs` | The commands `check` and `preview`. |
| `internal/bbt` | The BBS template parser, the front matter and the includes. |
| `internal/art` | The `.ans` loader and SAUCE. |
| `internal/screen` | Cells, colours, styles, the token scanner, the registry, conditions, layout, render, cell dump. |
| `internal/config` | `bbs.toml`. |
| `internal/tomlpos` | The TOML adapter: the values and the key positions. |

- The host language is Go. The Go module is `newbbs`.
- Imports go one way: `cmd/bbs`, then `internal/bbt`, then `internal/art`,
  then `internal/screen`. `internal/config` imports only `internal/screen`
  and `internal/tomlpos`. `internal/bbt` also imports `internal/tomlpos`.
- The model is immutable after the load. Sessions read it at the same time without locks.
- The checks are `go vet ./...`, `go test ./...` and `golangci-lint run ./...`
  (golangci-lint 2.15.0, configuration in `.golangci.yml`).

## 12. Libraries

| Task | Choice | Compared with | Reason |
|---|---|---|---|
| TOML | `github.com/pelletier/go-toml/v2` | `github.com/BurntSushi/toml` | `BurntSushi/toml` gives positions only for syntax errors. The `unstable` parser of `go-toml/v2` gives the position of each key. |
| Graphemes and width | `github.com/clipperhouse/displaywidth` | `github.com/rivo/uniseg`, `github.com/mattn/go-runewidth` | One iterator gives each grapheme with its width, the cell model. The release has Unicode 17; `uniseg` has Unicode 15. `go-runewidth` measures runes. |
| ANSI art and SAUCE | Own parser | `github.com/charmbracelet/x/ansi` | The loader needs about 12 sequences and the editor rules of section 8. A terminal parser adds a large dependency. |
| Conditions | Own parser | `github.com/expr-lang/expr` | The grammar has 7 rules and static types from the registry. `expr` adds arithmetic and functions that the format excludes. |
| Command line | `flag` | `github.com/spf13/cobra` | The standard library is sufficient for two commands. |

- The Go standard library and `golang.org/x/text` have no grapheme segmentation.
- The `unstable` package of `go-toml/v2` has no compatibility guarantee. This deviates
  from the Libraries rule (`docs/dev-flow/principles.md`, section 2). The operator accepted it.
  `go.mod` pins the version. Only `internal/tomlpos` uses the package, and tests cover it.

## 13. Not in the MVP

- `#each`, list rows in art, data providers and `{#if}` inside art:
  [lists](../ideas/lists.md). In the MVP, logic is in the BBS template only.
- CP437 art: [art-convert](../ideas/art-convert.md).
- Raw art and ANSI animation: [content-trust](../ideas/content-trust.md).
- Reload: [reload](../ideas/reload.md). A restart is sufficient for the MVP.
- Marks and arrow-key selection: [menu-actions](../ideas/menu-actions.md).
  The MVP selects items by hotkey only.

Legacy BBS display codes (pipe codes, `@` codes and MCI codes) are not supported.
There is no converter.
