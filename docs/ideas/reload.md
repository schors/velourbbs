# Idea: reload

Backlog. It extends [screen-format](../specs/screen-format.md).

## Goal

The sysop changes screens without a restart of the BBS.

## Decisions

- `bbs reload` or a file watch builds a new screen model.
- If the new model has errors, the old model stays and the host reports the errors.
- Sessions change to the new model at the next screen.
