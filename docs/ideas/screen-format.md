# Idea: screen format and template parser

MVP part 1 of 3. The other parts are [menu-nav](menu-nav.md) and [telnet-term](telnet-term.md).

## Goal

The host loads screen files into an in-memory screen model and renders it for one terminal.
The terminal can have any size. 80x25 is not a limit.

## Decisions

| Topic | Decision |
|---|---|
| Compile step | None. The host parses the source files at load. |
| Base file | `screens/NAME.bbt`, a UTF-8 text template ("BBS template"). |
| Screen data | TOML front matter at the start of the `.bbt` file: menu items, `width`, `narrow`, `[fields]`. One screen is one file. |
| Screen and menu | The screen only draws. A menu item is a choice with no position. The author writes the item text, for example `[M]essages`, into the screen. |
| Art | `.ans` files are art only. A `.bbt` file inserts them with `{>NAME.ans}`. |
| Art charset | UTF-8. If the file is not valid UTF-8, the loader reads it as CP437. This is input only. |
| Model | A cell holds one grapheme and attributes, never encoded bytes. A wide grapheme takes 2 cells. Colours are full colours: 16, 256 or RGB. |
| Colour syntax | `{!..}` takes a theme style, a 16-colour number, a 256-colour number or an RGB value. The design stage defines the syntax. |
| Output charset | UTF-8 now. An encoder at the session output (idea 3) adds 8-bit charsets later. |

Rejected options:
- **Compile step (`bbs compile`) or a disk cache.** It adds a stale-file risk and a format version.
  An 80x25 file parses in microseconds, so there is no speed gain.
- **Art and menu in one file.** Art editors (Moebius, PabloDraw) damage extra data on save.
- **`.txt` + `.toml` pair.** One file per screen is simpler for the sysop.
- **Fluid layout, as on the web.** A screen has width limits and is centred.

## Example

```
+++
width = [64, 100]
narrow = "main_small"
[fields]
w = "term.cols"
[[item]]
key = "A"
do = "goto about"
+++
{#if term.cols >= 100}
{>main_wide.ans}
{#else}
{>main.ans}
{/if}
{!title}Welcome! Your terminal has {w} columns.
{!prompt}[A]bout [G]oodbye
```

## Layout rules

- **Width.** `width = [min, max]` in the front matter. The render width is `term.cols`, clamped to
  `[min, max]`. When the terminal is wider than `max`, the screen is centred horizontally.
- **Default width.** Without `width`, the limits come from the widest art of the screen.
  A screen without art uses the default from `bbs.toml`.
- **Narrow terminal.** When `term.cols < min`, the host shows the screen named in `narrow`.
  Without `narrow`, the host cuts the screen on the right. The `narrow` screen is another view of
  the same screen: it uses the items of the main screen and has no `narrow` of its own.
- **Height.** There are no height limits. A screen that is higher than the terminal stops after
  `term.rows - 1` rows with a "More" prompt. Any key shows the next page.
- **Text.** A `.bbt` line that is wider than the render width wraps at word boundaries.
  The design stage defines the wrap details.
- **Art.** An art insert takes its own rows and starts at the left edge of the screen area.
  Art that is wider than the render width is cut on the right. `bbs check` gives a warning.
  The art width comes from SAUCE, else 80.
- **Block tags.** A line with only `{#if}`, `{#else}` or `{/if}` adds no row.
- **Tokens in art.** Fields only. A field keeps the width of its drawn span.
  `[fields]` aliases make short fields possible.
- **Resize.** After a NAWS resize, the host uses the new size at the next screen.
- **Source positions.** Layout keeps the source position of each token. `bbs check` errors use
  them. A later mark feature also uses them.
- **Reserved syntax.** `{@id}...{/@}` is reserved for marks. A mark is a named span. Layout
  will report its final cells, so arrow-key selection can find items. `bbs check` rejects
  marks in the MVP.

## MVP scope

In:
- `.bbt` tokens: fields `{path}`, `{#if}`, colour `{!..}`, include `{>NAME}`, art insert
  `{>NAME.ans}`;
- `.ans` loader: SAUCE, CP437 input, cell grid, fields in art;
- front matter: items, `width`, `narrow`, `[fields]`;
- `bbs check`: all load checks, errors with file, row and column;
- `bbs preview`: render a screen at a given terminal size to a cell dump. The cell dump is the
  golden file for tests.

Out (backlog):
- `#each` and list rows in art. They need data providers, which the MVP does not have.
- `#if` inside art. Logic is in the `.bbt` file only.
- Raw art (`raw = true`, ANSI animation). It cannot be cut or centred.
- Reload. A restart is sufficient for the MVP.
- Marks and arrow-key selection. The MVP selects items by hotkey only.

## Dependencies

- `term.cols`, `term.rows` and the charset come from [telnet-term](telnet-term.md).
- Menu items and input modes are defined in [menu-nav](menu-nav.md).

## Open questions

1. Which TOML parser gives positions for errors such as `main.bbt:9: item "S"`?
2. Where do the sample values for `bbs preview` come from?
3. `docs/screens.md` has the old design. The design stage updates it.
