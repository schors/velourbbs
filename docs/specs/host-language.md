# Host language

This specification records the language of the host and the reasons for it.
Related parts: [telnet-term](telnet-term.md), section 12, [screen-render](screen-render.md),
section 5, and [Lua mods](https://github.com/schors/velourbbs/issues/12).

## 1. Decision

The host is written in Go. The operator chose Go on 2026-11-12, after a study that compared
Go with Rust. Four reviewers read the study: a Go developer who later learned Rust,
a Rust developer, a terminal and protocol engineer and a security engineer.
Each of them recommended Go as the default.

## 2. Comparison

| Topic | Go | Rust | Result |
|---|---|---|---|
| Sessions | One goroutine for each session. Blocking reads and writes, with deadlines on `net.Conn`. | One tokio task for each session, with `.await`. A cancel drops the task. | Go is simpler. |
| SSH (not in the MVP) | `golang.org/x/crypto/ssh`, a part of the Go project. | `russh`, from a small community team. | Go. |
| HTTPS (not in the MVP) | `net/http` and `autocert`. | `axum`, `rustls` and `rustls-acme`. | Go needs fewer parts. |
| Telnet | Own code. | Own code. | Equal. |
| Terminal | Own art loader, cell model, update writer, key decoder and line input. | The same own code. `vte` and `vt100` are stronger parts. | Equal: the language does not decide it. |
| Build | A static binary without cgo. A fast build and a small dependency graph. | A slower build. `build.rs` and procedural macros run at build time. Luau needs a C++ toolchain. | Go. |
| Lua ([#12](https://github.com/schors/velourbbs/issues/12)) | gopher-lua: Lua 5.1 with `goto`, in pure Go. No memory limit for each state. | mlua with Luau or Lua 5.4. A memory limit for each state. | Rust is stronger. |
| Learning | — | The estimate of the study: 3–4 weeks to productive synchronous Rust, about 3 months to async server code. | Rust only when the goal is to learn Rust. |

## 3. Consequences

- Each session runs in its own goroutine. A panic in a session ends only that session
  ([telnet-term](telnet-term.md), section 10).
- The release builds with `CGO_ENABLED=0` ([release-deploy](release-deploy.md), section 3.2).
- A Lua state has no memory limit of its own. [#12](https://github.com/schors/velourbbs/issues/12)
  must give another limit.
