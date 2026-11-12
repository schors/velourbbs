# Screens, menus and scripts

This document defines how the BBS shows screens, binds keys to actions and runs scripts.

## 1. Summary

1. A screen is `screens/NAME.ans` (ANSI art) and/or `screens/NAME.txt`, plus an optional
   `screens/NAME.toml` for input, items, data and access.
2. Screens use one small logic-less token language: `{...}`. It has fields and `#if`. Text files
   also have colours, includes and `#each`. There is no arithmetic, no filter and no loop in art.
3. Art is WYSIWYG. The drawn span of a token is the field width.
4. Menu items are `[[item]]` records with a key, a label, `do = "verb target"` and typed `args`.
   The verbs are a fixed set.
5. `require` on a target controls access. The engine checks it when the action runs.
   `when` only hides an item.
6. Lua gives the dynamics. A script is a data provider or an action. `api = 1` has 23 calls.
7. All text is UTF-8: art, screens, user content and terminal output. There is no other charset.
8. Clients are telnet and SSH.

## 2. Decisions

| Topic | Decision |
|---|---|
| Script language | Lua 5.1 through gopher-lua. |
| Host language | Go. |
| Charset | UTF-8 only. CP437 terminals are not supported. |
| Clients | Telnet and SSH. |
| Doors (external programs) | Not supported. Games and utilities are Lua mods. |
| Old codes (pipe, `@`, MCI) | Not supported. No converter. |

Rejected options:
- **Own script language.** Errors, docs and tools cost months, and maintenance never ends.
- **WebAssembly.** Sysops need a compiler toolchain. Suspend for input is difficult. The trust model
  (section 6) does not need its isolation.
- **Compiled display files (MECCA style).** The host parses files at load. A compile step adds work.
- **Other engines** (JS, Starlark, Rhai, Janet, Wren). No suspend for input, a large engine,
  or a language that sysops do not know.

## 3. Template language

### Tokens

| Token | Art | Text | Meaning |
|---|---|---|---|
| `{user.name}` | yes | yes | Field. Always escaped (section 6). |
| `{user.name      }` | yes | yes | Field padded or cut to the token span. |
| `{   user.calls}` | yes | yes | Same, aligned right. |
| `{n}` | yes | yes | Alias from `[fields]` in the screen TOML, for example `n = "page.n"`. |
| `{#if cond}` `{#else}` `{/if}` | one row | yes | Show or hide a block. |
| `{bar:id}` | yes | no | Lightbar cell of item `id`. |
| `{!hi}`, `{!14/1}` | no | yes | Colour: theme style, or fg/bg numbers. |
| `{>header}` | no | yes | Include `screens/header.txt`. One level. The name matches `[a-z0-9_-]+`. |
| `{#each list}` `{/each}` | no | yes | One line for each item of the page. Fields are `row.*`. |

**Token start.** `{` starts a token only when a letter, a space, `#`, `/`, `>` or `!` follows it.
The token must close on the same row. In art, the token must also have one colour. Any other `{` is
a plain character, so logos with braces are safe. In text, `{{` gives a literal `{`.

**Width rule.** A field token with inner spaces has the width of its span. A field token without
inner spaces:
- has the width of its span in art, because the grid does not move;
- has its natural width in text and inside an art `#if` block.

**Conditions.** One grammar and one parser serve `#if`, `when` and `require`. A condition is `path`,
`not path`, or `path op literal` with `== != < <= > >=`. `and` and `or` join conditions.
There are no parentheses. More complex logic goes into a Lua data provider.

**Fields** come from a typed registry in the host: `user.*`, `node.*`, `term.*`, `area.*`, `page.*`,
`arg.*`, `row.*`. The host gives display strings. Dates use the format setting of the user.
An alias has 1-8 characters and cannot have the name of a registry root. `bbs fields` lists each
field with its type, a sample value and its maximum width.

### Main menu

`screens/main.ans`: the text that the sysop types into the art (art not shown):

