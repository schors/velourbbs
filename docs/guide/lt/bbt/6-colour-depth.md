# 6 užduotis. Spalvų gylis

## Terminalo spalvų gylis

Terminalas rodo 16, 256 arba 16777216 spalvų: tai jo „spalvų gylis“.
Prisijungiant BBS klausia terminalo ir skaito terminalo tipą, kurį siunčia klientas.
Gylį nurodo pirmoji tinkama lentelės eilutė:

| Terminalas | Gylis |
|---|---|
| Rodo RGB spalvą, pavyzdžiui, Windows Terminal, iTerm2, Ghostty, xterm su direct colour | 16777216 |
| PuTTY arba tmux | 16777216 |
| Tipas baigiasi `-direct` arba turi `truecolor` ar `24bit` | 16777216 |
| Tipas turi `256color` | 256 |
| Tipas baigiasi `-16color` arba `-8color` arba yra `ansi`, `dumb`, `vt52`, `vt100`, `vt102`, `vt220`, `vt320`, `rxvt`, `xterm-color`, `eterm-color` ar `cygwin` | 16 |
| Bet kuris kitas tipas, pavyzdžiui, `xterm` | 256 |

- Klientas be tipo gauna `ansi`. Jei jo terminalas neatsako, jis gauna 16 spalvų.
- Tera Term siunčia `xterm` ir nerodo RGB spalvų, todėl gauna 256 spalvas.

BBS pakeičia kiekvieną spalvą artimiausia to gylio spalva.
Esant 256 spalvoms, kreminis tekstas `#e9dccb` tampa neutralia pilka 253, ir ekranas praranda šilumą.
Todėl kiekvienam gyliui spalvas parink pats.
Indeksai 16–255 – tai xterm 256 spalvų paletės spalvos.
Indeksai 0–15 – tai VGA spalvos iš [telnet-term](../../../specs/telnet-term.md) 7.3 skyriaus.

Visos taisyklės: [telnet-term](../../../specs/telnet-term.md), 6 ir 7.3 skyriai.

## Gylio temos

```toml
[theme.256]
normal = "223/235"
title  = "230 bold"

[theme.16]
normal = "7/0"
title  = "15 bold"
```

- „Gylio tema“ duoda kitus stilius vienam gyliui.
  `[theme.256]` skirta 256 spalvoms. `[theme.16]` skirta 16 spalvų.
- Gylio temos stilius pakeičia to paties pavadinimo stilių iš `[theme.default]`.
  Kiti stiliai lieka.
- Kiekvienas gylio temos pavadinimas turi būti ir `[theme.default]` lentelėje arba būti `normal`, `more` ar `md_*` stilius (7 užduotis).
- Esant 16 spalvų, naudok indeksus 0–15. Esant 256 spalvoms, naudok indeksus 16–255.

Visos taisyklės: [screens-toml](../../../specs/screens-toml.md), 3.2 skyrius.

## Nuorodos į stilius

```toml
[theme.default]
dim   = "#9e8a7c"
title = "#f7c7a3 bold"
more  = "dim"
md_h1 = "title underline"

[theme.16]
dim   = "8"
title = "15 bold"
```

- Stilius gali nurodyti kitą stilių: tai „nuoroda“. `more = "dim"` suteikia `more` stilių `dim`.
- Nuoroda gali pridėti atributų: `md_h1 = "title underline"` yra `title` su pabraukimu.
- Nuoroda ima savo pavadinimo stilių iš tos pačios temos.
  Esant 16 spalvų, `more` yra `8`, o `md_h1` yra `15 bold underline`.
  Todėl jei stilius kiekviename gylyje yra kito stiliaus kopija, užrašyk jį kaip nuorodą į tą stilių.
- Nuoroda gali nurodyti kitą nuorodą.
  Nuoroda, kuri veda atgal į savo stilių, yra klaida: „ciklas“.

Visos taisyklės: [screens-toml](../../../specs/screens-toml.md), 3.2 skyrius.

## Eilutė „Toliau“

Ekranas, kuriame daugiau eilučių nei terminale, rodomas puslapiais. Kiekvienas puslapis, išskyrus paskutinį, baigiasi
eilute „Toliau“. Naudotojas paspaudžia bet kurį klavišą ir mato kitą puslapį.
Eilutėje „Toliau“ punktai neveikia.

```toml
[screen]
more = "    toliau ∙ bet kuris klavišas"
```

`more` nurodo eilutės „Toliau“ tekstą. Be `more` tekstas yra `-- More --`.
Stilius `more` nurodo jos spalvas.
Su toliau pateiktu failu `screens/screens.toml`, esant 12 eilučių, ekranas `about` turi du puslapius:

