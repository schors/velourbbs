# BBT vadovas

Šis vadovas moko rašyti Velour BBS ekranus.
Jis skirtas dailininkui ir ekranų autoriui.
Tau reikia teksto redaktoriaus, terminalo ir telnet kliento.

## Įvadas

### BBS katalogas

„BBS katalogas“ – tai katalogas su vienos BBS failais:

| Failas | Turinys |
|---|---|
| `bbs.toml` | Serverio nustatymai. Šiame vadove failas tuščias. |
| `screens/screens.toml` | Ekranų nustatymai: plotis, stiliai ir kitos reikšmės. |
| `screens/<name>.bbt` | „Ekranas“: tekstinis failas, kurį rodo BBS. |
| `screens/_<name>.bbt` | „Fragmentas“: ekrano dalis. Jo pavadinimas prasideda `_`. |

Kiekvieną `velourbbs` komandą paleisk BBS kataloge.

### Programa `velourbbs`

Sukompiliuok programą su Go. Velour BBS pirminio kodo saugykloje paleisk:

```
go build ./cmd/velourbbs
```

Komanda sukuria failą `velourbbs` (`velourbbs.exe` sistemoje Windows).
Padėk jį į katalogą, kuris yra tavo `PATH`. Arba paprašyk programos sistemos operatoriaus.

### Pavyzdys

Užduotys kuria pavyzdinę BBS „Lempa“. Jos failai yra
[`examples/lamp/lt/`](../../../../examples/lamp/lt/).

- Eilutė su keliu ir dvitaškiu rodo visą pavyzdžio failą.
  Pavyzdys: `` `examples/lamp/lt/screens/main.bbt`: ``.
  Įrašyk šį failą į savo BBS katalogą be `examples/lamp/lt/`: `screens/main.bbt`.
- Kiekvienas kitas blokas rodo failo dalį, komandą arba išvestį.

Kiekviena užduotis nurodo specifikacijas, kuriose yra visos taisyklės.

## Dalys

- [1 užduotis. BBS katalogas](1-bbs-directory.md)
- [2 užduotis. Pirmasis ekranas](2-first-screen.md)
- [3 užduotis. Galvutė](3-header.md)
- [4 užduotis. Meniu su sparčiaisiais klavišais](4-menu.md)
- [5 užduotis. Laukai ir sąlygos](5-fields.md)
- [6 užduotis. Spalvų gylis](6-colour-depth.md)
- [7 užduotis. Tekstai](7-texts.md)
- [8 užduotis. Sąrašai](8-lists.md)
- [9 užduotis. ANSI menas](9-art.md)
- [10 užduotis. Efektai](10-effects.md)
- [11 užduotis. Klaidos](11-errors.md)

## Ko nėra MVP

BBS dar neturi šių funkcijų:

- žymeklių (`{@<id>}`) ir punkto pasirinkimo rodyklių klavišais.
  Naudotojas pasirenka punktą tik jo klavišu;
- prieigos patikrų (`require`). Joms reikia naudotojų;
- skyrybos ženklų klavišų, pavyzdžiui, `?`;
- punkto pavadinimo. BBS nespausdina punkto teksto;
- Markdown failo turinio, aktyvių nuorodų į kitus failus
  ir pirminio kodo failų sintaksės spalvų;
- kitų BBS programų rodymo kodų: pipe kodų, `@` kodų ir MCI kodų.
  Keitiklio nėra.

Visos taisyklės: [bbt-format](../../../specs/bbt-format.md), 10 skyrius,
ir [file-view](../../../specs/file-view.md), 12 skyrius.
