# Task 4. A menu with hotkeys

## Items

Each `[[item]]` in the front matter is one item of the menu:

```
[[item]]
key = "A"
do = "gosub about"
```

- `key` is the key: a Latin letter, a digit or a named key, for example `ESC` or `ENTER`.
  Case does not matter. A Russian letter works as the Latin letter on the same key: `ф` is `A`.
- `do` is the action: `gosub <screen>` opens the screen, `back` goes back to the screen
  that opened it, `logoff` ends the session. The section "The navigation stack" has
  the other actions `goto` and `top`.
- The BBS does not show the items. Write the text of the items and the key row in the body.
  The key row shows all keys of the screen.

The short form writes each item on one line. It is the same menu:

```
item = [
  { key = "A", do = "gosub about" },
  { key = "G", do = "logoff" },
]
```

Use one form in a screen.

Full rules: [bbt-format](../../../specs/bbt-format.md), section 8,
and [menu-nav](../../../specs/menu-nav.md), sections 3–6.

## The screen `main`

The menu has two items: `A` goes to the screen `about`, `G` ends the session.
Task 5 adds the item `T`, task 7 adds the item `R`.

Write `screens/main.bbt` with this text:

```
+++
[[item]]
key = "A"
do = "gosub about"
[[item]]
key = "G"
do = "logoff"
+++
{>_head}
       Good evening. It's quiet and warm here.
       Pull up a chair by the lamp.


         {!key}A{!}   About                                   {!dim}who we are{!}

         {!key}G{!}   Goodbye                                 {!dim}the lamp will wait{!}






       {!dim}The light stays on till morning.{!}

  {!line}────────────────────────────────────────────────────────────────────────────{!}

    {!key}A G{!}  {!dim}choose{!}
```

The screen:

```text

  ────────────────────────────────────────────────────────────────────────────

       Lamp  ☼ a BBS by lamplight

       Good evening. It's quiet and warm here.
       Pull up a chair by the lamp.


         A   About                                   who we are

         G   Goodbye                                 the lamp will wait






       The light stays on till morning.

  ────────────────────────────────────────────────────────────────────────────

    A G  choose
```

The keys are in the orange `key`. The notes on the right and the word "choose" are in the grey-brown `dim`.

## The screen `about`

The keys `Esc` and `Enter` go back to the screen that opened `about`:

```
[[item]]
key = "ESC"
do = "back"
```

`examples/lamp/en/screens/about.bbt`:
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
       Lamp is a small station for friends.
       Come in, look around, stay a while.

       The station runs on Velour BBS.
       Its screens are plain text files.










  {!line}────────────────────────────────────────────────────────────────────────────{!}

    {!key}Esc Enter{!}  {!dim}back{!}
```

The screen:

```text

  ────────────────────────────────────────────────────────────────────────────

       Lamp  ☼ a BBS by lamplight

       Lamp is a small station for friends.
       Come in, look around, stay a while.

       The station runs on Velour BBS.
       Its screens are plain text files.










  ────────────────────────────────────────────────────────────────────────────

    Esc Enter  back
```

The keys `Esc Enter` are in the orange `key`. The word "back" is in the grey-brown `dim`.

## The navigation stack

The BBS keeps the open screens of a session in a stack. These actions change the stack:

| `do` | Action |
|---|---|
| `gosub <screen>` | Opens the screen over the current screen. |
| `back` | Closes the current screen and shows the screen below it again. On the first screen, it shows the same screen again. |
| `top` | Closes all screens over the first screen and shows the first screen. |
| `goto <screen>` | Shows the screen in place of the current screen. `back` then goes to the screen below. |

- A screen of `back` and `top` keeps its arguments (task 7) and the page of a text.
- The stack holds at most 16 screens. A `gosub` on the 16th screen removes the second screen.
  The first screen stays.

Example: `main` opens `texts` with `gosub`, `texts` opens `reader` with `gosub`.
In `reader`, `back` shows `texts` and `top` shows `main`.

Use `gosub` to go to a deeper screen, and `back` to come back. Then a screen can have
more than one way into it, and it does not need to know the screen before it.

Full rules: [menu-nav](../../../specs/menu-nav.md), section 4.1.

## The screen `goodbye`

After `logoff`, the BBS shows the screen `goodbye` and ends the session.
`goodbye` has no front matter and no items. It has no key row.

```
       {!title}See you soon.{!}
       The lamp stays on for you.
```

`examples/lamp/en/screens/goodbye.bbt`:
```
{>_head}
       {!title}See you soon.{!}
       The lamp stays on for you.













  {!line}────────────────────────────────────────────────────────────────────────────{!}
```

The screen has 21 rows:

```text

  ────────────────────────────────────────────────────────────────────────────

       Lamp  ☼ a BBS by lamplight

       See you soon.
       The lamp stays on for you.













  ────────────────────────────────────────────────────────────────────────────
```

"See you soon." is in the light peach `title` in bold.

## The screen `shutdown`

At SIGINT or SIGTERM, `velourbbs serve` stops. It shows the screen `shutdown`
to each open session and ends the session. Tell the visitors that the BBS restarts:

```
       {!title}The lamp goes out for a minute.{!}
       Call again soon.
```

The two rows:

```text
       The lamp goes out for a minute.
       Call again soon.
```

"The lamp goes out for a minute." is in the light peach `title` in bold.

- `shutdown` has the rules of `goodbye`: no items and no arguments.
  It is not a file screen: it has no key `file` (task 7).
- The BBS shows only the first page of `shutdown`. It does not wait for a key.
- Without `screens/shutdown.bbt`, the BBS ends the sessions with no text.
- After 5 s, the BBS closes each connection that is still open.

`velourbbs preview shutdown` shows the screen.

Full rules: [menu-nav](../../../specs/menu-nav.md), section 5.2,
and [telnet-term](../../../specs/telnet-term.md), section 10.2.

## Test the keys

`velourbbs preview <screen> --keys <keys>` presses the keys and shows each screen.
Separate the keys with commas:

```
velourbbs preview main --keys A,ESC,G
```

The output has an event line before each cell dump:

```
screen main
page 1 80x24 fill spaces
...
key A
screen about
page 1 80x24 fill spaces
...
key ESC
screen main
page 1 80x24 fill spaces
...
key G
logoff
screen goodbye
page 1 80x24 fill spaces
...
```

Full rules: [menu-nav](../../../specs/menu-nav.md), section 7.

## Result

- `velourbbs check` gives no output.
- `velourbbs preview main --keys A,ESC,G` gives the event lines above.
- `velourbbs preview goodbye` gives the 21 rows above.
- In the telnet client, `A` shows `about`, `Esc` goes back, `G` shows `goodbye` and ends the session.
