# 5 užduotis. Laukai ir sąlygos

Šioje užduotyje rašai ekraną `term`. Jis rodo, ką BBS žino apie naudotojo terminalą.

## Laukai

„Laukas“ yra žymė, kurią BBS pakeičia reikšme:

```
       {!dim}Dydis{!}    {term.cols}×{term.rows}
```

Terminale su 80 stulpelių ir 24 eilutėmis ši eilutė rodo `Dydis    80×24`.

| Laukas | Reikšmė |
|---|---|
| `term.cols` | Terminalo stulpelių skaičius, nuo 20 iki 255. |
| `term.rows` | Terminalo eilučių skaičius, nuo 10 iki 100. |
| `term.colors` | Spalvų gylis: `16`, `256` arba `16777216` (6 užduotis). |
| `term.type` | Terminalo tipas, kurį siunčia klientas, mažosiomis raidėmis. Be tipo – `ansi`. |
| `term.name` | Pavadinimas ir versija, kuriuos praneša terminalas, pavyzdžiui, `XTerm(412)`. Gali būti tuščias. Naudok jį tik rodymui. |
| `term.ambiguous` | `2`, kai terminalas rodo East Asian Ambiguous simbolius, pavyzdžiui, kirilicos raides ir `■`, dviejuose langeliuose. Kitaip `1`. Išdėstymas nesikeičia, todėl rėmeliai ir stulpeliai sugenda. Naudok jį, kad parodytum naudotojui, kaip išjungti šį nustatymą. |
| `term.bg_color` | Fono spalva, kurią praneša terminalas, kaip `#rrggbb` mažosiomis raidėmis. Tuščias, kai terminalas jos nepraneša. |
| `term.background` | `light`, kai `term.bg_color` yra šviesi spalva, pavyzdžiui, `#bcbcbc` ar šviesesnė pilka. `dark` – tamsiai spalvai. Tuščias, kai `term.bg_color` tuščias. |

- Laukas be tarpų riestiniuose skliaustuose užima savo reikšmės plotį.
- Laukas su tarpais riestiniuose skliaustuose turi „fiksuotą plotį“: žymės plotį kartu su skliaustais.
  BBS papildo trumpą reikšmę tarpais ir nukerpa ilgos reikšmės pabaigą.
  Tarpai po pavadinimo lygiuoja reikšmę pagal kairį kraštą. Tarpai prieš pavadinimą lygiuoja ją pagal dešinį kraštą.

```
[{term.type      }]
[{      term.type}]
```

Kiekviena žymė yra 17 langelių pločio. BBS rodo:

```text
[xterm            ]
[            xterm]
```

- `{` pradeda žymę, kai po jo eina lotyniška raidė, tarpas arba ASCII skyrybos ženklas,
  išskyrus `{` ir `}`. Skyrybos ženklas, kuris nėra `#`, `/`, `!` ar `>`, yra rezervuotas:
  jo žymė yra klaida. `{{` duoda paprastą `{`. Pavyzdys: `Įrašyk {{name} čia.` rodo `Įrašyk {name} čia.`

## Seanso ir BBS laukai

Kiti laukai rodo seanso ir BBS duomenis:

```
{bbs.name} {bbs.version}  {bbs.time}
Tavo adresas {session.ip}, ryšio trukmė {session.online}
Prisijungusių naudotojų {bbs.sessions}, veikia {bbs.uptime}
```

| Laukas | Reikšmė |
|---|---|
| `session.ip` | Naudotojo adresas. Pavyzdys: `203.0.113.7`. |
| `session.theme` | Seanso temos pasirinkimas: `dark`, `light` arba `auto` (6 užduotis). |
| `session.lang` | Seanso kalba, pavyzdžiui, `en` (7 užduotis). |
| `session.online` | Laikas nuo naudotojo prisijungimo: `hh:mm:ss`, nuo 24 valandų `<d>d hh:mm:ss`. |
| `bbs.name` | Lentelės `[bbs]` raktas `name` faile `bbs.toml`. Numatytoji reikšmė – `Kind Brothers BBS`. |
| `bbs.version` | `velourbbs` versija, pavyzdžiui, `v1.2.0`. Programa, sukompiliuota ne iš leidimo žymos (tag), duoda `dev`. |
| `bbs.sessions` | Kiek naudotojų dabar prisijungę, įskaitant šį naudotoją. |
| `bbs.time` | Serverio data ir laikas, laiko juostoje `[screen] zone`. |
| `bbs.uptime` | Laikas nuo `velourbbs serve` paleidimo, `session.online` forma. |

- BBS paima reikšmes, kai rodo ekraną. Reikšmė ekrane nesikeičia
  iki kito ekrano. Išimtis – gyvas laukas (žemiau).
- `[screen] time` faile `screens/screens.toml` nustato `bbs.time` formą pagal Go etaloninį laiką
  `2006-01-02 15:04:05`. Numatytoji reikšmė – `"2006-01-02 15:04:05"`. Pavyzdys formai `16.11.2026 15:04`:

