# Stage: Decomposition

1. Read `desc.md`, `plan.md` if it exists, and the specifications
   from the `Specification` line of `desc.md`.
2. Divide the design into blocks.
3. Gate: show the operator the blocks. If the operator does not accept the blocks, return to step 2.
4. Write a `<block>.md` file for each block from [template-block.md](template-block.md).
   If `plan.md` exists, update the blocks that are not done. Keep the done blocks.
5. Write `plan.md` from [template-plan.md](template-plan.md).
   Put the blocks in the order of execution.
