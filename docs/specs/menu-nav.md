# Menu and navigation

This specification defines the key match, the navigation of a session,
the menu checks between screens of `velourbbs check` and the key sequences of `velourbbs preview`.
[bbt-format](bbt-format.md), section 8, defines the menu items and the keys.
Related parts: [screens-toml](screens-toml.md), [screen-render](screen-render.md), [file-view](file-view.md),
[list-view](list-view.md), [session-language](session-language.md), [live-screen](live-screen.md)
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
- The "navigation stack" is the list of the frames of the session (section 4.1).
  A "frame" holds a screen, its arguments, the page of a file screen or of a list screen
  and the cursor of a list screen.
  The "first frame" is at the bottom of the stack. The "top frame" holds the current screen.

## 3. Keys

### 3.1. Key names

[bbt-format](bbt-format.md), section 8.1, defines the key names.
Other characters are keys too, but an item cannot have them. They match only `ANY`.

- Case does not matter in a key name of `--keys`, as in `key`. A letter is a key name,
  except in the line input (section 5.4): there it is a character with its case.
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
2. in a list screen with `select`, handles `UP`, `DOWN` and `ENTER` ([list-view](list-view.md), section 8)
   and stops;
3. runs the enabled item with this key, if there is one;
4. else runs the enabled `ANY` item, if there is one;
5. else ignores the key.

An item with `ask` first reads a line (section 5.4).

ESC has no implicit action.

## 4. Actions

The host runs the actions of [bbt-format](bbt-format.md), section 8.3:

- `goto`, `gosub`, `back` and `top` change the navigation stack (section 4.1).
- A `page` action shows a page of the current file screen (section 5.1) or list screen (section 5.3).
- `select` of a list screen runs its action as an item runs it ([list-view](list-view.md), section 8).
- `theme <choice>` sets the theme choice of the session to `<choice>`
  ([screens-toml](screens-toml.md), section 3.2). The choice lasts until the end of the session.
  Then the host shows the screen of the top frame again (section 5). The stack does not change.
  In a file screen, it shows the current page again (section 5.1).
  A list screen shows the page and the cursor of the frame (section 5.3).
- `goto <code>/<name>` and `gosub <code>/<name>` set the session language
  ([session-language](session-language.md), section 5) to `<code>` until the end of the session.
  Then they run `goto <name>` or `gosub <name>`. The new frame holds `<name>`, without `<code>`.
- `lang <code>` sets the session language to `<code>` until the end of the session. Then the host shows the screen of the top frame
  again from the screen set of the new language (section 5). The stack does not change.
  A file screen shows the page of the frame (section 5.1).
  A list screen shows the page and the cursor of the frame (section 5.3).
- `logoff` shows the screen `goodbye` if it exists ([session-language](session-language.md), section 4).
  Then the host disconnects.
  The host does not wait for a key after the last page of `goodbye`.

### 4.1. Navigation stack

- The stack has 16 frames at most.
- `goto <name>` replaces the top frame with a new frame: the screen `<name>`,
  the values of the `args` of the item, page 1 and the cursor on the first row.
- `gosub <name>` adds a new frame on top: the screen `<name>`, the values of the `args` of the item, page 1
  and the cursor on the first row.
  When the stack has 16 frames, the host first removes the second frame:
  the oldest frame above the first frame.
- The host takes the values of `args` before it changes the stack
  ([bbt-format](bbt-format.md), section 8.5): `{arg.<name>}` from the top frame,
  `{in}` from the line of `ask`, `{row.<field>}` from the selected row.
- `back` removes the top frame. On the first frame, it removes nothing.
- `top` removes each frame above the first frame.
- After each of these actions, the host shows the screen of the top frame
  with its arguments (section 5). A file screen shows the page of the frame (section 5.1).
  A list screen shows the page and the cursor of the frame (section 5.3).
- A frame holds the name of the screen. The host finds the file of the name
  in the session language each time it shows the frame ([session-language](session-language.md), section 4).
- A `goto` to the current screen also gives a new frame with page 1.
  The host reads the terminal size again (section 5).
- A `goto` on the first frame replaces the first frame. Then `top` shows its screen,
  not the start screen.

Example: `main` runs `gosub texts`, `texts` runs `gosub reader` and the user goes to page 3.
The stack is `main`, `texts`, `reader` on page 3. `back` shows `texts`. A second `back`
shows `main`. From `reader`, `top` shows `main`. From `reader`, `goto main` gives the stack
`main`, `texts`, `main`. `back` then shows `texts`.

