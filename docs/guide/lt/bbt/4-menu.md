# 4 užduotis. Meniu su sparčiaisiais klavišais

## Punktai

Kiekvienas `[[item]]` priekinėje dalyje yra vienas meniu punktas:

```
[[item]]
key = "A"
do = "gosub about"
```

- `key` yra klavišas: lotyniška raidė, skaitmuo arba klavišo pavadinimas, pavyzdžiui, `ESC` ar `ENTER`.
  Raidžių dydis nesvarbus. Rusiška raidė veikia kaip lotyniška raidė tame pačiame klaviše: `ф` yra `A`.
- `do` yra veiksmas: `gosub <screen>` atidaro ekraną, `back` grįžta į ekraną,
  kuris jį atidarė, `logoff` baigia seansą. Skyrius „Navigacijos dėklas“ aprašo
  kitus veiksmus `goto` ir `top`.
- BBS nerodo punktų. Punktų tekstą ir klavišų eilutę parašyk kūne.
  Klavišų eilutė rodo visus ekrano klavišus.

Trumpoji forma rašo kiekvieną punktą vienoje eilutėje:

```
item = [
  { key = "A", do = "gosub about" },
  { key = "G", do = "logoff" },
]
```

Viename ekrane naudok vieną formą.

Visos taisyklės: [bbt-format](../../../specs/bbt-format.md), 8 skyrius,
ir [menu-nav](../../../specs/menu-nav.md), 3–6 skyriai.

## Ekranas `main`

Meniu turi du punktus: `A` veda į ekraną `about`, `G` baigia seansą.
5 užduotis prideda punktą `T`, 7 užduotis – punktą `R`, 8 užduotis – punktą `L`.

Parašyk `screens/main.bbt` su šiuo tekstu:

```
+++
[[item]]
key = "A"
do = "gosub about"
[[item]]
key = "G"
do = "logoff"
+++
{>_head}
       Labas vakaras. Čia tylu ir šilta.
       Prisitrauk kėdę prie lempos.


         {!key}A{!}   Apie mus {!line}{~.}{!} {!dim}kas mes esame{!}  

         {!key}G{!}   Iki pasimatymo {!line}{~.}{!} {!dim}lempa palauks{!}  
{~}
       {!dim}Šviesa dega iki ryto.{!}

  {!line}{~─}{!}  

    {!key}A G{!}  {!dim}pasirinkti{!}

```

- `{~.}` tarp punkto pavadinimo ir pastabos yra „taškų linija“ (2 užduotis). Ji užpildo langelius
  tarp jų taškais, rudu stiliumi `line`.
- Kiekviena punkto eilutė baigiasi 2 tarpais. Todėl kiekviena pastaba baigiasi 78 stulpelyje, po plonos linijos pabaiga.

Ekranas:

```text

  ────────────────────────────────────────────────────────────────────────────

       Lempa  ☼ BBS prie lempos šviesos

       Labas vakaras. Čia tylu ir šilta.
       Prisitrauk kėdę prie lempos.


         A   Apie mus .......................................... kas mes esame

         G   Iki pasimatymo .................................... lempa palauks






       Šviesa dega iki ryto.

  ────────────────────────────────────────────────────────────────────────────

    A G  pasirinkti
```

Klavišai oranžiniai, stiliumi `key`. Pastabos dešinėje ir žodis „pasirinkti“ pilkai rudi, stiliumi `dim`.

## Ekranas `about`

Klavišai `Esc` ir `Enter` grįžta į ekraną, kuris atidarė `about`:

```
[[item]]
key = "ESC"
do = "back"
```

`examples/lamp/lt/screens/about.bbt`:
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
       „Lempa“ – maža stotis draugams.
       Užeik, apsidairyk, pabūk kiek ilgiau.

       Stotis veikia su Velour BBS.
       Jos ekranai – paprasti tekstiniai failai.
{~}
  {!line}{~─}{!}  

    {!key}Esc Enter{!}  {!dim}atgal{!}

```

Ekranas:

```text

  ────────────────────────────────────────────────────────────────────────────

       Lempa  ☼ BBS prie lempos šviesos

       „Lempa“ – maža stotis draugams.
       Užeik, apsidairyk, pabūk kiek ilgiau.

       Stotis veikia su Velour BBS.
       Jos ekranai – paprasti tekstiniai failai.










  ────────────────────────────────────────────────────────────────────────────

    Esc Enter  atgal
