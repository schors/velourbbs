# Session language

This specification defines the language of a session: the languages of the BBS,
the choice of the guest, the search of a screen and a content file by the language,
and the dictionaries of the engine texts.
Related parts: [screens-toml](screens-toml.md), [bbt-format](bbt-format.md), [menu-nav](menu-nav.md),
[file-view](file-view.md), [session-fields](session-fields.md), [session](session.md),
[screen-render](screen-render.md) and [home-bbs](home-bbs.md).

The [BBT guide](bbt-guide.md) teaches this specification. Keep the guide in sync:
[bbt-guide](bbt-guide.md), section 4.

## 1. Goal

A guest chooses a language. Each screen and each engine text then shows that language.
The session keeps the choice until its end.
The sysop writes a screen once for each language. A screen without a text,
for example a logo, is one file for all languages.

## 2. Concepts

- A "language code" matches `[a-z]{2,3}(-[a-z0-9]{2,8})?`. Examples: `en`, `ru`, `pt-br`.
- The "languages of the BBS" are the tables `[lang.<code>]` of `screens/screens.toml`
  ([screens-toml](screens-toml.md), section 3.5). A BBS without these tables has no languages.
- The "session language" is a language of the BBS, or "no language".
- A "language file" is a file in `screens/<code>/` or `content/<code>/`,
  where `<code>` is a language of the BBS.
  A "common file" is another file in `screens/` or `content/`.

## 3. Files

| Path | Content |
|---|---|
| `screens/<code>/<name>.bbt` | A screen of the language `<code>`. |
| `screens/<code>/_<name>.bbt` | A fragment of the language `<code>`. |
| `screens/<code>/<name>.ans` | Art of the language `<code>`. |
| `content/<code>/<path>` | A content file of the language `<code>`. |

- Each rule of a common file applies to a language file:
  [bbt-format](bbt-format.md), section 2, and [file-view](file-view.md), section 2.
- The load ignores a subdirectory and a `screens.toml` in `screens/<code>/`,
  as it ignores each other file that section 2 of [bbt-format](bbt-format.md) does not name.
- In a BBS with languages, a directory in `screens/` whose name is not a language of the BBS
  gives a warning of `velourbbs check`: `screens/rus:1:1: warning: "rus" is not a language of [lang]`.

Example:

```
screens/
  screens.toml
  main.bbt          the choice of the language
  _logo.bbt         one logo for all languages
  goodbye.bbt
  en/main.bbt       the menu in English
  en/texts.bbt
  ru/main.bbt       the menu in Russian
  ru/texts.bbt
content/
  en/features.md
  ru/features.md
```

## 4. Search

The host finds a file by its name and the session language:

1. In a language `<code>`: the language file `screens/<code>/<file>` or `content/<code>/<path>`.
2. Else the common file `screens/<file>` or `content/<path>`.

With no language, the host uses only step 2.

- The search applies to a screen of `goto`, `gosub` and `[menu] start`, to `goodbye` and `shutdown`,
  to an include, to an art insert and to the file of a file screen.
- An include, an art insert and a file of a screen use the session language,
  also in a common screen. Example: the common screen `main` with `{>_logo}` in the language `ru`
  shows `screens/ru/_logo.bbt` if it exists, else `screens/_logo.bbt`.
- The "screen set" of a session language is the result of the search for each screen name.
  The model holds one screen set for each language that a session can have (section 5).

## 5. Session

- The session starts with `[screen] lang` ([screens-toml](screens-toml.md), section 3.1).
  Without it, the session starts with no language.
- The action `lang <code>` sets the session language to `<code>`
  ([menu-nav](menu-nav.md), section 4). There is no action to go back to no language.
- The field `session.lang` gives the session language: the code, or empty for no language
  ([session-fields](session-fields.md), section 2).
- A session can have each language of the BBS. It can have no language only when
  `[screen] lang` is missing.

Example: the station starts with no language. The common screen `main` asks for the language.
The item `R` runs `lang ru`. The host shows `main` again: now `screens/ru/main.bbt`, the menu in Russian.

## 6. Checks

`velourbbs check` checks the screen set of each language that a session can have (section 5).

- Each screen set has the checks of [menu-nav](menu-nav.md), section 6, and of
  [file-view](file-view.md), section 7. The includes, the art inserts and the content files
  of the set use the search of its language.
- These are errors too:
  - `lang <code>` with a code that is not a language of the BBS: an invalid `do`
    ([bbt-format](bbt-format.md), section 8.4);
  - `[screen] lang` that is not a language of the BBS ([screens-toml](screens-toml.md), section 3.1);
  - a screen name of a screen set that is not in the screen set of each language of the BBS,
    or that has other arguments there ([bbt-format](bbt-format.md), section 8.5).
    The action `lang` shows the frames of the stack in the new screen set, so they must fit.
    The order of the sets is no language, then the codes in order. The error has the position 1:1
    of the screen file in the first set, with the suffix of the other set:
    `screens/ru/news.bbt:1:1: no screen "news" (language "en")`,
    `screens/en/read.bbt:1:1: arguments differ (language "ru")`.
- The values of an argument of a file screen ([file-view](file-view.md), section 3.1) are the values
  of its `goto` and `gosub` in all screen sets. Each screen set loads and checks each value
  with its own search. An error is an error of the item, with the suffix of the set:
  `screens/ru/texts.bbt:2:1: item "3": args: file: content file "x.md" does not exist (language "en")`.
