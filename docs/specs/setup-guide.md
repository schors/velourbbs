# Setup guide

This specification defines the setup guide: the guide to the setup and the run of a BBS
for the sysop.
Related parts: [bbs-toml](bbs-toml.md), [telnet-term](telnet-term.md),
[screen-render](screen-render.md), [bbt-guide](bbt-guide.md), [example-style](example-style.md)
and [session-fields](session-fields.md).

## 1. Goal

The sysop sets up, checks and runs a BBS without the specifications.
systemd limits the memory, the threads and the open files of the host process.

## 2. Reader

- The reader is the sysop.
- The reader has a Linux host with systemd and root access.
- The reader can edit a text file and run a command in a terminal.
- The reader does not read the specifications. A section refers to the sections
  of the specifications that hold its full rules.
- The reader does not write screens. The [BBT guide](bbt-guide.md) teaches the screens.

## 3. Files

| Path | Content |
|---|---|
| `docs/guide/en/setup/` | The guide in English, in parts. It is the source. |
| `docs/guide/ru/setup/` | The guide in Russian. It is a translation of the source. |
| `examples/deploy/velourbbs.service` | The systemd unit of section 5. |

- The rules of [bbt-guide](bbt-guide.md), section 3, for a specification change,
  for the two languages and for the parts (section 3.1) also apply to this guide.
  Part `<n>` is section `<n>` of section 4.
- The parts of this guide: `README.md`, `1-program.md`, `2-bbs-directory.md`, `3-bbs-toml.md`,
  `4-commands.md`, `5-systemd.md`.
- The English guide uses `examples/lamp/en/`. The Russian guide uses `examples/lamp/ru/`.
- The guide shows `examples/deploy/velourbbs.service` as a file block
  ([bbt-guide](bbt-guide.md), section 4.1).

## 4. Sections

The guide has these sections in this order. Each section ends with a check of the result.

| Section | Content | Check |
|---|---|---|
| 1. The program | Build `velourbbs` with Go, at least the version of `go.mod`, in a git checkout of a release tag without changes: then `bbs.version` shows the tag ([session-fields](session-fields.md), section 3.4). Install it to `/usr/local/bin/velourbbs`. | `velourbbs` without arguments prints the usage. |
| 2. The BBS directory | Copy the example BBS Lamp to `/srv/velourbbs`. Its files: [example-style](example-style.md), section 3. Make the files readable and the directories readable and searchable for each user: `chmod -R a+rX /srv/velourbbs`. Run each `velourbbs` command in this directory ([bbs-toml](bbs-toml.md), section 2). For the screens, the guide refers to the BBT guide. | `velourbbs check` gives no output. |
| 3. `bbs.toml` | The key `[bbs] name`. The keys of `[telnet]` in groups: the address (`listen`), the connection limits (`max_conns`, `max_conns_per_addr`, `max_connects_per_min`, `allow`), the PROXY protocol (`proxy`) and the timeouts. For each group: what the keys do and when the sysop changes them. For `listen`, the guide shows only the host part: empty for all addresses, or one address. An example file without `proxy` and without `allow`. Full rules: [bbs-toml](bbs-toml.md), sections 3.1 and 3.2. | `velourbbs check` gives no output. |
| 4. The commands | One short part for each command, with the task of the sysop. `velourbbs check`: after each change, before a restart ([screen-render](screen-render.md), section 3.1). `velourbbs preview` and `velourbbs preview --keys`: look at a screen and test its keys without a telnet client ([screen-render](screen-render.md), section 3.2). `velourbbs art import`: convert art from an art editor. `velourbbs serve`: the log, the exit statuses, the stop, and the read of the files only at start ([telnet-term](telnet-term.md), section 10). For the details of `preview`, `preview --keys` and `art import`, the guide refers to tasks 1, 4 and 8 of the BBT guide. | `velourbbs serve` writes the `listen` line. A telnet client on port 2323 shows the screen `main`. |
| 5. systemd | Install the unit to `/etc/systemd/system/velourbbs.service`. Start and enable it. Read the log with `journalctl -u velourbbs`. After a change of a file of the BBS directory: `chmod -R a+rX /srv/velourbbs`, `velourbbs check`, then `systemctl restart velourbbs`. After a change of the unit: `systemctl daemon-reload`, then `systemctl restart velourbbs`. The meaning of each line of the unit. The rules of section 5 for the limits. How to read the memory and its peak in `systemctl status velourbbs`. systemd 255 or later shows the peak. With an older systemd, the sysop reads the memory under load. | `systemctl is-active velourbbs` gives `active`. The log has the `listen` line. A telnet client on port 2323 shows the screen `main`. |

