# 10 užduotis. Efektai

„Efektas“ keičia ekrano langelius, kol ekranas laukia klavišo.
Šioje užduotyje tu pridedi efektus prie 9 užduoties lempos: lempos šviesa virpa,
aplink ją mirga žvaigždės, tamsoje atsiranda ir dingsta mažos kibirkštys, šalia klavišo
dega dvi mažos liepsnelės, gaubtas sužiba, o per lempą vieną kartą perbėga ryški juosta.
Naudok 9 užduoties katalogą `myart`.

## Lentelė `[[effect]]`

Kiekvienas efektas yra viena lentelė `[[effect]]` priekinėje dalyje:

```
[[effect]]
do = "twinkle"
at = [37, 3]
chars = "*+·"
```

- `do` nurodo efektą: `twinkle`, `sparkle`, `candle`, `glint`, `sweep` arba `lightning`.
- `at` yra „vieta“: langeliai, kuriuose efektas veikia (žr. toliau).
- Ekranas turi ne daugiau kaip 16 efektų. Efektai taikomi failo tvarka:
  efektas veikia langelius po aukščiau esančių efektų.
- Ekranai `goodbye` ir `shutdown` nelaukia klavišo, todėl negali turėti efektų.
- Efektai neveikia puslapyje su eilute „Toliau“ ir kol naudotojas rašo eilutę (8 užduotis).
- Kiekvieną kartą, kai BBS parodo ekraną, efektai prasideda iš naujo.

Visos taisyklės: [live-screen](../../../specs/live-screen.md), 5 ir 6.1 skyriai.

## Vieta

`at` turi vieną iš trijų formų:

| Forma | Pavyzdys | Langeliai |
|---|---|---|
| Taškas `[col, row]` | `[61, 3]` | Vienas langelis. |
| Stačiakampis `[col, row, width, height]` | `[1, 1, 80, 6]` | Stačiakampis nuo taško: `width` stulpelių į dešinę ir `height` eilučių žemyn. |
| Taškų masyvas | `[[61, 3], [70, 5]]` | Kiekvienas taškas yra vieta su savo laiku. |

- Su `in = "<name>.ans"` koordinatės yra mene: 1 stulpelis ir 1 eilutė yra
  pirmasis meno langelis. Vieta juda kartu su menu. Kai tekstas virš meno persikelia
  į daugiau eilučių, efektas lieka tuose pačiuose meno langeliuose.
- Su `in = "_<name>"` koordinatės yra fragmento `{>_<name>}` eilutėse:
  1 stulpelis yra kairysis ekrano kraštas, o 1 eilutė yra pirmoji fragmento eilutė.
  Vieta juda kartu su fragmentu ir lieka jo eilutėse. Pavyzdys: `in = "_logo"`
  leidžia `sweep` perbėgti per tekstinį logotipą temos spalvomis.
- Be `in` koordinatės yra puslapyje. 1 stulpelis yra kairysis ekrano kraštas,
  1 eilutė yra pirmoji puslapio eilutė. Persikeliantis tekstas pastumia žemiau esančias eilutes, todėl naudok
  šią formą tik eilutėms, kurios nepersikelia: galvutei, kampui ar klavišų eilutei.
- Neigiama koordinatė skaičiuojama nuo galo: `-1` yra paskutinis meno, fragmento
  arba ekrano ir puslapio stulpelis ar eilutė. Pavyzdys: `[-1, -1]` yra apatinio dešiniojo kampo langelis.
- Gyvas laukas (5 užduotis) ir platus simbolis, pavyzdžiui, `語`, nepriklauso vietai.
  Efektas niekada jų nekeičia.
- Kai ekranas nerodo langelio, langelis nepriklauso vietai: pavyzdžiui,
  meno langelis, nukirptas siauro terminalo krašte.

`velourbbs check` duoda klaidą, kai ekranas neįterpia `in` meno
arba neįterpia jo fragmento, arba kai vietos langelis yra už meno ribų.

Visos taisyklės: [live-screen](../../../specs/live-screen.md), 6.2 skyrius.

## Laikas

| Raktas | Numatytoji reikšmė | Reikšmė |
|---|---|---|
| `delay` | `0` | Sekundės iki laiko skaičiavimo pradžios: 0–3600. |
| `delay_spread` | `0` | Atsitiktinė `delay` dalis: delsa yra `delay` plius arba minus iki `delay_spread`. |
| `every` | Priklauso nuo efekto. | Sekundės tarp dviejų pakeitimų: 0.05–3600. |
| `spread` | Pusė `every`. | Atsitiktinė `every` dalis: kiekvienas tarpas yra `every` plius arba minus iki `spread`. |
| `times` | Nėra: be pabaigos. | Pakeitimų skaičius: 1–10000. Tada efektas sustoja. |

