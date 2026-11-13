# Idea: log-format

Backlog. It changes the logger of `bbs serve`.

## Goal

Write logs with `slog` in a format that fits the output:
JSON for machines, readable text for humans.

## Decisions

- There are two formats: JSON and readable text.
  Readable text is colored or not colored.
- By default, the format depends on stderr.
  If stderr is a terminal, the log is readable text. If not, the log is JSON.
- Readable text is colored only if the terminal supports color.
- A CLI flag sets any format: JSON, colored text or text with no color.
- Detect "no color" in all usual ways, for example `NO_COLOR` and `TERM=dumb`.
- Readable text renders each kind of data separately:
  - the color depends on the level;
  - values render by type: `error`, `time.Duration`, address, groups;
  - known events, for example `probe` and `close`, have their own view.
- A CLI flag sets the log level.
- Log output to a file is not in scope.

## Open questions

- The default level.
- The names and values of the flags.
- The full list of "no color" signs, and a way to force color.
- The view of the known events.
- The relation to [content-trust](content-trust.md): where to remove
  control characters from log lines.
