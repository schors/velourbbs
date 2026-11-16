# BBT guide

This specification defines the BBT guide: the guide to the screen files for the artist
and the screen author.
Related parts: [bbt-format](bbt-format.md), [art-format](art-format.md),
[art-import](art-import.md), [screens-toml](screens-toml.md), [menu-nav](menu-nav.md)
[screen-render](screen-render.md), [telnet-term](telnet-term.md)
and [example-style](example-style.md).

## 1. Goal

The artist and the screen author learn to write screens without the specifications.
The guide is in the MVP.

## 2. Reader

- The reader is the artist and the screen author.
- The reader can edit a text file and run a command in a terminal.
- The reader does not read the specifications. Each task refers to the sections
  of the specifications for the full rules.

## 3. Files

| Path | Content |
|---|---|
| `docs/guide/en/bbt.md` | The guide in English. It is the source. |
| `docs/guide/ru/bbt.md` | The guide in Russian. It is a translation of the source. |
| `examples/lamp/<lang>/` | The example BBS Lamp: [example-style](example-style.md). |
| `examples/art/` | The art example of section 5. |

- A change of a specification that a task of section 4 refers to also changes the guide,
  when the change affects what the task shows.
- Change the two languages in the same commit.
- The Russian guide shows the files of `examples/lamp/ru/`. The English guide shows
  the files of `examples/lamp/en/`. Both show the same files of `examples/art/`.
  The guide does not show the `.ans` files.
- The guide shows each screen as plain text, as [example-style](example-style.md), section 6.
  It names the styles of the text in words.

## 4. Tasks

The guide has these tasks in this order. Each task ends with a check of the result.

| Task | Content | Files of the example |
|---|---|---|
| 1. The BBS directory | The start directory: an empty `bbs.toml` and `screens/main.bbt` with one item `G` `logoff`. `velourbbs check` and the form of its errors and warnings ([screen-render](screen-render.md), section 3.1). `velourbbs preview <name>` and the cell dump: the rows and the style runs (section 3.3). `velourbbs serve` on `:2323` with PuTTY or Tera Term: they show UTF-8 and 256 colours ([telnet-term](telnet-term.md), sections 3 and 6). Restart `velourbbs serve` after each change. | — |
| 2. The first screen | Text lines, empty lines, the style tokens `{!name}` and `{!}`, the theme `[theme.default]`, `[screen] width` and `fill`. The item `G` of task 1 stays until task 4. | — |
| 3. The header | A fragment and an include. | `screens/_head.bbt` |
| 4. A menu with hotkeys | `[[item]]`, `key`, `do`, `goto` and `logoff`. The key row is text of the screen. The screen `goodbye`. `velourbbs preview --keys`. | `screens/main.bbt`, `screens/about.bbt`, `screens/goodbye.bbt` |
| 5. Fields and conditions | The fields `term.*`, `{#if}` and `{#else}`. A fixed-width field. `when` and `item.<KEY>`. | `screens/term.bbt` |
| 6. The colour depth | `[theme.256]` and `[theme.16]`. `velourbbs preview --set term.colors=<n>`. In `velourbbs serve`, the terminal type of the client gives the colour depth ([telnet-term](telnet-term.md), section 6). | `screens/screens.toml` |
| 7. Art | The ready art `screens/lamp.ans` and the art insert `{><name>.ans}`. A field in art: its token has one background and one foreground ([art-format](art-format.md), section 4). To change the art, use any editor. Convert the file of an art editor with `velourbbs art import` ([art-import](art-import.md), sections 2 and 3). | `examples/art/bbs.toml`, `examples/art/screens/screens.toml`, `examples/art/screens/main.bbt` |
| 8. Errors | Examples of frequent errors of `velourbbs check` and their fixes. | — |

- A task shows a part of a file as a snippet before it shows the whole file.
- The guide also shows these features in snippets: `{{`, `align`, `valign`, the More row,
  an alias of `[fields]`, and the art with a fragment for small terminals
  ([bbt-format](bbt-format.md), section 6.1).
- The guide names the features of [bbt-format](bbt-format.md), section 10, as not in the MVP.

### 4.1. File blocks

- A "file block" is a fenced code block after a line that holds only a path in backticks and a colon.
  The path is relative to the repository root.
  Example: `` `examples/lamp/en/screens/main.bbt`: ``.
- The fence starts on the line after the path line.
- A file block holds the whole file: the file equals the lines between the fences, each with LF.
- Each other code block is a snippet.

## 5. Art example

| File | Content |
|---|---|
| `examples/art/bbs.toml` | Empty. |
| `examples/art/screens/lamp.ans` | The art in the flat form ([art-format](art-format.md), section 1). |
| `examples/art/screens/screens.toml` | The styles `normal` and `key`. |
| `examples/art/screens/main.bbt` | The art insert, the key row `G  goodbye` and the item `G` `logoff`. |

- The art is a lamp, at most 40 columns and 12 rows. It has the field `{term.type}`.
- The texts are in English.

## 6. Acceptance

- A Go test in `cmd/velourbbs` reads `docs/guide/en/bbt.md` and `docs/guide/ru/bbt.md`.
  It compares each file block with its file. A missing file is a failure.
- A Go test in `cmd/velourbbs` checks `examples/art/`:
  - `velourbbs check` gives no output;
  - `velourbbs preview main --size 80x24 --set term.colors=256` equals the golden file
    `cmd/velourbbs/testdata/art/main-80x24-256.dump`.
- The golden dumps of Lamp do not change.
- Each task of section 4 is in both languages, in the same order.
