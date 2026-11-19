# BBT guide

This guide teaches you to write the screens of a Velour BBS.
It is for the artist and the screen author.
You need a text editor, a terminal and a telnet client.

## Introduction

### The BBS directory

A "BBS directory" holds the files of one BBS:

| File | Content |
|---|---|
| `bbs.toml` | The settings of the host. In this guide, it is empty. |
| `screens/screens.toml` | The settings of the screens: the width, the styles and other values. |
| `screens/<name>.bbt` | A "screen": a text file that the BBS shows. |
| `screens/_<name>.bbt` | A "fragment": a part of a screen. Its name starts with `_`. |

Run each `velourbbs` command in the BBS directory.

### The program `velourbbs`

Build the program with Go. In the source repository of Velour BBS, run:

```
go build ./cmd/velourbbs
```

The command writes the file `velourbbs` (`velourbbs.exe` on Windows).
Put it in a directory of your `PATH`. Or ask the sysop for the program.

### The example

The tasks build the example BBS "Lamp". Its files are in
[`examples/lamp/en/`](../../../../examples/lamp/en/).

- A line with a path and a colon shows a whole file of the example.
  Example: `` `examples/lamp/en/screens/main.bbt`: ``.
  Write this file in your BBS directory without `examples/lamp/en/`: `screens/main.bbt`.
- Each other block shows a part of a file, a command or an output.

Each task refers to the specifications for the full rules.

## Parts

- [Task 1. The BBS directory](1-bbs-directory.md)
- [Task 2. The first screen](2-first-screen.md)
- [Task 3. The header](3-header.md)
- [Task 4. A menu with hotkeys](4-menu.md)
- [Task 5. Fields and conditions](5-fields.md)
- [Task 6. The colour depth](6-colour-depth.md)
- [Task 7. Texts](7-texts.md)
- [Task 8. Lists](8-lists.md)
- [Task 9. Art](9-art.md)
- [Task 10. Errors](10-errors.md)

## Not in the MVP

The BBS does not have these features yet:

- conditions in art.
  In the MVP, only the `.bbt` files have conditions;
- marks (`{@<id>}`) and the choice of an item with the arrow keys.
  The user chooses an item only with its key;
- access checks (`require`). They need users;
- `ask` and a field token in an `args` value;
- punctuation keys, for example `?`;
- a label of an item. The BBS does not print the text of an item;
- the table of contents of a Markdown file, active links to other files,
  and syntax colours of source files;
- the display codes of other BBS programs: pipe codes, `@` codes and MCI codes.
  There is no converter.

Full rules: [bbt-format](../../../specs/bbt-format.md), section 10,
and [file-view](../../../specs/file-view.md), section 12.
