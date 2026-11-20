# 1. The program

You need Go 1.27.3 or later. In the source repository of Velour BBS, run:

```
go build ./cmd/velourbbs
sudo install velourbbs /usr/local/bin/
```

The first command writes the file `velourbbs`. The second command copies it
to `/usr/local/bin/velourbbs`.

Build in a git checkout of a release tag without changes, for example after `git checkout v1.2.0`.
Then the field `bbs.version` shows the tag. Each other build gives `dev`.
Full rules: [session-fields](../../../specs/session-fields.md), section 3.4.

## Result

`velourbbs` without arguments prints the usage:

```
usage:
  velourbbs art import <input> <output>
  velourbbs check [--size <cols>x<rows>]...
  velourbbs fields [--sample]
  velourbbs preview <name> [--size <cols>x<rows>] [--sample <set>] [--set <path>=<value>]... [--keys <keys>] [--view]
  velourbbs serve
```
