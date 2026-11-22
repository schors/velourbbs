# Velour galimybės

Velour - BBS programa. Štai ką ji moka dabar.

## Ryšys

Svečiai jungiasi per telnet. Velour veikia su PuTTY, Tera Term ir inetutils telnet UTF-8 terminale.

Velour nustato terminalo tipą, terminalo dydį ir spalvų gylį: 16, 256 arba 16 milijonų spalvų.
Terminalui reikia UTF-8 koduotės. Jei jos nėra, Velour paaiškina svečiui, kaip ją įjungti.

Terminalo dydis - nuo 20 iki 255 stulpelių ir nuo 10 iki 100 eilučių.
Naujas terminalo dydis įsigalioja nuo kito ekrano.

## Ekranai

Ekranai - tai sistemos operatoriaus tekstiniai failai BBT formatu. Faile yra tekstas, stiliai, laukai ir sąlygos.
Vienas ekranas gali turėti skirtingus pavidalus siauram ir plačiam terminalui.

Eilutė gali išsitempti per visą ekrano plotį: taškų linijos, skiriamosios linijos ir centruoti laukai.
Tarpo eilutė laiko paskutines eilutes kiekvieno terminalo apačioje.
Ekranas gali augti kartu su terminalu ir palikti tuščius stulpelius abiejuose šonuose.

Tema suteikia stilius, ir kiekvienas spalvų gylis gali turėti savo stilius.
Velour pakeičia spalvą, kurios terminalas neturi, artimiausia spalva.

Laukai rodo seanso ir stoties duomenis: svečio adresą, ryšio trukmę,
BBS pavadinimą, Velour versiją, atvirus seansus, serverio laiką ir veikimo trukmę.
Sistemos operatorius nustato serverio laiko formą ir laiko juostą.

## Meniu

Spartieji klavišai veda iš ekrano į ekraną. Raidės veikia ir rusiškame išdėstyme:
`й` yra `q`. Meniu punktas gali veikti tik esant sąlygai.
Pomeniu grįžta į ekraną, kuris jį atidarė: jo punktas `back` to ekrano nenurodo.

## Tekstai

Failo ekranas rodo Markdown, paprastą tekstą arba pirminį kodą puslapiais.
Markdown turi antraštes, sąrašus, citatas, kodą, lenteles ir nuorodas su išnašomis.
Velour suskaičiuoja skaitymo laiką.

Vienas ekranas rodo daug failų: meniu punktas jam perduoda failą.
Failo ekranas gali rodyti pirminio kodo failo dalį ir eilučių numerius.

Sąrašo ekranas rodo katalogo failus puslapiais: kiekvieno failo pavadinimą ir skaitymo laiką.
Velour nuskaito katalogą paleidimo metu.

## ANSI menas

`velourbbs art import` paverčia ANSI meną iš Moebius, icy_draw ir PabloDraw į Velour formatą.
Ekranas įterpia meną tarp savo teksto.

## Sistemos operatoriui

`velourbbs check` randa klaidas ekranų failuose ir nurodo failą, eilutę ir stulpelį.
`velourbbs preview` parodo ekraną be ryšio: bet kokiu terminalo dydžiu,
bet kokiu spalvų gyliu ir po klavišų paspaudimų.

Stotis riboja ryšius: iš viso, o iš vieno adreso - ir atvirus, ir naujus per minutę.
Velour gauna svečio adresą iš tarpinio serverio per PROXY protokolą.

Įdiegimo vadovas rodo, kaip paleisti Velour su systemd.
