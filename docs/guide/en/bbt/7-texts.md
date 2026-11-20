# Task 7. Texts

In this task, you add a reading room to Lamp. The screen `texts` lists two texts.
The screen `reader` shows a text in pages.

## Content files

A "content file" is a file that a screen shows.
Put the content files in the directory `content` of the BBS directory.

- A path in `content` can have directories, for example `xfer/src/crc16.c`.
  Each part of the path has Latin letters, digits, `_`, `.` and `-`.
  It starts with a letter, a digit or `_`.
- The file is UTF-8 text, at most 1 MB. A line ends with LF or CR LF.
  A control character other than TAB is an error.
- The extension gives the "kind" of the file. Case does not matter:

| Extension | Kind |
|---|---|
| `.md`, `.markdown` | Markdown |
| `.txt`, or no extension | Plain text |
| Each other extension | Source |

Create the directory `content`. Create the files `content/lamp.md` and `content/rules.txt`
from the sections "Markdown" and "A plain text" below.

Full rules: [file-view](../../../specs/file-view.md), section 2.

## The screen `main`

The item `R` goes to the screen `texts`.
Its row is between the rows of `A` and `T`. The spacer line gives 2 empty rows less.
The file `screens/main.bbt` after this task (task 8 shows its last form):

```
+++
[[item]]
key = "A"
do = "gosub about"
[[item]]
key = "R"
do = "gosub texts"
[[item]]
key = "T"
do = "gosub term"
[[item]]
key = "G"
do = "logoff"
+++
{>_head}
       Good evening. It's quiet and warm here.
       Pull up a chair by the lamp.


         {!key}A{!}   About {!line}{~.}{!} {!dim}who we are{!}  

         {!key}R{!}   Reading room {!line}{~.}{!} {!dim}stories by the lamp{!}  

         {!key}T{!}   Your terminal {!line}{~.}{!} {!dim}size and colours{!}  

         {!key}G{!}   Goodbye {!line}{~.}{!} {!dim}the lamp will wait{!}  
{~}
       {!dim}The light stays on till morning.{!}

  {!line}{~─}{!}  

    {!key}A R T G{!}  {!dim}choose{!}

```

The screen:

```text

  ────────────────────────────────────────────────────────────────────────────

       Lamp  ☼ a BBS by lamplight

       Good evening. It's quiet and warm here.
       Pull up a chair by the lamp.


         A   About ................................................ who we are

         R   Reading room ................................ stories by the lamp

         T   Your terminal .................................. size and colours

         G   Goodbye ...................................... the lamp will wait


       The light stays on till morning.

  ────────────────────────────────────────────────────────────────────────────

    A R T G  choose
```

## The screen `texts`

`args` gives an argument to the screen of a `gosub` or a `goto`:

```
[[item]]
key = "1"
do = "gosub reader"
args = { file = "lamp.md" }
```

- The screen `reader` gets the argument `file` with the value `lamp.md`.
  The field `{arg.file}` gives this value.
- An argument name has lower case Latin letters, digits and `_`, and starts with a letter.
  A value cannot have `{` or a control character.
- A `gosub` or a `goto` gives exactly the arguments that its screen uses.
  A missing argument or another argument is an error.

Full rules: [bbt-format](../../../specs/bbt-format.md), section 8.5.

`examples/lamp/en/screens/texts.bbt`:
```
+++
[[item]]
key = "1"
do = "gosub reader"
args = { file = "lamp.md" }
[[item]]
key = "2"
do = "gosub reader"
args = { file = "rules.txt" }
[[item]]
key = "ESC"
do = "back"
+++
{>_head}
       The texts of the station.
       Pick one and read it by the lamp.


         {!key}1{!}   How the lamps were lit {!line}{~.}{!} {!dim}a short history{!}  

         {!key}2{!}   Station rules {!line}{~.}{!} {!dim}read them once{!}  
{~}
  {!line}{~─}{!}  

    {!key}1 2{!}  {!dim}choose{!}   {!key}Esc{!}  {!dim}back{!}

```

