# Line input

This specification defines the line input: a field where the user types one line of text.
Related parts: [telnet-term](telnet-term.md), [menu-nav](menu-nav.md) and [bbt-format](bbt-format.md).

## 1. Goal

A user can type a line of text: a oneliner, a search text, an answer.
Without line input, the session reads only hotkeys ([menu-nav](menu-nav.md), section 3).

## 2. Concepts

- The "line" is the text that the user types. It has at most one row.
- The "field" is the place of the line on the terminal screen.
- The "caller" is the part of the host that asks for a line, for example `ask`
  ([args](https://github.com/schors/velourbbs/issues/66)).
- The "cursor position" is a position in the line, from 0 (before the first character)
  to the length of the line (after the last character).
- A "paste" and its marks: [telnet-term](telnet-term.md), section 8.4.
- The line input is user content of tier 1 ([content trust](https://github.com/schors/velourbbs/issues/5)):
  hostile input.

## 3. Field

The caller gives:

| Value | Meaning |
|---|---|
| Row, column | The terminal row and the terminal column of the first cell of the field, 1-based. |
| Maximum length M | The most characters of the line, M >= 1. |
| Style | The style of the cells of the field. |
| Masked | True: each character shows as `*`. |

- The field has M+1 cells on one terminal row. The first M cells hold the line.
  The last cell holds the cursor after a full line. So the line is always visible as a whole.
  There is no horizontal scroll.
- The field must fit the size of the current screen ([menu-nav](menu-nav.md), section 5, step 1):
  column + M <= `term.cols`, and row <= `term.rows`. Another field is a host bug:
  the line input panics, and `velourbbs serve` recovers the panic of one session.
- The field keeps its place until the input ends. A resize takes effect at the next screen.
- The line starts empty.

## 4. Keys

The line input takes the keys of [telnet-term](telnet-term.md), section 8, in this order:

| Key | Action |
|---|---|
| A character, `SPACE` | Insert the character at the cursor position (section 5). `SPACE` gives a space. The cursor moves after it. |
| `LEFT`, `RIGHT` | Move the cursor one character. At the start or the end of the line, no move. |
| `HOME`, `END` | Move the cursor to the start or the end of the line. |
| `BS` | Remove the character before the cursor. At the start of the line, no change. The cursor moves before the removed character. |
| `DEL` | Remove the character at the cursor. At the end of the line, no change. |
| `ENTER` | End the input. The caller gets the line. |
| `ESC` | Cancel the input. The caller gets "cancelled" and no line. |

- Each other key is ignored: `TAB`, `UP`, `DOWN`, `PGUP`, `PGDN` and `F1`–`F12`.
- The Russian layout map of [menu-nav](menu-nav.md), section 3.2, does not apply.
  `ы` gives `ы`.
- An empty line is a valid line. The caller decides what it does.
- The idle timeout of [telnet-term](telnet-term.md), section 8.3, applies.

## 5. Characters

- The "grapheme rule": each character of the line is one grapheme of width 1
  ([bbt-format](bbt-format.md), section 2.1).
- The line input inserts a character when the line has fewer than M characters,
  the character is not a control character ([bbt-format](bbt-format.md), section 2.1)
  and the line after the insert keeps the grapheme rule. Else it drops the character.
- The line input checks the grapheme rule with a letter before the line and a space after it.
  So a character at the start or at the end of the line does not join the text around the field:
  U+0903 at the start and U+0D4E (a Prepend character) at the end are invalid.
- `BS` and `DEL` make no change when the line after the removal breaks the grapheme rule.
  Example: in `ᅡxᅡ` (U+1161, `x`, U+1161), the removal of `x` joins the two U+1161
  into one grapheme.
- So the line input rejects wide characters (emoji, CJK), characters of width 0
  (combining marks, U+200B, variation selectors) and a character that joins
  its neighbour into one grapheme, for example U+0903 after a letter.
  Terminals give different widths to these graphemes ([telnet-term](telnet-term.md), section 7.4).
- East Asian Ambiguous characters have the width 1. They are valid.
- The line has M cells at most, so it passes the value rules of [bbt-format](bbt-format.md),
  section 4.3, without a change.

## 6. Paste

- At the start of the input, the host turns on the bracketed paste mode:
  `ESC[?2004h`. At the end of the input, after `ENTER` and after `ESC`, it turns the mode off:
  `ESC[?2004l`.
- In a paste, a character, `SPACE` and `ENTER` act as in section 4. So the first line end
  of a paste ends the input. Each other key of a paste is dropped, also `ESC` and `BS`.
- At the end of the input, the host clears the key queue. The keys of an open paste
  are dropped up to its end ([telnet-term](telnet-term.md), section 8.4).
  So the rest of a paste does not press the hotkeys of the next screen.
- A terminal without the mode sends a paste as typed keys. Then only the queue clear applies:
  the keys that arrive after the end of the input are keys of the next screen.
- The key queue holds a pasted line of M characters and its `ENTER`
  ([telnet-term](telnet-term.md), section 8.3).

## 7. Output

- At the start of the input, the host draws the field.
- After a key that changes the line or the cursor position, the host draws the field again.
  When the key queue has more keys, it first takes them, then draws once.
- `internal/line` gives the draw as data: the M+1 cells and the cursor position.
  `internal/term` writes it in one write:
  1. `ESC[<row>;<column>H`;
  2. the M+1 cells of the field in its style: each character of the line, or `*` when masked,
     then spaces;
  3. `ESC[0m`;
  4. `ESC[<row>;<column + cursor position>H`.
- The style goes out as in [telnet-term](telnet-term.md), sections 7.2 and 7.3.
- The field stays on the terminal screen after the end of the input. The caller shows
  the next page.

## 8. Code

| Package | Content |
|---|---|
| `internal/line` | The field, the line, the keys of section 4, the characters of section 5, the keys of a paste (section 6) and the cells of a draw (section 7). It has no I/O. |
| `internal/term` | The terminal of the line input: the keys with their paste mark, the bytes of a draw, the paste mode and the queue clear. |

- The terminal interface of the session ([menu-nav](menu-nav.md), section 8) does not change.
  The first caller adds the line input to it.

## 9. Acceptance

- A test covers each key of section 4 at the start, in the middle and at the end of the line,
  and the length limit M.
- A test covers the characters of section 5: a control character, a wide character,
  a character of width 0, U+0903 after a letter and at the start, U+0D4E at the end,
  an East Asian Ambiguous character
  and a Cyrillic letter. A test covers `BS` and `DEL` in `ᅡxᅡ`.
- A test covers the masked mode: the draw shows `*`, the caller gets the line.
- A test covers an empty `ENTER` and `ESC`.
- A test covers the bytes of a draw, with `ESC[?2004h` at the start and `ESC[?2004l`
  at the end of the input.
- A test covers a paste of two lines in two reads: the line gets the first line,
  and the next key after the input is a key after the paste end.
- A test covers a paste of M = 254 characters and its `ENTER` in one read.
- The tests of [telnet-term](telnet-term.md), section 8.4, pass.

## 10. Not in this specification

- A multi-line editor.
- A first value of the line.
- A caller in the screens: `ask` ([args](https://github.com/schors/velourbbs/issues/66)).
