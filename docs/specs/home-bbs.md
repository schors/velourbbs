# Home BBS

This specification defines the home BBS of Velour: the station that shows the product,
its guides and its features. The files of `station/` define its screens and texts.
Related parts: [brandbook](../brand/brandbook.md), [session-language](session-language.md)
and [example-style](example-style.md).

## 1. Goal

The home BBS is a live station and the site of the product. It follows the brandbook.
Each screen uses only what the engine can do now.
Lamp stays the example BBS of the guides ([example-style](example-style.md)).

## 2. Files

| Path | Content |
|---|---|
| `station/` | The BBS directory of the home BBS ([bbs-toml](bbs-toml.md), section 2). |
| `station/bbs.toml` | `[bbs] name = "Velour"` and `[store] path = "data/velourbbs.jsonl"`. No other keys. |
| `station/data/` | The directory of the data file. Not in git. The deploy makes it a symlink ([release-deploy](release-deploy.md), section 4.5). |
| `station/screens/` | The common screens, `screens.toml` and the screens of each language in `<lang>/`. |
| `station/content/<lang>/features.md` | The features text (section 4). |
| `station/content/<lang>/guide` | A relative symlink to `../../../docs/guide/<lang>`. It keeps one copy of the guides. |
| `cmd/velourbbs/testdata/station/` | The golden cell dumps of section 5. |

`<lang>` is `ru` or `en`.

## 3. Screens

- One BBS directory holds both languages ([session-language](session-language.md)).
  The common entry `main` sets the language with `gosub ru/menu` or `gosub en/menu`.
  The screens of `screens/ru/` and `screens/en/` have the same items.
- The texts follow the copy rules of the brandbook, section 7, for their language.
- Each screen except a file screen has an 80 form and a 132 form. The 132 form is
  for `term.cols >= 132 and term.rows >= 43`. A file screen has only the 80 form.
- With `term.ambiguous` 2, `main` shows a hint in both languages in place of the greeting:
  the screens break, and the setting for one cell in the common terminals.
  Each row of the hint fits 80 columns when each Cyrillic letter takes two cells.
- The guides and the features text are texts. The station shows each part of a guide
  as one file through the symlink. It does not copy the parts.
- The lists of guide parts use the provider `files`, so `README.md` is the first row.
  The footer hairline of a list shows its page when the list has more than one page.
- The reading screen shows the argument `title` in its kicker: the name of the guide,
  or `Velour` for the features text.
- The screen `about` shows the repository address, the session and BBS fields
  and each `term.*` value that the probe detects. `term.type` and `term.name` are
  fixed-width fields of 40 cells, so the 80 form keeps its 23 rows.
- The station facts (the SysOp name, the address, the port) come from the operator.
  The screens do not invent them.
- The theme gives the colours of the brandbook in each colour depth. The style names are
  the role names of the brandbook. The ground covers the whole terminal: `fill = "spaces"`.

## 4. Features text

- `content/<lang>/features.md` lists what the engine can do now, in short paragraphs.
  It names no feature that is not in a specification of `docs/specs/`.
- A `dev` issue that adds or changes a feature that a user or a sysop sees updates
  this file in both languages.

## 5. Acceptance

- A Go test in `cmd/velourbbs` runs `velourbbs check` in `station/`. It gives no output.
- The test compares `velourbbs preview` cell dumps with golden files
  `cmd/velourbbs/testdata/station/<screen>-<size>-<colors>.dump`. `<screen>` is a common screen,
  or `<lang>-<name>` with `--set session.lang=<lang>`. Each screen has a dump at 80x24
  with 256 colours, and at 132x43 when it has a 132 form. The test lists the other cases:
  one for each rule of section 3 that a dump shows.
- A golden output of `velourbbs preview main --keys R,ESC,E` shows the entry,
  the Russian menu, the entry and the English menu.
- Tests check the same items in `screens/ru/` and `screens/en/`, the guide symlinks
  and the title of each reading item.
