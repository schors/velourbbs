# Type: Dev

A `dev` issue is the parent of blocks. It has no branch and no code.
Each block is a sub-issue with the `block` label ([type-block.md](type-block.md)).

- The order of the sub-issues is the order of execution.
- The first block is the design: title `Design: <name>`.
  It posts the plan of the next blocks as a parent comment.
  After its merge, the parent steps create these blocks.
- `ghflow new block ... --parent <parent>` makes the block a sub-issue
  and makes the parent blocked by it. Add `--after <m>` for each block that it needs.
- After a block, the operator starts the next block with `/ticket <m>` in a new session.

## Create

`/ticket dev` does these steps.

1. Write the parent body from [template-issue.md](template-issue.md).
2. Run `ghflow new dev progress "<title>"`.
3. Create the design block: `ghflow new block todo "Design: <name>" --parent <parent>`.
4. Continue with the design block in this session (`/ticket`, "Create", steps 5 to 8).

## Parent steps

`/ticket-cleanup` of a block and `/ticket <parent>` do these steps.

1. Run `ghflow show <parent>`.
2. For each merged design block `#<d>`, read its parent comment `Plan of #<d>`.
   Create each block of the plan that is not a sub-issue, in the order of the plan,
   before the first open block that is not in the plan:
   `ghflow new block todo "<title>" --parent <parent> --before <m> --after <m>...`.
   With no such open block, omit `--before`. Ask the operator to drop each open block
   that the plan replaces (`/ticket-cleanup drop <m>`).
3. If all blocks are closed, close the parent as `completed`. Close its open sources
   as `completed`. If a block is closed as `not_planned` or `duplicate`,
   ask the operator first.
4. If not, tell the operator the first open block: `/ticket <m>`.