The screen:

```text

  ────────────────────────────────────────────────────────────────────────────

       Lamp  ☼ a BBS by lamplight

       The texts of the station.
       Pick one and read it by the lamp.


         1   How the lamps were lit .......................... a short history

         2   Station rules .................................... read them once








  ────────────────────────────────────────────────────────────────────────────

    1 2  choose   Esc  back
```

## The file screen `reader`

A "file screen" has the key `file` in its front matter.
`file` gives the content file: a path in `content`, or exactly one token `{arg.<name>}`.

```
+++
file = "{arg.file}"
margin = [7, 2]
```

- `{>file}` is the "file area": the place of the text. It is the only text of its line.
  A file screen has exactly one `{>file}`, in its own body: not in a fragment and not in `{#if}`.
- `margin = [left, right]` gives the empty cells left and right of the file area.
  `[7, 2]` gives the columns 8–78: the text column of Lamp. Without `margin`, the text starts at column 1.
  `left + right` must be less than the min width of the screen.
- The rows before `{>file}` are the "rows above". The rows after it are the "rows below".
  The file area takes the other rows of the terminal.
  At 24 rows, `reader` has 8 rows above and 5 rows below, so the file area has 11 rows.
- The BBS cuts the text into pages of the file area. A file screen has no More row:
  its items show the other pages.
- The last line of `reader.bbt` is empty. It gives the empty row 24 and keeps the lower line on row 21.
  Without it, the file area has 12 rows, and the lower line moves to row 22.
- With fewer than 14 rows, the BBS cuts each page at the last row of the terminal.
  With fewer than 13 rows, the key row is not visible.

Full rules: [file-view](../../../specs/file-view.md), sections 3, 3.1 and 3.2.

## Page actions

| `do` | Action |
|---|---|
| `page next` | Show the next page. On the last page, show the last page again. |
| `page prev` | Show the previous page. On the first page, show the first page again. |
| `page first` | Show the first page. |
| `page last` | Show the last page. |

- A `page` action works only in a file screen and in a list screen (task 8).
- `when = "not file.last"`: `Space` works on each page except the last.
- `when = "file.page > 1"`: `B` works from page 2.
- The key row shows a key only when its item works: `{#if item.SPACE}...{/if}`.
  So page 1 has no `B`, and the last page has no `Space`.

Full rules: [file-view](../../../specs/file-view.md), section 3.3,
and [menu-nav](../../../specs/menu-nav.md), section 5.1.

## The fields of a file

| Field | Value |
|---|---|
| `file.title` | The key `title` of the front matter. Else the text of the first level-1 heading. Else the file name, for example `rules.txt`. |
| `file.page` | The number of the page, from 1. |
| `file.pages` | The page count. |
| `file.last` | True on the last page. Use it only in a condition. |
| `file.section` | The text of the last heading of level 1 or 2 at the top of the page or above it. Else empty. |
| `file.minutes` | The reading time in minutes. |
| `meta.<key>` | The key `<key>` of the front matter. Else empty. |

- These fields work only in a file screen.
- The reading time is the word count divided by 200, rounded up, and at least 1.
  For Markdown, the BBS counts the words of the text, without the front matter and the addresses.
  For a plain text or source file, it counts the words of the lines that the screen shows.
- Row 7 of `reader` shows the author only when the file has `author`:
  `{#if meta.author != ""}{meta.author}  ∙  {/if}`.
- The lower line shows the page number in its centre:
  `{~─} {!dim}{file.page}/{file.pages}{!line} {~─}`. The two stretch tokens share the free cells.
  With an odd count of free cells, the right token gets one more cell.

Full rules: [file-view](../../../specs/file-view.md), sections 3.4 and 3.5.

