# Руководство по установке

Это руководство учит устанавливать, проверять и запускать Velour BBS на Linux-хосте.
Оно для сисопа.
Нужны Linux-хост с systemd, права root, текстовый редактор и telnet-клиент.

Каждый раздел заканчивается проверкой результата.
Раздел ссылается на спецификации, если в них есть его полные правила.
Экраны учит писать [руководство по BBT](bbt.md).

## 1. Программа

Нужен Go 1.27.3 или новее. В репозитории исходного кода Velour BBS выполни:

```
go build ./cmd/velourbbs
sudo install velourbbs /usr/local/bin/
```

Первая команда создаёт файл `velourbbs`. Вторая команда копирует его
в `/usr/local/bin/velourbbs`.

Собирай в git-checkout тега релиза без изменений, например после `git checkout v1.2.0`.
Тогда поле `bbs.version` показывает тег. Любая другая сборка даёт `dev`.
Полные правила: [session-fields](../../specs/session-fields.md), раздел 3.4.

### Результат

`velourbbs` без аргументов печатает usage:

```
usage:
  velourbbs art import <input> <output>
  velourbbs check
  velourbbs preview <name> [--size <cols>x<rows>] [--set <path>=<value>]... [--keys <keys>]
  velourbbs serve
```

## 2. Каталог BBS

«Каталог BBS» содержит файлы одной BBS. В этом руководстве это `/srv/velourbbs`.
Скопируй в него пример BBS «Лампа». В репозитории исходного кода выполни:

```
sudo cp -r examples/lamp/ru /srv/velourbbs
sudo chmod -R a+rX /srv/velourbbs
```

До копирования `/srv/velourbbs` не должен существовать.
`chmod` делает для каждого пользователя файлы доступными для чтения,
а каталоги — для чтения и поиска. BBS работает от пользователя без привилегий (раздел 5).
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
| `screens/goodbye.bbt` | Экран после конца сеанса. |

Как менять экраны, читай в [руководстве по BBT](bbt.md).

Запускай каждую команду `velourbbs` в каталоге BBS.
Запускай `velourbbs check` от своего пользователя, не от root: root может прочитать файл,
который пользователь BBS прочитать не может.
Полные правила: [bbs-toml](../../specs/bbs-toml.md), раздел 2.

### Результат

```
cd /srv/velourbbs
velourbbs check
```

Команда ничего не выводит.

## 3. `bbs.toml`

`bbs.toml` содержит таблицу `[bbs]`: имя BBS, и таблицу `[telnet]`: настройки telnet-сервера.
Каждый ключ необязателен. Ключ, которого нет в файле, получает значение по умолчанию.
Пустой файл даёт значения по умолчанию.

### Имя

`name` таблицы `[bbs]` — имя BBS. Экран показывает его в поле `bbs.name`.
По умолчанию `"Kind Brothers BBS"`.

### Адрес

`listen` — TCP-адрес сервера: `<host>:<port>`. По умолчанию `":2323"`.

- Пустой хост: сервер слушает все адреса хоста: `":2323"`.
- Один адрес: сервер слушает только этот адрес, например `"192.0.2.1:2323"`.
  Используй его, когда у хоста больше одного адреса и BBS должна использовать только один.

`velourbbs check` проверяет только форму `listen` и порт.
`velourbbs serve` проверяет адрес при запуске. Если у хоста нет этого адреса,
`velourbbs serve` останавливается.

### Лимиты подключений

| Ключ | По умолчанию | Что делает | Когда менять |
|---|---|---|---|
| `max_conns` | 64 | Наибольшее число подключений. | Одновременно подключается больше пользователей. Тогда измени и лимиты из раздела 5. |
| `max_conns_per_addr` | 4 | Наибольшее число подключений с одного адреса. | Много пользователей делят один адрес, например за NAT. |
| `max_connects_per_min` | 10 | Наибольшее число новых подключений с одного адреса за 60 с. | Клиент много раз подключается заново. |
| `allow` | `[]` | Адреса и префиксы без лимитов адреса. | Одному адресу нужно больше подключений, например известному NAT. |

- «Один адрес» — это один адрес IPv4 или один префикс IPv6 /64.
- `max_conns` действует и на адреса из `allow`.
- Подключение сверх лимита получает строку `Too many connections. Try again later.`
  Затем сервер закрывает его.

Пример:

```toml
[telnet]
allow = ["192.0.2.10", "2001:db8::/48"]
```

### Протокол PROXY

`proxy` содержит адреса и префиксы прокси, которые посылают заголовок протокола PROXY.
По умолчанию `[]`: сервер не читает заголовок.
Задай его, когда подключения к BBS передаёт прокси, например HAProxy.
Тогда лог и лимиты используют адрес клиента, а не адрес прокси.

