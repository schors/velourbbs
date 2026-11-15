# Development flow

Read this file. Then read only one more file: the file of the issue type,
or of the current stage for a `dev` issue.

## 1. Issue types and stages

| Type | Stage | File | Result |
|---|---|---|---|
| `idea` | `-` | [type-idea.md](type-idea.md) | Ideas in `docs/ideas/`. |
| `dev` | `design` | [stage-design.md](stage-design.md) | Specifications in `docs/specs/`. |
| `dev` | `decomposition` | [stage-decomposition.md](stage-decomposition.md) | `plan.md` and block files. |
| `dev` | `implementation` | [stage-implementation.md](stage-implementation.md) | One block, implemented and reviewed. |
| `one-shot` | `-` | [type-one-shot.md](type-one-shot.md) | A small task, from description to result. |

- "Operator" is the human who controls the work.
- A "gate" is a stop where the operator must reply before the work continues.
  Before the question, show in the chat, in Russian, the content that the operator
  decides on. A file name or an id is not content.
- `/ticket new <type>` starts an issue. `/ticket <id>` continues it.
  The 5 letters of the id are sufficient.
- Only a `dev` issue has stages, in the order of the table. Do each next stage in a new session.
  At the end of a stage, set the next stage in `.ai/issues.md`. Then stop.
  Tell the operator to run `/ticket <id>`.
- `/ticket-cleanup <id>` archives a complete issue. `/ticket-cleanup drop <id>` stops
  an incomplete issue.

## 2. Files

| Path | Tracked | Content |
|---|---|---|
| `docs/ideas/` | yes | Idea descriptions. The backlog of `dev` issues. |
| `docs/specs/` | yes | Designs and specifications (SDD). |
| `docs/brand/` | yes | Brandbook and its example screens. |
| `.ai/issues.md` | no | Index of issues. Template: [template-issues.md](template-issues.md). |
| `.ai/artifacts/` | no | Artifacts for more than one issue, or for no issue yet. |
| `.ai/issues/<id>/` | no | One issue. |
| `.ai/issues/<id>/artifacts/` | no | Artifacts for this issue only. |
| `.ai/issues/<id>/artifacts/desc.md` | no | What the issue is. Template: [template-desc.md](template-desc.md). |
| `.ai/issues/<id>/artifacts/<name>-review.md` | no | Review findings. `<name>` is the type, the stage or the block. |
| `.ai/issues/<id>/plan.md` | no | Index of blocks. Template: [template-plan.md](template-plan.md). |
| `.ai/issues/<id>/<block>.md` | no | Block description. Template: [template-block.md](template-block.md). |
| `.ai/issues/<id>/<block>.done.md` | no | Block result. Template: [template-block.done.md](template-block.done.md). |
| `.ai/archive/<id>/` | no | An archived or dropped issue. |
| `.ai/worktrees/<id>/` | no | The git worktree of the issue. |

The issue `<id>` is 5 random lowercase letters, a hyphen and a slug: `kqtxw-menu-parser`.
A `<block>` name is a slug.

## 3. Git

- Each issue has a branch `<id>` from `main` and its worktree. `/ticket new` creates them.
- Change tracked files only in the worktree of the issue.
- Change this flow (`docs/dev-flow/`, `.claude/skills/`) and `AGENTS.md` only in a one-shot issue.
- The "main checkout" is the first line of `git worktree list`.
  The `.ai/` files are only in the main checkout. In the worktree, use their absolute path.
- Commit in the worktree at the "Commit" steps of the type and stage files. Do not push.
- Sync: in the worktree, run `git merge --no-commit main`.
  If it fails and `MERGE_HEAD` does not exist, stop. Tell the operator.
  If `MERGE_HEAD` exists, resolve each conflict with the operator and run the tests
  and the linter. If they pass, commit `<id>: merge main`.
  If not, run `git merge --abort`. Tell the operator.
- Merge: do the sync. Then, in the main checkout, on `main` and with no uncommitted changes,
  run `git merge --no-ff <id> -m "Merge issue <id>"`.
  Then, in the worktree, run `git merge --ff-only main`.

### 3.1. Commit message

- Subject: `<id>: <summary>`, in English, in the imperative, with no period at the end,
  in 72 characters or fewer.
- Body, after a blank line: what changed and why, wrapped at 72 characters.
  A small change can have no body.
- The operator is the git author. Each AI model of the issue is a co-author:
  the model of each session, subagent and reviewer, from any vendor.
- The `AI` line of `desc.md` lists the co-authors as `<model> <email>`, separated with `; `.
  Before each commit and at the end of each session, add each missing co-author.
  Use the model name as the model gives it.
- Emails: Anthropic `noreply@anthropic.com`. For another vendor, ask the operator.
- End the message with a blank line and one line `Co-Authored-By: <model> <email>`
  for each co-author. These lines replace the default attribution of the tool.
- A `<id>: merge main` commit has the subject and the `Co-Authored-By` lines.
  A `Merge issue <id>` commit has the subject only.

```
kqtxw-menu-parser: parse menu TOML files

- Add the menu loader and its tests.
- Reject a menu item with no action.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
```

## 4. Gates

- Gate 1: show the plan or the researched solutions.
  Interview the operator about the open questions.
  Agree on the reviewers.
- Gate 2: show the operator the findings and the fixes. If the change has code,
  also show the results of the tests and the linter. The operator selects one option:
  - Return: return to the return step that the type or stage file names.
  - Fix: fix the findings in this session. Then do gate 2 again.
    Run the reviewers again only if the operator asks.
  - Accept: do the next step of the type or stage file.

## 5. Work rules

- Reviewers: give each reviewer a position: what to check and from which view.
  If a change affects security, changes a decision or has a large effect on the design,
  propose more than one reviewer.
- Reviewer brief: give the absolute paths of the files to review, the specification,
  `desc.md` and `<block>.md`. For a change in tracked files, also give the worktree path
  and the command `git diff --merge-base main`. First run `git add -N <path>` for each new file.
  Do not give your own conclusions. Ask the reviewer to give its model name in the report.
- Findings: ask each reviewer for defects only. A defect is a bug, or a break of
  the specification, of the `Acceptance` section, or of a rule in `AGENTS.md` or in this flow.
  A finding gives the place, the defect and a fix.
  Write the findings of each review to `artifacts/<name>-review.md`.
- Specification: if the code differs from the specification, update the specification
  in the same commit.
- Tests: write tests for each unit that you can test separately.
  After you implement or fix code, run the tests and the linter. Commit only if they pass.
