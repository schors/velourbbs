# Idea: content trust and raw art

Backlog. It extends [bbt-format](../specs/bbt-format.md) and [art-format](../specs/art-format.md).

## Goal

The host treats each kind of content by its trust tier. User content cannot change the terminal.

## Decisions

| Tier | What | Treatment |
|---|---|---|
| 1 | User content: names, messages, input | Hostile. Always escaped and filtered. |
| 2 | Screens, art, theme, TOML files | Files of the sysop. Limits catch mistakes. See [bbt-format](../specs/bbt-format.md), sections 1 and 9, [bbs-toml](../specs/bbs-toml.md), section 2, and [art-format](../specs/art-format.md), section 5. |
| 3 | Lua scripts ([lua-mods](lua-mods.md)) | Trusted code. Limits catch bugs only. |

## User content

- Field values follow [bbt-format](../specs/bbt-format.md), section 4.3.
- The host removes the same characters from log lines.
- Line input rejects control characters.
- Message bodies keep colour only (SGR). The host removes cursor moves and all other sequences.
  So message bodies lose ANSI layout.

## Raw art

- Art without fields can be "raw art". The host sends its bytes as they are,
  so ANSI animation is possible.
- Raw art cannot be cut or centred.
- Raw bytes pass the "raw filter". It keeps SGR, cursor moves, erase, save and restore.
  It removes OSC, DCS, APC, PM, SOS and report requests.

## Open questions

1. Where a sysop marks art as raw art. [bbt-format](../specs/bbt-format.md) has front matter only in BBS templates.
