# 8 užduotis. Sąrašai

Šioje užduotyje tu pridedi prie „Lempos“ prižiūrėtojo žurnalą. Ekranas `log` rodo katalogo
`content/log/` užrašų sąrašą. Kiekvienam užrašui eilutės rašyti nereikia:
BBS skaito katalogą ir pati sudaro eilutes. Rodyklių klavišai pasirenka užrašą,
o Enter atveria jį 7 užduoties skaityklėje.

## Užrašai

Sukurk katalogą `content/log/` ir jame septynis užrašus:
`1-first-night.md`, `2-rain.md`, `3-north.md`, `4-wick.md`, `5-winter.md`,
`6-moths.md` ir `7-tea.md`. Kiekvienas užrašas – Markdown failas su 1 lygio antrašte.
Antraštė – užrašo pavadinimas.

`examples/lamp/lt/content/log/1-first-night.md`:
```
# Pirmoji naktis

Uždegiau lempą devintą. Modemas sudainavo savo trumpą dainą,
ir pirmasis svečias užėjo ketvirtį po dešimtos.

Kalbėjomės apie orą ir senus terminalus.
Vidurnaktį parašiau šį užrašą ir truputį pritildžiau dagtį.
```

Kiti užrašai sudaryti taip pat. Jų tekstai yra `examples/lamp/lt/content/log/`.
Užrašas `5-winter.md` ilgesnis: jo skaitymo laikas – 2 minutės.

## Sąrašo ekranas

„Sąrašo ekranas“ – ekranas su raktu `data`. Jis rodo „tiekėjo“ eilutes puslapiais.

| Raktas | Reikšmė |
|---|---|
| `data` | Tiekėjas: `from` ir `args`. |
| `rows` | Didžiausias puslapio eilučių skaičius, nuo 1 iki 100. Sąrašo ekrane jis privalomas. |

Tiekėjas `files` duoda vieno katalogo turinio failus:

```
data = { from = "files", args = { dir = "log" } }
rows = 5
```

- `dir` – kelias kataloge `content/`, kaip failo ekrano raktas `file` (7 užduotis).
  Jei katalogo nėra, tai klaida.
- Eilutė – kiekvienas katalogo failas. BBS neina į pakatalogius.
  Ji praleidžia vardą, kuris netinka kelio daliai, pavyzdžiui, `.keep`.
- Tvarka: pirma `README.md`, paskui „natūrali tvarka“. Joje skaitmenys lyginami kaip skaičiai:
  `2-rain.md` eina prieš `10-art.md`.
- Kiekvienas failas – turinio failas: jo klaidos yra `velourbbs check` klaidos (7 užduotis).
- BBS skaito eilutes paleidimo metu. Pridėjęs užrašą, iš naujo paleisk `velourbbs serve`.
- Sąrašo ekranas negali turėti rakto `file`.

Visos taisyklės: [list-view](../../../specs/list-view.md), 3 ir 4 skyriai.

## Sąrašo linija

„Sąrašo linija“ – šaltinio eilutė, kuri prasideda `{#each}` ir baigiasi `{/each}`.
Tekstas tarp jų – „eilutės tekstas“: BBS rodo jį vieną kartą kiekvienai puslapio eilutei.

```
{#each}         {!dim}{row.num}{!}   {row.title                             }{!dim}{row.minutes} min{/each}
```

- Sąrašo linija duoda lygiai vieną eilučių puslapį: po vieną kiekvienai puslapio eilutei,
  paskui tuščias eilutes. Puslapyje yra `rows` eilučių, o mažame terminale mažiau (skyrius „Puslapiai“). Todėl eilutės po sąrašu kiekviename puslapyje lieka vietoje.
- Eilutė nekeliama į kitą eilutę. BBS nukerpa ją ties ekrano pločiu.
- Eilutės tekste gali būti laukų, stiliaus žymių ir blokų `{#if}`.
  Jo blokai turi užsidaryti toje pačioje eilutėje. Jame negali būti įterpimo ar meno įterpimo.
- Kiekviena eilutė prasideda stiliumi, kuris buvo prieš sąrašo liniją.
  Po sąrašo linijos stilius tas pats, kuris buvo prieš ją.
