# Velour features

Velour is a BBS program. This is what it can do now.

## Connection

Guests come by telnet. Velour works with PuTTY, Tera Term and inetutils telnet in a UTF-8 terminal.

Velour finds the terminal type, the terminal size and the colour depth: 16, 256 or 16 million colours.
The terminal needs the UTF-8 charset. Without it, Velour tells the guest how to turn it on.

The terminal size is from 20 to 255 columns and from 10 to 100 rows.
A new terminal size takes effect at the next screen.

## Screens

Screens are text files of the sysop in the BBT format. A file has text, styles, fields and conditions.
One screen can have different forms for a narrow and a wide terminal.

A theme gives the styles, and each colour depth can have its own.
Velour replaces a colour that the terminal does not have with the nearest one.

Fields show the data of the session and of the station: the address of the guest, the time online,
the BBS name, the Velour version, the open sessions, the server time and the uptime.
The sysop sets the form and the time zone of the server time.

## Menus

Hotkeys lead from screen to screen. Letters also work in the Russian layout:
`й` is `q`. A menu item can work only under a condition.

## Texts

A file screen shows Markdown, plain text or source code in pages.
Markdown has headings, lists, quotes, code, tables and links with notes.
Velour counts the reading time.

One screen shows many files: a menu item gives it the file.
A file screen can show a part of a source file and line numbers.

## Art

`velourbbs art import` converts ANSI art from Moebius, icy_draw and PabloDraw to the Velour format.
A screen inserts art between its text.

## For the sysop

`velourbbs check` finds the errors in the screen files and gives the file, the line and the column.
`velourbbs preview` shows a screen without a connection: at any terminal size,
at any colour depth and after key presses.

The station limits the connections: in total, and from one address both the open ones and the new ones per minute.
Velour takes the address of the guest from a proxy with the PROXY protocol.

The setup guide shows how to run Velour under systemd.
