# Idea: telnet and terminal detection

MVP part 3 of 3. The other parts are [screen-format](screen-format.md) and
[menu-nav](menu-nav.md).

## Goal

A user connects with telnet. The host finds the terminal type, the screen size, the charset and
the colour depth. Then it starts the session at the first screen.

## Concepts

- **Terminal.** The session view of the client: size, type, capabilities, key input and output.
  It does not depend on telnet, so SSH can give the same data later.
- **Probe.** The detection step at connect, before the first screen.
- **Encoder.** The last step of output. It converts cells to bytes for the charset of the terminal.
- **Key decoder.** It converts input bytes to key names (`ESC`, `ENTER`, letters).

## Decisions

| Topic | Decision |
|---|---|
| Client | Telnet only. SSH is in the backlog. |
| Telnet parser | Our own parser, about 200 lines. Libraries do not give server-side NAWS and TTYPE with bounded buffers. |
| Telnet options | BINARY in both directions, SGA, server ECHO, NAWS, TTYPE with MTTS. |
| Size | NAWS. Without NAWS, a CPR probe (`ESC[999C ESC[999B`, then a position request). Without both, 80x25. NAWS 0x0 means no size. The host clamps sizes to 20-255 columns and 10-100 rows. |
| Resize | A NAWS update changes `term.cols` and `term.rows`. The host uses the new size at the next screen. |
| Type | TTYPE gives the name. MTTS gives the capability bits. The host asks TTYPE 4 times at most and stops when a name repeats. |
| Charset check | MTTS bits, and a CPR probe with U+00C0: a UTF-8 terminal moves the cursor one column. The host erases the probe row after the probe. |
| Not UTF-8 | Only on positive evidence: the cursor moves by the byte count of the probe character. Then the host sends an ASCII message and disconnects. The message tells how to set UTF-8 in the client. |
| No answer | In all other cases, the host uses UTF-8 and ANSI. |
| Colour depth | From MTTS bits. Without MTTS, from the TTYPE name: `*-direct`, `*truecolor*`, `*24bit*` give truecolor; `*256color*` gives 256; other names give 16. |
| Colour model | Cells hold full colours. The output reduces each colour to the depth of the terminal. |
| Charset output | The encoder writes UTF-8. 8-bit charsets are later encoders with a fallback table. No other part changes. |
| Line mode client | A client that refuses character mode sends a line on Enter. The host reads the first key of the line. There is no special logic. |
| Enter key | CR LF, CR NUL, a single CR and a single LF from the client give one `ENTER`. The host always sends CR LF. |
| Fields | `term.type`, `term.cols`, `term.rows`, `term.charset`, `term.colors`. |
| `term.type` | Untrusted client input. The host filters it as user content and cuts it to 40 characters. |
| Log | The host logs the probe results, so the TTYPE colour rules can be tuned. |

## Connect sequence

1. Check the session and address limits. If a limit is reached, send an ASCII line and close.
2. Send all telnet requests in one write. Answer TTYPE subnegotiations.
3. Send the CPR probe if MTTS and NAWS do not give the charset and the size.
4. Wait for the replies. The probe ends when all replies arrive, or after about 1 s.
5. Decide the terminal values. Show the first screen. It clears the probe output.

## Protocol rules

- The telnet parser runs before the UTF-8 decoder and the key decoder.
  IAC can split an escape sequence or a multibyte character.
- IAC doubling runs after the encoder, because 0xFF is a character in some 8-bit charsets.
- The host never answers an acknowledgement, so negotiation cannot loop (RFC 1143).
- Subnegotiation buffers are bounded by size and by time.
- The host ignores NOP. It handles AYT, IP and BRK.

## Key decoder

- An unknown escape sequence is dropped as a whole. It never becomes separate keys.
- A CPR reply after the probe is dropped.
- The ESC timeout is 150 ms.
- The key queue is bounded.
- The Russian layout map (menu-nav) applies after UTF-8 decoding.

## Limits

All values are in `bbs.toml`.
- Idle timeout. The default is 10 minutes. Only keys reset it, not NOP or NAWS.
- Maximum number of connections, counted from accept, not from the session start.
- Maximum number of connections from one address. IPv6 addresses count by /64.
- Maximum connect rate from one address.
- An allowlist of addresses without the address limits.
- A hard total deadline for the probe.
- A bounded output queue. When the client does not read it in 60 s, the host disconnects.
- PROXY protocol (v1 and v2) for a TCP proxy in front of the BBS. The host accepts the header only
  from the proxy addresses in `bbs.toml`.

## MVP scope

Out (backlog):
- SSH.
- Charset negotiation (telnet CHARSET, RFC 2066).
- 8-bit charset output. CP437 is the first encoder, for SyncTERM and NetRunner.
- Terminals without ANSI and monochrome output.
- Detection of terminals that show ambiguous-width characters as wide.
- Mouse input.

## Dependencies

- [screen-format](screen-format.md) uses `term.*` and the colour model.
- [menu-nav](menu-nav.md) uses the key names from the key decoder.

## Open questions

1. The list of named keys: proposal `ESC ENTER TAB BS DEL UP DOWN LEFT RIGHT HOME END PGUP PGDN
   F1-F12`. 0x08 and 0x7F both give `BS`.
2. ESC followed by a letter (Alt+letter): one key or two.
3. The text of the "not UTF-8" message.
