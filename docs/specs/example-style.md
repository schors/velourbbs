# Example style

This specification defines the style of the example BBS "Lamp".
The guides build this BBS: [setup-guide](setup-guide.md) and [bbt-guide](bbt-guide.md).
The files of `examples/lamp/` define its screens, styles and texts.

## 1. Goal

The sysop reads the setup guide and builds a BBS that looks good at once.
The style is very simple, gentle and warm. The screen sources stay plain text that the sysop
can edit without counting cells.

## 2. Principles

- The example style is not a second brandbook. It does not use the [brandbook](../brand/brandbook.md).
- The style uses only what the engine can do. The screens have no art and no aliases.
- A dark warm ground, cream text and one warm accent for the keys.
  256 colours is the main mode. 16 colours are grey.
- One ornament `☼` on a screen: the light of the lamp. The list screen `log`
  also marks its selected row with it.
- Much empty space. A screen has few lines. The keys of a screen are always visible in the key row.
- The screen is 80 columns wide. The rows of the terminal below the screen have the ground colour.
- Grid at 80x24: the hairlines are rows 2 and 21, there are no side lines. The name, the ornament
  and the tagline are row 4. The text starts at column 8. A menu item has the key at column 10
  and the label at column 14. A dot leader in `line` joins the label and the note.
  The note ends at column 78. The key row is row 23, from column 5. Row 24 is empty.
- The hairlines and the dot leaders are stretch tokens. A spacer line gives the empty rows
  above the closing rows, so the closing rows stay at the bottom of each terminal
  ([bbt-format](bbt-format.md), section 6.6). The reader shows the page number
  in the centre of its bottom hairline.
- A key that works only on some pages is in the key row only on these pages.

## 3. Files

| Path | Content |
|---|---|
| `examples/lamp/<lang>/` | The BBS directory in the language `<lang>`: `en` or `ru`. |
| `cmd/velourbbs/testdata/lamp/<lang>/` | The golden cell dumps of section 4. |

- Each BBS directory has an empty `bbs.toml`, the screens `main`, `about`, `term`, `texts`,
  `reader` (a file screen), `log` (a list screen) and `goodbye`, the fragment `_head.bbt`
  with rows 1–5, and the content files `lamp.md`, `rules.txt` and `log/`.
- The two directories have the same files, layout, styles and keys. Only the texts differ.

## 4. Acceptance

- A Go test in `cmd/velourbbs` runs `velourbbs check` in each BBS directory. It gives no output.
- The test compares the `velourbbs preview` cell dump of each screen at 80x24 with 256 colours
  with its golden file `cmd/velourbbs/testdata/lamp/<lang>/<screen>-80x24-256.dump`.
  The test has no other cases.
- Each screen follows the grid of section 2.
