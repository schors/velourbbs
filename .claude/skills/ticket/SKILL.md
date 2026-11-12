---
name: ticket
description: Create an issue (idea, dev or one-shot) in .ai/issues/, or open an issue by its id to continue it.
argument-hint: "<id> | new <idea|dev|one-shot> [slug] [description]"
---

# Create or open an issue

Read `docs/dev-flow/flow.md`, sections 2 and 3. They define the paths, the issue id,
the issue types and the git rules.

Run the skill in the main checkout. The main checkout is the first line of `git worktree list`.

If the first argument is `new`, create an issue. Otherwise the first argument is an issue id
or its 5 letters: open the issue.

## Create

1. Get the type and the slug from the arguments after `new`.
   If there is no type, ask the operator. If there is no slug, propose one from the description.
2. Make the 5 letters with the command in `flow.md`.
   If `.ai/issues/` or `.ai/archive/` already has this prefix, make new letters.
3. Create `.ai/issues/<id>/artifacts/`.
4. Write `artifacts/desc.md` from `docs/dev-flow/template-desc.md`.
   Fill it from the idea file or the description. Leave unknown fields as placeholders.
5. If `.ai/issues.md` does not exist, create it from `docs/dev-flow/template-issues.md`.
6. Add a line for the issue to `.ai/issues.md`. Set the stage to the first stage of the type.
7. Run `git worktree add .ai/worktrees/<id> -b <id> main`.
8. Go to the worktree. Tell the operator the issue id and the worktree path.

If the operator called `/ticket` directly, stop here.

## Open

1. Find the issue directory in `.ai/issues/`.
   If it is in `.ai/archive/`, tell the operator and stop.
2. Read `artifacts/desc.md`, `plan.md` if it exists, and the line in `.ai/issues.md`.
3. Read the file of the current stage from `flow.md`, section 1.
4. Go to `.ai/worktrees/<id>/`. If the worktree does not exist, run
   `git worktree add .ai/worktrees/<id> <id>`.
5. Tell the operator the type, the stage and the next step.
   For the implementation stage, the next step is the first block that is not done.
6. Continue the stage when the operator confirms.
