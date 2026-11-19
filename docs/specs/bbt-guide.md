# BBT guide

This specification defines the BBT guide: the guide to the screen files for the artist
and the screen author.

## 1. Goal

The artist and the screen author learn to write screens without the specifications.
The guide is in the MVP.

## 2. Reader

- The reader is the artist and the screen author.
- The reader can edit a text file and run a command in a terminal.
- The reader does not read the specifications. A part links to the sections
  of the specifications that hold its full rules.

## 3. Files

| Path | Content |
|---|---|
| `docs/guide/en/bbt/` | The guide in English, in parts (section 3.1). It is the source. |
| `docs/guide/ru/bbt/` | The guide in Russian. It is a translation of the source. |
| `examples/lamp/<lang>/` | The example BBS Lamp ([example-style](example-style.md)). |
| `examples/art/` | The art example: a lamp of at most 40 columns and 12 rows with the field `{term.type}`. |

- A specification that the guide teaches says so in its header. A change of a section
  that a part links to also changes the part, when the change affects what the part shows.
- Change the two languages in the same commit.
- The guide in `<lang>` shows the files of `examples/lamp/<lang>/`. Both show the files
  of `examples/art/`. The guide does not show the `.ans` files.
- After each task, `velourbbs check` gives no output in the BBS directory of the reader.
- The guide shows each screen as plain text. It names the styles of the text in words.
- The `README.md` names the features not in the MVP: those of [bbt-format](bbt-format.md), section 10,
  and the table of contents, the active links and the syntax colours of [file-view](file-view.md), section 12.

### 3.1. Parts

A guide is a directory of Markdown files, the "parts": `README.md` and one part `<n>-<slug>.md`
for each task.

- The parts of this guide: `1-bbs-directory.md`, `2-first-screen.md`, `3-header.md`,
  `4-menu.md`, `5-fields.md`, `6-colour-depth.md`, `7-texts.md`, `8-lists.md`, `9-art.md`, `10-errors.md`.
- Both languages have the same parts.
- The first line of a part is a level-1 heading: its title. Outside the code blocks,
  a part has no other level-1 heading.
- A relative link is relative to the part file.
- Each task ends with a check of the result.

## 4. File blocks

- A "file block" is a fenced code block after a line that holds only a path in backticks and a colon.
  The path is relative to the repository root. Example: `` `examples/lamp/en/screens/main.bbt`: ``.
- The fence starts on the line after the path line.
- A file block holds the whole file: the file equals the lines between the fences, each with LF.
- A file block shows the last form of a file. Each earlier form, and a part of a file, is a snippet.

## 5. Acceptance

- A Go test in `cmd/velourbbs` compares each file block of the parts with its file.
  A missing file is a failure.
- The test checks the parts of each guide: the list of section 3.1, the same parts
  in both languages, and the level-1 heading rule.
- A Go test checks `examples/art/`: `velourbbs check` gives no output, and
  `velourbbs preview main --size 80x24 --set term.colors=256` equals its golden file.
