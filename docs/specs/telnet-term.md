# Telnet and terminal detection

This specification defines the telnet server, the detection of the terminal at connect,
the output of pages, the key decoder and the connection limits.
Related parts: [bbt-format](bbt-format.md), [bbs-toml](bbs-toml.md), [screen-render](screen-render.md),
[menu-nav](menu-nav.md), [line-input](line-input.md) and [live-screen](live-screen.md).

## 1. Goal

A user connects with telnet. The host finds the terminal type, the screen size, the charset
and the colour depth. Then it starts the session at the first screen ([menu-nav](menu-nav.md),
section 5).

## 2. Concepts

- The "terminal" is the session view of the client: size, type, capabilities, key input
  and output. It does not depend on telnet, so SSH can give the same data later.
- The "probe" is the detection step at connect, before the first screen.
- The "encoder" is the last step of output. It converts cells to bytes for the charset
  of the terminal.
- The "key decoder" converts input characters to keys: the key names and the characters
  of [bbt-format](bbt-format.md), section 8.1, and [menu-nav](menu-nav.md), section 3.1.
- The "control characters" are those of [bbt-format](bbt-format.md), section 2.1.
- In the byte sequences of this specification, spaces and commas only separate the parts.
  The host does not send them.

## 3. Target implementations

The host must work with these telnet clients: PuTTY, Tera Term, and inetutils telnet
in a UTF-8 terminal. inetutils telnet passes the bytes unchanged, so its terminal answers
the queries of section 5.7. The host must work with inetutils telnet in Windows Terminal (in WSL),
xterm and xterm-like terminals, tmux, iTerm2, Ghostty and cmux.

- The host supports other clients on a best-effort basis.
- The PROXY protocol senders are HAProxy, nginx, Traefik and stunnel.
  They send version 1 or 2 (section 9.3).

### 3.1. Client quirks

The rules of this specification handle these quirks of the clients and the terminals:

| Client | Quirk | Rule |
|---|---|---|
| All target terminals | They answer DA1 and send the replies in the order of the queries. | DA1 goes last (section 5.7). |
| All target terminals | No escape sequence switches the ambiguous width. A write to the second cell of a wide character erases the character. | The layout does not use `term.ambiguous` (section 6). |
| PuTTY, Tera Term, xterm | The terminfo entries have `xenl` and `bce`. In PuTTY, bce is a setting, on by default. | A page writes the last cell. `fill = "erase"` works (section 7.1). |
| PuTTY | TTYPE `XTERM`. It accepts BINARY but does not ask for it. | Section 4.1. |
| PuTTY | It sends no reply to an OSC 4 query with more than one index. | One query for each index (section 5.7). |
| PuTTY 0.71 and later | It shows RGB. It answers no DECRQSS, no OSC 10 and no OSC 11. It answers OSC 4 for 256 (the default foreground) and 258 (the default background). | The client table and the colours of section 6. |
| PuTTY 0.83 and earlier | It shows the DECRQSS query as the text `$qm`. | The erase of section 5.7. |
| PuTTY, inetutils telnet | They echo the input before character mode. | The queries go after step 3 of section 5.1. |
| Tera Term | TTYPE `xterm`. It starts in line mode. It answers DECRQSS with an index colour: `0;30`. | Section 4.1. It gets 256 colours (section 6). |
| inetutils telnet | TTYPE is `$TERM` in upper case, then aliases. It sends CR NUL for Enter in character mode. | Sections 5.3, 6 and 8.1. |
| Windows Terminal 1.19 and later, xterm, Ghostty | They send the RGB reply `38:2::1:2:3`. iTerm2 sends `38:2:1:1:2:3`. xterm without direct colour sends `0;38:5:16`. | The RGB reply of section 5.7. |
| Other target terminals | For the OSC 4 indexes 256 and 258, xterm and Tera Term give special colours, Windows Terminal 1.22 and later give reserved entries, iTerm2 gives black, tmux and Ghostty send no reply. No target terminal shows these queries as text. | The host sends them to each terminal and uses the reply for 258 only for PuTTY (section 6). |
| Windows Terminal 1.25 and later | It can send its own CPR query after an unknown sequence and take the next CPR reply. | The CPR probes go first (section 5.7). |
| Windows Terminal | Its wide mode keeps U+2500–U+259F narrow and makes Cyrillic wide. | The ambiguous probe uses a Cyrillic letter (section 5.4). |
| Windows Terminal 1.21 and earlier | Its ConPTY answers DA1, DA2 and DECRQSS itself, at once. It drops OSC 10 and OSC 11 and answers no OSC 4 and no XTVERSION. | No reply comes after the DA1 reply. |
| Windows Terminal | Its DA1 reply has 39 bytes. | The limit of 128 bytes (section 8.2). |
| tmux | It answers the queries itself. It sends only OSC 4 to the outer terminal. It answers OSC 10 and OSC 11 with the colours of the pane or of the outer terminal, or sends no reply. | Section 5.7. |
| tmux 3.6 and later | After OSC 4, it delays its next replies until the outer terminal answers: at most about 0.6 s in 3.7 and later, 3.5 s in 3.6, 3.6a and 3.6b. | OSC 4 goes directly before DA1, so only the DA1 reply can be late (section 5.7). |
| tmux | DA2 first parameter `84`. It reduces RGB for the outer terminal itself. 3.6 and later send an invalid DECRQSS reply. | The client table of section 6. |
| tmux | It answers the CPR with its own widths. | `term.ambiguous` is 1 also when the outer terminal shows wide characters (section 6). |
| tmux, GNU screen | No `bce`. | The fill rows of `fill = "erase"` have the default colours (section 7.1). |
| PuTTY, xterm, Tera Term, VTE | With the default settings, an emoji with VS16 (`❤️`) takes one cell. A ZWJ sequence (`👨‍👩‍👧`) takes two cells for each emoji in PuTTY, xterm and VTE. | The re-sync of section 7.4. |
| Windows telnet.exe | TTYPE `ANSI`. Its UTF-8 support depends on the console code page. Its escape sequence interpreter breaks ANSI art and frames. | Best effort. It gets 16 colours (section 6). |
| SyncTERM, NetRunner | No UTF-8. | Section 5.5. |
| AWS NLB | By default, it waits for client data before it sends the PROXY protocol header. | It does not work (section 9.3). |

