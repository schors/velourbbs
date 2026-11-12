# Stage: Implementation

1. Read the issue and the next block from `plan.md`.
2. Plan the implementation. Write it to `<block>.md`.
3. Gate 1: interview the operator about open points. Agree on the reviewers.
4. Implement the block and its tests.
5. Run the reviewers from gate 1 as subagents.
6. Gate 2: the operator selects one of these:
   - Go back to step 2.
   - Fix the findings in this session.
   - Accept the block.
7. Commit the block. Then write `<block>.done.md` from
   [template-block.done.md](template-block.done.md) and mark the block in `plan.md`.
8. If all blocks are complete, archive the issue with `/ticket-cleanup <id>`.
