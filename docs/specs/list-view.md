# List view

This specification defines the list screen: a screen that shows the rows of a data provider
in pages, and the selection of a row.
Related parts: [bbt-format](bbt-format.md), [art-format](art-format.md),
[screens-toml](screens-toml.md), [screen-render](screen-render.md), [menu-nav](menu-nav.md),
[file-view](file-view.md), [session-language](session-language.md) and [telnet-term](telnet-term.md).

The [BBT guide](bbt-guide.md) teaches this specification. Keep the guide in sync:
[bbt-guide](bbt-guide.md), section 3.

## 1. Goal

A screen shows a list of records and does not repeat them in items and in text.
The user pages through the list, moves a cursor over the rows and runs an action with a row.
Example: one screen lists the parts of a guide, and Enter opens the selected part.

## 2. Concepts

- A "provider" gives the rows of a list. Each provider is built into the host in Go.
- A "row" is one record: a value for each field of the provider.
- The "list" of a screen is the rows of its provider, in the order of the provider.
- A "list screen" is a screen with the key `data` in its front matter.
- A "list page" is P rows of the list, with the page size P of section 7. Page `n` holds the rows
  from `(n - 1) * P + 1` to `n * P`.

## 3. List screen

| Key | Type | Meaning |
|---|---|---|
| `data` | table | The provider: `from` and an optional `args`. Section 4. |
| `rows` | integer | The most rows of a list page, from 1 to 100. Section 7. |
| `select` | table | Optional. The action of a row: `do` and an optional `args`. Section 8. |

- `rows` is required in a list screen. `rows` and `select` are valid only in a list screen.
- A list screen cannot have the key `file` ([file-view](file-view.md), section 3).
- A list screen has one list. It has no More row ([bbt-format](bbt-format.md), section 6.4):
  the host cuts its rows at `term.rows`.
- A list screen has at least one list line (section 5) or list row in art (section 9),
  in its body, its fragments or its art.
- [menu-nav](menu-nav.md), section 5.3, shows a list screen.

Example `screens/parts.bbt`:

```
+++
data = { from = "files", args = { dir = "{arg.dir}" } }
rows = 12
select = { do = "gosub read", args = { file = "{row.path}" } }
[[item]]
key = "PGDN"
do = "page next"
when = "not list.last"
[[item]]
key = "PGUP"
do = "page prev"
when = "list.page > 1"
[[item]]
key = "ESC"
do = "back"
+++
{!title}Parts{!}  {!dim}page {list.page}/{list.pages}{!}

{#each}{#if row.selected}{!hi}∙ {!pick}{#else}  {!key}{/if}{  row.num}{!}  {row.title                              } {!dim}{row.minutes} min{/each}
{#if list.empty}{!dim}No parts yet.{/if}

{!prompt}Up Down Enter  PgUp PgDn  Esc
```

An item of the screen `texts` opens the list:

```
[[item]]
key = "2"
do = "gosub parts"
args = { dir = "guide/bbt" }
```

## 4. Providers

- `from` is the name of a provider. An unknown name is an error.
- `args` is a table of strings: the arguments of the provider. Each provider defines
  its arguments and their values. A missing required argument and an unknown argument are errors.
- A value is a literal or exactly one token `{arg.<name>}`: an argument of the screen
  ([bbt-format](bbt-format.md), section 8.5).
- The host reads the rows of each provider at its load. They do not change until the next start.
- For a value `{arg.<name>}`, the host reads the rows of each value of the argument
  ([bbt-format](bbt-format.md), section 8.5), in each screen set
  ([session-language](session-language.md), section 6).
  `velourbbs preview` also reads the rows of a `--set` value ([screen-render](screen-render.md), section 3.2).

### 4.1. `files`

`files` gives the content files of one directory ([file-view](file-view.md), section 2).

| Argument | Required | Value |
|---|---|---|
| `dir` | yes | The directory: a `<path>` of [file-view](file-view.md), section 2. |

- The host finds the directory with the search of [session-language](session-language.md), section 4:
  `content/<code>/<dir>` if a directory entry exists there, else `content/<dir>`.
  A missing directory is an error. The host does not merge the two directories.
- A row is each entry of the directory whose name is a segment of [file-view](file-view.md),
  section 2, and that is a regular file or a symlink to a regular file.
  The host does not go into subdirectories. It ignores each other entry.
- Each file of a row is a content file. It has the checks of [file-view](file-view.md), section 7.
- The order: `README.md` first, then the other names in the natural order.
  The natural order compares each run of digits as a number and the other characters by code point.
  Names that are equal in the natural order, for example `2.md` and `02.md`, compare by bytes.
  Example: `README.md`, `1-intro.md`, `2-menu.md`, `10-art.md`.

