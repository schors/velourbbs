# Type: One-shot

A small task that one session can complete. `desc.md` holds the plan and the result.

1. Write the plan to the `Plan` section of `desc.md`.
   If the task changes behavior, update the specification in `docs/specs/`.
2. Do gate 1 (`flow.md`, section 4). Write the reviewers to the `Plan` section.
3. Implement the task and its tests.
4. Run the reviewers from gate 1 as subagents.
5. Do gate 2 (`flow.md`, section 4). The return step is step 1.
6. Commit. Write the `Result` section of `desc.md`.
7. Archive the issue with `/ticket-cleanup <id>`.
