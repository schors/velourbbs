# Procesas

## Ką daro

Velour daugiausia rašo DI modeliai: specifikacijas, kodą, testus ir vadovus.
Žmogus operatorius valdo darbą ir priima sprendimus.
Visas darbas atviras: issues, pull request ir peržiūros yra
[github.com/schors/velourbbs](https://github.com/schors/velourbbs).

## Kaip

### Užduotys

Kiekvienas darbas yra GitHub issue. Issue tipas nurodo jo kelią:

- `idea` - idėja eilėje;
- `dev` - didelis darbas iš kelių blokų;
- `one-shot` - nedidelė užduotis;
- `docs` - vadovas naudotojams.

Šis stoties skyrius perrašytas viena `one-shot` užduotimi.

### Vartai

Vartai - sustojimas, kai darbas laukia operatoriaus atsakymo.
Pirmuose vartuose modelis parodo planą ir užduoda klausimus, o operatorius renkasi.
Priimtą planą modelis įrašo į issue. Taip sprendimą priima žmogus, dar prieš rašant kodą.
Antrieji vartai - pull request. Prižiūrėtojas jį priima arba rašo pastabas.

### Pirmiausia specifikacija

Programos taisyklės yra `docs/specs/`. Jei kodas skiriasi nuo specifikacijos,
specifikacija keičiama tame pačiame komite. Taip kiekviena taisyklė turi vieną apibrėžimą.

### Testai

Testai tikrina kiekvieną dalį, kurią galima patikrinti atskirai.
Prieš kiekvieną komitą vyksta `go vet`, `go test` ir `golangci-lint`.

Stoties testas paleidžia `velourbbs check` kataloge `station/`. Tada jis nufotografuoja
kiekvieną ekraną esant 80x24 ir 132x43 su 256 spalvomis ir lygina langelių nuotrauką su etalonu.
Etalonai yra git, todėl kiekvienas ekrano pakeitimas matomas pull request.
Be to, testas fiksuoja bėgančią šviesą ir gyvų laukų tiksėjimą kadras po kadro.

### Peržiūros

Kiekvieną pull request skaito DI peržiūrėtojai. Kiekvienas turi savo poziciją: ką tikrinti
ir kokiu požiūriu. Visus atsiliepimus modelis paskelbia pull request prieš pirmą pataisą.
Kiekvienoje atsiliepimo gijoje jis atsako: pataisyta ar ne ir kodėl.

## Ką gauna

- Atvirą kiekvieno pakeitimo istoriją: issue, planą, peržiūras ir atsakymus į jas.
- Ekranų etalonus, pagal kuriuos matomas kiekvienas stoties pakeitimas.
- Komitus, kuriuose kiekvienas modelis autorius įrašytas bendraautoriu.

## Ribos

- Atsiliepimas - tik informacija. Darbą priima merge, o jį daro prižiūrėtojas, ne modelis.
- Modelis dirba su operatoriaus žetonu, todėl jo įrašai yra operatoriaus įrašai.
  Kiekvienas modelio įrašas prasideda eilute `Agent: <modelis>`.
