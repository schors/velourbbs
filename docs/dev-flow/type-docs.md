# Type: Docs

A `docs` issue changes user documentation. It does the PR cycle (`flow.md`, section 5)
in one session.

- Run the program and use its real output.
- Gate 1 shows the reader, the outline, a sample of text or a picture, and the reviewers.
- Run only the checks that read the documents: file-block tests, `velourbbs check`,
  golden dumps. Run all checks only if code changed.
- Reviewers: a reader who does the tasks by hand, one reviewer for each language,
  a facts reviewer against the specifications.
- Change all languages in the same commit. Review a large document in parts, in one PR.
- Do not write a specification that repeats a guide.