## 4. Telnet

### 4.1. Options

At connect, the host sends these requests in one write:

| Request | Meaning |
|---|---|
| `WILL ECHO` | The host echoes, so the client stops its local echo. |
| `WILL SGA`, `DO SGA` | Character mode. |
| `WILL BINARY`, `DO BINARY` | 8-bit data in both directions. |
| `DO NAWS` | The client sends its window size. |
| `DO TTYPE` | The client sends its terminal type. |

The host keeps the state of each option for each side, as in RFC 1143:

- The host accepts these options: on its side ECHO, SGA and BINARY;
  on the client side SGA, BINARY, NAWS and TTYPE.
- A request to enable an accepted option that is disabled: the host enables it and agrees.
- A request to enable each other option: the host refuses it, `DONT` for `WILL`, `WONT` for `DO`.
- A request to disable an enabled option: the host disables it and agrees.
- The host does not answer a request that matches the current state, or that answers
  a request of the host. So the negotiation cannot loop.
- A client that refuses an option works without it, except character mode.
  The host does not support a client that refuses character mode. Such a client sends
  a line on Enter. The host reads the keys of the line, and the Enter of the line is
  one more key. So the next screen can run its `ANY` or `ENTER` item at once.

### 4.2. Parser

- The telnet parser runs before the UTF-8 decoder and the key decoder.
  An IAC can split an escape sequence or a multibyte character.
- `IAC IAC` gives the data byte 0xFF.
- AYT: the host sends BEL (0x07). It does not change the screen.
- IP and BRK end the session, as a hangup.
- The host ignores each other command.
- The host ignores a subnegotiation of an option that is not enabled on the client side.
- The data of a subnegotiation has at most 256 bytes, after `IAC IAC` gives 0xFF and without
  the option byte. The host disconnects when a subnegotiation is longer, or when it
  does not end within 10 s.
- `IAC` and a byte other than `SE` and `IAC` inside a subnegotiation drop the subnegotiation.
  The host handles the `IAC` and the byte as a command.
- The host ignores a NAWS subnegotiation whose data is not 4 bytes.

### 4.3. Output

- The host writes each reply to the client before it reads more input.
  `output_timeout` ([bbs-toml](bbs-toml.md), section 3.3) applies to each write.
- IAC doubling runs after the encoder, because 0xFF is a character in some 8-bit charsets.
- The host always sends CR LF as the line end.

## 5. Probe

### 5.1. Connect sequence

1. Check `max_conns`. Read the PROXY protocol header (section 9.3). Check the address limits.
   If a limit is reached, send an ASCII line and close (section 9.2).
2. Send the requests of section 4.1. After `WILL TTYPE`, send TTYPE `SEND` (section 5.3).
3. Wait for the replies (below).
4. Send the CPR probes (section 5.4) and the terminal queries (section 5.7) in one write.
   Wait for the replies (below).
5. Decide the terminal values (section 6). Show the first screen. It clears the probe output.