- Сервер читает заголовок только с адресов из `proxy`.
- Подключение с адреса из `proxy` без правильного заголовка закрывается без вывода.
  Поэтому, если `proxy` содержит `127.0.0.1`, telnet-клиент на том же хосте
  не может подключиться к `127.0.0.1`.
- Прокси должен послать заголовок при подключении. AWS NLB по умолчанию ждёт данных
  от клиента. Тогда он не работает.

Пример:

```toml
[telnet]
proxy = ["127.0.0.1"]
```

### Тайм-ауты

| Ключ | По умолчанию | Что делает | Когда менять |
|---|---|---|---|
| `idle_timeout` | 600 | Секунды без нажатия клавиши, после которых сервер отключает клиента. | Пользователи читают длинные экраны. |
| `probe_timeout` | 5 | Секунды от подключения, за которые сервер узнаёт терминал. | У клиентов медленная связь. |
| `output_timeout` | 60 | Секунды, за которые должна закончиться запись. Иначе сервер отключает клиента. | У клиентов медленная связь. |

### Пример файла

Этот файл содержит значения по умолчанию. Он задаёт каждый ключ, кроме `allow` и `proxy`:

```toml
[bbs]
name = "Kind Brothers BBS"

[telnet]
listen = ":2323"
max_conns = 64
max_conns_per_addr = 4
max_connects_per_min = 10
idle_timeout = 600
probe_timeout = 5
output_timeout = 60
```

Запиши его в `/srv/velourbbs/bbs.toml`. Затем измени нужные значения.
После каждого изменения выполни `sudo chmod -R a+rX /srv/velourbbs` и `velourbbs check`.

Полные правила: [bbs-toml](../../specs/bbs-toml.md), разделы 3.1 и 3.2.

### Результат

`velourbbs check` ничего не выводит.

## 4. Команды

### `velourbbs check`

Запускай её после каждого изменения файла и перед каждым перезапуском BBS.
Она проверяет `bbs.toml` и экраны. Когда файлы правильные, она ничего не выводит.
Иначе она выводит по одной строке на каждую ошибку или предупреждение:

```
screens/main.bbt:2:1: item "S": when: unknown field "term.colours"
screens/about.bbt:4:1: warning: art "logo.ans" is 132 cells wide, max width is 100
```

Строка содержит файл, номер строки, столбец и сообщение.
Предупреждение не останавливает BBS. Ошибка останавливает, и команда завершается
с кодом 1.
Полные правила: [screen-render](../../specs/screen-render.md), раздел 3.1.

### `velourbbs preview`

Она показывает экран как текст, без telnet-клиента:

```
velourbbs preview main
```

`--size` задаёт размер терминала, например `--size 132x43`.
`--set term.colors=16` показывает экран для терминала с 16 цветами.
Полные правила: [screen-render](../../specs/screen-render.md), раздел 3.2.
Подробно: задача 1 [руководства по BBT](bbt.md).

### `velourbbs preview --keys`

Она нажимает клавиши и показывает каждый экран. Используй её, чтобы проверить клавиши экрана:

```
velourbbs preview main --keys A,ESC,G
```

Подробно: задача 4 [руководства по BBT](bbt.md).

### `velourbbs art import`

Она переводит файл редактора арта в форму, которую читает BBS:

```
sudo velourbbs art import art/logo.ans screens/logo.ans
```

Подробно: задача 7 [руководства по BBT](bbt.md).

### `velourbbs serve`

Она запускает telnet-сервер. Сервер слушает порт 2323.

- Сначала она делает проверки `velourbbs check`. После ошибки она завершается с кодом 1.
- Если она не может слушать адрес из `listen`, она завершается с кодом 1.
- Она пишет лог в стандартный поток ошибок: строку `listen` с адресом,
  строку со значениями терминала каждого клиента и строку `close`
  для каждого подключения.
- Она читает файлы только при запуске. После изменения файла перезапусти её.
- Ctrl+C или SIGTERM останавливает её. Она закрывает подключения и завершается с кодом 0.

Пример лога:

```
time=2026-11-16T09:42:48.905Z level=INFO msg=listen addr=[::]:2323
time=2026-11-16T09:42:49.929Z level=INFO msg=probe addr=127.0.0.1 ttype=xterm-256color naws=80x24 cpr=1;2 type=xterm-256color cols=80 rows=24 charset=utf-8 colors=256 time=355.507µs
time=2026-11-16T09:42:51.231Z level=INFO msg=close addr=127.0.0.1 reason=logoff time=1.30238633s
```

Полные правила: [telnet-term](../../specs/telnet-term.md), раздел 10.

### Результат

1. В `/srv/velourbbs` выполни `velourbbs serve`. Она пишет строку `listen`.
2. Подключи telnet-клиент к хосту, порт 2323, например `telnet localhost 2323`.
   Клиент показывает экран `main`: шапку «Лампа» и меню `A T G`.
   Нажми `G`, чтобы завершить сеанс.
