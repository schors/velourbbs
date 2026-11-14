# Idea: load-memory

Backlog. It changes the memory of the screen load in `internal/bbt`.
Source: the review of the block `loader-resolve` (issue `pncaz-mvp-audit-fixes`).

## Goal

The memory of the screen load has a bound that fits a small host.

## Facts

- `Load` parses each screen file before the screen pass and keeps all of them.
  A 1 MB screen file takes about 40 MB of heap after the parse.
- The limit "all screens after includes" (screen-format §10) stops the load only
  in the screen pass, so it does not bound the parse.
- The limits of §10 allow 16 MB of body, about 640 MB of parsed nodes.
  The dev machine has 1.8 GB of RAM.

## Open questions

- Parse each screen file in the screen pass, and skip the parse after the total limit.
- A smaller form of the parsed nodes, or lower limits in §10.
