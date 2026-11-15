# Idea: show a file on a screen

Backlog. It extends [bbt-format](../specs/bbt-format.md), [menu-nav](../specs/menu-nav.md)
and [screen-render](../specs/screen-render.md).

## Goal

A screen shows a Markdown file, a plain text file, or a source file or a part of it.
The sysop does not copy the text into the screen.

## Decisions

- One idea for the three kinds: Markdown, plain text and source code.
- The screen author inserts the file into the body of a `.bbt` screen, as an art insert.
  The screen gives the title and the keys.
- The files belong to the sysop and are in the BBS directory
  ([content-trust](content-trust.md), tier 2). Users do not upload files to show.
- A long file has pages. The user pages forward and back.
  The screen shows the page number, for example `4/4`
  ([brandbook](../brand/brandbook.md), screen "Тексты: чтение").

## Open questions

1. Render the file at each show, or compile it to `.bbt` or to flat art first.
2. The Markdown elements that the host shows, and their styles.
3. How the screen names a part of a source file. Syntax colours and line numbers.
4. The paging keys and the More row of menu-nav. Shared paging with `page` of [lists](lists.md).
5. The encodings of a text file: UTF-8 only, or CP437 too.
