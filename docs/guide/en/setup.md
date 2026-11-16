# Setup guide

This guide teaches you to set up, check and run a Velour BBS on a Linux host.
It is for the sysop.
You need a Linux host with systemd, root access, a text editor and a telnet client.

Each section ends with a check of the result.
A section refers to the specifications that hold its full rules.
The [BBT guide](bbt.md) teaches you to write the screens.

## 1. The program

You need Go 1.27.3 or later. In the source repository of Velour BBS, run:

```
go build ./cmd/velourbbs
sudo install velourbbs /usr/local/bin/
```

The first command writes the file `velourbbs`. The second command copies it
to `/usr/local/bin/velourbbs`.

### Result

`velourbbs` without arguments prints the usage:

```
usage:
  velourbbs art import <input> <output>
  velourbbs check
  velourbbs preview <name> [--size <cols>x<rows>] [--set <path>=<value>]... [--keys <keys>]
  velourbbs serve
```

## 2. The BBS directory

A "BBS directory" holds the files of one BBS. This guide uses `/srv/velourbbs`.
Copy the example BBS "Lamp" to it. In the source repository, run:

```
sudo cp -r examples/lamp/en /srv/velourbbs
sudo chmod -R a+rX /srv/velourbbs
```

`/srv/velourbbs` must not exist before the copy.
`chmod` makes the files readable and the directories readable and searchable for each user.
The BBS runs as a user without privileges (section 5). This user must read the files.
The files belong to root. Edit them as root, for example with `sudoedit`.

The files of Lamp:

| File | Content |
|---|---|
| `bbs.toml` | The settings of the host (section 3). In Lamp, it is empty: the defaults. |
| `screens/screens.toml` | The settings of the screens: the width, the styles and other values. |
| `screens/_head.bbt` | The header. Each screen includes it. |
| `screens/main.bbt` | The first screen: the menu. |
| `screens/about.bbt` | The text about the BBS. |
| `screens/term.bbt` | The values of the terminal. |
| `screens/goodbye.bbt` | The screen after the end of the session. |

To change the screens, read the [BBT guide](bbt.md).

Run each `velourbbs` command in the BBS directory.
Run `velourbbs check` as your user, not as root: root can read a file
that the BBS user cannot read.
Full rules: [bbs-toml](../../specs/bbs-toml.md), section 2.

### Result

```
cd /srv/velourbbs
velourbbs check
```

The command gives no output.

## 3. `bbs.toml`

`bbs.toml` holds the table `[telnet]`: the settings of the telnet server.
Each key is optional. A key that is not set has its default. An empty file gives the defaults.

### The address

`listen` is the TCP address of the server: `<host>:<port>`. The default is `":2323"`.

- An empty host: the server listens on all addresses of the host: `":2323"`.
- One address: the server listens only on this address, for example `"192.0.2.1:2323"`.
  Use it when the host has more than one address and the BBS must use only one.

`velourbbs check` checks only the form of `listen` and the port.
`velourbbs serve` checks the address when it starts. If the host does not have
the address, `velourbbs serve` stops.

### The connection limits

| Key | Default | What it does | When to change it |
|---|---|---|---|
| `max_conns` | 64 | The maximum number of connections. | More users connect at the same time. Then also change the limits of section 5. |
| `max_conns_per_addr` | 4 | The maximum number of connections from one address. | Many users share one address, for example behind NAT. |
| `max_connects_per_min` | 10 | The maximum number of connects from one address in 60 s. | A client connects again many times. |
| `allow` | `[]` | Addresses and prefixes without the address limits. | One address must have more connections, for example a known NAT. |

- "One address" is one IPv4 address, or one IPv6 /64 prefix.
- `max_conns` applies also to the addresses in `allow`.
- A connection over a limit gets the line `Too many connections. Try again later.`
  Then the server closes it.

Example:

```toml
[telnet]
allow = ["192.0.2.10", "2001:db8::/48"]
```

### The PROXY protocol

`proxy` holds the addresses and prefixes of the proxies that send the PROXY protocol header.
The default is `[]`: the server reads no header.
Set it when a proxy, for example HAProxy, sends the connections to the BBS.
Then the log and the limits use the address of the client, not the address of the proxy.

- The server reads the header only from the addresses in `proxy`.
- A connection from an address in `proxy` without a valid header closes without output.
  So, if `proxy` holds `127.0.0.1`, a telnet client on the same host cannot connect
  to `127.0.0.1`.
- The proxy must send the header at the connect. AWS NLB waits for the data of the client
  by default. Then it does not work.

Example:

```toml
[telnet]
proxy = ["127.0.0.1"]
```

### The timeouts

| Key | Default | What it does | When to change it |
|---|---|---|---|
| `idle_timeout` | 600 | Seconds without a key until the server disconnects. | Users read long screens. |
| `probe_timeout` | 5 | Seconds from the connect until the server knows the terminal. | Clients have a slow link. |
| `output_timeout` | 60 | Seconds until a write must complete. Else the server disconnects. | Clients have a slow link. |

### An example file

