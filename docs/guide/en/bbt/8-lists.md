# Task 8. Lists

In this task, you add the keeper's log to Lamp. The screen `log` lists the notes
of the directory `content/log/`. You do not write a row for each note:
the BBS reads the directory and makes the rows. The arrow keys choose a note,
and Enter opens it in the reader of task 7.

## The notes

Create the directory `content/log/` and seven notes in it:
`1-first-night.md`, `2-rain.md`, `3-north.md`, `4-wick.md`, `5-winter.md`,
`6-moths.md` and `7-tea.md`. Each note is a Markdown file with a level-1 heading.
The heading is the title of the note.

`examples/lamp/en/content/log/1-first-night.md`:
```
# The first night

I lit the lamp at nine. The modem sang its short song,
and the first caller came in at a quarter past ten.

We talked about the weather and old terminals.
At midnight I wrote this note and turned the wick down a little.
```

The other notes have the same form. Their texts are in `examples/lamp/en/content/log/`.
The note `5-winter.md` is longer: its reading time is 2 minutes.

## A list screen

A "list screen" is a screen with the key `data`. It shows the rows of a "provider" in pages.

| Key | Meaning |
|---|---|
| `data` | The provider: `from` and `args`. |
| `rows` | The most rows of a page, from 1 to 100. A list screen must have it. |

The provider `files` gives the content files of one directory:

```
data = { from = "files", args = { dir = "log" } }
rows = 5
```

- `dir` is a path in `content/`, as the key `file` of a file screen (task 7).
  A missing directory is an error.
- A row is each file of the directory. The BBS does not go into subdirectories.
  It ignores a name that is not a valid part of a path, for example `.keep`.
- The order: `README.md` first, then the "natural order". It compares the digits as numbers:
  `2-rain.md` comes before `10-art.md`.
- Each file is a content file: its errors are errors of `velourbbs check` (task 7).
- The BBS reads the rows at the start. Restart `velourbbs serve` after you add a note.
- A list screen cannot have the key `file`.

Full rules: [list-view](../../../specs/list-view.md), sections 3 and 4.

## The list line

A "list line" is a source line that starts with `{#each}` and ends with `{/each}`.
The text between them is the "row text": the BBS shows it once for each row of the page.

```
{#each}         {!dim}{row.num}{!}   {row.title                             }{!dim}{row.minutes} min{/each}
```

- The list line gives exactly one page of rows: one row for each row of the page,
  then empty rows. A page has `rows` rows, or fewer in a small terminal (section "Pages"). So the rows below the list stay in place on each page.
- A row does not wrap. The BBS cuts it at the width of the screen.
- The row text can have fields, style tokens and `{#if}` blocks.
  Its blocks must close in the same line. It cannot have an include or an art insert.
- Each row starts with the style before the list line.
  After the list line, the style is the style before it.
- A list line can be in a block and in a fragment. A screen has one list:
  each list line shows the same page.

The fields of a row and of the list:

| Field | Value |
|---|---|
| `row.num` | The number of the row in the list, from 1. |
| `row.title` | The title of the file, as `file.title` (task 7). |
| `row.minutes` | The reading time of the file. |
| `row.path` | The path of the file in `content/`, for example `log/2-rain.md`. |
| `row.selected` | True on the selected row (section "Selection"). Only in a condition. |
| `list.page` | The number of the current page, from 1. |
| `list.pages` | The page count. An empty list has one page. |
| `list.last` | True on the last page. Only in a condition. |
| `list.empty` | True when the list has no rows. Only in a condition. |

- A `row.*` field is valid only in the row text, in a value of `args` of `select`
  and in the list row of art (task 9).
- A `list.*` field is valid only in a list screen: in its text, its fragments, its art and its `when`.

Full rules: [list-view](../../../specs/list-view.md), sections 5 and 6.

## Selection

The key `select` gives a list screen a "cursor": one selected row of the page.
Enter runs the action of `select` with the selected row:

```
select = { do = "gosub reader", args = { file = "{row.path}" } }
```

- `do` is an action, as the `do` of an item (task 4).
- A value of `args` is a text or exactly one field `{row.<field>}`.
  Enter puts the value of the selected row into the argument.
  Here `{row.path}` gives the screen `reader` of task 7 the file of the note.
