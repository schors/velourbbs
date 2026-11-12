# Stage: One-shot

Use this stage for a small task that needs one block.
The issue has no `plan.md` and no block files. `desc.md` holds the plan and the result.

1. The operator describes the task.
2. Create the issue with `/ticket new one-shot`.
3. Write the plan to the `Plan` section of `desc.md`.
   If the task changes behavior, update the specification in `docs/specs/`.
4. Gate 1: interview the operator about open points. Agree on the reviewers.
   Write them to the `Plan` section.
5. Implement the task and its tests.
6. Run the reviewers from gate 1 as subagents.
7. Gate 2: the operator selects one of these:
   - Go back to step 3.
   - Fix the findings in this session.
   - Accept the result.
8. Commit. Write the `Result` section of `desc.md`.
9. Archive the issue with `/ticket-cleanup <id>`.
