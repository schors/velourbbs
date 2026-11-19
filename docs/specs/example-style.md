# Example style

This specification defines the style and the screens of the example BBS "Lamp".
The guides build this BBS: [setup-guide](setup-guide.md) and [bbt-guide](bbt-guide.md).
Related parts: [bbt-format](bbt-format.md), [screens-toml](screens-toml.md),
[screen-render](screen-render.md), [telnet-term](telnet-term.md), [file-view](file-view.md)
and [list-view](list-view.md).

## 1. Goal

The sysop reads the setup guide and builds a BBS that looks good at once.
The style is very simple, gentle and warm. The screen sources stay plain text that the sysop
can edit without counting cells.

## 2. Principles

- The example style is not a second brandbook. It does not use the [brandbook](../brand/brandbook.md).
- The style uses only what the engine can do.
- A dark warm ground, cream text and one warm accent for the keys.
- Two hairlines frame the content: one above and one below. There are no side lines,
  so a text line needs no padding. There are no shadows.
- One ornament `☼` on a screen: the light of the lamp.
- Much empty space. A screen has few lines.
- The keys of a screen are always visible in the key row.

## 3. Files

| Path | Content |
|---|---|
| `examples/lamp/en/` | The BBS directory with the English screens. |
| `examples/lamp/ru/` | The BBS directory with the Russian screens. |
| `cmd/velourbbs/testdata/lamp/<lang>/` | The golden cell dumps of section 8. `<lang>` is `en` or `ru`. |

Each BBS directory has these files:

| File | Content |
|---|---|
| `bbs.toml` | Empty. The defaults of [bbs-toml](bbs-toml.md). |
| `screens/screens.toml` | The settings of section 4. |
| `screens/_head.bbt` | The fragment with rows 1–5 of section 5. Each screen includes it. |
| `screens/main.bbt` | The first screen: the menu. |
| `screens/about.bbt` | The text about the BBS. |
| `screens/term.bbt` | The values of the terminal. |
| `screens/texts.bbt` | The list of the texts of the reading room. |
| `screens/reader.bbt` | The file screen: it shows a text of the list. |
| `screens/log.bbt` | The list screen: the notes of the keeper's log. |
| `screens/goodbye.bbt` | The screen after `logoff`. |
| `content/lamp.md` | A Markdown text: a short history of the lamps (section 6.6). |
| `content/rules.txt` | A plain text: the rules of the station (section 6.6). |
| `content/log/` | The notes of the keeper's log: 7 Markdown files (section 6.7). |

- The two directories have the same layout, styles and keys. Only the texts differ (section 7).
- The screens have no art and no aliases.

## 4. `screens/screens.toml`

```toml
[screen]
width = [80, 80]
more = "    more ∙ any key"        # ru: "    ещё ∙ любая клавиша"
fill = "spaces"

[theme.default]
normal = "#e9dccb/#2a201d"
dim    = "#9e8a7c"
line   = "#8a6a5c"
title  = "#f7c7a3 bold"
key    = "#ffb088 bold"
orn    = "#f2a27e"
more   = "#9e8a7c"
md_h1     = "#f7c7a3 bold"
md_h2     = "#f2a27e"
md_h3     = "#d9b99b"
md_strong = "#f7c7a3 bold"
md_em     = "#d9b99b"
md_code   = "#c8b88a"
md_link   = "#d9b99b underline"
md_mark   = "#f2a27e"
md_frame  = "#8a6a5c"
md_dim    = "#9e8a7c"

[theme.256]
normal = "223/235"
dim    = "137"
line   = "95"
title  = "230 bold"
key    = "216 bold"
orn    = "216"
more   = "137"
md_h1     = "230 bold"
md_h2     = "216"
md_h3     = "180"
md_strong = "230 bold"
md_em     = "180"
md_code   = "144"
md_link   = "180 underline"
md_mark   = "216"
md_frame  = "95"
md_dim    = "137"

[theme.16]
normal = "7/0"
dim    = "8"
line   = "8"
title  = "15 bold"
key    = "15 bold"
orn    = "7"
more   = "8"
md_h1     = "15 bold"
md_h2     = "15"
md_h3     = "15"
md_strong = "15 bold"
md_em     = "7"
md_code   = "15"
md_link   = "15 underline"
md_mark   = "7"
md_frame  = "8"
md_dim    = "8"
```

