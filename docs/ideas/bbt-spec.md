# Idea: bbt-spec

Backlog. It moves the BBT format out of [screen-format](../specs/screen-format.md)
and [menu-nav](../specs/menu-nav.md) into its own specification.

## Goal

One specification defines the whole `.bbt` file format.
The screen author and the guides read the format in one place.

## Decisions

- The specification holds the whole file format: the structure, the front matter,
  the tokens, the includes, the art inserts, the fields, the conditions, the styles,
  and the items with `key` and `do`.
- The design specifications refer to it and do not repeat it.

## Open questions

- The border between the file format and the host behaviour:
  the layout, the text wrap, the colour reduction.
- The place of the key names (menu-nav §3) and of the art flat form (screen-format §8).
