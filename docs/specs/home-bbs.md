# Home BBS

This specification defines the home BBS of Velour: the station that shows the product,
its guides and its features. It also defines the changes of the brandbook for it.
Related parts: [brandbook](../brand/brandbook.md), [bbt-guide](bbt-guide.md),
[setup-guide](setup-guide.md), [file-view](file-view.md), [session-fields](session-fields.md),
[session-language](session-language.md) and [example-style](example-style.md).

## 1. Goal

The home BBS is a live station and the site of the product. It follows the brandbook.
Each screen uses only what the engine can do now.
Lamp stays the example BBS of the guides ([example-style](example-style.md)).

## 2. Files

| Path | Content |
|---|---|
| `station/` | The BBS directory of the home BBS ([bbs-toml](bbs-toml.md), section 2). |
| `station/bbs.toml` | `[bbs] name = "Velour"` and `[store] path = "data/velourbbs.jsonl"`. No `[telnet]` and no `[limits]` table. |
| `station/data/` | The directory of the data file ([data-store](data-store.md)). Not in git (`.gitignore`). The deploy makes it a symlink ([release-deploy](release-deploy.md), section 4.5). |
| `station/screens/screens.toml` | The theme of section 4, the screen settings and the languages `[lang.ru]` and `[lang.en]`. |
| `station/screens/` | The common screens of section 3 and their fragments. |
| `station/screens/<lang>/` | The screens of section 3 in the language `<lang>` ([session-language](session-language.md), section 3). |
| `station/content/<lang>/features.md` | The features text (section 6). |
| `station/content/<lang>/guide` | A relative symlink to `../../../docs/guide/<lang>` ([file-view](file-view.md), section 2). |
| `cmd/velourbbs/testdata/station/` | The golden cell dumps of section 8. |

- `<lang>` is `ru` or `en`.
- The symlink keeps one copy of the guides.
- A version tag deploys the directory with the symlinks resolved ([release-deploy](release-deploy.md)).
  To deploy by hand, copy it: `cp -rL station/. /srv/velourbbs/`, and replace
  `/srv/velourbbs/data/` with an empty directory. Then follow [part 2 of the setup guide](../guide/en/setup/2-bbs-directory.md)
  without its copy of Lamp, and parts 4 and 5. `station/bbs.toml` keeps the data file
  in `/srv/velourbbs/data/`.

## 3. Screens

One BBS directory holds both languages. `screens.toml` has no `[screen] lang`:
a session starts with no language and sees the common screens
([session-language](session-language.md), section 5). The entry `main` sets the language.
Then the session sees the screens of `screens/<lang>/`. The names below are screen names:
`main` of the language is `screens/<lang>/main.bbt`.
The station has two sections of the brandbook: Texts (Тексты) and About the station
(О станции). The guides and the features are texts.

| Screen | Content | Items |
|---|---|---|
| `main` (common) | The entry screen (section 3.2). | `R` `lang ru`, `E` `lang en`, `Q` `logoff`. |
| `main` (language) | The menu (section 3.2). | `1` `gosub texts`, `2` `gosub about`, `L` `lang en` (Russian) or `lang ru` (English), `Q` `logoff`. |
| `texts` | The catalog of the section Texts (section 3.3). | `1` `gosub setup`, `2` `gosub bbt`, `3` `gosub read` with the file `features.md` and the title `Velour`, `ESC` `back`. |
| `setup` | The parts of the setup guide. | `0` for `README.md`, `1`–`5` for the parts: `gosub read` with the part and the title of the guide. `ESC` `back`. |
| `bbt` | The parts of the BBT guide. | `0` for `README.md`, `1`–`9` for the parts: `gosub read` with the part and the title of the guide. `ESC` `back`. |
| `read` | The file screen of the argument `file`, with the argument `title` in the header (section 3.4). | `SPACE` `page next`, `B` `page prev`, `ESC` `back`. |
| `about` | The section About the station (section 3.5). | `ESC` `back`. |
| `goodbye` | The goodbye screen in both languages, with the repository address (section 3.5). | — |
| `shutdown` | The restart notice in both languages: the station restarts, call again in a minute. | — |

- The files of `args` have no language directory: `features.md`, `guide/setup/README.md`.
  The search finds them in `content/<lang>/` ([session-language](session-language.md), section 4).
- The home BBS shows each part of a guide as one file. It does not copy the parts.
- The texts follow the copy rules of the brandbook, section 7, for their language.
- The key rows:

| Screen | Russian | English |
|---|---|---|
| `main` (common) | `R Русский    E English    Q Отключиться / Disconnect` | the same |
| `main` (language) | `1 2 Открыть    L English    Q Выход` | `1 2 Open    L Русский    Q Quit` |
| `texts` | `1 2 3 Открыть    Esc В меню` | `1 2 3 Open    Esc To menu` |
| `setup`, `bbt` | `0-9 Открыть    Esc К текстам` | `0-9 Open    Esc To texts` |
| `read` | `B Назад    Пробел Дальше    Esc К списку` | `B Back    Space Next    Esc To list` |
| `about` | `Esc В меню` | `Esc To menu` |

