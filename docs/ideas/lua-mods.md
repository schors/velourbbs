# Idea: Lua mods

Backlog. It extends [bbt-format](../specs/bbt-format.md) and [menu-nav](../specs/menu-nav.md)
with the action `run`.

## Goal

Lua mods add interactive features to the BBS.
A sysop installs a mod as a directory and links it from a menu item.
Oneliners, last callers and a poll are stock mods.

## Decisions

| Topic | Decision |
|---|---|
| Runtime | Lua 5.1 with `goto`, through gopher-lua, in pure Go. The docs describe only this dialect. |
| Doors | Not supported. Games and utilities are Lua mods. |
| Roles | A data provider ([lists](lists.md)) or an action (`do = "run oneliners/main"`). |
| Data provider | It returns a table with 500 rows at most. A call to an input function is an error. |
| Action | It runs in a navigation frame. It can show screens, read input and run actions. |
| Trust | Lua scripts are trusted code (tier 3 of [content-trust](content-trust.md)). The sysop reviews a mod before install. |
| State | Each session has one Lua state. Sessions never share a state. Each script run gets a fresh environment. |
| Suspend | Each session has one goroutine. Input calls block until a key, a timeout, the idle timeout or a hangup. Coroutines are not necessary. |
| Panic | Each session goroutine recovers from panics. A panic ends one session, not the process. |
| Script error | The user sees "System error, returning to menu". The frame pops. The sysop log gets the file, the line and the traceback. |

Rejected options:
- Own script language. Errors, docs and tools cost months, and maintenance never ends.
- WebAssembly. Sysops need a compiler toolchain. Suspend for input is difficult.
  The trust tiers of [content-trust](content-trust.md) do not need its isolation.
- Other engines (JS, Starlark, Rhai, Janet, Wren). They have no suspend for input, a large engine,
  or a language that sysops do not know.

## Host API

`api = 1` has 23 calls. A call is in `api = 1` if a stock mod uses it or if it reads core data.
`api = 1` freezes after the three stock mods work.
All calls are in the table `bbs`. Example: `bbs.screen`.

| Group | Calls |
|---|---|
| Output | `screen(name, data)`, `print(markup, data)`, `cls()` |
| Input | `key(timeout)` returns a key name or `nil`. `input_line(opts)`. `select(items, opts)` uses the selection code of the host. |
| Actions | `action(verb, target, args)`. It passes the same `require` check as a menu item. |
| Users | `user`: a read-only snapshot of the session user. `users.get(id)`: public fields only. |
| Messages | `msg.areas`, `msg.list`, `msg.read`, `msg.post`. They use the access of the session user. The author of `msg.post` is always the session user. |
| System | `who()`, `calls.recent(n)` |
| Store | `store.get`, `store.set`, `store.update(key, fn)`, `store.keys(prefix)` |
| Other | `log`, `now()` (epoch seconds and the local board date `YYYY-MM-DD`), `random`, `sleep` (10 s at most for each call) |

`print` takes one line of BBS template body with style tokens and the fields of `data`.
It cannot have includes.
The host removes ESC and C0 characters from it.

Example:

```lua
-- mods/oneliners/main.lua
bbs.screen("oneliners/list", { lines = bbs.store.get("lines") or {} })
if bbs.select({ "Add a line", "Quit" }) ~= 1 then return end
local text = bbs.input_line({ max = 60, prompt = "{!prompt}Your line: " })
if text == "" then return end
bbs.store.update("lines", function(lines)    -- runs under the mod lock
  lines = lines or {}
  table.insert(lines, 1, { who = bbs.user.name, text = text })
  while #lines > 10 do table.remove(lines) end
  return lines
end)
```

## Store

- Each mod has one store. Values are JSON-like.
- Limits: 64 KB for each value, 128 bytes for each key, 16 MB for each mod.
- `update` runs `fn` under a lock for the mod and commits the result. `fn` cannot call `bbs.*`.
  The slice time limit (section Limits) limits it.
- Per-user keys prevent races, for example `vote/<user id>` in the poll mod.
- A mod resets daily data when the first user of the day starts it.
  It compares the date from `now()` with a stored date.

## Mods

- `mods/<name>/` holds `mod.toml`, a `screens/` directory and Lua files.
- `mod.toml` has `name`, `api = 1` and an optional `require` for each script.
- The directory name is the namespace: `run oneliners/main`, `bbs.screen("oneliners/list")`.
- Lua `require` loads only from the mod directory.
  Module names match `[a-z0-9_]+(\.[a-z0-9_]+)*`.
- `velourbbs check` lists the scripts of each mod and its `bbs.*` calls that write data.

## Limits

A "slice" is the run time between two waits for input.

- Libraries: `base` without `load`, `loadstring` and `dofile`; `string`, `table`, `math`.
  There is no `io`, `os`, `debug` or `coroutine`.
- Time: 2 s wall clock for each slice, through the context of the Lua state.
- Size caps: 1 MB for the results of `string.rep`, `table.concat`, `string.format` and `gsub`,
  and for pattern subjects. The sandbox replaces these functions with capped versions.
- Memory: gopher-lua has no memory limit for each state. A heap guard aborts all running slices
  when the process heap passes the configured limit.
- `sleep` counts as a wait. The session time limit applies inside every wait.
- An abort is sticky. A hangup, a limit or a guard abort sets the abort flag of the session and
  cancels the context of the Lua state. While the flag is set:
  - the VM raises the abort error at each instruction, so a pure Lua loop also stops;
  - each host call fails;
  - the sandbox versions of `pcall` and `xpcall` raise abort errors again.

Known gaps:
- Doubling a string with `..` in a loop can make a memory spike before the heap guard acts.
  This is accepted for trusted code.
- The loader cannot check the fields of a Lua provider. A missing field renders empty,
  and the host logs a warning.

## Node messaging

A later phase:
- `wait(opts)` with the events `{type = "key"}`, `{type = "timeout"}` and
  `{type = "msg", from = {node, user, mod}, data}`. `key(timeout)` does not change.
- `node.send`: the host sets `from`. The script cannot set it.
- The host delivers a message only to a session that waits in the same mod.
- Each session has an inbox of 32 messages. Each payload is 4 KB at most.
- When the inbox is full, `node.send` does not deliver. It returns the delivered count.
- A `msg` option for `select` and `input_line`, and a two-node demo game.

Other later calls: `gotoxy`, `users.find`, `user_set`.
`user_set` changes profile fields only, never the level, the flags or the password.

## Dependencies

- Users, login and message areas. There is no idea for them yet.
- Navigation frames and `require` come from [menu-actions](menu-actions.md).
- Data providers come from [lists](lists.md).

## Open questions

Check these with gopher-lua at the chosen version:
1. An abort stops `while true do pcall(f) end`, where `f` makes no host call.
2. The wall clock and long single operations (issue yuin/gopher-lua#521).
3. The heap guard.
4. Long calls inside host code.
5. A hangup and a reload while a script waits for input.

Other:

6. Do the aliases of `screens/screens.toml` apply in mod screens and in `print`?
   A field without a dot in mod art then means what the host defines.
