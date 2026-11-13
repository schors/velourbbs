# Menu and navigation

This specification defines the menu items, the keys, the navigation of a session,
the menu checks of `bbs check` and the key sequences of `bbs preview`.
Related parts: [screen-format](screen-format.md) and [telnet-term](../ideas/telnet-term.md).

## 1. Goal

A user goes from one screen to other screens with hotkeys and ends the session.
All navigation is explicit. The host has no implicit transitions.
All sessions are anonymous. Login is a later idea.

## 2. Concepts

- A "menu" is the `[[item]]` list in the front matter of a screen
  ([screen-format](screen-format.md), section 3.2).
- An "item" is one choice: `key`, `do` and an optional `when`.
  An item has no position on the screen. The screen author writes the item text,
  for example `[A]bout`, into the body.
- The "current screen" is the one screen of the session that waits for a key.
  There is no navigation stack.

Example `screens/main.bbt`:

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

Example `screens/sysinfo.bbt`:

```
+++
[[item]]
key = "ANY"
do = "goto main"
+++
Terminal: {term.cols}x{term.rows}. Press any key.
```

## 3. Keys

### 3.1. Key names

A key is one of these:

- a Latin letter `A`–`Z`;
- a digit `0`–`9`;
- a named key: `ESC`, `ENTER`, `SPACE`, `TAB`, `BS`, `DEL`, `UP`, `DOWN`, `LEFT`, `RIGHT`,
  `HOME`, `END`, `PGUP`, `PGDN`, `F1`–`F12`;
- `ANY`. It matches each key that has no enabled item of its own.

Other characters are keys too, but an item cannot have them. They match only `ANY`.

- Case does not matter in `key`, in `item.<KEY>` and in `--keys`. `s`, `S` and `esc`, `ESC`
  are the same keys.
- [telnet-term](../ideas/telnet-term.md) decodes the input bytes to these names
  and to characters. Both 0x08 and 0x7F give `BS`.

### 3.2. Russian layout

The host maps a Cyrillic letter to the Latin letter on the same key of the ЙЦУКЕН layout,
in both cases:

```
й q  ц w  у e  к r  е t  н y  г u  ш i  щ o  з p
ф a  ы s  в d  а f  п g  р h  о j  л k  д l
я z  ч x  с c  м v  и b  т n  ь m
```

The letters `х ъ ж э ё б ю` have no Latin letter. They match only `ANY`.

### 3.3. Key match

When a key arrives at the current screen, the host:

1. maps the key with section 3.2;
2. runs the enabled item with this key, if there is one;
3. else runs the enabled `ANY` item, if there is one;
4. else ignores the key.

ESC has no implicit action.

## 4. Items

| Key | Type | Meaning |
|---|---|---|
| `key` | string | The key of the item. Section 3.1. |
| `do` | string | The action. Section 5. |
| `when` | string | Optional. A condition ([screen-format](screen-format.md), section 5). |

- An item is "enabled" when it has no `when` or its `when` is true.
  A disabled item counts as absent, so `ANY` can match its key.
- The host evaluates each `when` once for each render of a screen.
  So the shown text and the enabled keys always agree.
- The condition path `item.<KEY>` is true when the item with the key `<KEY>` is enabled
  ([screen-format](screen-format.md), section 4.1). A template hides the text of a disabled
  item with `{#if item.<KEY>}`.

## 5. Actions

| `do` | Action |
|---|---|
| `goto <name>` | Show the screen `<name>`. |
| `logoff` | End the session. |

- The action words are lower case. `do` has no other spaces than the one space after `goto`.
- `<name>` matches the screen name of [screen-format](screen-format.md), section 2.
- `goto` to the current screen renders it again, for example after a resize.
- `logoff` shows `screens/goodbye.bbt` if it exists. Then the host disconnects.
  The host does not wait for a key after the last page of `goodbye`.

## 6. Session

- The first screen is `main`. `[menu] start` in `bbs.toml` can set another name.
  It matches the screen name of [screen-format](screen-format.md), section 2.
- The host shows a screen in these steps:
  1. Read `term.cols` and `term.rows`. A resize takes effect at the next screen.
  2. Evaluate the `when` of each item.
  3. Render the screen ([screen-format](screen-format.md), section 6),
     or its narrow screen (section 6.1 there).
  4. Clear the terminal screen and show the first page.
  5. After each page with a More row, wait for any key. Then clear the terminal screen
     and show the next page. Items do not work on a More row.
  6. After the last page, wait for keys and match them (section 3.3).
