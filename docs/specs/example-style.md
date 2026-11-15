# Example style

This specification defines the style and the screens of the example BBS "Lamp".
The guides build this BBS: [setup-guide](../ideas/setup-guide.md) and [bbt-guide](../ideas/bbt-guide.md).
Related parts: [bbt-format](bbt-format.md), [screens-toml](screens-toml.md),
[screen-render](screen-render.md) and [telnet-term](telnet-term.md).

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
| `screens/main.bbt` | The first screen: the menu. |
| `screens/about.bbt` | The text about the BBS. |
| `screens/term.bbt` | The values of the terminal. |
| `screens/goodbye.bbt` | The screen after `logoff`. |

- The two directories have the same layout, styles and keys. Only the texts differ (section 7).
- The screens have no art, no fragments and no aliases.

## 4. `screens/screens.toml`

```toml
[screen]
width = [80, 80]
more = "    more ∙ any key"        # ru: "    ещё ∙ любая клавиша"

[theme.default]
normal = "#e9dccb/#2a201d"
dim    = "#9e8a7c"
line   = "#8a6a5c"
title  = "#f7c7a3 bold"
key    = "#ffb088 bold"
orn    = "#f2a27e"
more   = "#9e8a7c"

[theme.256]
normal = "223/235"
dim    = "137"
line   = "95"
title  = "230 bold"
key    = "216 bold"
orn    = "216"
more   = "137"

[theme.16]
normal = "7/0"
dim    = "8"
line   = "8"
title  = "15 bold"
key    = "15 bold"
orn    = "7"
more   = "8"
```

| Style | Use | Truecolor | 256 | 16 |
|---|---|---|---|---|
| `normal` | The ground, the text, the labels of the menu items | `#e9dccb` on `#2a201d` | 223 on 235 | 7 on 0 |
| `dim` | The tagline, the notes of the menu items, the closing line of `main`, the labels of `term`, the words of the key row | `#9e8a7c` | 137 | 8 |
| `line` | The two hairlines | `#8a6a5c` | 95 | 8 |
| `title` | The name of the BBS, the farewell of `goodbye` | `#f7c7a3` bold | 230 bold | 15 bold |
| `key` | The keys | `#ffb088` bold | 216 bold | 15 bold |
| `orn` | The ornament `☼` | `#f2a27e` | 216 | 7 |
| `more` | The More row | `#9e8a7c` | 137 | 8 |

- 256 colours is the main mode: PuTTY and Tera Term get it ([telnet-term](telnet-term.md), section 6).
  `[theme.256]` gives hand-picked indexes. The reduction of the truecolor values gives
  the neutral greys 253 and 235 and rose notes, and loses the warmth.
- The xterm palette has no warm colour near `#2a201d`, so the ground in 256 colours is the grey 235.
  The cream text 223 keeps the warmth.
- PuTTY does not show `bold` on the indexes 16–255. So `title` differs from the text by its colour.
- 16 colours are grey: black, dark grey, light grey and white.
- `max = 80` keeps the screen 80 columns wide. On a wider terminal the host centres it,
  and the margins have the ground colour. At 132x43 the screen is the same 80-column screen.
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
| Key row | Row 23, from column 5: the keys, 2 spaces, one word | `key`, `dim` |
| More row | The last row of a page, from column 5 | `more` |

- Each screen has rows 1–5: row 1 empty, the hairline, row 3 empty, the name with the ornament
  and the tagline, row 5 empty.
- A screen has 23 rows. Empty source lines fill the rows between the content and the hairline of row 21.
  Rows 20 and 22 are empty.
- On a 24-row terminal, row 24 keeps the colours of the terminal. On a taller terminal,
  each row below row 23 keeps them: [screen-fill](../ideas/screen-fill.md).
- A space between two parts in one style keeps the style. Examples: `{!key}A T G{!}`,
  `{!orn}☼ {!dim}a BBS by lamplight{!}`.
- Each other cell is `normal`.

## 6. Screens

### 6.1. `main`

Items: `A` `goto about`, `T` `goto term`, `G` `logoff`.

```text

  ────────────────────────────────────────────────────────────────────────────

       Lamp  ☼ a BBS by lamplight

       Good evening. It's quiet and warm here.
       Pull up a chair by the lamp.


         A   About                                   who we are

         T   Your terminal                           size and colours

         G   Goodbye                                 the lamp will wait




       The light stays on till morning.

  ────────────────────────────────────────────────────────────────────────────

    A T G  choose
```

- Rows 10, 12 and 14 are the menu items. Row 19 is the closing line.

### 6.2. `about`

Items: `ESC` `goto main`, `ENTER` `goto main`. The key row is `Esc Enter  back`.

```text
       Lamp is a small station for friends.
       Come in, look around, stay a while.

       The station runs on Velour BBS.
       Its screens are plain text files.
```

- These are rows 6–10.

### 6.3. `term`

Items: `ESC` `goto main`, `ENTER` `goto main`. The key row is `Esc Enter  back`.

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

### 6.4. `goodbye`

No items. Rows 1–5 and the hairline of row 21 as in section 5.
Row 6 is the farewell in `title`. Row 7 is the last line in `normal`.
Rows 22 and 23 are empty. There is no key row.

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
| More row | more ∙ any key | ещё ∙ любая клавиша |

- In the Russian screens the values of `term` start at column 17 too.

## 8. Acceptance

- `velourbbs check` gives no errors and no warnings in each BBS directory.
- A Go test in `cmd/velourbbs` runs `velourbbs check` in each BBS directory.
  It runs `velourbbs preview <screen> --size <size> --set term.colors=<n>` and compares
  the cell dump with the golden file `cmd/velourbbs/testdata/lamp/<lang>/<screen>-<size>-<n>.dump`:
  - each screen at 80x24 with 16, 256 and 16777216 colours;
  - `main` at 80x25 and at 132x43 with 256 colours.
- Each screen matches the grid of section 5 at 80x24.

## 9. Not in this specification

- The rows below a screen on a terminal with more than 23 rows, and vertical centring:
  [screen-fill](../ideas/screen-fill.md).
- ANSI pictures of the example screens for the guides. `velourbbs preview` gives a cell dump only.
- The texts of the guides.