`examples/lamp/en/screens/reader.bbt`:
```
+++
file = "{arg.file}"
margin = [7, 2]
[[item]]
key = "SPACE"
do = "page next"
when = "not file.last"
[[item]]
key = "B"
do = "page prev"
when = "file.page > 1"
[[item]]
key = "ESC"
do = "back"
+++
{>_head}
       {!title}{file.title}{!}
       {!dim}{#if meta.author != ""}{meta.author}  ∙  {/if}{file.minutes} min{!}

{>file}

  {!line}{~─} {!dim}{file.page}/{file.pages}{!line} {~─}{!}  

    {#if item.SPACE}{!key}Space{!}  {!dim}next{!}   {/if}{#if item.B}{!key}B{!}  {!dim}prev{!}   {/if}{!key}Esc{!}  {!dim}list{!}

```

The screen with `lamp.md`, page 1:

```text

  ────────────────────────────────────────────────────────────────────────────

       Lamp  ☼ a BBS by lamplight

       How the lamps were lit
       Anna Lind  ∙  1 min

       People lit the first lamps long before they wrote the first words. For
       thousands of years, a lamp gave a small, steady light.

       Oil and wick

         ∙ a bowl of clay for the oil;
         ∙ a wick of flax or cotton;
         ∙ a hand that trimmed the wick each evening.

       │ A good lamp does not smoke. It only asks for a little care.


  ─────────────────────────────────── 1/3 ────────────────────────────────────

    Space  next   Esc  list

```

- The title is in the light peach `title` in bold. Row 7 is in the grey-brown `dim`.
- The heading "Oil and wick" is orange. The markers `∙` are orange too.

## Markdown

The BBS shows Markdown: CommonMark with the tables of GitHub.

| Markdown | Look | Style |
|---|---|---|
| A paragraph | The text. It wraps at the edge of the file area. | `normal` |
| `# Heading` | The text. | `md_h1` |
| `## Heading` | The text. | `md_h2` |
| `### Heading` to `###### Heading` | The text. | `md_h3` |
| `**strong**` | The text. | `md_strong` |
| `*emphasis*` | The text. | `md_em` |
| `` `code` `` | The text. | `md_code` |
| `- item` | `∙ item`. An inner list has `-`. | `md_mark`, inner `md_frame` |
| `1. item` | `1. item` | `md_mark`, inner `md_frame` |
| `> quote` | `│ quote` | the bar `md_mark`, the text `md_em` |
| A code block | `│ code`. A long row does not wrap: `...` ends it. | the bar `md_frame`, the code `md_code` |
| `---` | `∙   ∙   ∙` | `md_mark` |
| A table | The columns with 2 spaces between them. A line `─` under the header. | the header `md_strong`, the line `md_frame` |
| `[text](https://...)` | The text and a note number, for example `[1]`. | `md_link`, the number `md_dim` |
| `![text](picture.png)` | The text, as a link. | `md_link` |
| `<https://...>` | The address. | `md_link` |
| HTML | The text as written. | `normal` |

- The other extensions of GitHub show as text, for example `~~text~~`.
- A link with an address that has a scheme, for example `https:`, has a note.
  After the last block, the BBS shows the notes: `[1] <address>` in `md_dim`.
  A link without a scheme, for example to another Markdown file, has no note.
  An address in `<` and `>` has no note.
- A level-1 heading starts a new page. So `lamp.md` has no level-1 heading.
- A heading stays on one page with the first row of the next block.
- A page does not start with an empty row.

The "front matter" of a Markdown file gives the fields `meta.*`. Write it in TOML between `+++` lines:

```
+++
title = "How the lamps were lit"
author = "Anna Lind"
+++
```

- The front matter starts on the first line of the file.
- Each key with one value gives the field `meta.<key>`. The key `title` also gives `file.title`.
- A first line `+++` without its end line is an error (task 10).
- YAML front matter between `---` lines is an error.

Full rules: [file-view](../../../specs/file-view.md), section 5.

