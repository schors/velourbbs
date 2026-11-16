# Idea: home BBS of Velour

Backlog. It applies the [brandbook](../brand/brandbook.md) and changes it.

## Goal

The home BBS of Velour is a live station and the site of the product.
It shows the product, its guides and its features. It follows the Velour brandbook.

## Scope

- A BBS directory in the repo. The operator runs it at a public address.
- Two languages: ru and en.
- Sections:
  - Guides: the setup guide and the BBT guide, shown on the file screen.
  - Features: what Velour can do now.
  - About the station: a text about the project and the session fields.
- Lamp stays the example BBS of the guides. The home BBS does not replace it.

## Decisions

- The content does not go ahead of the engine. Each screen and text uses only
  what the engine can do now.
- The brandbook is the target design. Mark in it what the engine cannot do yet:
  Messages, Files, Code, Blog, the arrow-key menu, the table of contents,
  the syntax colours and the mono mode.
- Fix in the brandbook each error that the work on the BBS finds.
- Version news comes after the first release tag. There is no news section now.

## Known brandbook fixes

- The brandbook says that the example BBS follows it. Lamp does not follow it.
  Change the text: the home station follows the brandbook.
- The copy rules of the brandbook are for Russian only. Add rules for English.

## Open questions

- Where the BBS directory is in the repo.
- How the guides get into the BBS with one copy only, with no copies of `docs/guide/`.
- Which brandbook screens the engine can build now: the header with the sections,
  the widths 80 and 132, the mono mode.
- The station facts: the SysOp name, the address, the port. The operator gives them.
  The brandbook does not let the station invent them.
- Who updates the features section when the engine gets a new feature.
