# Задача 10. Ошибки

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

Для `gosub term` нужен файл `screens/term.bbt`. Создай файл или исправь имя.

## `screen has no items`

```
screens/x.bbt:1:1: screen has no items
```

Каждому экрану, кроме `goodbye` и `shutdown`, нужен хотя бы один пункт. Добавь пункт.

## `a user cannot reach this screen`

```
screens/old.bbt:1:1: warning: a user cannot reach this screen
```

`velourbbs check` проходит по каждому `goto`, `gosub`, `select` и `lang` от первого экрана `main`.
Ни один путь не ведёт на экран `old`.
Добавь пункт, который открывает экран, например `do = "gosub old"` в `main`. Или удали файл.

- Экраны, которые открывают только друг друга, тоже получают предупреждение.
- Пункт с `when` — это путь.
- `goodbye` и `shutdown` не получают предупреждения.
- В BBS на двух языках (задача 7) проверка идёт по каждому языку.
  В `screens/en/menu.bbt` есть пункт `gosub news`, а в `screens/ru/menu.bbt` его нет.
  Тогда общий файл `screens/news.bbt` получает предупреждение для `ru`:
  `screens/news.bbt:1:1: warning: a user cannot reach this screen (language "ru")`.
- Предупреждение не останавливает BBS.

## `content file ... does not exist`

```
screens/code.bbt:2:1: file: content file "src/lamp.c" does not exist
screens/texts.bbt:2:1: item "1": args: file: content file "lmap.md" does not exist
```

- Экран `code` содержит сниппет из задачи 7 («Исходный код»).
  Его ключ `file` называет `src/lamp.c`, а файла `content/src/lamp.c` нет.
  Положи файл по этому пути в `content/` или исправь путь.
- Пункт `1` передаёт `gosub reader` файл `lmap.md` в `args`, а этого файла нет в `content/`.
  Исправь имя в `args`.

## `directory ... does not exist`

```
screens/log.bbt:2:1: data: args: dir: directory "lgo" does not exist
```

Ключ `data` экрана-списка `log` называет каталог `lgo` (задача 8),
а каталога `content/lgo/` нет. Исправь имя или создай каталог.

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

Арт — это файл редактора арта. Переведи его (задача 9).

## Поле в арте меняет цвет

```
screens/size.ans:1:7: token "{w}" changes the foreground: 11 at col 7, 14 at col 9
```

У символов токена два цвета: 11 в столбце 7 и 14 в столбце 9.
В редакторе арта нарисуй весь токен одним цветом. Потом снова переведи арт.

## Раскладка при размере терминала

`velourbbs check --size <cols>x<rows>` отрисовывает каждый экран при этом размере и предупреждает об ошибках раскладки.
Укажи `--size` для каждого размера терминала, который ты ждёшь:

```
velourbbs check --size 80x24 --size 132x43
```

При этих размерах «Лампа» не даёт предупреждений. У её экранов `width = [80, 80]`,
поэтому терминал в 40 столбцов обрезает их строки:

```
screens/about.bbt:10:1: warning: at 40x24: the row is 46 cells wide, the screen shows 40
screens/_head.bbt:2:1: warning: screen "main": at 40x24: the row is 78 cells wide, the screen shows 40
```

Предупреждения и их исправления:

| Предупреждение | Проблема | Исправление |
|---|---|---|
| `the line wraps to 2 rows` | Строка текста шире экрана. | Сократи строку. |
| `31 rows, the screen gets a More row` | На экране больше строк, чем в терминале. | Убери строки или раздели экран. |
| `25 rows, the rows after 24 are cut` | Экран файла или экран списка не помещается в терминал. | Убери строки над или под областью файла или списком. |
| `the row is 43 cells wide, the screen shows 40` | BBS обрезает строку. | Сократи строку или уменьши `min` в `width`. |
| `art "logo.ans" is 100 cells wide, the screen shows 80` | BBS обрезает арт. | Покажи арт меньше на маленьком терминале (задача 9). |

- Предупреждение во фрагменте начинается с экрана, который его включает: `screen "main": `.
- Предупреждение не останавливает BBS.

## Результат

1. В строке 10 файла `screens/about.bbt` «Лампы» поставь `{!titel}` перед `«Лампа»`.
2. Выполни `velourbbs check`. Команда выводит:

```
screens/about.bbt:10:8: unknown style "titel"
```

3. Удали `{!titel}`. `velourbbs check` ничего не выводит.

Полные правила: [screen-render](../../../specs/screen-render.md), разделы 3.1 и 3.6,
[menu-nav](../../../specs/menu-nav.md), раздел 6,
[file-view](../../../specs/file-view.md), раздел 7,
и [list-view](../../../specs/list-view.md), раздел 10.
