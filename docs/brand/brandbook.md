# Velour brandbook

Visual system of the text-mode screens of Velour BBS: gold lettering, bronze hairlines and empty cells, no frames.

- Screens: 80x25 and 132x43; the station draws one row less than the terminal has.
- Colour modes: truecolor, 256 colours, 16 colours (grey), mono.
- Name: VELOUR. The Velour station, the guides and the example BBS follow this book.

Screens in this book are plain text, so they show layout but not colour. Every screen also exists as an ANSI file in four colour modes in `examples/` (see [Examples](#examples)). Person names, letters, posts and texts on the screens are design placeholders.

The screens of this book and the ANSI files show the same text. Change both together.

```text

    █   █ █▀▀▀▀ █     ▄▀▀▀▄ █   █ █▀▀▀▄
    █   █ █▄▄▄  █     █   █ █   █ █▄▄▄▀
    ▀▄ ▄▀ █     █     █   █ █   █ █ ▀▄
      ▀   ▀▀▀▀▀ ▀▀▀▀▀  ▀▀▀   ▀▀▀  ▀   ▀

    ────────────────────────────────────────────────────────── B B S

    01 ■  Сообщения ..................... Конференции и личная почта

    02 ■  Файлы ....................... ФайлЛист и новые поступления

  ∙ 03 ■  Тексты ............ Статьи и заметки для неспешного чтения

    04 ■  Код .................................. Исходники и примеры

    05 ■  Блог ........................ Заметки СисОпа и комментарии

    06 ■  О станции ..................... Немного о станции и СисОпе



    Enter Открыть    Esc Назад    Q Выход

```

## 1. Principles

1. **Air instead of frames.** No boxes, no double lines. Structure comes from empty rows, one hairline and a shared left edge.
2. **Gold marks, grey reads.** Gold is kept for the name, the rule under a title, list markers and hot keys. Running text is warm grey.
3. **One gem per section.** Each section owns one jewel colour, shown only as a small `■`. It is the only colour outside gold and grey.
4. **One column on a fixed edge.** Content hangs from column 4 at 80 and column 14 at 132. A wide screen adds a card beside the list; it never makes the reading line longer.
5. **Keys where the hand expects them.** The bottom row: paging on the left, `B` before `Пробел`; actions flush right. `Esc` always says where it leads.
6. **Four modes, one design.** Hierarchy is built from brightness first and hue second, so a screen still reads in 16 grey tones and in mono.

## 2. Name and logotype

The name is set in Latin capitals from a 5x7 pixel font with one empty pixel between letters, shaded with the gold ramp from light at the top to dark at the bottom.

| Size | Pixel | Rows | Ramp tones, top to bottom |
|---|---|---|---|
| 80 columns | half block `▀ ▄ █` | 4 | `g1 g3 g5 g7` |
| 132 columns | two full cells `██` | 7 | `g0 g1 g2 g4 g5 g6 g8` |

Under the name runs the **signature line**, one empty row below the logotype: gold (`gold`) exactly as wide as the name, then a bronze hairline (`frame`) to the end of the column, closed by `B B S` in gold.

```text

    █   █ █▀▀▀▀ █     ▄▀▀▀▄ █   █ █▀▀▀▄
    █   █ █▄▄▄  █     █   █ █   █ █▄▄▄▀
    ▀▄ ▄▀ █     █     █   █ █   █ █ ▀▄
      ▀   ▀▀▀▀▀ ▀▀▀▀▀  ▀▀▀   ▀▀▀  ▀   ▀

    ────────────────────────────────────────────────────────── B B S
```

On the entry screen at 132, with the greeting above and the gems and the motto below:

```text
              Добро пожаловать

              ██      ██  ██████████  ██            ██████    ██      ██  ████████
              ██      ██  ██          ██          ██      ██  ██      ██  ██      ██
              ██      ██  ██          ██          ██      ██  ██      ██  ██      ██
              ██      ██  ████████    ██          ██      ██  ██      ██  ████████
              ██      ██  ██          ██          ██      ██  ██      ██  ██  ██
                ██  ██    ██          ██          ██      ██  ██      ██  ██    ██
                  ██      ██████████  ██████████    ██████      ██████    ██      ██

              ────────────────────────────────────────────────────────────────────────────────────────────────── B B S

              ■ ■ ■ ■  Вечер, текст и немного золота
```

On section screens the name is letter-spaced in bright gold (`hi`) in the header, the current section in the brightest tone (`white`), the others dim. The signature line under it ends in the section's gem.

```text

    V E L O U R      Сообщения    Файлы    Тексты    Код    Блог
    ────────────────────────────────────────────────────────────── ■
```

| Rule | |
|---|---|
| Letters | Latin capitals of a 5x7 pixel font. Draw a new letter in the style of the VELOUR letters. |
| Length | A name of n letters is 6n-1 pixels wide. With ` B B S`, up to 8 letters fit the big logotype (104 columns at 132) and up to 9 the small one (64 columns at 80). |
| Clear space | One empty row between the logotype and its signature line, one more before anything else. |
| Never | Frames around the name, outlines, colours other than the gold ramp, centring on section screens. |

## 3. Colour

Screens never name a colour. They name a role, and each role has one definition for every mode.

| Mode | Ground | How roles are shown |
|---|---|---|
| truecolor | `#060a20` | The truecolor value of each role |
| 256 colours | index 16, black | The xterm-256 index; the palette has no navy this dark |
| 16 colours | black | DOS index 7 light grey, 8 dark grey, 15 white; gems turn grey |
| mono | terminal default | normal, bold or reverse; italic and underline where supported |

### Ground and text

| Role | Use | Truecolor | 256 | 16 | Mono |
|---|---|---|---|---|---|
| `text` | Running text | `#aeaba4` | 248 | 7 | normal |
| `dim` | Meta, dates, counts, the unselected right column | `#6f7180` | 243 | 8 | normal |
| `ctl` | Menu items, list entries, key labels | `#bcb197` | 144 | 7 | normal |
| `frame` | Hairlines, dot leaders, read entries | `#7a6230` | 94 | 8 | normal |

### Gold

| Role | Use | Truecolor | 256 | 16 | Mono |
|---|---|---|---|---|---|
| `gold` | Rules under titles, kickers, list markers, the signature line | `#b8902f` | 136 | 7 | normal |
| `hi` | Hot keys, the name in the header, first letters, the selection mark | `#f2c14e` | 220 | 15 | bold |
| `head` | Titles and headings in text | `#d4a94a` | 178 | 15 | bold |
| `white` | The current section in the header | `#ffe08a` | 222 | 15 | bold |
| `pick` | The selected entry | `#d4a94a` | 178 | 15 | bold |

### Inline marks in text

| Role | Use | Truecolor | 256 | 16 | Mono |
|---|---|---|---|---|---|
| `strong` | Strong, bold where the terminal can | `#d4a94a` | 178 | 15 | bold |
| `em` | Emphasis and quotes, italic where the terminal can | `#cfccc4` | 251 | 7 | italic |
| `code` | Inline code and code blocks | `#6e947c` | 65 | 7 | normal |
| `link` | Link text, underlined; the address goes to numbered notes at the end | `#c4a258` | 179 | 15 | underline |

### Gems

| Section | Role | Truecolor | 256 | 16 | Mono |
|---|---|---|---|---|---|
| Сообщения | `ruby` | `#9b1b30` | 124 | 7 | normal |
| Файлы | `sapph` | `#2a5caa` | 25 | 7 | normal |
| Тексты | `emer` | `#1f7a5a` | 29 | 7 | normal |
| Код | `ameth` | `#7a4b8c` | 96 | 7 | normal |
| Блог | `topaz` | `#b0602e` | 130 | 7 | normal |
| О станции | `gold` | `#b8902f` | 136 | 7 | normal |

On code screens the gems also colour C syntax: numbers in ruby, strings in emerald, preprocessor lines in amethyst. Nowhere else does a gem colour appear as text.

### Gold ramp for the logotype

| Tone | Truecolor | 256 | 16 |
|---|---|---|---|
| `g0` | `#ffe7a3` | 222 | 15 |
| `g1` | `#fbd77a` | 221 | 15 |
| `g2` | `#f2c14e` | 220 | 15 |
| `g3` | `#e6ad3a` | 214 | 15 |
| `g4` | `#d49a2c` | 178 | 15 |
| `g5` | `#bf8726` | 172 | 7 |
| `g6` | `#a87220` | 136 | 7 |
| `g7` | `#8f5f1b` | 130 | 7 |
| `g8` | `#744c16` | 94 | 7 |

## 4. Type and glyphs

A terminal has one font and one size. Hierarchy comes from role, position and a few glyphs.

```text
    ■  Тексты  ∙  Ремёсла

    Сусальное золото
    ────────────────
    5 минут чтения  ∙  1 из 3

    Золотой лист во много раз тоньше папиросной бумаги. Он не терпит
    сквозняка и суеты: стоит вздохнуть рядом, и лист сворачивается,
    липнет к пальцам, рвётся.

    Инструменты

      ∙ подушка из мягкой кожи, на которой лист режут
      ∙ нож с прямым лезвием
      ∙ широкая плоская кисть
          - из беличьего волоса
```

| Element | How it is set |
|---|---|
| Kicker | `■` in the section gem, two spaces, section and rubric in `gold` |
| Title | `head`, with a `gold` `─` rule exactly as long as the title on the next row |
| Meta | `dim`; parts joined by two spaces, `∙`, two spaces |
| First letter | The first letter of a text in `hi` |
| Headings | Levels 1-2 in `head`, level 3 in `gold`; on their own line, kept on the page with the lines after them |
| Lists | `∙` in `gold` at indent 2; nested `-` in `frame` at indent 6 |
| Quote | A `│` bar in `gold`, the text in `em` |
| Code block | A `│` bar in `frame`, the code in `code` |
| Break | `∙   ∙   ∙` in `gold` |
| Cut text | Ends in a dim `...`; a shortened preview says `... ещё 15 строк` |

| Glyph | Use |
|---|---|
| `─` | Hairlines, the signature line, the rule under a title |
| `■` | Gems: sections, unread entries, kickers |
| `∙` | The selection mark two columns left of an entry, list bullets, separators in meta |
| `.....` | Dot leaders from an entry to its right column; `frame`, `gold` on the selected entry |
| `│ ├─ └─` | Quote and code bars, the code tree |
| `█ ▀ ▄` | Logotype pixels only |
| `"` `-` `...` | Quotes, dash and ellipsis |

## 5. Grid

Rows are the same on every section screen, so the header, the page number and the keys are always in the same place. Rows and columns below are counted from 1.

```text
0         1         2         3         4         5         6         7         
01234567890123456789012345678901234567890123456789012345678901234567890123456789

    V E L O U R      Сообщения    Файлы    Тексты    Код    Блог
    ────────────────────────────────────────────────────────────── ■

    ■  Тексты  ∙  Ремёсла

    Сусальное золото
    ────────────────
    5 минут чтения  ∙  1 из 3

    Золотой лист во много раз тоньше папиросной бумаги. Он не терпит
    сквозняка и суеты: стоит вздохнуть рядом, и лист сворачивается,
    липнет к пальцам, рвётся.

    Инструменты

      ∙ подушка из мягкой кожи, на которой лист режут
      ∙ нож с прямым лезвием
      ∙ широкая плоская кисть
          - из беличьего волоса

    ───────────────────────────   1/6   ────────────────────────────
    B Назад    Пробел Дальше            O Оглавление    Esc К списку

```

| Row at 80x25 | Row at 132x43 | What |
|---|---|---|
| 2 | 2 | Header: the name letter-spaced, then the sections |
| 3 | 3 | Signature line, ending in the section gem |
| 5 - 9 | 6 - 10 | Kicker, title with its rule, meta |
| 11 - 21 | 12 - 39 | Content |
| 22 | 40 | Footer hairline with the page as `2/6` set into its middle; a single page gets the bare line |
| 23 | 41 | Keys |
| 25 | 43 | Left to the terminal |

| Columns | 80 | 132 |
|---|---|---|
| Left edge (counted from 0) | 4 | 14 |
| Reading column | 64 | 70 |
| Catalogs | 64; the lead of the selection in a field above the footer | 104: list 44, gap 8, card 52 |
| Menu and its signature line | 64 | 104 |
| Selection mark | 2 left of the edge | 2 left of the edge |

At 132 the list keeps its width and the card on the right shows the selected entry:

```text
0         1         2         3         4         5         6         7         8         9         0         1         2         3 
012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901

              V E L O U R      Сообщения    Файлы    Тексты    Код    Блог
              ────────────────────────────────────────────────────────────────────────────────────────────────────── ■


              ■  Тексты

              Ремёсла                                             ■  Ремёсла  ∙  1 из 3

            ∙ ■  Сусальное золото .................. 5 мин        Сусальное золото
                                                                  ────────────────
              ■  Чернь по серебру .................. 8 мин        5 минут чтения

              ■  Перегородчатая эмаль ............. 12 мин        Как ложится лист тоньше папиросной бумаги и почему
                                                                  мастер берёт его не руками, а кистью.
              Терминалы

              ■  Почему 80 колонок ................. 6 мин        Enter Читать

              ■  Золото в 256 цветах ............... 7 мин

              ■  Кириллица в CP866 ................. 5 мин

              Заметки

              ■  Медленный экран ................... 4 мин













              ────────────────────────────────────────────────────────────────────────────────────────────────────────
              Enter Читать    Esc В меню

```

## 6. Components

### Catalog entry

```text
    Ноябрь

  ∙ 14  ■  Рамки больше не нужны ............. 3 комментария
           Почему с экрана ушли двойные рамки и как без...

    11  ■  Код как дерево .................... 1 комментарий
           Исходники теперь показаны деревом: проекты,...

     6  ■  Пять тонов золота ................. 2 комментария
           Как градиент из пяти тонов держится в 256...
```

Day, gem (`topaz` when unread, `frame` when read), title, dot leader, count; the lead below, cut at a word. The selected entry turns `pick` and gets a `∙` in `hi` two columns left of the edge.

### Tree

```text
    ■  xfer/ ................................... 3 файла
    ├─ src/ .................................... 2 файла
    │  ├─ crc16.c ............................. 25 строк
  ∙ │  │  ├─ crc16() .......................... строка 6
    │  │  └─ main() .......................... строка 19
    │  └─ xmodem.c ............................... скоро
    └─ README.md .............................. 12 строк
```

Folders fold with Enter; their counts sit at the right edge like any other entry.

### Footer and keys

```text
    ───────────────────────────   4/4   ────────────────────────────
    B Назад    Пробел Дальше               K Написать    Esc К ленте
```

Paging keys on the left, `B` before `Пробел`; actions flush with the column's right edge. Key in `hi`, label in `ctl`, four spaces between pairs. A key appears only where it works: `K Написать` shows on pages with comments.

### Comment and the next entry

```text
    СисОп  ∙  14 ноя, 13:30
       Для ч/б подумаю над отбивками. Спасибо, что заметили.

    ■  Дальше  Код как дерево
       Исходники теперь показаны деревом: проекты, папки, файлы и...
```

The name in `gold`, the date `dim`, the text indented 3. At the end of the last page the next entry shows through after `Дальше`, so reading flows on.

### Contents

```text

    Оглавление

    Начало .................................................. стр. 1
  ∙ Инструменты ............................................. стр. 1
    Порядок работы .......................................... стр. 3
        Немного точности .................................... стр. 5


```

Replaces the page in place. Third-level headings are indented 4; the right column gives the page.

### Card at 132

```text
  ■  14 ноября

  Рамки больше не нужны
  ─────────────────────
  3 минуты чтения  ∙  #дизайн

  Первые наброски экранов были в двойных
  рамках. Рамка обещала порядок, но
  забирала место и внимание: глаз сначала
  видел линии, а уже потом текст.

  Что пришло на смену

    ∙ тонкая линия под заголовком
  ... ещё 15 строк

  Последний комментарий

  СисОп  ∙  14 ноя, 13:30
     Для ч/б подумаю над отбивками.
     Спасибо, что заметили.


  Enter Читать
```

Kicker, title, meta, the opening lines with what was cut, the last comment, one key.

## 7. Copy

All screen text is Russian except addresses and the station name. Short lines, sentence case, no full stop after menu items and keys. The station does not invent facts: ports, hosts, dates and schedules come from the sysop or stay out.

| Rule | Write | Not |
|---|---|---|
| The visitor is "Вы", capitalised | `Письма, адресованные лично Вам.` | `Письма для тебя` |
| Straight quotes, a hyphen, three dots | `"Дальше" - и так далее...` | `«Дальше» — и так далее…` |
| ё where it belongs | `Вперёд`, `Ремёсла` | `Вперед`, `Ремесла` |
| Fixed spellings | `СисОп`, `ФайлЛист` | `Сисоп`, `файл-лист` |
| Key, then a verb or a place | `Enter Читать`, `Esc К ленте` | `[Enter] - чтение`, `ESC: назад` |
| Esc names where it goes | `Esc В меню`, `Esc К списку` | `Esc Выход` |
| Numbers agree with words; zero is a phrase | `1 комментарий`, `3 комментария`, `без комментариев` | `комментариев: 0` |
| Dates in words | `14 ноября`; in lists `14 ноя, 13:30` | `14.11` |
| Meta joined by a dot | `3 минуты чтения  ∙  #дизайн` | `3 мин. \| #дизайн` |
| Headings in sentence case | `Что пришло на смену` | `ЧТО ПРИШЛО НА СМЕНУ` |

## 8. Screens

### Вход, 80x25

```text





        Добро пожаловать

        █   █ █▀▀▀▀ █     ▄▀▀▀▄ █   █ █▀▀▀▄
        █   █ █▄▄▄  █     █   █ █   █ █▄▄▄▀
        ▀▄ ▄▀ █     █     █   █ █   █ █ ▀▄
          ▀   ▀▀▀▀▀ ▀▀▀▀▀  ▀▀▀   ▀▀▀  ▀   ▀

        ────────────────────────────────────────────────────────── B B S

        ■ ■ ■ ■  Вечер, текст и немного золота






                                            Enter Войти    Q Отключиться


```

### Вход, 132x43

```text










              Добро пожаловать

              ██      ██  ██████████  ██            ██████    ██      ██  ████████
              ██      ██  ██          ██          ██      ██  ██      ██  ██      ██
              ██      ██  ██          ██          ██      ██  ██      ██  ██      ██
              ██      ██  ████████    ██          ██      ██  ██      ██  ████████
              ██      ██  ██          ██          ██      ██  ██      ██  ██  ██
                ██  ██    ██          ██          ██      ██  ██      ██  ██    ██
                  ██      ██████████  ██████████    ██████      ██████    ██      ██

              ────────────────────────────────────────────────────────────────────────────────────────────────── B B S

              ■ ■ ■ ■  Вечер, текст и немного золота














                                                                                          Enter Войти    Q Отключиться




```

### Меню, 80x25

```text

    █   █ █▀▀▀▀ █     ▄▀▀▀▄ █   █ █▀▀▀▄
    █   █ █▄▄▄  █     █   █ █   █ █▄▄▄▀
    ▀▄ ▄▀ █     █     █   █ █   █ █ ▀▄
      ▀   ▀▀▀▀▀ ▀▀▀▀▀  ▀▀▀   ▀▀▀  ▀   ▀

    ────────────────────────────────────────────────────────── B B S

    01 ■  Сообщения ..................... Конференции и личная почта

    02 ■  Файлы ....................... ФайлЛист и новые поступления

  ∙ 03 ■  Тексты ............ Статьи и заметки для неспешного чтения

    04 ■  Код .................................. Исходники и примеры

    05 ■  Блог ........................ Заметки СисОпа и комментарии

    06 ■  О станции ..................... Немного о станции и СисОпе



    Enter Открыть    Esc Назад    Q Выход

```

### Меню, 132x43

```text


              ██      ██  ██████████  ██            ██████    ██      ██  ████████
              ██      ██  ██          ██          ██      ██  ██      ██  ██      ██
              ██      ██  ██          ██          ██      ██  ██      ██  ██      ██
              ██      ██  ████████    ██          ██      ██  ██      ██  ████████
              ██      ██  ██          ██          ██      ██  ██      ██  ██  ██
                ██  ██    ██          ██          ██      ██  ██      ██  ██    ██
                  ██      ██████████  ██████████    ██████      ██████    ██      ██

              ────────────────────────────────────────────────────────────────────────────────────────────────── B B S


              01 ■  Сообщения ............................................................. Конференции и личная почта

              02 ■  Файлы ............................................................... ФайлЛист и новые поступления

            ∙ 03 ■  Тексты .................................................... Статьи и заметки для неспешного чтения

              04 ■  Код .......................................................................... Исходники и примеры

              05 ■  Блог ................................................................ Заметки СисОпа и комментарии

              06 ■  О станции ............................................................. Немного о станции и СисОпе
















              Enter Открыть    Esc Назад    Q Выход

```

### Сообщения: папки

```text

    V E L O U R      Сообщения    Файлы    Тексты    Код    Блог
    ────────────────────────────────────────────────────────────── ■

    Почта
  ∙ ■  Личные .............................................. 1 новое
    ■  Отправленные ....................................... 1 письмо

    Конференции
    ■  Общая ............................................... 4 новых
    ■  Терминалы .......................................... 2 письма







       Письма, адресованные лично Вам.


    ────────────────────────────────────────────────────────────────
    Enter Открыть    Esc В меню

```

### Сообщения: письма в папке

```text

    V E L O U R      Сообщения    Файлы    Тексты    Код    Блог
    ────────────────────────────────────────────────────────────── ■

    ■  Конференции  ∙  Общая

  ∙ ■  Палитра для 256 цветов ......................... Анна  13 ноя
    ■  Re: Палитра для 256 цветов .................... Борис  13 ноя
    ■  Re: Палитра для 256 цветов ..................... Анна  14 ноя
    ■  Шрифты для CP866 ............................... Вера  14 ноя
    ■  Как читать длинные тексты ...................... Глеб  14 ноя
    ■  Re: Шрифты для CP866 .......................... Борис  14 ноя






       Анна Котова  ->  Всем  ∙  13 ноя, 18:30
       Собрала палитру станции для 256 цветов. Золото идёт...

    ───────────────────────────   1/6   ────────────────────────────
    Enter Читать    R Написать    Esc К папкам

```

### Сообщения: письмо

```text

    V E L O U R      Сообщения    Файлы    Тексты    Код    Блог
    ────────────────────────────────────────────────────────────── ■

    ■  Сообщения  ∙  Личные

    Файлы со шрифтами
    ─────────────────
    Вера Сомова  ->  Борис Лунев  ∙  14 ноя, 10:05

    Борис, привет!

    Выложила шрифты в ФайлЛист, ищите FONTS866.ZIP. Там три размера,
    все с буквой ё.

    Вера

    ■  Дальше  Спасибо за палитру
       Глеб Орлов: Палитра встала как родная, золото больше не...


    ───────────────────────────   1/2   ────────────────────────────
    B Назад    Пробел Дальше              R Ответить    Esc К списку

```

### Тексты: каталог с карточкой

```text

              V E L O U R      Сообщения    Файлы    Тексты    Код    Блог
              ────────────────────────────────────────────────────────────────────────────────────────────────────── ■


              ■  Тексты

              Ремёсла                                             ■  Ремёсла  ∙  1 из 3

            ∙ ■  Сусальное золото .................. 5 мин        Сусальное золото
                                                                  ────────────────
              ■  Чернь по серебру .................. 8 мин        5 минут чтения

              ■  Перегородчатая эмаль ............. 12 мин        Как ложится лист тоньше папиросной бумаги и почему
                                                                  мастер берёт его не руками, а кистью.
              Терминалы

              ■  Почему 80 колонок ................. 6 мин        Enter Читать

              ■  Золото в 256 цветах ............... 7 мин

              ■  Кириллица в CP866 ................. 5 мин

              Заметки

              ■  Медленный экран ................... 4 мин













              ────────────────────────────────────────────────────────────────────────────────────────────────────────
              Enter Читать    Esc В меню

```

### Тексты: каталог

```text

    V E L O U R      Сообщения    Файлы    Тексты    Код    Блог
    ────────────────────────────────────────────────────────────── ■

    Ремёсла
  ∙ ■  Сусальное золото ...................................... 5 мин
    ■  Чернь по серебру ...................................... 8 мин
    ■  Перегородчатая эмаль ................................. 12 мин

    Терминалы
    ■  Почему 80 колонок ..................................... 6 мин
    ■  Золото в 256 цветах ................................... 7 мин
    ■  Кириллица в CP866 ..................................... 5 мин

    Заметки
    ■  Медленный экран ....................................... 4 мин


       Как ложится лист тоньше папиросной бумаги и почему мастер
       берёт его не руками, а кистью.

    ────────────────────────────────────────────────────────────────
    Enter Читать    Esc В меню

```

### Тексты: чтение

```text

    V E L O U R      Сообщения    Файлы    Тексты    Код    Блог
    ────────────────────────────────────────────────────────────── ■

    ■  Тексты  ∙  Ремёсла

    Сусальное золото
    ────────────────
    5 минут чтения  ∙  1 из 3

    Золотой лист во много раз тоньше папиросной бумаги. Он не терпит
    сквозняка и суеты: стоит вздохнуть рядом, и лист сворачивается,
    липнет к пальцам, рвётся.

    Инструменты

      ∙ подушка из мягкой кожи, на которой лист режут
      ∙ нож с прямым лезвием
      ∙ широкая плоская кисть
          - из беличьего волоса

    ───────────────────────────   1/6   ────────────────────────────
    B Назад    Пробел Дальше            O Оглавление    Esc К списку

```

### Тексты: оглавление

```text

    V E L O U R      Сообщения    Файлы    Тексты    Код    Блог
    ────────────────────────────────────────────────────────────── ■

    ■  Тексты  ∙  Ремёсла

    Сусальное золото
    ────────────────
    5 минут чтения  ∙  1 из 3

    Оглавление

    Начало .................................................. стр. 1
  ∙ Инструменты ............................................. стр. 1
    Порядок работы .......................................... стр. 3
        Немного точности .................................... стр. 5





    ───────────────────────────   1/6   ────────────────────────────
    Enter Перейти    O К тексту    Esc К списку

```

### Код: дерево с карточкой

```text

              V E L O U R      Сообщения    Файлы    Тексты    Код    Блог
              ────────────────────────────────────────────────────────────────────────────────────────────────────── ■


              ■  Код

              ■  xfer/ ................................... 3 файла        ■  xfer / src / crc16.c
              ├─ src/ .................................... 2 файла
              │  ├─ crc16.c ............................. 25 строк        crc16()
            ∙ │  │  ├─ crc16() .......................... строка 6        ───────
              │  │  └─ main() .......................... строка 19        строка 6  ∙  crc16.c
              │  └─ xmodem.c ............................... скоро
              └─ README.md .............................. 12 строк        uint16_t crc16(const uint8_t *buf, size_t...
                                                                          {
              ■  charset/ ................................ 3 файла            uint16_t crc = 0;
              ├─ utf8.c ................................ 32 строки
              │  ├─ utf8_seq() .......................... строка 6            while (len--) {
              │  ├─ utf8_len() ......................... строка 16                crc ^= (uint16_t)*buf++ << 8;
              │  └─ main() ............................. строка 28                for (int i = 0; i < 8; i++)
              ├─ koi2dos.c ................................. скоро                    crc = (crc & 0x8000) ? (crc <...
              └─ README.md ............................... 8 строк                                         : crc << 1;
                                                                              }
              ■  ansi/ ................................... 2 файла            return crc;
              ├─ ansi.c .................................... скоро        }
              └─ README.md ............................... 5 строк













              ────────────────────────────────────────────────────────────────────────────────────────────────────────
              Enter Открыть    Esc В меню

```

### Блог: лента

```text

    V E L O U R      Сообщения    Файлы    Тексты    Код    Блог
    ────────────────────────────────────────────────────────────── ■

    Ноябрь
  ∙ 14  ■  Рамки больше не нужны ..................... 3 комментария
           Почему с экрана ушли двойные рамки и как без них...
    11  ■  Код как дерево ............................ 1 комментарий
           Исходники теперь показаны деревом: проекты, папки,...
     6  ■  Пять тонов золота ......................... 2 комментария
           Как градиент из пяти тонов держится в 256 цветах и не...

    Октябрь
    29  ■  Письмо, которое читают с экрана
           Короткие абзацы, цитаты с инициалами и подпись: что...
    21  ■  Буква ё на своём месте .................... 2 комментария
           Где буква ё стоит в CP866 и KOI8-R и почему её так...
     9  ■  Восемьдесят колонок
           Узкий экран заставляет выбирать главное. Это скорее...


    ───────────────────────────   1/7   ────────────────────────────
    Enter Читать    Esc В меню

```

### Блог: лента с карточкой

```text

              V E L O U R      Сообщения    Файлы    Тексты    Код    Блог
              ────────────────────────────────────────────────────────────────────────────────────────────────────── ■


              ■  Блог

              Ноябрь                                                          ■  14 ноября

            ∙ 14  ■  Рамки больше не нужны ............. 3 комментария        Рамки больше не нужны
                     Почему с экрана ушли двойные рамки и как без...          ─────────────────────
                                                                              3 минуты чтения  ∙  #дизайн
              11  ■  Код как дерево .................... 1 комментарий
                     Исходники теперь показаны деревом: проекты,...           Первые наброски экранов были в двойных
                                                                              рамках. Рамка обещала порядок, но
               6  ■  Пять тонов золота ................. 2 комментария        забирала место и внимание: глаз сначала
                     Как градиент из пяти тонов держится в 256...             видел линии, а уже потом текст.

                                                                              Что пришло на смену
              Октябрь
                                                                                ∙ тонкая линия под заголовком
              29  ■  Письмо, которое читают с экрана                          ... ещё 15 строк
                     Короткие абзацы, цитаты с инициалами и...
                                                                              Последний комментарий
              21  ■  Буква ё на своём месте ............ 2 комментария
                     Где буква ё стоит в CP866 и KOI8-R и почему её...        СисОп  ∙  14 ноя, 13:30
                                                                                 Для ч/б подумаю над отбивками.
               9  ■  Восемьдесят колонок                                         Спасибо, что заметили.
                     Узкий экран заставляет выбирать главное. Это...

                                                                              Enter Читать
              Сентябрь

              30  ■  Первая запись ..................... 3 комментария
                     Зачем станции блог и о чём здесь будет.




              ───────────────────────────────────────────────   1/7   ────────────────────────────────────────────────
              Enter Читать    Esc В меню

```

### Блог: запись, страница с комментариями

```text

    V E L O U R      Сообщения    Файлы    Тексты    Код    Блог
    ────────────────────────────────────────────────────────────── ■

    ■  Блог  ∙  14 ноября

    Рамки больше не нужны
    ─────────────────────
    3 минуты чтения  ∙  3 комментария  ∙  #дизайн

    СисОп  ∙  14 ноя, 13:30
       Для ч/б подумаю над отбивками. Спасибо, что заметили.

    ■  Дальше  Код как дерево
       Исходники теперь показаны деревом: проекты, папки, файлы и...






    ───────────────────────────   4/4   ────────────────────────────
    B Назад    Пробел Дальше               K Написать    Esc К ленте

```

## Examples

Each screen as an ANSI file in four colour modes. To view one, set the terminal to the screen's size and run from the repository root:

```sh
cat docs/brand/examples/menu-80-tc.ans
```

Each file name ends in the mode: `tc.ans` for a truecolor terminal, `256.ans` for a 256-colour one, `16.ans` for a DOS-style 16-colour one, `mono.ans` where there is no colour.

| Screen | File, without the mode |
|---|---|
| Вход, 80x25 | `examples/welcome-80-` |
| Вход, 132x43 | `examples/welcome-132-` |
| Меню, 80x25 | `examples/menu-80-` |
| Меню, 132x43 | `examples/menu-132-` |
| Сообщения: папки | `examples/folders-80-` |
| Сообщения: письма в папке | `examples/letters-80-` |
| Сообщения: письмо | `examples/letter-80-` |
| Тексты: каталог с карточкой | `examples/texts-132-` |
| Тексты: каталог | `examples/texts-80-` |
| Тексты: чтение | `examples/text-80-` |
| Тексты: оглавление | `examples/toc-80-` |
| Код: дерево с карточкой | `examples/code-132-` |
| Блог: лента | `examples/blog-80-` |
| Блог: лента с карточкой | `examples/blog-132-` |
| Блог: запись, страница с комментариями | `examples/post-80-` |
