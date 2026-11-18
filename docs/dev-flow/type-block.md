# Type: Block

A block is one sub-issue of a `dev` issue ([type-dev.md](type-dev.md)).
It does the PR cycle (`flow.md`, section 5).

- A block is one unit of work. An agent can do it without help.
  A reviewer can review it in one pass. After its merge, `main` works.
  A block is not the smallest possible step.
- A feature block updates the guides that it affects. Its `Acceptance` names them.

## Design block

The title is `Design: <name>`. Its PR puts the specifications into `main`.

1. Read the parent and its source issues. Research the solutions for the open questions.
2. Gate 1 also shows the blocks after the design.
3. Write the design and the specifications to `docs/specs/`.
4. Before step 4 of the PR cycle, post the plan of the next blocks as a parent comment
   that starts with `Plan of #<n>`. List the blocks in the order of execution.
   For each block: the title, the blocks that it needs,
   the specification section, the task and the acceptance.
   Name each open block that the plan replaces.
   After the merge, the parent steps create the blocks.

Rules for the specifications:

- Libraries: use the standard library first. Use another library only if its authors
  maintain it and document each breaking change. Write the reason to the specification.
- Formats: for a protocol or a format that this project does not define, name
  the "target implementations" in the specification: the programs that this project
  must work with. Send a form of the standard that they accept. If there is none,
  send their form and write the reason. Accept each form that they send.

## Wrong specification

If a block shows that the specification is wrong, tell the operator.
Create a block `Design: <name>` before the first open block:
`ghflow new block todo "Design: <name>" --parent <parent> --before <m>`.
Ask the operator to keep this block open or to drop it (`/ticket-cleanup drop <n>`).
Tell the operator the new design block: `/ticket <m>`.
