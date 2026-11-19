# Setup guide

This specification defines the setup guide: the guide to the setup and the run of a BBS
for the sysop.

## 1. Goal

The sysop sets up, checks and runs a BBS without the specifications.
systemd limits the memory, the threads and the open files of the host process.

## 2. Reader

- The reader is the sysop.
- The reader has a Linux host with systemd and root access.
- The reader can edit a text file and run a command in a terminal.
- The reader does not read the specifications. A part links to the sections
  of the specifications that hold its full rules.
- The reader does not write screens. The [BBT guide](bbt-guide.md) teaches the screens.

## 3. Files

| Path | Content |
|---|---|
| `docs/guide/en/setup/` | The guide in English, in parts. It is the source. |
| `docs/guide/ru/setup/` | The guide in Russian. It is a translation of the source. |
| `examples/deploy/velourbbs.service` | The systemd unit of section 4. |

- The rules of [bbt-guide](bbt-guide.md), section 3, for a specification change,
  for the two languages and for the parts (section 3.1) also apply to this guide.
- The parts of this guide: `README.md`, `1-program.md`, `2-bbs-directory.md`, `3-bbs-toml.md`,
  `4-commands.md`, `5-systemd.md`.
- The guide in `<lang>` uses `examples/lamp/<lang>/`. It uses the default port 2323 only.
- The guide keeps the data file `velourbbs.jsonl` in the BBS directory.
- The guide shows `examples/deploy/velourbbs.service` as a file block
  ([bbt-guide](bbt-guide.md), section 4).

## 4. The systemd unit

- The unit is a user unit: the host runs as the user of the sysop, and the BBS directory
  `/srv/velourbbs` belongs to this user.
- `TasksMax` is at least the number of CPUs + 32. Go starts threads for each CPU.
  When Go cannot start a thread, the host stops with all sessions.
- `LimitNOFILE` is at least `2 × max_conns + 16`. The host has at most `max_conns` connections
  and at most `max_conns` connections that close after output ([telnet-term](telnet-term.md), section 10).
- `MemoryMax`: when the sysop changes `max_conns`, the sysop reads the memory peak
  under load and sets `MemoryMax` with a margin.
- The host needs systemd 240 or later with cgroup v2, and the controllers `memory`
  and `pids` in the user manager.

## 5. Acceptance

- The test `TestGuideFileBlocks` in `cmd/velourbbs` also reads the parts of this guide.
- `systemd-analyze --user verify` gives no error for a copy of `examples/deploy/velourbbs.service`
  whose `ExecStart` has the path of a built `velourbbs`.
