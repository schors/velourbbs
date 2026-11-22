# Išvestis ir įvestis

## Ką daro

Velour piešia ekraną atmintyje kaip langelių tinklelį, o terminalui siunčia baitus šiam tinkleliui.
Kita kryptimi ji paverčia baitus iš terminalo klavišais.

## Kaip: išvestis

### Langeliai

Langelis saugo vieną grafemą ir stilių. Grafema - tai, ką žmogus mato kaip vieną ženklą:
raidė su kirčiu, vėliava, jaustukas iš kelių kodų
([UAX #29](https://www.unicode.org/reports/tr29/)). Grafemos plotis - 0, 1 arba 2 langeliai,
pagal [Unicode 17](https://www.unicode.org/versions/Unicode17.0.0/).
Ambiguous pločio simboliai, kirilica ir linijos, užima vieną langelį.

### Puslapis

Kiekvieną puslapį Velour siunčia vienu rašymu:

1. `ESC[?2026h` - sinchroninės išvesties pradžia
   ([mode 2026](https://contour-terminal.org/vt-extensions/synchronized-output/)).
   Terminalas kaupia išvestį ir parodo puslapį iš karto, ne eilutę po eilutės.
   Terminalas be šio režimo jį praleidžia.
2. `ESC[0m ESC[2J` - stiliaus atstatymas ir ekrano išvalymas.
3. Kiekvienai eilutei: `ESC[<eilutė>;1H`, tada eilutės langeliai.
   Eilučių pabaigų nėra, todėl terminalas neslenka ekrano.
4. `ESC[0m ESC[?2026l` - stiliaus atstatymas ir sinchroninės išvesties pabaiga.

### Stiliai

Prieš langelį su nauju stiliumi Velour siunčia SGR: `ESC[0;<parametrai>m`.
Parametrai: `1` paryškintas, `4` pabrauktas, `5` mirksėjimas; `30`-`37` ir `90`-`97` 16 spalvų,
`38;5;n` 256 spalvoms, `38;2;r;g;b` RGB. Fonas - tie patys skaičiai, 10 didesni: `40`, `48;5;n`.

Spalvą, kurios terminalas neturi, Velour pakeičia artimiausia: su mažiausiu
RGB atstumo kvadratu. 256 spalvos - xterm paletė, 16 spalvų - VGA spalvos.
Todėl vienas ekranas atrodo teisingai ir truecolor, ir 256, ir 16 spalvų.

### Sudėtingos grafemos

Terminalai skirtingai skaičiuoja grafemos iš kelių kodų plotį.
Pavyzdžiui, `❤️` su VS16 PuTTY, xterm, Tera Term ir VTE užima vieną langelį.
Todėl po tokios grafemos Velour išspausdina tarpą ir pastato žymeklį į reikiamą vietą:
`ESC[<eilutė>;<stulpelis>H`. Kiti langeliai atsistoja į savo stulpelius esant bet kokiam pločiui.

### Atnaujinimas

Gyvas laukas keičiasi kas sekundę, o efektas keičia langelius, kol vyksta. Tada Velour
nepiešia puslapio iš naujo, o siunčia tik pakeistas atkarpas:

1. `ESC[?2026h ESC[?25l` - sinchroninė išvestis ir paslėptas žymeklis.
2. Kiekvienai atkarpai: `ESC[<eilutė>;<stulpelis>H` ir jos langeliai.
   Dvi atkarpos su 3 ar mažiau vienodų langelių tarp jų yra viena atkarpa.
3. `ESC[0m`, žymeklis į ankstesnę vietą, `ESC[?25h ESC[?2026l`.

Kadrai eina ne dažniau kaip kas 50 ms. Jei niekas nepasikeitė, Velour nieko nesiunčia.

## Kaip: įvestis

Baitai iš kliento eina per keturis žingsnius:

1. Telnet analizė.
2. Enter: CR LF, CR NUL, CR ir LF duoda vieną klavišą `ENTER`.
3. UTF-8 dekoderis ([RFC 3629](https://www.rfc-editor.org/rfc/rfc3629)).
   Neteisingus baitus jis atmeta.
4. Klavišų dekoderis.

Klavišų dekoderis žino rodyklių, Home, End, PgUp, PgDn, Del ir F1-F12 sekas
xterm, VT ir Linux konsolės pavidalais. Po ESC jis laukia kito baito 150 ms.
Jei baito nėra, tai klavišas Esc. Sekos viduje jis kiekvieno baito laukia iki 1 sekundės:
taip atsakymas, atėjęs dviem dalimis, netampa klavišais.
OSC, DCS ir APC eilutes jis atmeta visas: pavėlavęs apklausos atsakymas irgi netampa klavišais.

Įvesties eilutėje Velour įjungia bracketed paste: `ESC[?2004h`
([xterm](https://invisible-island.net/xterm/ctlseqs/ctlseqs.html)).
Terminalas apgaubia įklijavimą žymėmis `ESC[200~` ir `ESC[201~`.
Taip įklijavimas nespaudžia meniu punktų: už įvesties eilutės Velour jį atmeta.

## Ribos

- PuTTY 0.83, Tera Term, xterm ir tmux iki 3.7 nežino 2026 režimo.
  Jie rodo langelius, kai šie ateina. Žymeklis paslėptas ir nebėgioja per ekraną.
- Užpildui `fill = "erase"` reikia bce. tmux ir GNU screen jo neturi,
  ir tuščios eilutės lieka terminalo spalvos.
- Klavišus su modifikatoriais, pavyzdžiui Ctrl+rodyklė (`ESC[1;5A`), Velour atmeta.
  Alt+raidė duoda Esc ir raidę, išskyrus Alt+P.
- Seką, ilgesnę nei 128 baitai, Velour atmeta visą.
- Mirksėjimą terminalai rodo skirtingai, o PuTTY ir iTerm2 jį pagal nutylėjimą išjungia.
  Todėl vietoj mirksėjimo Velour turi efektus.
- Pelės nėra.