## 5. Session

- The first screen is `[menu] start` ([screens-toml](screens-toml.md), section 3.4).
  The session starts with one frame: the start screen with no arguments.
- The host shows a screen in these steps. Section 5.1 has the steps of a file screen.
  Section 5.3 has the steps of a list screen.
  1. Read `term.cols` and `term.rows`. A resize takes effect at the next screen.
  2. Evaluate the `when` of each item.
  3. Render the screen ([screen-render](screen-render.md), section 2).
  4. Clear the terminal screen and show the first page.
  5. After each page with a More row, wait for any key. Then clear the terminal screen
     and show the next page. Items do not work on a More row.
  6. After the last page, wait for keys and match them (section 3.3).
- Keys that arrive during a render stay in the input queue. A More row uses the next key
  of the queue.
- While the shown page waits for keys, the ticks and the effects of [live-screen](live-screen.md) run.
- A screen whose items are all disabled and that has no `ANY` item ignores all keys.
  The idle timeout of [telnet-term](telnet-term.md) ends such a session.
- A hangup or the idle timeout ends the session at once, without `goodbye`.

### 5.1. File screen

The host shows a file screen ([file-view](file-view.md), section 3) in these steps:

1. Read `term.cols` and `term.rows`. Lay out the file into file pages.
   The size and the file pages stay until the next screen.
2. For the current page, evaluate the `when` of each item, render the page and
   clear the terminal screen and show it. The first page that the host shows is the page
   of the top frame (section 4.1), or the last file page when the file has fewer pages.
   The host then sets the page of the top frame to the shown page.
3. Wait for keys and match them (section 3.3). A `page` action sets the current page
   and the page of the top frame, and continues at step 2. A `theme` action continues at step 2.
   Each other action leaves the screen.

### 5.2. Stop

When the host stops ([telnet-term](telnet-term.md), section 10.2), it shows
the screen `shutdown` to each open session if it exists
([session-language](session-language.md), section 4). Then it disconnects.

- The host renders `shutdown` with the size of the current screen and shows its first page.
  It does not wait for a key.
- Without the file, the host disconnects with no page.
  [telnet-term](telnet-term.md), section 8.4, can send `ESC[?2004l` before.

### 5.3. List screen

The host shows a list screen ([list-view](list-view.md), section 3) in these steps:

The current list page and the cursor are the page and the cursor of the top frame.

1. Read `term.cols` and `term.rows`. Find the page size ([list-view](list-view.md), section 7).
   The size and the page size stay until the next screen.
2. When the list has fewer pages than the current page, the current page is the last list page.
   When the cursor is not on the current page, it is on the last row of the page.
3. Evaluate the `when` of each item, render the current page,
   clear the terminal screen and show it.
4. Wait for keys and match them (section 3.3).
   - A `page` action or a cursor move to another list page ([list-view](list-view.md), section 8)
     continues at step 3. A `theme` action continues at step 3.
   - A cursor move on the same list page renders the page with the values of step 3
     and updates it ([telnet-term](telnet-term.md), section 7.5). Continue at step 4.
     On a live page, the slots have the values of the last tick ([live-screen](live-screen.md), section 4).
   - Each other action leaves the screen.

### 5.4. Ask

When an item with `ask` ([bbt-format](bbt-format.md), section 8.6) runs, the host:

1. Takes the "ask row": the last row of the shown page that the terminal shows.
   A page cut at `term.rows` has the row `term.rows`. A page without rows gets one row:
   the row where a page of 1 row starts ([bbt-format](bbt-format.md), section 6.5).
2. Shows the page with the ask row in place of that row.
   The ask row has the cells of the row outside the screen area. In the screen area,
   it has the prompt in the theme style `more`, then M + 1 spaces in the theme style `select`,
   then spaces in `normal`. The host cuts the row at `term.cols`.
3. Reads a line with the line input ([line-input](line-input.md)): the field is the M + 1 spaces
   of step 2, in `select`, not masked.
4. After `ENTER` with a line that is not empty, runs the action of the item
   with the line as `{in}`.
5. After `ESC` or `ENTER` with an empty line, shows the page again and waits for keys:
   step 6 of section 5, step 3 of section 5.1, step 4 of section 5.3.

- M is the free cells of the prompt ([bbt-format](bbt-format.md), section 8.6),
  in the screen area that the terminal shows, and at most `ask_max`.
