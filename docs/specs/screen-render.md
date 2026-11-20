# Screen render

This specification defines the screen model, the render and the commands
`velourbbs check` and `velourbbs preview`.
Related parts: [bbt-format](bbt-format.md), [art-format](art-format.md), [bbs-toml](bbs-toml.md),
[screens-toml](screens-toml.md), [menu-nav](menu-nav.md), [file-view](file-view.md), [list-view](list-view.md),
[session-fields](session-fields.md) and [telnet-term](telnet-term.md).

## 1. Goal

The host loads the screen files ([bbt-format](bbt-format.md) and [art-format](art-format.md))
into an in-memory screen model at start.
It renders a screen for one terminal of any size. 80x24 is not a limit.
`velourbbs check` finds the errors in the files. `velourbbs preview` renders a screen to a "cell dump",
or to the "line form" for the terminal of the sysop.

The host parses the screen files at load. There is no compile step and no cache.

## 2. Render

The host renders a screen by these rules of [bbt-format](bbt-format.md):

- the render width and the screen area: section 6.1;
- the rows of the body and of the art: section 6.2;
- the text wrap: section 6.3;
- the pages and the More row: section 6.4; the pages of a file screen:
  [file-view](file-view.md), section 3.2; the pages of a list screen:
  [list-view](list-view.md), section 7;
- the vertical position and the fill rows: section 6.5;
- the cells and the styles: section 7.

The host renders a screen page by page. A session holds one page in memory.

## 3. Commands

- A missing or unknown command, or an invalid argument of a command, is a usage error.
  `velourbbs` exits with status 2 after a usage error.
- `velourbbs art import` converts art to the flat form: [art-import](art-import.md).

### 3.1. `velourbbs check`

- `velourbbs check` reads `bbs.toml`, `screens/screens.toml` and each screen file in `screens/`
  and in the directory of each language ([session-language](session-language.md), section 3).
  It checks each screen with the fragments, the art and the content files that the screen uses
  ([file-view](file-view.md), section 7), in each screen set ([session-language](session-language.md), section 6).
- It warns about a fragment or an art file that no screen uses.
  After the limit error of a body ([bbt-format](bbt-format.md), section 9),
  the warning can be false: the load does not read the rest of the screen.
- It reports all errors and warnings, not only the first. A TOML document is an exception (below).
  It writes them to the standard error. It exits with status 1 after an error.
- An error of the load of `bbs.toml` or `screens/screens.toml` stops the check before the screens.
  The screens depend on these settings, so their errors can be false.
  The checks between files, for example of the start screen, run after the screens.
- One file gives at most 100 errors and warnings. Then it gives one error `too many diagnostics`.
  The loader drops the rest, and the load of the file goes on.
  The layout warnings of section 3.6 do not count.
- A TOML document (`bbs.toml`, `screens/screens.toml` or a front matter) gives one syntax
  or type error; else all unknown keys. Only a document without these gives its other errors.
- A key that differs from a known key only in case is an unknown key. A table `[item]`
  in place of `[[item]]` gives `item: must be an array of tables` with the unknown keys.
  If a key that differs only in case has a type error, the document gives only
  these two kinds of errors.
- A syntax error, a type error and an unknown key have the position that go-toml gives.
  The message of a type error starts with the dotted key. In an inline table,
  the key ends at the table.
- An error of the whole document has the position of its first line:
  1:1 in `bbs.toml` and `screens/screens.toml`, 2:1 in a front matter.
  Examples: the file is missing, is not a regular file or is over its size limit.
- The message of each other error of a TOML document starts with a key path, for example
  `theme.default.gold: ` or `item "S": do: `. The item label of [menu-nav](menu-nav.md),
  section 6, names the table of the item.
- The error has the position of the value at its key path. This includes an error
  between keys, or between a key and another file or screen.
  Examples: an invalid style, item key, `do`, address or `width`, an integer out of its range,
  an error of `when`, `lines: the screen has no file`.
- An error of the name of a key has the position of the key: a key that differs
  from a known key only in case, an invalid alias name, an invalid or unknown style name, `[item]`.
- A table has the position of its header, of the `{` of an inline table,
  or of the first dotted key that makes it.
- A key path to a key that is not set has the position of the last table on the path
  that the document sets. The root table has the position of the first line of the document.
  Examples: `item 3: key is missing` at the header of the third `[[item]]`,
  `menu: start: unknown screen "main"` at 1:1 without the table `[menu]`.
