# Idea: key-layouts

Backlog. It extends [menu-nav](../specs/menu-nav.md).

## Goal

A hotkey works when the user types it in a non-Latin keyboard layout.
A letter runs the item of the Latin letter on the same key.
Today only the Russian ЙЦУКЕН layout maps to Latin letters
([menu-nav](../specs/menu-nav.md), section 3.2). Other letters match only `ANY`.
Example: the Ukrainian `і` is on the key of `S`, but it does not run the item `S`.

## Scope

In: all non-Latin layouts: Ukrainian, Belarusian, other Cyrillic layouts
and other alphabets (Greek, Hebrew, Georgian and others).

## Decisions

- The design stage researches the methods and selects one.
- Priority: backlog. The Russian layout already works.

## Open questions

- What do other BBS hosts and terminal programs do with non-Latin layouts?
  Examples: more layout maps, physical key codes, other methods.
- Can an item have a non-Latin key, for example `key = "Ж"`?
  How does it work together with the layout maps?
- One letter can be on different keys in different layouts
  (Bulgarian phonetic and ЙЦУКЕН). The host sees only the character.
  Is a layout setting allowed: none, for the sysop, for the user?
- Some layouts (Kazakh, Georgian) have letters on the keys of digits or punctuation.
  Does such a letter run the item of that digit?
