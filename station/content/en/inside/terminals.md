# Terminals

## Connection

Now the guests come by telnet. Velour works with PuTTY, Tera Term and telnet
in a UTF-8 terminal: Windows Terminal, xterm, tmux, iTerm2, Ghostty.
With other clients, Velour works on a best-effort basis.

## What Velour detects itself

At connect, Velour asks the terminal and finds from the replies:

- the size of the terminal: 20 to 255 columns and 10 to 100 rows;
- the type of the terminal;
- the colour depth: 16, 256 or 16 million colours;
- the charset.

Velour replaces a colour that the terminal does not have with the nearest one.
So one screen looks right at each colour depth.

## UTF-8 only

Velour writes text only in UTF-8. UTF-8 has Cyrillic, Latin,
the lines and the blocks of art at once. An 8-bit charset has only 256 characters, and each has its own set.
For example, CP437 has no Cyrillic and no Lithuanian letters. With one charset, three languages live on one station,
and the screens, the texts and the art have one form.

Velour checks the charset so: it prints the letter `â` and asks where the cursor is.
In UTF-8, the cursor moves by one cell, in an 8-bit charset by two.
If the cursor moved by two cells, Velour tells how to change the charset
and closes the connection. If the terminal does not reply, Velour takes UTF-8.

## What does not work everywhere

- SyncTERM and NetRunner have no UTF-8. They get the explanation and do not come in.
- telnet.exe of Windows works on a best-effort basis: only 16 colours,
  and ANSI art and lines can break.
- MUD clients, for example Mudlet, do not work: they do not clear the screen and do not move the cursor.

## What comes next

- SSH.
- telnets: telnet over TLS.

There is no decision yet on CP437 terminals and terminals without ANSI.
