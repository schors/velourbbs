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
| `term.charset` | Always `utf-8`. |
| `term.name` | The name and the version that the terminal reports, for example `XTerm(412)`. Can be empty. Use it only to show it. |
| `term.images` | `sixel` when the terminal reports sixel graphics. Else empty. Some terminals report it, but do not show sixel. |
| `term.ambiguous` | `2` when the terminal shows East Asian Ambiguous characters, for example Cyrillic letters and `■`, in two cells. Else `1`. The layout does not change, so frames and columns break. Use it to show the user how to turn this setting off. |
| `term.bg_color` | The background colour that the terminal reports, as `#rrggbb` in lower case. Empty when the terminal does not report it. |
| `term.fg_color` | The foreground colour that the terminal reports, as `#rrggbb` in lower case. Empty when the terminal does not report it. |
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
  until the next screen.
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

Full rules: [bbt-format](../../../specs/bbt-format.md), sections 3.3 and 4,
and [screens-toml](../../../specs/screens-toml.md), section 3.3.

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
- The width is the most cells that a value can have. `-` means that the value has no limit.
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










  {!line}────────────────────────────────────────────────────────────────────────────{!}

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
Remove two empty lines below the row of `G`: the closing line stays on row 19.
Add `T` to the key row:

```
         {!key}A{!}   About                                   {!dim}who we are{!}

         {!key}T{!}   Your terminal                           {!dim}size and colours{!}

         {!key}G{!}   Goodbye                                 {!dim}the lamp will wait{!}
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
- In the telnet client, `T` shows the size of the window. Change the width of the window,
  keep at least 24 rows, and press `Esc` and `T` again: the screen shows the new size.
  With fewer rows, the screen comes in pages (task 6), and a key only shows the next page.