| Style | Use | Truecolor | 256 | 16 |
|---|---|---|---|---|
| `normal` | The ground, the text, the labels of the menu items | `#e9dccb` on `#2a201d` | 223 on 235 | 7 on 0 |
| `dim` | The tagline, the notes of the menu items, the closing line of `main`, the labels of `term`, row 7 of `reader` and of `log`, the row numbers and the reading time of `log`, the words of the key row | `#9e8a7c` | 137 | 8 |
| `line` | The two hairlines | `#8a6a5c` | 95 | 8 |
| `title` | The name of the BBS, the farewell of `goodbye`, the title of `reader` and of `log` | `#f7c7a3` bold | 230 bold | 15 bold |
| `key` | The keys | `#ffb088` bold | 216 bold | 15 bold |
| `orn` | The ornament `☼` | `#f2a27e` | 216 | 7 |
| `more` | The More row | `#9e8a7c` | 137 | 8 |
| `md_h1`, `md_strong` | Heading of level 1, strong emphasis: as `title` | `#f7c7a3` bold | 230 bold | 15 bold |
| `md_h2` | Heading of level 2: as `orn` | `#f2a27e` | 216 | 15 |
| `md_h3` | Heading of level 3 to 6: as `md_em` | `#d9b99b` | 180 | 15 |
| `md_mark` | List marker, quote bar, break: as `orn` | `#f2a27e` | 216 | 7 |
| `md_em` | Emphasis, the text of a quote | `#d9b99b` | 180 | 7 |
| `md_code` | Code | `#c8b88a` | 144 | 15 |
| `md_link` | Link text: as `md_em`, underlined. It differs from `key`: a link is not a key | `#d9b99b` underline | 180 underline | 15 underline |
| `md_frame` | Inner list marker, code bar, table rule: as `line` | `#8a6a5c` | 95 | 8 |
| `md_dim` | Note, cut mark: as `dim` | `#9e8a7c` | 137 | 8 |

- 256 colours is the main mode: PuTTY and Tera Term get it ([telnet-term](telnet-term.md), section 6).
  `[theme.256]` gives hand-picked indexes. The reduction of the truecolor values gives
  the neutral greys 253 and 235 and rose notes, and loses the warmth.
- The xterm palette has no warm colour near `#2a201d`, so the ground in 256 colours is the grey 235.
  The cream text 223 keeps the warmth.
- PuTTY does not show `bold` on the indexes 16–255. So `title` differs from the text by its colour,
  and `md_h1` and `md_h2` have different colours.
- 16 colours are grey: black, dark grey, light grey and white.
  With 16 colours, emphasis looks as the text.
- `max = 80` keeps the screen 80 columns wide. On a wider terminal the host centres it,
  and the margins have the ground colour. At 132x43 the screen is the same 80-column screen
  at the top, and the rows below it have the ground colour.
- `min = 80` keeps the layout on a narrower terminal: the host cuts the rows and does not wrap them
  ([bbt-format](bbt-format.md), section 6.1).

## 5. Grid

Columns and rows are 1-based in the screen area of 80 columns.

| Item | Place | Style |
|---|---|---|
| Margins | Columns 1–2 and 79–80 | `normal` |
| Hairlines | Rows 2 and 21, columns 3–78, `─` | `line` |
| Name | Row 4, from column 8 | `title` |
| Ornament and tagline | Row 4, after the name and 2 spaces: `☼ <tagline>` | `orn`, `dim` |
| Text | Rows 6–19, from column 8 | `normal` |
| Menu item | Key at column 10, label at column 14, note at column 54 | `key`, `normal`, `dim` |
| Key row | Row 23, from column 5: the keys, 2 spaces, one word. Groups of a key row have 3 spaces between them | `key`, `dim` |
| File area | `reader`: from row 9, columns 8–78 (section 6.6) | `md_*`, `normal` |
| More row | The last row of a page, from column 5 | `more` |

- Each screen has rows 1–5: row 1 empty, the hairline, row 3 empty, the name with the ornament
  and the tagline, row 5 empty.
- A screen has 23 rows, `goodbye` has 21. `reader` has 24 rows at 80x24 (section 6.6).
  Empty source lines fill the rows between the content and the hairline of row 21.
  Row 20 is empty. On a screen with a key row, row 22 is empty.
- The rows of the terminal below the screen are spaces in `normal`: `fill = "spaces"`
  ([bbt-format](bbt-format.md), section 6.5).
- A space between two parts in one style keeps the style. Examples: `{!key}A T G{!}`,
  `{!orn}☼ {!dim}a BBS by lamplight{!}`.
- Each other cell is `normal`.

## 6. Screens

### 6.1. `main`

Items: `A` `gosub about`, `R` `gosub texts`, `L` `gosub log`, `T` `gosub term`, `G` `logoff`.

```text

  ────────────────────────────────────────────────────────────────────────────

       Lamp  ☼ a BBS by lamplight

       Good evening. It's quiet and warm here.
       Pull up a chair by the lamp.

         A   About                                   who we are

         R   Reading room                            stories by the lamp

         L   Keeper's log                            notes of the nights

         T   Your terminal                           size and colours

         G   Goodbye                                 the lamp will wait

       The light stays on till morning.

  ────────────────────────────────────────────────────────────────────────────

    A R L T G  choose
```

