# 4. Komandos

## `velourbbs check`

Paleisk ją po kiekvieno failo pakeitimo ir prieš kiekvieną BBS paleidimą iš naujo.
Ji tikrina `bbs.toml`, ekranus ir duomenų failą. Duomenų failą ji skaito tik tada, kai failas egzistuoja.
Ji niekada nerašo duomenų failo. Kai failai teisingi, ji nieko neišveda.
Klaida faile `bbs.toml` ar `screens/screens.toml` sustabdo patikrą prieš ekranus.
Ištaisyk ją ir vėl paleisk patikrą.
Kitaip ji išveda po vieną eilutę kiekvienai klaidai ar įspėjimui:

```
screens/main.bbt:7:8: item "S": when: unknown field "term.colours"
screens/about.bbt:4:1: warning: art "logo.ans" is 132 cells wide, max width is 100
```

Eilutė nurodo failą, eilutės numerį, stulpelį ir pranešimą.
Įspėjimas nesustabdo BBS. Klaida ją sustabdo, ir komanda baigiasi su būsena 1.
Visos taisyklės: [screen-render](../../../specs/screen-render.md), 3.1 skyrius.

## `velourbbs preview`

Ji parodo ekraną kaip tekstą, be telnet kliento:

```
velourbbs preview main
```

`--size` nustato terminalo dydį, pavyzdžiui, `--size 132x43`.
`--set term.colors=16` parodo ekraną terminalui su 16 spalvų.
Visos taisyklės: [screen-render](../../../specs/screen-render.md), 3.2 skyrius.
Išsamiau: [BBT vadovo](../bbt/1-bbs-directory.md) 1 užduotis.

`--view` parodo ekraną spalvotai tavo terminale, vietoj tekstinės išklotinės:

```
velourbbs preview main --view
velourbbs preview main --view --set term.colors=16
```

Ji rašo puslapius vieną po kito ir nelaukia klavišo.
Kiekvienas puslapis prasideda savo eilute `page`.
Visos taisyklės: [screen-render](../../../specs/screen-render.md), 3.7 skyrius.

## `velourbbs preview --keys`

Ji paspaudžia klavišus ir parodo kiekvieną ekraną. Naudok ją ekrano klavišams išbandyti:

```
velourbbs preview main --keys A,ESC,G
```

`--keys` ir `--view` kartu parodo kiekvieną puslapį spalvotai.

Išsamiau: [BBT vadovo](../bbt/4-menu.md) 4 užduotis.

## `velourbbs fields`

Ji išvardija BBS laukus su jų pločiais ir pavyzdinėmis reikšmėmis:

```
velourbbs fields
```

`velourbbs fields --sample` įrašo `velourbbs preview` pavyzdžių rinkinius į `sample.toml`.
Išsamiau: [BBT vadovo](../bbt/5-fields.md) 5 ir 8 užduotys.

## `velourbbs art import`

Ji paverčia meno redaktoriaus failą į formą, kurią skaito BBS:

```
velourbbs art import art/logo.ans screens/logo.ans
```

Išsamiau: [BBT vadovo](../bbt/9-art.md) 9 užduotis.

## `velourbbs serve`

Ji paleidžia telnet serverį. Serveris klausosi prievadu 2323.

- Pirma ji atlieka `velourbbs check` patikras. Po klaidos ji baigiasi su būsena 1.
- Tada ji skaito duomenų failą. Kai failo nėra, ji jį sukuria.
  Po duomenų failo klaidos ji baigiasi su būsena 1.
- Jei ji negali klausytis `listen` adresu, ji baigiasi su būsena 1.
- Ji rašo žurnalą į standartinį klaidų srautą: eilutę `listen` su adresu ir jo šaltiniu,
  eilutę su kiekvieno kliento terminalo reikšmėmis ir eilutę `close`
  kiekvienam ryšiui.
- Ji skaito failus tik paleidimo metu. Po failo pakeitimo paleisk ją iš naujo.
- Ctrl+C arba SIGTERM ją sustabdo. Ji parodo ekraną `shutdown` kiekvienam atviram seansui,
  uždaro kiekvieną ryšį ne vėliau kaip po 5 s ir baigiasi su būsena 0.

Žurnalo pavyzdys:

```
time=2026-11-16T09:42:48.905Z level=INFO msg=listen addr=[::]:2323 source=bbs.toml
time=2026-11-16T09:42:49.929Z level=INFO msg=probe addr=127.0.0.1 ttype=xterm-256color naws=80x24 cpr=1;2 type=xterm-256color cols=80 rows=24 charset=utf-8 colors=256 time=355.507µs
time=2026-11-16T09:42:51.231Z level=INFO msg=close addr=127.0.0.1 reason=logoff time=1.30238633s
```

Visos taisyklės: [telnet-term](../../../specs/telnet-term.md), 10 skyrius.

## Rezultatas

1. Kataloge `/srv/velourbbs` paleisk `velourbbs serve`. Ji įrašo eilutę `listen`.
2. Prijunk telnet klientą prie serverio, prievadas 2323, pavyzdžiui, `telnet localhost 2323`.
   Klientas parodo ekraną `main`: galvutę „Lempa“ ir meniu `A R L T G`.
   Paspausk `G`, kad baigtum seansą.
3. Sustabdyk serverį klavišais Ctrl+C. 5 skyrius naudoja prievadą 2323.
