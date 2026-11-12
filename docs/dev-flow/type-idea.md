# Type: Idea

An `idea` issue develops an idea with the operator.

1. Read `docs/ideas/` to find related ideas.
2. Interview the operator until the idea is clear.
   Write the decisions of the operator to the `Scope` section of `desc.md`.
   Do not research or select solutions. The design stage does the research.
3. If parts of the idea have no effect on each other, make each part a separate idea.
4. Gate: show the operator the ideas. If the operator asks, run reviewers as subagents.
   If the operator does not accept the ideas, return to step 2.
5. Write each idea to `docs/ideas/<slug>.md`, for example `docs/ideas/menu-nav.md`:
   the goal, the scope, the decisions of the operator and the open questions.
   List the idea files in the `Ideas` line of `desc.md`. Commit.
6. Archive the issue with `/ticket-cleanup <id>`.
