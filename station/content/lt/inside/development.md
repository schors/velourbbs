# Kūrimas

Velour daugiausia parašyta su DI. DI modeliai rašo specifikacijas, kodą
ir vadovus. Žmogus operatorius valdo darbą ir priima sprendimus.
Visas darbas atviras: užduotys, pull request ir peržiūros yra [github.com/schors/velourbbs](https://github.com/schors/velourbbs).

## Užduotys

Kiekvienas darbas yra GitHub užduotis (issue). Užduoties tipas nurodo jos kelią:

- `idea` - idėja atsargoje;
- `dev` - didelis darbas iš kelių blokų;
- `one-shot` - nedidelė užduotis;
- `docs` - vadovas naudotojams.

Ši stoties dalis buvo sukurta viena `one-shot` užduotimi.

## Vartai

Vartai - tai sustojimas, kai darbas laukia operatoriaus.
Prie pirmųjų vartų modelis parodo planą ir užduoda klausimus, o operatorius renkasi.
Modelis paskelbia priimtą planą užduotyje. Taip sprendimą priima žmogus,
ir dar prieš kodą.

Antrieji vartai - pull request. Prižiūrėtojas jį sulieja arba parašo komentarus.

## Pirma specifikacija

Programos taisyklės yra `docs/specs/`. Jei kodas skiriasi nuo specifikacijos,
specifikacija keičiama tame pačiame komite. Taip kiekviena taisyklė turi vieną apibrėžimą.
Testai tikrina kiekvieną dalį, kurią galima patikrinti atskirai.

## Peržiūros

DI peržiūrėtojai skaito kiekvieną pull request. Kiekvienas turi poziciją: ką tikrinti ir kokiu požiūriu.
Modelis paskelbia visas peržiūras pull request prieš pirmą pataisymą.
Kiekvienoje peržiūros gijoje jis atsako: pataisyta, arba nepataisyta ir kodėl.
Taip kiekvienas pakeitimas turi viešą istoriją.

Peržiūra yra tik informacija. Darbą priima suliejimas, o sulieja prižiūrėtojas,
ne modelis. Kiekvienas komitas nurodo modelį kaip bendraautorių.
