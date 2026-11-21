# Task 5. Fields and conditions

In this task, you write the screen `term`. It shows what the BBS knows about the terminal of the user.

## Fields

A "field" is a token that the BBS replaces with a value:

```
       {!dim}Size{!}     {term.cols}×{term.rows}
```

On a terminal of 80 columns and 24 rows, this line shows `Size     80×24`.

| Field | Value |
|---|---|
| `term.cols` | The columns of the terminal, from 20 to 255. |
| `term.rows` | The rows of the terminal, from 10 to 100. |
| `term.colors` | The colour depth: `16`, `256` or `16777216` (task 6). |
| `term.type` | The terminal type that the client sends, in lower case. Without a type, `ansi`. |
| `term.name` | The name and the version that the terminal reports, for example `XTerm(412)`. Can be empty. Use it only to show it. |
| `term.ambiguous` | `2` when the terminal shows East Asian Ambiguous characters, for example Cyrillic letters and `■`, in two cells. Else `1`. The layout does not change, so frames and columns break. Use it to show the user how to turn this setting off. |
| `term.bg_color` | The background colour that the terminal reports, as `#rrggbb` in lower case. Empty when the terminal does not report it. |
| `term.background` | `light` when `term.bg_color` is a light colour, for example `#bcbcbc` or lighter grey. `dark` for a dark colour. Empty when `term.bg_color` is empty. |

- A field without spaces in its braces takes the width of its value.
- A field with spaces in its braces has a "fixed width": the width of the token with the braces.
  The BBS pads a short value with spaces and cuts the end of a long value.
  Spaces after the name align the value left. Spaces before the name align it right.

```
[{term.type      }]
[{      term.type}]
```

Each token is 17 cells wide. The BBS shows:

```text
[xterm            ]
[            xterm]
```

- `{` starts a token when a Latin letter, a space or an ASCII punctuation character follows it,
  except `{` and `}`. A punctuation character other than `#`, `/`, `!` and `>` is reserved:
  its token is an error. `{{` gives a plain `{`. Example: `Write {{name} here.` shows `Write {name} here.`

## Session and BBS fields

Other fields show the data of the session and of the BBS:

```
{bbs.name} {bbs.version}  {bbs.time}
Your address {session.ip}, online {session.online}
Users online {bbs.sessions}, up {bbs.uptime}
```

| Field | Value |
|---|---|
| `session.ip` | The address of the user. Example: `203.0.113.7`. |
| `session.theme` | The theme choice of the session: `dark`, `light` or `auto` (task 6). |
| `session.lang` | The language of the session, for example `en` (task 7). |
| `session.online` | The time from the connect of the user: `hh:mm:ss`, from 24 hours `<d>d hh:mm:ss`. |
| `bbs.name` | The key `name` of `[bbs]` in `bbs.toml`. The default is `Kind Brothers BBS`. |
| `bbs.version` | The version of `velourbbs`, for example `v1.2.0`. A build that is not of a release tag gives `dev`. |
| `bbs.sessions` | The users online now, with this user. |
| `bbs.time` | The date and the time of the server, in `[screen] zone`. |
| `bbs.uptime` | The time from the start of `velourbbs serve`, in the form of `session.online`. |

- The BBS takes the values when it shows the screen. A value on the screen does not change
  until the next screen. A live field is the exception (below).
- `[screen] time` in `screens/screens.toml` sets the form of `bbs.time` with the Go reference time
  `2006-01-02 15:04:05`. The default is `"2006-01-02 15:04:05"`. Example for `16.11.2026 15:04`:

```toml
[screen]
time = "02.01.2006 15:04"
```

- `[screen] zone` sets the time zone of `bbs.time`: `"Local"`, `"UTC"` or a zone name
  such as `"Europe/Berlin"`. The default `"Local"` is the zone of the server.
  `MST` in `time` shows the zone. Example for `14:05 UTC`:

```toml
[screen]
time = "15:04 MST"
zone = "UTC"
```

- `velourbbs preview` has no session. It shows `127.0.0.1`, `00:00:00`, `1` users
  and the time `2006-01-02 15:04:05` in `[screen] zone`, with UTC for `"Local"`. `--set` changes a value, for example
  `--set bbs.sessions=12`.

Full rules: [session-fields](../../../specs/session-fields.md), sections 2–4.

## Aliases

`[fields]` in `screens/screens.toml` gives a short name to a field, an "alias":

```toml
[fields]
w = "term.cols"
```

Then `{w}` shows the columns of the terminal. A short alias helps in art (task 9).

A screen can have its own aliases: `[fields]` in its front matter. They are valid in the screen,
its fragments and its art. An alias of the screen replaces the alias of `screens/screens.toml`
with the same name:

```
+++
[fields]
w = "term.rows"
+++
Rows: {w}
```

In this screen, its fragments and its art, `{w}` shows the rows of the terminal.

Full rules: [bbt-format](../../../specs/bbt-format.md), sections 3.2, 3.3 and 4,
and [screens-toml](../../../specs/screens-toml.md), section 3.3.

## Live fields

A field with the mark `live` changes while the screen waits for a key:

```
Time {bbs.time live "15:04:05"}, online {session.online live}
```

- Each second, the BBS writes the new value in place of the old value. The rest of the screen stays.
- These fields can be live: `bbs.time`, `session.online`, `bbs.uptime` and `bbs.sessions`.
  The mark on another field has no effect.
- A live field has the width of its largest value, so the text after it does not move.
  For example, `session.online` has 14 cells, the width of `9999d 23:59:59`.
