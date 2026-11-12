# Idea: menu and navigation

MVP part 2 of 3. The other parts are [screen-format](screen-format.md) and
[telnet-term](telnet-term.md).

## Goal

A user goes from one screen to other screens with hotkeys and ends the session.
All navigation is explicit. The engine has no implicit transitions.

## Concepts

- **Menu.** The `[[item]]` list in the front matter of a `.bbt` screen.
- **Item.** One choice: `key`, `do` and an optional `when`. An item has no position on the screen.
  The screen author writes the item text, for example `[A]bout`, into the screen.
- **Current screen.** The session holds the name of one screen. There is no navigation stack.

## Decisions

| Topic | Decision |
|---|---|
| Users | None. All sessions are anonymous. Login is a later idea. |
| Actions | `goto S` (show screen S) and `logoff` (end the session). |
| Navigation | Explicit only. Each transition is an item with a target screen. There is no stack, no `back` and no `top`. |
| Input | Hotkey only: one key runs the item, without Enter. |
| Keys | A Latin letter, a digit or a named key (`ESC`, `ENTER`, ...). Letters are case-insensitive. |
| Russian layout | The host maps a Cyrillic letter to the Latin letter at the same key position (`ы` → `S`). |
| `ANY` | A named key. It matches each key that has no item of its own. |
| Unknown key | The host ignores it. ESC has no implicit action. |
| Screen without items | A `bbs check` error, except `goodbye.bbt`. |
| Start | The first screen is `main`. `bbs.toml` can set another name. |
| `logoff` | The host shows `goodbye.bbt` if it exists, without a wait for a key. Then it disconnects. |
| `when` | It disables an item when false. A disabled item counts as absent, so `ANY` can match its key. |
| Condition time | The host evaluates `when` once for each render. Shown text and enabled keys always agree. |
| Item state in templates | The condition path `item.KEY` is true when the item is enabled. The template hides the item text with `{#if item.KEY}`. `item.*` is not permitted in `when` and in fields. |
| `narrow` screen | It uses the items of the main screen. The `narrow` file has no items and no `narrow` of its own. |
| "More" pause | Any key shows the next page. Items work after the last page. |
| Render | Each screen starts with a clear screen. Keys that arrive during a render are kept. |
| `label` | Not in the MVP. The engine does not print item text. |

`goto` to the current screen renders it again, for example after a resize.

## Example

```
+++
[[item]]
key = "S"
do = "goto sysinfo"
when = "term.cols >= 80"
[[item]]
key = "G"
do = "logoff"
+++
{>main.ans}
{#if item.S}[S]ystem info  {/if}[G]oodbye
```

`screens/sysinfo.bbt`:

```
+++
[[item]]
key = "ANY"
do = "goto main"
+++
Terminal: {term.cols}x{term.rows}. Press any key.
```

## MVP scope

In:
- the actions `goto` and `logoff`, items, `when`, `ANY`, the Russian layout map;
- `bbs check`: unknown action, missing screen, duplicate key (case-insensitive), invalid key,
  unknown `item.KEY`, a screen without items, a missing start screen;
- `bbs preview NAME --keys "S,ANY"`: run a key sequence without telnet and write a cell dump of
  each screen. It is the test harness for navigation.

Out (backlog):
- `gosub`, `back`, `top` and a navigation stack.
- `require` (access checks). It needs users.
- Arrow-key selection and marks (see screen-format).
- Typed command input.
- `args`, `ask` and `{arg.*}`.
- The actions `page`, `run` and `call`. They need lists, Lua or built-ins.
- Lua.

## Dependencies

- Screens and conditions come from [screen-format](screen-format.md).
- Key decoding (ESC, named keys) comes from [telnet-term](telnet-term.md).

## Open questions

1. The list of named keys is an open question of [telnet-term](telnet-term.md).
