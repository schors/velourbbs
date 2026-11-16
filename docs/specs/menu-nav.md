# Menu and navigation

This specification defines the key match, the navigation of a session,
the menu checks between screens of `velourbbs check` and the key sequences of `velourbbs preview`.
[bbt-format](bbt-format.md), section 8, defines the menu items and the keys.
Related parts: [screens-toml](screens-toml.md), [screen-render](screen-render.md), [file-view](file-view.md)
and [telnet-term](telnet-term.md).

The [BBT guide](bbt-guide.md) teaches this specification. Keep the guide in sync:
[bbt-guide](bbt-guide.md), section 3.

## 1. Goal

A user goes from one screen to other screens with hotkeys and ends the session.
All navigation is explicit. The host has no implicit transitions.
All sessions are anonymous. Login is a later idea.

## 2. Concepts

- A "menu", an "item" and an "enabled" item: [bbt-format](bbt-format.md), section 8.
- The "current screen" is the one screen of the session that waits for a key.
  There is no navigation stack.

## 3. Keys

### 3.1. Key names

[bbt-format](bbt-format.md), section 8.1, defines the key names.
Other characters are keys too, but an item cannot have them. They match only `ANY`.

- Case does not matter in `--keys`, as in `key`.
- [telnet-term](telnet-term.md), section 8, decodes the input bytes to these names
  and to characters.

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

## 4. Actions

The host runs the actions of [bbt-format](bbt-format.md), section 8.3:

- `goto` to the current screen renders it again, for example after a resize.
  With `args`, it shows the screen with the new arguments.
- A `page` action shows a page of the current file screen (section 5.1).
- `logoff` shows `screens/goodbye.bbt` if it exists. Then the host disconnects.
  The host does not wait for a key after the last page of `goodbye`.

## 5. Session

- The first screen is `[menu] start` ([screens-toml](screens-toml.md), section 3.4).
- The host shows a screen in these steps. Section 5.1 has the steps of a file screen.
  1. Read `term.cols` and `term.rows`. A resize takes effect at the next screen.
  2. Evaluate the `when` of each item.
  3. Render the screen ([screen-render](screen-render.md), section 2).
  4. Clear the terminal screen and show the first page.
  5. After each page with a More row, wait for any key. Then clear the terminal screen
     and show the next page. Items do not work on a More row.
  6. After the last page, wait for keys and match them (section 3.3).
- Keys that arrive during a render stay in the input queue. A More row uses the next key
  of the queue.
- A screen whose items are all disabled and that has no `ANY` item ignores all keys.
  The idle timeout of [telnet-term](telnet-term.md) ends such a session.
- A hangup or the idle timeout ends the session at once, without `goodbye`.

### 5.1. File screen

The host shows a file screen ([file-view](file-view.md), section 3) in these steps:

1. Read `term.cols` and `term.rows`. Lay out the file into file pages.
   The size and the file pages stay until the next screen.
2. For the current page, evaluate the `when` of each item, render the page and
   clear the terminal screen and show it. The first page is page 1.
3. Wait for keys and match them (section 3.3). A `page` action sets the current page
   and continues at step 2. Each other action leaves the screen.

## 6. `velourbbs check`

`velourbbs check` ([screen-render](screen-render.md), section 3.1) finds the item errors
of [bbt-format](bbt-format.md), section 8.4. It also finds these errors:

- a `goto` to a screen that does not exist;
- a `goto` to a screen without items, for example `goodbye`;
- a screen without items, except `goodbye`;
- items in `goodbye`;
- a file screen as `goodbye` ([file-view](file-view.md), section 3);
- a start screen that does not exist or has no items;
- a `goto` whose `args` differ from the arguments of the target screen
  ([bbt-format](bbt-format.md), section 8.5);
- arguments of the start screen or of `goodbye`.

- An error in an item has the position of the first line
  of the front matter ([screen-render](screen-render.md), section 3.1).
  It starts with `item "<KEY>"`, with the key in upper case.
  An item without a valid key gives `item <n>`, the 1-based number of the item.
- An error of a screen without items has the position 1:1.
- An error of the start screen has the position 1:1 in `screens/screens.toml`,
  also when the file is missing.
- The checks of a `goto`, of a screen without items and of the start screen use all screens.
  `velourbbs check` does them only when the files have no other errors.

Examples:

```
screens/main.bbt:2:1: item "S": do: unknown screen "sysinf"
screens/main.bbt:2:1: item 2: key: invalid key "?"
screens/main.bbt:2:1: item "S": key: duplicate key
screens/main.bbt:2:1: item 3: key is missing
screens/about.bbt:1:1: screen has no items
screens/screens.toml:1:1: menu: start: unknown screen "mian"
```

## 7. `velourbbs preview --keys`

`velourbbs preview <name> --keys <keys>` runs a session from the screen `<name>` without telnet.
It is the test harness for the navigation.

- `<keys>` is a list of keys, separated with commas. A key is a key name
  of [bbt-format](bbt-format.md), section 8.1, except `ANY`, or one other printable character
  except the comma and the space.
  Example: `S,ы,ESC,G`. An empty list is valid.
- `<name>` must be a valid start screen (section 6), except the rule on arguments.
  Else it is an invalid argument. `--set arg.<name>` gives the arguments of the screen
  ([screen-render](screen-render.md), section 3.2).
- The other arguments of `velourbbs preview` ([screen-render](screen-render.md), section 3.2) apply.
- The output has one part for each event:
  - `screen <name>` and the cell dump of the first page, when the host shows a screen;
  - the cell dump of the next page, after a key on a More row;
  - the cell dump of the page, after a `page` action;
  - `key <key>` before each key. `key <key> ignored` for an ignored key;
  - `logoff` when an item runs `logoff`. The pages of `goodbye` follow it.
- The output gives a key after the map of section 3.2. A letter, a digit or a key name
  is in upper case. Each other character is as it is. Example: `ы` gives `key S`.
- When the list ends, `velourbbs preview` stops after the current page and exits with status 0.
- A key after the end of the session at `logoff` (section 4) is an error.
  `velourbbs preview` exits with status 1.

Example: `velourbbs preview main --keys S,X,G --size 80x25`, with the screens
of [bbt-format](bbt-format.md), section 8:

```
screen main
page 1 80x25 fill none
...
key S
screen sysinfo
page 1 80x25 fill none
...
key X
screen main
page 1 80x25 fill none
...
key G
logoff
screen goodbye
page 1 80x25 fill none
...
```

## 8. Code

| Package | Content |
|---|---|
| `internal/key` | The key names, the parse of `key` and the Russian layout map. |
| `internal/bbt` | The parse of `key` and `do` and the checks of one file. |
| `internal/menu` | The checks of section 6 between files, and the session of section 5 with a terminal interface. |

- The model (`screen.Item`) holds the key in upper case and the parsed action.
  `internal/screen` uses `internal/key` for the case rule of `item.<KEY>`.
- The terminal interface gives the size and the keys and shows a page.
  `velourbbs preview` and telnet-term give its implementations.
- A terminal can also take the other events of section 7: the screen name, a key,
  an ignored key and `logoff`. The session sends them only to a terminal that implements
  this second interface. `velourbbs preview` implements it.

## 9. Not in the MVP

- `gosub`, `back`, `top` and a navigation stack: [menu-actions](../ideas/menu-actions.md).
- Typed command input.
- `page` for lists, and the actions `run` and `call`. They need lists, Lua or built-ins.
