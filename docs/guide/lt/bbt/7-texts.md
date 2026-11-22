# 7 užduotis. Tekstai

Šioje užduotyje pridedi „Lempai“ skaityklą. Ekranas `texts` rodo dviejų tekstų sąrašą.
Ekranas `reader` rodo tekstą puslapiais.

## Turinio failai

„Turinio failas“ – tai failas, kurį rodo ekranas.
Turinio failus dėk į BBS katalogo katalogą `content`.

- Kelias kataloge `content` gali turėti katalogų, pavyzdžiui, `xfer/src/crc16.c`.
  Kiekviena kelio dalis susideda iš lotyniškų raidžių, skaitmenų, `_`, `.` ir `-`.
  Ji prasideda raide, skaitmeniu arba `_`.
- Failas yra UTF-8 tekstas, ne didesnis nei 1 MB. Eilutė baigiasi LF arba CR LF.
  Valdymo simbolis, išskyrus TAB, yra klaida.
- Plėtinys nurodo failo „rūšį“. Raidžių dydis nesvarbus:

| Plėtinys | Rūšis |
|---|---|
| `.md`, `.markdown` | Markdown |
| `.txt` arba be plėtinio | Paprastas tekstas |
| Bet kuris kitas plėtinys | Pirminis kodas |

Sukurk katalogą `content`. Sukurk failus `content/lamp.md` ir `content/rules.txt`
pagal toliau esančius skyrius „Markdown“ ir „Paprastas tekstas“.

Visos taisyklės: [file-view](../../../specs/file-view.md), 2 skyrius.

## Ekranas `main`

Punktas `R` veda į ekraną `texts`.
Jo eilutė yra tarp `A` ir `T` eilučių. Tarpo eilutė duoda 2 tuščiomis eilutėmis mažiau.
Failas `screens/main.bbt` po šios užduoties (galutinį jo pavidalą rodo 8 užduotis):

```
+++
[[item]]
key = "A"
do = "gosub about"
[[item]]
key = "R"
do = "gosub texts"
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

         {!key}T{!}   Tavo terminalas {!line}{~.}{!} {!dim}dydis ir spalvos{!}  

         {!key}G{!}   Iki pasimatymo {!line}{~.}{!} {!dim}lempa palauks{!}  
{~}
       {!dim}Šviesa dega iki ryto.{!}

  {!line}{~─}{!}  

    {!key}A R T G{!}  {!dim}pasirinkti{!}

```

Ekranas:

```text

  ────────────────────────────────────────────────────────────────────────────

       Lempa  ☼ BBS prie lempos šviesos

       Labas vakaras. Čia tylu ir šilta.
       Prisitrauk kėdę prie lempos.


         A   Apie mus .......................................... kas mes esame

         R   Skaitykla ................................. istorijos prie lempos

         T   Tavo terminalas ................................ dydis ir spalvos

         G   Iki pasimatymo .................................... lempa palauks


       Šviesa dega iki ryto.

  ────────────────────────────────────────────────────────────────────────────

    A R T G  pasirinkti
```

## Ekranas `texts`

`args` perduoda argumentą `gosub` arba `goto` ekranui:

```
[[item]]
key = "1"
do = "gosub reader"
args = { file = "lamp.md" }
```

- Ekranas `reader` gauna argumentą `file` su reikšme `lamp.md`.
  Laukas `{arg.file}` duoda šią reikšmę.
- Argumento pavadinimas susideda iš mažųjų lotyniškų raidžių, skaitmenų ir `_` ir prasideda raide.
  Tekstinėje reikšmėje negali būti `{` ir valdymo simbolių.
  Reikšmė taip pat gali perduoti toliau dabartinio ekrano argumentą (8 užduotis).
- `gosub` arba `goto` perduoda lygiai tuos argumentus, kuriuos naudoja jo ekranas.
  Trūkstamas arba kitas argumentas yra klaida.

Visos taisyklės: [bbt-format](../../../specs/bbt-format.md), 8.5 skyrius.

`examples/lamp/lt/screens/texts.bbt`:
```
+++
[[item]]
key = "1"
do = "gosub reader"
args = { file = "lamp.md" }
[[item]]
key = "2"
do = "gosub reader"
args = { file = "rules.txt" }
[[item]]
key = "ESC"
do = "back"
+++
{>_head}
       Stoties tekstai.
       Išsirink ir skaityk prie lempos.


         {!key}1{!}   Kaip buvo uždegamos lempos {!line}{~.}{!} {!dim}trumpa istorija{!}  

         {!key}2{!}   Stoties taisyklės {!line}{~.}{!} {!dim}perskaityk vieną kartą{!}  
{~}
  {!line}{~─}{!}  

    {!key}1 2{!}  {!dim}pasirinkti{!}   {!key}Esc{!}  {!dim}atgal{!}

```

