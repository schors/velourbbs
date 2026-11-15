# Idea: screen-fill

Backlog. It defines the cells of the terminal outside the rendered rows of a screen.
It changes [bbt-format](../specs/bbt-format.md), sections 6.2 and 6.4,
[screen-render](../specs/screen-render.md) and [telnet-term](../specs/telnet-term.md), section 7.1.

## Goal

A screen with a filled background fills the whole terminal at each size.
A template has no empty lines only to fill the rows below it.

## Decisions

- The cells outside a screen are this idea, not a part of the [example style](../specs/example-style.md).

## Facts

- The host starts each page with `ESC[0m ESC[2J`. The terminal clears the screen
  in its default colours.
- The margins left and right of the screen area are spaces in `normal`
  ([bbt-format](../specs/bbt-format.md), section 6.2).
- The rows below the last rendered row keep the default colours of the terminal.
- A page with a More row has `term.rows` rows: `term.rows - 1` rows of the screen and the More row.
  The host writes all W cells of the More row, also on the last row of the terminal.
- The example style has 23 rows. On a 24-row terminal, row 24 keeps the default colours.

## Open questions

1. Fill each page to `term.rows - 1` rows or to `term.rows` rows in `normal`.
   A write to the last cell of the last row can scroll some terminals.
2. A screen that does not want the fill, for example a screen with the colour `default`.
3. Vertical centring of a short screen on a tall terminal, for example 23 rows at 132x43.
4. The cells right of the screen area when `term.cols` < `min`: the host cuts the rows now.
5. The More row and the fill.
