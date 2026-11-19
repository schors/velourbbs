# 5. systemd

You need systemd 240 or later with cgroup v2.
Debian 12, Ubuntu 24.04 and RHEL 9 have them by default.

The BBS runs as your user, in your user manager of systemd: a user unit.
Root is necessary only once, for the linger.

## The unit

`examples/deploy/velourbbs.service`:
```
[Unit]
Description=Velour BBS

[Service]
Type=exec
ExecStart=/usr/local/bin/velourbbs serve
WorkingDirectory=/srv/velourbbs
Restart=on-failure
RestartSec=5
MemoryMax=128M
TasksMax=64
LimitNOFILE=1024

[Install]
WantedBy=default.target
```

| Line | Meaning |
|---|---|
| `Description=Velour BBS` | The name in `systemctl` and in the log. |
| `Type=exec` | `systemctl start` fails when systemd cannot start the program. |
| `ExecStart=/usr/local/bin/velourbbs serve` | The command. |
| `WorkingDirectory=/srv/velourbbs` | The BBS directory. |
| `Restart=on-failure` | systemd starts the BBS again after a failure. |
| `RestartSec=5` | systemd waits 5 s before it starts the BBS again. A user unit cannot wait for the network at boot. If the address of `listen` does not exist yet, `velourbbs serve` stops. systemd then tries again until the address exists. |
| `MemoryMax=128M` | The memory limit. It is for the defaults of `bbs.toml`. Lamp used 6.8 MB at start, 11.9 MB with 64 sessions and 15.8 MB with 300 sessions. |
| `TasksMax=64` | The thread limit. |
| `LimitNOFILE=1024` | The open file limit. |
| `WantedBy=default.target` | `systemctl --user enable` starts the BBS when your user manager starts. |

## Install and start

Turn on the linger of your user. Then your user manager starts at boot,
also when you are not logged in:

```
sudo loginctl enable-linger "$USER"
```

In the source repository, run:

```
mkdir -p ~/.config/systemd/user
cp examples/deploy/velourbbs.service ~/.config/systemd/user/
systemctl --user daemon-reload
systemctl --user enable --now velourbbs
```

`enable --now` starts the BBS and starts it when your user manager starts.

The BBS writes the log to the standard error. systemd writes it to the journal.
Read the log:

```
journalctl --user -u velourbbs
```

## Changes

After a change of a file in the BBS directory:

```
cd /srv/velourbbs
velourbbs check && systemctl --user restart velourbbs
```

After a change of `~/.config/systemd/user/velourbbs.service`:

```
systemctl --user daemon-reload
systemctl --user restart velourbbs
```

## The data file

The BBS writes the data file `/srv/velourbbs/velourbbs.jsonl` while it runs.
The file is text: one JSON object on each line. Stop the BBS before you copy or edit the file.

Make a backup:

```
systemctl --user stop velourbbs
cp /srv/velourbbs/velourbbs.jsonl ~/velourbbs.jsonl.bak
systemctl --user start velourbbs
```

To restore it, stop the BBS and copy the backup back to `/srv/velourbbs/velourbbs.jsonl`.

A new release of `velourbbs` can change the format of the file. Then an older release
cannot read the file. Make a backup before you install a new release.

To edit the file, stop the BBS:

```
systemctl --user stop velourbbs
```

Edit `/srv/velourbbs/velourbbs.jsonl` in a text editor. Then run:

```
cd /srv/velourbbs
velourbbs check && systemctl --user start velourbbs
```

Full rules: [data-store](../../../specs/data-store.md), section 8.

## The limits

- `TasksMax` is at least the number of CPUs + 32. `nproc` gives the number of CPUs.
  Go starts threads for each CPU. When Go cannot start a thread, the BBS stops
  with all sessions. `TasksMax=64` is sufficient for 32 CPUs or fewer.
- `LimitNOFILE` is at least `2 × max_conns + 16`. The BBS has at most `max_conns`
  connections, and at most `max_conns` connections that close after the output.
  `LimitNOFILE=1024` is sufficient for `max_conns = 504` or less.
- `MemoryMax`: when you change `max_conns`, read the memory peak under load.
  Then set `MemoryMax` with a margin.

These limits do not replace the connection limits of `bbs.toml`.
The limits of `bbs.toml` count the connections. The limits of systemd limit the process.

`systemctl --user status velourbbs` shows the threads and the memory. Example:

```
      Tasks: 6 (limit: 64)
     Memory: 11.9M (max: 128M, available: 116.1M, peak: 15.8M)
```

`Tasks` gives the threads and `TasksMax`. `Memory` gives the memory now,
`MemoryMax` and the peak. systemd 255 or later shows the peak.
With an older systemd, read `Memory` while the BBS is under load.

Full rules: [setup-guide](../../../specs/setup-guide.md), section 4,
and [telnet-term](../../../specs/telnet-term.md), section 10.

## Result

- `systemctl --user is-active velourbbs` gives `active`.
- `journalctl --user -u velourbbs` shows the `listen` line.
- A telnet client on port 2323 shows the screen `main`.

`systemctl start` and `systemctl restart` succeed also when `velourbbs serve`
stops at once. So check the state and the log after each start.