`examples/lamp/en/content/lamp.md`:
```
+++
title = "How the lamps were lit"
author = "Anna Lind"
+++

People lit the first lamps long before they wrote the first words.
For thousands of years, a lamp gave a small, *steady* light.

## Oil and wick

- a bowl of clay for the oil;
- a wick of flax or cotton;
- a hand that trimmed the wick each evening.

> A good lamp does not smoke. It only asks for a little care.

## The glass chimney

In the 1780s, Aimé Argand put a glass chimney over a round wick.
The air came from below, and the flame became **much brighter**.
Seventy years later, the [kerosene lamp](https://en.wikipedia.org/wiki/Kerosene_lamp)
came to each home.

| Lamp | Fuel | Since |
|---|---|---|
| Clay lamp | olive oil | 5000 years ago |
| Argand lamp | whale oil | the 1780s |
| Kerosene lamp | kerosene | the 1850s |

---

## By the lamp tonight

Electric light came later, and the old lamps went to the shelf.
But a lamp in a window still says the same thing:
someone is home, and the door is open.

Pull up a chair. The light stays on.
```

Page 2, the file area:

```text
       The glass chimney

       In the 1780s, Aimé Argand put a glass chimney over a round wick. The
       air came from below, and the flame became much brighter. Seventy years
       later, the kerosene lamp[1] came to each home.

       Lamp           Fuel       Since
       ────────────────────────────────────────
       Clay lamp      olive oil  5000 years ago
       Argand lamp    whale oil  the 1780s
       Kerosene lamp  kerosene   the 1850s
```

Page 3, the file area:

```text
       ∙   ∙   ∙

       By the lamp tonight

       Electric light came later, and the old lamps went to the shelf. But a
       lamp in a window still says the same thing: someone is home, and the
       door is open.

       Pull up a chair. The light stays on.

       [1] https://en.wikipedia.org/wiki/Kerosene_lamp
```

## The styles of a text

The `md_*` styles give the colours of a text. In Lamp, `screens/screens.toml` has them (task 6):

- `md_h1` and `md_strong` are as `title`: light peach in bold;
- `md_h2` and `md_mark` are as `orn`: orange. `md_mark` is the list marker, the quote bar and the break;
- `md_h3`, `md_em` and `md_link` are light brown. A link is underlined: it is not a key;
- `md_code` is light olive;
- `md_frame` is as `line`, and `md_dim` is as `dim`.

A theme without `md_*` styles uses the default styles: white headings, a yellow list marker.
`[theme.256]` and `[theme.16]` of Lamp give the `md_*` styles for each depth.
With 16 colours, the emphasis looks as the text.

Full rules: [file-view](../../../specs/file-view.md), section 4.

## A plain text

A plain text file shows each line as it is. A long line wraps. The text is in `normal`.

`examples/lamp/en/content/rules.txt`:
```
1. Be kind. Everyone here is a guest.
2. Write as you speak to a friend at the table.
3. No ads and no spam.
4. Do not share the words of others without their consent.
5. Keep the secrets of the station's guests.
6. If you see a problem, tell the sysop.

Thank you, and welcome to Lamp.
```

The screen with `rules.txt`:

```text

  ────────────────────────────────────────────────────────────────────────────

       Lamp  ☼ a BBS by lamplight

       rules.txt
       1 min

       1. Be kind. Everyone here is a guest.
       2. Write as you speak to a friend at the table.
       3. No ads and no spam.
       4. Do not share the words of others without their consent.
       5. Keep the secrets of the station's guests.
       6. If you see a problem, tell the sysop.

       Thank you, and welcome to Lamp.




  ─────────────────────────────────── 1/1 ────────────────────────────────────

    Esc  list

```

The title is the file name: a plain text file has no front matter.

## A source file

A source file shows each line in `md_code`. A long line does not wrap: `...` ends it.
A TAB goes to the next multiple of 8 cells.

A screen can show a fixed file or a part of it:

```
+++
file = "src/lamp.c"
lines = [4, 8]
numbers = true
margin = [7, 2]
[[item]]
key = "ESC"
do = "back"
+++
{>file}
```

- `lines = [first, last]` shows the lines from `first` to `last`.
- `numbers = true` shows the line numbers in `md_dim`.
- `lines` and `numbers` work only for a plain text or source file.

The file area:

