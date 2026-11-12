# Stage: Implementation

1. Read the issue and the next block from `plan.md`.
2. Plan the implementation. Write it to `<block>.md`.
3. Gate 1: interview the operator about open points. Agree on the reviewers.
   Write them to `Reviews` in `<block>.md`.
4. Implement the block and its tests.
5. Run the reviewers from gate 1 as subagents.
6. Gate 2: the operator selects one of these:
   - Go back to step 2.
   - Fix the findings in this session.
   - Go back to design: write the reason to `<block>.md`, set the stage of the issue
     to `design` and stop.
   - Accept the block.
7. Commit the block. Merge the branch into `main` (`flow.md`, section 3).
   Then write `<block>.done.md` from [template-block.done.md](template-block.done.md)
   and mark the block in `plan.md`.
8. If all blocks are complete, archive the issue with `/ticket-cleanup <id>`.