- `setup` shows `0-5` in place of `0-9`.
- The More row of each language is in `[lang.<lang>] more`: `Любая клавиша Дальше`
  and `Any key Next`, with the indent of the screens. The common screens keep the bilingual
  `[screen] more`: `Любая клавиша Дальше    Any key Next`.
- A file screen shows `B` and `Пробел` (`Space`) only where they work
  (`when` and `{#if item.<KEY>}`). Paging keys are on the left, the other keys flush right.

### 3.1. Width

- `screens.toml` sets `width = [80, 80]` and `align = "left"`.
- Each screen except a file screen has `width = [80, 132]` and two forms:
  the 132 form when `term.cols >= 132 and term.rows >= 43`, else the 80 form.
  The 132 form follows the grid of the brandbook, section 5, at 132.
- A file screen has only the 80 form: its `margin` is one for each width
  ([file-view](file-view.md), section 3).

### 3.2. Entry and menu

- The common `main` is the brandbook screen "Вход". The greeting and the motto are in both languages.
  The language keys replace `Enter Войти`.
- With `term.ambiguous` 2 ([telnet-term](telnet-term.md), section 6), the common `main` shows a hint
  in place of the greeting, the logo and the motto, in both forms. The hint tells in both languages
  that the screens break, and gives the setting for one cell in PuTTY, Tera Term, xterm, iTerm2,
  Windows Terminal, GNOME Terminal and Ptyxis. The key row stays.
  Each row fits 80 columns when each Cyrillic letter takes two cells.
  The last row of the hint is empty: the key row with its padding is wider than `term.cols`
  and wraps into it, so the page does not scroll with 23 rows.
- The `main` of a language is the brandbook screen "Меню" with two items:
  Texts (gem `emer`) and About the station (gem `gold`).
  A hotkey is the digit of the item. The menu has no selection mark.
- The menu is the first screen of the language, so it has no `Esc Ко входу` of the brandbook.
  `L` changes the language in its place.
- The key row of the menu ends in the clock of the brandbook, section 6:
  `bbs.time` with `time = "15:04 MST"` and `zone = "UTC"` in `screens.toml`.

### 3.3. Section screens

- The header and the signature line of the brandbook, section 5.
  The header lists the section Texts. In the section Texts it is in `white`.
  The signature line ends in the gem of the section: `emer` for Texts, `gold` for About the station.
- The catalog `texts` has two groups: Guides (the setup guide and the BBT guide)
  and Velour (the features text). Each entry has the gem `emer`, the title, a dot leader
  and a short note in `dim`. It has no lead.
- A list of guide parts shows the number and the title of each part.
- Each entry of the catalog and of a list of guide parts starts with the digit of its hotkey
  in `gold`, in place of the selection.

### 3.4. Reading

- The brandbook screen "Тексты: чтение" at 80: the kicker, the meta, the file area,
  the footer with `{file.page}/{file.pages}` and the keys.
- The kicker is `Тексты  ∙  {arg.title}` (English: `Texts  ∙  {arg.title}`). The item gives the title:
  the name of the guide, or `Velour` for the features text.
- The meta is `{file.minutes} мин чтения` (English: `{file.minutes} min read`).
- The title is the level-1 heading at the start of the file. The rows above the file area
  do not repeat it.
- `margin` gives the left edge 4 and the reading column 64.

### 3.5. About the station

- A short text about the project: what Velour is and that the station runs on it.
- The repository address `https://github.com/schors/velourbbs` as a label and value row
  (`Исходный код`, `Source code`), above the session and BBS fields.
- The session and BBS fields as label and value rows: `session.ip`, `session.online`,
  `bbs.version`, `bbs.sessions` and `bbs.uptime` ([session-fields](session-fields.md), section 2).
  A label does not count the value, so it needs no plural form.
- Below them, after an empty row, the terminal rows show the `term.*` fields
  ([telnet-term](telnet-term.md), section 6):

  | Label | Value |
  |---|---|
  | `Терминал`, `Terminal` | `{term.cols}x{term.rows}, {term.type}` |
  | `Программа`, `Program` | `term.name`. When it is empty, `нет ответа`, `no reply`. |
  | `Цвета`, `Colours` | `term.colors`, `truecolor` for 16777216. `, sixel` follows when `term.images` is `sixel`. |
  | `Фон`, `Background` | `тёмный`, `dark` or `светлый`, `light` and `, {term.bg_color}`. When `term.background` is empty, `не определён`, `unknown`. |
  | `Ширина кириллицы`, `Cyrillic width` | `1 клетка`, `1 cell` or `2 клетки`, `2 cells`: `term.ambiguous`. |

