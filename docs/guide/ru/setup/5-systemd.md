# 5. systemd

Нужен systemd 240 или новее с cgroup v2.
В Debian 12, Ubuntu 24.04 и RHEL 9 они есть по умолчанию.

## Unit

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

## Установка и запуск

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

## Изменения

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

## Лимиты

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

Полные правила: [setup-guide](../../../specs/setup-guide.md), раздел 5,
и [telnet-term](../../../specs/telnet-term.md), раздел 10.

## Результат

- `systemctl is-active velourbbs` выводит `active`.
- `sudo journalctl -u velourbbs` показывает строку `listen`.
- Telnet-клиент на порту 2323 показывает экран `main`.

`systemctl start` и `systemctl restart` завершаются успешно, даже когда `velourbbs serve`
сразу останавливается. Поэтому после каждого запуска проверяй состояние и лог.
