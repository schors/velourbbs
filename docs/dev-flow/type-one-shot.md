# Type: One-shot

Use this type for a small task that one session can complete.
The issue has no stages, no `plan.md` and no block files. `desc.md` holds the plan and the result.

1. Write the plan to the `Plan` section of `desc.md`.
   If the task changes behavior, update the specification in `docs/specs/`.
2. Gate 1: interview the operator about open questions. Agree on the reviewers.
   Write them to the `Plan` section.
3. Implement the task and its tests.
4. Run the reviewers from gate 1 as subagents.
5. Gate 2: the operator selects one of these:
   - Go back to step 1.
   - Fix the findings in this session.
   - Accept the result.
6. Commit. Write the `Result` section of `desc.md`.
7. Archive the issue with `/ticket-cleanup <id>`.