- When `term.cols` < `min`, M can be less than 1. Then the host cuts the end of the prompt
  at a grapheme boundary until M is at least 1. `ask_max` applies after the cut.
- The key of the item does not go into the line. The keys after it in the input queue do.
- The line input clears the key queue at its end ([line-input](line-input.md), section 6).
  The `--keys` list of `velourbbs preview` is not a queue: the clear drops no key of it (section 7).

Example: the screen `area` at 80x25 ends with an empty row 24. `J` of [bbt-format](bbt-format.md),
section 8.6, shows `Message number: ` and 6 cells of the field on row 24.
The user types `12` and `ENTER`. The host runs `gosub msg` with `n = "12"`.

## 6. `velourbbs check`

`velourbbs check` ([screen-render](screen-render.md), section 3.1) finds the item errors
of [bbt-format](bbt-format.md), section 8.4. It also finds these errors in the screen set
of each language of the BBS ([session-language](session-language.md), section 6):

- a `goto` or a `gosub` to a screen that does not exist;
- a `goto` or a `gosub` to a screen without items, for example `goodbye` or `shutdown`;
- a screen without items, except `goodbye` and `shutdown`;
- items in `goodbye` or `shutdown`;
- a file screen or a list screen as `goodbye` or `shutdown` ([file-view](file-view.md), section 3,
  [list-view](list-view.md), section 3);
- a start screen that does not exist or has no items;
- a `goto` or a `gosub` whose `args` differ from the arguments of the target screen
  ([bbt-format](bbt-format.md), section 8.5);
- an error of the values of [bbt-format](bbt-format.md), section 8.5: an open value
  or a `{row.<field>}` value that reaches `file` or `data`;
- arguments of the start screen, of `goodbye` or of `shutdown`.

- An error in an item has the position of its key path ([screen-render](screen-render.md), section 3.1).
  It starts with `item "<KEY>"`, with the key in upper case.
  An item without a valid key gives `item <n>`, the 1-based number of the item.
- An error of a screen without items has the position 1:1.
- An error of the start screen is an error of the key `menu.start` of `screens/screens.toml`,
  also when the file is missing.
- The checks of a `goto` and a `gosub`, of the values, of a screen without items and of the start screen use all screens.
  `velourbbs check` does them only when the files have no other errors.

Examples:

```
screens/main.bbt:8:6: item "S": do: unknown screen "sysinf"
screens/main.bbt:11:7: item 2: key: invalid key "?"
screens/main.bbt:14:7: item "S": key: duplicate key
screens/main.bbt:16:1: item 4: key is missing
screens/about.bbt:1:1: screen has no items
screens/screens.toml:12:9: menu: start: unknown screen "mian"
```

### 6.1. Unreachable screen

`velourbbs check` also warns about each screen that a user cannot reach:
`screens/old.bbt:1:1: warning: a user cannot reach this screen`.

- The check walks pairs of a language and a screen name. The walk starts at the start screen
  in `[screen] lang` ([session-language](session-language.md), section 5).
- The "actions" of a screen are the action of each item and the action of its `select`.
- From a reached pair of the language L and the screen S, the walk reaches:
  - the target of each `goto` and `gosub` of the actions of S:
    the language `<code>` and the screen `<name>` for `goto <code>/<name>` and `gosub <code>/<name>`,
    else L and `<name>`;
  - the language `<code>` and S, for each `lang <code>` of the actions of S;
  - the language M and S, when the walk from the target of a `gosub` of the actions of S
    reaches a pair of the language M. S stays on the stack below the target,
    and `back` shows S again in the session language (section 4.1).
- The walk ignores `when`: each item is a path. It ignores the limit of 16 frames.
  It does not need a `back` item. So the check can miss a screen that a user cannot reach,
  but it does not warn about a screen that a user can reach.
- Each pair that the walk does not reach gives the warning at 1:1 of the file of the screen
  in the screen set of the language, except `goodbye` and `shutdown`.
  The language suffix of [session-language](session-language.md), section 6, applies.
- The warning does not change the exit status. The check runs only when the files,
  the other checks of this section and the checks of [session-language](session-language.md),
  section 6, give no error.

Example with the languages `en` and `ru`, `[screen] lang = "en"` and the start screen `main`.
The common screen `main` runs `gosub en/menu` and `gosub ru/menu`.
`screens/en/menu.bbt` runs `gosub news`. `screens/ru/menu.bbt` has no item for `news`:

