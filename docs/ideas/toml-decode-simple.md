# Idea: toml-decode-simple

Backlog. It changes `internal/tomlcheck`, `internal/config`, `internal/bbt/front.go`
and [screen-format](../specs/screen-format.md), sections 9.1 and 11.
Source: the second MVP audit (issue `jgqjh-mvp-reaudit`).

## Goal

Each TOML key has one definition in the code. The specs do not repeat the code.

## Facts

- `tomlcheck.Decode` (249 lines) decodes each document three times:
  into plain types, into types with `UnmarshalText`, and into `map[string]any`.
  It reads go-toml error text (`"cannot decode TOML "`, `" of type "`). 40 tests pin that text.
- Each key with a value check has two structs: `frontDoc` and `frontValues`,
  `telnetDoc` and `telnetValues`, aligned by index (`internal/bbt/front.go:103`).
- go-toml v2.4.3, checked with a probe:
  - a strict decode gives a type error with a position (`key = 5` into `string`: 1:7);
  - a strict decode gives all unknown keys with positions;
  - a type with `UnmarshalText` takes an integer as text (`hot = 5` gives `"5"`);
  - `KEY` matches the field `key`, and strict mode does not stop it;
  - a table `[item]` fills `[]Item` with one element.
- screen-format §11 "Code" and the position rules in §9.1 repeat the code.
- The backlog adds keys: `data`, `select`, `args`, `mod.toml`.

## Options

1. Simplify: one strict decode into plain types; value checks in Go after it.
   - A value error names the key path, with no line number. A front matter has about 10 lines.
   - `KEY` for `key` and a single `[item]` are accepted.
   - Saves about 260 code lines and 100 test lines (estimates).
2. Keep: strict and with positions. Review the error text at each go-toml upgrade.

In both options: cut screen-format §11; shrink §9.1 to the list of errors.

## Open questions

- Option 1 or 2. The audit recommends option 1.
