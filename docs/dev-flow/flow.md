# Development flow

This document defines the issue types, the stages, the files, the git rules, the gates
and the work rules.
Read this file first. Then read only one more file: the file of the issue type,
or the file of the current stage for a `dev` issue.

## 1. Issue types and stages

| Type | Stage | File | Result |
|---|---|---|---|
| `idea` | `-` | [type-idea.md](type-idea.md) | Ideas in `docs/ideas/`. |
| `dev` | `design` | [stage-design.md](stage-design.md) | Specifications in `docs/specs/`. |
| `dev` | `decomposition` | [stage-decomposition.md](stage-decomposition.md) | `plan.md` and block files. |
| `dev` | `implementation` | [stage-implementation.md](stage-implementation.md) | One block, implemented and reviewed. |
| `one-shot` | `-` | [type-one-shot.md](type-one-shot.md) | A small task, from description to result. |

- Only a `dev` issue has stages. The stages go in the order of the table.
- `docs/ideas/` is the backlog. An `idea` issue writes ideas to it. A `dev` issue takes
  one or more ideas from it.
- Start an issue with `/ticket new <type>`. In the same session, do the steps of the type file,
  or of the design stage file for a `dev` issue.
- Do each next stage of a `dev` issue in a new session. Start it with `/ticket <id>`.
  The 5 letters of the id are sufficient.
- "Operator" is the human who controls the work.

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
| `.ai/archive/<id>/` | no | A complete issue. |
| `.ai/worktrees/<id>/` | no | The git worktree of the issue. |

- The issue `<id>` is 5 random lowercase letters, a hyphen and a slug.
  The 5 letters do not include `i`, `l` and `o`. Example: `kqtxw-menu-parser`.
- Make the 5 letters with `LC_ALL=C tr -dc 'a-hjkmnp-z' </dev/urandom | head -c5`.
- A "block" is one unit of work with one complex task. An agent can do it without help.
  Each reviewer can review it in one pass. After its merge, `main` works.
  A block is not the smallest possible step.
- A `<block>` name is a slug. Example: `toml-loader`.
- Each issue with `plan.md` has a `<block>.md` file for each block.
- `/ticket new` sets the stage in `.ai/issues.md`: `design` for a `dev` issue,
  `-` for the other types.
- At the end of a `dev` stage, set the next stage in `.ai/issues.md`.
  Then stop. Tell the operator to run `/ticket <id>`.
- The stage `archived` or `dropped` is for all types. Only `/ticket-cleanup` sets it.
- Move a complete issue directory to `.ai/archive/` with `/ticket-cleanup <id>`.
  To stop an incomplete issue, run `/ticket-cleanup drop <id>`.

### 2.1. Complete issue

| Type | Complete when |
|---|---|
| `idea` | `desc.md` lists one or more ideas. All of them are in `docs/ideas/` on the branch of the issue. |
| `dev` | The stage is `implementation`. `plan.md` has one or more blocks. All blocks are done. Each block has `<block>.done.md`. |
| `one-shot` | `desc.md` has a filled `Result` section. |

Also, the worktree of a complete issue has no uncommitted changes.

## 3. Git

- Each issue has a branch `<id>` from `main` and a worktree in `.ai/worktrees/<id>/`.
  `/ticket new` creates them.
- Change tracked files only in the worktree of the issue.
- Change this flow (`docs/dev-flow/`, `.claude/skills/`) and `AGENTS.md` only in a one-shot issue.
  Exception: each issue can add a vendor row to the table in section 3.1.
- The "main checkout" is the first line of `git worktree list`.
  The `.ai/` files are only in the main checkout. In the worktree, use their absolute path.
- Commit in the worktree at the "Commit" steps of the type and stage files.
  Write each commit message as section 3.1 tells.
- Merge the branch into `main` after each block commit. `/ticket-cleanup` merges the rest
  of the branch, removes the worktree and deletes the branch.
- To merge, use the main checkout. It must be on `main` and have no uncommitted changes.
  Run `git merge --no-ff <id> -m "Merge issue <id>"`. Then run the tests in the main checkout.
- If the merge has a conflict, run `git merge --abort`. In the worktree, run `git merge main`.
  Resolve the conflict with the operator. Commit with `<id>: merge main`. Then merge again.
- Do not push.

### 3.1. Commit message

- Write the subject as `<id>: <summary>`. Write the summary in English, in the imperative.
  Do not put a period at the end. The subject has 72 characters or fewer.
- Put a blank line after the subject. Then write the body: what changed and why.
  The body can be a list. A small change can have no body. Wrap the body at 72 characters.
- The operator is the git author. Each AI model of the issue is a co-author:
  the model of each session, subagent and reviewer, from any vendor.
- The `AI` line of `desc.md` has the list of co-authors. Separate the items with `; `.
  Example: `- AI: Claude Opus 5.5 <noreply@anthropic.com>; Fable 5.2 <noreply@anthropic.com>`.
  If `desc.md` has no `AI` line, add it.
- At the end of each session and before each commit, add each missing co-author
  to the `AI` line.
- Use the model name as the model gives it, for example `Claude Opus 5.5`.
  Use the email of the vendor from the table below.
  If the vendor is not in the table, ask the operator for the email.
  Add a row to the table (section 3, exception).
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

## 4. Gates

A "gate" is a stop where the operator must reply before the work continues.
This section defines gate 1 and gate 2. A type or stage file defines its other gates
in its steps.

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

- Review: if a change affects security, changes a decision or has a large effect on the design,
  propose more than one reviewer.
- Reviewer brief: give each reviewer a "position": what to check and from which view.
  Give the absolute paths of the files to review, the specification, `desc.md` and `<block>.md`.
  For a change in tracked files, also give the worktree path.
  Run `git add -N <path>` for each new file of the change.
  Then give the command `git diff --merge-base main`.
  Do not give your own conclusions. Ask the reviewer to give its model name in the report.
- Style: if a change affects a tracked document for AI, propose the reviewer position `style`.
  `AGENTS.md` defines a document for AI and the style.
  Check the style of the files in `.ai/` yourself.
- Principles: for each design and each change with code, propose the reviewer position
  `principles`. `AGENTS.md` defines the principles.
- Write the findings of each review to `artifacts/<name>-review.md`.
- Specification: if the code differs from the specification, update the specification
  in the same commit.
- Tests: write tests for each unit that you can test separately.
  After you implement or fix code, run the tests and the linter.
  If they fail, fix the cause and run them again. Commit only if they pass.
