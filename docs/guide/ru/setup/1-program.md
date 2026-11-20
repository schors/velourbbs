# 1. Программа

Нужен Go 1.27.3 или новее. В репозитории исходного кода Velour BBS выполни:

```
go build ./cmd/velourbbs
sudo install velourbbs /usr/local/bin/
```

Первая команда создаёт файл `velourbbs`. Вторая команда копирует его
в `/usr/local/bin/velourbbs`.

Собирай в git-checkout тега релиза без изменений, например после `git checkout v1.2.0`.
Тогда поле `bbs.version` показывает тег. Любая другая сборка даёт `dev`.
Полные правила: [session-fields](../../../specs/session-fields.md), раздел 3.4.

## Результат

`velourbbs` без аргументов печатает usage:

```
usage:
  velourbbs art import <input> <output>
  velourbbs check [--size <cols>x<rows>]...
  velourbbs fields [--sample]
  velourbbs preview <name> [--size <cols>x<rows>] [--sample <set>] [--set <path>=<value>]... [--keys <keys>] [--view]
  velourbbs serve
```
