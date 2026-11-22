# 5. systemd

Tau reikia systemd 240 ar naujesnės versijos su cgroup v2.
Debian 12, Ubuntu 24.04 ir RHEL 9 juos turi pagal numatymą.

BBS veikia tavo naudotojo vardu, tavo systemd naudotojo tvarkytuve: tai naudotojo vienetas.
Root teisės reikalingos tik vieną kartą, linger nustatymui.

## Vienetas

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

| Eilutė | Reikšmė |
|---|---|
| `Description=Velour BBS` | Pavadinimas `systemctl` išvestyje ir žurnale. |
| `Type=exec` | `systemctl start` baigiasi klaida, kai systemd negali paleisti programos. |
| `ExecStart=/usr/local/bin/velourbbs serve` | Komanda. |
| `WorkingDirectory=/srv/velourbbs` | BBS katalogas. |
| `Restart=on-failure` | Po gedimo systemd vėl paleidžia BBS. |
| `RestartSec=5` | systemd laukia 5 s ir tik tada vėl paleidžia BBS. Naudotojo vienetas negali laukti tinklo įkrovos metu. Jei `listen` adreso dar nėra, `velourbbs serve` sustoja. Tada systemd bando vėl, kol adresas atsiranda. |
| `MemoryMax=128M` | Atminties riba. Ji skirta `bbs.toml` numatytosioms reikšmėms. „Lempa“ naudojo 6,8 MB paleidimo metu, 11,9 MB su 64 seansais ir 15,8 MB su 300 seansų. |
| `TasksMax=64` | Gijų riba. |
| `LimitNOFILE=1024` | Atvirų failų riba. |
| `WantedBy=default.target` | `systemctl --user enable` paleidžia BBS, kai pasileidžia tavo naudotojo tvarkytuvė. |

## Įdiegimas ir paleidimas

Įjunk savo naudotojo linger. Tada tavo naudotojo tvarkytuvė pasileidžia įkrovos metu,
net kai tu nesi prisijungęs:

```
sudo loginctl enable-linger "$USER"
```

Pirminio kodo saugykloje paleisk:

```
mkdir -p ~/.config/systemd/user
cp examples/deploy/velourbbs.service ~/.config/systemd/user/
systemctl --user daemon-reload
systemctl --user enable --now velourbbs
```

`enable --now` paleidžia BBS ir paleidžia ją, kai pasileidžia tavo naudotojo tvarkytuvė.

BBS rašo žurnalą į standartinį klaidų srautą. systemd jį įrašo į journal.
Perskaityk žurnalą:

```
journalctl --user -u velourbbs
```

## Pakeitimai

Po failo pakeitimo BBS kataloge:

```
cd /srv/velourbbs
velourbbs check && systemctl --user restart velourbbs
```

Po `~/.config/systemd/user/velourbbs.service` pakeitimo:

```
systemctl --user daemon-reload
systemctl --user restart velourbbs
```

## Duomenų failas

Kol BBS veikia, ji rašo duomenų failą `/srv/velourbbs/velourbbs.jsonl`.
Failas yra tekstinis: kiekvienoje eilutėje vienas JSON objektas. Sustabdyk BBS, prieš kopijuodamas ar redaguodamas failą.

Padaryk atsarginę kopiją:

```
systemctl --user stop velourbbs
cp /srv/velourbbs/velourbbs.jsonl ~/velourbbs.jsonl.bak
systemctl --user start velourbbs
```

Norėdamas ją atkurti, sustabdyk BBS ir nukopijuok atsarginę kopiją atgal į `/srv/velourbbs/velourbbs.jsonl`.

Nauja `velourbbs` laida gali pakeisti failo formatą. Tada senesnė laida
negali perskaityti failo. Padaryk atsarginę kopiją, prieš įdiegdamas naują laidą.

Norėdamas redaguoti failą, sustabdyk BBS:

```
systemctl --user stop velourbbs
```

Redaguok `/srv/velourbbs/velourbbs.jsonl` teksto redaktoriumi. Tada paleisk:

```
cd /srv/velourbbs
velourbbs check && systemctl --user start velourbbs
```

Visos taisyklės: [data-store](../../../specs/data-store.md), 8 skyrius.

## Ribos

- `TasksMax` yra ne mažiau kaip procesorių skaičius + 32. `nproc` duoda procesorių skaičių.
  Go paleidžia gijas kiekvienam procesoriui. Kai Go negali paleisti gijos, BBS sustoja
  su visais seansais. `TasksMax=64` pakanka 32 ar mažiau procesorių.
- `LimitNOFILE` yra ne mažiau kaip `2 × max_conns + 16`. BBS turi ne daugiau kaip `max_conns`
  ryšių ir ne daugiau kaip `max_conns` ryšių, kurie užsidaro po išvesties.
  `LimitNOFILE=1024` pakanka, kai `max_conns = 504` ar mažiau.
- `MemoryMax`: kai keiti `max_conns`, perskaityk atminties piką esant apkrovai.
  Tada nustatyk `MemoryMax` su atsarga.

Šios ribos nepakeičia `bbs.toml` ryšių ribų.
`bbs.toml` ribos skaičiuoja ryšius. systemd ribos riboja procesą.

`systemctl --user status velourbbs` rodo gijas ir atmintį. Pavyzdys:

```
      Tasks: 6 (limit: 64)
     Memory: 11.9M (max: 128M, available: 116.1M, peak: 15.8M)
```

`Tasks` rodo gijas ir `TasksMax`. `Memory` rodo dabartinę atmintį,
`MemoryMax` ir piką. systemd 255 ar naujesnė versija rodo piką.
Su senesne systemd versija skaityk `Memory`, kol BBS veikia esant apkrovai.

Visos taisyklės: [setup-guide](../../../specs/setup-guide.md), 4 skyrius,
ir [telnet-term](../../../specs/telnet-term.md), 10 skyrius.

## Rezultatas

- `systemctl --user is-active velourbbs` duoda `active`.
- `journalctl --user -u velourbbs` rodo eilutę `listen`.
- Telnet klientas prievadu 2323 rodo ekraną `main`.

`systemctl start` ir `systemctl restart` pavyksta ir tada, kai `velourbbs serve`
iš karto sustoja. Todėl po kiekvieno paleidimo patikrink būseną ir žurnalą.