- Reikšmė yra sekundžių skaičius, pavyzdžiui, `0.5`.
- Pirmasis pakeitimas ateina praėjus `delay` ir vienam tarpui po parodymo.
  Pavyzdys: `delay = 1` ir `every = 1` su `spread = 0` duoda pirmąjį pakeitimą po 2 sekundžių.
- Su `spread = 0` kiekvienas pakeitimas ateina lygiai po `every` sekundžių po ankstesniojo.
- `every - spread` turi būti ne mažiau kaip 0.05. `delay - delay_spread` turi būti ne mažiau kaip 0.
- Kiekvienas taškų masyvo taškas gauna savo atsitiktinę delsą. Pavyzdys: `delay = 2`
  ir `delay_spread = 1` paleidžia kiekvieną tašką po 1–3 sekundžių ir dar vieno tarpo.
- Pavyzdys: `times = 3` duoda tris pakeitimus. Po jų efektas nedaro naujų pakeitimų:
  `twinkle` ar `candle` išlaiko savo paskutinį pakeitimą, o kitų efektų įvykiai baigiasi įprastai.
- Pakeitimai atsitiktiniai. Kiekvienas seansas turi savo atsitiktines reikšmes.

Visos taisyklės: [live-screen](../../../specs/live-screen.md), 6.3 skyrius.

## Efektas `twinkle`

Mirgėjimas: žvaigždė, kuri visada matoma ir keičia savo simbolį.

| Raktas | Numatytoji reikšmė | Reikšmė |
|---|---|---|
| `at` | Privalomas. | Taškas, stačiakampis arba taškų masyvas. |
| `chars` | Privalomas. | 1–32 vieno langelio simboliai, iš jų 2 ar daugiau skirtingų. |
| `colors` | Nėra. | 1–16 spalvų: stilius, pavyzdžiui, `"215"` ar `"#ffcc00"`, arba temos stiliaus pavadinimas. |
| `every` | `1` | |
| `share` | `0.1` | Stačiakampio žvaigždžių dalis, kuri keičiasi kiekvieną kartą: daugiau nei 0, ne daugiau kaip 1. |

- Taškas rodo žvaigždę nuo pradžios: savo simbolį, kai jis yra `chars`, kitaip pirmąjį
  `chars` simbolį. Kiekvieno pakeitimo metu žvaigždė gauna kitą `chars` simbolį.
- Stačiakampyje žvaigždės yra langeliai su `chars` simboliu.
  Kiekvieno pakeitimo metu `share` jų dalis gauna kitą `chars` simbolį.
- Su `colors` kiekvienas naujas simbolis gauna atsitiktinę `colors` spalvą.
  Kitaip jis išlaiko langelio spalvą. Fonas lieka.

## Efektas `candle`

Žvakė: liepsna, kuri plazda.

| Raktas | Numatytoji reikšmė | Reikšmė |
|---|---|---|
| `at` | Privalomas. | Taškas, stačiakampis arba taškų masyvas. |
| `chars` | Privalomas. | 1–32 vieno langelio simboliai. |
| `colors` | Privalomas. | 1–16 spalvų, kaip `twinkle`. |
| `every` | `0.2` | |
| `share` | `0.3` | Stačiakampio langelių dalis, kuri keičiasi kiekvieną kartą. |

- Nuo pradžios kiekvienas vietos langelis rodo pirmąjį `chars` simbolį
  pirmąja `colors` spalva.
- Kiekvieno pakeitimo metu taškas gauna atsitiktinį `chars` simbolį ir atsitiktinę `colors` spalvą.
  Stačiakampyje taip daro `share` langelių dalis.

Visos taisyklės: [live-screen](../../../specs/live-screen.md), 6.4 ir 6.6 skyriai.

## Efektas `sparkle`

Žybsniai: žvaigždė, kuri atsiranda ir dingsta.

| Raktas | Numatytoji reikšmė | Reikšmė |
|---|---|---|
| `at` | Nėra: visas menas ar puslapis. | Taškas, stačiakampis arba taškų masyvas. |
| `chars` | Privalomas. | 1–32 vieno langelio simboliai. |
| `colors` | Nėra. | 1–16 spalvų, kaip `twinkle`. |
| `over` | `false` | `true`: žvaigždė gali uždengti bet kurį langelį. |
| `count` | `1` | Stačiakampio žvaigždės kiekvieną kartą: 1–100. Ne su tašku. |
| `time` | `2` | Sekundės, kiek žvaigždė rodoma: 0.05–60. |
| `every` | `1` | |

