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
  Section 3.1 has one exception.
- The `.ai/` files are only in the main checkout. In the worktree, use their absolute path.
- At the end of a stage, commit the changes in the worktree. In the implementation stage,
  make one commit for each block. Write the commit message as section 3.1 tells.
- Merge the branch into `main` after each block commit. `/ticket-cleanup` merges the rest
  of the branch, removes the worktree and deletes the branch.
- To merge, use the main checkout. It must be on `main` and have no uncommitted changes.
  Run `git merge --no-ff <id> -m "Merge issue <id>"`. Then run the tests in the main checkout.
- If the merge has a conflict, run `git merge --abort`. In the worktree, run `git merge main`.
  Resolve the conflict with the operator, commit with `<id>: merge main`, then merge again.
- Do not push.

### 3.1. Commit message

- Write the subject as `<id>: <summary>`. Write the summary in English, in the imperative.
  Do not put a period at the end. The subject has 72 characters or fewer.
- Put a blank line after the subject. Then write the body: what changed and why.
  A list is permitted. A small change can have no body. Wrap the body at 72 characters.
- The operator is the git author. Each AI model of the issue is a co-author.
  The models of all sessions, subagents and reviewers are co-authors, from all vendors.
- The `AI` line of `desc.md` has the list of co-authors. Separate the items with `; `.
  Example: `- AI: Claude Opus 5.5 <noreply@anthropic.com>`.
  If `desc.md` has no `AI` line, add it.
- At the end of each session and before each commit, add to the `AI` line
  the model of the session and the models of its subagents that are not in the list.
- Use the model name as the model gives it, for example `Claude Opus 5.5`.
  Use the email of the vendor from the table below.
  If the vendor is not in the table, ask the operator for the email.
  Add a row to the table in the current issue. This is an exception to section 3.
- At the end of the message, put a blank line. Then put one line
  `Co-Authored-By: <model> <email>` for each item of the `AI` line.
  Do not wrap these lines. They replace the default attribution of the tool.
- A `<id>: merge main` commit has the subject and the `Co-Authored-By` lines.
- Exception: a `Merge issue <id>` commit has the subject only.

| Vendor | Email |
|---|---|
| Anthropic | `noreply@anthropic.com` |

Example of a commit message:

```
kqtxw-menu-parser: parse menu TOML files

- Add the menu loader and its tests.
- Reject a menu item with no action.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
```

## 4. Rules

- Review: if a change affects security, changes a decision or has a large effect on the design,
  propose more than one reviewer.
- Reviewer brief: give each reviewer a position, that is, what to check and from which view.
  Give the absolute paths of the files to review, the specification and `<block>.md`.
  For code, also give the worktree path and the command `git diff main...<id>`.
  Do not give your own conclusions. Ask the reviewer to give its model name in the report.
- Write the findings of each review to `artifacts/<name>-review.md`.
- Specification: if the code differs from the specification, update the specification
  in the same commit.
- Tests: write tests for each unit that you can test separately.
- Before a commit that changes code, run the tests and the linter. Commit only if they pass.