- Sąrašo linija gali būti bloke ir fragmente. Ekranas turi vieną sąrašą:
  kiekviena sąrašo linija rodo tą patį puslapį.

Eilutės ir sąrašo laukai:

| Laukas | Reikšmė |
|---|---|
| `row.num` | Eilutės numeris sąraše, nuo 1. |
| `row.title` | Failo pavadinimas, kaip `file.title` (7 užduotis). |
| `row.minutes` | Failo skaitymo laikas. |
| `row.path` | Failo kelias kataloge `content/`, pavyzdžiui, `log/2-rain.md`. |
| `row.selected` | Teisinga pasirinktoje eilutėje (skyrius „Pasirinkimas“). Tik sąlygoje. |
| `list.page` | Dabartinio puslapio numeris, nuo 1. |
| `list.pages` | Puslapių skaičius. Tuščias sąrašas turi vieną puslapį. |
| `list.last` | Teisinga paskutiniame puslapyje. Tik sąlygoje. |
| `list.empty` | Teisinga, kai sąraše nėra eilučių. Tik sąlygoje. |

- Laukas `row.*` galioja tik eilutės tekste, `select` rakto `args` reikšmėje
  ir meno sąrašo eilutėje (9 užduotis).
- Laukas `list.*` galioja tik sąrašo ekrane: jo tekste, fragmentuose, mene ir `when`.

Visos taisyklės: [list-view](../../../specs/list-view.md), 5 ir 6 skyriai.

## Pasirinkimas

Raktas `select` duoda sąrašo ekranui „žymeklį“: vieną pasirinktą puslapio eilutę.
Enter vykdo `select` veiksmą su pasirinkta eilute:

```
select = { do = "gosub reader", args = { file = "{row.path}" } }
```

- `do` – veiksmas, kaip meniu punkto `do` (4 užduotis).
- `args` reikšmė – tekstas, lygiai vienas laukas `{row.<field>}`
  arba lygiai vienas laukas `{arg.<name>}` (skyrius „Sąrašas su argumentu“).
  Enter įdeda pasirinktos eilutės reikšmę į argumentą.
  Čia `{row.path}` perduoda 7 užduoties ekranui `reader` užrašo failą.
- BBS paleidimo metu įkelia kiekvienos eilutės failą ekranui `reader`.
  Klaida faile yra `velourbbs check` klaida.
- `Up` ir `Down` stumia žymeklį. Paskutinėje puslapio eilutėje `Down` rodo kitą puslapį.
  Pirmoje puslapio eilutėje `Up` rodo ankstesnį puslapį.
  Sąrašo galuose klavišai nieko nedaro. Enter tuščiame sąraše nieko nedaro.
- Ekrane su `select` punktas negali turėti klavišo `UP`, `DOWN` ar `ENTER`.
- Veiksmas `page` padeda žymeklį į pirmą naujo puslapio eilutę.
  `back` iš `reader` rodo tą patį puslapį ir tą pačią eilutę.
- Kai žymeklis juda tame pačiame puslapyje, BBS siunčia tik dvi pasikeitusias eilutes.

`row.selected` teisinga pasirinktoje eilutėje. Ekrano `log` eilutės tekstas rodo ženklą `☼`
stiliumi `orn` prieš pasirinktos eilutės numerį, o jos pavadinimą – stiliumi `title`:

```
{#each}{#if row.selected}       {!orn}☼{!} {#else}         {/if}{!dim}{row.num}{!}   {#if row.selected}{!title}{/if}{row.title                             }{!dim}{row.minutes} min{/each}
```

Abi pirmojo bloko šakos duoda po 9 langelius, todėl numeris lieka 10 stulpelyje.

Visos taisyklės: [list-view](../../../specs/list-view.md), 8 skyrius,
ir [menu-nav](../../../specs/menu-nav.md), 3.3 ir 5.3 skyriai.

## Ekranas `log`

