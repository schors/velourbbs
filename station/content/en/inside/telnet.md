# Telnet

## What it does

Velour puts the client into character mode and finds the size and the type of the terminal.

## How

Telnet sends commands inside the data stream. A command starts with the byte IAC (255)
([RFC 854](https://www.rfc-editor.org/rfc/rfc854)). The commands
`WILL`, `WONT`, `DO` and `DONT` turn options on and off. The data of an option goes between `IAC SB` and `IAC SE`.

Directly after the connection, Velour sends in one write:

- `WILL ECHO` ([RFC 857](https://www.rfc-editor.org/rfc/rfc857)).
  The server echoes, and the client stops its own echo of the input.
- `WILL SGA` and `DO SGA` ([RFC 858](https://www.rfc-editor.org/rfc/rfc858)).
  With ECHO, this is character mode: the client sends each key at once, not a line on Enter.
- `WILL BINARY` and `DO BINARY` ([RFC 856](https://www.rfc-editor.org/rfc/rfc856)).
  8-bit data in both directions: so the bytes of UTF-8 pass.
- `DO NAWS` ([RFC 1073](https://www.rfc-editor.org/rfc/rfc1073)): the client sends the window size.
- `DO TTYPE` ([RFC 1091](https://www.rfc-editor.org/rfc/rfc1091)): the client sends the terminal type.
  When the client answers `WILL TTYPE`, Velour sends `IAC SB TTYPE SEND IAC SE` once.

Velour keeps the state of each option for each side, by the Q method
of [RFC 1143](https://www.rfc-editor.org/rfc/rfc1143). It does not answer a request that matches
the current state. So the negotiation cannot loop.
Velour refuses the options that are not in the list: `DONT` for `WILL`, `WONT` for `DO`.

The telnet parser runs before the UTF-8 decoder and the key decoder:
an IAC can split a multibyte character or an escape sequence.
`IAC IAC` gives the byte 255. Velour answers AYT with BEL. IP and BRK end the session.

## What it gets

- Size. NAWS gives the width and the height as two 16-bit numbers.
  Velour limits them to 20-255 columns and 10-100 rows. Without NAWS, the size is 80x24.
  A new NAWS during the session changes the size from the next screen.
- Type. The first TTYPE name in lower case becomes `term.type`. Without a name, it is `ansi`.
  The clients name themselves in different ways: PuTTY `XTERM`, Tera Term `xterm`,
  inetutils telnet `$TERM` in capitals, Windows telnet.exe `ANSI`.
- Enter. CR LF, CR NUL, a single CR and a single LF give one Enter.
  inetutils telnet sends CR NUL in character mode.

On output, Velour always sends CR LF and doubles each byte 255.

## Limits

- The data of one option is at most 256 bytes. When it is longer or does not end
  within 10 seconds, Velour closes the connection.
- There is no CHARSET option ([RFC 2066](https://www.rfc-editor.org/rfc/rfc2066)).
  Velour checks the charset with the terminal probe.
- A client that does not accept character mode is not supported:
  it sends the whole line on Enter.
- The telnet parser is own code, about 200 lines. The Go libraries do not give a server
  NAWS and TTYPE with bounded buffers.
