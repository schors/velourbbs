# File view

This specification defines the file screen: a screen that shows a Markdown file,
a plain text file, or a source file or a part of it, in pages.
Related parts: [bbt-format](bbt-format.md), [screens-toml](screens-toml.md),
[screen-render](screen-render.md) and [menu-nav](menu-nav.md).

## 1. Goal

The sysop shows a file on a screen and does not copy its text into the screen.
One screen shows many files: a menu item gives the file as an argument.
The user pages forward and back. A Markdown file has theme styles and front matter fields.

## 2. Files

| Path | Content |
|---|---|
| `content/<path>` | A "content file": a file that a file screen shows. |

- The host finds `<path>` in the session language: [session-language](session-language.md), section 4.
- The content files are files of the sysop. The host trusts them ([bbt-format](bbt-format.md), section 1).
- `<path>` is one or more segments separated with `/`. A segment matches `[A-Za-z0-9_][A-Za-z0-9_.-]*`.
  Example: `xfer/src/crc16.c`.
- A content file that is not a regular file is an error.
- The host follows a symlink in `<path>`: a symlink to a file or to a directory.
  Its target can be outside the BBS directory.
- The extension of the file gives its "kind". Case does not matter in the extension.

| Extension | Kind |
|---|---|
| `.md`, `.markdown` | Markdown (section 5) |
| `.txt`, or no extension | Plain text (section 6) |
| Each other extension | Source (section 6) |

### 2.1. Content

- The file is UTF-8. The loader removes a BOM. Invalid UTF-8 is an error.
- A line ends with LF or CR LF. A LF at the end of the file ends the last line.
  It does not start a new line. The "line count" is the number of lines.
- The rules of [bbt-format](bbt-format.md), section 2.1, apply to the graphemes.
  A control character except TAB is an error.
- In an error of a content file, the column counts code points.
- The host loads each content file that a file screen can show (section 3.1) at start.
  `velourbbs check` reports its errors ([screen-render](screen-render.md), section 3.1).

## 3. File screen

A "file screen" is a screen with the key `file` in its front matter.

| Key | Type | Meaning |
|---|---|---|
| `file` | string | The content file: a `<path>` of section 2, or exactly one token `{arg.<name>}`. |
| `lines` | array of 2 integers | Optional. `[first, last]`: the part of the file, with 1 <= `first` <= `last`. |
| `numbers` | boolean | Optional. True shows line numbers. The default is false. |
| `margin` | array of 2 integers | Optional. `[left, right]`: the empty cells left and right of the file area (section 3.1). The default is `[0, 0]`. |

- `lines`, `numbers` and `margin` are valid only in a file screen.
- `lines` and `numbers` are valid only for a plain text or source file.
- `last` must be at most the line count of the file.
- `left` and `right` are at least 0. `left + right` must be less than `min` of the width limits
  ([bbt-format](bbt-format.md), section 6.1).

Example `screens/reader.bbt`:

```
+++
file = "{arg.file}"
[[item]]
key = "B"
do = "page prev"
when = "file.page > 1"
[[item]]
key = "SPACE"
do = "page next"
when = "not file.last"
[[item]]
key = "ESC"
do = "back"
+++
{!md_mark}■  {!}{file.section}
{!md_h1}{file.title}
{!md_dim}{meta.author}  ∙  {file.minutes} min
{>file}
{!md_frame}──────────────   {file.page}/{file.pages}   ──────────────
{#if item.B}{!md_mark}B {!}Back    {/if}{#if item.SPACE}{!md_mark}Space {!}Next    {/if}{!md_mark}Esc {!}List
```

An item of the screen `texts` opens a file in it:

```
[[item]]
key = "1"
do = "gosub reader"
args = { file = "gold.md" }
```

### 3.1. File area

- `{>file}` is the "file area". It must be the only content of its source line.
- A file screen has exactly one `{>file}`, in its own body. A `{>file}` in a block,
  in a fragment or in a screen without `file` is an error.
- The file area starts `left` cells after the left edge of the screen area.
  The "area width" A is W − `left` − `right`. Sections 5 and 6 lay out the file at A:
  in them, W is A.
- The margin cells of a file row are spaces in `normal`.
- The host can show each file that `file` names: the path of `file`, or each value
  of the argument `<name>` that a `goto` or a `gosub` to the screen gives ([bbt-format](bbt-format.md), section 8.5).