- In step 3, the host waits for an answer to `DO NAWS` and `DO TTYPE`, for the first NAWS
  after `WILL NAWS` and for the first TTYPE name after `WILL TTYPE`.
  In step 4, the host waits for the reply to DA1 (section 5.7), or for a CPR reply
  that means "not UTF-8" (section 5.4).
- A wait ends when all these replies arrive, or when no reply arrives for 1 s.
  In step 4, the first reply to each query of section 5.7 is a reply.
- The PROXY protocol header and the probe end within `probe_timeout`
  ([bbs-toml](bbs-toml.md), section 3.1)
  from accept. Then the host decides with the replies that it has.
- Keys that arrive during the probe go to the key queue.

### 5.2. Size

- NAWS gives the size. A value with 0 in a dimension gives no size.
- Without a NAWS size, the size is 80x24.
- The host clamps the size to 20–255 columns and 10–100 rows.
- A NAWS update after the probe changes `term.cols` and `term.rows`.
  An update with 0 in a dimension does not change them.
  The session uses the new size at the next screen ([menu-nav](menu-nav.md), section 5).

### 5.3. Type

- The host sends `SEND` once. The first TTYPE name is the terminal name.
- The host ignores each later TTYPE subnegotiation.

### 5.4. CPR probes

Each CPR probe writes one character at the start of the line and asks for the cursor position:

| Probe | Bytes | Measures |
|---|---|---|
| Charset probe | `CR`, U+00E2 (`â`, bytes `C3 A2`), `ESC[6n` | The charset. |
| Ambiguous probe | `CR`, U+0436 (`ж`, bytes `D0 B6`), `ESC[6n` | The cells of an East Asian Ambiguous character. |

The charset probe goes first. `CR ESC[K` after the queries of section 5.7 erases the probes.
The reply is a CPR: `ESC[<row>;<col>R`. The first CPR reply is the reply to the charset probe.
The second CPR reply is the reply to the ambiguous probe.

- A number of the reply has 1–5 digits. Else the host drops the reply.
- Charset: the reply column 2 means UTF-8. The column 3 means "not UTF-8":
  the cursor moved by the byte count of `â`. Each other reply, or no reply, means UTF-8.
- Ambiguous width: the reply column 3 gives `term.ambiguous` 2. Each other reply,
  or no reply, gives 1.
- U+00E2 has the East Asian Width N, so a CJK terminal also shows it in one cell.
- U+0436 has the East Asian Width A. The probe uses a Cyrillic letter,
  not a box drawing character (section 3.1).
- The bytes of both characters are printable characters in ISO-8859-1, CP437, CP866 and KOI8-R.
  A byte from 0x80 to 0x9F is a C1 control in some 8-bit terminals, so the probes do not use one.

### 5.5. Not UTF-8

When the charset probe gives "not UTF-8", the host sends this ASCII text and disconnects:

```
This BBS needs a UTF-8 terminal. Your terminal uses another charset.
Set the charset (encoding) of your client to UTF-8 and connect again.
SyncTERM has no UTF-8. Use PuTTY, Tera Term or telnet in a UTF-8 terminal.
```

Each line ends with CR LF. The text starts with `CR LF`.

### 5.6. Log

At the end of the probe, the host logs one line: the address, the TTYPE name,
the NAWS size, the CPR replies, the charset (`utf-8` or `not utf-8`), the replies of section 5.7,
the terminal values of section 6 and the probe time. So the colour rules can be tuned.

- In the TTYPE name and in the replies to DECRQSS, DA1, DA2 and XTVERSION, the host removes
  the control characters and cuts the text to 40 graphemes, with no case change.
  The log has each reply from the byte after the introducer
  (`ESC P`, `ESC[`) to the byte before the terminator: for example `1$r0;38:2::1:2:3m`.
- The palette is the 16 colours of the OSC 4 replies as `rrggbb`, separated by commas.
  An index without a reply is `-`.
- The replies to OSC 10, OSC 11 and OSC 4 for 256 and 258 are colours as `rrggbb`.
- A missing reply is `-`.

### 5.7. Terminal queries

The host asks the terminal for its colours and its name.
Step 4 of section 5.1 sends this write:

