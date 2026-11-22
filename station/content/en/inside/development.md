# Development

Velour is written mostly with AI. The AI models write the specifications, the code
and the guides. A human operator controls the work and makes the decisions.
All the work is open: the issues, the pull requests and the reviews are in [github.com/schors/velourbbs](https://github.com/schors/velourbbs).

## Issues

Each piece of work is a GitHub issue. The type of the issue gives its path:

- `idea` - an idea in the backlog;
- `dev` - a large work in several blocks;
- `one-shot` - a small task;
- `docs` - a guide for the users.

This section of the station was made in one `one-shot` issue.

## Gates

A gate is a stop where the work waits for the operator.
At the first gate, the model shows the plan and asks its questions, and the operator chooses.
The model posts the accepted plan in the issue. So a human makes the decision,
and before the code.

The second gate is the pull request. A maintainer merges it or writes comments.

## Specification first

The rules of the program are in `docs/specs/`. If the code differs from the specification,
the specification changes in the same commit. So each rule has one definition.
Tests check each part that can be tested separately.

## Reviews

AI reviewers read each pull request. Each has a position: what to check and from which view.
The model posts all reviews in the pull request before the first fix.
In each review thread, it replies: fixed, or not fixed and why.
So each change keeps a public history.

A review is information only. The merge accepts the work, and a maintainer merges,
not a model. Each commit names the model as a co-author.