- Kiekvieną kartą taškas rodo atsitiktinį `chars` simbolį `time` sekundžių.
  Stačiakampyje taip daro `count` atsitiktinių langelių.
- Su `over = false` žvaigždė atsiranda tik tuščiame langelyje: tarpe.
  Todėl menas ir tekstas lieka. Su `over = true` žvaigždė gali uždengti bet kurį vietos langelį.
- Langelis su žvaigžde negauna antros to paties efekto žvaigždės.
- Su `colors` kiekviena žvaigždė gauna atsitiktinę `colors` spalvą. Kitaip ji išlaiko langelio spalvą.
  Fonas lieka.

Visos taisyklės: [live-screen](../../../specs/live-screen.md), 6.5 skyrius.

## Ryškumas

`glint` ir `sweep` padaro langelį ryškų ir išlaiko jo simbolį.
Jie pakeičia simbolio spalvą iki spalvos `peak` ir atgal.
Fonas lieka.

- `peak` reikšmė `"#ffffff"` duoda baltą. `peak` reikšmė `"230"` duoda kreminę. Temos stiliaus pavadinimas
  duoda to stiliaus spalvą.
- Langelis su spalva `default` gauna `normal` spalvą.
- Terminale su 16 spalvų tarpinių spalvų mažai.
  Su 256 spalvomis ir su true color perėjimai sklandūs.

Visos taisyklės: [live-screen](../../../specs/live-screen.md), 6.3 skyrius.

## Efektas `glint`

Sužibimas: langelis tampa ryškus ir užgęsta.

| Raktas | Numatytoji reikšmė | Reikšmė |
|---|---|---|
| `at` | Nėra: visas menas ar puslapis. | Taškas, stačiakampis arba taškų masyvas. |
| `peak` | `#ffffff` | Ryškiausios akimirkos spalva: stilius arba temos stiliaus pavadinimas. |
| `time` | `0.6` | Sekundės nuo ryškiausios akimirkos iki pabaigos: 0.05–10. |
| `every` | `3` | |

- Kiekvieną kartą vienas atsitiktinis vietos langelis tampa ryškus. Jis užgęsta per `time` sekundžių.
- Tarpas nesužiba. Sužibti gali tik langelis su simboliu.
- Taškas visada sužiba tame pačiame langelyje.

## Efektas `sweep`

Bėganti šviesa: ryški dėmė bėga iš kairės į dešinę per vietos simbolius.

| Raktas | Numatytoji reikšmė | Reikšmė |
|---|---|---|
| `at` | Nėra: visas menas ar puslapis. | Stačiakampis. |
| `delay` | `0.5` | Sekundės iki laiko skaičiavimo pradžios, kaip kituose efektuose. |
| `every` | Nėra: vienas perbėgimas. | Sekundės nuo vieno perbėgimo pradžios iki kito perbėgimo pradžios. |
| `spread` | `0` | Tik su `every`. |
| `times` | Nėra: be pabaigos. | Perbėgimų skaičius. Tik su `every`. |
| `peak` | `#ffffff` | Kaip `glint`. |
| `time` | `0.6` | Sužibimo kiekvienos eilutės gale sekundės. |
| `width` | `3` | Dėmės pusplotis langeliais: 1–20. |
| `speed` | `80` | Stulpeliai per sekundę: 1–1000. |

- Dėmė bėga per visas vietos eilutes tuo pačiu metu. Per kelias eilutes
  ji yra juosta.
- Tarpas netampa ryškus. Todėl juosta parodo simbolių formą.
- Kai dėmė pasiekia paskutinį eilutės simbolį, tas simbolis sužiba `time` sekundžių.
- Pirmasis perbėgimas ateina po `delay` sekundžių nuo parodymo, su `delay_spread` kaip kituose efektuose.
- Be `every` juosta perbėga vieną kartą.
- Su `every` kiti perbėgimai ateina po `every` sekundžių nuo ankstesnio perbėgimo pradžios.
  Be `times` perbėgimai neturi pabaigos.
  Pavyzdys: `delay = 1`, `every = 5` ir `times = 3` duoda perbėgimus po 1, 6 ir 11 sekundžių.

Visos taisyklės: [live-screen](../../../specs/live-screen.md), 6.7 ir 6.8 skyriai.

## Efektas `lightning`

Žaibas: žaibo kirtis, kuris trumpam blyksteli.

