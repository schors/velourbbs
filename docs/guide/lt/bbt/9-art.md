# 9 užduotis. ANSI menas

„ANSI menas“ – tai paveikslėlis iš simbolių ir spalvų.
Šioje užduotyje tu rodai meną ekrane. Naudok naują BBS katalogą, pavyzdžiui, `myart`.
Failai yra [`examples/art/`](../../../../examples/art/). Įrašyk kiekvieną failą be `examples/art/`.
Šio pavyzdžio tekstai – anglų kalba.

1. Sukurk katalogą `myart` ir jame katalogą `screens`.
2. Nukopijuok failą `examples/art/screens/lamp.ans` iš pirminio kodo saugyklos
   į `myart/screens/lamp.ans`. Arba paprašyk failo sistemos operatoriaus.
   Šis vadovas failo nerodo: jame yra valdymo sekos.

## Meno įterpimas

```
{>lamp.ans}
```

- `{><name>.ans}` rodo meną `screens/<name>.ans`. Eilutėje neturi būti kito teksto.
- Kiekviena meno eilutė prasideda ties kairiuoju ekrano kraštu.
- Menas nekeičia po juo esančio teksto stiliaus.
- BBS nukerpa meną, kuris platesnis už ekraną.
  `velourbbs check` įspėja, kai menas platesnis už didžiausią ekrano plotį.

Visos taisyklės: [bbt-format](../../../specs/bbt-format.md), 3.4 ir 6.2 skyriai.

## Failai

`examples/art/bbs.toml`:
```
```

Failas tuščias: BBS turi numatytuosius nustatymus.

`examples/art/screens/screens.toml`:
```
[theme.default]
normal = "223/235"
key    = "215 bold"
```

Stilius `key` – 256 spalvų indeksas 215. Oranžinė 215 lieka šilta ir esant 16 spalvų: BBS pakeičia ją artimiausiu indeksu 0–15.

Ekranas `screens/main.bbt`. 10 užduotis prideda prie jo efektus:

```
+++
width = [40, 40]
fill = "spaces"
[[item]]
key = "G"
do = "logoff"
+++
{>lamp.ans}

               {!key}G{!}  goodbye
```

- Menas yra 40 stulpelių pločio. `width = [40, 40]` priekinėje dalyje daro šį ekraną
  40 stulpelių pločio.
- Raktai `width`, `align`, `valign`, `fill` ir `gutter` priekinėje dalyje nustato vieno ekrano išdėstymą.
  Failo `screens/screens.toml` lentelėje `[screen]` jie nustato visų ekranų išdėstymą.
- Čia `screens/screens.toml` neturi `[screen]`. Todėl ekranas turi numatytąjį `align = "center"`:
  80 stulpelių terminale BBS padeda ekraną per vidurį.

Ekranas:

```text

                ▄▄▄▄▄▄▄▄
             ▄████████████▄
          ▄██████████████████▄
       ░▒▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▒░
          ░░▒▒▒▒▒▒ ██ ▒▒▒▒▒▒░░
            ░░░░   ██   ░░░░
                   ██
               ▄▄▄▄██▄▄▄▄
            ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀

             lit for xterm

               G  goodbye
```

- Fonas tamsiai pilkas. Lempos gaubtas oranžinis su kreminiu viršumi.
  Šviesa po juo geltona ir tamsiai oranžinė.
- Kotas pilkai rudas. Pagrindas rudas.
- Žodžiai „lit for“ pilkai rudi. Lauko reikšmė „xterm“ kreminė.
  Klavišas `G` – oranžinis ir pusjuodis, stiliumi `key`.

## Ekrano padėtis

```
+++
width = [40, 40]
align = "left"
valign = "middle"
+++
```

- `align` – padėtis platesniame terminale: `"left"`, `"center"` (numatytoji) arba `"right"`.
- `valign` – ekrano, kuris turi mažiau eilučių nei terminalas, padėtis:
  `"top"` (numatytoji), `"middle"` arba `"bottom"`. Ekranas su tarpo eilute (2 užduotis)
  užpildo terminalą, todėl `valign` jo neveikia.

Ekranas 80 ir 132 stulpelių terminalams gali augti kartu su terminalu:

```
+++
width = [68, 108]
gutter = 6
+++
```

- `gutter` – langeliai, kuriuos ekranas palieka kiekvienoje terminalo pusėje: nuo 0 (numatytoji reikšmė) iki 100.
  Ekrano plotis – terminalo plotis minus 2 `gutter`, `width` ribose.
- Esant 80 stulpelių, ekranas yra 68 stulpelių pločio, su 6 tuščiais stulpeliais kiekvienoje pusėje.
  Esant 132 stulpeliams, ekranas yra 108 stulpelių pločio, per vidurį: po 12 tuščių stulpelių kiekvienoje pusėje.
