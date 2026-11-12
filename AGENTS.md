# Working instructions

- Speak briefly in Russian; write code comments and repository documentation in English.
- A "document for AI" is a file in one of these paths: `AGENTS.md`, `docs/dev-flow/`,
  `.claude/skills/`, `docs/specs/`, `docs/ideas/`, `.ai/issues.md`, `.ai/artifacts/`
  and `.ai/issues/<id>/`.
- The "style" is the wording rules and the structure rules below.
  A document for AI must follow the style.
  For other English text, the style is a guideline: use another form if the content needs it.
- Apply the style to text that you write or change. Do not rewrite other text only for style.
- Do not change code, identifiers, commands, grammars or quoted output to fit the style.
- Wording rules (ASD-STE100, Simplified Technical English):
  - Use short sentences: up to 20 words for instructions, 25 for descriptions.
    Count a literal in backticks as one word.
  - Put one instruction or one fact in each sentence.
  - Use the active voice. Write instructions in the imperative.
  - Use one term for each thing, with one meaning. Avoid synonyms and phrasal verbs.
  - Headings, labels and table cells can be fragments.
- Structure rules (the files in `docs/dev-flow/` are examples):
  - Use a numbered list for steps in sequence. Use a bullet list for rules.
    Use a table for a mapping. Use a paragraph for a description.
  - Number the sections that other files refer to.
  - Define a term one time, in quotation marks. Then use only this term.
  - Refer to a rule by its file and section.
  - Put paths, commands and literal values in backticks. Write a variable part as `<name>`.
    Give an example for a file name or a line format that has a `<name>` part.
  - In a template, write each fill instruction as an HTML comment.
  - Do not use bold, italics or emoji.
  - Keep each line at 100 characters or fewer. Table rows, front matter lines and URLs
    can be longer.
- Write Russian replies in plain language. This is a guideline:
  - Use short sentences, with one fact in each. Use verbs, not verbal nouns
    ("проверил", not "была произведена проверка").
  - Do not use filler, bureaucratic phrases or evaluations without facts.
  - Use one term for each thing. Keep established English terms and identifiers as they are.
- Keep changes scoped and run relevant checks.
- Never weaken or bypass a check, precondition or its test to make a change pass.
- Keep one definition for each rule. Retain only current rules and necessary usage
  instructions; remove research and review drafts after use.
- Keep it simple: add no checks, options or documentation beyond what a change needs.
- Follow the development flow in `docs/dev-flow/flow.md`.
