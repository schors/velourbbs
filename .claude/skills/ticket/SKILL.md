---
name: ticket
description: Create an issue (idea, dev or one-shot) in .ai/issues/, or open an issue by its id to continue it.
argument-hint: "<id> | new <idea|dev|one-shot> [slug] [description | idea files]"
---

# Create or open an issue

Read `docs/dev-flow/flow.md`. It defines the issue types, the stages, the paths,
the issue id, the git rules and the gates.

Run the skill in the main checkout (`flow.md`, section 3).

If the first argument is `new`, create an issue. Otherwise the first argument is an issue id
or its 5 letters: open the issue.

## Create

1. Get the type, the slug and the description from the arguments after `new`.
   For a `dev` issue, the description can be the paths of idea files in `docs/ideas/`.
   If there is no type or no description, ask the operator.
   If there is no slug, make one from the description or from the names of the idea files.
2. Make the 5 letters with the command in `flow.md`, section 2.
   If an issue in `.ai/issues/` or `.ai/archive/` has these 5 letters, make new letters.
3. Create `.ai/issues/<id>/artifacts/`.
4. Write `artifacts/desc.md` from `docs/dev-flow/template-desc.md`.
   Fill it from the idea files or the description. Leave unknown fields as placeholders.
5. If `.ai/issues.md` does not exist, create it from `docs/dev-flow/template-issues.md`.
6. Add a line for the issue to `.ai/issues.md`. Set the stage as `flow.md`, section 2, tells.
7. Run `git worktree add .ai/worktrees/<id> -b <id> main`.
8. Go to the worktree. Tell the operator the issue id and the worktree path.
9. Read the file of the type, or of the `design` stage, from `flow.md`, section 1.
   Do its steps in this session.

## Open

1. Find the issue directory in `.ai/issues/`.
   If it is in `.ai/archive/`, stop. Tell the operator.
2. Read `artifacts/desc.md`, `plan.md` if it exists, and the line in `.ai/issues.md`.
3. Read the file of the type, or of the current stage, from `flow.md`, section 1.
4. Go to `.ai/worktrees/<id>/`. If the worktree does not exist, run
   `git worktree add .ai/worktrees/<id> <id>`.
5. Tell the operator the type, the stage and the next step.
   For the implementation stage, the next step is the first block that is not done.
   For an `idea` or `one-shot` issue, find the next step from `desc.md` and the worktree.
6. Do the next step and the steps after it.