- Rows 9, 11, 13, 15 and 17 are the menu items. Row 19 is the closing line.

### 6.2. `about`

Items: `ESC` `back`, `ENTER` `back`. The key row is `Esc Enter  back`.

```text
       Lamp is a small station for friends.
       Come in, look around, stay a while.

       The station runs on Velour BBS.
       Its screens are plain text files.
```

- These are rows 6–10.

### 6.3. `term`

Items: `ESC` `back`, `ENTER` `back`. The key row is `Esc Enter  back`.

```text
       Size     80×24
       Colours  256
       Type     xterm

       This is how the station sees you.
```

- These are rows 6–10. The labels are `dim`, the values are `normal`. The values start at column 17.
- The values are fields with their natural width: `{term.cols}×{term.rows}`, `{term.colors}`
  and `{term.type}`.
- The colours row shows `16 million` for 16777216:
  `{#if term.colors == 16777216}16 million{#else}{term.colors}{/if}`.

### 6.4. `texts`

Items: `1` `gosub reader` with `args = { file = "lamp.md" }`, `2` `gosub reader`
with `args = { file = "rules.txt" }`, `ESC` `back`.
The key row is `1 2  choose   Esc  back`.

```text
       The texts of the station.
       Pick one and read it by the lamp.


         1   How the lamps were lit                  a short history

         2   Station rules                           read them once
```

- These are rows 6–12. The menu items follow the grid of `main`.

### 6.5. `goodbye`

No items. Rows 1–5 and the hairline of row 21 as in section 5.
Row 6 is the farewell in `title`. Row 7 is the last line in `normal`.
The hairline is the last row. There is no key row.

### 6.6. `reader`

The file screen: `file = "{arg.file}"`, `margin = [7, 2]`
([file-view](file-view.md), section 3).

Items: `SPACE` `page next` with `when = "not file.last"`, `B` `page prev`
with `when = "file.page > 1"`, `ESC` `back`.

At 80x24:

```text
       How the lamps were lit
       Anna Lind  ∙  1 min  ∙  page 1/3

       <the file area, rows 9-19>

  ────────────────────────────────────────────────────────────────────────────

    Space  next   Esc  list

```

- Rows 1–5 as in section 5. Row 6 is `{file.title}` in `title`. Row 7 is in `dim`:
  the author, the reading time and the page. A file without `meta.author` has no author part:
  `{#if meta.author != ""}{meta.author}  ∙  {/if}`.
- Row 8 is empty. The file area follows: 11 rows at 80x24.
- The rows below the file area: an empty row, the hairline, an empty row, the key row
  and an empty row. So the hairline is row 21 and the key row is row 23 at 80x24.
- The key row is `Space  next   B  prev   Esc  list`. It shows the group of `SPACE`
  only when `item.SPACE` is true, and the group of `B` only when `item.B` is true.
  So page 1 has no `B`, and the last page has no `Space`.
- The empty last line of the source gives the empty row 24. Without it, the file area
  has 12 rows and the hairline moves to row 22.
- When `term.rows` is less than 14, the host cuts the page ([file-view](file-view.md), section 3.2).
  When it is less than 13, the key row is not visible.

`content/lamp.md`:

- TOML front matter with `title` and `author`. `file.title` comes from `title`.
- 2 or 3 pages at 80x24. Its text has no level-1 heading.
- The English and the Russian file have the same blocks and the same page count at 80x24.
- Paragraphs, 2 or 3 level-2 headings, a bullet list, a block quote, strong emphasis,
  emphasis, a table with 3 columns, a thematic break and one link with a web address: it gets a note.

`content/rules.txt`:

- 5 to 7 short rules, one line each, and an empty line before the last line. One page at 80x24.
- `file.title` is the file name `rules.txt`.

### 6.7. `log`

The list screen: `data = { from = "files", args = { dir = "log" } }`, `rows = 5`
([list-view](list-view.md), section 3).

Items: `SPACE` `page next` with `when = "not list.last"`, `B` `page prev`
with `when = "list.page > 1"`, `ESC` `back`.

At 80x24, page 1:

```text
       The keeper's log
       notes of the nights  ∙  page 1/2

         1   The first night                         1 min
         2   Rain on the roof                        1 min
         3   A caller from the north                 1 min
         4   A new wick                              1 min
         5   A long winter evening                   2 min
```

