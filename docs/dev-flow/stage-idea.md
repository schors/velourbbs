# Stage: Idea

The `/idea` skill starts this stage.

1. The operator describes the idea.
2. Create the issue with `/ticket new idea`.
3. Interview the operator until the idea is clear.
4. Select one or more reviewers. Run them as subagents.
5. Go back to step 3 until the operator confirms.
6. Write each idea to `docs/ideas/<slug>.md`. List the idea files in `desc.md`. Commit.
7. Archive the issue with `/ticket-cleanup <id>`.

Result: one or more ideas, ready for design.
Make two ideas when they have no effect on each other.