```
velourbbs preview about --size 80x12
```

```
page 1 80x12 fill spaces
...
|    toliau ∙ bet kuris klavišas                                                 |
styles
...
page 2 80x12 fill spaces
...
```

Be tarpo eilutės ekranas turi 14 eilučių. Todėl tarpo eilutė neduoda eilučių.
1 puslapyje yra ekrano 1–11 eilutės ir eilutė „Toliau“. 2 puslapyje yra 12–14 eilutės.

Visos taisyklės: [bbt-format](../../../specs/bbt-format.md), 6.4 skyrius,
ir [menu-nav](../../../specs/menu-nav.md), 5 skyrius.

## Failas `screens.toml`

Iš naujo įrašyk `screens/screens.toml` su šiuo tekstu:

`examples/lamp/lt/screens/screens.toml`:
```
[screen]
width = [80, 80]
more = "    toliau ∙ bet kuris klavišas"
fill = "spaces"

[theme.default]
normal = "#e9dccb/#2a201d"
dim    = "#9e8a7c"
line   = "#8a6a5c"
title  = "#f7c7a3 bold"
key    = "#ffb088 bold"
orn    = "#f2a27e"
more   = "dim"
md_h1     = "title"
md_h2     = "#f2a27e"
md_h3     = "#d9b99b"
md_strong = "title"
md_em     = "#d9b99b"
md_code   = "#c8b88a"
md_link   = "#d9b99b underline"
md_mark   = "orn"
md_frame  = "line"
md_dim    = "dim"

[theme.256]
normal = "223/235"
dim    = "137"
line   = "95"
title  = "230 bold"
key    = "216 bold"
orn    = "216"
md_h2     = "216"
md_h3     = "180"
md_em     = "180"
md_code   = "144"
md_link   = "180 underline"

[theme.16]
normal = "7/0"
dim    = "8"
line   = "8"
title  = "15 bold"
key    = "15 bold"
orn    = "7"
md_h2     = "15"
md_h3     = "15"
md_em     = "7"
md_code   = "15"
md_link   = "15 underline"
```

- Esant 256 spalvoms, fonas yra tamsiai pilka 235, o tekstas – kreminis 223, ne pilka 253.
- Esant 16 spalvų, ekranas pilkas: juodas fonas, šviesiai pilkas tekstas,
  tamsiai pilki `dim` ir `line`, balti `title` ir `key`.
- Stiliai `md_*` yra tekstų stiliai. Juos naudoja 7 užduotis.
- `more`, `md_h1`, `md_strong`, `md_mark`, `md_frame` ir `md_dim` yra nuorodos.
  Kiekviename gylyje jos ima to gylio stilius `dim`, `title`, `orn` ir `line`.

## Pažiūrėk į kiekvieną gylį

`velourbbs preview` naudoja tipą `xterm`, todėl 256 spalvas. Kitą gylį nustatyk su `--set`:

```
velourbbs preview main --set term.colors=16
velourbbs preview main --set term.type=vt100
```

Abi komandos rodo ekraną su `[theme.16]` stiliais.
Langelių išklotinė rodo temos spalvas taip, kaip jas užrašei.
Pridėk `--view`, kad savo terminale pamatytum gylio spalvas:

```
velourbbs preview main --view
velourbbs preview main --view --set term.colors=16
```

`--view` duoda artimiausias gylio spalvas, kaip jas gauna telnet klientas.
Esant 256 ir 16777216 spalvų, menas kiekviename terminale turi tas pačias spalvas.
Esant 16 spalvų, terminalas rodo indeksus 0–15 savo paletėje.

## Šviesi tema

Kai kurių terminalų fonas šviesus. Tamsus fonas šviesiame terminale atrodo netinkamai.
„Šviesi tema“ duoda tokiems terminalams kitus stilius.

```toml
[screen]
theme = "auto"

[theme.light.default]
normal = "#3a2a24/#f4ece0"
title  = "#8a3a1c bold"

[theme.light.16]
normal = "0/15"
title  = "4 bold"
```

- `[screen] theme` parenka temą seanso pradžioje: `dark`, `light` arba `auto`.
  Be `theme` seansas naudoja tamsią temą.
- Su `auto` temą parenka laukas `term.background` (5 užduotis).
  `light` duoda šviesią temą. Terminalas be atsakymo gauna tamsią temą.
- „Šviesiosios lentelės“ `[theme.light.default]`, `[theme.light.16]` ir `[theme.light.256]`
  duoda kitus stilius šviesioje temoje.
