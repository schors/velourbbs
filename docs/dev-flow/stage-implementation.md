# Stage: Implementation

1. Do the sync (`flow.md`, section 3).
2. Read `desc.md`, `plan.md` and the `<block>.md` file of the first block that is not done.
   Read the specification from the `Specification` line of `<block>.md`.
3. Plan the implementation. Write it to the `Implementation plan` section of `<block>.md`.
4. Do gate 1 (`flow.md`, section 4). Write the reviewers to the `Reviews` line of `<block>.md`.
5. Implement the block and its tests.
6. Run the reviewers from gate 1 as subagents.
7. Do gate 2 (`flow.md`, section 4). The return step is step 3. Offer one more option:
   - Return to design: write the reason to the `Scope` section of `desc.md`.
     Ask the operator to keep or discard the changes of the block.
     Set the stage of the issue to `design`. Then stop.
8. Commit the block. Do the merge (`flow.md`, section 3).
   Then write `<block>.done.md` from [template-block.done.md](template-block.done.md).
   Mark the block in `plan.md`.
9. If all blocks are done, archive the issue with `/ticket-cleanup <id>`.
   If not, stop. Tell the operator to run `/ticket <id>` in a new session.