- Keys that arrive during a render stay in the input queue. A More row uses the next key
  of the queue.
- A screen whose items are all disabled and that has no `ANY` item ignores all keys.
  The idle timeout of [telnet-term](../ideas/telnet-term.md) ends such a session.
- A hangup or the idle timeout ends the session at once, without `goodbye`.

`bbs.toml` example:

```
[menu]
start = "welcome"
```

## 7. `bbs check`

`bbs check` ([screen-format](screen-format.md), section 9.1) also finds these errors:

- an invalid key;
- a duplicate key in a menu, also `ANY`;
- an invalid `do`: an unknown action, or a wrong argument of the action;
- an `item.<KEY>` path without an item in the screen
  ([screen-format](screen-format.md), section 6.1, has the rule for a narrow screen);
- a `goto` to a screen that does not exist;
- a `goto` to a screen without items, for example a narrow screen or `goodbye`;
- a screen without items, except `goodbye` and the narrow screens;
- items in `goodbye`;
- a start screen that does not exist or has no items.

- An error of an item starts with `item "<KEY>"` when its key is valid, with the key
  in upper case. Else it starts with `item <n>`, the 1-based number of the item.
- An error of an item has the position of the key with the error.
  An error of a screen without items has the position 1:1.
- An error of the start screen has the position of `start`.
  When `bbs.toml` has no `start`, the position is 1:1.

Examples:

```
screens/main.bbt:4:1: item "S": do: unknown screen "sysinf"
screens/main.bbt:7:1: item 2: key: invalid key "?"
screens/main.bbt:7:1: item "S": key: duplicate key
screens/about.bbt:1:1: screen has no items
bbs.toml:2:1: menu: start: unknown screen "mian"
bbs.toml:1:1: menu: start: unknown screen "main"
```

## 8. `bbs preview --keys`

`bbs preview <name> --keys <keys>` runs a session from the screen `<name>` without telnet.
It is the test harness for the navigation.

- `<keys>` is a list of keys, separated with commas. A key is a key name of section 3.1
  except `ANY`, or one other printable character except the comma and the space.
  Example: `S,ы,ESC,G`. An empty list is valid.
- `<name>` must be a valid start screen (section 7). Else it is an invalid argument.
- The other arguments of `bbs preview` ([screen-format](screen-format.md), section 9.2) apply.
- The output has one part for each event:
  - `screen <name>` and the cell dump of the first page, when the host shows a screen.
    `<name>` is the rendered screen: the narrow screen, when the host shows it;
  - the cell dump of the next page, after a key on a More row;
  - `key <key>` before each key. `key <key> ignored` for an ignored key;
  - `logoff` when an item runs `logoff`. The pages of `goodbye` follow it.
- The output gives a key after the map of section 3.2. A letter, a digit or a key name
  is in upper case. Each other character is as it is. Example: `ы` gives `key S`.
- When the list ends, `bbs preview` stops after the current page and exits with status 0.
- A key after the last page of `goodbye` is an error. `bbs preview` exits with status 1.

Example: `bbs preview main --keys S,X,G --size 80x25`:

```
screen main
page 1/1 80x25
...
key S
screen sysinfo
page 1/1 80x25
...
key X
screen main
page 1/1 80x25
...
key G
logoff
screen goodbye
page 1/1 80x25
...
```

## 9. Code

| Package | Content |
|---|---|
| `internal/config` | `[menu]` in `bbs.toml`. |
| `internal/key` | The key names, the parse of `key` and the Russian layout map. |
| `internal/bbt` | The parse of `key` and `do` and the checks of one file. |
| `internal/menu` | The checks of section 7 between files, and the session of section 6 with a terminal interface. |

- The import order is in [screen-format](screen-format.md), section 11.
- The model (`screen.Item`) holds the key in upper case and the parsed action.
  `internal/screen` uses `internal/key` for the case rule of `item.<KEY>`.
- The terminal interface gives the size and the keys. The session tells it each event
  of section 8: a page with the name of the rendered screen, a key, an ignored key and `logoff`.
  `bbs preview` and telnet-term give its implementations.

## 10. Not in the MVP

- `gosub`, `back`, `top` and a navigation stack: [menu-actions](../ideas/menu-actions.md).
- `require` (access checks). It needs users.
- Arrow-key selection and marks: [menu-actions](../ideas/menu-actions.md).
- Typed command input.
- `args`, `ask` and `{arg.*}`.
- The actions `page`, `run` and `call`. They need lists, Lua or built-ins.
- Punctuation keys, for example `?`.
- A `label` for an item. The host does not print item text.