| Query | Bytes | Reply |
|---|---|---|
| Charset probe | Section 5.4. | CPR. |
| Ambiguous probe | Section 5.4. | CPR. |
| DECRQSS for SGR | `ESC[38;2;1;2;3m`, `ESC P $ q m ST`, `ESC[0m` | `ESC P 1 $ r <parameters> m ST`, or `ESC P 0 $ r ST`. |
| XTVERSION | `ESC[>0q` | `ESC P >| <text> ST`. |
| DA2 | `ESC[>c` | `ESC[> <parameters> c`. |
| OSC 10 | `ESC]10;? ST` | `ESC]10;rgb:<r>/<g>/<b>` and ST or BEL. |
| OSC 11 | `ESC]11;? ST` | `ESC]11;rgb:<r>/<g>/<b>` and ST or BEL. |
| OSC 4 | `ESC]4;<n>;? ST` for each `<n>` from 0 to 15, then 256 and 258 | `ESC]4;<n>;rgb:<r>/<g>/<b>` and ST or BEL. |
| DA1 | `ESC[c` | `ESC[? <parameters> c`. |
| Erase | `CR ESC[K` | None. |

`ST` is `ESC \`.

- The CPR probes go first, in the order of section 5.4. DA1 goes last, so the DA1 reply
  comes after the other replies. OSC 4 goes directly before DA1.
- Each OSC 4 index has its own query.
- The host sends the OSC 4 queries for 256 and 258 to each terminal, because the one write
  goes before the DA2 reply. It uses the reply for 258 only for PuTTY (section 6).
- The host only logs the replies to OSC 10 and to OSC 4 for 256 (section 5.6).
- The erase removes a query that a terminal shows as text.
- The host sends the queries after step 3, so the client is in character mode
  and does not echo the replies.
- Section 3.1 gives the client quirks behind these rules.
- The host takes the first reply to each query, and for OSC 4 to each index.
  It drops a reply before step 4, a second reply and a reply that does not match this section.

Reply rules:

- DECRQSS: the "RGB reply" is a reply with `1$r` whose parameters, split at `;`, have
  the colour 1, 2, 3 in one of these forms: the five parameters `38`, `2`, `1`, `2`, `3`,
  or one parameter `38:2:1:2:3` or `38:2:<id>:1:2:3`, where `<id>` is empty or a number.
  A terminal that maps RGB to its palette sends an index colour (section 3.1).
- OSC 4: `<n>` is 0–15, 256 or 258. Each of `<r>`, `<g>` and `<b>` has 1–4 hex digits.
  A value v with d digits gives the 8-bit value `round(v × 255 / (16^d − 1))`.
  For example, `ffff` and `ff` give 255, `8080` gives 128.
- OSC 10 and OSC 11: `<r>`, `<g>` and `<b>` have the rules of OSC 4.
- XTVERSION: `<text>` is the name and the version of the terminal, for example `XTerm(412)`
  or `tmux 3.6`.
- DA1 and DA2: the parameters are numbers separated by `;`.
- The "palette" is the colours of the OSC 4 replies for 0–15. The palette is "not VGA" when
  a reply has a colour other than the VGA colour of its index (section 7.3).

## 6. Terminal values

| Field | Value |
|---|---|
| `term.type` | The TTYPE name in lower case. Without a name, `ansi`. |
| `term.cols`, `term.rows` | Section 5.2. |
| `term.colors` | `16`, `256` or `16777216`. |
| `term.name` | The `<text>` of the XTVERSION reply (section 5.7). Without a reply, empty. |
| `term.ambiguous` | `1` or `2`: the ambiguous probe (section 5.4). |
| `term.bg_color` | The background colour of the terminal as `#rrggbb` in lower case: the OSC 11 reply. Without it, the OSC 4 reply for 258 when the DA2 reply is the reply of PuTTY in the client table below. Else empty. |
| `term.background` | `light` when the relative luminance of `term.bg_color` is more than 0.5. Else `dark`. Empty when `term.bg_color` is empty. |

- `term.type` and `term.name` are untrusted client input. The host removes the control
  characters and cuts the text to 40 graphemes. A name that is empty after this counts as no name.
  The value rules of [bbt-format](bbt-format.md), section 4.3, also apply.
- `term.name` is for display only. Screens must not select output by it.
- The relative luminance of the colour R, G, B (0–255) is `0.2126 R' + 0.7152 G' + 0.0722 B'`.
  For each channel, with c = value / 255: c' is c / 12.92 when c ≤ 0.04045,
  else ((c + 0.055) / 1.055)^2.4. So `#bcbcbc` and lighter greys are light, `#bbbbbb` is dark.
- `term.background` selects the session theme with `[screen] theme = "auto"`
  ([screens-toml](screens-toml.md), section 3.2).
- `term.ambiguous` 2 comes from a setting of the terminal, off by default.
  The probe sees only the width of U+0436.
- The layout does not use `term.ambiguous` (section 3.1): with 2, the columns of a page move,
  and frames and art break. A screen can show the user how to turn the setting off.
