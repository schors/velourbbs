# Idea: example style and the light BBS

Backlog. It needs [theme-depth](theme-depth.md).
[bbt-guide](bbt-guide.md) and [setup-guide](setup-guide.md) build this BBS.

## Goal

The guides build a light example BBS. The reader sees the design and the controls at once.
The screens are pleasant to read.

## Decisions

- The example style is not a second brandbook. It is the light BBS that the guides build.
  It does not use the [brandbook](../brand/brandbook.md).
- It uses only what the engine can do, and [theme-depth](theme-depth.md).
- Sizes: 80x24 and 80x25, with 23 rows on a screen, and 132x43.
- Colours: truecolor, 256, and 16 as grey: black, dark grey, light grey and white.
- One template for the three colour modes.
- Frames and a filled background. No shadows.
- The keys are always visible on the screen.
- The screens are not overloaded.

## Open questions

1. The design: the frames, the palette and the layout.
2. The screens: the minimum that shows the design and the controls.
3. The content of the BBS.
4. Bright backgrounds 8–15 in 16 colours, for example a dark grey background 8.
5. `velourbbs preview` renders the ANSI files of the examples, not a human.
6. The place of the files.