### 3.2. Pages

- The rows of the body before `{>file}` are the "rows above". The rows after it are the "rows below".
- The host renders the rows above and below with the fields of page 1, with `file.pages` equal to 1.
  The "area height" H is `term.rows` minus their rows, and at least 1.
- The host lays out the file into rows of the area width A (sections 5 and 6).
  It splits the rows into "file pages" of at most H rows (section 5.6).
  A file page with fewer than H rows has empty rows after its rows, up to H.
- Each file page gives one page of the screen: the rows above, the H rows of the file page
  and the rows below.
- The host renders the rows above and below again for each page, with the fields of that page.
- A page of the screen with more than `term.rows` rows is cut at `term.rows`.
  A page with fewer rows starts at the terminal row 1 and has fill rows below it
  ([bbt-format](bbt-format.md), section 6.5).
- A file screen has no More row.
- An empty file has one file page.
- The layout stops after 10000 file rows. The rows after the limit are not shown.

### 3.3. Page actions

| `do` | Action |
|---|---|
| `page next` | Show the next page. On the last page, show the last page again. |
| `page prev` | Show the previous page. On the first page, show the first page again. |
| `page first` | Show the first page. |
| `page last` | Show the last page. |

- A `page` action is valid only in a file screen and in a list screen ([list-view](list-view.md), section 7).
- [menu-nav](menu-nav.md), section 5.1, defines the show of a page of a file screen,
  and section 5.3 of a list screen.

### 3.4. Fields

| Path | Type | Value |
|---|---|---|
| `file.title` | string | The front matter key `title`, else the text of the first level-1 heading, else the file name without the directories. |
| `file.page` | integer | The number of the current page, from 1. |
| `file.pages` | integer | The page count. |
| `file.last` | boolean | True on the last page. |
| `file.section` | string | The text of the last heading of level 1 or 2 that gives rows, at or before the first file row of the page. Else empty. |
| `file.minutes` | integer | The reading time in minutes (section 3.5). |
| `meta.<key>` | string | The front matter key `<key>` (section 5.1). Else empty. |

- The fields of this section are valid only in a file screen.
- `file.last` is valid only in a condition: a field token cannot show a boolean.
- `<key>` matches `[a-z][a-z0-9_]*`.
- The values follow [bbt-format](bbt-format.md), section 4.3.

### 3.5. Reading time

- A "word" is a run of graphemes between spaces.
- The host counts the words of the text of the file before the layout:
  - Markdown: the text of the blocks and the inline elements, without the front matter
    and the addresses of the links and images;
  - plain text and source: the lines that the screen shows (`lines`).
- The reading time is the word count divided by 200, rounded up. It is at least 1.

## 4. Theme styles

The base theme has these styles ([screens-toml](screens-toml.md), section 3.2):

| Style | Default | Use |
|---|---|---|
| `md_h1` | `15 bold` | Heading of level 1. |
| `md_h2` | `15 bold` | Heading of level 2. |
| `md_h3` | `15` | Heading of level 3 to 6. |
| `md_strong` | `15 bold` | Strong emphasis. The header row of a table. |
| `md_em` | `14` | Emphasis. The text of a quote. |
| `md_code` | `10` | Code span, code block, source file. |
| `md_link` | `14 underline` | Link text, image text. |
| `md_mark` | `11` | List marker of level 1, quote bar, break. |
| `md_frame` | `8` | List marker of level 2 and more, code bar, the rule of a table. |
| `md_dim` | `8` | Line number, note number, note, the cut mark. |

- Plain text has the style `normal`.
- Inside nested elements, the style of the innermost element applies.

## 5. Markdown

- The Markdown is CommonMark 0.31.2 with the tables of GitHub Flavored Markdown.
  The target implementations are the CommonMark reference implementation and GitHub.
- Other extensions of GitHub show as text. Examples: `~~text~~`, `[ ]` of a task list,
  a web address without `<` and `>`.
- A character reference can give any character. The host treats the text and the addresses
  after the parse as values: [bbt-format](bbt-format.md), section 4.3.

### 5.1. Front matter

- If the first line is `+++`, the lines up to the next `+++` line are TOML front matter.
  A first line `+++` without its end line is an error.
- If the first line is `---`, the second line is not empty, and a `---` or `...` line follows,
  the lines up to it are YAML front matter. YAML front matter is an error: Velour reads only TOML.
