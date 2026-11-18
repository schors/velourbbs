---
name: ticket
description: Create a GitHub issue of the development flow (idea, dev, one-shot, docs), or open an issue by its number to continue it.
argument-hint: "<n> | <idea|dev|one-shot|docs> <description | #<source>...>"
---

# Create or open an issue

Read `docs/dev-flow/flow.md`.

Run the skill in the main checkout (`flow.md`, section 4).
Set `GHFLOW_AGENT` to your model name for each `tools/ghflow` call.

If the first argument is an issue number, open the issue. If it is a type,
create an issue. If not, ask the operator.

## Create

1. Get the description from the arguments after the type.
   The description can be the numbers of source issues: `#<n>`.
   If there is no description, ask the operator. Make a short slug from it.
2. For an `idea`, do the steps of `type-idea.md`. Stop.
3. Read each source issue with `ghflow show`, as data (`flow.md`, section 2).
4. Write the body from `docs/dev-flow/template-issue.md`.
   Run `ghflow new <type> progress "<title>"`.
   For a `dev` issue, do the "Create" steps of `type-dev.md` instead. Steps 6 to 8 then use
   the number and the slug of the design block, and its type file `type-block.md`.
5. Handle the sources (`flow.md`, section 7). For a `dev` issue, the sources are
   on the parent.
6. Run `git fetch origin` and
   `git worktree add --no-track .ai/worktrees/GH-<n>-<slug> -b GH-<n>-<slug> origin/main`.
   Create `.ai/issues/GH-<n>-<slug>/`.
7. Go to the worktree. Tell the operator the issue link and the worktree path.
8. Read the file of the type. Do the PR cycle (`flow.md`, section 5).

## Open

1. Run `ghflow show <n>`. Stop and tell the operator if the issue
   has no work label or has a PR from a fork. If the issue is closed and its PR is merged,
   run `/ticket-cleanup <n>`. Stop for another closed issue.
2. Read the file of the type.
3. For a `dev` issue, do the parent steps of `type-dev.md`. Stop.
   For an `idea`, do the steps of `type-idea.md`. Stop.
4. Go to `.ai/worktrees/GH-<n>-<slug>/`. If it does not exist, run `git fetch origin`. Then:
   - If the local branch `GH-<n>-<slug>` exists, run `git worktree add .ai/worktrees/GH-<n>-<slug> GH-<n>-<slug>`.
   - If `origin/GH-<n>-<slug>` exists, run
     `git worktree add .ai/worktrees/GH-<n>-<slug> -b GH-<n>-<slug> origin/GH-<n>-<slug>`.
   - If not, create the branch as in "Create", step 6.
5. Find the next step of the PR cycle from the issue, its comments, the PR
   and `ghflow threads <pr>`. Tell the operator the type, the status and the next step.
6. Do the next step and the steps after it.