- `term.type` and `term.name` are fixed-width fields of 40 cells, so a long value
  does not wrap.
- The 80 form has 23 rows: its text is one paragraph of two rows, and the repository row
  has no empty row below it.
- The station facts (the SysOp name, the address, the port) come from the operator.
  The screen does not invent them. Until the operator gives them, the screen has no facts.

## 4. Theme

- `[theme.default]` gives the truecolor values of the brandbook, section 3.
  `[theme.256]` and `[theme.16]` give its 256 and 16 values.
- The style names are the role names of the brandbook: `text` as `normal`, `dim`, `ctl`,
  `frame`, `gold`, `hi`, `head`, `white`, the gems and the ramp `g0`–`g8`.
- The `md_*` styles follow the brandbook, sections 3 and 4: `md_h1` and `md_h2` as `head`,
  `md_h3` as `gold`, `md_strong` as `strong`, `md_em` as `em`, `md_code` as `code`,
  `md_link` as `link`, `md_mark` as `gold`, `md_frame` as `frame`, `md_dim` as `dim`.
- The ground of the brandbook covers the whole terminal: `normal` and `more` have the ground
  as background in each theme, and `fill = "spaces"`.

## 5. Guide parts

The guides are directories of parts ([bbt-guide](bbt-guide.md), section 3.1).
`setup` and `bbt` list the parts in the order of their numbers.

## 6. Features text

- `content/<lang>/features.md` lists what the engine can do now, in short paragraphs.
  It names no feature that is not in a specification of `docs/specs/`.
- A `dev` issue that adds or changes a feature that a user or a sysop sees updates
  this file in both languages.

## 7. Brandbook changes

| Place | Change |
|---|---|
| Header list, line 7 | The home station follows this book. The example BBS Lamp does not. The copy rules apply to screens, not to the guides. |
| Principle 6 | "16 grey tones" becomes the greys of the 16-colour mode. |
| Section 7 | The rules apply to screen text in Russian or in English. Add the English rules (section 7.1). |
| The opening menu (line 36), the menu screens and their `.ans` files | `Esc Назад` becomes `Esc Ко входу`. |
| New section "Engine support" | One table: each brandbook element that the engine cannot show yet, and the specification or idea that would add it, or "none yet". |

The table of "Engine support" lists: Messages, Files, Code, Blog, the arrow-key selection
and the selection mark, the lead of the selection, the card at 132, the reading screen at 132,
the table of contents, the syntax colours, the mono mode, italic,
the title with its rule above the meta, the first letter in `hi`,
and the plural forms of the numbers.

### 7.1. English rules

The English table of section 7 of the brandbook:

| Rule | Write | Not |
|---|---|---|
| The visitor is "you" | `Letters sent to you.` | `Letters for the user` |
| Straight quotes, a hyphen, three dots | `"Next" - and so on...` | `“Next” — and so on…` |
| Key, then a verb or a place | `Enter Read`, `Esc To list` | `[Enter] - read`, `ESC: back` |
| Esc names where it goes | `Esc To menu`, `Esc To list` | `Esc Exit` |
| Numbers agree with words; zero is a phrase | `1 comment`, `3 comments`, `no comments` | `comments: 0` |
| Dates in words | `14 November`; in lists `14 Nov, 13:30` | `11/14` |
| Meta joined by a dot | `3 min read  ∙  #design` | `3 min. \| #design` |
| Headings in sentence case | `What came next` | `What Came Next` |

## 8. Acceptance

- A Go test in `cmd/velourbbs` runs `velourbbs check` in `station/`. It gives no output.
- The test compares `velourbbs preview` cell dumps with golden files
  `cmd/velourbbs/testdata/station/<screen>-<size>-<colors>.dump`.
  `<screen>` is the name of a common screen, or `<lang>-<name>` for a screen of a language
  with `--set session.lang=<lang>`. Example: `ru-main` is the Russian menu.
  - each screen at 80x24 with 256 colours;
  - each screen except a file screen also at 132x43 with 256 colours;
  - `main` and `ru-texts` also with 16 and 16777216 colours;
  - `ru-about` and `en-about` also with 16777216 colours, `term.ambiguous` 2,
    a `term.name`, sixel and a `term.bg_color` (a dark one and a light one);
  - `ru-about` also at 80x24 with a `term.type` and a `term.name` of 40 wide graphemes;
  - each file screen with the first file of each catalog.
- A golden output of `velourbbs preview main --keys R,L` shows the entry,
  the Russian menu and the English menu.
- The screens of `screens/ru/` and of `screens/en/` have the same names and the same items.
- `content/ru/guide` and `content/en/guide` are symlinks into `docs/guide/`.

## 9. Not in this specification

- Version news. It comes after the first release tag.
- The station facts: the operator gives them later.
- The elements of the "Engine support" table of the brandbook.
