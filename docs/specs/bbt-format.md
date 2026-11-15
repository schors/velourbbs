# BBT format

This specification defines the BBT files and the layout of a screen.
Related parts: [art-format](art-format.md), [screens-toml](screens-toml.md),
[bbs-toml](bbs-toml.md), [screen-render](screen-render.md), [menu-nav](menu-nav.md)
and [telnet-term](telnet-term.md).

## 1. Goal

A screen author reads the whole file format in this specification.
Screen files are the files of the sysop. The host trusts them.
The limits of section 9 catch mistakes, for example a file that is too large.

## 2. Files

| Path | Content |
|---|---|
| `screens/<name>.bbt` | A "screen": a "BBS template" that the host shows. Section 3. |
| `screens/_<name>.bbt` | A "fragment": a BBS template that an include inserts. Section 3.4. |
| `screens/<name>.ans` | ANSI art in the flat form: [art-format](art-format.md). |
| `screens/screens.toml` | Settings of the screens: [screens-toml](screens-toml.md). |
| `bbs.toml` | Host settings: [bbs-toml](bbs-toml.md). |

- `<name>` matches `[a-z0-9][a-z0-9_-]*`. Example: `main_wide`.
- An `.ans` file is art only. Only an art insert in a BBS template shows it.
- A `.bbt` or `.ans` file that is not a regular file is an error.
- The first screen and the screen `goodbye` have special roles: [menu-nav](menu-nav.md),
  sections 4 and 5.

### 2.1. Characters

