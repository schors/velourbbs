# 2. Каталог BBS

«Каталог BBS» содержит файлы одной BBS. В этом руководстве это `/srv/velourbbs`.
Скопируй в него пример BBS «Лампа». В репозитории исходного кода выполни:

```
sudo cp -r examples/lamp/ru /srv/velourbbs
sudo chmod -R a+rX /srv/velourbbs
```

До копирования `/srv/velourbbs` не должен существовать.
`chmod` делает для каждого пользователя файлы доступными для чтения,
а каталоги — для чтения и поиска. BBS работает от пользователя `velourbbs` (ниже).
Этот пользователь должен читать файлы.
Файлы принадлежат root. Редактируй их от root, например с помощью `sudoedit`.

Файлы «Лампы»:

| Файл | Содержимое |
|---|---|
| `bbs.toml` | Настройки хоста (раздел 3). В «Лампе» файл пустой: значения по умолчанию. |
| `screens/screens.toml` | Настройки экранов: ширина, стили и другие значения. |
| `screens/_head.bbt` | Шапка. Каждый экран включает её. |
| `screens/main.bbt` | Первый экран: меню. |
| `screens/about.bbt` | Текст о BBS. |
| `screens/term.bbt` | Значения терминала. |
| `screens/texts.bbt` | Список текстов читальни. |
| `screens/reader.bbt` | Экран, который показывает текст из списка. |
| `screens/goodbye.bbt` | Экран после конца сеанса. |
| `content/lamp.md` | Текст в Markdown: короткая история ламп. |
| `content/rules.txt` | Простой текст: правила станции. |

Как менять экраны, читай в [руководстве по BBT](../bbt/README.md).

## Пользователь BBS

BBS работает от системного пользователя `velourbbs`, без привилегий.
У пользователя нет входа в систему и нет домашнего каталога.
BBS пишет свой файл данных в каталог `/var/lib/velourbbs` (раздел 3).
Создай пользователя и каталог:

```
sudo useradd --system --no-create-home --shell /usr/sbin/nologin velourbbs
sudo install -d -o velourbbs -g velourbbs /var/lib/velourbbs
```

Запускай каждую команду `velourbbs` в каталоге BBS.
Запускай `velourbbs check` и `velourbbs serve` от пользователя `velourbbs`:
`sudo -u velourbbs /usr/local/bin/velourbbs check`. Root может прочитать файл,
который пользователь `velourbbs` прочитать не может.
Указывай полный путь: `sudo` на некоторых хостах, например в RHEL 9, не ищет в `/usr/local/bin`.
Полные правила: [bbs-toml](../../../specs/bbs-toml.md), раздел 2.

## Результат

```
cd /srv/velourbbs
sudo -u velourbbs /usr/local/bin/velourbbs check
```

Команда ничего не выводит.