```
  {bar:msg          }   {bar:file         }   {bar:who          }
  {bar:one          }   {bar:sys          }   {bar:bye          }
  Welcome back, {user.name           }   Time left: {t } min
  {#if user.new_mail}You have {user.new_mail} new mail.          {/if}
```

`screens/main.toml` (two of six items shown):

```toml
input = "lightbar"                 # hotkey | lightbar
[fields]
t = "user.time_left"

[[item]]
key   = "M"
bar   = "msg"
label = "Message areas"
do    = "gosub msg_areas"

[[item]]
key   = "S"
bar   = "sys"
label = "Sysop menu"
do    = "gosub sysop"
when  = "user.level >= 90"         # hides the item; screens/sysop.toml has require
```

### Message list with paging

`screens/msglist.ans` (the frame continues for 15 rows):

```
 ╔═══════════════════════════════════════════════════════════════╗
 ║ {area.name                          }          page {n }/{c } ║
 ║ {m}{ row.num} {row.from        } {row.subject                }║
 ║                                                               ║
 ╚═[N]ext═[P]rev═[J]ump═[Q]uit═══════════════════════════════════╝
```

`screens/msglist.toml` (items N, P and Q not shown):

```toml
input   = "lightbar"                         # the cursor moves over list rows
rows    = 15
require = "area.can_read"
data    = { from = "msg.list", args = { area = "{arg.area}" } }
select  = { do = "gosub msg_read", args = { area = "{arg.area}", id = "{row.id}" } }
[fields]
n = "page.n"
c = "page.count"
m = "row.mark"                               # "*" for unread, else a space

[[item]]
key  = "J"
ask  = "Message number: "
do   = "gosub msg_read"
args = { area = "{arg.area}", n = "{in}" }
```

`screens/msglist.txt` is the fallback for narrow terminals. It uses `#each`:

```
{!title} {area.name} {!dim}page {page.n}/{page.count}
{#each msgs}{row.mark}{   row.num} {row.from        } {row.subject                  }{/each}
{#if page.empty}{!dim}  No messages in this area.{/if}
{!prompt}[N]ext [P]rev [J]ump [Q]uit, Enter = read
```

### User information

`screens/userinfo.txt`. It has no TOML file, so any key goes back:

```
{>header}
{!title}User information{!normal}
  Name .....: {user.name                    }  Level ..: {user.level}
  Location .: {user.location                }  Calls ..: {     user.calls}
  Terminal .: {term.cols}x{term.rows}
{#if user.expires}  Expires ..: {user.expires}{/if}
{#if not user.email}{!warn}  No e-mail address. Press E in the profile menu.{/if}
```

### ANSI art

An art file is an ANSI file in UTF-8. The sysop draws the art and types tokens into it as plain
characters. The loader does these steps:

1. It removes a SAUCE record at the end of the file, if there is one. The art width is `width` in
   the screen TOML. The default is 80.
2. It clamps the input: file size 1 MB, grid 255 x 1000 cells, 500 tokens.
3. It plays the ANSI stream into a cell grid.
4. It finds the tokens. A token that changes colour is a load error.
5. A field and a bar take the colour of their first cell. A selected bar uses the theme style
   `bar_selected`.
6. **`#if` block.** A block is a fixed box on one row, from `{#if` to `{/if}`. When the condition is
   true, the cells between the markers become one line of text with their painted colours. Fields
   follow the width rule. The line is padded or cut to the box width. When the condition is false,
   the box shows spaces in the colour of its first cell. Nothing outside the box moves.
7. **List row.** The row with `row.*` fields is the list row. `data` in the TOML names the list.
   The list rectangle has the columns from the first to the last `row.*` token and `rows` rows.
   The cells below the list row inside the rectangle must be blank. Else, the loader gives an error.
   Cells outside the rectangle, for example frame borders, stay. A screen has one list row.
   Conditions inside a list row come from display fields such as `row.mark`.

Blocks and lists in art have one row and no nesting. This limit is fixed.

