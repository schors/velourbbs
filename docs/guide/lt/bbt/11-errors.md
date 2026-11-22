# 11 užduotis. Klaidos

`velourbbs check` randa klaidas prieš BBS paleidimą.
Kiekviena eilutė nurodo failą, eilutės numerį, stulpelį ir pranešimą (1 užduotis).
Mene eilutė ir stulpelis yra meno eilutė ir stulpelis.
Ši užduotis rodo dažnas klaidas ir jų taisymus.

## Klaida faile `bbs.toml` ar `screens/screens.toml`

```
screens/screens.toml:10:10: unexpected character U+0023 '#' at start of value
```

`screens/screens.toml` 10 eilutė yra `title  = #f7c7a3 bold`.
Tekstinei reikšmei TOML kalboje reikia dvigubų kabučių: `title  = "#f7c7a3 bold"`.

Ekranai naudoja `bbs.toml` ir `screens/screens.toml` nustatymus.
Todėl po klaidos šiuose failuose `velourbbs check` sustoja prieš ekranus
ir neišveda ekranų klaidų.
Ištaisyk klaidas ir vėl paleisk `velourbbs check`. Tada ji patikrina ekranus.

## `unknown style`

```
screens/about.bbt:10:8: unknown style "titel"
```

Stiliaus žymė `{!titel}` nurodo stilių, kurio temoje nėra.
Ištaisyk pavadinimą arba pridėk stilių į `[theme.default]`.

## `token has no "}"`

```
screens/about.bbt:10:30: token has no "}"
```

`{` pradeda žymę (5 užduotis). Uždaryk žymę ženklu `}` arba rašyk `{{`, jei reikia paprasto `{`.

## `unknown field`

```
screens/about.bbt:11:15: unknown field "term.colours"
screens/about.bbt:14:12: unknown field "note"
screens/term.bbt:14:67: unknown field "item.S"
```

- Pavadinimas nėra nei laukas, nei pseudonimas (5 užduotis). Ištaisyk pavadinimą.
- Tekstui riestiniuose skliaustuose rašyk `{{note}`. Mene `{{` nepadeda:
  pakeisk tekstą meno redaktoriuje ir vėl paversk meną.
- `item.S` reikia meniu punkto su klavišu `S` tame pačiame ekrane.

## `{#if} has no {/if}`

```
screens/about.bbt:12:8: {#if} has no {/if}
```

Pridėk `{/if}`. Blokas baigiasi tame faile, kuriame prasideda.

## `control character U+0009`

```
screens/about.bbt:13:1: control character U+0009
```

Eilutėje yra TAB. Pakeisk jį tarpais. Nustatyk redaktorių, kad klavišas Tab rašytų tarpus.

## Sąlyga

```
screens/term.bbt:14:8: expected an integer or a string, found "xterm"
screens/term.bbt:14:38: integer field "term.colors" needs an operator
```

- `{#if term.type == xterm}`: tekstą rašyk dvigubose kabutėse: `term.type == "xterm"`.
- `{#if term.colors = 256}`: lygink su `==`.

## `item: must be an array of tables`

```
screens/main.bbt:2:2: item: must be an array of tables
screens/main.bbt:5:2: table item already exists
screens/main.bbt:5:2: table item already exists as an array of tables
screens/main.bbt:5:3: key item already exists as a table, but should be an array table
```

Priekinėje dalyje yra `[item]`. Rašyk `[[item]]`.
Pirmasis pranešimas ateina iš ekrano su vienu meniu punktu. Kiti pranešimai nurodo lentelės eilutę.

## `unknown screen`

```
screens/main.bbt:13:6: item "T": do: unknown screen "term"
```

`gosub term` reikia failo `screens/term.bbt`. Sukurk failą arba ištaisyk pavadinimą.

## `screen has no items`

```
screens/x.bbt:1:1: screen has no items
```

Kiekvienam ekranui, išskyrus `goodbye` ir `shutdown`, reikia bent vieno meniu punkto. Pridėk punktą.

## `a user cannot reach this screen`

```
screens/old.bbt:1:1: warning: a user cannot reach this screen
```

`velourbbs check` seka kiekvieną `goto`, `gosub`, `select` ir `lang` nuo pirmojo ekrano `main`.
Joks kelias neveda į ekraną `old`.
Pridėk punktą, kuris atveria ekraną, pavyzdžiui, `do = "gosub old"` ekrane `main`.
Arba ištrink failą ir kiekvieną punktą, kuris jį atveria.

