# BBT format

This specification defines the BBT files and the layout of a screen.
Related parts: [art-format](art-format.md), [screens-toml](screens-toml.md),
[bbs-toml](bbs-toml.md), [screen-render](screen-render.md), [menu-nav](menu-nav.md),
[file-view](file-view.md), [list-view](list-view.md), [session-fields](session-fields.md) and [telnet-term](telnet-term.md).

The [BBT guide](bbt-guide.md) teaches this specification. Keep the guide in sync:
[bbt-guide](bbt-guide.md), section 3.

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
| `content/<path>` | A content file of a file screen: [file-view](file-view.md), section 2. |
| `screens/<code>/`, `content/<code>/` | The files of a language: [session-language](session-language.md), section 3. |

- `<name>` matches `[a-z0-9][a-z0-9_-]*`. Example: `main_wide`.
- An `.ans` file is art only. Only an art insert in a BBS template shows it.
- A `.bbt` or `.ans` file that is not a regular file is an error.
- The first screen and the screens `goodbye` and `shutdown` have special roles:
  [menu-nav](menu-nav.md), sections 4 and 5.

### 2.1. Characters

- A "grapheme" is an extended grapheme cluster (Unicode UAX #29).
  A grapheme of more than 32 bytes is an error in a file.
- The width of a grapheme is 0, 1 or 2 cells, by Unicode 17. East Asian Ambiguous characters
  have the width 1. A terminal that shows them as wide gives `term.ambiguous` 2
  ([telnet-term](telnet-term.md), section 5.4). The layout does not change.
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
| `align` | string | The horizontal position: `"left"`, `"center"` or `"right"`. Section 6.1. |
| `valign` | string | The vertical position: `"top"`, `"middle"` or `"bottom"`. Section 6.5. |
| `fill` | string | The fill rows: `"none"`, `"spaces"` or `"erase"`. Section 6.5. |
| `gutter` | integer | The cells that the render width leaves on each side of the screen area, from 0 to 100. Section 6.1. |
| `file`, `lines`, `numbers`, `margin` | — | The file of a file screen: [file-view](file-view.md), section 3. |
| `data`, `rows`, `select` | — | The list of a list screen: [list-view](list-view.md), section 3. |
| `[[item]]` | array of tables | Menu items with `key`, `do`, an optional `when` and optional `args`. Section 8. An array of inline tables is the same: `item = [{ key = "G", do = "logoff" }]`. |

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

`{` starts a token when an ASCII letter, a space or ASCII punctuation other than `{` and `}` follows it.
Each other `{` is a plain character. In the body, `{{` gives a plain `{`.
A token ends at the first `}` on the same line. A token without `}` is an error.
In art, only [art-format](art-format.md), sections 4 and 5, defines the tokens.

| Token | Meaning |
|---|---|
| `{<path>}` | Field with its natural width. Section 4. |
| `{<path>      }` | Field with the width of the token, aligned left. |
| `{      <path>}` | Field with the width of the token, aligned right. |
| `{#if <cond>}` `{#else}` `{/if}` | Conditional block. Section 5. |
| `{!<style>}` | Style change. Section 7.3. |
| `{>_<name>}` | Include of the fragment `_<name>.bbt`. Section 3.4. |
| `{~}` `{~<c>}` | Stretch token: free cells of the row. Section 6.6. A source line `{~}` is a spacer line. |
| `{><name>.ans}` | Art insert of the art `<name>.ans`. Sections 3.4 and 6.2. |
| `{>file}` | The file area of a file screen: [file-view](file-view.md), section 3.1. |
| `{#each}` `{/each}` | A list line of a list screen: [list-view](list-view.md), section 5. |
| `{@<id>}` `{/@}` | Reserved for marks. An error in the MVP. |
| `{<punctuation>...}` | Each other token that starts with ASCII punctuation is reserved for new tokens. An error. |

- The width of a fixed-width field includes the braces. Example: `{term.type      }` is 17 cells.
- A field token with spaces on both sides of the path is an error.
- "Block tokens" are `{#if}`, `{#else}` and `{/if}`. Blocks can nest and can span lines.
  `{#else}` is optional. The blocks of each file must balance.

### 3.4. Includes and art inserts

- An include or an art insert must be the only content of its source line.
- An include inserts the body lines of the fragment in place of its line.
- The host finds the file of an include and of an art insert in the session language
  ([session-language](session-language.md), section 4).
- An include of a name without `_`, except `{>file}`, is an error. A missing file is an error
  in each screen set that does not find it ([session-language](session-language.md), section 6).
- Includes can nest. An include cycle is an error.
- After an include, the style is the style before the include.

## 4. Fields

### 4.1. Registry

| Path | Type | Value |
|---|---|---|
| `term.type` | string | Terminal type. |
| `term.cols` | integer | Terminal columns. |
| `term.rows` | integer | Terminal rows. |
| `term.colors` | integer | Colour depth: `16`, `256` or `16777216`. |
| `term.name` | string | Terminal name and version, for display only. Can be empty. |
| `term.ambiguous` | integer | Cells of an East Asian Ambiguous character in the terminal: `1` or `2`. |
| `term.background` | string | Background of the terminal: `dark`, `light` or empty. |
| `term.bg_color` | string | Background colour of the terminal: `#rrggbb` or empty. |
| `item.<KEY>` | boolean | True when the item with the key `<KEY>` exists and is enabled (section 8.2). |
| `arg.<name>` | string | The argument `<name>` of the screen (section 8.5). |
| `file.*`, `meta.<key>` | — | The fields of a file screen: [file-view](file-view.md), section 3.4. |
| `row.*`, `list.*` | — | The fields of a list screen: [list-view](list-view.md), section 6. |
| `session.*`, `bbs.*` | — | The fields of the session and of the BBS: [session-fields](session-fields.md), section 2. |

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

- The "render width" W is `term.cols - 2 * gutter`, clamped to the "width limits" `[min, max]`.
- The width limits are `width` in the front matter, else `[screen] width`
  ([screens-toml](screens-toml.md), section 3.1).
  The same rule gives `align`, `valign`, `fill` and `gutter`.
- When `term.cols` > W, the "screen area" has a left margin. F is `term.cols - W`.
  The margin is `F - min(gutter, F)` cells with `align = "right"`, `F / 2` cells, rounded down,
  with `"center"`, and `min(gutter, F)` cells with `"left"`.
- Example: `width = [68, 108]`, `align = "center"` and `gutter = 6` give W = 68 and a margin
  of 6 at 80 columns, and W = 108 and a margin of 12 at 132 columns.
  The text of a source line that starts with 2 spaces is then at the columns 9 and 15,
  counted from 1, and a mark 2 cells left of the text is in the screen area.
- When `term.cols` < `min`, the host renders at W = `min` and cuts each row at `term.cols`.
- A screen with art of width `a` for smaller terminals has `width = [40, a]`, `gutter = 0` and this body:

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
  A list line is an exception ([list-view](list-view.md), section 5).
- An art insert gives the rows of the art. Each row starts at the left edge of the screen area.
  The host cuts the art at W. `velourbbs check` warns when the art is wider than `max`.
- A cut of a row at W or at `term.cols` through a wide grapheme leaves a space in its style.
- A spacer line gives the rows of section 6.6.
- A source line with a stretch token that the render reaches gives one row (section 6.6),
  also with a block token.
- Each other source line gives one or more rows (section 6.3). An empty line gives an empty row.
- Each cell of a row of a page that the screen does not draw is a space in the theme style `normal`.
  Section 6.5 gives the cells of the other terminal rows.

### 6.3. Text wrap

- A row has W cells. A source line with a stretch token does not wrap (section 6.6).
- A text line breaks at a space. The spaces at a break are removed.
- The spaces at the start of the source line stay up to W. They join the first word.
- The spaces at the end of the source line stay up to W.
- A word that is wider than W starts on a new row and breaks at a grapheme boundary.
- A wide grapheme that does not fit at the end of a row goes to the next row.
  When W is 1, a wide grapheme becomes a space.
- A fixed-width field is one word. If it is wider than W, it has the width W.
- The style continues on the next row.

### 6.4. Height

- A screen has no height limit.
- A file screen has its own pages: [file-view](file-view.md), section 3.2.
- A list screen has its own pages: [list-view](list-view.md), section 7.
- When another screen has more rows than `term.rows`, the host shows it in pages of `term.rows - 1` rows.
  After each page except the last, the host shows the "More row" ([menu-nav](menu-nav.md), section 5).
- The More row is the text of [session-language](session-language.md), section 7.2,
  in the theme style `more`, at the left edge of the screen area.
  The host cuts it at the screen area.

### 6.5. Vertical position and fill

- A file screen has its own rule: [file-view](file-view.md), section 3.2.
- A screen of R rows with R < `term.rows` has one page. Its first row is the terminal row
  1 with `valign = "top"`, `(term.rows - R) / 2`, rounded down, plus 1 with `"middle"`,
  and `term.rows - R + 1` with `"bottom"`.
- Each other page starts at the terminal row 1.
- The terminal rows above and below the rows of a page are the "fill rows".
  `fill` gives their cells:
  - `"none"`: the default colours of the terminal;
  - `"spaces"`: a space in the theme style `normal`;
  - `"erase"`: the background of `normal`. The host clears the terminal screen in it
    ([telnet-term](telnet-term.md), section 7.1). A terminal without "background colour erase"
    (bce) shows its default colours.

### 6.6. Stretch

- A "stretch token" is `{~}` or `{~<c>}`. `<c>` is one grapheme of width 1, except `}`.
  `{~}` is the same as `{~ }`. Each other token that starts with `{~` is an error.
  Example: `{~.}` gives a dot leader, `{~─}` a rule.
- A stretch token in a false block has no effect.
- The "free cells" of a row are W minus the cells of the row without its stretch tokens.
  The spaces at the start and at the end of the source line count.
- The stretch tokens of a row share the free cells. With n tokens and F free cells,
  each token gets `F / n` cells, rounded down. The last `F mod n` tokens get one more cell.
  So `{~}{file.page}{~}` puts the field in the centre, rounded down to the left.
- A token fills its cells with `<c>` in the style at the token.
- When the cells of a row without its stretch tokens are W or more, each stretch token has 0 cells.
  The host cuts the row at W.
- A row of a list line has these rules too ([list-view](list-view.md), section 5).
- Art has no stretch token. In art, `{~` is a plain character ([art-format](art-format.md), section 4).

Example: one source line gives one row at each W.

```
{!ctl}Texts {!frame}{~.} {!dim}Guides and features
{!hi}Q {!ctl}Quit{~}{!dim}{bbs.time}
```

A "spacer line" is a source line that is exactly `{~}`.

- The "free rows" are `term.rows` minus the rows of the screen without the spacer lines.
  When the free rows are 0 or less, each spacer line gives 0 rows.
- The spacer lines that the render reaches share the free rows, by the rule of the free cells.
  Each row of a spacer line is an empty row.
- So a screen with a spacer line and fewer rows than `term.rows` fills its page.
  `valign` has no effect on it (section 6.5).
- A list screen counts the free rows after the page size P ([list-view](list-view.md), section 7).
- In a file screen, a spacer line gives 0 rows: the file area takes the free rows
  ([file-view](file-view.md), section 3.2).

Example: the key row stays at the bottom of the terminal at 24 and at 43 rows.

```
{!hi}V E L O U R
{~}
{!hi}Q {!ctl}Quit
```

## 7. Colours and styles

### 7.1. Colour model

- A cell holds one grapheme and a style, never encoded bytes. A wide grapheme takes 2 cells.
- A "colour" is `default` (the colour of the terminal), an index from 0 to 255, RGB,
  or a VGA colour.
  Indexes 0–7 are the standard colours, 8–15 the bright colours,
  16–255 the xterm 256-colour palette.
- A "VGA colour" is a colour 0–15 of art ([art-format](art-format.md), section 3).
  [telnet-term](telnet-term.md), section 7.3, gives its output with the VGA values.
  So art keeps the colours of its artist.
  The theme and the styles of this format give indexes, never VGA colours.
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
  styles for one `term.colors`. A light table replaces styles in the light theme.
  [screens-toml](screens-toml.md), section 3.2, gives the session theme.
- A theme style is a style of the session theme.
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
do = "gosub sysinfo"
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
do = "back"
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
| `args` | table | Optional. The arguments of a `goto` or a `gosub` (section 8.5). |
| `ask` | string | Optional. The prompt of a line that the host asks for before the action (section 8.6). |
| `ask_max` | integer | Optional. The most characters of the line of `ask` (section 8.6). |

- An item is "enabled" when it has no `when` or its `when` is true.
  A disabled item counts as absent.
- The `when` of the items and the `{#if}` of the body take their values at the same render
  (section 5). So the shown text and the enabled keys always agree.
- A template hides the text of a disabled item with `{#if item.<KEY>}` (section 4.1).

### 8.3. Actions

| `do` | Action |
|---|---|
| `goto <name>` | Show the screen `<name>` in place of the current screen. |
| `gosub <name>` | Show the screen `<name>` over the current screen: [menu-nav](menu-nav.md), section 4.1. |
| `goto <code>/<name>`, `gosub <code>/<name>` | Show the screen `<name>` in the language `<code>`: [menu-nav](menu-nav.md), section 4. |
| `back` | Show the screen below the current screen again: [menu-nav](menu-nav.md), section 4.1. |
| `top` | Show the first screen of the navigation stack again: [menu-nav](menu-nav.md), section 4.1. |
| `logoff` | End the session. |
| `page next`, `page prev`, `page first`, `page last` | Show another page of a file screen or a list screen: [file-view](file-view.md), section 3.3. |
| `theme dark`, `theme light`, `theme auto` | Set the theme choice of the session: [screens-toml](screens-toml.md), section 3.2. |
| `lang <code>` | Set the session language: [menu-nav](menu-nav.md), section 4. |

- The action words are lower case. `do` has no other spaces than the one space after `goto`, `gosub`, `page`, `theme` or `lang`.
- `<name>` matches the screen name of section 2. `<code>` is a language code
  of [session-language](session-language.md), section 2.
- [menu-nav](menu-nav.md), section 4, defines how the host runs the actions.

### 8.4. Errors

Each of these is an error in a screen:

- an item without `key` or without `do`;
- an invalid key;
- a duplicate key in a menu, also `ANY`;
- an invalid `do`: an unknown action, or a wrong argument of the action;
- an `item.<KEY>` path without an item in the screen;
- `args` with an action other than `goto` or `gosub`, an invalid argument name or an invalid value (section 8.5);
- an error of `ask` or `ask_max` (section 8.6).

[menu-nav](menu-nav.md), section 6, has the other menu errors: the `goto` and `gosub` targets, the screens
without items, the items of `goodbye` and `shutdown`, and the start screen. It also has the form of the item errors.

### 8.5. Arguments

- `args` is a table of strings: an inline table or the table `[item.args]`.
  Example: `args = { file = "gold.md" }`.
- A name matches `[a-z][a-z0-9_]*`.
- A value is a literal or exactly one "value token":
  - `{arg.<name>}`: the value of the argument `<name>` of the current screen;
  - `{in}`: the line of `ask` (section 8.6). It is valid only in an item with `ask`;
  - `{row.<field>}`: a field of the selected row. It is valid only in `select`
    of a list screen ([list-view](list-view.md), section 8).
- A literal cannot have a control character or `{`. Each other token is an error,
  also an alias and a field with a width.
- The host puts the value into the argument as one string. It does not parse the value again.
- The "arguments of a screen" are the names of the `arg.<name>` fields of the screen:
  in its body, its fragments, its art, its `when`, its `file`
  ([file-view](file-view.md), section 3), its `data` ([list-view](list-view.md), section 4)
  and the `args` values of its items and of its `select`.
- A `goto` or a `gosub` gives exactly the arguments of its target screen. A missing argument
  or another argument is an error.
- The field `arg.<name>` gives the value of the `goto` or the `gosub` that shows the screen.
  `back` and `top` show a screen with the same values.

Values at load:

- The "values" of an argument of a screen are the values that each `goto` and `gosub`
  to the screen can give it, in all screen sets:
  - a literal gives itself;
  - `{arg.<name>}` gives each value of the argument `<name>` of its own screen.
    A cycle of screens gives no new values;
  - `{row.<field>}` gives the field of each row of each list of its screen;
  - `{in}` gives each line that the user can type: an "open" value.
- The host loads the content files of `file` and the rows of `data` for each value at load
  ([file-view](file-view.md), section 3.1, [list-view](list-view.md), section 4).
  So an argument that a screen uses in `file` or in `data` cannot have an open value,
  and an argument that a screen uses in `data` cannot have a value of `{row.<field>}`.
  So the rows of a provider never depend on the rows of another provider.
- Such a value is an error of the `{in}` or the `{row.<field>}` that gives it,
  also through the `{arg.<name>}` of other screens.
- An error of a content file or of a provider row of a value is an error of the literal
  that gives the value, also through the `{arg.<name>}` of other screens.

Example: `texts` gives `dir` to `parts`, and `parts` gives it on to `read`.
So `dir` of `read` has the value `guide/bbt`, and `file` of `read` has the value
of `row.path` of each row of the list of `parts`:

```
# screens/texts.bbt
[[item]]
key = "2"
do = "gosub parts"
args = { dir = "guide/bbt" }

# screens/parts.bbt
data = { from = "files", args = { dir = "{arg.dir}" } }
rows = 12
select = { do = "gosub read", args = { dir = "{arg.dir}", file = "{row.path}" } }
```

Examples of errors:

```
screens/area.bbt:14:34: item "J": args: n: the value reaches the file of screen "msg"
screens/dirs.bbt:3:50: select: args: dir: the value reaches the data of screen "parts"
screens/texts.bbt:9:16: item "2": args: dir: directory "guide/bbtt" does not exist
```

### 8.6. Ask

`ask` asks the user for one line before the action of the item.

Example: `J` asks for a message number and shows the message.

```
[[item]]
key = "J"
ask = "Message number: "
ask_max = 5
do = "gosub msg"
args = { area = "{arg.area}", n = "{in}" }
```

- `ask` is the "prompt": text without tokens. `{` in it is a plain character.
  It has the rules of section 2.1 and no control character.
- `ask_max` is at least 1. It is valid only with `ask`.
- An item with `ask` must have the action `goto` or `gosub` and a value `{in}` in `args`.
  Else the line has no use.
- The "free cells" of the prompt are the cells of the screen area after the prompt, minus 1.
  At W = `min` of the width limits (section 6.1), the free cells must be at least 1.
- [menu-nav](menu-nav.md), section 5.4, shows the prompt and reads the line.

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
- [art-format](art-format.md), section 6, has the limits of art.

## 10. Not in the MVP

- Marks and arrow-key selection of items: [marks](https://github.com/schors/velourbbs/issues/67).
  Items have hotkeys only. A list screen selects a row with the arrow keys ([list-view](list-view.md), section 8).
- `require` (access checks). It needs users.
- Punctuation keys, for example `?`.
- A `label` for an item. The host does not print item text.

Legacy BBS display codes (pipe codes, `@` codes and MCI codes) are not supported.
There is no converter.
