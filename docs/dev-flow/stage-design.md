# Stage: Design

1. Read `desc.md` and the idea files from its `Ideas` line.
   If there are no idea files, use the `Goal` and `Scope` sections.
2. Research the solutions for the open questions.
3. Gate 1: interview the operator about open questions. Agree on the reviewers.
4. Select the solutions. Write the design and the specifications to `docs/specs/`.
   List the specification files in `desc.md`.
   Delete the idea files that the specifications replace.
5. Run the reviewers from gate 1 as subagents.
6. Gate 2: show the operator the findings. Go back to step 3 until the operator confirms.
7. Commit.

Result: an issue with a design and specifications, ready for decomposition.
