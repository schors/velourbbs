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
- Keep it simple: add no checks, options or documentation beyond what a change needs.
- Follow the development flow in `docs/dev-flow/flow.md`.