| Field | Type | Value |
|---|---|---|
| `row.path` | string | The path of the file under `content/` or `content/<code>/`, the form of `file`. Example: `guide/bbt/2-first-screen.md`. |
| `row.title` | string | The title of the file: the rule of `file.title` ([file-view](file-view.md), section 3.4). |
| `row.minutes` | integer | The reading time of the whole file ([file-view](file-view.md), section 3.5). |

## 5. Text list

A "list line" is a source line of a BBS template that starts with `{#each}` and ends with `{/each}`.

- The text between `{#each}` and `{/each}` is the "row text". It can have each token of
  [bbt-format](bbt-format.md), section 3.3, except an include, an art insert, `{>file}`
  and `{#each}`. Its blocks must balance.
- A list line gives exactly P terminal rows: one row for each row of the list page,
  then empty rows up to P. An empty row is spaces in `normal`.
- A row of a list line does not wrap ([bbt-format](bbt-format.md), section 6.3).
  The host cuts it at W.
- The style of each row starts as the style before the list line. The style after the list line
  is the style before it.
- A list line can be in a block and in a fragment.
  A screen can have more than one list line, for example one in each layout
  of `{#if term.cols >= 132}`. Each list line that the render reaches shows the same list page.
- `{#each}` or `{/each}` in another place is an error.

## 6. Fields

| Path | Type | Value |
|---|---|---|
| `row.num` | integer | The position of the row in the list, from 1. |
| `row.selected` | boolean | True on the row of the cursor (section 8). |
| `row.<field>` | — | The field `<field>` of the provider (section 4). |
| `list.page` | integer | The number of the current list page, from 1. |
| `list.pages` | integer | The count of list pages. An empty list has one page. |
| `list.last` | boolean | True on the last list page. |
| `list.empty` | boolean | True when the list has no rows. |

- A `row.*` field is valid only in a row text (section 5), in the list row of art (section 9)
  and in an `args` value of `select` (section 8).
  A field that the provider does not define is an error.
- A `list.*` field is valid only in a list screen.
- `row.selected`, `list.last` and `list.empty` are valid only in a condition.
  `row.selected` is valid only in a screen with `select`.
- The values follow [bbt-format](bbt-format.md), section 4.3.

## 7. Pages

- The page size P is `rows`. The host renders page 1 with P = `rows`. When it has R rows
  and R > `term.rows`, P is `rows - (R - term.rows)`, and at least 1.
  So the list fits the terminal, as the file area of [file-view](file-view.md), section 3.2.
- The page actions of [file-view](file-view.md), section 3.3, also show a list page of a list screen.
  [menu-nav](menu-nav.md), section 5.3, shows a list page.

## 8. Selection

`select` gives the screen a "cursor": one selected row of the list page.

| Key | Type | Meaning |
|---|---|---|
| `do` | string | The action ([bbt-format](bbt-format.md), section 8.3). |
| `args` | table | Optional. The arguments of a `goto` or a `gosub`. |

- An `args` value is a literal, `{arg.<name>}` or `{row.<field>}`
  ([bbt-format](bbt-format.md), section 8.5).
  For `{row.<field>}`, the host puts the value of the field of the selected row into the argument as it is.
  An integer gives its decimal form.
- The other rules of `args` of an item apply ([bbt-format](bbt-format.md), section 8.5).
  The values of a `{row.<field>}` argument are the values of the field in each row of the list
  of each screen set. A file screen checks each value ([file-view](file-view.md), section 3.1).
- A `{row.<field>}` value that reaches the `data` of a screen is an error
  ([bbt-format](bbt-format.md), section 8.5).

Keys:

- In a screen with `select`, the host takes the keys `UP`, `DOWN` and `ENTER`
  before the items ([menu-nav](menu-nav.md), section 3.3). An item with one of these keys is an error.
- `DOWN` selects the next row. On the last row of a list page, it shows the next list page
  with the cursor on its first row. On the last row of the list, it does nothing.
- `UP` selects the previous row. On the first row of a list page, it shows the previous list page
  with the cursor on its last row. On the first row of the list, it does nothing.
- `ENTER` runs `select` with the selected row. In an empty list, it does nothing.
- A key that does nothing is an ignored key ([menu-nav](menu-nav.md), section 7).
- A `page` action puts the cursor on the first row of the new list page.
- A row text shows the selected row with `{#if row.selected}`. Example: the brandbook mark
  `∙` and the style `pick` in section 3.

## 9. List row in art

- In art of a list screen, the art row with `row.*` field tokens is the "list row".
  An art file has at most one list row.
