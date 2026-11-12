# Style

The "style" is six qualities of text. A document for AI must have them.
For other English text, the style is a guideline.

| Quality | Test |
|---|---|
| Brevity | A sentence has no word that you can remove without a loss of meaning. |
| Compactness | A document has no sentence that you can remove without a loss for the reader. It keeps each precondition and each exception. |
| Coherence | The order of the parts has a reason. Each part serves the goal of the document. |
| Expressiveness | Words are exact and concrete. A format with a `<name>` part has an example. A rule that a reader can misread has an example. |
| Consistency | In all documents for AI, one thing has one term and one form, and one term has one meaning. |
| Clarity | A reader without the session context understands each sentence in one reading. |

Use these techniques. The files in `docs/dev-flow/` are examples.
Only the Sentences and Lists techniques can yield to a quality, and only where they break it.

- Wording (ASD-STE100): use the active voice. Write instructions in the imperative.
  Do not use phrasal verbs.
- Sentences: put one instruction or one fact in each sentence. Use up to 20 words
  for an instruction and 25 for a description. Count a literal in backticks as one word.
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

Apply the style to text that you write or change. Do not rewrite other text only for style.
Do not change code, identifiers, commands, grammars or quoted output to fit the style.