The host sends the grid, not the source bytes. When both `NAME.ans` and `NAME.txt` exist, a terminal
that is narrower than the art gets the `.txt` file. Art without tokens can set `raw = true` in its
TOML, so ANSI animation is possible. Raw bytes pass the raw filter (section 6).

## 4. Action binding

**Item record:** `key`, `label`, `do`, `args`, and the optional fields `bar`, `when` and `ask`.
Named keys such as `ESC`, `ENTER` and `UP` are valid keys. The hotkey label comes from the record.

**Typed arguments.** `do` is `verb target`. Both parts are literals. Each `args` value is a literal
or exactly one `{path}` token. The engine puts the value into its argument slot as one opaque
string. It never parses a string after substitution. `{in}` is the answer to `ask`. `{row.*}` is
the selected list row.

| Verb | Effect |
|---|---|
| `goto S` / `gosub S` | Replace the top frame with screen S / push screen S. |
| `back` / `top` | Pop one frame / pop to the first frame. |
| `page next\|prev\|first\|last` | Move the list page. |
| `run MOD/SCRIPT` | Run a Lua action script. |
| `call NAME` | Run a built-in: `msg.read`, `msg.post`, `msg.scan`, `file.list`, `file.get`, `user.edit`, `chat`. |
| `logoff` | End the session. |

The loader checks verbs, targets and argument names. An error for a wrong verb names the correct
verb: `oneliners/main is a script: use "run"`.

**Access.** A screen has `require` in its TOML. A script has `require` in its `mod.toml`.
Built-ins check area access themselves. The engine checks `require` when the action runs. The check
is the same for menu items and for `bbs.action`. `when` only hides an item and disables its key.
`bbs check` gives a warning for a `when` item whose target has no `require`.

**Navigation stack.** A frame holds the screen, its arguments, the page and the cursor. `back`
restores all of them. The stack has 16 frames at most. A `bbs.action("gosub", ...)` call returns
when the stack pops past its frame.

**Input modes.**
- `hotkey`: one key runs the item, without Enter.
- `lightbar`: the arrow keys move over `{bar:id}` cells or list rows. Enter runs the item.
  Hotkeys also work. A screen without art shows its bars as a vertical list.

## 5. Lua

**Runtime.** gopher-lua: Lua 5.1 with `goto`, in pure Go. The docs describe only this dialect.

**Roles.**
- Data provider (`data = { from = "lua:lastcallers/list" }`). It returns a table with 500 rows at
  most. A call to an input function is an error.
- Action (`do = "run oneliners/main"`). It runs in a frame. It can show screens, read input and run
  actions.

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

```lua
-- mods/poll/main.lua: one key for each user, so votes do not race
local poll = bbs.store.get("poll")
local choice = bbs.select(poll.options, { title = poll.question })
if not choice then return end
bbs.store.set("vote/" .. bbs.user.id, choice)
local counts = {}
for _, k in ipairs(bbs.store.keys("vote/")) do
  local c = bbs.store.get(k)
  counts[c] = (counts[c] or 0) + 1
end
bbs.screen("poll/results", { rows = result_rows(poll.options, counts) })  -- local helper
```

**Host API, `api = 1`: 23 calls.** A call is in `api = 1` if a stock mod uses it or if it reads
core data.
- Output (3): `screen(name, data)`, `print(markup, data)`, `cls()`.
- Input (3): `key(timeout)` returns a key name or `nil`. `input_line(opts)`. `select(items, opts)`
  uses the lightbar and hotkey code of the engine.
- Actions (1): `action(verb, target, args)`. It passes the same `require` check as a menu item.
- Users (2): `user` (a read-only snapshot of the session user), `users.get(id)` (public fields only).
- Messages (4): `msg.areas`, `msg.list`, `msg.read`, `msg.post`. They use the access of the session
  user, the same as the built-ins. The author of `msg.post` is always the session user.
