# Idea: table of contents, links and syntax colours in the file view

Backlog. It extends [file-view](../specs/file-view.md).

## Goal

A Markdown file has a table of contents and active links to other Markdown files.
A source file has syntax colours.

## Decisions

- The table of contents (TOC) is a separate view of the file. The user opens it with a key
  and goes to a section with Enter ([brandbook](../brand/brandbook.md), screen
  "Тексты: оглавление").
- The screen author turns the TOC on or off for each file screen. A file without the TOC has no key.
- The sysop writes the TOC template as a `.bbt` screen. One template row repeats for each
  heading, as `{#each}` of [lists](lists.md). The row has the heading text, its level and
  its page.
- A link to another Markdown file of the sysop opens that file in the same file screen.
  `velourbbs check` gives an error for a link to a missing file.
- Syntax colours use theme styles. The [brandbook](../brand/brandbook.md), section 3, colours
  numbers, strings and preprocessor lines of C.

## Open questions

1. How the TOC template finds its file and returns to the text at the selected page.
2. The headings in the TOC: all levels, or a maximum level.
3. How the user selects a link: arrow keys over the links, as the marks of
   [menu-actions](menu-actions.md), or a number.
4. Syntax colours: a library (`github.com/alecthomas/chroma/v2` adds about 6 MB to the binary)
   or an own highlighter for a few languages.

## Dependencies

- `{#each}` and the selection of a row come from [lists](lists.md).
