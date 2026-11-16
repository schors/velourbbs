# Idea: show a file on a screen

Backlog. It extends [bbt-format](../specs/bbt-format.md), [menu-nav](../specs/menu-nav.md),
[screen-render](../specs/screen-render.md) and [screens-toml](../specs/screens-toml.md).

## Goal

A screen shows a Markdown file, a plain text file, or a source file or a part of it.
The sysop does not copy the text into the screen.
A Markdown file has theme styles and a table of contents.

## Decisions

- One idea for the three kinds: Markdown, plain text and source code.
- The screen author inserts the file into the body of a `.bbt` screen, as an art insert.
  The screen gives the title and the keys.
- The insert names the file by a fixed path or by an argument of the screen, for example
  `{arg.file}` of [menu-actions](menu-actions.md). One reader screen shows many files.
- The files belong to the sysop and are in the BBS directory
  ([content-trust](content-trust.md), tier 2). Users do not upload files to show.
- A long file has pages. The user pages forward and back.
  The screen shows the page number, for example `4/4`
  ([brandbook](../brand/brandbook.md), screen "Тексты: чтение").
- The screen lines above and below the insert stay on each page.
  The file pages in the area between them.

## Markdown

- Theme styles for the Markdown elements are in `[theme]` of `screens/screens.toml`,
  with the depth themes. All Markdown files share them.
- A Markdown file can have front matter, for example `title`, `author` and `date`.
  The screen shows its keys as fields.
- The screen gets fields of the file: the title, the page and the page count,
  the current section and the reading time
  ([brandbook](../brand/brandbook.md), "5 минут чтения").
- A top-level heading starts a new page.
- A link to another Markdown file of the sysop opens that file in the same reader.

## Table of contents

- The table of contents (TOC) is a separate view of the file. The user opens it with a key
  and goes to a section with Enter ([brandbook](../brand/brandbook.md), screen
  "Тексты: оглавление").
- The screen author turns the TOC on or off for each insert. A file without the TOC has no key.
- The sysop writes the TOC template as a `.bbt` screen. One template row repeats for each
  heading, as `{#each}` of [lists](lists.md). The row has the heading text, its level and
  its page.

## Open questions

1. Render the file at each show, or compile it to `.bbt` or to flat art first.
2. The Markdown elements that the host shows, and the names of their theme styles.
3. How the screen names a part of a source file. Syntax colours and line numbers.
4. The paging keys and the More row of menu-nav. Shared paging with `page` of [lists](lists.md).
5. The encodings of a text file: UTF-8 only, or CP437 too.
6. The syntax of the insert: the file, the argument and the TOC switch.
7. How the TOC template finds its file and returns to the text at the selected page.
8. The headings in the TOC: all levels, or a maximum level.
9. The format of the Markdown front matter: TOML as in `.bbt`, or YAML.
10. A link to a web page or to a missing file: show the text only, or the address too.
11. How the host counts the reading time.

## Dependencies

- `{#each}` comes from [lists](lists.md).
- `args` and `{arg.*}` come from [menu-actions](menu-actions.md).
