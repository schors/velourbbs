# 4. The commands

## `velourbbs check`

Run it after each change of a file and before each restart of the BBS.
It checks `bbs.toml`, the screens and the data file. It reads the data file only when the file exists.
It never writes the data file. When the files are correct, it gives no output.
An error in `bbs.toml` or `screens/screens.toml` stops the check before the screens.
Correct it, then run the check again.
Else it gives one line for each error or warning:

```
screens/main.bbt:7:8: item "S": when: unknown field "term.colours"
screens/about.bbt:4:1: warning: art "logo.ans" is 132 cells wide, max width is 100
```

The line gives the file, the line number, the column and the message.
A warning does not stop the BBS. An error stops it, and the command exits with status 1.
Full rules: [screen-render](../../../specs/screen-render.md), section 3.1.

## `velourbbs preview`

It shows a screen as text, without a telnet client:

```
velourbbs preview main
```

`--size` sets the size of the terminal, for example `--size 132x43`.
`--set term.colors=16` shows the screen for a terminal with 16 colours.
Full rules: [screen-render](../../../specs/screen-render.md), section 3.2.
Details: task 1 of the [BBT guide](../bbt/1-bbs-directory.md).

`--view` shows the screen in colour in your terminal, in place of the text dump:

```
velourbbs preview main --view
velourbbs preview main --view --set term.colors=16
```

It writes the pages one after another and does not wait for a key.
Each page starts with its line `page`.
Full rules: [screen-render](../../../specs/screen-render.md), section 3.7.

## `velourbbs preview --keys`

It presses the keys and shows each screen. Use it to test the keys of a screen:

```
velourbbs preview main --keys A,ESC,G
```

`--keys` and `--view` together show each page in colour.

Details: task 4 of the [BBT guide](../bbt/4-menu.md).

## `velourbbs fields`

It lists the fields of the BBS with their widths and sample values:

```
velourbbs fields
```

`velourbbs fields --sample` writes the sample sets of `velourbbs preview` to `sample.toml`.
Details: tasks 5 and 8 of the [BBT guide](../bbt/5-fields.md).

## `velourbbs art import`

It converts a file of an art editor to the form that the BBS reads:

```
velourbbs art import art/logo.ans screens/logo.ans
```

Details: task 9 of the [BBT guide](../bbt/9-art.md).

## `velourbbs serve`

It starts the telnet server. The server listens on port 2323.

- It does the checks of `velourbbs check` first. After an error, it exits with status 1.
- Then it reads the data file. When the file does not exist, it creates it.
  After an error of the data file, it exits with status 1.
- If it cannot listen on the address of `listen`, it exits with status 1.
- It writes the log to the standard error: a `listen` line with the address and its source,
  a line with the values of the terminal of each client, and a `close` line
  for each connection.
- It reads the files only when it starts. After a change of a file, restart it.
- Ctrl+C or SIGTERM stops it. It shows the screen `shutdown` to each open session,
  closes each connection after 5 s at most and exits with status 0.

Example of the log:

```
time=2026-11-16T09:42:48.905Z level=INFO msg=listen addr=[::]:2323 source=bbs.toml
time=2026-11-16T09:42:49.929Z level=INFO msg=probe addr=127.0.0.1 ttype=xterm-256color naws=80x24 cpr=1;2 type=xterm-256color cols=80 rows=24 charset=utf-8 colors=256 time=355.507µs
time=2026-11-16T09:42:51.231Z level=INFO msg=close addr=127.0.0.1 reason=logoff time=1.30238633s
```

Full rules: [telnet-term](../../../specs/telnet-term.md), section 10.

## Result

1. In `/srv/velourbbs`, run `velourbbs serve`. It writes the `listen` line.
2. Connect a telnet client to the host, port 2323, for example `telnet localhost 2323`.
   The client shows the screen `main`: the header "Lamp" and the menu `A R T G`.
   Press `G` to end the session.
3. Stop the server with Ctrl+C. Section 5 uses port 2323.
