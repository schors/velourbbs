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
- The file is UTF-8 text in the JSON Lines form: each line is one JSON object
  and ends with `\n`. The loader removes a BOM.
- Line 1 is the header: `{"format":"velourbbs-store","version":<n>}`.
  `<n>` is the format version, an integer from 1.
- Each next line is one record: a JSON object with the string `kind` and the fields of the kind.
  An empty line has no effect. A line with only spaces, tabs and `\r` is an empty line.
- The order of the records of a kind is their order in the file, from the oldest to the newest.
  The records of different kinds can be mixed. The write groups them by kind,
  in the order of the kinds in the code.
- The file is at most 16 MB.

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
- A field name is in snake case. A time is RFC 3339 in UTC, with seconds and no fraction.
- When a kind has more records than its limit, the store drops the oldest records.
- This specification defines no kind. The format version 1 has no kind.

## 4. Checks

The read (section 7) and each change (section 5) use these checks.

- These are errors:
  - the file is not a regular file, or it is larger than 16 MB. A symlink is not a regular file:
    the rename of a write replaces the link. A symlink to the directory of the file works;
  - a line is not valid UTF-8, or it is not one JSON object;
  - line 1 is not a valid header;
  - the version is newer than the version of the code: `version <n> is newer than this velourbbs (<m>)`;
  - an unknown kind; an unknown, missing or duplicate field; a value of a wrong type,
    also `null`; a value that breaks a check of its kind.
    A field name that differs from a known name only in case is an unknown field.
- A read error has the form of [screen-render](screen-render.md), section 3.1, with its limit:
  `<path>:<line>:1: <message>`. `<path>` is the value of `[store] path`.
  An error of the header ends the read.
- More records than the limit of a kind are not an error (section 3).

## 5. Write

- The store keeps all records in memory. Each change writes the whole file:
  1. Check the new records (section 4). An error ends the change.
  2. Remove an old `<file>.tmp`. Create the new `<file>.tmp` in the same directory
     with the mode 600. Write the records and sync the file.
  3. Rename it over the file. The rename is the commit point.
  4. Sync the directory. On Windows, skip this step.
- An error before the rename removes `<file>.tmp`. The records and the file stay as before.
  The caller gets the error.
- After the rename, the store keeps the new records. A sync error of the directory goes to the log.
- After a crash on Linux or macOS, the file has the records before the change or after it.
  On Windows, Go does not make the rename atomic, and the rename fails while another program
  has the file open. Then the change gets a write error.
- One lock serialises the changes.
- The write does not escape `<`, `>` and `&`. It keeps each non-ASCII character as UTF-8,
  except U+2028 and U+2029: it escapes them.
- The store does not read the file again after the start. A change of the file by the sysop
  while `velourbbs serve` runs goes away at the next write.
- Each data file has one `velourbbs serve`. The store does not lock the file.

## 6. Migrations

- The code has the current format version and an ordered list of migrations.
  Migration `<n>` changes the records of version `<n>` to version `<n>+1`.
- Each change of a kind increases the version: a new or removed kind, a new, removed
  or changed field, and a stricter check. The migration fixes or drops each record that breaks
  the new rules. The migration of a new kind changes no record.
  Then an older `velourbbs` refuses the file with the error of the version.
- The read checks the header first. Then it applies the migrations from the version
  of the file to the current version. Then it checks the records with the current kinds.
  An error gives the line of the record in the file.
- After a migration, an older `velourbbs` cannot read the file.
  Before the deploy of a release that increases the version, make the backup of section 8.
  To go back to an older release, restore this backup.

## 7. Commands

The data file is not a part of the model ([session](session.md), section 2):
it changes while the host runs.

- `velourbbs serve` reads the file after the load of the BBS:
  - When the file does not exist, the store starts with no records and writes the file.
    When its directory does not exist, it is an error.
  - After a migration, the store writes the file.
  - An error stops the start. The host exits with status 1.
- `velourbbs check` reads the file when it exists and `bbs.toml` has no error,
  also after errors in other files. It applies the migrations in memory.
  It never writes the file. A missing file is not an error.
- `velourbbs preview` uses a store in memory with no records. It does not read
  or write the file. Its changes go away at the exit.

## 8. Backup and edit

- Backup: stop the BBS, copy the file, start the BBS. There is no backup code.
- Edit: stop the BBS, edit the file, run `velourbbs check`, start the BBS.
- The stop of the home BBS stops `velourbbs.socket` and `velourbbs.service`
  ([release-deploy](release-deploy.md), section 5). Else a connect starts the host again.
- The home BBS keeps its data file outside the releases ([release-deploy](release-deploy.md),
  section 4.5).
- The setup guide keeps the data file in the BBS directory
  ([setup-guide](setup-guide.md), section 3).

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
| `internal/store` | The kinds, the checks, the read, the write and the migrations. |
| `internal/config` | The key `[store] path`. |
| `cmd/velourbbs` | The use of the store in `serve`, `check` and `preview` (section 7). `run` also gets the BBS directory as an OS path for the store. Tests use a temporary directory. |

## 11. Acceptance

- Tests of `internal/store` with test kinds and test migrations:
  - a valid file gives its records; a BOM and an empty line have no effect;
  - each error of section 4 gives its line;
  - a kind over its limit drops the oldest records;
  - a file of an older version gives the records of the current version;
  - a change that breaks a check gives an error; the records and the file stay as before;
  - after a write error before the rename, the records and the file stay as before,
    and `<file>.tmp` does not exist;
  - a written file reads back with the same records, with `<` and non-ASCII text unchanged.
- `velourbbs check` with a valid data file gives no output and does not change the file.
  With an invalid data file, it gives the error and exits with status 1.
- `velourbbs serve` with a missing file creates it with the header.
  With a missing directory, it exits with status 1.
- `go vet ./...`, `go test ./...` and `golangci-lint run ./...` pass.

## 12. Not in this specification

- The kinds of [#81](https://github.com/schors/velourbbs/issues/81): the oneliners
  and the last callers.
- The data of mods ([#12](https://github.com/schors/velourbbs/issues/12)), accounts.
- A command that changes the data. The sysop edits the file (section 8).