- Rows 1–5 as in section 5. Row 6 is the title in `title`. Row 7 is in `dim`.
- Row 8 is empty. The list line gives rows 9–13: `{row.num}` in `dim` at column 10,
  `{row.title}` at column 14 in a field of 40 cells, `{row.minutes}` and the word in `dim` at column 54.
  The row number is in `dim`, not in `key`: it is not a key.
- Empty source lines fill rows 14–20. The hairline is row 21, the key row is row 23.
- The key row is `Space  next   B  prev   Esc  back`, with the groups of `SPACE` and `B`
  as in `reader` (section 6.6).

`content/log/`:

- 7 notes `1-<slug>.md` to `7-<slug>.md`. Each note has a level-1 heading: its title.
  So the list has 2 pages: 5 rows and 2 rows.
- The notes are a few short paragraphs. Note 5 is longer: its reading time is 2 minutes.
  The other notes have 1 minute.
- The English and the Russian notes have the same names and the same reading times.

## 7. Texts

| Place | English | Russian |
|---|---|---|
| Name | Lamp | Лампа |
| Tagline | a BBS by lamplight | BBS при свете лампы |
| `main`, rows 6–7 | Good evening. It's quiet and warm here. / Pull up a chair by the lamp. | Добрый вечер. Здесь тихо и тепло. / Садись поближе к свету. |
| `main`, items | About — who we are; Your terminal — size and colours; Goodbye — the lamp will wait | О станции — кто мы; Твой терминал — размер и цвета; До свидания — лампа подождёт |
| `main`, row 19 | The light stays on till morning. | Свет горит до утра. |
| `main`, key row | choose | выбор |
| `about` | Lamp is a small station for friends. / Come in, look around, stay a while. / (empty) / The station runs on Velour BBS. / Its screens are plain text files. | «Лампа» — маленькая станция для друзей. / Заходи, осмотрись, посиди немного. / (пусто) / Станция работает на Velour BBS. / Её экраны — простые текстовые файлы. |
| `term`, labels | Size, Colours, Type | Размер, Цветов, Тип |
| `term`, 16777216 | 16 million | 16 млн |
| `term`, row 10 | This is how the station sees you. | Так тебя видит станция. |
| Key row of `about` and `term` | back | назад |
| `goodbye` | See you soon. / The lamp stays on for you. | До встречи. / Лампа будет гореть для тебя. |
| `main`, item `R` | Reading room — stories by the lamp | Читальня — истории у лампы |
| `main`, item `L` | Keeper's log — notes of the nights | Журнал смотрителя — заметки о ночах |
| `log`, rows 6–7 | The keeper's log / notes of the nights ∙ page | Журнал смотрителя / заметки о ночах ∙ стр. |
| `log`, list rows | `min` | `мин` |
| `log`, key row | next, prev, back | дальше, пред., назад |
| `content/log/` | The first night, Rain on the roof, A caller from the north, A new wick, A long winter evening, Moths at the glass, Morning tea | Первая ночь, Дождь по крыше, Гость с севера, Новый фитиль, Долгий зимний вечер, Мотыльки у стекла, Утренний чай |
| `texts`, rows 6–7 | The texts of the station. / Pick one and read it by the lamp. | Тексты станции. / Выбери и читай при лампе. |
| `texts`, items | How the lamps were lit — a short history; Station rules — read them once | Как зажигали лампы — короткая история; Правила станции — прочти один раз |
| `texts`, key row | choose, back | выбор, назад |
| `reader`, row 7 | `min`, `page` | `мин`, `стр.` |
| `reader`, key row | next, prev, list | дальше, пред., список |
| `content/lamp.md` | Title "How the lamps were lit", author "Anna Lind" | Заголовок «Как зажигали лампы», автор «Анна Линд» |
| More row | more ∙ any key | ещё ∙ любая клавиша |

- In the Russian screens the values of `term` start at column 17 too.

## 8. Acceptance

- `velourbbs check` gives no errors and no warnings in each BBS directory.
- A Go test in `cmd/velourbbs` runs `velourbbs check` in each BBS directory.
  It runs `velourbbs preview <screen> --size <size> --set term.colors=<n>` and compares
  the cell dump with the golden file `cmd/velourbbs/testdata/lamp/<lang>/<screen>-<size>-<n>.dump`:
  - each screen at 80x24 with 16, 256 and 16777216 colours.
    `reader` has `--set arg.file=lamp.md`. The dump of `log` has its 2 pages;
  - `reader` with `--set arg.file=rules.txt` at 80x24 with 256 colours,
    in the golden file `reader-rules-80x24-256.dump`;
  - `main` at 80x25 and at 132x43 with 256 colours.
- Each screen matches the grid of section 5 at 80x24.

## 9. Not in this specification

- ANSI pictures of the example screens for the guides. `velourbbs preview` gives a cell dump only.
- The texts of the guides.
