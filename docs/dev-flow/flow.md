# Development flow

This document defines the files and the stages of development.
Read this file first. Then read only the file of the current stage.

## 1. Stages

| Stage | File | Result |
|---|---|---|
| Idea | [stage-idea.md](stage-idea.md) | Ideas in `docs/ideas/`. An archived issue. |
| Design | [stage-design.md](stage-design.md) | An issue and specifications in `docs/specs/`. |
| Decomposition | [stage-decomposition.md](stage-decomposition.md) | `plan.md` and block files. |
| Implementation | [stage-implementation.md](stage-implementation.md) | One block, implemented and reviewed. |
| One-shot | [stage-one-shot.md](stage-one-shot.md) | A small task, from idea to result. |

Start each stage in a new session.
To continue an issue in a new session, run `/ticket <id>`. The 5 letters of the id are sufficient.
When `/ticket <id>` opens an issue, skip the stage step that creates the issue.
"Operator" is the human who controls the work.
A "gate" is a stop where the operator must reply before the work continues.

## 2. Files

| Path | Tracked | Content |
|---|---|---|
| `docs/ideas/` | yes | Idea descriptions. |
| `docs/specs/` | yes | Designs and specifications (SDD). |
| `.ai/issues.md` | no | Index of issues. Template: [template-issues.md](template-issues.md). |
| `.ai/artifacts/` | no | Artifacts for more than one issue, or for no issue yet. |
| `.ai/issues/<id>/` | no | One issue. |
| `.ai/issues/<id>/artifacts/` | no | Artifacts for this issue only. |
| `.ai/issues/<id>/artifacts/desc.md` | no | What the issue is. Template: [template-desc.md](template-desc.md). |
| `.ai/issues/<id>/artifacts/<name>-review.md` | no | Review findings. `<name>` is the stage or the block. |
| `.ai/issues/<id>/plan.md` | no | Index of blocks. Template: [template-plan.md](template-plan.md). |
| `.ai/issues/<id>/<block>.md` | no | Block description. Template: [template-block.md](template-block.md). |
| `.ai/issues/<id>/<block>.done.md` | no | Block result. Template: [template-block.done.md](template-block.done.md). |
| `.ai/archive/<id>/` | no | A completed issue. |
| `.ai/worktrees/<id>/` | no | The git worktree of the issue. |

- The issue `<id>` is 5 random lowercase letters, a hyphen and a slug.
  The letters `i`, `l` and `o` are not used. Example: `kqtxw-menu-parser`.
- Make the 5 characters with `LC_ALL=C tr -dc 'a-hjkmnp-z' </dev/urandom | head -c5`.
- A `<block>` name is a slug. Example: `toml-loader`.
- Each issue with `plan.md` has a `<block>.md` file for each block.
- `/ticket new` sets the first stage of the issue in `.ai/issues.md`.
  At the end of a stage, set the next stage of the type (section 2.1) in `.ai/issues.md`.
- Move a complete issue directory to `.ai/archive/` with `/ticket-cleanup <id>`.
  To stop an incomplete issue, run `/ticket-cleanup drop <id>`.

### 2.1. Issue types

| Type | Created at | Stages | Complete when |
|---|---|---|---|
| `idea` | [stage-idea.md](stage-idea.md) | idea | `desc.md` lists one or more ideas. All of them are in `docs/ideas/`. |
| `dev` | [stage-design.md](stage-design.md) | design, decomposition, implementation | `plan.md` has one or more blocks. All blocks are done. Each block has `<block>.done.md`. |
| `one-shot` | [stage-one-shot.md](stage-one-shot.md) | one-shot | The `Result` section of `desc.md` is filled. |

An issue is complete only if it is in the last stage of its type and the condition above is true.
Also, its worktree must have no uncommitted changes.
After the last stage, the stage of each type is `archived`. The stage of a stopped issue is `dropped`.

## 3. Git

- Each issue has a branch `<id>` from `main` and a worktree in `.ai/worktrees/<id>/`.
  `/ticket new` creates them.
- Change tracked files only in the worktree of the issue. Do not change them in the main checkout.
- Change this flow (`docs/dev-flow/`, `.claude/skills/`) and `AGENTS.md` only in a one-shot issue.
- The `.ai/` files are only in the main checkout. In the worktree, use their absolute path.
- At the end of a stage, commit the changes in the worktree. In the implementation stage,
  make one commit for each block.
- Start the commit message with `<id>: `.
- Merge the branch into `main` after each block commit. `/ticket-cleanup` merges the rest
  of the branch, removes the worktree and deletes the branch.
- To merge, use the main checkout. It must be on `main` and have no uncommitted changes.
  Run `git merge --no-ff <id> -m "Merge issue <id>"`. Then run the tests in the main checkout.
- If the merge has a conflict, run `git merge --abort`. In the worktree, run `git merge main`.
  Resolve the conflict with the operator, commit with `<id>: merge main`, then merge again.
- Do not push.

## 4. Rules

- Review: if a change affects security, changes a decision or has a large effect on the design,
  propose more than one reviewer.
- Reviewer brief: give each reviewer a position, that is, what to check and from which view.
  Give the absolute paths of the files to review, the specification and `<block>.md`.
  For code, also give the worktree path and the command `git diff main...<id>`.
  Do not give your own conclusions.
- Write the findings of each review to `artifacts/<name>-review.md`.
- Specification: if the code differs from the specification, update the specification
  in the same commit.
- Tests: write tests for each unit that you can test separately.
- Before a commit that changes code, run the tests and the linter. Commit only if they pass.