| Raktas | Numatytoji reikšmė | Reikšmė |
|---|---|---|
| `at` | Nėra: visas menas ar puslapis. | Stačiakampis. |
| `chars` | `╲│╱` | 3 simboliai: žingsniai žemyn į dešinę, tiesiai žemyn ir žemyn į kairę. |
| `peak` | `#ffffff` | Žaibo spalva: stilius arba temos stiliaus pavadinimas. |
| `over` | `false` | `true`: žaibas gali uždengti bet kurį langelį. |
| `time` | `0.3` | Sekundės, kiek žaibas rodomas: 0.05–10. |
| `every` | `30` | |

- Žaibas prasideda atsitiktiniame stačiakampio viršutinės eilutės stulpelyje. Jis eina žemyn
  iki apatinės stačiakampio pusės ar toliau, atsitiktiniais žingsniais į kairę ir į dešinę.
  Jis turi 1–3 šakas.
- Su `over = false` žaibas piešia tik tuščiuose langeliuose: menas ir tekstas jį uždengia.
  Su `over = true` jis uždengia bet kurį vietos langelį.
- Žaibas išlaiko kiekvieno langelio foną. Jis neturi paryškinimo ir jokių kitų atributų.
- ASCII simboliams rašyk `chars = "\\|/"`: TOML kalboje `\\` yra vienas `\`.

Pavyzdys: žaibas danguje virš teksto, maždaug kas 20 sekundžių, puslapyje su 80 stulpelių:

```
[[effect]]
do = "lightning"
at = [1, 1, 80, 8]
every = 20
```

Visos taisyklės: [live-screen](../../../specs/live-screen.md), 6.9 skyrius.

## Failas

Pakeisk `screens/main.bbt`:

`examples/art/screens/main.bbt`:
```
+++
width = [40, 40]
fill = "spaces"
[[effect]]
do = "twinkle"
in = "lamp.ans"
at = [8, 5, 26, 3]
chars = "░▒"
share = 0.2
[[effect]]
do = "twinkle"
in = "lamp.ans"
at = [[4, 2], [37, 3], [2, 9], [39, 10]]
chars = "·+*"
colors = ["223", "215"]
every = 1.5
[[effect]]
do = "sparkle"
in = "lamp.ans"
chars = "·."
colors = ["137", "223"]
count = 2
time = 1.5
[[effect]]
do = "candle"
at = [[13, -1], [28, -1]]
chars = "▲▴"
colors = ["208", "214", "220"]
[[effect]]
do = "glint"
in = "lamp.ans"
at = [11, 2, 20, 3]
every = 2
[[effect]]
do = "sweep"
in = "lamp.ans"
[[item]]
key = "G"
do = "logoff"
+++
{>lamp.ans}

               {!key}G{!}  goodbye
```

- Pirmasis `twinkle` yra šviesa po gaubtu: meno stačiakampis, 5–7 eilutės.
  Jo žvaigždės yra langeliai `░` ir `▒`. Maždaug kas sekundę penktadalis jų pasikeičia.
- Antrasis `twinkle` duoda keturias žvaigždes tamsiame meno fone, kreminės ir oranžinės spalvos.
- `sparkle` veikia visą meną. Maždaug kas sekundę atsiranda dvi kibirkštys 1,5 sekundės.
  Jos atsiranda tik tuščiuose langeliuose, todėl lieka tamsiame fone ir neuždengia lempos.
- `candle` duoda dvi liepsneles paskutinėje puslapio eilutėje, šalia klavišo `G`, oranžinės ir geltonos spalvos.
  Čia puslapio koordinatės tinka: ekranas yra 40 stulpelių pločio, todėl jo eilutės nepersikelia.
- `glint` veikia gaubtą: meno stačiakampis, 2–4 eilutės.
  Maždaug kas 2 sekundes vienas gaubto langelis tampa baltas ir užgęsta.
- `sweep` perbėga vieną kartą per visą lempą, praėjus pusei sekundės po parodymo.
  Jis yra paskutinis, todėl padaro ryškias ir aukščiau esančių efektų žvaigždes bei šviesą.

## Rezultatas

- `velourbbs check` nieko neišveda.
- `velourbbs preview main` duoda 9 užduoties ekraną: peržiūra rodo ekraną
  prieš efektus.
- `velourbbs preview main --keys +1s` parodo ekraną po vienos sekundės: žvaigždes, kibirkštis,
  liepsneles ir paskutinius bėgančios šviesos sužibimus dešiniajame lempos krašte.
  `+1s,+1s` parodo jį po vienos sekundės ir po dviejų sekundžių.
  Peržiūros atsitiktinės reikšmės kiekvieno paleidimo metu tos pačios.
- Paleisk `velourbbs serve`. PuTTY parodo virpančią šviesą, žvaigždes, kibirkštis, liepsneles,
  sužibimus ir juostą.
