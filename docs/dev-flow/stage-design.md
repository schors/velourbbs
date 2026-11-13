# Stage: Design

1. Read `desc.md` and the idea files from its `Ideas` line.
   If there are no idea files, use the `Goal` and `Scope` sections.
2. Research the solutions for the open questions.
3. Do gate 1 (`flow.md`, section 4).
4. Select the solutions. Write the design and the specifications to `docs/specs/`.
   List the specification files in the `Specification` line of `desc.md`.
   Delete the idea files that the specifications replace.
5. Run the reviewers from gate 1 as subagents.
6. Do gate 2 (`flow.md`, section 4). The return step is step 3.
7. Commit.

Rules for step 4:

- Libraries: use the standard library first. Use another library only if its authors
  maintain it and document each breaking change. Write the reason to the specification.
- Formats: for a protocol or a format that this project does not define, name
  the "target implementations" in the specification: the programs that this project
  must work with. Send a form of the standard that they accept. If there is none,
  send their form and write the reason. Accept each form that they send.