- Colour depth: the first rule that matches gives it:
  1. the DECRQSS reply is the RGB reply (section 5.7): 16777216;
  2. the DA2 reply is in the client table below: 16777216;
  3. a `term.type` that ends with `-direct`, or that has `truecolor` or `24bit`: 16777216;
  4. a `term.type` that has `256color`: 256;
  5. a `term.type` that ends with `-16color` or `-8color`, or that is `ansi`, `dumb`, `vt52`,
     `vt100`, `vt102`, `vt220`, `vt320`, `rxvt`, `xterm-color`, `eterm-color` or `cygwin`: 16;
  6. each other `term.type`: 256.
- The client table has the terminals that show RGB but do not answer DECRQSS with it:

  | Client | DA2 reply |
  |---|---|
  | PuTTY | The parameters are `0;136;0`. |
  | tmux | The first parameter is `84`. |

- Rule 5 has only names where 16 colours are certain. A terminal without `38;5` reads
  `38;5;n` as other SGR parameters, for example `5` as blink. Without a name, `term.type`
  is `ansi`, so rule 5 gives 16.

## 7. Output

### 7.1. Page

The host shows a page ([menu-nav](menu-nav.md), section 5) in this order:

1. `ESC[?2026h ESC[0m ESC[2J`: start the synchronized output, reset the style and clear
   the terminal screen.
   With `fill = "erase"`, the SGR of `normal` (sections 7.2 and 7.3) in place of `ESC[0m`:
   the terminal clears the screen in the background of `normal`.
2. For each terminal row from the top, with the terminal rows of [bbt-format](bbt-format.md),
   section 6.5:
   - a row of the page: `ESC[<row>;1H`, then the cells of the row;
   - a fill row with `fill = "spaces"`: `ESC[<row>;1H`, then `term.cols` spaces in `normal`.
3. `ESC[0m ESC[?2026l`: reset the style and end the synchronized output.

- An update of a page has its own steps (section 7.5).
- With the synchronized output (mode 2026), the terminal shows the page at once, not row
  by row. The host sends it without a query. A terminal without the mode ignores it.
- A wide grapheme appears once. Its second cell gives no output.
- Each row starts at its own position, so a page has no line ends
  and the terminal does not scroll (one exception: section 7.4).
- A page writes the last cell of the last terminal row when its rows or its fill rows
  with `fill = "spaces"` reach that row. A terminal with `xenl` delays the wrap
  after the last column, so it does not scroll (section 3.1).
- `fill = "erase"` needs bce. In a terminal without bce, the fill rows have the default colours.

### 7.2. Styles

Before a cell whose style differs from the style of the previous cell, the host sends
`ESC[0;<params>m`, or `ESC[0m` when there are no parameters. After step 1 of section 7.1,
the style is `default` colours without attributes, or `normal` with `fill = "erase"`.
The parameters are the attributes, then the foreground, then the background:

| Item | Parameter |
|---|---|
| `bold`, `underline`, `blink` | `1`, `4`, `5`. |
| Colour `default` | None. |
| Index 0–7 | `30+n`, background `40+n`. |
| Index 8–15 | `90+n-8`, background `100+n-8`. |
| Index 16–255 | `38;5;n`, background `48;5;n`. |
| RGB | `38;2;r;g;b`, background `48;2;r;g;b`. |

### 7.3. Colour reduction

The host reduces each colour of a cell to `term.colors` before the output:

- First, a VGA colour ([bbt-format](bbt-format.md), section 7.1) becomes its RGB value
  of the table below when the palette is not VGA (section 5.7). Else it becomes its index.
  With 16 colours, the RGB value becomes the same index again.
- 16777216: no change.
- 256: an RGB colour becomes the nearest index from 16 to 255.
- 16: an index from 16 to 255 and an RGB colour become the nearest index from 0 to 15.
- An index 0–15 and `default` do not change.
- "Nearest" is the smallest squared RGB distance. On a tie, the smaller index wins.
- Indexes 16–255 have the RGB values of the xterm palette. Indexes 0–15 have the VGA values:

```
0 000000  1 aa0000  2 00aa00  3 aa5500  4 0000aa  5 aa00aa  6 00aaaa  7 aaaaaa
8 555555  9 ff5555 10 55ff55 11 ffff55 12 5555ff 13 ff55ff 14 55ffff 15 ffffff
```

### 7.4. Encoder

The encoder writes each grapheme in UTF-8.

A "re-sync" follows a grapheme of more than one code point when a cell of the row follows
the last cell of the grapheme. The re-sync is a space in the style of the grapheme,
then `ESC[<row>;<col>H`. `<row>` is the terminal row. `<col>` is the terminal column
after the last cell of the grapheme.