Ekranas:

```text

  ────────────────────────────────────────────────────────────────────────────

       Lempa  ☼ BBS prie lempos šviesos

       Stoties tekstai.
       Išsirink ir skaityk prie lempos.


         1   Kaip buvo uždegamos lempos ...................... trumpa istorija

         2   Stoties taisyklės ........................ perskaityk vieną kartą








  ────────────────────────────────────────────────────────────────────────────

    1 2  pasirinkti   Esc  atgal
```

## Failo ekranas `reader`

„Failo ekranas“ turi raktą `file` savo priekinėje dalyje.
`file` nurodo turinio failą: kelią kataloge `content` arba lygiai vieną žymę `{arg.<name>}`.

```
+++
file = "{arg.file}"
margin = [7, 2]
```

- `{>file}` yra „failo sritis“: teksto vieta. Jo eilutėje nėra kito teksto.
  Failo ekranas turi lygiai vieną `{>file}` savo paties kūne: ne fragmente ir ne `{#if}` viduje.
- `margin = [left, right]` nurodo tuščius langelius kairėje ir dešinėje nuo failo srities.
  `[7, 2]` duoda 8–78 stulpelius: „Lempos“ teksto stulpelį. Be `margin` tekstas prasideda 1 stulpelyje.
  `left + right` turi būti mažesnis už mažiausią ekrano plotį.
- Eilutės prieš `{>file}` yra „viršutinės eilutės“. Eilutės po jo yra „apatinės eilutės“.
  Failo sritis užima likusias terminalo eilutes.
  Esant 24 eilutėms, `reader` turi 8 viršutines ir 5 apatines eilutes, todėl failo sritis turi 11 eilučių.
- BBS suskaido tekstą į failo srities dydžio puslapius. Failo ekranas neturi eilutės „Toliau“:
  kitus puslapius rodo jo punktai.
- Paskutinė `reader.bbt` eilutė tuščia. Ji duoda tuščią 24 eilutę ir palieka apatinę liniją 21 eilutėje.
  Be jos failo sritis turi 12 eilučių, o apatinė linija pasislenka į 22 eilutę.
- Kai eilučių mažiau nei 14, BBS kiekvieną puslapį nukerpa ties paskutine terminalo eilute.
  Kai eilučių mažiau nei 13, klavišų eilutės nesimato.

Visos taisyklės: [file-view](../../../specs/file-view.md), 3, 3.1 ir 3.2 skyriai.

## Puslapių veiksmai

| `do` | Veiksmas |
|---|---|
| `page next` | Rodyti kitą puslapį. Paskutiniame puslapyje vėl rodyti paskutinį puslapį. |
| `page prev` | Rodyti ankstesnį puslapį. Pirmajame puslapyje vėl rodyti pirmąjį puslapį. |
| `page first` | Rodyti pirmąjį puslapį. |
| `page last` | Rodyti paskutinį puslapį. |

- Veiksmas `page` veikia tik failo ekrane ir sąrašo ekrane (8 užduotis).
- `when = "not file.last"`: `Space` veikia kiekviename puslapyje, išskyrus paskutinį.
- `when = "file.page > 1"`: `B` veikia nuo 2 puslapio.
- Klavišų eilutė rodo klavišą tik tada, kai jo punktas veikia: `{#if item.SPACE}...{/if}`.
  Todėl 1 puslapyje nėra `B`, o paskutiniame puslapyje nėra `Space`.

Visos taisyklės: [file-view](../../../specs/file-view.md), 3.3 skyrius,
ir [menu-nav](../../../specs/menu-nav.md), 5.1 skyrius.

## Failo laukai

| Laukas | Reikšmė |
|---|---|
| `file.title` | Priekinės dalies raktas `title`. Kitaip – pirmosios 1 lygio antraštės tekstas. Kitaip – failo vardas, pavyzdžiui, `rules.txt`. |
| `file.page` | Puslapio numeris, nuo 1. |
| `file.pages` | Puslapių skaičius. |
| `file.last` | Teisinga paskutiniame puslapyje. Naudok jį tik sąlygoje. |
| `file.section` | Paskutinės 1 ar 2 lygio antraštės puslapio viršuje arba virš jo tekstas. Kitaip – tuščia. |
| `file.minutes` | Skaitymo laikas minutėmis. |
| `meta.<key>` | Priekinės dalies raktas `<key>`. Kitaip – tuščia. |