- The server listens on the default port 2323. The guide does not show other ports.
- With `Type=exec`, `systemctl start` and `systemctl restart` succeed also when
  `velourbbs serve` exits at once. So the check of section 5 reads the state and the log.

## 5. The systemd unit

The unit has these lines:

| Section | Line | Reason |
|---|---|---|
| `[Unit]` | `Description=Velour BBS` | The name in `systemctl` and in the log. |
| `[Unit]` | `Wants=network-online.target`, `After=network-online.target` | The address of `listen` exists before the start. Else `listen` fails and the host exits ([telnet-term](telnet-term.md), section 10). |
| `[Service]` | `Type=exec` | `systemctl start` fails when systemd cannot start the program. |
| `[Service]` | `ExecStart=/usr/local/bin/velourbbs serve` | The command. |
| `[Service]` | `WorkingDirectory=/srv/velourbbs` | The BBS directory ([bbs-toml](bbs-toml.md), section 2). |
| `[Service]` | `Restart=on-failure` | systemd starts the host again after a failure. |
| `[Service]` | `DynamicUser=yes` | The host runs as a new user without privileges. The system is read-only for it, and `/home` is read-only. The host only reads the BBS directory, so this does not stop it. |
| `[Service]` | `MemoryMax=128M` | The memory limit for the defaults of [bbs-toml](bbs-toml.md). The host with Lamp used 6.8 MB at start, 11.9 MB with 64 sessions and 15.8 MB with 300 sessions. |
| `[Service]` | `TasksMax=64` | The thread limit. |
| `[Service]` | `LimitNOFILE=1024` | The open file limit. |
| `[Install]` | `WantedBy=multi-user.target` | `systemctl enable` starts the host at boot. |

- The log goes to the journal: the host writes it to the standard error
  ([telnet-term](telnet-term.md), section 10).
- `TasksMax` is at least the number of CPUs + 32. Go starts threads for each CPU.
  When Go cannot start a thread, the host stops with all sessions.
- `LimitNOFILE` is at least `2 × max_conns + 16`. The host has at most `max_conns` connections
  and at most `max_conns` connections that close after output ([telnet-term](telnet-term.md), section 10).
- `MemoryMax`: when the sysop changes `max_conns`, the sysop reads the memory peak
  under load and sets `MemoryMax` with a margin.
- These limits do not repeat a host limit of [telnet-term](telnet-term.md), section 9.2.
  The host limits count connections. The systemd limits limit the process.

### 5.1. Target implementations

- systemd 240 or later with cgroup v2. `Type=exec` needs systemd 240.
  `MemoryMax` needs cgroup v2.
- Debian 12, Ubuntu 24.04 and RHEL 9 have them by default.

## 6. Acceptance

- The test `TestGuideFileBlocks` in `cmd/velourbbs` also reads the parts
  of `docs/guide/en/setup/` and `docs/guide/ru/setup/`.
- `examples/deploy/velourbbs.service` has the lines of section 5 and no other lines.
- `systemd-analyze verify` gives no error for a copy of `examples/deploy/velourbbs.service`
  whose `ExecStart` has the path of a built `velourbbs`.
- Each section of section 4 is in both languages, in the same order.

## 7. Not in this guide

- nftables and `tc`: [deploy-limits](https://github.com/schors/velourbbs/issues/6).
- A port other than 2323.
