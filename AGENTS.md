# Working instructions

- Speak briefly in Russian. Write code comments and repository documentation in English.
- Write Russian replies in plain language. This is a guideline:
  - Use short sentences, with one fact in each. Use verbs, not verbal nouns
    ("проверил", not "была произведена проверка").
  - Do not use filler, bureaucratic phrases or evaluations without facts.
  - Use one term for each thing. Keep established English terms and identifiers as they are.
  - Name an item by its content, not only by its number:
    "вариант 2 (merge main в worktree)", not "вариант 2".
- Write English documents in plain style (ASD-STE100 as a guideline): short sentences
  with one fact in each, active voice, imperative instructions, one term for each thing.
  This is not a review check. Do not rewrite other text only for style.
- Keep changes in scope. Run the relevant checks.
- Never weaken or bypass a check, precondition or its test to make a change pass.
- Keep it simple: add no check, option or documentation that the change does not need.
- Do not shape a decision by an accidental limit: the current code layout,
  a plan that the operator did not accept or a limit of a tool.
  If the better option is outside the accepted scope, show it at the next gate.
- Keep one definition for each rule. Keep only current rules and necessary usage instructions.
  Remove research and review drafts after use.
- Follow the development flow in `docs/dev-flow/flow.md`.