- The BBS loads the file of each row for `reader` at the start.
  An error in a file is an error of `velourbbs check`.
- `Up` and `Down` move the cursor. On the last row of a page, `Down` shows the next page.
  On the first row of a page, `Up` shows the previous page.
  At the ends of the list, the keys do nothing. Enter in an empty list does nothing.
- In a screen with `select`, an item cannot have the key `UP`, `DOWN` or `ENTER`.
- A `page` action puts the cursor on the first row of the new page.
  `back` from `reader` shows the same page and the same row.
- When the cursor moves on the same page, the BBS sends only the two rows that change.

`row.selected` is true on the selected row. The row text of `log` shows the sign `☼`
in `orn` before the number of the selected row, and its title in `title`:

```
{#each}{#if row.selected}       {!orn}☼{!} {#else}         {/if}{!dim}{row.num}{!}   {#if row.selected}{!title}{/if}{row.title                             }{!dim}{row.minutes} min{/each}
```

Both branches of the first block give 9 cells, so the number stays in column 10.

Full rules: [list-view](../../../specs/list-view.md), section 8,
and [menu-nav](../../../specs/menu-nav.md), sections 3.3 and 5.3.

## The screen `log`

`examples/lamp/en/screens/log.bbt`:
```
+++
data = { from = "files", args = { dir = "log" } }
rows = 5
select = { do = "gosub reader", args = { file = "{row.path}" } }
[[item]]
key = "SPACE"
do = "page next"
when = "not list.last"
[[item]]
key = "B"
do = "page prev"
when = "list.page > 1"
[[item]]
key = "ESC"
do = "back"
+++
{>_head}
       {!title}The keeper's log{!}
       {!dim}notes of the nights  ∙  page {list.page}/{list.pages}{!}

{#each}{#if row.selected}       {!orn}☼{!} {#else}         {/if}{!dim}{row.num}{!}   {#if row.selected}{!title}{/if}{row.title                             }{!dim}{row.minutes} min{/each}







  {!line}────────────────────────────────────────────────────────────────────────────{!}

    {!key}↑↓{!}  {!dim}choose{!}   {!key}Enter{!}  {!dim}read{!}   {#if item.SPACE}{!key}Space{!}  {!dim}next{!}   {/if}{#if item.B}{!key}B{!}  {!dim}prev{!}   {/if}{!key}Esc{!}  {!dim}back{!}
```

The screen:

```text

  ────────────────────────────────────────────────────────────────────────────

       Lamp  ☼ a BBS by lamplight

       The keeper's log
       notes of the nights  ∙  page 1/2

       ☼ 1   The first night                         1 min
         2   Rain on the roof                        1 min
         3   A caller from the north                 1 min
         4   A new wick                              1 min
         5   A long winter evening                   2 min







  ────────────────────────────────────────────────────────────────────────────

    ↑↓  choose   Enter  read   Space  next   Esc  back
```

- The rows of the list are the rows 9–13. The empty lines below the list line
  keep the hairline on row 21.
- `{row.title                             }` is a fixed-width field of 40 cells (task 5).
  So the reading time starts at column 54, as the notes of the menu items.

## Pages

The `page` actions of task 7 also work in a list screen.
`when` and `{#if item.<KEY>}` use `list.last` and `list.page`, as `file.last` and `file.page`.

The list has 7 rows, and a page has 5 rows. So page 2 has two notes and three empty rows.

When the screen does not fit the terminal, a page has fewer rows.
The BBS renders page 1 and counts its rows. When the screen is longer than the terminal
by `n` rows, a page has `rows - n` rows, at least 1. At 80x20 the screen `log` has 23 rows,
so a page has 2 rows, and the list has 4 pages:

```text
       The keeper's log
       notes of the nights  ∙  page 1/4

       ☼ 1   The first night                         1 min
         2   Rain on the roof                        1 min
```

A list screen has no More row.

An empty list has one page with empty rows. Show a text for it with `list.empty`:

```
{#if list.empty}       {!dim}No notes yet.{!}{/if}
```

