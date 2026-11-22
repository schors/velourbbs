# Terminal detection

## What it does

Telnet gives only the size and the type name. Velour asks the terminal itself for the rest,
with escape sequences: the charset, the width of Cyrillic, the colour depth, the program name
and the background colour. The probe runs before the first screen.

## How

When telnet ends the negotiation, the client is in character mode and does not print the replies.
Then Velour sends the queries in one write, in this order:

1. Charset check: CR, the letter `â` (bytes `C3 A2`), the cursor position query `ESC[6n`.
   This is DSR of ECMA-48 ([ECMA-48](https://ecma-international.org/publications-and-standards/standards/ecma-48/)).
   The reply CPR is `ESC[<row>;<column>R`
   ([DSR-CPR, VT510](https://vt100.net/docs/vt510-rm/DSR-CPR.html)).
2. Width of Cyrillic: CR, the letter `ж` (bytes `D0 B6`), `ESC[6n`.
3. Truecolor: Velour sets the colour `ESC[38;2;1;2;3m` and asks which colour is set:
   DECRQSS for SGR, `ESC P $ q m ESC \`
   ([DECRQSS](https://vt100.net/docs/vt510-rm/DECRQSS.html),
   reply [DECRPSS](https://vt100.net/docs/vt510-rm/DECRPSS.html)).
   Then it resets the colour: `ESC[0m`.
4. Program name: XTVERSION, `ESC[>0q`.
5. Secondary attributes: DA2, `ESC[>c`.
6. Text and background colours: OSC 10 and OSC 11 with `?`, for example `ESC]11;?ESC \`.
7. Palette: OSC 4 with `?`, one query for each number: 0-15, 256 and 258.
8. Primary attributes: DA1, `ESC[c`.
9. Line erase: `CR ESC[K`. It removes a query that the terminal printed as text.

The xterm reference describes queries 4-8
([xterm control sequences](https://invisible-island.net/xterm/ctlseqs/ctlseqs.html)).

DA1 goes last. Each target terminal answers it, and the replies come
in the order of the queries. So the DA1 reply means: no more replies will come.
Velour waits for it or for 1 second of silence.

## What it gets

### Charset

In UTF-8, the letter `â` takes one cell, and the cursor goes to column 2.
In an 8-bit charset, the two bytes are two characters, and the cursor goes to column 3.
Then Velour prints an explanation in English on how to turn on UTF-8 and closes the connection.
With no reply, Velour takes UTF-8.

Why the letter `â`: by Unicode its width is N, narrow also in a CJK terminal
([UAX #11](https://www.unicode.org/reports/tr11/)). Its bytes are printable characters
in ISO-8859-1, CP437, CP866 and KOI8-R. The bytes 0x80-0x9F can be C1 controls
in 8-bit terminals, so the probe has none.

### Width of Cyrillic

By UAX #11, Cyrillic has the width A, Ambiguous: a CJK terminal can draw it in two cells.
When the cursor is in column 3 after `ж`, `term.ambiguous` is 2.
Then the screens break, and the entry screen tells how to turn this setting off.
The probe measures Cyrillic, not lines: in the wide mode of Windows Terminal, the lines
U+2500-U+259F stay narrow, and Cyrillic becomes wide.

### Colour depth

The first rule that matches gives `term.colors`:

1. The DECRQSS reply has the colour 1, 2, 3: the terminal kept RGB. This is truecolor.
   A terminal that maps RGB to its palette answers with a number: Tera Term sends `0;30`.
2. The DA2 reply is from PuTTY (`0;136;0`) or from tmux (first number `84`). They show RGB
   but do not answer DECRQSS. This is truecolor.
3. `term.type` ends with `-direct` or has `truecolor` or `24bit`: truecolor.
4. `term.type` has `256color`: 256 colours.
5. `term.type` is `ansi`, `vt100`, `xterm-color` or another name where 16 colours are certain: 16.
6. Each other type: 256.

Rule 5 has only certain names: a terminal without 256 colours reads `38;5;n`
as other attributes, for example `5` as blink.

### Name and background

- XTVERSION gives the name and the version of the program, for example `XTerm(412)` or `tmux 3.6`.
  This is `term.name`. It is for display only: the screens do not select output by it.
- OSC 11 gives the background colour, `term.bg_color`. PuTTY does not answer OSC 11,
  so for PuTTY the background comes from OSC 4 for the number 258.
- The background is light when its relative luminance is more than 0.5. The formula is from WCAG 2.2
  ([relative luminance](https://www.w3.org/TR/WCAG22/#dfn-relative-luminance)).
  The theme `auto` then takes the light colours.
- The palette 0-15 comes from OSC 4. When it differs from VGA, the colours of art go as RGB.
  So art keeps the colours of its artist.

Velour writes all replies in one line of the log. With them, the colour rules can be tuned.

## Limits

- PuTTY does not answer DECRQSS, OSC 10 and OSC 11. Version 0.83 and earlier
  print the DECRQSS query as the text `$qm`. Step 9 erases it.
- PuTTY does not answer OSC 4 with more than one number. So each number has its own query.
- tmux answers the queries itself and sends only OSC 4 to the outer terminal. After OSC 4, versions 3.6
  and later delay the next replies: up to 3.5 seconds in 3.6, about 0.6 seconds in 3.7.
  So OSC 4 is directly before DA1, and only the DA1 reply can be late.
- tmux answers CPR with its own widths. So `term.ambiguous` is 1
  also when the outer terminal draws Cyrillic wide.
- Windows Terminal 1.25 and later can send its own CPR query after an unknown sequence
  and take the next CPR reply. So the CPR probes go first.
- Windows Terminal 1.21 and earlier answers DA1 itself and at once. Nothing comes after the DA1 reply.
- SyncTERM and NetRunner have no UTF-8. They get the explanation and do not enter.
- Windows telnet.exe names itself `ANSI` and gets 16 colours. ANSI art and lines
  can break in it.
- MUD clients, for example Mudlet, do not do `ESC[2J` and cursor positions:
  a page becomes one line.
- There is no decision yet on terminals in CP437 and terminals without ANSI.