- Gylio šviesi tema yra to gylio tema su `[theme.light.default]` stiliais,
  paskui su to gylio šviesiosios lentelės stiliais.
- Kiekvienas šviesiosios lentelės pavadinimas turi būti temos pavadinimas, kaip ir gylio temoje.
- Šviesioje temoje `normal` trūkstamą spalvą ima iš `0/15`: juodas tekstas baltame fone.

„Lempa“ lieka su tamsia tema. Kad išbandytum šviesią temą, pridėk raktą `theme` į failo `screens/screens.toml`
lentelę `[screen]`, o šviesiąsias lenteles – į failo pabaigą.
Paskui nustatyk terminalo foną:

```
velourbbs preview main --set term.background=light
velourbbs preview main --set term.background=light --set term.colors=16
```

Pirmoji komanda rodo `[theme.light.default]` stilius.
Antroji komanda dar rodo ir `[theme.light.16]` stilius.

### Naudotojo tema

Meniu punktai su veiksmu `theme` leidžia naudotojui pasirinkti temą.
Kad juos išbandytum, pridėk šiuos punktus į `screens/main.bbt` priekinę dalį,
o klavišų eilutę – į jo tekstą:

```toml
[[item]]
key = "L"
do = "theme light"
[[item]]
key = "D"
do = "theme dark"
[[item]]
key = "B"
do = "theme auto"
```

```
[L] šviesi  [D] tamsi  [B] pagal terminalą: {session.theme}
```

- `theme light` ir `theme dark` parenka temą. `theme auto` grąžina pasirinkimą laukui `term.background`.
- BBS iš naujo parodo dabartinį ekraną su nauja tema. Failo ekrane BBS rodo tą patį puslapį.
- Pasirinkimas galioja iki seanso pabaigos.
- Laukas `session.theme` duoda pasirinkimą: `dark`, `light` arba `auto`.
  Seanso pradžioje jis yra `[screen] theme`.

Paleisk punktus su `--keys`. `--set session.theme` nustato pasirinkimą pradžioje:

```
velourbbs preview main --keys D,L,B --set term.background=light
velourbbs preview main --set session.theme=light
```

Pirmoji komanda rodo `main` su tamsia tema po `D` ir su šviesia tema po `L`.
Po `B` ji vėl rodo šviesią temą, nes `term.background` yra `light`.

Visos taisyklės: [screens-toml](../../../specs/screens-toml.md), 3.1 ir 3.2 skyriai,
ir [menu-nav](../../../specs/menu-nav.md), 4 skyrius.

## Atributas `blink`

Stilius gali turėti atributą `blink`, bet terminalai jį rodo skirtingai:

| Terminalas | `blink` |
|---|---|
| PuTTY | Išjungtas pagal numatytuosius nustatymus. Vietoj mirksėjimo rodo ryškų foną. |
| Tera Term | Nemirksi. Rodo mirksėjimo spalvą. |
| Windows Terminal | Mirksi. Visi langeliai mirksi kartu. |
| xterm | Mirksi, jei versija palaiko mirksėjimą. Kitaip – paryškintas. |
| iTerm2 | Išjungtas pagal numatytuosius nustatymus. |
| Ghostty | Nemirksi. |
| tmux | Perduoda atributą išoriniam terminalui. |

Todėl nenaudok `blink` tekstui, kuris kiekviename terminale turi atrodyti vienodai.
Kad ekranas judėtų, vietoj `blink` naudok [10 užduoties](10-effects.md) efektus.
Jie kiekviename terminale atrodo vienodai.

Visos taisyklės: [live-screen](../../../specs/live-screen.md), 8 skyrius.

## Rezultatas

- `velourbbs check` nieko neišveda.
- `velourbbs preview main` išveda 4 eilutę su `[theme.256]` stiliais:

```
4: 1-7 223/235, 8-12 230/235 bold, 13-14 223/235, 15-16 216/235, 17-39 137/235, 40-80 223/235
```

- `velourbbs preview main --set term.colors=16` išveda 4 eilutę su `[theme.16]` stiliais:

```
4: 1-7 7/0, 8-12 15/0 bold, 13-16 7/0, 17-39 8/0, 40-80 7/0
```

- PuTTY rodo tamsiai pilką foną ir kreminį, ne pilką tekstą.
  PuTTY nustatyk Connection > Data > „Terminal-type string“ į `vt100` ir vėl prisijunk:
  PuTTY rodo pilką 16 spalvų ekraną.
  Paskui grąžink „Terminal-type string“ reikšmę `xterm`.
