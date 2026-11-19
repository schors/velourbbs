# 5. systemd

Нужен systemd 240 или новее с cgroup v2.
В Debian 12, Ubuntu 24.04 и RHEL 9 они есть по умолчанию.

BBS работает от твоего пользователя, в твоём пользовательском менеджере systemd: это user unit.
Root нужен только один раз, для linger.

## Unit

`examples/deploy/velourbbs.service`:
```
[Unit]
Description=Velour BBS

[Service]
Type=exec
ExecStart=/usr/local/bin/velourbbs serve
WorkingDirectory=/srv/velourbbs
Restart=on-failure
RestartSec=5
MemoryMax=128M
TasksMax=64
LimitNOFILE=1024

[Install]
WantedBy=default.target
```

| Строка | Значение |
|---|---|
| `Description=Velour BBS` | Имя в `systemctl` и в логе. |
| `Type=exec` | `systemctl start` завершается ошибкой, когда systemd не может запустить программу. |
| `ExecStart=/usr/local/bin/velourbbs serve` | Команда. |
| `WorkingDirectory=/srv/velourbbs` | Каталог BBS. |
| `Restart=on-failure` | systemd снова запускает BBS после сбоя. |
| `RestartSec=5` | systemd ждёт 5 с перед новым запуском BBS. User unit не может ждать сеть при загрузке. Если адреса из `listen` ещё нет, `velourbbs serve` останавливается. Тогда systemd пробует снова, пока адрес не появится. |
| `MemoryMax=128M` | Лимит памяти. Он рассчитан на значения `bbs.toml` по умолчанию. «Лампа» использовала 6,8 МБ при запуске, 11,9 МБ с 64 сеансами и 15,8 МБ с 300 сеансами. |
| `TasksMax=64` | Лимит потоков. |
| `LimitNOFILE=1024` | Лимит открытых файлов. |
| `WantedBy=default.target` | `systemctl --user enable` запускает BBS, когда запускается твой пользовательский менеджер. |

## Установка и запуск

Включи linger для своего пользователя. Тогда твой пользовательский менеджер запускается
при загрузке системы, даже когда ты не вошёл в систему:

```
sudo loginctl enable-linger "$USER"
```

В репозитории исходного кода выполни:

```
mkdir -p ~/.config/systemd/user
cp examples/deploy/velourbbs.service ~/.config/systemd/user/
systemctl --user daemon-reload
systemctl --user enable --now velourbbs
```

`enable --now` запускает BBS и запускает её, когда запускается твой пользовательский менеджер.

BBS пишет лог в стандартный поток ошибок. systemd записывает его в журнал.
Прочитай лог:

```
journalctl --user -u velourbbs
```

## Изменения

После изменения файла в каталоге BBS:

```
cd /srv/velourbbs
velourbbs check && systemctl --user restart velourbbs
```

После изменения `~/.config/systemd/user/velourbbs.service`:

```
systemctl --user daemon-reload
systemctl --user restart velourbbs
```

## Файл данных

BBS пишет файл данных `/srv/velourbbs/velourbbs.jsonl`, пока работает.
Это текстовый файл: по одному объекту JSON на строку. Останови BBS, прежде чем копировать
или редактировать файл.

Сделай резервную копию:

```
systemctl --user stop velourbbs
cp /srv/velourbbs/velourbbs.jsonl ~/velourbbs.jsonl.bak
systemctl --user start velourbbs
```

Чтобы восстановить её, останови BBS и скопируй копию обратно в `/srv/velourbbs/velourbbs.jsonl`.

Новый выпуск `velourbbs` может изменить формат файла. Тогда старый выпуск
не может прочитать файл. Сделай резервную копию перед установкой нового выпуска.

Чтобы отредактировать файл, останови BBS:

```
systemctl --user stop velourbbs
```

Отредактируй `/srv/velourbbs/velourbbs.jsonl` в текстовом редакторе. Затем выполни:

```
cd /srv/velourbbs
velourbbs check && systemctl --user start velourbbs
```

Полные правила: [data-store](../../../specs/data-store.md), раздел 8.

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

`systemctl --user status velourbbs` показывает потоки и память. Пример:

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

- `systemctl --user is-active velourbbs` выводит `active`.
- `journalctl --user -u velourbbs` показывает строку `listen`.
- Telnet-клиент на порту 2323 показывает экран `main`.

`systemctl start` и `systemctl restart` завершаются успешно, даже когда `velourbbs serve`
сразу останавливается. Поэтому после каждого запуска проверяй состояние и лог.
