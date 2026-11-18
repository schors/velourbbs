# Issue body templates

Write the body in English. `ghflow` adds the marker line.
Add a `Source: #<n>` line at the top for each source issue.
Remove the comments.

## `one-shot`, `docs`

```markdown
## Goal

<!-- What the issue gives, in 1–3 sentences. -->

## Scope

<!-- What is in the issue. What is not in the issue. -->

## Acceptance

<!-- Checks that show the issue is complete. -->
```

## `idea`

```markdown
## Goal

## Scope

## Decisions of the operator

## Open questions
```

## `dev`

```markdown
## Goal

## Scope
```

## `block`

```markdown
Parent: #<n>
Specification: `docs/specs/<slug>.md#<section>`

## Task

<!-- What the block does. -->

## Acceptance

<!-- Checks that show the block is complete. -->
```