- Tempimo žymė (2 užduotis) užpildo ekrano plotį esant bet kokiam dydžiui.
  Todėl `{~─}` duoda 68 langelių ploną liniją esant 80 stulpelių ir 108 langelių – esant 132 stulpeliams.

Visos taisyklės: [bbt-format](../../../specs/bbt-format.md), 6.1, 6.5 ir 6.6 skyriai.

## Laukas mene

`lamp.ans` 12 eilutėje yra tekstas `lit for {term.type    }`.
BBS randa lauką mene ir įrašo reikšmę į jo vietą.

- Laukas užima savo žymės langelius: čia 15 langelių. BBS nukerpa ilgesnę reikšmę.
- Tarpai prieš vardą lygiuoja reikšmę dešinėje. Kitaip reikšmė lygiuojama kairėje.
- Visi žymės langeliai turi turėti vieną fono spalvą.
  Visi jos simboliai, išskyrus tarpus, turi turėti vieną teksto spalvą.
  Kitaip `velourbbs check` praneša klaidą (11 užduotis).
- Reikšmė gauna pirmojo žymės simbolio spalvas.
- Mene gali būti tik laukai ir sritys (žemiau): be stiliaus žymių ir be įterpimų.
  Trumpas pseudonimas (5 užduotis) duoda trumpą lauką, pavyzdžiui, `{w}`.

Visos taisyklės: [art-format](../../../specs/art-format.md), 4 skyrius.

## Sąrašo eilutė mene

Sąrašo ekranas (8 užduotis) gali rodyti savo eilutes mene. Pavyzdys: ekranas `shelf`
rodo `content/log/` užrašus rėmelyje.

Menas `screens/shelf.ans`, 40 stulpelių pločio:

```text
╔══════════════════════════════════════╗
║ Prižiūrėtojo žurnalas psl. { n}/{c } ║
║ { row.num}  {row.title            }  ║
║                                      ║
║                                      ║
╚═[Enter]═[Space]═[Esc]════════════════╝
```

Pseudonimai `screens/screens.toml` lentelėje `[fields]` (5 užduotis):

```
[fields]
n = "list.page"
c = "list.pages"
```

Ekranas `screens/shelf.bbt`:

```
+++
width = [40, 40]
data = { from = "files", args = { dir = "log" } }
rows = 3
select = { do = "gosub reader", args = { file = "{row.path}" } }
[[item]]
key = "SPACE"
do = "page next"
when = "not list.last"
[[item]]
key = "ESC"
do = "back"
+++
{>shelf.ans}
```

Ekranas:

```text
╔══════════════════════════════════════╗
║ Prižiūrėtojo žurnalas psl.    1/3    ║
║          1  Pirmoji naktis           ║
║          2  Lietus ant stogo         ║
║          3  Svečias iš šiaurės       ║
╚═[Enter]═[Space]═[Esc]════════════════╝
```

3 eilutė – pasirinkta eilutė: ji juoda ant pilko fono, stilius `select`.
Kitos eilutės turi meno spalvas.

- Meno eilutė su laukais `row.*` – tai „sąrašo eilutė“. Mene yra ne daugiau kaip viena sąrašo eilutė.
- „Sąrašo stačiakampis“ prasideda pirmuoju pirmos žymės `row.*` langeliu
  ir baigiasi paskutiniuoju paskutinės žymės `row.*` langeliu. Jame yra `rows` meno eilučių, pradedant sąrašo eilute ir žemyn.
  Čia jis apima 3–37 stulpelius 3–5 eilutėse.
- Kiekviena stačiakampio eilutė rodo vieną sąrašo puslapio eilutę: sąrašo eilutės langelius
  su tos eilutės reikšmėmis.
- Kiekvienas stačiakampio langelis po sąrašo eilute turi būti tarpas. Kitaip `velourbbs check`
  praneša klaidą. Stačiakampis, kuris baigiasi žemiau meno, irgi yra klaida.
- Langeliai už stačiakampio lieka: čia tai rėmelis.
- Stačiakampio eilutė be puslapio eilutės – tarpai sąrašo eilutės spalvomis.
- Pasirinkta eilutė gauna temos stilių `select`. Numatytasis – `0/7`: juoda ant pilko.
  Nustatyk jį lentelėje `[theme.default]`, pavyzdžiui, `select = "235/215"`.
- Sritis negali būti sąrašo eilutėje, todėl `row.selected` mene neveikia.
- Sąrašo ekranui reikia sąrašo linijos (8 užduotis) arba sąrašo eilutės mene.

Visos taisyklės: [list-view](../../../specs/list-view.md), 9 skyrius.

## Sritis mene

