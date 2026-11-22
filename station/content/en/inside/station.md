# Station

This station is the home BBS of Velour. It runs on the same program that you can
install yourself, and it uses only what Velour can do now.
So the station is the site of the product and its live example.

## Files

The whole station is the directory `station/` in the Velour repository:
[github.com/schors/velourbbs](https://github.com/schors/velourbbs). It has no code,
only the files of the sysop:

- `bbs.toml` - the name of the BBS and the path of the data file;
- `screens/` - the screens in the BBT format and `screens.toml` with the theme;
- `content/` - the texts in Markdown, as this one.

## Screens

A screen is a BBT text file. The file starts with the settings and the menu items.
Then comes the text with styles, fields and conditions.
One file gives the screen for 80 and for 132 columns: stretch tokens draw the rules,
the dot leaders and the edges.

The brandbook gives the look of the station: gold lettering, bronze hairlines,
empty cells and no frames. The theme turns the roles of the brandbook into colours
for truecolor, 256 and 16 colours.

## Two languages

One BBS directory holds both languages. The entry screen asks for the language
and opens the menu in it. The screens of the two languages have the same menu items.
A test checks this.

## Texts

The guides are in `docs/guide/` of the repository. The station reads them through
a symlink and does not copy them. So the guide on the station and on GitHub is one file.

## Checks

A test runs `velourbbs check` on the station. Then it takes each screen at 80x24
and 132x43 and compares the cells with a golden dump. The golden dumps are in git,
so each change of a screen shows in the pull request.
