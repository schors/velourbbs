# Data store

This specification defines the data store: the data file that the BBS keeps between sessions,
its format, its read and write, and its migrations.
Related parts: [bbs-toml](bbs-toml.md), [session](session.md), [screen-render](screen-render.md),
[release-deploy](release-deploy.md) and [setup-guide](setup-guide.md).

## 1. Goal

The BBS keeps data between sessions: the oneliners and the last callers
([#81](https://github.com/schors/velourbbs/issues/81)).
The sysop reads the data file without a third-party tool.

## 2. Data file

- `[store] path` gives the data file ([bbs-toml](bbs-toml.md), section 3.4).
  A relative path is relative to the BBS directory.
- The store resolves the symlinks of the path once at start. It writes to the target,
  so a symlink stays a symlink.
- The file is UTF-8 text in the JSON Lines form: each line is one JSON object
  and ends with `\n`. An empty line has no effect.
- The first line is the header: `{"format":"velourbbs-store","version":<n>}`.
  `<n>` is the format version, an integer from 1.
- Each next line is one record: a JSON object with the string `kind` and the fields of the kind.
- The file is at most 16 MB.
- The store writes the records grouped by kind, in the order of the kinds in the code.
  In a kind, the records go from the oldest to the newest.
- The store creates the file with the mode 600.

Example (the kind `oneliner` is an example; [#81](https://github.com/schors/velourbbs/issues/81)
defines the kinds):

```
{"format":"velourbbs-store","version":2}
{"kind":"oneliner","time":"2026-11-19T09:00:00Z","name":"Ann","text":"привет <3"}
```

## 3. Kinds

- A kind is a type of record. It has a name, its fields, its checks and a limit:
  the maximum number of its records.
- The package `internal/store` defines each kind. It is the owner of the schema.
- A field name is in snake case. A time is RFC 3339 in UTC, with seconds.
- When a kind has more records than its limit, the store drops the oldest records.
- This specification defines no kind. The format version 1 has no kind.

## 4. Read

The store reads the file at the start of `velourbbs serve` and in `velourbbs check` (section 7).

- These are errors:
  - the file is not a regular file, or it is larger than 16 MB;
  - a line is not valid UTF-8, or it is not one JSON object;
  - the first line is not a valid header;
  - the version is newer than the version of the code: `version <n> is newer than this velourbbs (<m>)`;
  - an unknown kind, an unknown field, a missing field, a value of a wrong type,
    or a value that breaks a check of its kind.
- An error has the form of [screen-render](screen-render.md), section 3.1:
  `<path>:<line>:1: <message>`. `<path>` is the value of `[store] path`.
- The store reports all errors of the file, at most 100, as [screen-render](screen-render.md),
  section 3.1. An error of the header ends the read.
- More records than the limit of a kind are not an error (section 3).

## 5. Write

- The store keeps all records in memory. Each change writes the whole file:
  1. Write `<file>.tmp` in the same directory and sync it.
  2. Rename it over the file.
  3. Sync the directory. On Windows, skip this step.

  After a crash, the file has the records before the change or after it.
- One lock serialises the changes. The store changes the records in memory only after
  the write succeeds. After a write error, the records and the file stay as before.
  The caller gets the error.
- The write does not escape `<`, `>` and `&`, and keeps each non-ASCII character as UTF-8.
- The store does not read the file again after the start. A change of the file by the sysop
  while `velourbbs serve` runs goes away at the next write.
- Each data file has one `velourbbs serve`. The store does not lock the file.

## 6. Migrations

- The code has the current format version and an ordered list of migrations.
  Migration `<n>` changes the records of version `<n>` to version `<n>+1`.
- A new kind, a new field or a change of a field increases the version.
  The migration of a new kind changes no record. Then an older `velourbbs` refuses the file
  with the error of the version, not with an unknown kind.
- The read (section 4) checks the header first. Then it applies the migrations from the version
  of the file to the current version. Then it checks the records with the current kinds.
  An error gives the line of the record in the file.
- After a migration, an older `velourbbs` cannot read the file.
  To go back to an older release, restore a copy of the file of its version (section 8).

## 7. Commands

The data file is not a part of the model ([session](session.md), section 2):
it changes while the host runs. The commands read it after the load of the BBS.

- `velourbbs serve` reads the file at start:
  - When the file does not exist, the store starts with no records and writes the file.
    When its directory does not exist, it is an error.
  - After a migration, the store writes the file.
  - An error stops the start. The host exits with status 1.
- `velourbbs check` reads the file when it exists, with the migrations in memory.
  It never writes the file. A missing file is not an error.
- `velourbbs preview` uses a store in memory with no records. It does not read
  or write the file. Its changes go away at the exit.

## 8. Backup and edit

- Backup: stop the BBS, copy the file, start the BBS. There is no backup code.
- Edit: stop the BBS, edit the file, run `velourbbs check`, start the BBS.
- The home BBS keeps its data file outside the releases ([release-deploy](release-deploy.md),
  section 4.5).
- The setup guide keeps the data file in the state directory of systemd
  ([setup-guide](setup-guide.md), section 5).

## 9. Libraries

- The store uses only the standard library: `encoding/json`, `bufio` and `os`.
- Reason: the sysop reads and edits the file with a text editor
  ([#80](https://github.com/schors/velourbbs/issues/80)). The data is small:
  each kind has a limit, so the whole file fits in memory.
- Not used: SQLite (`modernc.org/sqlite`) and bbolt. Their files need a tool or a command
  to read. `modernc.org/sqlite` also adds 6 MB to the binary and 10 modules.

## 10. Code

| Package | Content |
|---|---|
| `internal/store` | The kinds, the read, the write and the migrations. |
| `internal/config` | The key `[store] path`. |
| `cmd/velourbbs` | The use of the store in `serve`, `check` and `preview` (section 7). |

## 11. Acceptance

- Tests of `internal/store` with test kinds and test migrations:
  - a valid file gives its records; an empty line has no effect;
  - each error of section 4 gives its line;
  - a kind over its limit drops the oldest records;
  - a file of an older version gives the records of the current version;
  - after a write error, the records and the file stay as before;
  - a written file reads back with the same records, with `<` and non-ASCII text unchanged;
  - a symlink to the file stays a symlink after a write.
- `velourbbs check` with a valid data file gives no output and does not change the file.
  With an invalid data file, it gives the error and exits with status 1.
- `velourbbs serve` with a missing file creates it with the header.
  With a missing directory, it exits with status 1.
- `go vet ./...`, `go test ./...` and `golangci-lint run ./...` pass.

## 12. Not in this specification

- The kinds of [#81](https://github.com/schors/velourbbs/issues/81): the oneliners
  and the last callers.
- The data of mods ([#12](https://github.com/schors/velourbbs/issues/12)), accounts.
- A lock against two hosts on one data file.
- A command that changes the data. The sysop edits the file (section 8).
