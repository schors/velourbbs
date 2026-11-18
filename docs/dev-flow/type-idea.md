# Type: Idea

An idea is an issue with `idea`, `backlog` and a priority (`flow.md`, section 2).
It has no branch and no PR.

1. Run `ghflow list idea`. Read the related ideas with `ghflow show`.
2. Interview the operator until the idea is clear.
   Do not research or select solutions. The design block does that.
3. If parts of the idea have no effect on each other, make each part a separate idea.
4. Gate: show the operator each idea: the goal, the scope, the decisions of the operator,
   the open questions and the priority. If the operator does not accept them, return to step 2.
5. Create each idea with `ghflow new idea backlog "<title>"`
   and the body from [template-issue.md](template-issue.md).
   Set its priority with `ghflow priority`.
   Give the operator the links.
   If you opened an existing idea, change it, do not create it again:
   `ghflow body` for an idea with the marker, a comment for another idea.

To start the work on an idea, create a `dev`, `one-shot` or `docs` issue
with `Source: #<n>` (`flow.md`, section 7).
