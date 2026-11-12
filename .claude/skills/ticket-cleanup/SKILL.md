---
name: ticket-cleanup
description: Merge completed issues into main, remove their worktrees and branches, and move them to .ai/archive/.
argument-hint: "[issue id]"
---

# Archive completed issues

Read `docs/dev-flow/flow.md`, sections 2 and 3. They define the paths, a complete issue
and the git rules.

Run the skill in the main checkout. The main checkout is the first line of `git worktree list`.
If the current directory is a worktree, go to the main checkout first.

1. Select the issues. Use the issue from the arguments, or all issues in `.ai/issues/`.
2. Check each issue against the definition of a complete issue.
3. Show the operator the complete issues and the incomplete issues with the reason.
   If the operator named an incomplete issue, continue with it only after the operator confirms.
4. Make sure that the main checkout is on `main` and has no uncommitted changes.
   If not, stop and tell the operator.
5. For each issue:
   1. Run `git merge --no-ff <id> -m "Merge issue <id>"`.
      If there is a conflict, run `git merge --abort` and skip the issue.
      Tell the operator the next steps: in the worktree, run `git merge main`, resolve
      the conflict, commit with `<id>: merge main`, then run `/ticket-cleanup <id>` again.
   2. Run `git worktree remove .ai/worktrees/<id>` and `git branch -d <id>`.
   3. Move the issue directory to `.ai/archive/`.
   4. In `.ai/issues.md`, set the stage of the issue to `archived`.
6. Tell the operator what merged and moved. Show `git log --oneline -n 5`.