- A "grapheme" is an extended grapheme cluster (Unicode UAX #29).
  A grapheme of more than 32 bytes is an error in a file.
- The width of a grapheme is 0, 1 or 2 cells, by Unicode 17. East Asian Ambiguous characters
  have the width 1. Terminals that show them as wide are not in the MVP
  ([telnet-term](telnet-term.md), section 13).
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
| `[[item]]` | array of tables | Menu items with `key`, `do` and an optional `when`. Section 8. |

Example:

```
+++
width = [64, 100]
[[item]]
key = "A"
do = "goto about"
[[item]]
key = "G"
do = "logoff"
+++
{#if term.cols >= 100}
{>main_wide.ans}
{#else}
{>main.ans}
{/if}
{!title}Welcome! Your terminal has {term.cols} columns.
{!prompt}[A]bout [G]oodbye
```

### 3.3. Tokens

`{` starts a token when an ASCII letter, a space, `#`, `/`, `!`, `>` or `@` follows it.
Each other `{` is a plain character. In the body, `{{` gives a plain `{`.
A token ends at the first `}` on the same line. A token without `}` is an error.
In art, only [art-format](art-format.md), section 4, defines the field tokens.

| Token | Meaning |
|---|---|
| `{<path>}` | Field with its natural width. Section 4. |
| `{<path>      }` | Field with the width of the token, aligned left. |
| `{      <path>}` | Field with the width of the token, aligned right. |
| `{#if <cond>}` `{#else}` `{/if}` | Conditional block. Section 5. |
| `{!<style>}` | Style change. Section 7.3. |
| `{>_<name>}` | Include of `screens/_<name>.bbt`. Section 3.4. |
| `{><name>.ans}` | Art insert of `screens/<name>.ans`. Sections 3.4 and 6.2. |
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
| `item.<KEY>` | boolean | True when the item with the key `<KEY>` exists and is enabled (section 8.2). |

- A path matches `[a-z][a-z0-9_]*(\.[A-Za-z0-9_]+)?`. A path without a dot is an alias.
- [telnet-term](telnet-term.md) gives the `term.*` values.
  `term.cols` is from 20 to 255. `term.rows` is from 10 to 100.
- `item.<KEY>` is valid only in `{#if}`. Section 8.1 defines `<KEY>`.
- An unknown path is an error.
- A field token shows an integer in decimal and a string as it is.

### 4.2. Aliases

- `[fields]` in `screens/screens.toml` maps an alias to a registry path:
  [screens-toml](screens-toml.md), section 3.3.
- An alias is valid in each screen, fragment and art. Short aliases let art have short fields.

### 4.3. Values

- The host inserts a value after the parse, so a value cannot make a token.
- The host replaces each invalid UTF-8 byte of a value with U+FFFD.
- The host removes the control characters and the graphemes of width 0 from a value.
  It replaces a grapheme of more than 32 bytes with U+FFFD.
- A fixed-width field pads its value with spaces or cuts the end of the value
  at a grapheme boundary. A wide grapheme that does not fit becomes a space.

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
- An integer literal must fit in a 64-bit signed integer.
- `and` binds tighter than `or`. There are no parentheses.
- A path without an operator and a path after `not` must be boolean.
- An integer field takes an integer literal and each operator.
- A string field takes a string literal and only `==` or `!=`.
- Each other combination is an error.
- The host evaluates the conditions of a screen at each render.

## 6. Layout

### 6.1. Width

- The "render width" W is `term.cols`, clamped to the "width limits" `[min, max]`.
- The width limits are `width` in the front matter, else `[screen] width`
  ([screens-toml](screens-toml.md), section 3.1).
- When `term.cols` > W, the "screen area" has a left margin of `(term.cols - W) / 2` cells,
  rounded down.
- When `term.cols` < `min`, the host renders at W = `min` and cuts each row at `term.cols`.
- A screen with art of width `a` for smaller terminals has `width = [40, a]` and this body:

```
{#if term.cols < a}
{>_main_small}
{#else}
{>main.ans}
{/if}
```

### 6.2. Rows

The body gives rows in source order:

- The content of a false block gives nothing: no text, no line end, no art and no include.
- A source line with a block token gives no row when it has only spaces after the evaluation.
- An art insert gives the rows of the art. Each row starts at the left edge of the screen area.
  The host cuts the art at W. `velourbbs check` warns when the art is wider than `max`.
- A cut of a row at W or at `term.cols` through a wide grapheme leaves a space in its style.
- Each other source line gives one or more rows (section 6.3). An empty line gives an empty row.
- Each cell that the screen does not draw is a space in the theme style `normal`.

### 6.3. Text wrap

- A row has W cells. A text line breaks at a space. The spaces at a break are removed.
- The spaces at the start of the source line stay up to W. They join the first word.
- The spaces at the end of the source line stay up to W.
- A word that is wider than W starts on a new row and breaks at a grapheme boundary.
- A wide grapheme that does not fit at the end of a row goes to the next row.
  When W is 1, a wide grapheme becomes a space.
- A fixed-width field is one word. If it is wider than W, it has the width W.
- The style continues on the next row.

### 6.4. Height

- A screen has no height limit.
- When a screen has more rows than `term.rows`, the host shows it in pages of `term.rows - 1` rows.
  After each page except the last, the host shows the "More row" ([menu-nav](menu-nav.md), section 5).
- The More row is `[screen] more` ([screens-toml](screens-toml.md), section 3.1),
  in the theme style `more`, at the left edge of the screen area.
  The host cuts it at the screen area.

## 7. Colours and styles

### 7.1. Colour model

- A cell holds one grapheme and a style, never encoded bytes. A wide grapheme takes 2 cells.
- A "colour" is `default` (the colour of the terminal), an index from 0 to 255, or RGB.
  Indexes 0–7 are the standard colours, 8–15 the bright colours,
  16–255 the xterm 256-colour palette.
- A "style" is a foreground colour, a background colour and the attributes
  `bold`, `underline` and `blink`.
- [telnet-term](telnet-term.md) reduces each colour to the colour depth of the terminal.

### 7.2. Style syntax

```
style  = colour [ "/" colour ] { " " attr } | "/" colour { " " attr } ;
colour = integer | "#" hex hex hex hex hex hex | "-" ;
attr   = "bold" | "underline" | "blink" ;
```

- An integer colour is an index from 0 to 255. `-` is the colour `default`.
- Examples: `14`, `14/1`, `/1`, `#ff8800/#000020`, `15/4 bold`, `-/-`.

### 7.3. Theme and style tokens

- The theme in `screens/screens.toml` maps a style name to a style. A depth theme replaces
  styles for one `term.colors`: [screens-toml](screens-toml.md), section 3.2.
- A theme style is a style of the session theme of the `term.colors` of the terminal.
- `{!<name>}` sets the whole theme style `<name>`. A name that is not a name of the theme
  is an error.
- `{!14}`, `{!14/1}` and `{!/1}` set only the colours that they name.
  The other colour and the attributes stay. A style token cannot have attributes.
- `{!}` sets the theme style `normal`.
- Each screen starts with `normal`. The style continues across lines and blocks.
  A style token in a false block has no effect.
- An art insert does not change the style of the text after it.

## 8. Items

- A "menu" is the `[[item]]` list in the front matter of a screen (section 3.2).
- An "item" is one choice: `key`, `do` and an optional `when`.
  An item has no position on the screen. The screen author writes the item text,
  for example `[A]bout`, into the body.

Example `screens/main.bbt`:

```
+++
[[item]]
key = "S"
do = "goto sysinfo"
when = "term.cols >= 80"
[[item]]
key = "G"
do = "logoff"
+++
{>main.ans}
{#if item.S}[S]ystem info  {/if}[G]oodbye
```

Example `screens/sysinfo.bbt`:

```
+++
[[item]]
key = "ANY"
do = "goto main"
+++
Terminal: {term.cols}x{term.rows}. Press any key.
```

### 8.1. Keys

A key is one of these:

- a Latin letter `A`–`Z`;
- a digit `0`–`9`;
- a named key: `ESC`, `ENTER`, `SPACE`, `TAB`, `BS`, `DEL`, `UP`, `DOWN`, `LEFT`, `RIGHT`,
  `HOME`, `END`, `PGUP`, `PGDN`, `F1`–`F12`;
- `ANY`.

[menu-nav](menu-nav.md), sections 3.2 and 3.3, defines how a key matches an item and `ANY`.

Case does not matter in `key` and in `item.<KEY>`. `s`, `S` and `esc`, `ESC`
are the same keys.

### 8.2. Item keys

| Key | Type | Meaning |
|---|---|---|
| `key` | string | The key of the item. Section 8.1. |
| `do` | string | The action. Section 8.3. |
| `when` | string | Optional. A condition (section 5). |

- An item is "enabled" when it has no `when` or its `when` is true.
  A disabled item counts as absent.
- The `when` of the items and the `{#if}` of the body take their values at the same render
  (section 5). So the shown text and the enabled keys always agree.
- A template hides the text of a disabled item with `{#if item.<KEY>}` (section 4.1).

### 8.3. Actions

| `do` | Action |
|---|---|
| `goto <name>` | Show the screen `<name>`. |
| `logoff` | End the session. |

- The action words are lower case. `do` has no other spaces than the one space after `goto`.
- `<name>` matches the screen name of section 2.
- [menu-nav](menu-nav.md), section 4, defines how the host runs the actions.

### 8.4. Errors

Each of these is an error in a screen:

- an item without `key` or without `do`;
- an invalid key;
- a duplicate key in a menu, also `ANY`;
- an invalid `do`: an unknown action, or a wrong argument of the action;
- an `item.<KEY>` path without an item in the screen.

[menu-nav](menu-nav.md), section 6, has the other menu errors: the `goto` targets, the screens
without items, the items of `goodbye` and the start screen. It also has the form of the item errors.

## 9. Limits

| Content | Limit |
|---|---|
| `.bbt` file | 1 MB. |
| Front matter | 64 KB. |
| Body of a screen after includes | 10000 lines. An include or an art insert counts its own line, and the lines of the fragment or the rows of the art. |
| Rendered screen | 10000 rows. |
| Grapheme | 32 bytes. |

- A file or a part of a file over a limit is an error. The loader counts the lines while it expands
  the includes, and stops at the limit.
- The render stops at the row limit.
- [art-format](art-format.md), section 5, has the limits of art.

## 10. Not in the MVP

- `#each`, list rows in art, data providers and `{#if}` inside art:
  [lists](../ideas/lists.md). In the MVP, logic is in the BBS template only.
- Marks and arrow-key selection: [menu-actions](../ideas/menu-actions.md).
  The MVP selects items by hotkey only.
- `require` (access checks). It needs users.
- `args`, `ask` and `{arg.*}`.
- Punctuation keys, for example `?`.
- A `label` for an item. The host does not print item text.

Legacy BBS display codes (pipe codes, `@` codes and MCI codes) are not supported.
There is no converter.