This file has the defaults. It sets each key except `allow` and `proxy`:

```toml
[telnet]
listen = ":2323"
max_conns = 64
max_conns_per_addr = 4
max_connects_per_min = 10
idle_timeout = 600
probe_timeout = 5
output_timeout = 60
```

Write it to `/srv/velourbbs/bbs.toml`. Then change the values that you need.
After each change, run `sudo chmod -R a+rX /srv/velourbbs` and `velourbbs check`.

Full rules: [bbs-toml](../../specs/bbs-toml.md), section 3.1.

### Result

`velourbbs check` gives no output.

## 4. The commands

### `velourbbs check`

Run it after each change of a file and before each restart of the BBS.
It checks `bbs.toml` and the screens. When the files are correct, it gives no output.
Else it gives one line for each error or warning:

```
screens/main.bbt:2:1: item "S": when: unknown field "term.colours"
screens/about.bbt:4:1: warning: art "logo.ans" is 132 cells wide, max width is 100
```

The line gives the file, the line number, the column and the message.
A warning does not stop the BBS. An error stops it, and the command exits with status 1.
Full rules: [screen-render](../../specs/screen-render.md), section 3.1.

### `velourbbs preview`

It shows a screen as text, without a telnet client:

```
velourbbs preview main
```

`--size` sets the size of the terminal, for example `--size 132x43`.
`--set term.colors=16` shows the screen for a terminal with 16 colours.
Full rules: [screen-render](../../specs/screen-render.md), section 3.2.
Details: task 1 of the [BBT guide](bbt.md).

### `velourbbs preview --keys`

It presses the keys and shows each screen. Use it to test the keys of a screen:

```
velourbbs preview main --keys A,ESC,G
```

Details: task 4 of the [BBT guide](bbt.md).

### `velourbbs art import`

It converts a file of an art editor to the form that the BBS reads:

```
sudo velourbbs art import art/logo.ans screens/logo.ans
```

Details: task 7 of the [BBT guide](bbt.md).

### `velourbbs serve`

It starts the telnet server. The server listens on port 2323.

- It does the checks of `velourbbs check` first. After an error, it exits with status 1.
- If it cannot listen on the address of `listen`, it exits with status 1.
- It writes the log to the standard error: a `listen` line with the address,
  a line with the values of the terminal of each client, and a `close` line
  for each connection.
- It reads the files only when it starts. After a change of a file, restart it.
- Ctrl+C or SIGTERM stops it. It closes the connections and exits with status 0.

Example of the log:

```
time=2026-11-16T09:42:48.905Z level=INFO msg=listen addr=[::]:2323
time=2026-11-16T09:42:49.929Z level=INFO msg=probe addr=127.0.0.1 ttype=xterm-256color naws=80x24 cpr=1;2 type=xterm-256color cols=80 rows=24 charset=utf-8 colors=256 time=355.507µs
time=2026-11-16T09:42:51.231Z level=INFO msg=close addr=127.0.0.1 reason=logoff time=1.30238633s
```

Full rules: [telnet-term](../../specs/telnet-term.md), section 10.

### Result

1. In `/srv/velourbbs`, run `velourbbs serve`. It writes the `listen` line.
2. Connect a telnet client to the host, port 2323, for example `telnet localhost 2323`.
   The client shows the screen `main`: the header "Lamp" and the menu `A T G`.
   Press `G` to end the session.
3. Stop the server with Ctrl+C. Section 5 uses port 2323.

## 5. systemd

You need systemd 240 or later with cgroup v2.
Debian 12, Ubuntu 24.04 and RHEL 9 have them by default.

### The unit

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
DynamicUser=yes
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
| `DynamicUser=yes` | The BBS runs as a new user without privileges. The system is read-only for it, and `/home` is read-only. The BBS only reads the BBS directory, so this does not stop it. |
| `MemoryMax=128M` | The memory limit. It is for the defaults of `bbs.toml`. Lamp used 6.8 MB at start, 11.9 MB with 64 sessions and 15.8 MB with 300 sessions. |
| `TasksMax=64` | The thread limit. |
| `LimitNOFILE=1024` | The open file limit. |
| `WantedBy=multi-user.target` | `systemctl enable` starts the BBS at boot. |

### Install and start

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

### Changes

After a change of a file in the BBS directory:

```
cd /srv/velourbbs
sudo chmod -R a+rX /srv/velourbbs
velourbbs check && sudo systemctl restart velourbbs
```

After a change of `/etc/systemd/system/velourbbs.service`:

```
sudo systemctl daemon-reload
sudo systemctl restart velourbbs
```

### The limits

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

Full rules: [setup-guide](../../specs/setup-guide.md), section 5,
and [telnet-term](../../specs/telnet-term.md), section 10.

### Result

- `systemctl is-active velourbbs` gives `active`.
- `sudo journalctl -u velourbbs` shows the `listen` line.
- A telnet client on port 2323 shows the screen `main`.

`systemctl start` and `systemctl restart` succeed also when `velourbbs serve`
stops at once. So check the state and the log after each start.
