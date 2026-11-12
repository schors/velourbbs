# Principles

The "principles" are the "principle qualities" and the rules below.
A "work product" is a design, an implementation decision, code or a document.
Apply the principles to each work product that you write or change.
The "goal" of a work product is the part of the `Goal` section of `desc.md`
or of the specification that it serves.
The goal of code includes the tests and the linter that `flow.md`, section 5, requires.
The goal of code that reads untrusted input includes its validation.

## 1. Principle qualities

For a document for AI, Clear means Clarity and Compact means Compactness ([style.md](style.md)).

| Quality | Test |
|---|---|
| Fit | It meets its goal. |
| Clear | A reader who knows the language or the format, its standards and the specifications, but not the session context, understands it in one reading. |
| Minimal | It has no option, feature, validation or documentation that its goal does not need. |
| Compact | It has no part that you can remove or merge without a loss. |
| Idiomatic | It uses the forms that the authors of the language or format use in its documentation or its standard library. Example: in Go, report an expected failure with an `error` value, not a panic. |
| Efficient | Its time, memory and I/O grow with the size of the input no faster than the goal needs. It meets each limit in the specification. |
| Elegant | One idea covers all the cases, or one decision removes a problem instead of handling it. Example: a linked list with a sentinel head node needs no special case for the first item. |

## 2. Rules

- Priority: if two principle qualities conflict, Fit has priority, then Clear.
  The Technology and Libraries rules have priority over the Typical solutions rule.
  For any other conflict, select one option, write a reason and show the conflict
  to the operator at the next gate (`flow.md`, section 4).
- Effort: the work to get the principle qualities can be large. It can include research,
  a comparison of options and a second version. It stays inside the steps of the type
  or stage file. The scope of the issue limits the change, not the work.
  The operator can limit the work at a gate.
- Limits: a "meaningful limit" comes from a decision of the operator, a specification,
  a standard, `AGENTS.md` or the flow. Each other limit is "accidental".
  Examples: the current split of code into files and functions, a scope or a plan
  that the operator did not accept, the context size of the model.
  Make each decision as you would make it without the accidental limits.
  Example: if one tool call accepts up to 4 questions, ask the other questions in the next call.
  If the decision crosses a meaningful limit, such as an accepted scope,
  select the best option inside the limit. At the next gate, show the operator
  the option outside the limit and the defect that this option removes.
- Typical solutions: compare solutions by the principle qualities, not by popularity.
  If two solutions are equal, select the "typical" one: the one that most projects use.
  If you select a solution that is not typical, write a reason.
  The reason names the typical solution and its defect that your choice removes.
- Technology: this rule is for a protocol or a format that this project does not define.
  Its "standard" is the set of documents that define it, update it or state requirements for it.
  In the specification, name its "target implementations": the programs that this project
  must work with. Example: a list of telnet clients with their versions.
  If only this project reads the format, the target implementation is its parser.
  If the specification names none, ask the operator.
  Send or write a form that the standard defines and the target implementations accept.
  If there is no such form, use the form of the target implementations.
  Accept each form that the target implementations send or write.
  Use a feature that a target implementation does not support only with a fallback
  for that implementation.
  If you deviate from the standard, write a reason that names the target implementations.
- Libraries: use the standard library of the language first.
  Example: in Go, `golang.org/x/` is not in the standard library.
  If its documentation recommends another library for the task, use that library.
  If the standard library does not meet the goal, compare libraries and new project code.
  Of the options that meet the goal, select the one with the lowest cost of maintenance.
  Use only a library whose authors fix defects, support the current version of the language
  and document each breaking change.
  Write a reason for the choice between a library and new project code.
  The reason names the compared options.
- Reasons: write each reason when you write the work product that it explains.
  Write it to the specification, to the `Task` or `Implementation plan` section
  of `<block>.md`, or to the `Plan` section of `desc.md`.
  Repeat it in the body of the next commit message.
