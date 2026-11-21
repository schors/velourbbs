---
name: ticket-cleanup
description: Remove the local files of merged issues and close their sources, or drop an issue.
argument-hint: "[<n>] | drop <n>"
---

# Clean up issues

Read `docs/dev-flow/flow.md`, sections 2 to 4.

Run the skill in the main checkout (`flow.md`, section 4).
Set `GHFLOW_AGENT` to your model name for each `tools/ghflow` call.

1. Select the issues: the issue from the arguments, or each worktree
   in `.ai/worktrees/` and each folder in `.ai/issues/` whose name starts with `GH-`.
2. Run `ghflow show <n>` for each issue. An issue is complete when its PR is merged.
   A closed issue with no worktree and no merged PR has only a folder to delete.
3. Show the operator the complete issues, the folders to delete
   and the incomplete issues with the reason.
4. Make sure that the main checkout is on `main` and has no uncommitted changes.
   If not, stop. Tell the operator. Run `git fetch origin` and `git merge --ff-only origin/main`.
5. For each complete issue:
   1. If the worktree has uncommitted changes or commits that are not in `main`,
      show them to the operator. Continue only after the operator confirms.
   2. If they exist, run `git worktree remove --force .ai/worktrees/GH-<n>-<slug>`
      and `git branch -D GH-<n>-<slug>`.
   3. Delete `.ai/issues/GH-<n>-<slug>/`.
   4. Close each open source issue (`Source: #<s>`) as `completed`.
   5. For a block, do the parent steps (`type-dev.md`).
6. For each closed issue that has only a folder, delete `.ai/issues/GH-<n>-<slug>/`.
7. Tell the operator what was removed.

## Drop

If the arguments are `drop <n>`, stop the issue with no merge:

1. Show the operator the issue, its PR, the commits of the branch that are not in `main`
   and the uncommitted changes. Continue only after the operator confirms.
2. Close the PR with `ghflow close <pr> not_planned` and the issue
   with `ghflow close <n> not_planned`.
   If `origin/GH-<n>-<slug>` exists, run `git push origin --delete GH-<n>-<slug>`.
3. Close each open source issue (`Source: #<s>`) as `not_planned`.
4. Run `git worktree remove --force .ai/worktrees/GH-<n>-<slug>` and `git branch -D GH-<n>-<slug>`.
5. Delete `.ai/issues/GH-<n>-<slug>/`.
6. For a block, do the parent steps (`type-dev.md`).
