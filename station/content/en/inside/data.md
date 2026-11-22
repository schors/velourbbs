# Data

## What it does

Velour keeps the data of the station in one text file. The SysOp reads and edits it
in a text editor.

## How

The file is JSON Lines ([jsonlines.org](https://jsonlines.org/)): UTF-8, one record is
one JSON object on a line. The first line is the header with the format version:

```
{"format":"velourbbs-store","version":2}
```

Each next record has a kind, the field `kind`. The code defines each kind: its fields, checks and the limit
of the number of records. When there are more records than the limit, the oldest go.
A time is RFC 3339 in UTC, with seconds.

Velour keeps all records in memory. Each change writes the whole file:

1. It checks the new records.
2. It writes them to the temporary file `<file>.tmp` with mode 600 and syncs it to the disk.
3. It renames the temporary file over the old one. This is the commit point.
4. It syncs the directory to the disk.

## What it gets

- After a crash on Linux or macOS, the file has the records before the change or after it, not a mix.
- An error before the rename keeps the old file and the old records.
- Each change of the format increases the version. Migrations change an old file to the new version
  at the read. An error shows the line of the file.

## Limits

- The file is at most 16 MB, because all of it is in memory.
- One file has one `velourbbs serve`. The file is not locked.
- An edit of the file while the server runs goes away at the next write.
- An older Velour does not read a file of a newer version. Before a deploy with a migration,
  make a copy of the file.
- On Windows, Go does not make the rename atomic, and it fails
  while another program has the file open.
- Velour does not use SQLite and bbolt: their file cannot be read without a separate program.
  Also, `modernc.org/sqlite` adds 6 MB and 10 modules to the program.
