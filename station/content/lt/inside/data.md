# Duomenys

## Ką daro

Velour saugo stoties duomenis viename teksto faile. Sistemos operatorius skaito ir taiso jį
teksto redaktoriumi.

## Kaip

Failas yra JSON Lines ([jsonlines.org](https://jsonlines.org/)): UTF-8, vienas įrašas -
vienas JSON objektas eilutėje. Pirma eilutė - antraštė su formato versija:

```
{"format":"velourbbs-store","version":2}
```

Kiekvienas kitas įrašas turi rūšį, lauką `kind`. Kodas aprašo kiekvieną rūšį: jos laukus, patikras ir
įrašų skaičiaus ribą. Kai įrašų daugiau nei riba, seniausi išeina.
Laikas - RFC 3339 UTC, su sekundėmis.

Velour laiko visus įrašus atmintyje. Kiekvienas pakeitimas rašo visą failą:

1. Patikrina naujus įrašus.
2. Įrašo juos į laikiną failą `<failas>.tmp` su teisėmis 600 ir išsaugo jį diske.
3. Pervadina laikiną failą ant senojo. Tai įsipareigojimo taškas.
4. Išsaugo katalogą diske.

## Ką gauna

- Po gedimo Linux ar macOS sistemoje faile yra įrašai prieš pakeitimą arba po jo, ne jų mišinys.
- Klaida prieš pervadinimą palieka seną failą ir senus įrašus.
- Kiekvienas formato pakeitimas didina versiją. Migracijos skaitymo metu paverčia seną failą nauja versija.
  Klaida parodo failo eilutę.

## Ribos

- Failas - ne didesnis nei 16 MB, nes visas yra atmintyje.
- Vienas failas - vienas `velourbbs serve`. Failas nerakinamas.
- Failo pataisa, kol serveris veikia, dings po kito rašymo.
- Senesnė Velour versija neskaito naujesnės versijos failo. Prieš diegimą su migracija
  padarykite failo kopiją.
- Windows sistemoje Go nedaro pervadinimo atominio, ir jis nepavyksta,
  kol failą atidariusi kita programa.
- Velour nenaudoja SQLite ir bbolt: jų failo neperskaitysi be atskiros programos.
  Be to, `modernc.org/sqlite` prideda programai 6 MB ir 10 modulių.
