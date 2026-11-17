# Задача 9. Ошибки

`velourbbs check` находит ошибки до запуска BBS.
Каждая строка вывода содержит файл, номер строки, номер столбца и сообщение (задача 1).
В арте это номер строки и номер столбца арта.
Эта задача показывает частые ошибки и их исправления.

## `unknown style`

```
screens/about.bbt:10:8: unknown style "titel"
```

Токен стиля `{!titel}` называет стиль, которого нет в теме.
Исправь имя или добавь стиль в `[theme.default]`.

## `token has no "}"`

```
screens/about.bbt:10:35: token has no "}"
```

`{` начинает токен (задача 5). Закрой токен знаком `}` или напиши `{{` для простой `{`.

## `unknown field`

```
screens/about.bbt:11:13: unknown field "term.colours"
screens/about.bbt:14:10: unknown field "note"
screens/term.bbt:14:67: unknown field "item.S"
```

- Это имя — не поле и не алиас (задача 5). Исправь имя.
- Для текста в фигурных скобках напиши `{{note}`. В арте `{{` не помогает:
  измени текст в редакторе арта и снова переведи арт.
- Для `item.S` нужен пункт с клавишей `S` в том же экране.

## `{#if} has no {/if}`

```
screens/about.bbt:12:8: {#if} has no {/if}
```

Добавь `{/if}`. Блок кончается в том же файле, где начался.

## `control character U+0009`

```
screens/about.bbt:13:1: control character U+0009
```

В строке есть TAB. Замени его пробелами. Настрой редактор, чтобы клавиша Tab писала пробелы.

## Условие

```
screens/term.bbt:14:8: expected an integer or a string, found "xterm"
screens/term.bbt:14:38: integer field "term.colors" needs an operator
```

- `{#if term.type == xterm}`: пиши текст в двойных кавычках: `term.type == "xterm"`.
- `{#if term.colors = 256}`: сравнивай через `==`.

## `item: must be an array of tables`

```
screens/main.bbt:2:1: item: must be an array of tables
screens/main.bbt:5:2: table item already exists
screens/main.bbt:5:2: table item already exists as an array of tables
screens/main.bbt:5:3: key item already exists as a table, but should be an array table
```

Во front matter стоит `[item]`. Напиши `[[item]]`.
Первое сообщение выводит экран с одним пунктом. Остальные сообщения дают строку таблицы.

## `unknown screen`

```
screens/main.bbt:2:1: item "T": do: unknown screen "term"
```

Для `goto term` нужен файл `screens/term.bbt`. Создай файл или исправь имя.

## `screen has no items`

```
screens/x.bbt:1:1: screen has no items
```

Каждому экрану, кроме `goodbye` и `shutdown`, нужен хотя бы один пункт. Добавь пункт.

## `content file ... does not exist`

```
screens/code.bbt:2:1: file: content file "src/lamp.c" does not exist
screens/texts.bbt:2:1: item "1": args: file: content file "lmap.md" does not exist
```

- Экран `code` содержит сниппет из задачи 7 («Исходный код»).
  Его ключ `file` называет `src/lamp.c`, а файла `content/src/lamp.c` нет.
  Положи файл по этому пути в `content/` или исправь путь.
- Пункт `1` передаёт `goto reader` файл `lmap.md` в `args`, а этого файла нет в `content/`.
  Исправь имя в `args`.

## `front matter: no end line`

```
content/lamp.md:1:1: front matter: no end line "+++"
```

Первая строка файла — `+++`, а другой строки `+++` нет.
Добавь строку `+++` после последнего ключа front matter (задача 7).

## `not in the flat form`

```
screens/logo.ans:1:1: not in the flat form, convert it with velourbbs art import
```

Арт — это файл редактора арта. Переведи его (задача 8).

## Поле в арте меняет цвет

```
screens/size.ans:1:7: token "{w}" changes the foreground: 11 at col 7, 14 at col 9
```

У символов токена два цвета: 11 в столбце 7 и 14 в столбце 9.
В редакторе арта нарисуй весь токен одним цветом. Потом снова переведи арт.

## Результат

1. В строке 10 файла `screens/about.bbt` «Лампы» поставь `{!titel}` перед `«Лампа»`.
2. Выполни `velourbbs check`. Команда выводит:

```
screens/about.bbt:10:8: unknown style "titel"
```

3. Удали `{!titel}`. `velourbbs check` ничего не выводит.

Полные правила: [screen-render](../../../specs/screen-render.md), раздел 3.1,
[menu-nav](../../../specs/menu-nav.md), раздел 6,
и [file-view](../../../specs/file-view.md), раздел 7.