```text
       4 int main(void)
       5 {
       6         puts("The light stays on.");
       7         return 0;
       8 }
```

Full rules: [file-view](../../../specs/file-view.md), section 6.

## A BBS in two languages

A BBS can show its screens and texts in more than one language.
The user chooses a language. The BBS keeps the choice until the end of the session.

Each table `[lang.<code>]` of `screens/screens.toml` adds a language.
The code is two or three lower-case letters, for example `en` or `ru`.
`[screen] lang` is the language at the start of a session. Its default is `en`.

```toml
[screen]
lang = "en"

[lang.en]

[lang.ru]
more = "    дальше ∙ любая клавиша"
```

`more` of a language table gives the More row of this language.
Without it, the More row is `[screen] more`, then `-- More --`.

Put the files of a language into a directory with the name of its code:

```
screens/
  screens.toml
  main.bbt        the choice of the language
  _logo.bbt       one logo for all languages
  en/menu.bbt     the menu in English
  ru/menu.bbt     the menu in Russian
content/
  en/lamp.md
  ru/lamp.md
```

The BBS finds a file in the language of the session first, then in the common directory.
Example: in the language `ru`, the screen `menu` is `screens/ru/menu.bbt`, else `screens/menu.bbt`.
The same rule applies to a fragment, to art and to a content file, also in a common screen.
A file without text, for example a logo, stays in `screens/` once for all languages.

The common screen `main` asks for the language:

```
+++
[[item]]
key = "E"
do = "gosub en/menu"
[[item]]
key = "R"
do = "gosub ru/menu"
+++
{>_logo}
E  English   R  Русский
```

`gosub ru/menu` sets the language `ru` and shows the screen `menu` in this language.
The stack holds only the name `menu`. So `Esc` with `back` in the menu shows `main` again.

The action `lang <code>` changes the language on any screen.
The BBS then shows the same screen in the new language. A file screen keeps its page.

```
[[item]]
key = "L"
do = "lang en"
```

The field `session.lang` gives the language of the session:

```
{#if session.lang == "ru"}Язык: русский{#else}Language: English{/if}
```

- Each language must have the same screens with the same arguments.
  `velourbbs check` checks the screens of each language.
  Example: `screens/ru/news.bbt:1:1: no screen "news" (language "en")`.
- An error of a common file that only some languages find ends with the language.
  Example: `screens/about.bbt:6:6: item "M": do: unknown screen "menu" (language "de")`.
- `velourbbs preview menu --set session.lang=ru` shows the screen `menu` in Russian.

Full rules: [session-language](../../../specs/session-language.md), sections 2 to 6,
and [menu-nav](../../../specs/menu-nav.md), section 4.

## Test the pages

`reader` needs the argument `file`. Give it with `--set`:

```
velourbbs preview reader --set arg.file=lamp.md
```

The command gives the cell dumps of all pages of the text, one after the other.
Without `--set`, `velourbbs preview reader` gives the error `missing argument "file" of screen "reader"`.

`--keys` shows the pages. Start at `texts`: `1` opens the text in `reader`, `Esc` goes back to `texts`:

```
velourbbs preview texts --keys 1,SPACE,SPACE,B,ESC
```

```
screen texts
page 1 80x24 fill spaces
...
key 1
screen reader
page 1 80x24 fill spaces
...
key SPACE
page 2 80x24 fill spaces
...
key SPACE
page 3 80x24 fill spaces
...
key B
page 2 80x24 fill spaces
...
key ESC
screen texts
page 1 80x24 fill spaces
...
```

## Result

- `velourbbs check` gives no output.
- `velourbbs preview reader --set arg.file=lamp.md` gives three pages. Page 1 is the screen above, with `1/3` on the lower line.
- `velourbbs preview reader --set arg.file=lamp.md --keys SPACE,SPACE` gives page 3. Its key row is `B  prev   Esc  list`.
- In the telnet client, `R` shows the list, `1` shows the text, `Space` and `B` turn the pages,
  `Esc` goes back to the list.