```

Klavišai `Esc Enter` oranžiniai, stiliumi `key`. Žodis „atgal“ pilkai rudas, stiliumi `dim`.

## Navigacijos dėklas

BBS laiko atidarytus seanso ekranus dėkle. Šie veiksmai keičia dėklą:

| `do` | Veiksmas |
|---|---|
| `gosub <screen>` | Atidaro ekraną virš dabartinio ekrano. |
| `back` | Uždaro dabartinį ekraną ir vėl rodo ekraną po juo. Pirmajame ekrane vėl rodo tą patį ekraną. |
| `top` | Uždaro visus ekranus virš pirmojo ekrano ir rodo pirmąjį ekraną. |
| `goto <screen>` | Rodo ekraną dabartinio ekrano vietoje. Tada `back` grįžta į ekraną po juo. |

- Ekranas, kurį rodo `back` ir `top`, išlaiko savo argumentus (7 užduotis) ir teksto puslapį.
- Dėkle telpa ne daugiau kaip 16 ekranų. `gosub` 16-ajame ekrane pašalina antrąjį ekraną.
  Pirmasis ekranas lieka.

Pavyzdys: `main` atidaro `texts` su `gosub`, `texts` atidaro `reader` su `gosub`.
Ekrane `reader` `back` rodo `texts`, o `top` rodo `main`.

Naudok `gosub`, kad eitum į gilesnį ekraną, ir `back`, kad grįžtum. Tada į ekraną galima patekti
skirtingais keliais, ir jam nereikia žinoti ankstesnio ekrano.

Visos taisyklės: [menu-nav](../../../specs/menu-nav.md), 4.1 skyrius.

## Ekranas `goodbye`

Po `logoff` BBS rodo ekraną `goodbye` ir baigia seansą.
`goodbye` neturi priekinės dalies ir punktų. Jis neturi klavišų eilutės.

```
       {!title}Iki greito.{!}
       Lempa degs tau.
```

`examples/lamp/lt/screens/goodbye.bbt`:
```
{>_head}
       {!title}Iki greito.{!}
       Lempa degs tau.
{~}
  {!line}{~─}{!}  



```

3 tuščios eilutės failo pabaigoje laiko ploną liniją 21 eilutėje.
Ekranas turi 24 eilutes. 22–24 eilutės tuščios:

```text

  ────────────────────────────────────────────────────────────────────────────

       Lempa  ☼ BBS prie lempos šviesos

       Iki greito.
       Lempa degs tau.













  ────────────────────────────────────────────────────────────────────────────
```

„Iki greito.“ šviesiai persikinės spalvos ir pusjuodis, stiliumi `title`.

## Ekranas `shutdown`

Gavusi SIGINT arba SIGTERM, komanda `velourbbs serve` sustoja. Ji rodo ekraną `shutdown`
kiekvienam atviram seansui ir baigia seansą. Pasakyk svečiams, kad BBS paleidžiama iš naujo:

```
       {!title}Lempa užges minutei.{!}
       Užsuk vėl netrukus.
```

Dvi eilutės:

```text
       Lempa užges minutei.
       Užsuk vėl netrukus.
```

„Lempa užges minutei.“ šviesiai persikinės spalvos ir pusjuodis, stiliumi `title`.

- Ekranui `shutdown` galioja ekrano `goodbye` taisyklės: jokių punktų ir jokių argumentų.
  Tai ne failo ekranas: jis neturi rakto `file` (7 užduotis).
- BBS rodo tik pirmąjį `shutdown` puslapį. BBS nelaukia klavišo.
- Be `screens/shutdown.bbt` BBS baigia seansus be teksto.
- Po 5 s BBS uždaro kiekvieną ryšį, kuris dar atviras.

`velourbbs preview shutdown` rodo ekraną.

Visos taisyklės: [menu-nav](../../../specs/menu-nav.md), 5.2 skyrius,
ir [telnet-term](../../../specs/telnet-term.md), 10.2 skyrius.

## Išbandyk klavišus

`velourbbs preview <screen> --keys <keys>` paspaudžia klavišus ir rodo kiekvieną ekraną.
Atskirk klavišus kableliais:

```
velourbbs preview main --keys A,ESC,G
```

Išvestyje prieš kiekvieną langelių išklotinę yra įvykio eilutė:

```
screen main
page 1 80x24 fill spaces
...
key A
screen about
page 1 80x24 fill spaces
...
key ESC
screen main
page 1 80x24 fill spaces
...
key G
logoff
screen goodbye
page 1 80x24 fill spaces
...
```

Visos taisyklės: [menu-nav](../../../specs/menu-nav.md), 7 skyrius.

## Rezultatas

- `velourbbs check` nieko neišveda.
- `velourbbs preview main --keys A,ESC,G` išveda įvykio eilutes, kaip parodyta aukščiau.
- `velourbbs preview goodbye` išveda 24 eilutes, kaip parodyta aukščiau: 22–24 eilutės tuščios.
- Telnet kliente `A` rodo `about`, `Esc` grįžta atgal, `G` rodo `goodbye` ir baigia seansą.
