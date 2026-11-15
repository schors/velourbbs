# Idea: bbt-guide

Backlog. It adds a guide to the [BBT format](../specs/bbt-format.md) and the [art format](../specs/art-format.md).

## Goal

The artist and the screen author learn to write screens without the specifications.

## Decisions

- The reader is the artist and the screen author.
- The guide shows tasks: a screen, a menu with hotkeys, an include, the styles
  and the theme, an art insert, a check with `velourbbs check` and `velourbbs preview`.
- The guide refers to bbt-format and art-format for the full rules.
- The guide is in English and Russian. English is the source, Russian is a translation.
  Change both in the same commit.
- The guide builds the light example BBS of [example-style](example-style.md).

## Open questions

- The place of the guide files.
- A flow rule that a change of the format also changes the guide.