- Else the file has no front matter.
- A delimiter line can have spaces at its end.
- A syntax error is an error. A top-level key two times is an error.
- The front matter has no limit on its keys. A key that matches `[a-z][a-z0-9_]*` and has
  a scalar value gives the field `meta.<key>`. The value of the field:

| Value | Field |
|---|---|
| String | The string. |
| Integer, float | The number as the file writes it. Example: `0x1F`. |
| Boolean | `true` or `false`. |
| Date | `YYYY-MM-DD`. |
| Date and time with an offset | RFC 3339. |
| Date and time without an offset | RFC 3339 without the offset. |
| Time | `HH:MM:SS`, with the fraction when the file has it. |

- The other keys and values have no effect.
- Target implementation: Hugo (TOML). The rule for `---` follows Jekyll and Pandoc,
  so their YAML front matter gives an error, not a page with the front matter as text.

### 5.2. Blocks

- One empty row separates two blocks. Inside a tight list, the blocks of the list
  have no empty row between them. Inside a quote, the empty row has the bar.
- A block that gives no rows does not take a separator.
- Text wraps as [bbt-format](bbt-format.md), section 6.3. A soft line break is a space.
  A hard line break starts a new row. A TAB in text is a space.
- The "prefix" of a row is the indent, the list markers and the bars before its text.
  It is at most W / 2 cells, rounded down. A longer prefix keeps its last W / 2 cells.
- A paragraph or a heading with no cells gives no rows. An empty list item gives one row with its marker.

| Block | Rows |
|---|---|
| Paragraph | The text. |
| Heading | The text in `md_h1`, `md_h2` or `md_h3`. |
| Bullet list | The marker at the indent 2: at level 1 `∙` in `md_mark`, at level n > 1 `-` in `md_frame`. One space follows the marker, then the text. |
| Ordered list | As a bullet list, with the number and `.` in place of the marker. The numbers of a list align right to the widest number. |
| Block quote | `│` in `md_mark` and a space before each row. The content in `md_em`. |
| Code block | `│` in `md_frame` and a space before each row. The code in `md_code`. Section 5.4. |
| Thematic break | `∙   ∙   ∙` in `md_mark`, cut to the space. |
| Table | Section 5.3. |
| HTML block | The text as written, in `normal`. Each line wraps as a text line. An HTML comment gives no rows. The text after its `-->` shows when it is not only spaces. |
| Link reference definition | No rows. |

- A block inside a list item starts at the text of the item. A block inside a quote starts after the bar.
- The "space" of a block is W minus its prefix.

### 5.3. Table

- The width of a column is the width of its widest cell.
- Two spaces separate the columns. A cell aligns as its column in the separator row.
- The header row is in `md_strong`. A row of `─` in `md_frame` under it has the width of the table.
- A row wider than its space is cut: the cut mark (section 5.4) ends it.

### 5.4. Code rows

- A code row does not wrap. A TAB goes to the next multiple of 8 cells from the start of the line.
- A row wider than its space is cut. Its last 3 cells are the "cut mark" `...` in `md_dim`.
  When the space is less than 3 cells, the cut mark is cut to the space.

### 5.5. Inline elements

| Element | Cells |
|---|---|
| Strong emphasis | `md_strong` |
| Emphasis | `md_em` |
| Code span | `md_code` |
| Link | The text in `md_link`. A note number `[<n>]` in `md_dim` follows the text of a link with a note. |
| Image | The description as the text of a link. |
| Autolink | The address in `md_link`. It has no note. |
| Inline HTML | The text as written. An HTML comment gives no cells. |

- A link or an image with an address that has a scheme, for example `https:`, has a note.
  A note number is the number of the address in the order of the note numbers in the text.
- The "notes" are the rows after the last block: an empty row, then for each address
  `[<n>] <address>` in `md_dim`. The address wraps as text after `[<n>] `,
  or from the start of the row when `[<n>] ` is wider than W / 2.
- A link without a scheme, for example to another Markdown file, has no note.

### 5.6. Page breaks

- A heading of level 1 starts a new file page, except at the first file row.
- A heading and the first row of the next block are on one file page. When they do not fit
  in the rest of the page and fit in H rows, the heading moves to the next file page.
- A file page does not start with an empty row.

## 6. Plain text and source

