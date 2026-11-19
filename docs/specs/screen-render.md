# Screen render

This specification defines the screen model, the render and the commands
`velourbbs check` and `velourbbs preview`.
Related parts: [bbt-format](bbt-format.md), [art-format](art-format.md), [bbs-toml](bbs-toml.md),
[screens-toml](screens-toml.md), [menu-nav](menu-nav.md), [file-view](file-view.md),
[session-fields](session-fields.md) and [telnet-term](telnet-term.md).

## 1. Goal

The host loads the screen files ([bbt-format](bbt-format.md) and [art-format](art-format.md))
into an in-memory screen model at start.
It renders a screen for one terminal of any size. 80x24 is not a limit.
`velourbbs check` finds the errors in the files. `velourbbs preview` renders a screen to a "cell dump".

The host parses the screen files at load. There is no compile step and no cache.

## 2. Render

The host renders a screen by these rules of [bbt-format](bbt-format.md):

- the render width and the screen area: section 6.1;
- the rows of the body and of the art: section 6.2;
- the text wrap: section 6.3;
- the pages and the More row: section 6.4; the pages of a file screen:
  [file-view](file-view.md), section 3.2;
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
- One file gives at most 100 errors and warnings. Then it gives one error `too many diagnostics`.
  The loader drops the rest, and the load of the file goes on.
- A TOML document (`bbs.toml`, `screens/screens.toml` or a front matter) gives one syntax
  or type error; else all unknown keys. Only a document without these gives its other errors.
- A key that differs from a known key only in case is an unknown key. A table `[item]`
  in place of `[[item]]` gives `item: must be an array of tables` with the unknown keys.
  If a key that differs only in case has a type error, the document gives only
  these two kinds of errors.
- A syntax error has the position that go-toml gives. A type error and an unknown key
  have the position of the key or value. The message of a type error starts with
  the dotted key. In an inline table, the key ends at the table.
- A key that differs only in case, `[item]` and a type error in a nested array
  have the position of the first line of the document.
- Each other error of a TOML document has the position of the first line of the document:
  1:1 in `bbs.toml` and `screens/screens.toml`, 2:1 in a front matter. The message names the key.
  Examples: an invalid style, item key, `do`, address or `width`, an integer out of its range,
  an invalid alias or style name, an error of `when`, an error between keys, or between a key
  and another file or screen.
- An error has the form `<file>:<row>:<col>: <message>`, with 1-based numbers.
  In a `.bbt` file, the column counts code points.
  For art, the row and the column are grid positions.
  A warning has `warning: ` before the message.
- A diagnostic escapes each invalid UTF-8 byte and each rune that is not printable,
  as a Go quoted string does. Example: a file name with ESC gives `\x1b`.
- A check of a fragment or an art file can depend on the screen that uses it:
  on its items or its width limits. Such an error starts with `screen "<name>": `.
- The host runs the same checks at start. An error stops the start.

Examples:

```
screens/main.bbt:2:1: item "S": when: unknown field "term.colours"
screens/logo.ans:1:1: not in the flat form, convert it with velourbbs art import
screens/main.ans:4:15: token "{w}" changes the foreground: 11 at col 15, 15 at col 16
screens/about.bbt:4:1: warning: art "logo.ans" is 132 cells wide, max width is 100
```

### 3.2. `velourbbs preview`

`velourbbs preview <name> [--size <cols>x<rows>] [--set <path>=<value>]...` renders a screen
and writes its cell dump to the standard output.

- It does the checks of `velourbbs check` first. After an error it writes no dump
  and exits with status 1.
- The default size is 80x24. `--size` is the only way to set the size.
- `--set` sets `term.type`, `term.colors`, `term.name`, `term.images`, `term.ambiguous`,
  `term.bg_color`, `term.fg_color`, `term.background`,
  `arg.<name>` or a field of [session-fields](session-fields.md), section 4.
  Example: `--set term.colors=256`.
  The defaults are `xterm`, 256, empty, empty, 1, empty, empty and empty. `term.charset` is `utf-8`.
  For `term.name`, the host removes the control characters and cuts the text to 40 graphemes
  ([telnet-term](telnet-term.md), section 6), with no case change and no default.
  `term.images` is `sixel` or empty. `term.ambiguous` is `1` or `2`.
  The layout does not use it.
- `term.bg_color` and `term.fg_color` are `#rrggbb` or empty. The host makes them lower case.
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

### 3.3. Cell dump

The cell dump is the golden file for the tests. It has one part for each page.
A file screen has one part for each file page ([file-view](file-view.md), section 3.2).
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

- The first line gives the page number (`file.page` on a file screen), the terminal size and the value of `fill`
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