- The "list rectangle" has the columns from the first cell of the first `row.*` token
  to the last cell of the last `row.*` token, and `rows` art rows from the list row.
- Each cell of the list rectangle below the list row must be a space. Else the art gives an error
  with the position of the first such cell. A rectangle that ends below the art grid is an error.
- The cells outside the list rectangle stay, for example the borders of a frame.
- Art row `r + i` of the rectangle, with the list row `r`, shows the row `i + 1` of the list page:
  the cells of the list row in the rectangle, with the values of that row in its fields.
- A rectangle row without a row of the list page has spaces in the style of each cell of the list row.
  With P < `rows`, the rectangle rows after P have no row.
- The cells of the rectangle row of the selected row have the theme style `select`
  ([screens-toml](screens-toml.md), section 3.2). Art has no conditions in the list row
  ([art-format](art-format.md), section 5).

Example `screens/parts.ans` with `rows = 3` and `[fields]` `n = "list.page"`, `c = "list.pages"`:

```
 ╔══════════════════════════════════════════════╗
 ║ Parts                            page {n}/{c}║
 ║ { row.num} {row.title                       }║
 ║                                              ║
 ║                                              ║
 ╚═[PgDn]═[PgUp]═[Esc]══════════════════════════╝
```

## 10. Errors

Each of these is an error of `velourbbs check` ([screen-render](screen-render.md), section 3.1),
in each screen set:

- an error of the keys of section 3: `data` without `from`, an unknown provider, an error of the provider
  arguments, `rows` out of its range or missing, `rows` or `select` without `data`, `data` with `file`;
- a list screen without a list line and without a list row in art (section 3);
- a list screen as `goodbye` or `shutdown` ([menu-nav](menu-nav.md), section 6);
- an error of a provider row (section 4.1), as an error of the key `data`;
- an error of a list line (section 5) and of a field of section 6;
- an error of `select` and an item with the key `UP`, `DOWN` or `ENTER` in a screen with `select` (section 8);
- an error of the list row in art (section 9).

Positions:

- An error of a key of the screen has the position of its key path
  ([screen-render](screen-render.md), section 3.1).
- An error of a row of a provider value `{arg.<name>}` is an error of the item that gives the value,
  as in [file-view](file-view.md), section 7.
- An error of a value of `select` in a file screen is an error of the key `select`.
- An error of the list rectangle depends on `rows` of the screen. It starts with `screen "<name>": `
  ([screen-render](screen-render.md), section 3.1).

Examples:

```
screens/parts.bbt:4:8: data: from: unknown provider "file"
screens/parts.bbt:5:16: data: args: dir: directory "guide/bbtt" does not exist
screens/texts.bbt:9:16: item "2": args: dir: directory "guide/bbtt" does not exist
screens/parts.bbt:9:3: field "row.size": provider "files" has no field "size"
screens/parts.ans:4:3: screen "parts": list rectangle: cell is not a space
```

## 11. Code

| Package | Content |
|---|---|
| `internal/provider` | The providers and their rows. |
| `internal/bbt` | The parse of `data`, `rows`, `select` and of the list line. |
| `internal/screen` | The model of the list, the fields of section 6, the render of the list line and of the list row of art. |
| `internal/menu` | The list pages, the cursor and the update of a page. |
| `internal/term` | The output of the changed rows. |

- The model holds the rows of each provider value. It is immutable after the load.

## 12. Acceptance

- A test BBS in `cmd/velourbbs/testdata/` has a list screen with `files`, a value `{arg.dir}`,
  a list line in a block and in a fragment, `select` to a file screen, and art with a list row.
- Golden cell dumps of `velourbbs preview` show the first page, the last page, an empty list
  and a list line wider than W.
- Golden cell dumps show P smaller than `rows` at 80x10.
- Golden output of `velourbbs preview --keys` shows each `page` action, `UP` and `DOWN` over a page border
  and at the ends of the list, `ENTER` to a file screen and `back` to the same cursor.
- A test of telnet output shows a cursor move that sends only the changed rows.
- A test covers each error of section 10.
- A test covers the natural order and `README.md` first.

## 13. Not in this specification

- Lua rows: a Lua mod passes its rows in the data of `bbs.screen` ([lua-mods](https://github.com/schors/velourbbs/issues/12)).
- Providers of live rows, for example the last callers ([presence](https://github.com/schors/velourbbs/issues/81)).
- More than one list in a screen. A name in `{#each}` can add it later.
- Art as a background with fields at a row and a column. Fields are tokens in art
  ([art-format](art-format.md), section 4).
- A hotkey for each row.
- Format filters. Each user sets the date and the number formats later.
