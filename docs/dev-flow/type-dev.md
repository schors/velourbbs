# Type: Dev

A `dev` issue is the parent of blocks. It has no branch and no code.
Each block is a sub-issue with the `block` label ([type-block.md](type-block.md)).

- The parent body has the block list as a task list.
  The order of the list is the order of execution.
- The first block is the design: title `Design: <name>`.
  It posts the plan of the next blocks as a parent comment.
  After its merge, create these blocks.
- `ghflow new block ... --parent <parent>` makes the block a sub-issue
  and makes the parent blocked by it. Add `--after <m>` for each block that it needs.

## Create

`/ticket new dev` does these steps.

1. Write the parent body from [template-issue.md](template-issue.md).
   The block list has one item: `Design: <name>`.
2. Run `ghflow new dev progress "<title>"`.
3. Create the design block: `ghflow new block todo "Design: <name>" --parent <parent>`.
   Put its number into the item: `- [ ] #<m>`. Use `ghflow body <parent>`.
4. Continue with the design block in this session (`/ticket`, "Create", steps 5 to 8).

## Open

1. Run `ghflow show <parent>`.
2. For each item with no issue number: if the PR of its design block is merged,
   create the block from the plan comment of that design block:
   `ghflow new block todo "<title>" --parent <parent> --after <m>...`.
   Put its number into the item.
3. Mark the items of closed blocks: `- [x] #<m>`.
4. If all blocks are closed, close the parent as `completed`. Close its open sources
   as `completed`. If a block is closed as `not_planned` or `duplicate`,
   ask the operator first.
5. If not, tell the operator the first open block: `/ticket <m>`.

If the parent has no marker, `ghflow body` refuses. Show the operator the new list
and ask the operator to change the body.