Full rules: [list-view](../../../specs/list-view.md), section 7,
and [menu-nav](../../../specs/menu-nav.md), section 5.3.

## The screen `main`

The item `L` goes to the screen `log`. Its row is between the rows of `R` and `T`.
Remove one empty line above the row of `A` and one below the row of `G`.

`examples/lamp/en/screens/main.bbt`:
```
+++
[[item]]
key = "A"
do = "gosub about"
[[item]]
key = "R"
do = "gosub texts"
[[item]]
key = "L"
do = "gosub log"
[[item]]
key = "T"
do = "gosub term"
[[item]]
key = "G"
do = "logoff"
+++
{>_head}
       Good evening. It's quiet and warm here.
       Pull up a chair by the lamp.

         {!key}A{!}   About                                   {!dim}who we are{!}

         {!key}R{!}   Reading room                            {!dim}stories by the lamp{!}

         {!key}L{!}   Keeper's log                            {!dim}notes of the nights{!}

         {!key}T{!}   Your terminal                           {!dim}size and colours{!}

         {!key}G{!}   Goodbye                                 {!dim}the lamp will wait{!}

       {!dim}The light stays on till morning.{!}

  {!line}────────────────────────────────────────────────────────────────────────────{!}

    {!key}A R L T G{!}  {!dim}choose{!}
```

## A list with an argument

A value of `args` in `data` can be one argument of the screen: `{arg.<name>}`.
So one screen can show more than one directory:

```
data = { from = "files", args = { dir = "{arg.dir}" } }
```

```
[[item]]
key = "1"
do = "gosub notes"
args = { dir = "log" }
```

- The BBS reads the rows of each value of `dir` that a `goto` or a `gosub` gives.
- `velourbbs preview notes --set arg.dir=log` shows the rows of this value.

## Sample data

`velourbbs fields` (task 5) also lists the fields of the rows:

```text
row.num          integer  1   -      1
row.path         string   20  files  log/1-first-night.md
row.title        string   23  files  The first night
row.minutes      integer  1   files  1
```

The provider column names the provider of a `row.*` field.
The width of `row.title` is the widest title of the notes.

A "sample set" gives `velourbbs preview` other rows and other field values.
So you can test an empty list and a long title without new notes.
The sets are in the file `sample.toml` in the BBS directory.
`velourbbs fields --sample` writes this file with the set `default`:
the value of each field and one row. It does not replace an existing file.

Add your sets to the file:

```toml
[empty]
rows = []

[long]
fields = { "term.colors" = "16" }
rows = [
  { path = "log/2-rain.md", title = "A very long title of a note that does not fit in the row", minutes = 120 },
  { path = "log/3-north.md", title = "Short" },
]
```

- Each table is one set. A set name has lower case letters, digits, `_` and `-`.
- `fields` gives values to fields, as `--set`.
- `rows` replaces the rows of each list screen of the preview.
  A row gives the fields of the provider without `row.`. A missing field is empty or 0.
- `select` of `log` opens `path` in `reader`. So each `path` must be a content file.
- `velourbbs check` and `velourbbs serve` do not read `sample.toml`.

```
velourbbs preview log --sample empty
velourbbs preview log --sample long
```

The first command shows one page with empty rows. The second command shows the cut long title
in the 16 colours. `--set` replaces a value of the set, in each order of the options.

Full rules: [screen-render](../../../specs/screen-render.md), sections 3.2 and 3.5.

## Test the pages

```
velourbbs preview log
```

The command gives the two pages of the list, one after the other.
`--size 80x20` gives the four pages of the small terminal.

`--keys` moves the cursor, opens a note and turns the pages:

```
velourbbs preview main --keys L,DOWN,ENTER,ESC,SPACE,B,ESC
```

A cursor move gives the whole page, with the cursor on the new row.

## Result

- `velourbbs check` gives no output.
- `velourbbs preview log` gives two pages. Page 1 is the screen above.
- In the telnet client, `L` shows the log. `Up` and `Down` move the cursor,
  Enter opens the note, `Esc` goes back to the same row.
  `Space` and `B` turn the pages, `Esc` goes back to the menu.
