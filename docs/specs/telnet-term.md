# Telnet and terminal detection

This specification defines the telnet server, the detection of the terminal at connect,
the output of pages, the key decoder and the connection limits.
Related parts: [bbt-format](bbt-format.md), [screen-render](screen-render.md) and [menu-nav](menu-nav.md).

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

The host must work with these telnet clients:

| Client | Notes |
|---|---|
| PuTTY | TTYPE `XTERM`. Accepts BINARY but does not ask for it. |
| Tera Term | TTYPE `xterm`. Starts in line mode. |
| inetutils telnet in a UTF-8 terminal | TTYPE is `$TERM` in upper case, then aliases. Sends CR NUL for Enter in character mode. |

- SyncTERM and NetRunner show no UTF-8. The probe finds them as "not UTF-8" (section 5.4).
- The PROXY protocol senders are HAProxy, nginx, Traefik and stunnel.
  They send version 1 or 2, so the host accepts both (section 9.3).
  AWS NLB works only when it sends the header at connect (section 9.3).

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
  `output_timeout` (section 9.1) applies to each write.
- IAC doubling runs after the encoder, because 0xFF is a character in some 8-bit charsets.
- The host always sends CR LF as the line end.

## 5. Probe

### 5.1. Connect sequence

1. Check `max_conns`. Read the PROXY protocol header (section 9.3). Check the address limits.
   If a limit is reached, send an ASCII line and close (section 9.2).
2. Send the requests of section 4.1. After `WILL TTYPE`, send TTYPE `SEND` (section 5.3).
3. Wait for the replies (below).
4. Send the charset probe (section 5.4). Wait for its reply (below).
5. Decide the terminal values (section 6). Show the first screen. It clears the probe output.

- In step 3, the host waits for an answer to `DO NAWS` and `DO TTYPE`, for the first NAWS
  after `WILL NAWS` and for the first TTYPE name after `WILL TTYPE`.
  In step 4, the host waits for the reply to the charset probe.
- A wait ends when all these replies arrive, or when no reply arrives for 1 s.
- The PROXY protocol header and the probe end within `probe_timeout` (section 9.1)
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

### 5.4. Charset probe

The charset probe is `CR`, U+00E2 (`â`, bytes `C3 A2`), `ESC[6n`. Then `CR ESC[K` erases the probe.
The reply is a CPR: `ESC[<row>;<col>R`.

- A number of the reply has 1–5 digits. Else the host drops the reply.
- Charset: the reply column 2 means UTF-8. The column 3 means "not UTF-8":
  the cursor moved by the byte count of `â`. Each other reply, or no reply, means UTF-8.
- U+00E2 has the East Asian Width N, so a CJK terminal also shows it in one cell. Both
  of its bytes are printable characters in ISO-8859-1, CP437, CP866 and KOI8-R. A byte from
  0x80 to 0x9F is a C1 control in some 8-bit terminals, so the probe does not use one.

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
the NAWS size, the CPR reply, the terminal values of section 6
and the probe time. So the TTYPE colour rules can be tuned.
The TTYPE name in the log is filtered and cut as `term.type` (section 6).

## 6. Terminal values

| Field | Value |
|---|---|
| `term.type` | The TTYPE name in lower case. Without a name, `ansi`. |
| `term.cols`, `term.rows` | Section 5.2. |
| `term.charset` | `utf-8`. |
| `term.colors` | `16`, `256` or `16777216`. |

- `term.type` is untrusted client input. The host removes the control characters and cuts
  the name to 40 graphemes. A name that is empty after this counts as no name.
  The value rules of [bbt-format](bbt-format.md), section 4.3, also apply.
- Colour depth: the first rule that matches `term.type` gives it:
  1. a name that ends with `-direct`, or that has `truecolor` or `24bit`: 16777216;
  2. a name that has `256color`: 256;
  3. a name that ends with `-16color` or `-8color`, or that is `ansi`, `dumb`, `vt52`,
     `vt100`, `vt102`, `vt220`, `vt320`, `rxvt`, `xterm-color`, `eterm-color` or `cygwin`: 16;
  4. each other name: 256.
- Rule 3 has only names where 16 colours are certain. A terminal without `38;5` reads
  `38;5;n` as other SGR parameters, for example `5` as blink. Without a name, `term.type`
  is `ansi`, so the depth is 16.
- PuTTY (`xterm`) and Tera Term (`xterm`) get 256. They show 256 colours.
  Truecolor art is reduced to 256 for them.

## 7. Output

### 7.1. Page

The host shows a page ([menu-nav](menu-nav.md), section 5) in this order:

1. `ESC[0m ESC[2J`: reset the style and clear the terminal screen.
2. For each row of the page: `ESC[<row>;1H`, then the cells of the row.
3. `ESC[0m`.

- A wide grapheme appears once. Its second cell gives no output.
- Each row starts at its own position, so a page has no line ends
  and the terminal does not scroll.

### 7.2. Styles

Before a cell whose style differs from the style of the previous cell, the host sends
`ESC[0;<params>m`, or `ESC[0m` when there are no parameters. After step 1 of section 7.1,
the style is `default` colours without attributes. The parameters are the attributes,
then the foreground, then the background:

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

