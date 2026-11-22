# 2 užduotis. Pirmasis ekranas

Šioje užduotyje rašai pirmąjį „Lempos“ ekraną. Meniu lieka toks pat: vienas punktas `G`. 4 užduotis prideda kitus punktus.

## Plotis ir stiliai

Sukurk `screens/screens.toml` su šiuo tekstu:

```toml
[screen]
width = [80, 80]
fill = "spaces"

[theme.default]
normal = "#e9dccb/#2a201d"
dim    = "#9e8a7c"
line   = "#8a6a5c"
title  = "#f7c7a3 bold"
key    = "#ffb088 bold"
orn    = "#f2a27e"
```

- `width = [80, 80]` nurodo mažiausią ir didžiausią ekrano plotį.
  Todėl ekranas visada yra 80 stulpelių pločio. Platesniame terminale BBS jį padeda per vidurį.
  Siauresniame terminale BBS nukerpa eilutes.
- `fill = "spaces"` užpildo terminalo eilutes po ekranu stiliumi `normal`.
- `[theme.default]` yra „tema“: ji suteikia pavadinimą kiekvienam stiliui.
  Stilius – tai teksto spalva, fono spalva po `/` ir atributai, pavyzdžiui, `bold`.
  Spalva yra `#rrggbb` arba skaičius nuo 0 iki 255.
- `normal` yra ekrano fono ir paprasto teksto stilius.
  Stilius be fono spalvos ima `normal` foną.

Visos taisyklės: [screens-toml](../../../specs/screens-toml.md), 3.1 ir 3.2 skyriai,
ir [bbt-format](../../../specs/bbt-format.md), 6.1–6.3, 6.5 ir 7 skyriai.

## Tekstas ir stiliaus žymės

- Kiekviena kūno eilutė yra viena ekrano eilutė. Tuščia eilutė yra tuščia ekrano eilutė.
  Už ekraną platesnė eilutė lūžta ties tarpu į kelias ekrano eilutes.
- Tarpai eilutės pradžioje išlieka.
- `{!name}` pradeda temos stilių `name`. `{!}` vėl pradeda stilių `normal`.
  Stilius tęsiasi kitose eilutėse iki kitos stiliaus žymės.

Pavyzdys:

```
       {!title}Lempa{!}  {!orn}☼ {!dim}BBS prie lempos šviesos{!}
```

Ši eilutė rodo „Lempa“ šviesiai persikine spalva ir pusjuodžiu šriftu, stiliumi `title`,
saulę `☼` šiltai oranžine spalva, stiliumi `orn`, ir šūkį pilkai ruda spalva, stiliumi `dim`.

## Tempimo žymės ir tarpo eilutės

- `{~─}` yra „tempimo žymė“. Ji užpildo laisvus savo eilutės langelius simboliu `─`.
  Laisvi langeliai – tai ekrano plotis atėmus kitus eilutės langelius.
  Tarpai eilutės pradžioje ir pabaigoje irgi skaičiuojami.
- `{~.}` užpildo laisvus langelius taškais. `{~}` užpildo juos tarpais.
- Eilutė su tempimo žyme visada yra viena ekrano eilutė. Ji nelūžta.
- Eilutė, kurioje yra tik `{~}`, yra „tarpo eilutė“. Ji duoda tuščias eilutes, kurios terminalui lieka
  po kitų ekrano eilučių. Todėl eilutės po jos lieka terminalo apačioje.

Pavyzdys:

```
  {!line}{~─}{!}  
```

Eilutė prasideda ir baigiasi 2 tarpais. Kai plotis yra 80, tempimo žymė duoda 76 simbolius `─`
3–78 stulpeliuose.

Visos taisyklės: [bbt-format](../../../specs/bbt-format.md), 6.6 skyrius.

## Ekranas

Perrašyk `screens/main.bbt` su šiuo tekstu:

```
+++
[[item]]
key = "G"
do = "logoff"
+++

  {!line}{~─}{!}  

       {!title}Lempa{!}  {!orn}☼ {!dim}BBS prie lempos šviesos{!}

       Labas vakaras. Čia tylu ir šilta.
       Prisitrauk kėdę prie lempos.
{~}
       {!dim}Šviesa dega iki ryto.{!}

  {!line}{~─}{!}  

    {!key}G{!}  {!dim}iki pasimatymo{!}

```

Abi plonos linijos baigiasi 2 tarpais. Failas baigiasi tuščia eilute po klavišų eilutės.

Kai dydis yra 80x24, tarpo eilutė duoda 11 tuščių eilučių, ir ekranas turi 24 eilutes:

```text

  ────────────────────────────────────────────────────────────────────────────

       Lempa  ☼ BBS prie lempos šviesos

       Labas vakaras. Čia tylu ir šilta.
       Prisitrauk kėdę prie lempos.











       Šviesa dega iki ryto.

  ────────────────────────────────────────────────────────────────────────────

    G  iki pasimatymo
```

- Fonas tamsiai rudas. Tekstas kreminis.
- Abi linijos rudos, stiliumi `line`. Jos užima 3–78 stulpelius.
- 24 eilutė tuščia: tai tuščia eilutė failo pabaigoje.
  Terminale su 43 eilutėmis tarpo eilutė duoda 30 tuščių eilučių.
  Baigiamoji eilutė, plona linija ir klavišų eilutė lieka apačioje.
- Baigiamoji eilutė ir žodžiai „iki pasimatymo“ pilkai rudi, stiliumi `dim`.
- Klavišas `G` oranžinis ir pusjuodis, stiliumi `key`.

Tera Term rodo 256 spalvas. Ten fonas tamsiai pilkas.
6 užduotis nurodo stilius 256 spalvoms.

## Rezultatas

- `velourbbs check` nieko neišveda.
- `velourbbs preview main` išveda 24 eilutes. 2 ir 4 eilučių stilių atkarpos tokios:

```
2: 1-2 #e9dccb/#2a201d, 3-78 #8a6a5c/#2a201d, 79-80 #e9dccb/#2a201d
4: 1-7 #e9dccb/#2a201d, 8-12 #f7c7a3/#2a201d bold, 13-14 #e9dccb/#2a201d, 15-16 #f2a27e/#2a201d, 17-39 #9e8a7c/#2a201d, 40-80 #e9dccb/#2a201d
```

- Telnet klientas rodo ekraną su tamsiu fonu visose eilutėse.

Išsaugok išvestį 3 užduočiai:

```
velourbbs preview main > task2.txt
```
