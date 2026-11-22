# Išleidimas

Nauja stoties versija išleidžiama pagal žymą. Prižiūrėtojas įkelia žymą, pavyzdžiui `v0.1.0`,
ant `main` komito. Visa kita atlieka GitHub Actions.

## Surinkimas

Užduotis patikrina, ar žyma yra ant `main` komito. Tada ji paleidžia `go vet` ir testus
ir surenka Velour sistemoms Linux, macOS ir Windows.
Programos ir jų SHA-256 sumos patenka į GitHub Release.

## Pristatymas

Užduotis per SSH nusiunčia į serverį vieną archyvą: programą, katalogą `station/`,
systemd vienetus ir diegimo scenarijų. Serverio adresas ir SSH raktas yra
GitHub paslaptys. Užduoties žurnalas viešas, bet jame nėra paslapčių.

## Diegimas

Scenarijus serveryje pirmiausia paleidžia `velourbbs check` naujai stočiai.
Jei patikra randa klaidą, diegimas sustoja, o ankstesnė versija toliau veikia.
Tada scenarijus nukreipia nuorodą `current` į naują versiją ir paleidžia tarnybą iš naujo.

Duomenų failas yra už versijų ribų ir išlieka po kiekvieno diegimo.
Ekranas "Apie stotį" rodo žymą eilutėje "Velour versija".

## Be prarastų ryšių

Prievadą laiko systemd, ne Velour. Paleidžiant iš naujo prievadas lieka atviras:
nauji ryšiai laukia eilėje, ir juos priima nauja versija.
Atviri seansai parodo paleidimo iš naujo ekraną ir užsidaro.

## Grąžinimas

Norėdamas grąžinti ankstesnę versiją, prižiūrėtojas iš naujo paleidžia jos vykdymą GitHub Actions.
GitHub tai leidžia 30 dienų.
