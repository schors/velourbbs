# Idea: art converter

Backlog. It extends [screen-format](../specs/screen-format.md), section 8.

## Goal

A sysop converts CP437 ANSI art to the UTF-8 art that the host loads.
Old art from archives and art from PabloDraw then work without an art editor.

## Decisions

| Topic | Decision |
|---|---|
| Output | UTF-8 art that the loader of the screen format accepts. |
| Input | CP437 ANSI art, with or without SAUCE. |
| Use | A one-time conversion. The host never reads CP437. |

## Facts from the design research

- The art editors draw the bytes 0x01–0x1F and 0x7F as CP437 glyphs, for example 0x01 as U+263A.
  Only LF, CR, ESC and 0x1A are controls in all of them.
- `golang.org/x/text/encoding/charmap` decodes 0x00–0x1F and 0x7F as controls.
  The converter needs its own table for them.
- PabloDraw writes RGB colours as `ESC[0;R;G;Bt` (background) and `ESC[1;R;G;Bt` (foreground).
- A CP437 file can be valid UTF-8. Example: the CP437 bytes of `─┐` are C4 BF, which is U+013F.
  So the converter cannot detect the input charset from the bytes.

## Open questions

1. The command: `bbs convert` or a separate tool.
2. Where the converter gets the input charset when the file has no SAUCE.
