# Idea: update fields on an open screen

Backlog. It extends [session-fields](session-fields.md).

## Goal

The online time, the number of sessions, the clock and the uptime change on the screen
without a move to another screen.

## Decisions

- Not in the first version of [session-fields](session-fields.md). The change is too large.

## Open questions

- A wait for a key or a timer. Now the session waits for a key with no timeout.
- A redraw of the field cells only, with no flicker. Now the host draws a full page.
- A `{#if}` condition on a live field changes the layout and the enabled items.
- The update period.
