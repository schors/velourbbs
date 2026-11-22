# Home BBS

This specification defines the home BBS of Velour: the station that shows the product,
its guides, its features and how the station is made.
The files of `station/` define its screens and texts.
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
| `station/content/<lang>/inside/` | The texts of the section Inside (section 4). |
| `station/content/<lang>/guide` | A relative symlink to `../../../docs/guide/<lang>`. It keeps one copy of the guides. |
| `cmd/velourbbs/testdata/station/` | The golden cell dumps of section 5. |

`<lang>` is `ru` or `en`.

## 3. Screens

- One BBS directory holds both languages ([session-language](session-language.md)).
  The common entry `main` sets the language with `gosub ru/menu` or `gosub en/menu`.
  The screens of `screens/ru/` and `screens/en/` have the same items.
- A screen of a language stays in its language: no item names a language
  (`lang`, `goto <lang>/`, `gosub <lang>/`), and no file argument has a language directory.
- The texts follow the copy rules of the brandbook, section 7, for their language.
- Each screen except a file screen has one layout for 80 and 132 columns and for 24 and 43 rows.
  `[screen]` has `width = [68, 108]`, `align = "center"` and `gutter = 6`.
  A source line starts with 2 spaces, and a row with a stretch token ends with 2 spaces.
  So the column is 8 to 71 at 80 columns and 14 to 117 at 132, counted from 0,
  as in the grid of the brandbook, section 5.
- Stretch tokens give the rules, the dot leaders and the right edge of the key rows.
  A spacer line above the footer takes the free rows ([bbt-format](bbt-format.md), section 6.6).
- On each screen with a header except a file screen, a block with `term.rows >= 43`
  gives the empty row above the kicker.
  On each screen with keys, a block with `term.rows >= 25` gives the empty row below the keys.
  So, counted from 1, the kicker is at row 5 at 80x24 and 80x25 and at row 6 at 132x43.
  The keys are at row 23 at 80x24 and 80x25 and at row 41 at 132x43.
- The menus have the sections "Документация" / "Documentation" and "Устройство" / "Inside",
  and the screen `about`. The gem of Documentation is `emer`, of Inside `ameth`, of `about` `gold`.
- Each screen of a language has the header of the brandbook, section 2.
  The header of a section screen names both sections: the current section in `white`,
  the other `dim`. The header of the menu and of `about` has both sections `dim`.
  The signature line of the menu ends in `B B S`, of `about` in the gem `gold`.
- The menu has the kicker "Меню" / "Menu" in `gold` above its items.
- The common screens have the logotype: `_logo-132` with `term.cols >= 132 and term.rows >= 43`,
  else `_logo-80`. `main` has one `sweep` with `peak = "g0"` for each logotype fragment
  ([live-screen](live-screen.md), section 6.8). The sweep has no `at`: the light runs over the letters
  and the signature line.
- Each screen with a header has one `sweep` with `peak = "g0"` over the signature line:
  `at = [1, 3, 108, 1]`, `delay = 10`, `delay_spread = 8`, `every = 45` and `spread = 20`.
- Each screen with a header ends its key row with `{clock}` in `dim`, flush with the right edge
  of the column. On a file screen, `Esc` stays at the right: 4 spaces before the clock.
  `main` has no clock: its key row has no space for it at 80 columns.
- A file screen has `width = [68, 68]`, `align = "left"` and `margin = [2, 2]`.
  So its column is 8 to 71 at each width: the brandbook has no reading screen at 132 yet.
- With `term.ambiguous` 2, `main` shows a hint in both languages in place of the greeting:
  the screens break, and the setting for one cell in the common terminals.
  The hint follows the layout rules above. Each row of the hint fits 80 columns
  also when each Cyrillic letter takes two cells.
- The guides and the features text are texts. The station shows each part of a guide
  as one file through the symlink. It does not copy the parts.
- The lists of guide parts use the provider `files`, so `README.md` is the first row.
  The footer hairline of a list shows its page when the list has more than one page.
- The reading screen `read` of Documentation and `inside-read` of Inside show
  the argument `title` in their kicker: the name of the guide, `Velour` for the features text,
  or the name of the text of Inside.
- The screen `about` shows the repository address, the session and BBS fields
  and each `term.*` value that the probe detects. `session.online`, `bbs.sessions`
  and `bbs.uptime` are live fields ([live-screen](live-screen.md), section 3). `term.type` and `term.name` are
  fixed-width fields of 36 cells, so each of their rows fits the column of 64 cells
  and the screen fits 24 rows.
- The station facts (the SysOp name, the address, the port) come from the operator.
  The screens do not invent them.
- The theme gives the colours of the brandbook in each colour depth. The style names are
  the role names of the brandbook. The ground covers the whole terminal: `fill = "spaces"`.

## 4. Features text and Inside texts

- `content/<lang>/features.md` lists what the engine can do now, in short paragraphs.
  It names no feature that is not in a specification of `docs/specs/`.
- A `dev` issue that adds or changes a feature that a user or a sysop sees updates
  this file in both languages.
- `content/<lang>/inside/` tells how the station is made and why: `station.md` (the files,
  the screens and the tests of the station), `development.md` (the development flow)
  and `deploy.md` (the release and the deploy, [release-deploy](release-deploy.md)).
  Each fact comes from the repository. The texts name no host, address or port.
- A change of the station, of `docs/dev-flow/` or of the deploy that makes a fact
  of an Inside text wrong updates the text in both languages.

## 5. Acceptance

- A Go test in `cmd/velourbbs` runs `velourbbs check` in `station/`. It gives no output.
- The test compares the `velourbbs preview` cell dump of each screen at 80x24 and at 132x43
  with 256 colours with its golden file `cmd/velourbbs/testdata/station/<screen>-<size>-256.dump`.
  `<screen>` is a common screen, `<lang>-<name>` with `--set session.lang=<lang>`,
  or `main-ambiguous`: `main` with `--set term.ambiguous=2`.
- The test compares the output of `velourbbs preview main` with `--keys +750ms,+250ms,+500ms,+1500ms`
  at 80x24 and at 132x43 with 256 colours with its golden file
  `cmd/velourbbs/testdata/station/main-sweep-<size>-256.dump`. It shows the sweep over the logotype.
- The test compares the output of `velourbbs preview menu --set session.lang=ru`
  with `--keys +14750ms,+500ms,+1500ms` at 80x24 and at 132x43 with 256 colours with its golden file
  `cmd/velourbbs/testdata/station/ru-menu-sweep-<size>-256.dump`. It shows the sweep over the signature line.
- The test has no other cases.
- The station tests use a guide fixture in place of `content/<lang>/guide`.
  An edit of `docs/guide/` does not change their result.
- Tests check the same items in `screens/ru/` and `screens/en/`, the guide symlinks
  and the title of each reading item.
