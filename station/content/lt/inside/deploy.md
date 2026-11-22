# Išleidimas

## Ką daro

Nauja stoties versija išleidžiama pagal žymą. Prižiūrėtojas įkelia anotuotą žymą,
pavyzdžiui `v0.1.0`, ant `main` komito. Visa kita atlieka GitHub Actions.

## Kaip

### Surinkimas

1. Užduotis patikrina žymą: jos pavidalą ir tai, kad žymos komitas yra `main`.
2. Ji paleidžia `go vet` ir testus.
3. Ji surenka Velour su `CGO_ENABLED=0` platformoms linux/amd64, linux/arm64,
   darwin/arm64 ir windows/amd64.
4. `go version -m` patikrina, kad kiekvienoje programoje įrašyta ši žyma.
5. Programos ir failas `SHA256SUMS` patenka į GitHub Release.
   Leidimo tekstas - žymos anotacija.

Užduotis naudoja tik `actions/checkout` ir `actions/setup-go`.
Kiekvienas užfiksuotas pilnu komito SHA. Svetimų veiksmų SSH ir leidimams nėra:
jie gautų SSH raktą.

### Pristatymas

Užduotis per SSH nusiunčia į serverį vieną archyvą: programą, katalogą `station/`,
systemd vienetus ir diegimo scenarijų. SSH raktas ir serverio adresas yra GitHub
aplinkos paslaptys. Aplinka leidžia tik žymas `v*`.

### Diegimas

1. Scenarijus paleidžia `velourbbs check` naujai stočiai. Patikra skaito ir duomenų failą.
   Klaida sustabdo diegimą, o ankstesnė versija toliau veikia.
2. Jis padeda versiją į `releases/<žyma>/` ir nukreipia nuorodą `current` į ją.
3. Jis paleidžia tarnybą iš naujo ir po 3 sekundžių patikrina, kad ji veikia.
4. Jis palieka 3 naujausias versijas.

Prievadą laiko systemd vienetas `velourbbs.socket`, ne programa. Paleidžiant iš naujo
prievadas lieka atviras: nauji ryšiai laukia eilėje, ir juos priima nauja versija.
Atviri seansai parodo paleidimo iš naujo ekraną ir užsidaro.

## Ką gauna

- Versiją, kurią ekranas "Apie stotį" rodo eilutėje "Velour versija".
- Viešą užduoties žurnalą: jame yra žingsniai ir rezultatas, bet nėra paslapčių.
- Duomenų failą, kuris yra už versijų ribų ir išlieka po kiekvieno diegimo.

## Ribos

- Grąžinimas - ankstesnės versijos vykdymo pakartojimas GitHub Actions. GitHub tai leidžia 30 dienų.
  Vėliau - revert `main` šakoje ir nauja žyma.
- Po duomenų failo migracijos ankstesnė versija nepraeis patikros.
  Tada reikia grąžinti tos versijos failo kopiją.
- Tas, kas gali įkelti žymą `v*`, gali pakeisti ir užduotį. Rašyti į saugyklą
  gali tik prižiūrėtojai.