`examples/lamp/lt/screens/log.bbt`:
```
+++
data = { from = "files", args = { dir = "log" } }
rows = 5
select = { do = "gosub reader", args = { file = "{row.path}" } }
[[item]]
key = "SPACE"
do = "page next"
when = "not list.last"
[[item]]
key = "B"
do = "page prev"
when = "list.page > 1"
[[item]]
key = "ESC"
do = "back"
+++
{>_head}
       {!title}Prižiūrėtojo žurnalas{!}
       {!dim}naktų užrašai  ∙  psl. {list.page}/{list.pages}{!}

{#each}{#if row.selected}       {!orn}☼{!} {#else}         {/if}{!dim}{row.num}{!}   {#if row.selected}{!title}{/if}{row.title                             }{!dim}{row.minutes} min{/each}
{~}
  {!line}{~─}{!}  

    {!key}↑↓{!}  {!dim}pasirinkti{!}   {!key}Enter{!}  {!dim}skaityti{!}   {#if item.SPACE}{!key}Space{!}  {!dim}toliau{!}   {/if}{#if item.B}{!key}B{!}  {!dim}ankst.{!}   {/if}{!key}Esc{!}  {!dim}atgal{!}

```

Ekranas:

```text

  ────────────────────────────────────────────────────────────────────────────

       Lempa  ☼ BBS prie lempos šviesos

       Prižiūrėtojo žurnalas
       naktų užrašai  ∙  psl. 1/2

       ☼ 1   Pirmoji naktis                          1 min
         2   Lietus ant stogo                        1 min
         3   Svečias iš šiaurės                      1 min
         4   Naujas dagtis                           1 min
         5   Ilgas žiemos vakaras                    2 min







  ────────────────────────────────────────────────────────────────────────────

    ↑↓  pasirinkti   Enter  skaityti   Space  toliau   Esc  atgal
```

- Sąrašo eilutės – 9–13 eilutės. Tarpo eilutė po sąrašo linija
  laiko ploną liniją 21 eilutėje.
- `{row.title                             }` – fiksuoto pločio laukas, 40 langelių (5 užduotis).
  Todėl skaitymo laikas prasideda 54 stulpelyje.

## Puslapiai

7 užduoties veiksmai `page` veikia ir sąrašo ekrane.
`when` ir `{#if item.<KEY>}` naudoja `list.last` ir `list.page`, kaip `file.last` ir `file.page`.

Sąraše yra 7 eilutės, o puslapyje – 5 eilutės. Todėl 2 puslapyje yra du užrašai ir trys tuščios eilutės.

Kai ekranas netelpa terminale, puslapyje yra mažiau eilučių.
BBS piešia 1 puslapį ir skaičiuoja jo eilutes. Kai ekranas už terminalą ilgesnis
`n` eilučių, puslapyje yra `rows - n` eilučių, bet ne mažiau kaip 1. BBS skaičiuoja eilutes
be tarpo eilutės. Esant 80x15, ekranas `log` turi 17 eilučių,
todėl puslapyje yra 3 eilutės, o sąraše – 3 puslapiai:

```text
       Prižiūrėtojo žurnalas
       naktų užrašai  ∙  psl. 1/3

       ☼ 1   Pirmoji naktis                          1 min
         2   Lietus ant stogo                        1 min
         3   Svečias iš šiaurės                      1 min
```

Tarpo eilutė duoda laisvas eilutes po puslapio dydžio skaičiavimo. Esant 80x15, ji neduoda eilučių.

Sąrašo ekranas neturi eilutės „Toliau“.

Tuščias sąrašas turi vieną puslapį su tuščiomis eilutėmis. Parodyk jam tekstą su `list.empty`:

```
{#if list.empty}       {!dim}Užrašų dar nėra.{!}{/if}
```

Visos taisyklės: [list-view](../../../specs/list-view.md), 7 skyrius,
ir [menu-nav](../../../specs/menu-nav.md), 5.3 skyrius.

## Ekranas `main`

Punktas `L` veda į ekraną `log`. Jo eilutė yra tarp `R` ir `T` eilučių.
Pašalink vieną tuščią eilutę virš `A` eilutės. Tarpo eilutė duoda viena tuščia eilute mažiau.