- The decode of go-toml gives the positions of its own errors only, not of the keys and
  the values. The parser `unstable.Parser` of go-toml gives them for the other errors.
  This API does not promise backward compatibility. `go.mod` pins the version,
  and the tests of the positions find a change.
- An error has the form `<file>:<row>:<col>: <message>`, with 1-based numbers.
  In a `.bbt` file, the column counts code points.
  For art, the row and the column are grid positions.
  A warning has `warning: ` before the message.
- A diagnostic escapes each invalid UTF-8 byte and each rune that is not printable,
  as a Go quoted string does. Example: a file name with ESC gives `\x1b`.
- A check of a fragment or an art file can depend on the screen that uses it:
  on its items, its width limits or its `rows` ([list-view](list-view.md), section 9).
  Such an error starts with `screen "<name>": `.
- The host runs the same checks at start. An error stops the start.
- `velourbbs check --size <cols>x<rows>` also gives the layout warnings of section 3.6.

Examples:

```
screens/main.bbt:9:8: item "S": when: unknown field "term.colours"
screens/screens.toml:22:10: theme.default.gold: invalid style "#b8902g": invalid colour "#b8902g"
screens/logo.ans:1:1: not in the flat form, convert it with velourbbs art import
screens/main.ans:4:15: token "{w}" changes the foreground: 11 at col 15, 15 at col 16
screens/about.bbt:4:1: warning: art "logo.ans" is 132 cells wide, max width is 100
```

### 3.2. `velourbbs preview`

`velourbbs preview <name> [--size <cols>x<rows>] [--sample <set>] [--set <path>=<value>]... [--keys <keys>] [--view]`
renders a screen and writes its cell dump to the standard output.

- It does the checks of `velourbbs check` first. After an error it writes no dump
  and exits with status 1.
- The default size is 80x24. `--size` is the only way to set the size.
- `--set` sets `term.type`, `term.colors`, `term.name`, `term.ambiguous`,
  `term.bg_color`, `term.background`,
  `arg.<name>` or a field of [session-fields](session-fields.md), section 4.
  Example: `--set term.colors=256`.
  The defaults are `xterm`, 256, empty, 1, empty and empty.
  For `term.name`, the host removes the control characters and cuts the text to 40 graphemes
  ([telnet-term](telnet-term.md), section 6), with no case change and no default.
  `term.ambiguous` is `1` or `2`.
  The layout does not use it.
- `term.bg_color` is `#rrggbb` or empty. The host makes it lower case.
  `--set term.bg_color` gives `term.background` by the rule of [telnet-term](telnet-term.md),
  section 6. `term.background` is `dark`, `light` or empty.
  `--set term.background` overrides it, in each order of the options.
- Each argument of the screen needs `--set arg.<name>=<value>`
  ([bbt-format](bbt-format.md), section 8.5). A missing argument, another argument and
  an invalid value are invalid arguments. A content file of a value has the checks of
  [file-view](file-view.md), section 7. After an error, the command writes no dump and exits with status 1.
- `--set term.type` applies the rules of [telnet-term](telnet-term.md), section 6:
  it cleans the name, cuts it and makes it lower case. An empty name gives `ansi`.
  The name gives `term.colors` by rules 3–6. `--set term.colors` overrides it, in each order of the options.
- `item.<KEY>` comes from the items of the screen and their `when`.
- A size outside the ranges of [bbt-format](bbt-format.md), section 4.1, and an unknown screen
  are invalid arguments.
- `--keys` runs a session: [menu-nav](menu-nav.md), section 7.
- `--view` writes each page in the line form (section 3.7) in place of the cell dump.
- `--sample <set>` takes the values of a sample set (section 3.5). `--set` replaces them,
  in each order of the options.
- A list screen shows the rows of its provider ([list-view](list-view.md), section 4).
  For a `--set arg.<name>` value of `data`, the command reads the rows of the value.
  An error of the provider is an invalid argument.

### 3.3. Cell dump

The cell dump is the golden file for the tests. `go test <package> -update` rewrites
the golden files of `cmd/velourbbs` and `internal/screen`. The dump has one part for each page.
A file screen has one part for each file page ([file-view](file-view.md), section 3.2).
A list screen has one part for each list page ([list-view](list-view.md), section 7),
with the cursor on the first row of each page.
This dump shows the screen `+++` `[[item]]` `key = "G"` `do = "logoff"` `+++`
with the body `{!title}Welcome!{!}` and `{!14}[G]oodbye`, the theme of
[screens-toml](screens-toml.md), section 4, and the other keys of `[screen]` at their defaults:

