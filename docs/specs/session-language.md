# Session language

This specification defines the language of a session: the languages of the BBS,
the choice of the guest, the search of a screen and a content file by the language,
and the dictionaries of the engine texts.
Related parts: [screens-toml](screens-toml.md), [bbt-format](bbt-format.md), [menu-nav](menu-nav.md),
[file-view](file-view.md), [session-fields](session-fields.md), [session](session.md),
[screen-render](screen-render.md) and [home-bbs](home-bbs.md).

The [BBT guide](bbt-guide.md) teaches this specification. Keep the guide in sync:
[bbt-guide](bbt-guide.md), section 3.

## 1. Goal

A guest chooses a language. Each screen and each engine text then shows that language.
The session keeps the choice until its end.
The sysop writes a screen once for each language. A screen without a text,
for example a logo, is one file for all languages.

## 2. Concepts

- A "language code" matches `[a-z]{2,3}`. Examples: `en`, `ru`.
- The "languages of the BBS" are `[screen] lang` and the codes of the tables `[lang.<code>]`
  of `screens/screens.toml` ([screens-toml](screens-toml.md), sections 3.1 and 3.5).
  `[screen] lang` is `en` by default. So a BBS without `[lang]` has one language: `en`.
- The "session language" is a language of the BBS. A session always has one.
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
- In a BBS with `[lang]` tables, a directory in `screens/` whose name is not a language of the BBS
  gives a warning of `velourbbs check`: `screens/rus:1:1: warning: "rus" is not a language of [lang]`.

Example:

```
screens/
  screens.toml
  main.bbt          the choice of the language
  _logo.bbt         one logo for all languages
  goodbye.bbt
  en/menu.bbt       the menu in English
  en/texts.bbt
  ru/menu.bbt       the menu in Russian
  ru/texts.bbt
content/
  en/features.md
  ru/features.md
```

## 4. Search

The host finds a file by its name in the session language `<code>`:

1. the language file `screens/<code>/<file>` or `content/<code>/<path>`;
2. else the common file `screens/<file>` or `content/<path>`.

- The search takes the language path when a directory entry exists there, also a broken symlink.
  The rules of a file then apply to it ([bbt-format](bbt-format.md), section 2,
  [file-view](file-view.md), section 2): a broken symlink or a file that is not a regular file
  is an error, not a step to the common file.

- The search applies to a screen of `goto`, `gosub` and `[menu] start`, to `goodbye` and `shutdown`,
  to an include, to an art insert, to the file of a file screen and to the directory
  of the provider `files` ([list-view](list-view.md), section 4.1).
- An include, an art insert and a file of a screen use the session language,
  also in a common screen. Example: the common screen `main` with `{>_logo}` in the language `ru`
  shows `screens/ru/_logo.bbt` if it exists, else `screens/_logo.bbt`.
- The name of a `goto` or a `gosub` is a screen name. A code before it, as in `gosub ru/menu`,
  is not a path: it sets the language ([menu-nav](menu-nav.md), section 4).
- The "screen set" of a language is the result of the search for each screen name.
  The model holds one screen set for each language of the BBS.

## 5. Session

- The session starts with `[screen] lang` ([screens-toml](screens-toml.md), section 3.1).
- The actions `goto <code>/<name>`, `gosub <code>/<name>` and `lang <code>` set the session language
  ([menu-nav](menu-nav.md), section 4).
- The field `session.lang` gives the session language ([session-fields](session-fields.md), section 2).

Example: the station starts in `en`. The common screen `main` asks for the language.
The item `R` runs `gosub ru/menu`: the session language is `ru`, and the host shows `screens/ru/menu.bbt`.
`ESC` in the menu runs `back` and shows the common `main` again.

## 6. Checks

`velourbbs check` checks the screen set of each language of the BBS.

- Each screen set has the checks of [menu-nav](menu-nav.md), section 6, and of
  [file-view](file-view.md), section 7. The includes, the art inserts and the content files
  of the set use the search of its language.
- A `goto <code>/<name>` or a `gosub <code>/<name>` has the checks of a `goto` or a `gosub`
  to `<name>` in the screen set of `<code>`.
  Its error names the target as `<code>/<name>`: `item "R": do: unknown screen "ru/menu"`.
