# Process

## What it does

AI models write most of Velour: the specifications, the code, the tests and the guides.
A human operator controls the work and makes the decisions.
All work is open: the issues, the pull requests and the reviews are in
[github.com/schors/velourbbs](https://github.com/schors/velourbbs).

## How

### Tasks

Each piece of work is a GitHub issue. The type of the issue gives its path:

- `idea`: an idea in the queue;
- `dev`: a large piece of work of several blocks;
- `one-shot`: a small task;
- `docs`: a guide for users.

This section of the station was rewritten in one `one-shot` task.

### Gates

A gate is a stop where the work waits for the answer of the operator.
At the first gate, the model shows the plan and asks questions, and the operator chooses.
The model writes the accepted plan to the issue. So a human makes the decision, and before the code.
The second gate is the pull request. A maintainer accepts it or writes comments.

### Specification first

The rules of the program are in `docs/specs/`. When the code differs from the specification,
the specification changes in the same commit. So each rule has one definition.

### Tests

Tests check each part that can be tested separately.
`go vet`, `go test` and `golangci-lint` run before each commit.

The station test runs `velourbbs check` on the directory `station/`. Then it takes
each screen at 80x24 and 132x43 with 256 colours and compares the cell dump with its golden file.
The golden files are in git, so each change of a screen shows in the pull request.
The test also takes the running light and the ticks of the live fields frame by frame.

### Reviews

AI reviewers read each pull request. Each has its own position: what to check
and from which view. The model posts all reviews to the pull request before the first fix.
In each review thread, it answers: fixed or not, and why.

## What it gets

- An open history of each change: the issue, the plan, the reviews and the answers to them.
- Golden screen files that show each change of the station.
- Commits where each author model is a co-author.

## Limits

- A review is only information. A merge accepts the work, and a maintainer does it, not a model.
- The model works with the token of the operator, so its posts are posts of the operator.
  Each post of a model starts with the line `Agent: <model>`.