`examples/lamp/lt/screens/main.bbt`:
```
+++
[[item]]
key = "A"
do = "gosub about"
[[item]]
key = "R"
do = "gosub texts"
[[item]]
key = "L"
do = "gosub log"
[[item]]
key = "T"
do = "gosub term"
[[item]]
key = "G"
do = "logoff"
+++
{>_head}
       Labas vakaras. Čia tylu ir šilta.
       Prisitrauk kėdę prie lempos.

         {!key}A{!}   Apie mus {!line}{~.}{!} {!dim}kas mes esame{!}  

         {!key}R{!}   Skaitykla {!line}{~.}{!} {!dim}istorijos prie lempos{!}  

         {!key}L{!}   Prižiūrėtojo žurnalas {!line}{~.}{!} {!dim}naktų užrašai{!}  

         {!key}T{!}   Tavo terminalas {!line}{~.}{!} {!dim}dydis ir spalvos{!}  

         {!key}G{!}   Iki pasimatymo {!line}{~.}{!} {!dim}lempa palauks{!}  
{~}
       {!dim}Šviesa dega iki ryto.{!}

  {!line}{~─}{!}  

    {!key}A R L T G{!}  {!dim}pasirinkti{!}

```

## Sąrašas su argumentu

`args` reikšmė rakte `data` gali būti vienas ekrano argumentas: `{arg.<name>}`.
Taip vienas ekranas gali rodyti daugiau nei vieną katalogą:

```
data = { from = "files", args = { dir = "{arg.dir}" } }
```

```
[[item]]
key = "1"
do = "gosub notes"
args = { dir = "log" }
```

- BBS skaito kiekvienos `dir` reikšmės, kurią duoda `goto` ar `gosub`, eilutes.
- `velourbbs preview notes --set arg.dir=log` rodo šios reikšmės eilutes.

Ekranas gali perduoti savo argumentą kitam ekranui. Punkto ar `select`
`args` reikšmė gali būti lygiai vienas laukas `{arg.<name>}`. Čia `notes` perduoda savo `dir`
ir eilutės failą ekranui `note`:

```
select = { do = "gosub note", args = { dir = "{arg.dir}", file = "{row.path}" } }
```

- `{arg.dir}` duoda dabartinio ekrano argumento `dir` reikšmę.
  Po punkto `1` ekranas `note` gauna `dir = "log"`.
  Vardas `dir` – ekrano `notes` argumentas, kaip rakte `data`.
- Reikšmė gali pereiti per kelis ekranus. BBS seka kiekvieną `goto` ir `gosub`
  ir paleidimo metu įkelia kiekvienos reikšmės failus ir eilutes.
- Reikšmės klaida yra klaida ties tekstu ar lauku `{row.<field>}`, kuris ją duoda,
  net po kitų ekranų. Pavyzdžiui, `velourbbs check` praneša apie trūkstamą katalogą
  ties punkto `1` `dir = "log"`.
- Lauko `{row.<field>}` reikšmė negali pasiekti ekrano `data`, net per kitus ekranus.
  Todėl sąrašo eilutės niekada nepriklauso nuo kito sąrašo eilučių.
- Rakte `args` kiekvienas kitas laukas yra klaida: pseudonimas, laukas su pločiu.

Visos taisyklės: [bbt-format](../../../specs/bbt-format.md), 8.5 skyrius.

## Eilutė iš naudotojo

Punktas gali paprašyti vienos teksto eilutės prieš savo veiksmą. `ask` – raginimas,
o `{in}` rakte `args` perduoda eilutę kitam ekranui.
Šio pavyzdžio nėra „Lempoje“. Norėdamas jį išbandyti,
pridėk punktą į savo BBS failą `screens/main.bbt`:

```
[[item]]
key = "N"
ask = "Your name: "
ask_max = 20
do = "gosub hello"
args = { name = "{in}" }
```

Paskui sukurk ekraną `hello`. `screens/hello.bbt`:

```
+++
[[item]]
key = "ESC"
do = "back"
+++
Good evening, {arg.name}.
```

- Klavišas `N` rodo raginimą ir lauką paskutinėje puslapio eilutėje.
  Naudotojas įveda eilutę ir paspaudžia Enter. Tada `hello` gauna eilutę argumente `name`.
- Esc arba Enter su tuščia eilute atšaukia įvedimą. Puslapis lieka.
- `ask_max` – didžiausias eilutės simbolių skaičius. Be jo laukas užpildo likusią eilutės dalį.
- Punktui su `ask` reikia `goto` ar `gosub` ir `{in}` rakte `args`.
  `{in}` galioja tik punkte su `ask`.
