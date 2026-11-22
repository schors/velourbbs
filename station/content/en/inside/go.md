# Go and libraries

Velour is written in Go. It is one program without cgo: it takes guests over telnet,
draws the screens, keeps the data and checks the files of the SysOp.
The source is open: [github.com/schors/velourbbs](https://github.com/schors/velourbbs).

## Why Go

The language was chosen on 12 November 2026. Go was compared with Rust, and four reviewers
chose Go: a Go developer, a Rust developer, a terminal and protocol engineer
and a security engineer. The comparison is in `docs/specs/host-language.md`.

- Sessions. Each session is its own goroutine. The code reads and writes with blocking calls,
  and the deadlines of `net.Conn` limit the time. In Rust, a session is a tokio task with `.await`.
- SSH for later. Go has `golang.org/x/crypto/ssh` from the Go project.
  Rust has `russh` from a small team.
- Build. Go builds a static program without cgo, fast and with a small dependency graph.
  Rust builds slower, and the `build.rs` files and procedural macros of the dependencies run at build time.
- Terminal. Here the languages are equal: the art parser, the cell model, the update output
  and the key decoder are own code in both.

## Where Rust is stronger

- Lua for mods. In Rust, `mlua` sets a memory limit for each Lua state.
  `gopher-lua` for Go has no such limit. Lua mods are only a plan now.
- Ready parts for a terminal parser: `vte` and `vt100`.

## Libraries

Apart from the Go standard library, Velour has three direct dependencies:

- `go-toml/v2` reads the settings and the screens. It was compared with `BurntSushi/toml`.
  `go-toml` won: it gives the line and the column of an unknown key,
  so `velourbbs check` shows the place of the error.
- `displaywidth` splits text into graphemes and gives the width of each by Unicode 17.
  It was compared with `uniseg` (Unicode 15) and `go-runewidth` (it measures runes, not graphemes).
- `goldmark/v2` reads Markdown: CommonMark 0.31.2 and the GFM tables.
  It has no dependencies of its own. Version 2 is young, and a patch can break the API,
  so `go.mod` pins the version.

`log/slog` writes the log, `flag` parses the commands, `encoding/json` encodes the data.

## Own code in place of libraries

- Telnet. The Go libraries do not give a server NAWS and TTYPE with bounded buffers.
  The own parser has about 200 lines.
- PROXY protocol. `go-proxyproto` has an API of version 0.x, and its 2026 releases
  changed the default behaviour. The own receiver of versions 1 and 2 has about 200 lines.
- Terminal output. `tcell` draws the whole screen as an application, and Velour shows
  pages with a "Next" row. Also, `tcell` does not query the terminal at connection.
- ANSI art. The server needs only SGR styles, and the import about 15 commands.
  The full terminal parser `charmbracelet/x/ansi` is a large dependency.
- Screen conditions. Their grammar has 7 rules. `expr` would add arithmetic
  and functions that the format does not have.
- Socket activation. The code reads two environment variables and calls `net.FileListener`:
  about 30 lines in place of `go-systemd`.

## Limits

- Own code means own tests and own maintenance. Each decision is in the section
  "Libraries" of its specification.
- The Go standard library and `golang.org/x/text` have no grapheme segmentation.
  So the width comes from the third-party `displaywidth`.
