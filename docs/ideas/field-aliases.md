# Idea: field-aliases

Backlog. It changes field aliases ([screen-format](../specs/screen-format.md), sections 4.1 and 4.2)
and the loader in `internal/bbt` and `internal/screen`.
Source: the second MVP audit (issue `jgqjh-mvp-reaudit`).

## Goal

A fragment and an art file mean the same in each screen. The loader resolves each field once.

## Facts

- `[fields]` in the front matter of a screen maps an alias to a registry path: `w = "term.cols"`.
  The alias is valid in the screen, its fragments and its art. Short aliases let art have short fields.
- Fragments and art are shared by screens, so their fields have no meaning without a screen.
  The file pass gives `node` values (`internal/bbt/file.go:19-52`). The screen pass copies them
  into `screen.Seg` values with a closed `Ref` (`internal/screen/field.go:46-71`),
  with caches for each screen, `errorOnce` and the `screen "<name>": ` error prefix.
- A new token touches five places: the scanner, a node type, a case in `resolve`,
  a `Seg` type, a case in the render.
- The registry has five `term.*` fields. Lists need `row.*` and `page.*`,
  menu actions need `arg.*`, Lua gives `data` tables. None fits a closed `Ref`.

## Options

1. Global aliases in `bbs.toml`. Fields resolve once, in the file pass.
2. No aliases until lists.
3. The model keeps the path. An open registry maps a path or a prefix to a type.
   All screens share a fragment body. The screen pass only checks.
   Do it as the first step of lists.

Each option saves about 120–200 code lines (estimates).

## Open questions

- Option 1, 2 or 3. The audit recommends option 3 at the start of lists,
  and option 1 if lists are not the next feature.
