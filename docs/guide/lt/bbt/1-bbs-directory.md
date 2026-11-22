# 1 užduotis. BBS katalogas

1. Sukurk katalogą, pavyzdžiui, `mybbs`. Jame sukurk katalogą `screens`.
2. Sukurk tuščią failą `bbs.toml`.
3. Sukurk `screens/main.bbt` su šiuo tekstu:

```
+++
[[item]]
key = "G"
do = "logoff"
+++
Labas! Tai mano BBS.

G  iki pasimatymo
```

Eilutės tarp dviejų `+++` eilučių yra „priekinė dalis“. Ji nurodo ekrano meniu.
Čia meniu turi vieną punktą: klavišas `G` baigia seansą (`logoff`).
Ekranas turi turėti bent vieną punktą. Punktų neturi tik ekranai `goodbye` ir `shutdown` (4 užduotis).
Kitos eilutės yra „kūnas“: ekrano tekstas.
BBS pirmiausia rodo ekraną `main`.

## Patikrink failus

Paleisk:

```
velourbbs check
```

Kai failai teisingi, komanda nieko neišveda.
Kitaip ji išveda po vieną eilutę kiekvienai klaidai ar įspėjimui:

```
screens/_x.bbt:1:1: warning: no screen uses this file
screens/main.bbt:4:6: item "G": do: unknown action "logof"
```

Eilutėje yra failas, eilutės numeris, stulpelis ir pranešimas.
Punkto klaida nurodo neteisingos reikšmės vietą: čia `"logof"` 4 eilutėje.
Įspėjimas turi `warning:` prieš pranešimą. Įspėjimas nesustabdo BBS. Klaida sustabdo.
Visos taisyklės: [screen-render](../../../specs/screen-render.md), 3.1 skyrius.

## Pažiūrėk į ekraną

Paleisk:

```
velourbbs preview main
```

Komanda išveda ekrano „langelių išklotinę“ 80 stulpelių ir 24 eilučių terminalui:

```
page 1 80x24 fill none
|Labas! Tai mano BBS.                                                            |
|                                                                                |
|G  iki pasimatymo                                                               |
styles
1: 1-80 7/0
2: 1-80 7/0
3: 1-80 7/0
```

- Pirmoje eilutėje yra puslapis, terminalo dydis ir eilučių po ekranu užpildas.
- Kiekviena ekrano eilutė yra tarp dviejų `|`.
- Kiekvienoje eilutėje po `styles` yra eilutės numeris ir tos eilutės „stilių atkarpos“.
  Atkarpoje yra pirmas stulpelis, paskutinis stulpelis ir stilius. `7/0` – šviesiai pilka ant juodos.

Kitiems dydžiams naudok `--size`, pavyzdžiui, `velourbbs preview main --size 132x43`.
Visos taisyklės: [screen-render](../../../specs/screen-render.md), 3.2 ir 3.3 skyriai.

## Pažiūrėk į ekraną telnet kliente

1. Paleisk `velourbbs serve`. Serveris klausosi 2323 prievado.
2. Prijunk telnet klientą prie `localhost`, prievadas 2323. Naudok PuTTY arba Tera Term:
   jie rodo UTF-8. PuTTY rodo 16777216 spalvų, Tera Term – 256 spalvas.
   SyncTERM ir NetRunner nerodo UTF-8.
   PuTTY programoje pasirink ryšio tipą „Telnet“.
   Nustatyk „Close window on exit“ į „Never“: kitaip PuTTY uždaro langą seanso pabaigoje.
3. Paspausk `G`, kad baigtum seansą.

Serveris skaito failus tik paleidimo metu. Po kiekvieno failo pakeitimo
sustabdyk serverį klavišais Ctrl+C ir paleisk iš naujo.
Visos taisyklės: [telnet-term](../../../specs/telnet-term.md), 3 ir 6 skyriai.

## Rezultatas

- `velourbbs check` nieko neišveda.
- Telnet klientas rodo dvi teksto eilutes. `G` baigia seansą.