- After the mark, `bbs.time` can have its own form in double quotes, for example `"15:04:05"`.
  Without it, the field has the form of `[screen] time`.
- An alias can hold the mark. Then `{clock}` is a live field:

```toml
[fields]
clock = 'bbs.time live "15:04:05"'
```

- A token with an alias takes no mark: `{clock live}` is an error.
- A condition, `{#if}` or `when`, takes the value when the BBS shows the screen. The condition does not change.
- A live field does not change on a page with a More row and while the user types a line (task 8).
- `velourbbs preview` has a clock that stands still. In `--keys`, `+1s` moves it one second
  and shows the screen after the second: `velourbbs preview main --keys +1s`.

Full rules: [live-screen](../../../specs/live-screen.md), sections 3, 4 and 7.

## Conditions

```
{#if term.colors == 16777216}16 million{#else}{term.colors}{/if}
```

This text shows `16 million` when the terminal shows 16777216 colours. Else it shows the number.

- `{#if <condition>}` shows the text up to `{#else}` or `{/if}` when the condition is true.
- `{#else}` shows the text up to `{/if}` when the condition is false. `{#else}` is optional.
- A block can span lines and can hold other blocks. Each `{#if}` needs its `{/if}` in the same file.
- A line with `{#if}`, `{#else}` or `{/if}` gives no row when only spaces stay on it.
- The operators are `==`, `!=`, `<`, `<=`, `>` and `>=`.
  A number field takes a number and each operator.
  A text field takes text in double quotes and only `==` or `!=`: `term.type == "xterm"`.
- `and` and `or` join conditions. `and` binds tighter. There are no parentheses.

Full rules: [bbt-format](../../../specs/bbt-format.md), sections 5 and 6.2.

## Items with conditions

`when` gives a condition to an item. When the condition is false, the item does not work.
`{#if item.<KEY>}` is true when the item `<KEY>` works. So the text of an item and its key agree:

```
+++
[[item]]
key = "W"
do = "gosub wide"
when = "term.cols >= 132"
[[item]]
key = "G"
do = "logoff"
+++
{#if item.W}W  wide screen{/if}
G  goodbye
```

On a terminal of 132 columns or more, the screen shows `W  wide screen`, and `W` works.
On a narrower terminal, the line gives no row, and the BBS ignores `W`.
`not item.W` is true when the item does not work.

Full rules: [bbt-format](../../../specs/bbt-format.md), section 8.2.

## The list of fields

`velourbbs fields` lists the fields of your BBS. A part of its output:

```
velourbbs fields
```

```text
term.type        string   -   -      xterm
term.cols        integer  3   -      80
term.colors      integer  8   -      256
term.background  string   5   -
session.lang     string   2   -      en
bbs.name         string   -   -      Kind Brothers BBS
```

- Each line gives the path, the type, the width, the provider (task 8) and the sample value.
- The width is the most cells that a value can have. `-` means that the command gives no width.
  Use the width for a fixed-width field: `{term.cols}` needs 3 cells, `{term.background}` needs 5.
- The sample value is the value of `velourbbs preview`.
- The command does the checks of `velourbbs check` first. After an error, it gives only the errors.

Full rules: [screen-render](../../../specs/screen-render.md), section 3.4.

## The screen `term`

Create `screens/term.bbt` with this text:

`examples/lamp/en/screens/term.bbt`:
```
+++
[[item]]
key = "ESC"
do = "back"
[[item]]
key = "ENTER"
do = "back"
+++
{>_head}
       {!dim}Size{!}     {term.cols}×{term.rows}
       {!dim}Colours{!}  {#if term.colors == 16777216}16 million{#else}{term.colors}{/if}
       {!dim}Type{!}     {term.type}

       This is how the station sees you.
{~}
  {!line}{~─}{!}  

    {!key}Esc Enter{!}  {!dim}back{!}

```

The screen at 80 columns and 24 rows:

```text

  ────────────────────────────────────────────────────────────────────────────

       Lamp  ☼ a BBS by lamplight

       Size     80×24
       Colours  256
       Type     xterm

       This is how the station sees you.










  ────────────────────────────────────────────────────────────────────────────

    Esc Enter  back
```

The labels "Size", "Colours" and "Type" are in the grey-brown `dim`. The values are in the cream `normal`.

## The item `T` in `main`

Add the item `T` to the front matter of `screens/main.bbt`, after the item `A`:

```
[[item]]
key = "T"
do = "gosub term"
```

In the body, add the row of `T` between the rows of `A` and `G`.
The spacer line gives 2 empty rows less, so the closing line stays on row 19.
Add `T` to the key row:

```
         {!key}A{!}   About {!line}{~.}{!} {!dim}who we are{!}  

         {!key}T{!}   Your terminal {!line}{~.}{!} {!dim}size and colours{!}  

         {!key}G{!}   Goodbye {!line}{~.}{!} {!dim}the lamp will wait{!}  
```

```
    {!key}A T G{!}  {!dim}choose{!}
```

## Result

- `velourbbs check` gives no output.
- `velourbbs preview main --keys T` shows the screen `term`.
- `velourbbs preview term` gives the rows `Size     80×24`, `Colours  256` and `Type     xterm`.
- `velourbbs preview term --set term.type=ansi` gives `Colours  16` and `Type     ansi`.
- `velourbbs preview term --set term.colors=16777216` gives `Colours  16 million`.
- In the telnet client, `T` shows the size of the window. Change the size of the window
  and press `Esc` and `T` again: the screen shows the new size, and the key row stays at the bottom.
  With fewer than 14 rows, the screen comes in pages (task 6), and a key only shows the next page.