- Šie laukai veikia tik failo ekrane.
- Skaitymo laikas yra žodžių skaičius, padalytas iš 200, suapvalintas į didesnę pusę, ir ne mažiau nei 1.
  Markdown faile BBS skaičiuoja teksto žodžius be priekinės dalies ir adresų.
  Paprasto teksto ar pirminio kodo faile BBS skaičiuoja eilučių, kurias rodo ekranas, žodžius.
- `reader` 7 eilutė rodo autorių tik tada, kai failas turi `author`:
  `{#if meta.author != ""}{meta.author}  ∙  {/if}`.
- Apatinė linija rodo puslapio numerį savo centre:
  `{~─} {!dim}{file.page}/{file.pages}{!line} {~─}`. Dvi tempimo žymės pasidalija laisvus langelius.
  Kai laisvų langelių skaičius nelyginis, dešinioji žymė gauna vienu langeliu daugiau.

Visos taisyklės: [file-view](../../../specs/file-view.md), 3.4 ir 3.5 skyriai.

`examples/lamp/lt/screens/reader.bbt`:
```
+++
file = "{arg.file}"
margin = [7, 2]
[[item]]
key = "SPACE"
do = "page next"
when = "not file.last"
[[item]]
key = "B"
do = "page prev"
when = "file.page > 1"
[[item]]
key = "ESC"
do = "back"
+++
{>_head}
       {!title}{file.title}{!}
       {!dim}{#if meta.author != ""}{meta.author}  ∙  {/if}{file.minutes} min{!}

{>file}

  {!line}{~─} {!dim}{file.page}/{file.pages}{!line} {~─}{!}  

    {#if item.SPACE}{!key}Space{!}  {!dim}toliau{!}   {/if}{#if item.B}{!key}B{!}  {!dim}ankst.{!}   {/if}{!key}Esc{!}  {!dim}sąrašas{!}

```

Ekranas su `lamp.md`, 1 puslapis:

```text

  ────────────────────────────────────────────────────────────────────────────

       Lempa  ☼ BBS prie lempos šviesos

       Kaip buvo uždegamos lempos
       Anna Lind  ∙  1 min

       Žmonės uždegė pirmąsias lempas gerokai anksčiau, nei parašė pirmuosius
       žodžius. Tūkstančius metų lempa skleidė mažą, ramią šviesą.

       Aliejus ir dagtis

         ∙ molio dubenėlis aliejui;
         ∙ lininis arba medvilninis dagtis;
         ∙ ranka, kuri kas vakarą pakirpdavo dagtį.

       │ Gera lempa nerūksta. Ji prašo tik truputėlio priežiūros.


  ─────────────────────────────────── 1/3 ────────────────────────────────────

    Space  toliau   Esc  sąrašas

```

- Pavadinimas – šviesiai persikinis ir paryškintas, stilius `title`. 7 eilutė – pilkai ruda, stilius `dim`.
- Antraštė „Aliejus ir dagtis“ – oranžinė. Žymekliai `∙` irgi oranžiniai.

## Markdown

BBS rodo Markdown: CommonMark su GitHub lentelėmis.

| Markdown | Vaizdas | Stilius |
|---|---|---|
| Pastraipa | Tekstas. Jis keliamas į kitą eilutę ties failo srities kraštu. | `normal` |
| `# Antraštė` | Tekstas. | `md_h1` |
| `## Antraštė` | Tekstas. | `md_h2` |
| Nuo `### Antraštė` iki `###### Antraštė` | Tekstas. | `md_h3` |
| `**paryškinta**` | Tekstas. | `md_strong` |
| `*pabrėžta*` | Tekstas. | `md_em` |
| `` `kodas` `` | Tekstas. | `md_code` |
| `- punktas` | `∙ punktas`. Vidinis sąrašas turi `-`. | `md_mark`, vidinis `md_frame` |
| `1. punktas` | `1. punktas` | `md_mark`, vidinis `md_frame` |
| `> citata` | `│ citata` | brūkšnys `md_mark`, tekstas `md_em` |
| Kodo blokas | `│ kodas`. Ilga eilutė nekeliama: ją baigia `...`. | brūkšnys `md_frame`, kodas `md_code` |
| `---` | `∙   ∙   ∙` | `md_mark` |
| Lentelė | Stulpeliai su 2 tarpais tarp jų. Po antraštės eilute – linija `─`. | antraštės eilutė `md_strong`, linija `md_frame` |
| `[tekstas](https://...)` | Tekstas ir išnašos numeris, pavyzdžiui, `[1]`. | `md_link`, numeris `md_dim` |
| `![tekstas](picture.png)` | Tekstas, kaip nuoroda. | `md_link` |
| `<https://...>` | Adresas. | `md_link` |
| HTML | Tekstas taip, kaip parašytas. | `normal` |