„Sritis“ rodo meno eilutės dalį tik tada, kai jos sąlyga teisinga.
Pavyzdys: pakeisk apatinę `shelf.ans` eilutę šia eilute. Ji rodo `[Space]`
tik tada, kai punktas `SPACE` veikia:

```text
╚═[Esc]═{#if item.SPACE}[Space]{/if}═══╝
```

Apatinė eilutė 1 ir 2 puslapiuose ir paskutiniame 3 puslapyje:

```text
╚═[Esc]═[Space]                     ═══╝
╚═[Esc]═                            ═══╝
```

- Sritis eina nuo `{#if <sąlyga>}` iki `{/if}` vienoje meno eilutėje.
  Sąlyga – tai `{#if}` sąlyga (5 užduotis).
- Kai sąlyga teisinga, sritis rodo tekstą tarp dviejų žymių nuo pirmojo srities
  langelio. Likusi srities dalis – tarpai.
- Kai sąlyga klaidinga, sritis – tarpai.
- Tarpai turi pirmojo srities langelio spalvas. Langeliai po srities nepasislenka.
- Srities tekste gali būti laukų.
- Sritys nesideda viena į kitą ir neturi `{#else}`. Sritis negali būti sąrašo eilutėje.
- Kiekviena kita žymė, kuri prasideda `{#`, `{/`, `{!` ar `{>`, mene yra klaida,
  pavyzdžiui, `{!title}`.

Visos taisyklės: [art-format](../../../specs/art-format.md), 5 skyrius.

## Meno spalvos

Meno spalvos 0–15 – tai VGA spalvos: meno redaktorių spalvos.
Terminalas gali rodyti spalvas 0–15 kitomis reikšmėmis. Pavyzdžiui, Windows Terminal
rodo spalvą 1 kaip `c50f1f`, o ne kaip VGA `aa0000`.

- Prisijungiant BBS klausia terminalo jo spalvų 0–15.
- Kai jos skiriasi nuo VGA, BBS siunčia meno VGA reikšmes kaip RGB spalvas.
  256 spalvų terminalui BBS siunčia artimiausią indeksą 16–255.
- 16 spalvų terminalui ir tada, kai terminalas neatsako, BBS siunčia
  indeksus 0–15. Terminalas rodo juos savo spalvomis.
- Tema ir stiliaus žymės, pavyzdžiui, `{!14}`, visada duoda indeksus.
  Terminalas rodo juos savo spalvomis.

Visos taisyklės: [telnet-term](../../../specs/telnet-term.md), 7.3 skyrius.

## Menas mažiems terminalams

Ekranas gali rodyti kitą turinį siaurame terminale. Pavyzdys su 80 stulpelių menu:

```
+++
width = [40, 80]
[[item]]
key = "G"
do = "logoff"
+++
{#if term.cols < 80}
{>_main_small}
{#else}
{>main.ans}
{/if}
```

Terminale, siauresniame nei 80 stulpelių, ekranas rodo fragmentą `screens/_main_small.bbt`:
tekstą arba mažesnį meną.

Kai `screens/screens.toml` lentelėje `[screen]` yra `gutter`, pridėk `gutter = 0` į priekinę dalį.
Taip menas gauna visą terminalo plotį.

Visos taisyklės: [bbt-format](../../../specs/bbt-format.md), 6.1 skyrius.

## Naujas menas

BBS skaito meną tik „plokščia forma“. `velourbbs art import` paverčia meno redaktoriaus
failą plokščia forma.

1. Nupiešk meną viename iš šių redaktorių. Įrašyk jį su numatytosiomis parinktimis:
   - Moebius 1.0.29: Save (`.ans`);
   - icy_draw 0.6.2: Export į `.ans`;
   - PabloDraw 3.3.14-beta: Save as ANSI (`.ans`).
2. Laikyk šį „šaltinio failą“ už `screens/` ribų, pavyzdžiui, `art/logo.ans`.
3. Paversk jį:

```
velourbbs art import art/logo.ans screens/logo.ans
```

4. Įterpk meną su `{>logo.ans}`.

Moebius ir PabloDraw negali atverti plokščios formos. Norėdamas pakeisti meną, keisk šaltinio failą
ir paversk jį dar kartą.

Visos taisyklės: [art-import](../../../specs/art-import.md), 2 ir 3 skyriai.

## Rezultatas

- `velourbbs check` nieko neišveda.
- `velourbbs preview main` išveda ekraną, kaip aukščiau, su 20 fono stulpelių kiekvienoje pusėje.
- `velourbbs preview main --set term.type=ansi` išveda užrašą `lit for ansi`.
- Sustabdyk „Lempos“ serverį. Paleisk `velourbbs serve` kataloge `myart`.
  PuTTY rodo lempą lango viduryje.