```
page 1 80x24 fill none
|Welcome!                                                                        |
|[G]oodbye                                                                       |
styles
1: 1-8 15/4 bold, 9-80 7/0
2: 1-9 14/0, 10-80 7/0
```

- The first line gives the page number (`file.page` on a file screen, `list.page` on a list screen), the terminal size and the value of `fill`
  ([bbt-format](bbt-format.md), section 6.5).
- Each rendered row of the page is between two `|`. A wide grapheme appears once
  and takes 2 columns.
- Each line of `styles` gives the terminal row number and the runs of equal styles,
  in the syntax of [bbt-format](bbt-format.md), section 7.2, with 1-based columns.
  A run has the form `<first>-<last> <style>`, also for one column. Example: `5-5 7/0`.
- The dump has the colours of the model with the session theme
  ([screens-toml](screens-toml.md), section 3.2), before the reduction to `term.colors`.
  A VGA colour appears as its index. The tests of the output
  ([telnet-term](telnet-term.md), section 7.3) check the VGA colours.
- A page with a More row includes it as its last row.

### 3.4. `velourbbs fields`

`velourbbs fields [--sample]` writes the fields of the BBS to the standard output.

- It does the checks of `velourbbs check` first. After an error it writes nothing
  and exits with status 1.
- It writes one line for each field: the path, the type, the width, the provider and the sample value,
  in columns separated with two or more spaces. The sample value is the rest of the line.
  The provider is the provider of a `row.*` field, else `-`.
- The fields are each path of the registry ([bbt-format](bbt-format.md), section 4.1)
  without `<...>`, in the order of the registry: `term.*`, `file.*`, `list.*`, `session.*` and `bbs.*`.
  Then `arg.<name>` for each argument of a screen, in name order.
  Then `row.num` and the fields of each provider that a screen uses ([list-view](list-view.md), section 4),
  the providers in name order.
- The "width" is the most cells that a value can have:
  - an integer with a range or a fixed set of values: the digits of the widest value.
    Example: `term.cols` has 3, `term.colors` has 8;
  - a string with a fixed set of values: the widest value. Example: `term.background` has 5.
    The set of `session.lang` is the languages of the BBS;
  - a `row.*` field: the widest value of the rows of the BBS. `row.num`, `list.page`
    and `list.pages`: the digits of the largest row count;
  - else `-`.
- The "sample value" is the value of `velourbbs preview` without `--set` (section 3.2,
  [session-fields](session-fields.md), section 4). The screen set of `[screen] lang` gives the first
  values, with the screens in name order. For `arg.<name>`, it is the first value
  that a `goto` or a `gosub` gives: the items in source order, then `select` with the first row
  of the screen. For a `row.*` field, it is the value of the first row of the provider, else empty.
- `file.*` and `list.*` have the values of page 1 of 1: `file.title` and `file.section` are empty,
  `file.minutes` is 0, and `list.empty` is false when a list of the BBS has a row.

Example:

```
term.cols      integer  3   -      80
term.type      string   -   -      xterm
bbs.sessions   integer  -   -      1
row.title      string   34  files  Introduction
```

- `--sample` writes the file `sample.toml` (section 3.5) with the set `default` and exits.
  The set has `fields` with the sample value of each field that `--set` takes, except `arg.<name>`,
  and `rows` with one row: the sample value of each field of each provider that a screen uses.
  An existing file is an error: the command does not replace it.

### 3.5. Sample sets

`sample.toml` in the BBS directory holds the "sample sets" of `velourbbs preview`.
Each top-level table is one set. The set name matches `[a-z][a-z0-9_-]*`.

| Key | Type | Meaning |
|---|---|---|
| `fields` | table of strings | Optional. A value for each path, with the rules of `--set` (section 3.2). |
| `rows` | array of tables | Optional. The rows of each list screen of the preview: the screen of the command, and each list screen of `--keys`. A table maps a field name of the provider, without `row.`, to a string or an integer. A missing field is empty or 0. Without `rows`, a list screen shows the rows of its provider. |

Example:

