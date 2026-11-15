# Idea: menu actions beyond goto

Backlog. It extends [bbt-format](../specs/bbt-format.md) and [menu-nav](../specs/menu-nav.md).

## Goal

Menu items get a navigation stack, access checks, typed arguments, built-in actions and
arrow-key selection.

## Decisions

| Topic | Decision |
|---|---|
| Actions | `goto <screen>`, `gosub <screen>`, `back`, `top`, `page ...` ([lists](lists.md)), `run <mod>/<script>` ([lua-mods](lua-mods.md)), `call <name>`, `logoff`. The actions are a fixed set. |
| `gosub <screen>` | Push the screen. `goto <screen>` replaces the top frame. |
| `back` / `top` | Pop one frame / pop to the first frame. |
| Frame | A frame holds the screen, its arguments, the page and the cursor. `back` restores all of them. |
| Stack size | 16 frames at most. |
| `call <name>` | Run a built-in: `msg.read`, `msg.post`, `msg.scan`, `file.list`, `file.get`, `user.edit`, `chat`. |
| Wrong action | The loader error names the correct action: `oneliners/main is a script: use "run"`. |

## Typed arguments

- `do` is `<action>` or `<action> <target>`. Both parts are literals.
- Each `args` value is a literal or exactly one `{<path>}` token.
- The host puts the value into its argument slot as one opaque string.
  It never parses a string after substitution.
- `ask = "Message number: "` asks for a line before the action. `{in}` is the answer.
  `in` has no dot, so it shares the names of the aliases in `screens/screens.toml`.
- `{arg.*}` gives the arguments of the current screen.
- The loader checks actions, targets and argument names.

```toml
[[item]]
key  = "J"
ask  = "Message number: "
do   = "gosub msg_read"
args = { area = "{arg.area}", n = "{in}" }
```

## Access

- A screen has `require` in its front matter. A script has `require` in its `mod.toml`.
- Built-ins check area access themselves.
- The host checks `require` when the action runs.
  The check is the same for menu items and for the Lua call `bbs.action`.
- `when` only disables the item and its key. The template hides the item text with
  `{#if item.<KEY>}`.
- `velourbbs check` gives a warning for a `when` item whose target has no `require`.

## Arrow-key selection

- A mark `{@<id>}...{/@}` is a named span in a screen. [bbt-format](../specs/bbt-format.md), section 3.3, reserves this syntax.
- Layout reports the final cells of each mark.
- Arrow keys move over the marks. Enter runs the item of the selected mark. Hotkeys also work.
- A selected mark uses a theme style.

## Dependencies

- Users and login give the data for `require`. There is no idea for them yet.
- Typed command input is in the backlog of [menu-nav](../specs/menu-nav.md).