- Terminals give different widths to these graphemes (section 3.1).
- After the re-sync, the next cells are in their columns.
  - A terminal that shows the grapheme narrow writes the space after it.
    The space has the background of the grapheme.
  - A terminal that shows the grapheme wide writes the next cells over the end of the grapheme.
    When the grapheme passes the last column, the terminal wraps it. The output of the next
    terminal row overwrites the wrapped part. A fill row with `fill = "erase"` has no output,
    so the part stays there. On the last terminal row, the wrap scrolls the page.
  - When no cell follows the grapheme, a narrow terminal leaves the last cell empty.

### 7.5. Update

An update shows a new grid ([live-screen](live-screen.md), section 2) in place of the grid
of the last page or update: a cursor move in a list screen ([menu-nav](menu-nav.md), section 5.3)
and each frame of a live page ([live-screen](live-screen.md), section 5).
The host sends it in this order:

1. `ESC[?2026h ESC[?25l`: start the synchronized output and hide the cursor.
2. For each terminal row from the top, for each run of the row: `ESC[<row>;<col>H`,
   then the cells of the run.
3. `ESC[0m`, `ESC[<row>;<col>H` to the cursor place, `ESC[?25h ESC[?2026l`.

- A "run" starts and ends at a cell that differs from the shown grid. Two runs with at most
  3 equal cells between them are one run. Then each run grows until these rules hold:
  - a run that starts or ends in a grapheme of width 2 takes the whole grapheme;
  - a run that starts directly after a grapheme of more than one code point takes that grapheme.
    A terminal that shows the grapheme wide uses the next cell for it;
  - a run that ends with a grapheme of more than one code point takes the next cell of the row, if there is one.
    The re-sync of section 7.4 writes a space over that cell.
- The styles have the rule of section 7.2. Before the first cell of step 2, the style is `default`
  colours without attributes. A cell of a fill row with `fill = "none"` is a space in this style.
- The encoder of section 7.4 writes the cells, with the re-sync.
- The "cursor place" is the place where the last page leaves the cursor: the last terminal row
  that the page writes, at the column `term.cols`. When the page writes no row, it is row 1, column 1.
- An update without a changed cell sends nothing.
- PuTTY 0.83, Tera Term, xterm and tmux before 3.7 have no mode 2026. They show the cells
  as they arrive. The hidden cursor does not run over the screen.
- In a terminal without bce, a changed cell of a fill row with `fill = "erase"` gets the background of `normal`.

## 8. Key decoder

### 8.1. Input

The input goes through these steps:

1. Telnet parser (section 4.2).
2. Enter: CR gives `ENTER`. An LF or a NUL directly after a CR is dropped. Each other LF
   gives `ENTER`. So CR LF, CR NUL, a single CR and a single LF each give one `ENTER`.
3. UTF-8 decoder. It drops each byte that is not valid UTF-8 (RFC 3629).
4. Key decoder.

The key decoder gives:

| Input | Key |
|---|---|
| 0x08, 0x7F | `BS` |
| 0x09 | `TAB` |
| 0x20 | `SPACE` |
| 0x1B | `ESC`, or the start of a sequence (section 8.2) |
| Other control characters | None. They are dropped. |
| Other characters | The character. |

### 8.2. Escape sequences

- After ESC and after the byte that starts a sequence or a control string, the decoder waits
  150 ms for the next byte. After each later byte, it waits 1 s. So a reply that arrives
  in two parts does not become keys. When no byte arrives, ESC alone gives `ESC`,
  and the decoder drops a started sequence or control string.
- After ESC, `[` starts a CSI sequence and `O` starts an SS3 sequence.
  `]` (OSC), `P` (DCS) and `_` (APC) start a control string. Each other byte
  gives `ESC`, and then the decoder handles the byte as new input.
  So Alt+letter gives `ESC` and the letter, except Alt+`P`. Alt+`]`, Alt+`P` and Alt+`_`
  give no key.
