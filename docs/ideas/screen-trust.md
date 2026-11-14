# Idea: screen-trust

Backlog. It changes the trust premise of screen files in [screen-format](../specs/screen-format.md),
sections 1 and 10, and tier 2 of [content-trust](content-trust.md).
It changes `internal/bbt`, `internal/screen`, `internal/art` and `internal/menu`.
Source: the second MVP audit (issue `jgqjh-mvp-reaudit`).

## Goal

The screen load has a trust premise that the code keeps, at the lowest cost.

## Facts

- The premise now: screen files are untrusted data. Code that exists only for it:
  - the body and total counters in lines and bytes, `cut`, `totalOver`
    (`internal/bbt/screen.go:99-120`, `internal/bbt/bbt.go:106-111`);
  - the nesting limit (`internal/bbt/file.go:209`);
  - `FileDiags` with `Full()` guards (`internal/screen/diag.go:86-118`), and the second sort
    of the diagnostics into files in `internal/menu/check.go:20-54`;
  - the row reuse in render (`internal/screen/render.go`, `free`, `out`): the rows of a page
    are valid until the next page (`internal/menu/session.go:18`);
  - the art player speed paths: lazy `CSI 2 J` with generations, bulk `CSI n b`
    (`internal/art/play.go:55-58`, `:206-267`, `:462-511`).
- The code does not keep the premise. A 1 MB screen file takes about 38 MB of heap after `Load`.
  The cause is `screen.Text` as `[]Grapheme`, 24 bytes for each grapheme.
  `Load` parses each screen file before the total limit applies.
  The number of screen files has no limit. The dev machine has 1.8 GB of RAM.
- Without the speed paths, a crafted 1 MB art file loaded in up to 2 min; with them, up to 6 s.
  Moebius never writes `CSI 2 J` or `CSI n b`. icy_draw writes `CSI 2 J` once at the start,
  and `CSI n b` only at the levels IcyTerm and Utf8Terminal.
- A pack with a Lua mod is trusted code (tier 3). Screens come in the same pack.
  A hostile screen file can only slow the start of the host. `bbs check` shows it before the start.

## Options

1. Drop the premise: "screen files are trusted; the limits catch mistakes".
   - Keep: 1 MB for a file, the include cycle check, 10000 body lines, 1000 art rows,
     10000 render rows.
   - Cut: the byte counter, the total limits, the nesting limit, `cut`, `FileDiags`
     (cap the output in `writeDiags`), the row reuse in render.
   - Art player: cut the speed paths. Add a write budget: count the written and erased cells,
     and stop with one error over about 16 grids. The load time with a budget is not measured.
   - Saves about 200 code lines and 200–350 test lines (estimates).
2. Keep the premise and make it true: store a text run as a string and segment it in render;
   parse a file only below the total limit; limit the number of files. Effort M.

## Open questions

- Option 1 or 2. The audit recommends option 1.
- Option 1: the budget value, from the load time of the largest real art file.
