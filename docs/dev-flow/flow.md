# Development flow

Read this file, [style.md](style.md) and [principles.md](principles.md) first.
Then read only one more file: the file of the issue type, or of the current stage
for a `dev` issue.

## 1. Issue types and stages

| Type | Stage | File | Result |
|---|---|---|---|
| `idea` | `-` | [type-idea.md](type-idea.md) | Ideas in `docs/ideas/`. |
| `dev` | `design` | [stage-design.md](stage-design.md) | Specifications in `docs/specs/`. |
| `dev` | `decomposition` | [stage-decomposition.md](stage-decomposition.md) | `plan.md` and block files. |
| `dev` | `implementation` | [stage-implementation.md](stage-implementation.md) | One block, implemented and reviewed. |
| `one-shot` | `-` | [type-one-shot.md](type-one-shot.md) | A small task, from description to result. |

- "Operator" is the human who controls the work.
- Only a `dev` issue has stages. The stages go in the order of the table.
- `docs/ideas/` is the backlog. An `idea` issue writes ideas to it. A `dev` issue takes
  one or more ideas from it.
- Start an issue with `/ticket new <type>`. In the same session, do the steps of the type file,
  or of the design stage file for a `dev` issue.
- Do each next stage of a `dev` issue in a new session. At the end of a stage,
  set the next stage in `.ai/issues.md`. Then stop. Tell the operator to run `/ticket <id>`.
  The 5 letters of the id are sufficient.
- `/ticket-cleanup <id>` archives a complete issue. `/ticket-cleanup drop <id>` stops
  an incomplete issue. Only `/ticket-cleanup` sets the stage `archived` or `dropped`, for any type.

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
| `.ai/issues/<id>/artifacts/<name>-review.md` | no | Review findings. `<name>` is the type, the stage or the block. Example: `toml-loader-review.md`. |
| `.ai/issues/<id>/plan.md` | no | Index of blocks. Template: [template-plan.md](template-plan.md). |
| `.ai/issues/<id>/<block>.md` | no | Block description. Template: [template-block.md](template-block.md). |
| `.ai/issues/<id>/<block>.done.md` | no | Block result. Template: [template-block.done.md](template-block.done.md). |
| `.ai/archive/<id>/` | no | An archived or dropped issue. |
| `.ai/worktrees/<id>/` | no | The git worktree of the issue. |

- The issue `<id>` is 5 random lowercase letters (not `i`, `l` or `o`), a hyphen and a slug.
  Example: `kqtxw-menu-parser`.
- A "block" is one unit of work with one complex task. An agent can do it without help.
  Each reviewer can review it in one pass. After its merge, `main` works.
  A block is not the smallest possible step.
  A `<block>` name is a slug. Example: `toml-loader`.

## 3. Git

- Each issue has a branch `<id>` from `main` and its worktree. `/ticket new` creates them.
- Change tracked files only in the worktree of the issue.
- Change this flow (`docs/dev-flow/`, `.claude/skills/`) and `AGENTS.md` only in a one-shot issue.
  Exception: each issue can add a vendor email to section 3.1.
- The "main checkout" is the first line of `git worktree list`.
  The `.ai/` files are only in the main checkout. In the worktree, use their absolute path.
- Commit in the worktree at the "Commit" steps of the type and stage files.
  Write each commit message as section 3.1 tells.
- Merge the branch into `main` after each block commit.
  `/ticket-cleanup` merges the rest of the branch.
- To merge, use the main checkout. It must be on `main` and have no uncommitted changes.
  Run `git merge --no-ff <id> -m "Merge issue <id>"`.
  Run the tests in the main checkout only if the branch has a code change
  and `main` had commits outside the branch.
  A "code change" changes a file that the build or the tests read.
- If the merge has a conflict, run `git merge --abort`. In the worktree, run `git merge main`.
  Resolve the conflict with the operator.
  If the branch has a code change, run the tests in the worktree.
  Commit with `<id>: merge main`. Then merge again.
- Do not push.

### 3.1. Commit message

- Write the subject as `<id>: <summary>`: in English, in the imperative, with no period
  at the end, in 72 characters or fewer.
- After a blank line, write the body: what changed and why. Wrap it at 72 characters.
  The body can be a list. A small change can have no body.
- The operator is the git author. Each AI model of the issue is a co-author:
  the model of each session, subagent and reviewer, from any vendor.
- The `AI` line of `desc.md` lists the co-authors as `<model> <email>`, separated with `; `.
  Before each commit and at the end of each session, add each missing co-author.
  If `desc.md` has no `AI` line, add it. Use the model name as the model gives it.
  Example: `- AI: Claude Opus 5.5 <noreply@anthropic.com>; Fable 5.2 <noreply@anthropic.com>`.
- Vendor emails: Anthropic `noreply@anthropic.com`. For another vendor, ask the operator
  and add the email here (section 3, exception).
- End the message with a blank line and one line `Co-Authored-By: <model> <email>`
  for each co-author. Do not wrap these lines. They replace the default attribution of the tool.
- A `<id>: merge main` commit has the subject and the `Co-Authored-By` lines.
  A `Merge issue <id>` commit has the subject only.

Example:

```
kqtxw-menu-parser: parse menu TOML files

- Add the menu loader and its tests.
- Reject a menu item with no action.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
```

## 4. Gates

A "gate" is a stop where the operator must reply before the work continues.
A type or stage file defines gates other than gate 1 and gate 2 in its steps.

- Gate 1: interview the operator about the open questions of the plan or the design.
  Agree on the reviewers.
- Gate 2: show the operator the findings and the fixes. If the change has code,
  also show the results of the tests and the linter. The operator selects one option:
  - Return: return to the "return step" that the type or stage file names.
  - Fix: fix the findings in this session. Run the reviewers again only if the operator asks.
    Then do gate 2 again.
  - Accept: do the next step of the type or stage file.

  A type or stage file can add an option.

## 5. Work rules

- Reviewers: give each reviewer a "position": what to check and from which view.
  Propose the position `principles` for each design and each change with code.
  Propose the position `style` for each change in a tracked document for AI.
  Check the style of the files in `.ai/` yourself.
  If a change affects security, changes a decision or has a large effect on the design,
  propose more than one reviewer.
- Reviewer brief: for `style`, name [style.md](style.md).
  For `principles`, name [principles.md](principles.md).
  Give the absolute paths of the files to review, the specification, `desc.md` and `<block>.md`.
  For a change in tracked files, also give the worktree path and the command
  `git diff --merge-base main`. First run `git add -N <path>` for each new file.
  Do not give your own conclusions. Ask the reviewer to give its model name in the report.
- Write the findings of each review to `artifacts/<name>-review.md`.
- Specification: if the code differs from the specification, update the specification
  in the same commit.
- Tests: write tests for each unit that you can test separately.
  After you implement or fix code, run the tests and the linter.
  If they fail, fix the cause and run them again. Commit only if they pass.