- The check reports an equal diagnostic once.
- An error of a common file that the check finds only in some screen sets ends with
  ` (language "<code>")` or ` (no language)`, one diagnostic for each screen set.
  Example: `screens/about.bbt:2:1: item "M": do: unknown screen "menu" (no language)`.
- The warning "no screen uses this file" ([screen-render](screen-render.md), section 3.1)
  applies to a fragment or an art file that no screen set uses.
  A common screen that no screen set holds gives the warning `no screen set uses this file`.
  Example: `screens/main.bbt` next to `screens/en/main.bbt` and `screens/ru/main.bbt`
  with `[screen] lang`.

## 7. Engine texts

An "engine text" is a text that the host shows and the sysop does not write.

| Engine text | Where |
|---|---|
| `-- More --` | The More row without `more` ([screens-toml](screens-toml.md), section 3.1). |

- The English text is the key. English needs no dictionary.
- A dictionary is a GNU gettext `.po` file in the binary: `internal/i18n/po/<code>.po`.
  The host uses the dictionary of the session language. With no dictionary,
  with no language, or with no translation of a text, the host shows the English text.
- A new engine text adds a row to the table of this section.

### 7.1. Dictionary format

A dictionary is a subset of the PO format of GNU gettext.
The target implementations are GNU gettext `msgfmt` and `msgmerge`, and Poedit:
the subset accepts the files that they write for the engine texts.

- The file is UTF-8.
- A line that starts with `#` is a comment. This includes the comments `#,`, `#:`, `#|` and `#~`.
- An entry is `msgid "<text>"`, then `msgstr "<text>"`. The next `msgid` starts the next entry.
  Empty lines and comments can be between entries.
- A line with only a string `"<text>"` continues the string of the line before it.
- A string has the escapes `\\`, `\"`, `\n` and `\t`.
- The entry with an empty `msgid` is the header. The host ignores it.
- The host ignores an entry with the flag `fuzzy` in its `#,` comment, as `msgfmt` does.
  Example: `#, fuzzy, c-format`. An empty `msgstr` is no translation.
- These are errors: another keyword (`msgctxt`, `msgid_plural`, `msgstr[<n>]`), each other line,
  a string without its end quote or with text after it, a continued string with no string before it,
  a `msgid` without `msgstr`, a `msgstr` without `msgid`, another escape,
  invalid UTF-8, a duplicate `msgid`, a `msgid` that is not an engine text,
  and a translation that breaks a rule of the engine text.
  The rules of a More row are the rules of `more` ([screens-toml](screens-toml.md), section 3.1).
- A test reads each dictionary of the binary. An error fails the test.

The parser uses the standard library of Go. The subset needs no library:
the engine has few texts and no plural forms.

Example `ru.po`:

```
msgid ""
msgstr "Content-Type: text/plain; charset=UTF-8\n"

msgid "-- More --"
msgstr "-- Дальше --"
```

### 7.2. More row

The text of the More row is the first of:

1. `more` of `[lang.<code>]` of the session language ([screens-toml](screens-toml.md), section 3.5);
2. `[screen] more` ([screens-toml](screens-toml.md), section 3.1);
3. the engine text `-- More --` in the session language.

## 8. `velourbbs preview`

- `--set session.lang=<code>` sets the session language ([screen-render](screen-render.md), section 3.2).
  An empty value gives no language. The default is `[screen] lang`, else no language.
  A value that a session cannot have (section 5) is an invalid argument.
- `velourbbs preview <name>` shows the screen `<name>` of the screen set of the session language.
- In `--keys` ([menu-nav](menu-nav.md), section 7), the action `lang` changes the session language.

## 9. Code

| Package | Content |
|---|---|
| `internal/config` | `[screen] lang` and `[lang.<code>]` with their checks. |
| `internal/bbt` | The search of section 4 and the load of the screens of one screen set. |
| `internal/bbs` | The screen sets in the model and the checks of section 6. |
| `internal/i18n` | The dictionaries in the binary, their parser and the engine texts. |
| `internal/menu` | The session language, the action `lang` and the More text of section 7.2. |
| `internal/screen` | The field `session.lang`. |
| `cmd/velourbbs` | `--set session.lang`. |

## 10. Acceptance

- A test covers the search: a language file, a common file, no language,
  and an include and an art insert of a common screen in a language.
- A test covers the checks of section 6, each error and the language suffix.
- A test covers the dictionary format: each rule of section 7.1, `fuzzy`, the header,
  a continued string and each error.
- A golden output of `velourbbs preview --keys` shows the action `lang` in a screen and in a file screen.
- Golden cell dumps show the More row of section 7.2: `[lang.<code>] more`, `[screen] more`,
  and the engine text in Russian and in English.
- `velourbbs check` in a BBS directory without `[lang]` gives the same output as before.
- `go vet ./...`, `go test ./...` and `golangci-lint run ./...` pass.

## 11. Not in this specification

- Dictionaries for the text of screens.
- The language of a user account.
- A start language from the client: the variable `LANG` of the telnet option NEW-ENVIRON or of ssh.
- A fallback from a region code to its language, for example from `pt-br` to `pt`.
- Plural forms and the names of months and days.
- A `.pot` template of the engine texts.