- Kiti GitHub plėtiniai rodomi kaip tekstas, pavyzdžiui, `~~tekstas~~`.
- Nuoroda su adresu, kuris turi schemą, pavyzdžiui, `https:`, turi išnašą.
  Po paskutinio bloko BBS rodo išnašas: `[1] <adresas>` stiliumi `md_dim`.
  Nuoroda be schemos, pavyzdžiui, į kitą Markdown failą, išnašos neturi.
  Adresas tarp `<` ir `>` išnašos neturi.
- 1 lygio antraštė pradeda naują puslapį. Todėl `lamp.md` neturi 1 lygio antraštės.
- Antraštė lieka viename puslapyje su pirmąja kito bloko eilute.
- Puslapis neprasideda tuščia eilute.

Markdown failo „priekinė dalis“ duoda laukus `meta.*`. Rašyk ją TOML kalba tarp `+++` eilučių:

```
+++
title = "Kaip buvo uždegamos lempos"
author = "Anna Lind"
+++
```

- Priekinė dalis prasideda pirmoje failo eilutėje.
- Kiekvienas raktas su viena reikšme duoda lauką `meta.<key>`. Raktas `title` dar duoda ir `file.title`.
- Pirmoji eilutė `+++` be pabaigos eilutės yra klaida (11 užduotis).
- YAML priekinė dalis tarp `---` eilučių yra klaida.

Visos taisyklės: [file-view](../../../specs/file-view.md), 5 skyrius.

`examples/lamp/lt/content/lamp.md`:
```
+++
title = "Kaip buvo uždegamos lempos"
author = "Anna Lind"
+++

Žmonės uždegė pirmąsias lempas gerokai anksčiau, nei parašė pirmuosius žodžius.
Tūkstančius metų lempa skleidė mažą, *ramią* šviesą.

## Aliejus ir dagtis

- molio dubenėlis aliejui;
- lininis arba medvilninis dagtis;
- ranka, kuri kas vakarą pakirpdavo dagtį.

> Gera lempa nerūksta. Ji prašo tik truputėlio priežiūros.

## Stiklinis gaubtas

XVIII a. 9-ajame dešimtmetyje Aimé Argand uždėjo stiklinį gaubtą ant apvalaus dagčio.
Oras ėjo iš apačios, ir liepsna tapo **daug ryškesnė**.
Po septyniasdešimties metų [žibalinė lempa](https://lt.wikipedia.org/wiki/Žibalinė_lempa)
atkeliavo į kiekvienus namus.

| Lempa | Kuras | Nuo |
|---|---|---|
| Molinė lempa | alyvuogių aliejus | prieš 5000 metų |
| Argand lempa | banginių taukai | XVIII a. 9-ojo dešimtmečio |
| Žibalinė lempa | žibalas | XIX a. 6-ojo dešimtmečio |

---

## Prie lempos šį vakarą

Elektros šviesa atėjo vėliau, ir senosios lempos iškeliavo į lentyną.
Bet lempa lange vis dar sako tą patį:
kažkas yra namie, ir durys atviros.

Prisitrauk kėdę. Šviesa dega.
```

2 puslapis, failo sritis:

```text
       Stiklinis gaubtas

       XVIII a. 9-ajame dešimtmetyje Aimé Argand uždėjo stiklinį gaubtą ant
       apvalaus dagčio. Oras ėjo iš apačios, ir liepsna tapo daug ryškesnė. Po
       septyniasdešimties metų žibalinė lempa[1] atkeliavo į kiekvienus namus.

       Lempa           Kuras              Nuo
       ─────────────────────────────────────────────────────────────
       Molinė lempa    alyvuogių aliejus  prieš 5000 metų
       Argand lempa    banginių taukai    XVIII a. 9-ojo dešimtmečio
       Žibalinė lempa  žibalas            XIX a. 6-ojo dešimtmečio
```

3 puslapis, failo sritis:

```text
       ∙   ∙   ∙

       Prie lempos šį vakarą

       Elektros šviesa atėjo vėliau, ir senosios lempos iškeliavo į lentyną.
       Bet lempa lange vis dar sako tą patį: kažkas yra namie, ir durys
       atviros.

       Prisitrauk kėdę. Šviesa dega.

       [1] https://lt.wikipedia.org/wiki/Žibalinė_lempa
```

## Teksto stiliai

Stiliai `md_*` nurodo teksto spalvas. „Lempoje“ juos turi `screens/screens.toml` (6 užduotis):

- `md_h1` ir `md_strong` kaip `title`: šviesiai persikiniai ir paryškinti;
- `md_h2` ir `md_mark` kaip `orn`: oranžiniai. `md_mark` yra sąrašo žymeklis, citatos brūkšnys ir skirtukas;
- `md_h3`, `md_em` ir `md_link` šviesiai rudi. Nuoroda pabraukta: tai ne klavišas;
- `md_code` šviesiai alyvuogių spalvos;
- `md_frame` kaip `line`, o `md_dim` kaip `dim`.

Tema be `md_*` stilių naudoja numatytuosius stilius: baltas antraštes, geltoną sąrašo žymeklį.
„Lempos“ `[theme.256]` ir `[theme.16]` kiekvienam gyliui nurodo `md_h2`, `md_h3`, `md_em`, `md_code` ir `md_link`.
Kiti `md_*` stiliai yra nuorodos (6 užduotis): jie ima savo pavadinimo stilių tame pačiame gylyje.
Esant 16 spalvų, pabrėžimas atrodo kaip tekstas.

Visos taisyklės: [file-view](../../../specs/file-view.md), 4 skyrius.

## Paprastas tekstas

Paprasto teksto failas rodo kiekvieną eilutę tokią, kokia ji yra. Ilga eilutė keliama į kitą eilutę. Tekstas rodomas stiliumi `normal`.

`examples/lamp/lt/content/rules.txt`:
```
1. Būk malonus. Kiekvienas čia – svečias.
2. Rašyk taip, kaip kalbi su draugu prie stalo.
3. Jokių reklamų ir šlamšto.
4. Neskelbk kitų žodžių be autoriaus sutikimo.
5. Saugok stoties svečių paslaptis.
6. Matai problemą – pasakyk sistemos operatoriui.

Ačiū, ir sveiki atvykę į „Lempą“.
```

Ekranas su `rules.txt`:

```text

  ────────────────────────────────────────────────────────────────────────────

       Lempa  ☼ BBS prie lempos šviesos

       rules.txt
       1 min

       1. Būk malonus. Kiekvienas čia – svečias.
       2. Rašyk taip, kaip kalbi su draugu prie stalo.
       3. Jokių reklamų ir šlamšto.
       4. Neskelbk kitų žodžių be autoriaus sutikimo.
       5. Saugok stoties svečių paslaptis.
       6. Matai problemą – pasakyk sistemos operatoriui.

       Ačiū, ir sveiki atvykę į „Lempą“.




  ─────────────────────────────────── 1/1 ────────────────────────────────────

    Esc  sąrašas

```

Pavadinimas yra failo vardas: paprasto teksto failas neturi priekinės dalies.

## Pirminio kodo failas

Pirminio kodo failas rodo kiekvieną eilutę stiliumi `md_code`. Ilga eilutė nekeliama: ją baigia `...`.
TAB perkelia iki kito 8 langelių kartotinio.

Ekranas gali rodyti fiksuotą failą arba jo dalį:

```
+++
file = "src/lamp.c"
lines = [4, 8]
numbers = true
margin = [7, 2]
[[item]]
key = "ESC"
do = "back"
+++
{>file}
```

- `lines = [first, last]` rodo eilutes nuo `first` iki `last`.
- `numbers = true` rodo eilučių numerius stiliumi `md_dim`.
- `lines` ir `numbers` veikia tik paprasto teksto ar pirminio kodo failui.

Failo sritis:

```text
       4 int main(void)
       5 {
       6         puts("The light stays on.");
       7         return 0;
       8 }
```

Visos taisyklės: [file-view](../../../specs/file-view.md), 6 skyrius.

## BBS dviem kalbomis

BBS gali rodyti savo ekranus ir tekstus daugiau nei viena kalba.
Naudotojas pasirenka kalbą. BBS išlaiko pasirinkimą iki seanso pabaigos.

