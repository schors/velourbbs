# Stage: Implementation

1. Read the issue and the next block from `plan.md`.
2. Plan the implementation. Write it to `<block>.md`.
3. Gate 1: interview the operator about open questions. Agree on the reviewers.
   Write them to `Reviews` in `<block>.md`.
4. Implement the block and its tests.
   If the change has code, run the tests and the linter.
   If they fail, fix the cause and run them again.
5. Run the reviewers from gate 1 as subagents.
6. Gate 2: show the operator the findings, the fixes and the results
   of the tests and the linter.
   The operator selects one of these:
   - Go back to step 2.
   - Fix the findings in this session. If the change has code, run the tests and the linter.
     Run the reviewers again only if the operator asks. Then do this gate again.
   - Go back to design: write the reason to the `Scope` section of `desc.md`.
     Ask the operator to keep or discard the changes of the block.
     Set the stage of the issue to `design` and stop.
   - Accept the block.
7. When the operator accepts the block, commit it.
   Merge the branch into `main` (`flow.md`, section 3).
   Then write `<block>.done.md` from [template-block.done.md](template-block.done.md)
   and mark the block in `plan.md`.
8. If all blocks are complete, archive the issue with `/ticket-cleanup <id>`.
   If not, stop. Do the next block in a new session with `/ticket <id>`.
