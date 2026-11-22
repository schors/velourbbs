# Stotis

Ši stotis - Velour namų BBS. Ji veikia su ta pačia programa, kurią galite
įdiegti patys, ir naudoja tik tai, ką Velour moka dabar.
Taip stotis yra ir produkto svetainė, ir jo gyvas pavyzdys.

## Failai

Visa stotis yra katalogas `station/` Velour saugykloje:
[github.com/schors/velourbbs](https://github.com/schors/velourbbs). Joje nėra kodo,
tik sistemos operatoriaus failai:

- `bbs.toml` - BBS pavadinimas ir duomenų failo kelias;
- `screens/` - ekranai BBT formatu ir `screens.toml` su tema;
- `content/` - tekstai Markdown formatu, kaip šis.

## Ekranai

Ekranas yra BBT tekstinis failas. Failas prasideda nustatymais ir meniu punktais.
Toliau eina tekstas su stiliais, laukais ir sąlygomis.
Vienas failas duoda ekraną 80 ir 132 stulpeliams: tempimo žymės nubrėžia linijas,
taškų linijas ir kraštus.

Stoties išvaizdą nustato prekės ženklo knyga: aukso raidės, bronzos plonos linijos,
tušti langeliai ir jokių rėmelių. Tema paverčia prekės ženklo knygos vaidmenis spalvomis
truecolor, 256 ir 16 spalvų režimams.

## Trys kalbos

Viename BBS kataloge yra visos trys kalbos: rusų, anglų ir lietuvių. Įėjimo ekranas klausia kalbos
ir ja atidaro meniu. Visų kalbų ekranai turi tuos pačius meniu punktus.
Tai tikrina testas.

## Tekstai

Vadovai yra saugyklos `docs/guide/`. Stotis juos skaito per
simbolinę nuorodą ir jų nekopijuoja. Taip vadovas stotyje ir GitHub yra vienas failas.

## Patikros

Testas paleidžia `velourbbs check` stočiai. Tada jis paima kiekvieną ekraną 80x24
ir 132x43 dydžiu ir palygina langelius su etalonine kopija. Etaloninės kopijos yra git,
todėl kiekvienas ekrano pakeitimas matyti pull request.
