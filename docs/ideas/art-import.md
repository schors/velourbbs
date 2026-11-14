# Idea: strict art load and art import

Backlog. It extends [screen-format](../specs/screen-format.md), section 8.

## Goal

The host skips nothing in art silently. Art editors and old art files stay before the BBS:
the host reads one flat form of art.

## Decisions

The idea has two steps, in this order.

### Step 1. Strict art load

| Topic | Decision |
|---|---|
| Unknown input | An unknown sequence, an unknown SGR parameter or mode, an incomplete sequence and an invalid colour are errors with a position. Now they are warnings (section 8.3). |
| Editor no-ops | The host accepts the known sequences of the target editors that do nothing, with no diagnostic. Example: the icy_draw font page `CSI 0;1 SP D`. |
| Animation | A write to a cell that the play already wrote keeps the last frame. `bbs check` warns. |

### Step 2. Art import

| Topic | Decision |
|---|---|
| Command | `bbs art import <input> <output>`. |
| Input | Any ANSI art: CP437 or UTF-8, with or without SAUCE, with cursor moves. |
| Output | The "flat form": UTF-8 graphemes, SGR, LF and SAUCE. |
| Host | The host accepts only the flat form. A cursor move is an error. |
| SAUCE | The import copies SAUCE to the output: title, author, group and comments. |
| Animation | The import keeps the last frame and warns. |
| Use | A one-time conversion for each file. The host never reads CP437. |

## Facts from the design research

- The art editors draw the bytes 0x01–0x1F and 0x7F as CP437 glyphs, for example 0x01 as U+263A.
  Only LF, CR, ESC and 0x1A are controls in all of them.
- `golang.org/x/text/encoding/charmap` decodes 0x00–0x1F and 0x7F as controls.
  The import needs its own table for them.
- PabloDraw writes RGB colours as `ESC[0;R;G;Bt` (background) and `ESC[1;R;G;Bt` (foreground).
- A CP437 file can be valid UTF-8. Example: the CP437 bytes of `─┐` are C4 BF, which is U+013F.
  So the import cannot detect the input charset from the bytes.
- icy_draw writes `CSI 0;1 SP D` (font page) in its files, for example
  `internal/art/testdata/icydraw.ans`.
- The host loads 30 screens and 30 art files of 80x25 (about 750 KB) in 84 ms.
  So a compiled screen file or a cache gives no gain (measured on 2026-11-14).

## Open questions

1. The full list of the editor no-op sequences.
2. Whether static art from Moebius and icy_draw writes cursor moves or writes a cell twice.
   If it writes a cell twice, the animation warning is false.
   If it writes cursor moves, the sysop must import these files too.
3. The exact flat form, for example iCE colours and the width from SAUCE.
4. Where the import gets the input charset when the file has no SAUCE.
5. Whether the sysop keeps the source file next to the imported file.