- A control string ends with ST (`ESC \`) or BEL. The decoder drops it as a whole.
  So a late reply to a query of section 5.7 does not become keys.
  In a control string, ESC followed by a byte other than `\` ends the string. The decoder drops
  the string and handles ESC as new input. Other control characters do not end the string.
- A CSI sequence has parameter bytes 0x30–0x3F, intermediate bytes 0x20–0x2F and one final byte
  0x40–0x7E. An SS3 sequence has one final byte.
- An invalid byte ends the sequence. The decoder drops the sequence and handles the byte
  as new input.
- The decoder drops a sequence or a control string longer than 128 bytes up to its end,
  also its final byte or its terminator.
- The decoder drops each sequence that is not in the table as a whole.
  It never becomes separate keys. Keys with modifiers, for example `ESC[1;5A`, are dropped.
- A CPR reply after the probe is dropped.

| Key | Sequences |
|---|---|
| `UP`, `DOWN`, `RIGHT`, `LEFT` | `CSI A`–`CSI D`, `SS3 A`–`SS3 D` |
| `HOME` | `CSI H`, `SS3 H`, `CSI 1~`, `CSI 7~` |
| `END` | `CSI F`, `SS3 F`, `CSI 4~`, `CSI 8~` |
| `DEL` | `CSI 3~` |
| `PGUP`, `PGDN` | `CSI 5~`, `CSI 6~` |
| `F1`–`F4` | `SS3 P`–`SS3 S`, `CSI 11~`–`CSI 14~` |
| `F1`–`F5` | `CSI [A`–`CSI [E` (Linux console) |
| `F5`–`F12` | `CSI 15~`, `CSI 17~`–`CSI 21~`, `CSI 23~`, `CSI 24~` |

`CSI [` is a CSI sequence with the final byte `[`. One more byte from `A` to `E` follows it.

`CSI 200~` and `CSI 201~` are the paste marks (section 8.4). They are not keys.

### 8.3. Key queue

- The key queue holds at most 256 keys. The host drops a key when the queue is full.
  So the queue holds a pasted line of the line input ([line-input](line-input.md), section 6):
  `term.cols` is at most 255. Section 8.4 has the rule
  of a full queue in a paste.
- After `idle_timeout` seconds without a key, the host disconnects.
- The idle timeout starts at the end of the probe.
- Keys reset the idle timeout, also a key that the full queue drops.
  NOP, NAWS, the paste marks and dropped input do not reset it.
  A key of a paste is a key, also when section 8.4 drops it.

### 8.4. Paste

The line input turns on the bracketed paste mode ([line-input](line-input.md), section 6).
Then the terminal sends a paste between two marks: `CSI 200~` (the paste start)
and `CSI 201~` (the paste end).

- A "paste" is the keys from a paste start to the next paste end or the next paste start.
  It has no time limit: a late part of a paste is still a part of the paste.
- The line input takes the keys of a paste with their paste mark.
- Outside the line input, the host drops each key of a paste.
  A paste that the user starts in the line input can end after it.
- When the key queue is full in a paste, the host drops the rest of the paste.
  So a later part of a paste does not continue the line after a lost `ENTER`.
- When the mode is on and the host ends the session or shows `shutdown` without the end
  of the input, it first sends `ESC[?2004l`: at the idle timeout, at IP and BRK (section 4.2),
  at a protocol error and at a stop (section 10.2). It ignores a write error.
- A terminal without the mode ignores `ESC[?2004h` and sends no marks.
- Tests cover each rule of this section: the marks, a paste in two reads, a full queue
  in a paste, the keys of a paste outside the line input and `ESC[?2004l` at each end of the session.

## 9. Limits

### 9.1. `bbs.toml`

The keys of `[telnet]` and `[limits]` are in [bbs-toml](bbs-toml.md), sections 3.1 and 3.3.

### 9.2. Connection limits

- `max_conns` limits the connections of the host. It counts a connection from accept.
- The "address limits" are `max_conns_per_addr` and `max_connects_per_min`.
  `max_conns_per_addr` limits the open connections of one address.
  `max_connects_per_min` limits the connects of one address in 60 s.
- An address in `allow` has no address limits. `max_conns` applies to it.
- "One address" is one IPv4 address, or one IPv6 /64 prefix.
  An IPv4-mapped IPv6 address is the IPv4 address.
- A connect that `max_conns` refuses does not count for the address limits.
- Each other connect counts for `max_connects_per_min`, also a refused one.
  The host counts the connects in steps of 1 s. The 60 s are the second of the connect
  and the 59 seconds before it.
- The connect counter holds at most 10000 addresses. When it is full, the host removes
  the address with the oldest connect.
- A connection over a limit gets one ASCII line and CR LF. Then the host closes it:
  `Too many connections. Try again later.`

### 9.3. PROXY protocol

- The host reads a PROXY protocol header only from a connection whose peer address
  is in `proxy`. From other peers, it reads no header.
- The header is version 1 or version 2 of the HAProxy specification.
- The header gives the client address. The limits and the log use it.
- These headers keep the peer address: `UNKNOWN` (version 1), and `LOCAL`, `UNSPEC`
  and `UNIX` (version 2). A connection with such a header has no address limits.
- The host skips the TLVs of version 2.
- The bytes after the header are telnet data.
- Version 1 has at most 107 bytes. Version 2 has at most 16 + 536 bytes.
- An invalid or missing header closes the connection without output.
- The proxy must send the header at connect. The host sends first, so a proxy that waits
  for client data before the header does not work.

### 9.4. Output

- The host writes each page and each update (section 7.5) in one write. A write that does not complete within
  `output_timeout` ends the session.

## 10. Command

`velourbbs serve` starts the telnet server for the BBS directory, the current directory.

- It does the checks of [screen-render](screen-render.md), section 3.1, at start.
  It writes the errors and the warnings to the standard error.
- It writes the log to the standard error: a `listen` line with the address and its source,
  the probe line (section 5.6) and a `close` line for each connection.
  The `close` line has the client address, the reason of the end and the time.
- After an accept error, it logs the error and accepts again after 1 s.
- A panic in a session ends only this session. The host logs it with the stack
  and closes the connection.
- After output that ends a connection (section 5.5, section 9.2, `logoff`), the host
  ends its write side and discards the input until EOF, for 1 s at most. Then it closes
  the connection. A close with unread input sends a TCP reset, and the client can lose
  the output. The connection does not count for the limits of section 9.2 during this time.
  At most `max_conns` connections do this at the same time. The host closes each other
  connection at once.
- It exits with status 1 after an error of the checks, of the data file
  ([data-store](data-store.md), section 7) or of `listen`.
- It runs until SIGINT or SIGTERM. Then it closes the listener, stops the sessions
  (section 10.2) and exits with status 0.

### 10.1. Listener

- On Linux, when `LISTEN_PID` is the process ID and `LISTEN_FDS` is set, systemd passes
  the listener (socket activation). The host uses file descriptor 3 and ignores `telnet.listen`.
  The source in the `listen` line is `systemd`.
- `LISTEN_FDS` must be `1`, and file descriptor 3 must be a listening TCP socket.
  Else it is an error of `listen`.
- Else the host listens on `telnet.listen` ([bbs-toml](bbs-toml.md), section 3.1).
  The source in the `listen` line is `bbs.toml`.
- The other rules of this specification are the same for the two sources.

### 10.2. Stop

At SIGINT or SIGTERM, the host does these steps:

1. Close the listener. A connection in the probe closes at once.
2. Stop each open session ([session-fields](session-fields.md), section 3.2) at its next
   wait for a key, or after the output of its current page. Show the screen `shutdown`
   ([menu-nav](menu-nav.md), section 5.2).
3. After 5 s, close each connection that is still open. Then exit.

- The reason in the `close` line of a connection of steps 1 and 2 is `shutdown`.
- The close after output of section 10 applies after `shutdown`. Its 1 s counts in the 5 s.

## 11. Code

| Package | Content |
|---|---|
| `internal/telnet` | The telnet parser, the option states and the subnegotiations. |
| `internal/term` | The probe, the encoder, the colour reduction, the key decoder, the paste, the draw of the line input, the update of section 7.5, the terminal of [menu-nav](menu-nav.md), section 8, and the line form of [screen-render](screen-render.md), section 3.7. |
| `internal/server` | The listener, the limits, the PROXY protocol header and the sessions. |

- `cmd/velourbbs` loads the model ([session](session.md), section 3) and gives it to `internal/server`.
  It opens the listener of section 10.1.

## 12. Libraries

| Task | Choice | Compared with | Reason |
|---|---|---|---|
| Telnet | Own parser | Go telnet libraries | Libraries do not give server-side NAWS and TTYPE with bounded buffers. The parser has about 200 lines. |
| PROXY protocol | Own parser | `github.com/pires/go-proxyproto` | The library has API v0.x. Its 2026 releases changed the default policy and the parser, and not all changes are marked as breaking. A receiver of versions 1 and 2 has about 200 lines of the standard library. |
| Terminal output | Own encoder | `github.com/gdamore/tcell` | The full-screen model of tcell differs from pages with a More row. It does not do the probe. |
| Log | `log/slog` | Other log libraries | The standard library is sufficient. |
| Socket activation | Own code | `github.com/coreos/go-systemd/v22/activation` | The host takes one socket. The code reads two variables and calls `net.FileListener`: about 30 lines of the standard library. |

## 13. Not in the MVP

- SSH. The size and the terminal type come from `pty-req` and `window-change`.
- Web client.
- Charset negotiation (telnet CHARSET, RFC 2066).
- Terminals without ANSI and monochrome output.
- A layout for terminals that show ambiguous-width characters as wide (section 6).
- Mouse input.
- MUD clients (Mudlet, TinTin++). Mudlet does not execute `ESC[2J` and cursor positions,
  so a page becomes one line.
- The size by a CPR for a client without NAWS. Each target client sends NAWS.
