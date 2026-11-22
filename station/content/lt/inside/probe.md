# Terminalo atpažinimas

## Ką daro

Telnet duoda tik dydį ir tipo pavadinimą. Kitko Velour klausia paties terminalo
valdymo sekomis: koduotės, kirilicos pločio, spalvų gylio, programos pavadinimo
ir fono spalvos. Apklausa vyksta prieš pirmą ekraną.

## Kaip

Kai telnet derybos baigiasi, klientas jau simbolių režime ir nespausdina atsakymų.
Tada Velour vienu rašymu siunčia užklausas tokia tvarka:

1. Koduotės patikra: CR, raidė `â` (baitai `C3 A2`), žymeklio vietos užklausa `ESC[6n`.
   Tai DSR iš ECMA-48 ([ECMA-48](https://ecma-international.org/publications-and-standards/standards/ecma-48/)).
   Atsakymas CPR - `ESC[<eilutė>;<stulpelis>R`
   ([DSR-CPR, VT510](https://vt100.net/docs/vt510-rm/DSR-CPR.html)).
2. Kirilicos plotis: CR, raidė `ж` (baitai `D0 B6`), `ESC[6n`.
3. Truecolor: Velour nustato spalvą `ESC[38;2;1;2;3m` ir klausia, kokia spalva nustatyta:
   DECRQSS SGR, `ESC P $ q m ESC \`
   ([DECRQSS](https://vt100.net/docs/vt510-rm/DECRQSS.html),
   atsakymas [DECRPSS](https://vt100.net/docs/vt510-rm/DECRPSS.html)).
   Tada atstato spalvą: `ESC[0m`.
4. Programos pavadinimas: XTVERSION, `ESC[>0q`.
5. Antriniai atributai: DA2, `ESC[>c`.
6. Teksto ir fono spalvos: OSC 10 ir OSC 11 su `?`, pavyzdžiui `ESC]11;?ESC \`.
7. Paletė: OSC 4 su `?`, atskira užklausa kiekvienam numeriui: 0-15, 256 ir 258.
8. Pirminiai atributai: DA1, `ESC[c`.
9. Eilutės trynimas: `CR ESC[K`. Jis pašalina užklausą, kurią terminalas išspausdino kaip tekstą.

4-8 užklausas aprašo xterm žinynas
([xterm control sequences](https://invisible-island.net/xterm/ctlseqs/ctlseqs.html)).

DA1 eina paskutinė. Į ją atsako visi tiksliniai terminalai, ir atsakymai ateina
užklausų tvarka. Todėl atsakymas į DA1 reiškia: daugiau atsakymų nebus.
Velour laukia jo arba 1 sekundės tylos.

## Ką gauna

### Koduotė

UTF-8 koduotėje raidė `â` užima vieną langelį, ir žymeklis atsiduria 2 stulpelyje.
8 bitų koduotėje du baitai yra du simboliai, ir žymeklis atsiduria 3 stulpelyje.
Tada Velour angliškai paaiškina, kaip įjungti UTF-8, ir uždaro ryšį.
Jei atsakymo nėra, Velour laiko, kad tai UTF-8.

Kodėl raidė `â`: pagal Unicode jos plotis N, siauras ir CJK terminale
([UAX #11](https://www.unicode.org/reports/tr11/)). Jos baitai yra spausdinami simboliai
ISO-8859-1, CP437, CP866 ir KOI8-R koduotėse. Baitai 0x80-0x9F 8 bitų terminaluose gali būti
C1 valdymo simboliai, todėl apklausoje jų nėra.

### Kirilicos plotis

Pagal UAX #11 kirilicos plotis yra A, Ambiguous: CJK terminalas gali ją piešti dviem langeliais.
Jei po `ж` žymeklis yra 3 stulpelyje, `term.ambiguous` lygus 2.
Tada ekranai sugenda, ir įėjimo ekranas paaiškina, kaip išjungti šį nustatymą.
Apklausa matuoja kirilicą, ne linijas: Windows Terminal plačiajame režime linijos
U+2500-U+259F lieka siauros, o kirilica tampa plati.

### Spalvų gylis

Pirma tinkanti taisyklė duoda `term.colors`:

1. DECRQSS atsakymas turi spalvą 1, 2, 3: terminalas išsaugojo RGB. Tai truecolor.
   Terminalas, kuris RGB suveda į paletę, atsako numeriu: Tera Term siunčia `0;30`.
2. DA2 atsakymas yra iš PuTTY (`0;136;0`) arba iš tmux (pirmas skaičius `84`). Jie rodo RGB,
   bet į DECRQSS neatsako. Tai truecolor.
3. `term.type` baigiasi `-direct` arba turi `truecolor` ar `24bit`: truecolor.
4. `term.type` turi `256color`: 256 spalvos.
5. `term.type` yra `ansi`, `vt100`, `xterm-color` ar kitas pavadinimas, kur 16 spalvų tikra: 16.
6. Bet kuris kitas tipas: 256.

5 taisyklėje tik tikri pavadinimai: terminalas be 256 spalvų skaito `38;5;n`
kaip kitus atributus, pavyzdžiui `5` kaip mirksėjimą.

### Pavadinimas ir fonas

- XTVERSION duoda programos pavadinimą ir versiją, pavyzdžiui `XTerm(412)` arba `tmux 3.6`.
  Tai `term.name`. Jis skirtas tik rodyti: ekranai pagal jį nesirenka išvesties.
- OSC 11 duoda fono spalvą, `term.bg_color`. PuTTY į OSC 11 neatsako,
  todėl jam fonas imamas iš OSC 4 numeriui 258.
- Fonas šviesus, jei jo santykinis skaistis didesnis nei 0,5. Formulė - iš WCAG 2.2
  ([relative luminance](https://www.w3.org/TR/WCAG22/#dfn-relative-luminance)).
  Tema `auto` tada ima šviesias spalvas.
- Paletė 0-15 iš OSC 4. Jei ji nesutampa su VGA, meno spalvos siunčiamos RGB.
  Taip menas išlaiko dailininko spalvas.

Visus atsakymus Velour rašo viena žurnalo eilute. Pagal juos spalvų taisykles galima patikslinti.

## Ribos

- PuTTY neatsako į DECRQSS, OSC 10 ir OSC 11. 0.83 ir senesnės versijos
  išspausdina DECRQSS užklausą kaip tekstą `$qm`. Ją ištrina 9 žingsnis.
- PuTTY neatsako į OSC 4 su keliais numeriais. Todėl kiekvienas numeris turi savo užklausą.
- tmux į užklausas atsako pats ir į išorę siunčia tik OSC 4. Po OSC 4 3.6 ir naujesnės versijos
  vėlina kitus atsakymus: iki 3,5 sekundės 3.6 versijoje, apie 0,6 sekundės 3.7 versijoje.
  Todėl OSC 4 stovi iškart prieš DA1, ir vėluoti gali tik DA1 atsakymas.
- tmux atsako į CPR savo pločiais. Todėl `term.ambiguous` lygus 1,
  net jei išorinis terminalas piešia kirilicą plačią.
- Windows Terminal 1.25 ir naujesnė po nežinomos sekos gali nusiųsti savo CPR užklausą
  ir pasiimti kitą CPR atsakymą. Todėl CPR patikros eina pirmos.
- Windows Terminal 1.21 ir senesnė pati iš karto atsako į DA1. Po DA1 atsakymo nieko neateina.
- SyncTERM ir NetRunner nemoka UTF-8. Jie gauna paaiškinimą ir neįeina.
- Windows telnet.exe vadina save `ANSI` ir gauna 16 spalvų. ANSI menas ir linijos
  jame gali sugesti.
- MUD klientai, pavyzdžiui Mudlet, nevykdo `ESC[2J` ir žymeklio vietų:
  puslapis tampa viena eilute.
- Dėl CP437 terminalų ir terminalų be ANSI sprendimo kol kas nėra.