```toml
[screen]
time = "02.01.2006 15:04"
```

- `[screen] zone` nustato `bbs.time` laiko juostą: `"Local"`, `"UTC"` arba juostos pavadinimą,
  pavyzdžiui, `"Europe/Berlin"`. Numatytoji reikšmė `"Local"` yra serverio laiko juosta.
  `MST` reikšmėje `time` rodo laiko juostą. Pavyzdys formai `14:05 UTC`:

```toml
[screen]
time = "15:04 MST"
zone = "UTC"
```

- Komanda `velourbbs preview` neturi seanso. Ji rodo `127.0.0.1`, `00:00:00`, `1` naudotoją
  ir laiką `2006-01-02 15:04:05` juostoje `[screen] zone`, o `"Local"` atveju – UTC. `--set` pakeičia reikšmę, pavyzdžiui,
  `--set bbs.sessions=12`.

Visos taisyklės: [session-fields](../../../specs/session-fields.md), 2–4 skyriai.

## Pseudonimai

`[fields]` faile `screens/screens.toml` suteikia laukui trumpą pavadinimą – „pseudonimą“:

```toml
[fields]
w = "term.cols"
```

Tada `{w}` rodo terminalo stulpelių skaičių. Trumpas pseudonimas praverčia ANSI mene (9 užduotis).

Ekranas gali turėti savo pseudonimus: `[fields]` savo priekinėje dalyje. Jie galioja ekrane,
jo fragmentuose ir jo mene. Ekrano pseudonimas pakeičia tokio pat pavadinimo
`screens/screens.toml` pseudonimą:

```
+++
[fields]
w = "term.rows"
+++
Eilučių: {w}
```

Šiame ekrane, jo fragmentuose ir jo mene `{w}` rodo terminalo eilučių skaičių.

Visos taisyklės: [bbt-format](../../../specs/bbt-format.md), 3.2, 3.3 ir 4 skyriai,
ir [screens-toml](../../../specs/screens-toml.md), 3.3 skyrius.

## Gyvi laukai

Laukas su požymiu `live` keičiasi, kol ekranas laukia klavišo:

```
Laikas {bbs.time live "15:04:05"}, ryšio trukmė {session.online live}
```

- Kas sekundę BBS įrašo naują reikšmę senos vietoje. Likusi ekrano dalis nesikeičia.
- Gyvi gali būti šie laukai: `bbs.time`, `session.online`, `bbs.uptime` ir `bbs.sessions`.
  Kitame lauke požymis neturi jokio poveikio.
- Gyvas laukas užima savo didžiausios reikšmės plotį, todėl tekstas po juo nepasislenka.
  Pavyzdžiui, `session.online` užima 14 langelių – `9999d 23:59:59` plotį.
- Po požymio `bbs.time` gali turėti savo formą dvigubose kabutėse, pavyzdžiui, `"15:04:05"`.
  Be jos laukas turi `[screen] time` formą.
- Pseudonimas gali turėti požymį. Tada `{clock}` yra gyvas laukas:

```toml
[fields]
clock = 'bbs.time live "15:04:05"'
```

- Žymė su pseudonimu negali turėti požymio: `{clock live}` yra klaida.
- Sąlyga, `{#if}` arba `when`, paima reikšmę, kai BBS rodo ekraną. Sąlyga nesikeičia.
- Gyvas laukas nesikeičia puslapyje su eilute „Toliau“ ir kol naudotojas rašo eilutę (8 užduotis).
- Komandos `velourbbs preview` laikrodis stovi vietoje. Sąraše `--keys` elementas `+1s` pastumia jį viena sekunde
  ir rodo ekraną po tos sekundės: `velourbbs preview main --keys +1s`.

Visos taisyklės: [live-screen](../../../specs/live-screen.md), 3, 4 ir 7 skyriai.

## Sąlygos

```
{#if term.colors == 16777216}16 mln.{#else}{term.colors}{/if}
```

Šis tekstas rodo `16 mln.`, kai terminalas rodo 16777216 spalvų. Kitaip jis rodo skaičių.

- `{#if <condition>}` rodo tekstą iki `{#else}` arba `{/if}`, kai sąlyga teisinga.
- `{#else}` rodo tekstą iki `{/if}`, kai sąlyga klaidinga. `{#else}` neprivalomas.
- Blokas gali apimti kelias eilutes ir turėti kitų blokų. Kiekvienam `{#if}` reikia savo `{/if}` tame pačiame faile.
- Eilutė su `{#if}`, `{#else}` arba `{/if}` neduoda ekrano eilutės, kai joje lieka tik tarpai.
- Operatoriai yra `==`, `!=`, `<`, `<=`, `>` ir `>=`.
  Skaitinis laukas priima skaičių ir bet kurį operatorių.
  Tekstinis laukas priima tekstą dvigubose kabutėse ir tik `==` arba `!=`: `term.type == "xterm"`.
- `and` ir `or` jungia sąlygas. `and` jungia stipriau. Skliaustų nėra.

Visos taisyklės: [bbt-format](../../../specs/bbt-format.md), 5 ir 6.2 skyriai.