Kiekviena `screens/screens.toml` lentelė `[lang.<code>]` prideda kalbą.
Kodas – dvi arba trys mažosios raidės, pavyzdžiui, `en` arba `ru`.
`[screen] lang` yra kalba seanso pradžioje. Numatytoji reikšmė – `en`.

```toml
[screen]
lang = "en"

[lang.en]

[lang.ru]
more = "    дальше ∙ любая клавиша"
```

Kalbos lentelės `more` nurodo tos kalbos eilutę „Toliau“.
Be jo eilutė „Toliau“ yra `[screen] more`, paskui `-- More --`.

Kalbos failus dėk į katalogą, kurio pavadinimas yra jos kodas:

```
screens/
  screens.toml
  main.bbt        kalbos pasirinkimas
  _logo.bbt       vienas logotipas visoms kalboms
  en/menu.bbt     meniu anglų kalba
  ru/menu.bbt     meniu rusų kalba
content/
  en/lamp.md
  ru/lamp.md
```

BBS pirmiausia ieško failo seanso kalba, paskui bendrame kataloge.
Pavyzdys: kalba `ru` ekranas `menu` yra `screens/ru/menu.bbt`, kitaip – `screens/menu.bbt`.
Ta pati taisyklė galioja fragmentui, menui ir turinio failui, taip pat bendrame ekrane.
Failas be teksto, pavyzdžiui, logotipas, lieka `screens/` vieną kartą visoms kalboms.

Bendras ekranas `main` klausia kalbos:

```
+++
[[item]]
key = "E"
do = "gosub en/menu"
[[item]]
key = "R"
do = "gosub ru/menu"
+++
{>_logo}
E  English   R  Русский
```

`gosub ru/menu` nustato kalbą `ru` ir rodo ekraną `menu` ta kalba.
Navigacijos dėklas laiko tik pavadinimą `menu`. Todėl `Esc` su `back` meniu vėl rodo `main`.

Veiksmas `lang <code>` pakeičia kalbą bet kuriame ekrane.
Tada BBS rodo tą patį ekraną nauja kalba. Failo ekranas išlaiko savo puslapį.

```
[[item]]
key = "L"
do = "lang en"
```

Laukas `session.lang` duoda seanso kalbą:

```
{#if session.lang == "ru"}Язык: русский{#else}Language: English{/if}
```

- Kiekviena kalba turi turėti tuos pačius ekranus su tais pačiais argumentais.
  `velourbbs check` tikrina kiekvienos kalbos ekranus.
  Pavyzdys: `screens/ru/news.bbt:1:1: no screen "news" (language "en")`.
- Bendro failo klaida, kurią randa tik kai kurios kalbos, baigiasi kalba.
  Pavyzdys: `screens/about.bbt:6:6: item "M": do: unknown screen "menu" (language "de")`.
- `velourbbs preview menu --set session.lang=ru` rodo ekraną `menu` rusų kalba.

Visos taisyklės: [session-language](../../../specs/session-language.md), 2–6 skyriai,
ir [menu-nav](../../../specs/menu-nav.md), 4 skyrius.

## Patikrink puslapius

`reader` reikia argumento `file`. Perduok jį su `--set`:

```
velourbbs preview reader --set arg.file=lamp.md
```

Komanda išveda visų teksto puslapių langelių išklotines vieną po kitos.
Be `--set` komanda `velourbbs preview reader` išveda klaidą `missing argument "file" of screen "reader"`.

`--keys` rodo puslapius. Pradėk nuo `texts`: `1` atidaro tekstą ekrane `reader`, `Esc` grąžina į `texts`:

```
velourbbs preview texts --keys 1,SPACE,SPACE,B,ESC
```

```
screen texts
page 1 80x24 fill spaces
...
key 1
screen reader
page 1 80x24 fill spaces
...
key SPACE
page 2 80x24 fill spaces
...
key SPACE
page 3 80x24 fill spaces
...
key B
page 2 80x24 fill spaces
...
key ESC
screen texts
page 1 80x24 fill spaces
...
```

## Rezultatas

- `velourbbs check` nieko neišveda.
- `velourbbs preview reader --set arg.file=lamp.md` išveda tris puslapius. 1 puslapis yra ekranas aukščiau, su `1/3` apatinėje linijoje.
- `velourbbs preview reader --set arg.file=lamp.md --keys SPACE,SPACE` išveda 3 puslapį. Jo klavišų eilutė yra `B  ankst.   Esc  sąrašas`.
- Telnet kliente `R` rodo sąrašą, `1` rodo tekstą, `Space` ir `B` verčia puslapius,
  `Esc` grąžina į sąrašą.