The encoder writes each grapheme in UTF-8. 8-bit charsets are later encoders.

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

- After ESC and after each byte of a sequence, the decoder waits 150 ms for the next byte.
  When no byte arrives, ESC alone gives `ESC`, and the decoder drops a started sequence.
- After ESC, `[` starts a CSI sequence and `O` starts an SS3 sequence. Each other byte
  gives `ESC`, and then the decoder handles the byte as new input.
  So Alt+letter gives `ESC` and the letter.
- A CSI sequence has parameter bytes 0x30–0x3F, intermediate bytes 0x20–0x2F and one final byte
  0x40–0x7E. An SS3 sequence has one final byte.
- An invalid byte ends the sequence. The decoder drops the sequence and handles the byte
  as new input.
- The decoder drops a sequence longer than 32 bytes up to its end, also its final byte.
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

### 8.3. Key queue

- The key queue holds at most 64 keys. The host drops a key when the queue is full.
- The idle timeout starts at the end of the probe.
- Keys reset the idle timeout, also a key that the full queue drops.
  NOP, NAWS and dropped input do not reset it.

## 9. Limits

### 9.1. `bbs.toml`

```
[telnet]
listen = ":2323"
max_conns = 64
max_conns_per_addr = 4
max_connects_per_min = 10
allow = ["192.0.2.10", "2001:db8::/48"]
proxy = ["127.0.0.1"]
idle_timeout = 600
probe_timeout = 5
output_timeout = 60
```

| Key | Default | Meaning |
|---|---|---|
| `listen` | `":2323"` | The TCP address of the server: `<host>:<port>`. The host can be empty. |
| `max_conns` | 64 | The maximum number of connections, counted from accept. |
| `max_conns_per_addr` | 4 | The maximum number of connections from one address. |
| `max_connects_per_min` | 10 | The maximum number of connects from one address in 60 s. |
| `allow` | `[]` | Addresses and prefixes without the address limits. |
| `proxy` | `[]` | Addresses and prefixes of the PROXY protocol senders. |
| `idle_timeout` | 600 | Seconds without a key until the host disconnects. |
| `probe_timeout` | 5 | Seconds until the probe ends (section 5.1). |
| `output_timeout` | 60 | Seconds until a write must complete. Else the host disconnects. |

- Each integer is from 1 to 1000000. The port of `listen` is from 1 to 65535.
- An address is an IPv4 or IPv6 address. A prefix is an address with `/<bits>`.
- `velourbbs check` checks `[telnet]` with the rules of [screen-render](screen-render.md), section 3.1.

### 9.2. Connection limits

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
  for client data before the header does not work. AWS NLB waits by default.

### 9.4. Output

- The host writes each page in one write. A write that does not complete within
  `output_timeout` ends the session.

## 10. Command

`velourbbs serve` starts the telnet server for the BBS directory, the current directory.

- It does the checks of [screen-render](screen-render.md), section 3.1, at start.
  It writes the errors and the warnings to the standard error.
- It writes the log to the standard error: a `listen` line with the address,
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
- It exits with status 1 after an error of the checks or of `listen`.
- It runs until SIGINT or SIGTERM. Then it closes the listener and the connections and exits
  with status 0.

## 11. Code

| Package | Content |
|---|---|
| `internal/telnet` | The telnet parser, the option states and the subnegotiations. |
| `internal/term` | The probe, the encoder, the colour reduction, the key decoder and the terminal of [menu-nav](menu-nav.md), section 8. |
| `internal/server` | The listener, the limits, the PROXY protocol header and the sessions. |
| `internal/config` | `[telnet]` in `bbs.toml`. |

- `cmd/velourbbs` loads the screens and gives them to `internal/server`.

## 12. Libraries

| Task | Choice | Compared with | Reason |
|---|---|---|---|
| Telnet | Own parser | Go telnet libraries | Libraries do not give server-side NAWS and TTYPE with bounded buffers. The parser has about 200 lines. |
| PROXY protocol | Own parser | `github.com/pires/go-proxyproto` | The library has API v0.x. Its 2026 releases changed the default policy and the parser, and not all changes are marked as breaking. A receiver of versions 1 and 2 has about 200 lines of the standard library. |
| Terminal output | Own encoder | `github.com/gdamore/tcell` | The full-screen model of tcell differs from pages with a More row. It does not do the probe. |
| Log | `log/slog` | Other log libraries | The standard library is sufficient. |

## 13. Not in the MVP

- SSH. The size and the terminal type come from `pty-req` and `window-change`.
- Web client.
- Charset negotiation (telnet CHARSET, RFC 2066).
- 8-bit charset output. CP437 is the first encoder, for SyncTERM and NetRunner.
- Terminals without ANSI and monochrome output.
- Detection of terminals that show ambiguous-width characters as wide.
- Mouse input.
- MUD clients (Mudlet, TinTin++). Mudlet does not execute `ESC[2J` and cursor positions,
  so a page becomes one line.
- Windows telnet.exe. Its UTF-8 support is not known.
