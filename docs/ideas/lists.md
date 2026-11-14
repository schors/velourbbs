# Idea: lists and data providers

Backlog. It extends [screen-format](../specs/screen-format.md) and
[menu-nav](../specs/menu-nav.md).

## Goal

A screen shows a paged list of records, for example messages of an area.
The user pages through the list and selects a row.

## Decisions

| Topic | Decision |
|---|---|
| Data | `data = { from = "msg.list", args = { area = "{arg.area}" } }` in the front matter. `from` names a built-in provider or a Lua provider (`lua:lastcallers/list`, see [lua-mods](lua-mods.md)). |
| Page size | `rows = 15` in the front matter. |
| Fields | `row.*` for the fields of one row. `page.n`, `page.count` and `page.empty` for the page. |
| Text list | `{#each <list>}...{/each}` in a BBS template: one line for each row of the page. |
| Art list | The art row with `row.*` fields is the "list row". A screen has one list row. |
| Paging | The action `page next\|prev\|first\|last`. |
| Selection | `select = { do = "...", args = { id = "{row.id}" } }`. The cursor moves over list rows. Enter runs the action with the selected row. |
| Conditions in a list row | Only display fields such as `row.mark`. |
| Formats | There are no format filters. Each user sets the date and number formats. |

## List row in art

- The list rectangle has the columns from the first to the last `row.*` token and `rows` rows.
- The cells below the list row inside the rectangle must be blank. Else, the loader gives an error.
- Cells outside the rectangle stay, for example frame borders.

## `#if` inside art

- An `#if` block in art is a fixed box on one row, from `{#if` to `{/if}`.
- When the condition is true, the cells between the markers become one line of text with their
  styles. The line is padded or cut to the box width.
- When the condition is false, the box shows spaces in the style of its first cell.
- Nothing outside the box moves.
- Blocks and lists in art have one row and no nesting.

## Sample data

- `bbs fields` lists each field with its type, a sample value and its maximum width.
- `bbs fields --sample` writes `sample.toml` with sample values.
- The sysop adds named sample sets to `sample.toml`, for example one set with new mail and one
  without. `bbs preview <name> --sample newmail` uses a set.

## Example

`screens/msglist.ans` (the frame continues for 15 rows), with `[fields]` of `bbs.toml` `n = "page.n"`,
`c = "page.count"` and `m = "row.mark"`:

```
 ╔═══════════════════════════════════════════════════════════════╗
 ║ {area.name                          }          page {n }/{c } ║
 ║ {m}{ row.num} {row.from        } {row.subject                }║
 ║                                                               ║
 ╚═[N]ext═[P]rev═[J]ump═[Q]uit═══════════════════════════════════╝
```

A `.bbt` text list:

```
{!title} {area.name} {!dim}page {page.n}/{page.count}
{#each msgs}{row.mark}{   row.num} {row.from        } {row.subject                  }{/each}
{#if page.empty}{!dim}  No messages in this area.{/if}
```

## Open questions

1. How `<list>` in `{#each <list>}` names the `data` of the screen.
   A name without a dot shares the names of the aliases in `bbs.toml`.

## Dependencies

- Lua providers come from [lua-mods](lua-mods.md).
- `args` and `{arg.*}` come from [menu-actions](menu-actions.md).