- Raginimas – paprastas tekstas: žymė jame yra tekstas. Temos stiliai `more` ir `select`
  duoda raginimo ir lauko spalvas.
- Esant mažiausiam ekrano pločiui, raginimas turi palikti bent 2 langelius.
- Eilutė negali pasiekti `file` ar `data`, net per kitus ekranus:
  BBS įkelia failus ir eilutes paleidimo metu, o naudotojas gali įvesti bet kokią eilutę.

Išbandyk eilutę ekrane `main` su šiuo punktu: `velourbbs preview main --keys N,a,n,n,ENTER`. Išvestyje yra `ask`,
puslapis su raginimu, kiekvienas eilutės klavišas ir `line "ann"`.
Eilutės raidė išlaiko savo dydį.

Visos taisyklės: [bbt-format](../../../specs/bbt-format.md), 8.6 skyrius,
ir [menu-nav](../../../specs/menu-nav.md), 5.4 skyrius.

## Pavyzdiniai duomenys

`velourbbs fields` (5 užduotis) išveda ir eilučių laukus:

```text
row.num          integer  1   -      1
row.path         string   20  files  log/1-first-night.md
row.title        string   20  files  Pirmoji naktis
row.minutes      integer  1   files  1
```

Tiekėjo stulpelis nurodo lauko `row.*` tiekėją.
`row.title` plotis – ilgiausio užrašų pavadinimo plotis.

„Pavyzdžių rinkinys“ duoda `velourbbs preview` kitas eilutes ir kitas laukų reikšmes.
Taip gali išbandyti tuščią sąrašą ir ilgą pavadinimą be naujų užrašų.
Rinkiniai yra faile `sample.toml` BBS kataloge.
`velourbbs fields --sample` įrašo šį failą su rinkiniu `default`:
kiekvieno lauko, kurį priima `--set`, reikšmę ir vieną eilutę. Esamo failo komanda nepakeičia.

Pridėk į failą savo rinkinius:

```toml
[empty]
rows = []

[long]
fields = { "term.colors" = "16" }
rows = [
  { path = "log/2-rain.md", title = "Labai ilgas užrašo pavadinimas, kuris netelpa eilutėje", minutes = 120 },
  { path = "log/3-north.md", title = "Trumpas" },
]
```

- Kiekviena lentelė – vienas rinkinys. Rinkinio vardą sudaro mažosios raidės, skaitmenys, `_` ir `-`.
- `fields` duoda laukams reikšmes, kaip `--set`.
- `rows` pakeičia kiekvieno peržiūros sąrašo ekrano eilutes.
  Eilutė duoda tiekėjo laukus be `row.`. Trūkstamas laukas tuščias arba 0.
- Ekrano `log` `select` atveria `path` ekrane `reader`. Todėl kiekvienas `path` turi būti turinio failas.
- `velourbbs check` ir `velourbbs serve` neskaito `sample.toml`.

```
velourbbs preview log --sample empty
velourbbs preview log --sample long
```

Pirmoji komanda rodo vieną puslapį su tuščiomis eilutėmis. Antroji komanda rodo nukirptą ilgą pavadinimą
16 spalvų. `--set` pakeičia rinkinio reikšmę, nesvarbu, kokia parinkčių tvarka.

Visos taisyklės: [screen-render](../../../specs/screen-render.md), 3.2 ir 3.5 skyriai.

## Išbandyk puslapius

```
velourbbs preview log
```

Komanda išveda abu sąrašo puslapius, vieną po kito.
`--size 80x15` išveda tris mažo terminalo puslapius.

`--keys` stumia žymeklį, atveria užrašą ir verčia puslapius:

```
velourbbs preview main --keys L,DOWN,ENTER,ESC,SPACE,B,ESC
```

Žymeklio judesys išveda visą puslapį su žymekliu naujoje eilutėje.

## Rezultatas

- `velourbbs check` nieko neišveda.
- `velourbbs preview log` išveda du puslapius. 1 puslapis – ekranas aukščiau.
- Telnet kliente `L` rodo žurnalą. `Up` ir `Down` stumia žymeklį,
  Enter atveria užrašą, `Esc` grąžina į tą pačią eilutę.
  `Space` ir `B` verčia puslapius, `Esc` grąžina į meniu.
