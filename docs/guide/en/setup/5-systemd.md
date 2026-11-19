# 5. systemd

You need systemd 240 or later with cgroup v2.
Debian 12, Ubuntu 24.04 and RHEL 9 have them by default.

## The unit

`examples/deploy/velourbbs.service`:
```
[Unit]
Description=Velour BBS
Wants=network-online.target
After=network-online.target

[Service]
Type=exec
ExecStart=/usr/local/bin/velourbbs serve
WorkingDirectory=/srv/velourbbs
Restart=on-failure
User=velourbbs
ProtectSystem=strict
ProtectHome=read-only
StateDirectory=velourbbs
MemoryMax=128M
TasksMax=64
LimitNOFILE=1024

[Install]
WantedBy=multi-user.target
```

| Line | Meaning |
|---|---|
| `Description=Velour BBS` | The name in `systemctl` and in the log. |
| `Wants=network-online.target`, `After=network-online.target` | The address of `listen` exists before the start. Else `velourbbs serve` cannot listen and stops. |
| `Type=exec` | `systemctl start` fails when systemd cannot start the program. |
| `ExecStart=/usr/local/bin/velourbbs serve` | The command. |
| `WorkingDirectory=/srv/velourbbs` | The BBS directory. |
| `Restart=on-failure` | systemd starts the BBS again after a failure. |
| `User=velourbbs` | The BBS runs as the user of section 2, without privileges. The data file belongs to this user. |
| `ProtectSystem=strict`, `ProtectHome=read-only` | The system and `/home` are read-only for the BBS. The BBS only reads the BBS directory, so this does not stop it. |
| `StateDirectory=velourbbs` | `/var/lib/velourbbs/` stays writable for the BBS. The data file is there. |
| `MemoryMax=128M` | The memory limit. It is for the defaults of `bbs.toml`. Lamp used 6.8 MB at start, 11.9 MB with 64 sessions and 15.8 MB with 300 sessions. |
| `TasksMax=64` | The thread limit. |
| `LimitNOFILE=1024` | The open file limit. |
| `WantedBy=multi-user.target` | `systemctl enable` starts the BBS at boot. |

## Install and start

In the source repository, run:

```
sudo cp examples/deploy/velourbbs.service /etc/systemd/system/
sudo systemctl enable --now velourbbs
```

`enable --now` loads the unit, starts the BBS and starts it at each boot.

The BBS writes the log to the standard error. systemd writes it to the journal.
Read the log:

```
sudo journalctl -u velourbbs
```

## Changes

After a change of a file in the BBS directory:

```
cd /srv/velourbbs
sudo chmod -R a+rX /srv/velourbbs
sudo -u velourbbs /usr/local/bin/velourbbs check && sudo systemctl restart velourbbs
```

After a change of `/etc/systemd/system/velourbbs.service`:

```
sudo systemctl daemon-reload
sudo systemctl restart velourbbs
```

## The data file

The BBS writes the data file `/var/lib/velourbbs/velourbbs.jsonl` while it runs.
The file is text: one JSON object on each line. Stop the BBS before you copy or edit the file.
Use `sudo -u velourbbs cp` and `sudoedit`: the file then keeps the owner `velourbbs`.

Make a backup:

```
sudo systemctl stop velourbbs
cd /var/lib/velourbbs
sudo -u velourbbs cp velourbbs.jsonl velourbbs.jsonl.bak
sudo systemctl start velourbbs
```

To restore it, stop the BBS and copy `velourbbs.jsonl.bak` back to `velourbbs.jsonl` the same way.

A new release of `velourbbs` can change the format of the file. Then an older release
cannot read the file. Make a backup before you install a new release.

Edit the file:

```
sudo systemctl stop velourbbs
sudoedit /var/lib/velourbbs/velourbbs.jsonl
cd /srv/velourbbs
sudo -u velourbbs /usr/local/bin/velourbbs check && sudo systemctl start velourbbs
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

`systemctl status velourbbs` shows the threads and the memory. Example:

```
      Tasks: 6 (limit: 64)
     Memory: 11.9M (max: 128M, available: 116.1M, peak: 15.8M)
```

`Tasks` gives the threads and `TasksMax`. `Memory` gives the memory now,
`MemoryMax` and the peak. systemd 255 or later shows the peak.
With an older systemd, read `Memory` while the BBS is under load.

Full rules: [setup-guide](../../../specs/setup-guide.md), section 5,
and [telnet-term](../../../specs/telnet-term.md), section 10.

## Result

- `systemctl is-active velourbbs` gives `active`.
- `sudo journalctl -u velourbbs` shows the `listen` line.
- A telnet client on port 2323 shows the screen `main`.

`systemctl start` and `systemctl restart` succeed also when `velourbbs serve`
stops at once. So check the state and the log after each start.
