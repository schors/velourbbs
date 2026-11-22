# Ekranai

## Ką daro

Ekranas - BBT teksto failas, kurį rašo sistemos operatorius. Kiekvieną kartą rodydama
Velour iš jo sudaro puslapį svečio terminalui: plotį, aukštį, spalvas ir kalbą.

## Kaip

### Failas

BBT failas yra UTF-8. Viršuje tarp eilučių `+++` yra nustatymai TOML 1.1 formatu
([TOML](https://toml.io/en/v1.1.0)): plotis, meniu punktai, efektai.
Žemiau eina tekstas su žymėmis:

- `{term.cols}` - laukas, `{term.type      }` - fiksuoto pločio laukas;
- `{!gold}` - stilius pagal pavadinimą iš temos;
- `{#if term.cols >= 100}`, `{#else}`, `{/if}` - sąlyga;
- `{>_foot}` - fragmento įterpimas, `{>logo.ans}` - meno įterpimas;
- `{~}`, `{~.}`, `{~─}` - tempimas.

### Plotis ir aukštis

Ekranas nurodo pločio intervalą. Ši stotis turi `width = [68, 108]`, `align = "center"`
ir `gutter = 6`: teksto stulpelis yra 64 langeliai esant 80 stulpelių ir 104 esant 132.
Tempimai dalijasi laisvais eilutės langeliais: `{~.}` piešia taškų liniją, `{~─}` - liniją.
Eilutė `{~}` dalijasi laisvomis ekrano eilutėmis. Taip klavišų eilutė lieka apačioje
ir esant 24, ir esant 43 eilutėms.

### Sąlygos

Sąlygų gramatika turi 7 taisykles: `or`, `and`, `not` ir palyginimai `==`, `!=`, `<`, `<=`, `>`, `>=`.
Skaičiai lyginami bet kuriuo operatoriumi, eilutės - tik `==` ir `!=`.
Skliaustų ir aritmetikos nėra. Laukų tipai žinomi iš anksto, todėl `velourbbs check`
randa tipo klaidą prieš paleidimą.

### Tema

`screens.toml` paverčia prekės ženklo knygos vaidmenų pavadinimus stiliais. Lentelės `[theme.16]` ir `[theme.256]`
duoda kitas spalvas 16 ir 256 spalvų, `[theme.light]` - šviesiam fonui.
Su `theme = "auto"` Velour pasirenka šviesią temą pagal fono spalvą iš terminalo apklausos.

### Gyvas ekranas

Ekranas gyvas, kol laukia klavišo, ir tam nereikia scenarijaus.

- Gyvas laukas, pavyzdžiui `{bbs.time live}`, kas sekundę atsinaujina savo vietoje.
  Taip eina laikrodis skyrių apačioje. Ekrane "Apie stotį" taip pat keičiasi
  prisijungimo laikas, atvirų seansų skaičius ir stoties veikimo laikas.
- Efektų yra šeši: `twinkle` (žvaigždės mirga), `sparkle` (žvaigždės užsidega ir užgęsta),
  `candle` (žvakė virpa), `glint` (raidė sužimba), `sweep` (šviesa bėga eilute)
  ir `lightning` (trenkia žaibas).
- Ši stotis naudoja `sweep`: logotipe prie įėjimo, antraštėje su jos linija
  ir linijoje virš klavišų. Šviesa antraštėje bėga maždaug kas 45 sekundes.

### Trys kalbos

Vienas BBS katalogas laiko rusų, anglų ir lietuvių kalbas. Kalbos ekranai yra
`screens/<kalba>/`, tekstai - `content/<kalba>/`. Failas be teksto, pavyzdžiui logotipas,
yra vienas visoms kalboms. Įėjimo ekranas klausia kalbos ir atidaro meniu ta kalba.
Testas tikrina, kad visų kalbų ekranai turi tuos pačius punktus.

### Patikra

`velourbbs check` perskaito visus ekranus prieš paleidimą. Klaida ateina su failu,
eilute ir stulpeliu. Nežinomas TOML raktas - irgi klaida su vieta.

## Ką gauna

Vieną puslapį konkrečiam terminalui. Pavyzdys - ekranas "Apie stotį":
jis rodo dydį, tipą, programos pavadinimą, spalvas, foną ir kirilicos plotį,
kuriuos Velour nustatė prisijungimo metu.

## Ribos

- Išdėstymas nesikeičia, kai `term.ambiguous` lygus 2: stulpeliai pasislenka, rėmai ir menas sugenda.
- Grafema, kurios plotis 0 arba kuri ilgesnė nei 32 baitai, faile yra klaida.
- Scenarijų kol kas nėra. Lua moduliai - planuose.