- System (2): `who()`, `calls.recent(n)`.
- Store (4): `store.get`, `store.set`, `store.update(key, fn)`, `store.keys(prefix)`.
- Other (4): `log`, `now()` (epoch seconds and the local date of the board, `YYYY-MM-DD`),
  `random`, `sleep` (10 s at most for each call).

**Store.** Each mod has one store. Values are JSON-like. Limits: 64 KB for each value, 128 bytes
for each key, 16 MB for each mod. `update` runs `fn` under a lock for the mod and commits the
result. `fn` cannot call `bbs.*`. The slice wall clock limits it. A mod resets daily data when the
first player of the day starts it: it compares the date from `now()` with a stored date.

**Mods.** `mods/<name>/` holds `mod.toml` (`name`, `api = 1`, an optional `require` for each
script), a `screens/` folder and Lua files. The directory name is the namespace:
`run oneliners/main`, `bbs.screen("oneliners/list")`. Lua `require` loads only from the mod
directory. Module names match `[a-z0-9_]+(\.[a-z0-9_]+)*`. The sysop installs or removes a mod as a
directory and links it from a menu item. Oneliners, last callers and the poll are stock mods.
`api = 1` freezes after these three mods work.

**Runtime markup.** `print` markup can have colours and the fields of its `data` argument. It
cannot have includes. The host removes ESC and C0 characters from it.

**Suspend.** Each session has one goroutine. `key()`, `select()` and `input_line()` block until a
key, the timeout, the idle timeout or a hangup. Coroutines are not necessary. Each session
goroutine recovers from panics, so a panic ends one session and not the process.

**Limits.** A slice is the run time between two waits for input.
- Each session has one Lua state. Sessions never share a state. Each script run gets a fresh
  environment.
- Libraries: `base` (without `load`, `loadstring` and `dofile`), `string`, `table`, `math`.
  There is no `io`, `os`, `debug` or `coroutine`.
- CPU: 2 s wall clock for each slice, through the context of the Lua state.
- Size caps: 1 MB for the results of `string.rep`, `table.concat`, `string.format` and `gsub`, and
  for pattern subjects. The sandbox replaces these functions with capped versions.
- Memory: gopher-lua has no memory limit for each state. The host uses a heap guard: when the
  process heap passes the configured limit, the host aborts all running slices.
- An abort is sticky. A hangup, a limit or a guard abort sets the abort flag of the session and
  cancels the context of the Lua state. While the flag is set:
  - the VM raises the abort error at each instruction, so a pure Lua loop also stops;
  - every host call fails;
  - the sandbox versions of `pcall` and `xpcall` raise abort errors again.
- `sleep` counts as a wait. The session time limit applies inside every wait.
- After a script error, the user sees "System error, returning to menu". The frame pops. The sysop
  log gets the file, the line and the traceback.

## 6. Terminal and trust

**Trust tiers.**

| Tier | What | Treatment |
|---|---|---|
| 1 | User content: names, messages, input | Hostile. Always escaped and filtered. |
| 2 | Screens, art, theme, TOML files | Untrusted data. Safe to install without review. Clamped and filtered. |
| 3 | Lua scripts | Trusted code. Limits catch bugs only. The sysop reviews mods before install. |

**Charset.** All input files and all output are UTF-8. Input is normalised to NFC.

**Width.** One function, `cells(s)`, measures width:
- East Asian Wide and Fullwidth characters take 2 cells;
- East Asian Ambiguous characters take 1 cell (box drawing, blocks and most art symbols are in
  this class);
- combining marks take 0 cells, and a base character has 2 combining marks at most.

A terminal must show ambiguous characters as narrow. Truncation never splits a wide character.

**User content.** Values go into a template after parsing, so they cannot add tokens. The host
removes U+0000-001F, U+007F-009F and the bidi controls (U+202A-202E, U+2066-2069) from values and
log lines. Line input rejects control characters. Message bodies keep colour only (SGR). The host
removes cursor moves and all other sequences from them.