```toml
[empty]
rows = []

[long]
fields = { "bbs.sessions" = "12" }
rows = [
  { title = "A very long title of a guide part that does not fit", minutes = 120 },
  { title = "Short", minutes = 1 },
]
```

- `velourbbs check` and `velourbbs serve` do not read `sample.toml`.
- A missing file, an invalid set name, an unknown set, an unknown key and an invalid value
  are invalid arguments of `velourbbs preview`. The command checks the names, the keys and the types
  of each set, and the `fields` values of the set of `--sample`.
- A field of `rows` that the provider of a list screen of the preview does not have,
  or a value of the wrong type, is an error of the preview. After an error of the preview,
  the command exits with status 1.
- A value of `rows` that `select` gives to a file screen must be a content file
  ([list-view](list-view.md), section 8). Else it is an error of the preview.

### 3.6. Layout warnings

`velourbbs check --size <cols>x<rows> [--size <cols>x<rows>]...` renders each screen at each size
and warns about layout mistakes that the host does not report.

- A size outside the ranges of [bbt-format](bbt-format.md), section 4.1, is an invalid argument.
  A repeated size counts once.
- The command renders only after the checks of section 3.1 give no error.
  `velourbbs serve` does not render at start.
- It renders each screen of each screen set ([session-language](session-language.md), section 6)
  at each size. `session.lang` is the language of the screen set. The other fields have
  the values of `velourbbs preview` without `--set` and `--sample` (section 3.2).
  So the conditions and the `when` of the items see the size, as in a session.
- The arguments of a screen are the arguments of the first action that shows the screen:
  the screen set of `[screen] lang` first, then the other screen sets in the order of the codes;
  in each set, the screens in name order, the items in source order, then `select`
  with the first row of its list. The list of a screen with arguments needs the arguments
  of an action before its `select`.
  A screen with arguments that no action shows is not rendered.
- It renders each page of the screen, as the cell dump has (section 3.3).

The warnings. Each row of the table is one "kind".

| Kind | Warning | When | Position |
|---|---|---|---|
| Wrap | `the line wraps to <n> rows` | A text line gives more than one row ([bbt-format](bbt-format.md), section 6.3). | The source line. |
| More | `<n> rows, the screen gets a More row` | A screen other than a file screen and a list screen has more rows than `term.rows` ([bbt-format](bbt-format.md), section 6.4). | The first line of the screen file. |
| Height | `<n> rows, the rows after <rows> are cut` | A page of a file screen or a list screen has more rows than `term.rows` ([file-view](file-view.md), section 3.2; [list-view](list-view.md), section 3). | The first line of the screen file. |
| Width | `the row is <n> cells wide, the screen shows <m>` | The host cuts a row, and the cut drops a grapheme other than a space. Rows: a row of text, a fixed-width field wider than W, a stretch line, a list line and a row of the file area. | The source line; for the file area, the line of `{>file}`. |
| Art | `art "<name>.ans" is <n> cells wide, the screen shows <m>` | An art insert is wider than `<m>`, and `<m>` is less than `max` of the width limits. Section 3.1 already warns about art wider than `max`. | The line of the art insert. |

- `<m>` is the smaller of W and `term.cols` ([bbt-format](bbt-format.md), section 6.1).
- `<n>` of Width is the cells of the row before the cut, up to its last grapheme other than a space.
  A fixed-width field wider than W has its own width in it. A position with more than one row
  on a page gives the widest row.
- The column of each position is 1.
- The message starts with `at <cols>x<rows>: `. In a fragment, `screen "<name>": `
  comes before it (section 3.1).
- Each position, screen, size, kind and screen set gives at most one warning,
  with the values of the first page where it happens.
  Then the rules of [session-language](session-language.md), section 6, apply:
  an equal warning appears once, and a warning of a common file that only some screen sets give
  ends with ` (language "<code>")`.
- A warning does not change the exit status.
- The file area of a file screen gives only the warning Width at a cut at `term.cols`.
  The text of a Markdown file wraps by design, and its code rows have a cut mark.
- The More row gives no warning of its own. The host cuts its text at the screen area,
  and the warning More already names the screen.

Example at 80x24 with `width = [68, 108]` and `gutter = 6`:

```
screens/main.bbt:12:1: warning: at 80x24: the line wraps to 2 rows
screens/_logo.bbt:3:1: warning: screen "main": at 80x24: the row is 75 cells wide, the screen shows 68
screens/about.bbt:1:1: warning: at 80x24: 31 rows, the screen gets a More row
```