## Punktai su sąlygomis

`when` suteikia punktui sąlygą. Kai sąlyga klaidinga, punktas neveikia.
`{#if item.<KEY>}` yra teisinga, kai punktas `<KEY>` veikia. Taip punkto tekstas ir jo klavišas sutampa:

```
+++
[[item]]
key = "W"
do = "gosub wide"
when = "term.cols >= 132"
[[item]]
key = "G"
do = "logoff"
+++
{#if item.W}W  platus ekranas{/if}
G  iki pasimatymo
```

Terminale su 132 ar daugiau stulpelių ekranas rodo `W  platus ekranas`, ir `W` veikia.
Siauresniame terminale ši eilutė neduoda ekrano eilutės, ir BBS nepaiso `W`.
`not item.W` yra teisinga, kai punktas neveikia.

Visos taisyklės: [bbt-format](../../../specs/bbt-format.md), 8.2 skyrius.

## Laukų sąrašas

`velourbbs fields` išvardija tavo BBS laukus. Jos išvesties dalis:

```
velourbbs fields
```

```text
term.type        string   -   -      xterm
term.cols        integer  3   -      80
term.colors      integer  8   -      256
term.background  string   5   -
session.lang     string   2   -      en
bbs.name         string   -   -      Kind Brothers BBS
```

- Kiekvienoje eilutėje yra kelias, tipas, plotis, tiekėjas (8 užduotis) ir pavyzdinė reikšmė.
- Plotis yra didžiausias langelių skaičius, kurį gali užimti reikšmė. `-` reiškia, kad komanda nenurodo pločio.
  Naudok plotį fiksuoto pločio laukui: `{term.cols}` reikia 3 langelių, `{term.background}` – 5.
- Pavyzdinė reikšmė yra `velourbbs preview` reikšmė.
- Pirmiausia komanda atlieka `velourbbs check` patikras. Po klaidos ji išveda tik klaidas.

Visos taisyklės: [screen-render](../../../specs/screen-render.md), 3.4 skyrius.

## Ekranas `term`

Sukurk `screens/term.bbt` su šiuo tekstu:

`examples/lamp/lt/screens/term.bbt`:
```
+++
[[item]]
key = "ESC"
do = "back"
[[item]]
key = "ENTER"
do = "back"
+++
{>_head}
       {!dim}Dydis{!}    {term.cols}×{term.rows}
       {!dim}Spalvos{!}  {#if term.colors == 16777216}16 mln.{#else}{term.colors}{/if}
       {!dim}Tipas{!}    {term.type}

       Taip tave mato stotis.
{~}
  {!line}{~─}{!}  

    {!key}Esc Enter{!}  {!dim}atgal{!}

```

Ekranas 80 stulpelių ir 24 eilučių terminale:

```text

  ────────────────────────────────────────────────────────────────────────────

       Lempa  ☼ BBS prie lempos šviesos

       Dydis    80×24
       Spalvos  256
       Tipas    xterm

       Taip tave mato stotis.










  ────────────────────────────────────────────────────────────────────────────

    Esc Enter  atgal
```

Pavadinimai „Dydis“, „Spalvos“ ir „Tipas“ pilkai rudi, stiliumi `dim`. Reikšmės kreminės, stiliumi `normal`.

## Punktas `T` ekrane `main`

Pridėk punktą `T` į failo `screens/main.bbt` priekinę dalį po punkto `A`:

```
[[item]]
key = "T"
do = "gosub term"
```

Kūne pridėk punkto `T` eilutę tarp punktų `A` ir `G` eilučių.
Tarpo eilutė duoda 2 tuščiomis eilutėmis mažiau, todėl baigiamoji eilutė lieka 19 eilutėje.
Pridėk `T` į klavišų eilutę:

```
         {!key}A{!}   Apie mus {!line}{~.}{!} {!dim}kas mes esame{!}  

         {!key}T{!}   Tavo terminalas {!line}{~.}{!} {!dim}dydis ir spalvos{!}  

         {!key}G{!}   Iki pasimatymo {!line}{~.}{!} {!dim}lempa palauks{!}  
```

```
    {!key}A T G{!}  {!dim}pasirinkti{!}
```

## Rezultatas

- `velourbbs check` nieko neišveda.
- `velourbbs preview main --keys T` rodo ekraną `term`.
- `velourbbs preview term` išveda eilutes `Dydis    80×24`, `Spalvos  256` ir `Tipas    xterm`.
- `velourbbs preview term --set term.type=ansi` išveda `Spalvos  16` ir `Tipas    ansi`.
- `velourbbs preview term --set term.colors=16777216` išveda `Spalvos  16 mln.`.
- Telnet kliente `T` rodo lango dydį. Pakeisk lango dydį,
  paspausk `Esc` ir vėl `T`: ekranas rodo naują dydį, o klavišų eilutė lieka apačioje.
  Kai eilučių mažiau nei 14, ekranas rodomas puslapiais (6 užduotis), ir klavišas tik rodo kitą puslapį.
