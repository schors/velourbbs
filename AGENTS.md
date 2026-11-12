# Working instructions

- Speak briefly in Russian. Write code comments and repository documentation in English.
- A "document for AI" is a file in one of these paths: `AGENTS.md`, `docs/dev-flow/`,
  `.claude/skills/`, `docs/specs/`, `docs/ideas/`, `.ai/issues.md`, `.ai/artifacts/`
  and `.ai/issues/<id>/`.
- The "style" is six qualities of text. A document for AI must have them.
  For other English text, the style is a guideline.
  - Brevity: a sentence has no word that you can remove without a loss of meaning.
  - Compactness: a document has no sentence that you can remove without a loss for the reader.
    It keeps each precondition and each exception.
  - Coherence: the order of the parts has a reason. Each part serves the goal of the document.
  - Expressiveness: words are exact and concrete. A format with a `<name>` part has an example.
    A rule that a reader can misread has an example.
  - Consistency: in all documents for AI, one thing has one term and one form,
    and one term has one meaning.
  - Clarity: a reader without the session context understands each sentence in one reading.
- Use these techniques to get the style. The files in `docs/dev-flow/` are examples.
  The Sentences and Lists techniques can yield to a quality. The other techniques have
  no exceptions.
  - Wording (from ASD-STE100, Simplified Technical English): use the active voice.
    Write instructions in the imperative. Do not use phrasal verbs.
  - Sentences: use up to 20 words for an instruction and 25 for a description.
    Put one instruction or one fact in each sentence. Count a literal in backticks as one word.
    Headings, labels and table cells can be fragments.
  - Lists: use a numbered list for steps in sequence, a bullet list for rules,
    a table for a mapping and a paragraph for a description.
  - Terms: define a term in quotation marks.
  - References: refer to a rule by its file and section.
    Number the sections that other files refer to.
  - Literals: put paths, commands and literal values in backticks.
    Write a variable part as `<name>`.
  - Templates: write each fill instruction as an HTML comment.
  - Markup: do not use bold, italics or emoji. Keep each line at 100 characters or fewer.
    Table rows, front matter lines and URLs can be longer.
- Apply the style to text that you write or change. Do not rewrite other text only for style.
- Do not change code, identifiers, commands, grammars or quoted output to fit the style.
- Write Russian replies in plain language. This is a guideline:
  - Use short sentences, with one fact in each. Use verbs, not verbal nouns
    ("проверил", not "была произведена проверка").
  - Do not use filler, bureaucratic phrases or evaluations without facts.
  - Use one term for each thing. Keep established English terms and identifiers as they are.
- Keep changes in scope. Run the relevant checks.
- Never weaken or bypass a check, precondition or its test to make a change pass.
- Keep one definition for each rule. Keep only current rules and necessary usage instructions.
  Remove research and review drafts after use.
- The "principles" are the "principle qualities" and the rules below.
  A "work product" is a design, an implementation decision, code or a document.
  Apply the principles to each work product that you write or change.
  In the principles, the "goal" of a work product is the part of the `Goal` section
  of `desc.md` or of the specification that it serves.
  The tests and the linter that `docs/dev-flow/flow.md`, section 5, requires are part of
  the goal of code.
  Validation of untrusted input is part of the goal of the code that reads the input.
  The "standard library" is the library in the distribution of the language.
  - Principle qualities:
    - Fit: it meets its goal.
    - Clear: a reader who knows the language or the format, its standards and the
      specifications, but not the session context, understands it in one reading.
      For a document for AI, Clear means the style quality Clarity.
    - Minimal: it has no option, feature, validation or documentation that its goal
      does not need.
    - Compact: it has no part that you can remove or merge without a loss.
      For a document for AI, Compact means the style quality Compactness.
    - Idiomatic: it uses the idioms of its language or format. An "idiom" is a form that
      the authors of the language or format use in its documentation,
      or in the standard library of the language.
      Example: in Go, report an expected failure with an `error` value, not a panic.
    - Efficient: its time, memory and I/O grow with the size of the input no faster
      than the goal needs. It meets each limit in the specification.
    - Elegant: one idea covers all the cases, or one decision removes a problem
      instead of handling it.
      Example: a linked list with a sentinel head node needs no special case for the first item.
  - Rules:
    - Priority: if two principle qualities conflict, Fit has priority, then Clear.
      The Technology rule and the Libraries rule have priority over the Typical solutions rule.
      For any other conflict, select one option, write a reason and show the conflict
      to the operator at the next gate (`docs/dev-flow/flow.md`, section 4).
    - Effort: the work to get the principle qualities can be large and complex.
      It can include research, a comparison of options and a second version of a work product.
      It stays inside the steps of the type or stage file.
      The scope of the issue limits the change, not the work.
      The operator can limit the work at a gate.
    - Typical solutions: a "typical" solution is the one that most projects use.
      Compare solutions by the principle qualities, not by popularity.
      If two solutions are equal, select the typical one.
      If a design or an implementation decision selects a solution that is not typical,
      write a reason. The reason names the typical solution and its defect that the choice removes.
    - Technology: this rule is for a protocol or a format that this project does not define.
      Its "standard" is the set of documents that define it, update it or state requirements for it.
      In the specification, name its "target implementations": the programs that this project
      must work with. Example: a list of telnet clients with their versions.
      If only this project reads the format, the target implementation is its parser.
      If the specification names none, ask the operator.
      Send or write a form that the standard defines and the target implementations accept.
      If there is no standard or no such form, use the form of the target implementations.
      Accept each form that the target implementations send or write.
      Use a feature that a target implementation does not support only with a fallback
      for that implementation.
      If you deviate from the standard, write a reason that names the target implementations.
    - Libraries: use the standard library first.
      Example: in Go, the `golang.org/x/` libraries are not parts of the standard library.
      If the documentation of the standard library recommends another library for a task,
      use that library.
      If the standard library does not meet the goal, compare libraries and new project code.
      Of the options that meet the goal, select the one with the lowest cost of maintenance.
      A "maintained" library has authors that fix defects. It works with the current version
      of the language. It documents each breaking change of its API.
      Use only a maintained library.
      Write a reason for each choice between a library and new project code.
      The reason names the compared options.
    - Reasons: write each reason when you write the work product that it explains.
      Write it to the specification, to the `Task` or `Implementation plan` section
      of `<block>.md`, or to the `Plan` section of `desc.md`.
      Repeat it in the body of the next commit message.
- Follow the development flow in `docs/dev-flow/flow.md`.