- These are errors too:
  - `lang <code>`, `goto <code>/<name>` or `gosub <code>/<name>` with a code that is not
    a language of the BBS: an invalid `do` ([bbt-format](bbt-format.md), section 8.4);
  - a screen name of a screen set that is not in the screen set of each other language,
    or that has other arguments there ([bbt-format](bbt-format.md), section 8.5).
    The action `lang` shows the frames of the stack in the new screen set, so they must fit.
    For each screen name, the "reference set" is the first set that holds the name,
    in the order of the codes. The check compares each other set with the reference set.
    Each difference gives one error at 1:1 of the screen file in the reference set,
    with the suffix of the other set. Examples with the languages `en` and `ru`:
    `screens/ru/news.bbt:1:1: no screen "news" (language "en")` when only `ru` has `news`,
    `screens/en/read.bbt:1:1: arguments differ (language "ru")`.
- The values of an argument of a file screen ([file-view](file-view.md), section 3.1) are the values
  of its `goto` and `gosub` in all screen sets. Each screen set loads and checks each value
  with its own search. An error is an error of the item, with the suffix of the set:
  `screens/ru/texts.bbt:12:17: item "3": args: file: content file "x.md" does not exist (language "en")`.
- The check of a screen that a user cannot reach does not run for each screen set separately.
  It walks the screen sets of all languages together ([menu-nav](menu-nav.md), section 6.1).
- The check reports an equal diagnostic once.
- In a BBS with more than one language, an error of a common file that the check finds
  only in some screen sets ends with ` (language "<code>")`, one diagnostic for each screen set.
  Example: `screens/about.bbt:6:6: item "M": do: unknown screen "menu" (language "de")`.
- The warning `no screen uses this file` ([screen-render](screen-render.md), section 3.1)
  applies to a fragment, an art file and a common screen that no screen set uses.
  Example: `screens/menu.bbt` next to `screens/en/menu.bbt` and `screens/ru/menu.bbt`
  in a BBS with the languages `en` and `ru`.

## 7. Engine texts

An "engine text" is a text that the host shows and the sysop does not write.

| Engine text | Where |
|---|---|
| `-- More --` | The More row without `more` ([screens-toml](screens-toml.md), section 3.1). |

- The English text is the key. English needs no dictionary.
- A dictionary is a GNU gettext `.po` file in the binary: `internal/i18n/po/<code>.po`.
  The host uses the dictionary of the session language. With no dictionary
  or with no translation of a text, the host shows the English text.
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
  The default is `[screen] lang`. A value that is not a language of the BBS is an invalid argument.
- `velourbbs preview <name>` shows the screen `<name>` of the screen set of the session language.
- The model of an argument file ([screen-render](screen-render.md), section 3.2) loads and checks
  the value in the screen set of each language, as section 6 does for the values of an item.
  The errors merge as in section 6. After an error, the command exits as after an error
  of a content file of a value ([screen-render](screen-render.md), section 3.2).

## 9. Code

| Package | Content |
|---|---|
| `internal/config` | `[screen] lang` and `[lang.<code>]` with their checks, and the More text of section 7.2. |
| `internal/bbt` | The search of section 4, the load of the screens of one screen set, and the parse of `lang <code>`, `goto <code>/<name>` and `gosub <code>/<name>`. |
| `internal/bbs` | The screen sets in the model and the checks of section 6. |
| `internal/i18n` | The dictionaries in the binary, their parser and the engine texts. |
| `internal/menu` | The session language and its actions ([menu-nav](menu-nav.md), section 4). |
| `internal/screen` | The field `session.lang`. |
| `cmd/velourbbs` | `--set session.lang`. |

## 10. Acceptance

- A test covers the search: a language file, a common file,
  and an include and an art insert of a common screen in a language.
- A test covers the checks of section 6, each error and the language suffix.
- A test covers the dictionary format: each rule of section 7.1, `fuzzy`, the header,
  a continued string and each error.
- A golden output of `velourbbs preview --keys` shows `gosub <code>/<name>`, and the action `lang`
  in a screen and in a file screen.
- Golden cell dumps show the More row of section 7.2: `[lang.<code>] more`, `[screen] more`,
  and the engine text in Russian and in English.
- `velourbbs check` in a BBS directory without `[lang]`, without `screens/en/`
  and without `content/en/` gives the same output as before.
- `go vet ./...`, `go test ./...` and `golangci-lint run ./...` pass.

## 11. Not in this specification

- Dictionaries for the text of screens.
- The language of a user account.
- A start language from the client: the variable `LANG` of the telnet option NEW-ENVIRON or of ssh.
- A region in a language code, for example `pt-br`.
- Plural forms and the names of months and days.
- A `.pot` template of the engine texts.