- Ekranai, kurie atveria tik vienas kitą, irgi gauna įspėjimą.
- Punktas su `when` yra kelias.
- `goodbye` ir `shutdown` negauna įspėjimo.
- Dviejų kalbų BBS (7 užduotis) patikra seka kiekvieną kalbą.
  `screens/en/menu.bbt` turi punktą `gosub news`, o `screens/ru/menu.bbt` jo neturi.
  Tada bendras failas `screens/news.bbt` gauna įspėjimą kalbai `ru`:
  `screens/news.bbt:1:1: warning: a user cannot reach this screen (language "ru")`.
  Pridėk punktą `gosub news` į `screens/ru/menu.bbt`.
- Įspėjimas nesustabdo BBS.

## `content file ... does not exist`

```
screens/code.bbt:2:8: file: content file "src/lamp.c" does not exist
screens/texts.bbt:5:17: item "1": args: file: content file "lmap.md" does not exist
```

- Ekranas `code` laiko 7 užduoties pavyzdį („Pirminio kodo failas“).
  Jo raktas `file` nurodo `src/lamp.c`, o `content/src/lamp.c` neegzistuoja.
  Padėk failą šiuo keliu į `content/` arba ištaisyk kelią.
- Punktas `1` perduoda `gosub reader` failą `lmap.md` per `args`, o failo nėra kataloge `content/`.
  Ištaisyk pavadinimą `args` reikšmėje.

## `directory ... does not exist`

```
screens/log.bbt:2:41: data: args: dir: directory "lgo" does not exist
```

Sąrašo ekrano `log` raktas `data` nurodo katalogą `lgo` (8 užduotis),
o `content/lgo/` neegzistuoja. Ištaisyk pavadinimą arba sukurk katalogą.

## `front matter: no end line`

```
content/lamp.md:1:1: front matter: no end line "+++"
```

Pirmoji failo eilutė yra `+++`, o jokia kita eilutė nėra `+++`.
Pridėk eilutę `+++` po paskutinio priekinės dalies rakto (7 užduotis).

## `not in the flat form`

```
screens/logo.ans:1:1: not in the flat form, convert it with velourbbs art import
```

Menas yra meno redaktoriaus failas. Paversk jį (9 užduotis).

## Laukas mene keičia spalvą

```
screens/size.ans:1:7: token "{w}" changes the foreground: 11 at col 7, 14 at col 9
```

Žymės simboliai turi dvi spalvas: 11 – 7 stulpelyje ir 14 – 9 stulpelyje.
Meno redaktoriuje nupiešk visą žymę viena spalva. Tada vėl paversk meną.

## Išdėstymas esant terminalo dydžiui

`velourbbs check --size <cols>x<rows>` atvaizduoja kiekvieną ekraną šiuo dydžiu ir įspėja apie išdėstymo klaidas.
Nurodyk `--size` kiekvienam terminalo dydžiui, kurio tikiesi:

```
velourbbs check --size 80x24 --size 132x43
```

Esant šiems dydžiams, „Lempa“ neduoda įspėjimų. Jos ekranai turi `width = [80, 80]`,
todėl 40 stulpelių terminalas nukerpa jų eilutes:

```
screens/about.bbt:11:1: warning: at 40x24: the row is 44 cells wide, the screen shows 40
screens/_head.bbt:2:1: warning: screen "main": at 40x24: the row is 78 cells wide, the screen shows 40
```

Įspėjimai ir jų taisymai:

| Įspėjimas | Problema | Taisymas |
|---|---|---|
| `the line wraps to 2 rows` | Teksto eilutė platesnė už ekraną. | Sutrumpink eilutę. |
| `31 rows, the screen gets a More row` | Ekranas turi daugiau eilučių nei terminalas. | Pašalink eilutes arba padalyk ekraną. |
| `25 rows, the rows after 24 are cut` | Failo ekranas ar sąrašo ekranas netelpa terminale. | Pašalink eilutes virš failo srities ar sąrašo arba po jais. |
| `the row is 43 cells wide, the screen shows 40` | BBS nukerpa eilutę. | Sutrumpink eilutę arba sumažink `width` reikšmę `min`. |
| `art "logo.ans" is 100 cells wide, the screen shows 80` | BBS nukerpa meną. | Mažame terminale rodyk mažesnį meną (9 užduotis). |

- Įspėjimas fragmente prasideda ekranu, kuris jį įterpia: `screen "main": `.
- Įspėjimas nesustabdo BBS.

## Rezultatas

1. „Lempos“ faile `screens/about.bbt`, 10 eilutėje, įrašyk `{!titel}` prieš `„Lempa“`.
2. Paleisk `velourbbs check`. Ji išveda:

```
screens/about.bbt:10:8: unknown style "titel"
```

3. Pašalink `{!titel}`. `velourbbs check` nieko neišveda.

Visos taisyklės: [screen-render](../../../specs/screen-render.md), 3.1 ir 3.6 skyriai,
[menu-nav](../../../specs/menu-nav.md), 6 skyrius,
[file-view](../../../specs/file-view.md), 7 skyrius,
ir [list-view](../../../specs/list-view.md), 10 skyrius.