- `main` in `ru` gets no warning: the walk from `ru/menu` reaches `ru`,
  and `back` from `ru/menu` shows `main` in `ru`.
- `screens/ru/news.bbt:1:1: warning: a user cannot reach this screen`.
  With a common file `screens/news.bbt`, the warning is
  `screens/news.bbt:1:1: warning: a user cannot reach this screen (language "ru")`.

## 7. `velourbbs preview --keys`

`velourbbs preview <name> --keys <keys>` runs a session from the screen `<name>` without telnet.
It is the test harness for the navigation.

- `<keys>` is a list of keys, separated with commas. A key is a key name
  of [bbt-format](bbt-format.md), section 8.1, except `ANY`, or one other printable character
  except the comma and the space.
  Example: `S,ы,ESC,G`. An empty list is valid.
  An element `+<duration>` is a wait, not a key ([live-screen](live-screen.md), section 7). Example: `S,+2s,G`.
- `<name>` must be a valid start screen (section 6), except the rule on arguments.
  Else it is an invalid argument. The session starts with one frame: this screen.
  `--set arg.<name>` gives the arguments of the screen ([screen-render](screen-render.md), section 3.2).
- The other arguments of `velourbbs preview` ([screen-render](screen-render.md), section 3.2) apply.
- The output has one part for each event:
  - `screen <name>` and the cell dump of the first shown page, when the host shows a screen,
    also after `back`, `top` and `lang`;
  - the cell dump of the next page, after a key on a More row;
  - the cell dump of the page, after a `page` action or a `theme` action in a file screen
    or a list screen, and after a cursor move in a list screen;
  - `key <key>` before each key. `key <key> ignored` for an ignored key;
  - `wait <duration>` and the grid dump after each wait ([live-screen](live-screen.md), section 7);
  - `logoff` when an item runs `logoff`. The pages of `goodbye` follow it;
  - `ask` and the cell dump of the page with the ask row, when an item with `ask` runs
    (section 5.4);
  - `key <key>` before each key of the line input, without the map of section 3.2 and without `ignored`.
    A character is as it is, with its case;
  - `line "<line>"` after `ENTER` with a line, and `line cancelled` after `ESC`
    or `ENTER` with an empty line. Before `"` and `\` in `<line>`, the output has a `\`.
    After `line cancelled`, the cell dump of the page follows.
- With `--view`, each page is in the line form ([screen-render](screen-render.md), section 3.7)
  in place of the cell dump. The other lines stay as they are.
- Outside the line input, the output gives a key after the map of section 3.2. A letter, a digit or a key name
  is in upper case. Each other character is as it is. Example: `ы` gives `key S`.
- When the list ends, `velourbbs preview` stops after the current page and exits with status 0.
- A key after the end of the session at `logoff` (section 4) is an error.
  `velourbbs preview` exits with status 1.

Example: `velourbbs preview main --keys S,X,G --size 80x25`, with the screens
of [bbt-format](bbt-format.md), section 8. `S` runs `gosub sysinfo`, `X` runs `back`:

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
| `internal/bbs` | The checks of section 6 between files ([session](session.md), section 3). |
| `internal/menu` | The session of section 5 with a terminal interface ([session](session.md), section 4). |

- The model (`screen.Item`) holds the key in upper case and the parsed action.
  `internal/screen` uses `internal/key` for the case rule of `item.<KEY>`.
- The terminal interface gives the size and the keys, shows a page and updates a page
  (section 5.3, [live-screen](live-screen.md), section 5).
  The wait for a key takes the context of the session ([session](session.md), section 6)
  and a deadline: the time of the next frame of a live page ([live-screen](live-screen.md), section 5).
  At the deadline, it gives no key. A More row and the line input wait without a deadline.
  The terminal also gives the clock of the ticks: the server clock, or the fixed clock
  of `velourbbs preview` ([live-screen](live-screen.md), section 7).
  `velourbbs preview` and telnet-term give its implementations.
- The terminal interface also reads a line with the line input (section 5.4,
  [line-input](line-input.md), section 8).
- A terminal can also take the other events of section 7: the screen name, a key,
  an ignored key, `logoff`, `ask` and the end of a line. The session sends them only to a terminal that implements
  this second interface. `velourbbs preview` implements it.

## 9. Not in the MVP

- Typed command input.
- A theme choice that lasts after the session. It needs users and login.
- The actions `run` and `call`. They need Lua or built-ins.