- A plain text file: each line wraps as [bbt-format](bbt-format.md), section 6.3, in `normal`.
- A source file: each line is a code row (section 5.4) in `md_code`.
- A TAB goes to the next multiple of 8 cells from the start of the line, before the wrap
  and without the line number.
- `lines` shows only the lines from `first` to `last`.
- `numbers` shows the line number of the file at the start of the first row of each line:
  right-aligned to the width of the number `last`, else of the line count, then one space.
  The other rows of a line have spaces there. This number column is in `md_dim`.
  A number column wider than W / 2 is not shown.
- An empty line gives an empty row. Empty rows at the start of a file page stay.

## 7. Errors

Each of these is an error of `velourbbs check` ([screen-render](screen-render.md), section 3.1):

- a content file that does not exist, that a rule of section 2 forbids, or that is over its limit (section 8);
- an error of the content (section 2.1) or of the front matter (section 5.1),
  at its position in the content file;
- an error of the file screen keys and of `{>file}` (sections 3 and 3.1);
- a `page` action outside a file screen and a list screen, a field of section 3.4 outside a file screen.

Positions:

- An error of a key of the screen has the position of the first line of its front matter
  ([screen-render](screen-render.md), section 3.1).
- An error of a whole content file (it does not exist, it is not a regular file, it is over 1 MB)
  is an error of the key `file`.
- An error of a file that a `goto` or a `gosub` gives is an error of its item:
  the position of the first line of the front matter of its screen,
  and the start `item "<KEY>": args: <name>: ` ([menu-nav](menu-nav.md), section 6).
  This includes an error of `lines` and `numbers` for this file.

Examples:

```
screens/reader.bbt:2:1: file: content file "gold.md" does not exist
screens/texts.bbt:2:1: item "1": args: file: content file "gild.md" does not exist
screens/code.bbt:2:1: item "2": args: file: lines: "gold.md" is a Markdown file
content/gold.md:1:1: front matter: no end line "+++"
```

## 8. Limits

| Content | Limit |
|---|---|
| Content file | 1 MB. |
| Front matter | 64 KB. |
| Grapheme | 32 bytes. |
| File rows | 10000 (section 3.2). |

## 9. Code

| Package | Content |
|---|---|
| `internal/content` | The load of the content files, the front matter and the Markdown parse. |
| `internal/screen` | The layout of a file into rows and pages. |

- The model holds each content file after its parse. It is immutable after the load.
- At the show of a file screen, the host lays out the file once and keeps the first row
  of each file page. It holds one page in memory ([screen-render](screen-render.md), section 2).

## 10. Libraries

| Task | Choice | Compared with | Reason |
|---|---|---|---|
| Markdown | `github.com/yuin/goldmark/v2` | `github.com/yuin/goldmark` v1, `github.com/gomarkdown/markdown`, own parser | CommonMark 0.31.2 and the GFM tables. An AST without HTML. No dependencies. v2 is the current line: v1 needs a migration later. v2 is young, and a patch release can break the API (v2.0.2 changed `parser.Nil`). `go.mod` pins the version; update it with the tests. An own CommonMark parser is too large. |
| TOML front matter | `github.com/pelletier/go-toml/v2` | — | The TOML library of the screens ([screen-render](screen-render.md), section 5). Its parser gives the number as the file writes it. |

## 11. Acceptance

- A test BBS in `cmd/velourbbs/testdata/` has a file screen and content files:
  a Markdown file with each element of section 5 and a TOML front matter,
  a plain text file and a source file.
- Golden cell dumps of `velourbbs preview` cover each kind at 80x24 and at 40x10,
  `lines`, `numbers` and `margin`, and `--keys` with each `page` action.
- A test covers each error of section 7.
- A test shows a content file through a symlink to a directory outside the BBS directory.
- The BBT guide in English and Russian does not name `args` and `{arg.*}` as not in the MVP
  ([bbt-guide](bbt-guide.md), section 3).

## 12. Not in this specification

- The table of contents and active links to other Markdown files: [file-view](https://github.com/schors/velourbbs/issues/7).
  They need the selection of a row ([list-view](list-view.md), section 8).
- Syntax colours of source files: [file-view](https://github.com/schors/velourbbs/issues/7).
- Italic. The style model has no italic ([bbt-format](bbt-format.md), section 7.1).
- Files of users.