3. Останови сервер клавишами Ctrl+C. Раздел 5 использует порт 2323.

## 5. systemd

Нужен systemd 240 или новее с cgroup v2.
В Debian 12, Ubuntu 24.04 и RHEL 9 они есть по умолчанию.

### Unit

`examples/deploy/velourbbs.service`:
```
[Unit]
Description=Velour BBS
Wants=network-online.target
After=network-online.target

[Service]
Type=exec
ExecStart=/usr/local/bin/velourbbs serve
WorkingDirectory=/srv/velourbbs
Restart=on-failure
DynamicUser=yes
MemoryMax=128M
TasksMax=64
LimitNOFILE=1024

[Install]
WantedBy=multi-user.target
```

| Строка | Значение |
|---|---|
| `Description=Velour BBS` | Имя в `systemctl` и в логе. |
| `Wants=network-online.target`, `After=network-online.target` | Адрес из `listen` существует до запуска. Иначе `velourbbs serve` не может слушать и останавливается. |
| `Type=exec` | `systemctl start` завершается ошибкой, когда systemd не может запустить программу. |
| `ExecStart=/usr/local/bin/velourbbs serve` | Команда. |
| `WorkingDirectory=/srv/velourbbs` | Каталог BBS. |
| `Restart=on-failure` | systemd снова запускает BBS после сбоя. |
| `DynamicUser=yes` | BBS работает от нового пользователя без привилегий. Система для него только для чтения, и `/home` только для чтения. BBS только читает каталог BBS, поэтому это ей не мешает. |
| `MemoryMax=128M` | Лимит памяти. Он рассчитан на значения `bbs.toml` по умолчанию. «Лампа» использовала 6,8 МБ при запуске, 11,9 МБ с 64 сеансами и 15,8 МБ с 300 сеансами. |
| `TasksMax=64` | Лимит потоков. |
| `LimitNOFILE=1024` | Лимит открытых файлов. |
| `WantedBy=multi-user.target` | `systemctl enable` запускает BBS при загрузке системы. |

### Установка и запуск

В репозитории исходного кода выполни:

```
sudo cp examples/deploy/velourbbs.service /etc/systemd/system/
sudo systemctl enable --now velourbbs
```

`enable --now` загружает unit, запускает BBS и запускает её при каждой загрузке системы.

BBS пишет лог в стандартный поток ошибок. systemd записывает его в журнал.
Прочитай лог:

```
sudo journalctl -u velourbbs
```

### Изменения

После изменения файла в каталоге BBS:

```
cd /srv/velourbbs
sudo chmod -R a+rX /srv/velourbbs
velourbbs check && sudo systemctl restart velourbbs
```

После изменения `/etc/systemd/system/velourbbs.service`:

```
sudo systemctl daemon-reload
sudo systemctl restart velourbbs
```

### Лимиты

- `TasksMax` не меньше числа CPU + 32. `nproc` выводит число CPU.
  Go запускает потоки для каждого CPU. Когда Go не может запустить поток, BBS
  останавливается со всеми сеансами. `TasksMax=64` достаточно для 32 CPU или меньше.
- `LimitNOFILE` не меньше `2 × max_conns + 16`. У BBS не больше `max_conns`
  подключений и не больше `max_conns` подключений, которые закрываются после вывода.
  `LimitNOFILE=1024` достаточно для `max_conns = 504` или меньше.
- `MemoryMax`: когда меняешь `max_conns`, прочитай пик памяти под нагрузкой.
  Затем задай `MemoryMax` с запасом.

Эти лимиты не заменяют лимиты подключений из `bbs.toml`.
Лимиты `bbs.toml` считают подключения. Лимиты systemd ограничивают процесс.

`systemctl status velourbbs` показывает потоки и память. Пример:

```
      Tasks: 6 (limit: 64)
     Memory: 11.9M (max: 128M, available: 116.1M, peak: 15.8M)
```

`Tasks` показывает потоки и `TasksMax`. `Memory` показывает текущую память,
`MemoryMax` и пик. Пик показывает systemd 255 или новее.
Со старым systemd читай `Memory`, пока BBS под нагрузкой.

Полные правила: [setup-guide](../../specs/setup-guide.md), раздел 5,
и [telnet-term](../../specs/telnet-term.md), раздел 10.

### Результат

- `systemctl is-active velourbbs` выводит `active`.
- `sudo journalctl -u velourbbs` показывает строку `listen`.
- Telnet-клиент на порту 2323 показывает экран `main`.

`systemctl start` и `systemctl restart` завершаются успешно, даже когда `velourbbs serve`
сразу останавливается. Поэтому после каждого запуска проверяй состояние и лог.
