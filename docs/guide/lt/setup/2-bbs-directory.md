# 2. BBS katalogas

„BBS katalogas“ laiko vienos BBS failus. Šis vadovas naudoja `/srv/velourbbs`.
Jis priklauso tavo naudotojui. BBS taip pat veikia tavo naudotojo vardu (5 skyrius):
ji skaito failus ir rašo savo duomenų failą į šį katalogą.
Sukurk katalogą ir nukopijuok į jį BBS pavyzdį „Lempa“. Šaltinio kodo saugykloje paleisk:

```
sudo install -d -o "$USER" -g "$(id -gn)" /srv/velourbbs
cp -r examples/lamp/lt/. /srv/velourbbs/
```

Redaguok failus savo naudotojo vardu. Nenaudok `sudo` BBS katalogo failams.

„Lempos“ failai:

| Failas | Turinys |
|---|---|
| `bbs.toml` | Serverio nustatymai (3 skyrius). „Lempoje“ jis tuščias: numatytosios reikšmės. |
| `screens/screens.toml` | Ekranų nustatymai: plotis, stiliai ir kitos reikšmės. |
| `screens/_head.bbt` | Galvutė. Kiekvienas ekranas ją įterpia. |
| `screens/main.bbt` | Pirmasis ekranas: meniu. |
| `screens/about.bbt` | Tekstas apie BBS. |
| `screens/term.bbt` | Terminalo reikšmės. |
| `screens/texts.bbt` | Skaityklos tekstų sąrašas. |
| `screens/reader.bbt` | Ekranas, kuris rodo sąrašo tekstą. |
| `screens/goodbye.bbt` | Ekranas po seanso pabaigos. |
| `content/lamp.md` | Markdown tekstas: trumpa lempų istorija. |
| `content/rules.txt` | Paprastas tekstas: stoties taisyklės. |

Kaip keisti ekranus, skaityk [BBT vadove](../bbt/README.md).

Kiekvieną `velourbbs` komandą paleisk BBS kataloge, savo naudotojo vardu.
Visos taisyklės: [bbs-toml](../../../specs/bbs-toml.md), 2 skyrius.

## Rezultatas

```
cd /srv/velourbbs
velourbbs check
```

Komanda nieko neišveda.