### 3.7. Line form

The line form shows the pages in colour in the terminal of the sysop.
The page output of the host ([telnet-term](telnet-term.md), section 7.1) clears the terminal screen
and puts each row at its position, so the next page replaces the page before it.
The line form writes the pages one after another, and each page stays in the scrollback.

For each page, the line form writes:

1. The first line of the cell dump of the page (section 3.3), then a line end.
2. For each terminal row from the top, with the terminal rows of [bbt-format](bbt-format.md),
   section 6.5:
   - a row of the page: the cells of the row;
   - a fill row with `fill = "none"`: no cells;
   - a fill row with `fill = "spaces"` or `"erase"`: `term.cols` spaces in `normal`.

   Then `ESC[0m` and a line end.

- A line end is LF.
- Before a cell whose style differs from the style of the previous cell, the line form
  sends the SGR of [telnet-term](telnet-term.md), section 7.2. Each row starts with `default` colours
  without attributes.
- The colours have the reduction to `term.colors` of [telnet-term](telnet-term.md), section 7.3.
  A VGA colour becomes its RGB value, as for a terminal whose palette is not VGA.
  So with 256 and 16777216 colours, the palette of the terminal of the sysop does not change
  the colours of the art. With 16 colours, the art has the indexes 0–15 in that palette.
- The encoder of [telnet-term](telnet-term.md), section 7.4, writes the cells.
  The re-sync moves to the column with `ESC[<col>G` in place of `ESC[<row>;<col>H`.
  The re-sync is not an SGR sequence: a viewer that passes only SGR, such as `less -R`,
  shows it as text.
- The line form has no synchronized output, no clear and no positions of rows.
- The `ESC[0m` before the line end stops a terminal with bce from filling
  the next row in the background of the last cell.
- When the terminal of the sysop shows a row wider than its columns, the row wraps
  to extra lines, and the columns after a re-sync are not correct. Examples: a terminal
  with fewer columns than `term.cols`, a grapheme wider than in the model, ambiguous characters
  shown wide. No later output overwrites the extra lines: the wrap rule of
  [telnet-term](telnet-term.md), section 7.4, does not apply.

Example: the page of section 3.3, with the colours of the example, at 80x10 with `fill = "none"`.
`␛` is ESC. `...` stands for the spaces up to column 80. The terminal rows 3–10 are fill rows:

```
page 1 80x10 fill none
␛[0;1;97;44mWelcome!␛[0;37;40m    ...␛[0m
␛[0;96;40m[G]oodbye␛[0;37;40m    ...␛[0m
␛[0m
␛[0m
␛[0m
␛[0m
␛[0m
␛[0m
␛[0m
␛[0m
```

## 4. Code

- The host language is Go. The Go module is `github.com/schors/velourbbs`.
- The model is immutable after the load. Sessions read it at the same time without locks.
- The checks are `go vet ./...`, `go test ./...` and `golangci-lint run ./...`
  (golangci-lint 2.15.0, configuration in `.golangci.yml`).

## 5. Libraries

| Task | Choice | Compared with | Reason |
|---|---|---|---|
| TOML | `github.com/pelletier/go-toml/v2` | `github.com/BurntSushi/toml` | The decode into structs with `DisallowUnknownFields` gives the positions of unknown keys and type errors (section 3.1). `BurntSushi/toml` gives no positions of unknown keys. |
| Graphemes and width | `github.com/clipperhouse/displaywidth` | `github.com/rivo/uniseg`, `github.com/mattn/go-runewidth` | One iterator gives each grapheme with its width, the cell model. The release has Unicode 17; `uniseg` has Unicode 15. `go-runewidth` measures runes. |
| ANSI art and SAUCE | Own parser | `github.com/charmbracelet/x/ansi` | The loader needs SGR only. The import needs about 15 sequences and the editor rules of [art-import](art-import.md). A terminal parser adds a large dependency. |
| Conditions | Own parser | `github.com/expr-lang/expr` | The grammar has 7 rules and static types from the registry. `expr` adds arithmetic and functions that the format excludes. |
| Command line | `flag` | `github.com/spf13/cobra` | The standard library is sufficient for the commands of section 3. |

- The Go standard library and `golang.org/x/text` have no grapheme segmentation.

## 6. Not in the MVP

- Reload: [reload](https://github.com/schors/velourbbs/issues/14). A restart is sufficient for the MVP.
