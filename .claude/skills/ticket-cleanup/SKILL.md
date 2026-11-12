---
name: ticket-cleanup
description: Merge complete issues into main, remove their worktrees and branches, and move them to .ai/archive/.
argument-hint: "[issue id] | drop <issue id>"
---

# Archive complete issues

Read `docs/dev-flow/flow.md`, sections 2 and 3. They define the paths, a complete issue
and the git rules.

Run the skill in the main checkout (`flow.md`, section 3).
If the current directory is a worktree, go to the main checkout first.

1. Select the issues. Use the issue from the arguments, or all issues in `.ai/issues/`.
2. Check each issue against the definition of a complete issue.
3. Show the operator the complete issues and the incomplete issues with the reason.
   If the operator named an incomplete issue, continue with it only after the operator confirms.
4. Make sure that the main checkout is on `main` and has no uncommitted changes.
   If not, stop. Tell the operator.
5. For each issue:
   1. Merge the branch into `main` as `flow.md`, section 3, tells.
      If there is a conflict, skip the issue. Tell the operator the steps from section 3.
   2. Run `git worktree remove .ai/worktrees/<id>` and `git branch -d <id>`.
   3. Move the issue directory to `.ai/archive/`.
   4. In `.ai/issues.md`, set the stage of the issue to `archived`.
6. Tell the operator what merged and moved. Show `git log --oneline -n 5`.

## Drop

If the arguments are `drop <id>`, stop the issue without a merge:

1. Show the operator the commits of the branch that are not in `main`
   and the uncommitted changes in the worktree. Continue only after the operator confirms.
2. Run `git worktree remove --force .ai/worktrees/<id>` and `git branch -D <id>`.
3. Move the issue directory to `.ai/archive/`.
4. In `.ai/issues.md`, set the stage of the issue to `dropped`.