**Raw filter.** This filter applies to raw art. It keeps SGR, cursor moves, erase, save and
restore. It removes OSC, DCS, APC, PM, SOS and report requests.

**Clients.**
- Telnet: double IAC (0xFF), NAWS, TTYPE, bounded subnegotiation buffers.
- SSH: the size and TERM come from `pty-req` and `window-change`.
- The host clamps sizes to 20-255 columns and 10-100 rows.
- Each session has a bounded output queue. When the client does not drain it in 60 s, the host
  hangs up.

## 7. Sysop tools

**Files:**
- `screens/NAME.ans` and/or `NAME.txt`, plus an optional `NAME.toml`;
- `mods/<name>/`;
- `bbs.toml`, with `[theme]`;
- `sample.toml`.

**Errors.** Errors appear at load time. Each error gives the file, a 1-based row and column, and
the correction:

```
screens/main.toml:9: item "S": gosub sysop: screen "sysop" not found
screens/main.ans row 4 col 15: unknown field "user.nmae" (did you mean "user.name"?)
screens/main.ans row 7 col 3: token "{bar:sys}" changes colour: 11/1 at col 3, 15/1 at col 8
screens/msglist.ans row 5 col 4: list rectangle is not blank below the list row
screens/msglist.toml:4: expected "=" after key (TOML)
```

**Commands.**
- `bbs check` runs all load checks without a running BBS. For each mod, it lists the scripts and
  the `bbs.*` calls that write data.
- `bbs fields` lists the registry. `bbs fields --sample` writes `sample.toml` with sample values.
  The sysop adds named sample sets to it, for example one set with new mail and one without.
- `bbs preview NAME --set newmail --keys "DOWN,ENTER" --out x.ans` renders a screen. It writes an
  `.ans` file and a plain cell dump. The cell dump is the golden file for tests.

**Reload.** `bbs reload` or a file watch builds a new snapshot. If the snapshot has errors, the old
snapshot stays and the host reports the errors. Sessions change to the new snapshot at the next
screen.

## 8. Known limits

1. gopher-lua has no memory limit for each session. Doubling a string with `..` in a loop can make
   a memory spike before the heap guard acts. This is accepted for tier 3 code.
2. In art, blocks and lists have one row and no nesting.
3. There are no format filters. Date and number formats are set for each user.
4. The loader cannot check fields from Lua providers. A missing field renders empty and the host
   logs a warning.
5. Message bodies lose ANSI layout. Only the colour stays.
6. `api = 1` freezes after three stock mods only.

## 9. Check before you build

1. Moebius and PabloDraw: can they save art as UTF-8 ANSI? If not, a separate one-time converter
   is necessary. The BBS does not include it.
2. The art grid loader with real art: cursor moves, auto-wrap at the art width, iCE colour. Keep a
   corpus of real art files as loader tests from the first week.
3. Ambiguous-width characters in common terminals (xterm, Windows Terminal, PuTTY, iTerm2, macOS
   Terminal).
4. gopher-lua at the chosen version:
   - an abort stops `while true do pcall(f) end`, where `f` makes no host call;
   - the wall clock and long single operations (issue yuin/gopher-lua#521);
   - the heap guard;
   - long calls inside host code.
5. Hangup and reload while a script waits for input.

## 10. Backlog

- Node messaging. Phase 2 adds `wait(opts)` with the events `{type = "key"}`, `{type = "timeout"}`
  and `{type = "msg", from = {node, user, mod}, data}`. `key(timeout)` does not change. Rules for
  `node.send`:
  - the host sets `from`, and the script cannot set it;
  - the host delivers a message only to a session that waits in the same mod;
  - each session has an inbox of 32 messages, and each payload is 4 KB at most;
  - when the inbox is full, `node.send` does not deliver and returns the delivered count.
- `msg` option for `select` and `input_line`, and a two-node demo game.
- `gotoxy`, `users.find`, `user_set` (profile fields only, never level, flags or password).
- `line` input mode (typed commands).
- Web client.
