# Technologijos

Velour parašyta Go kalba. Tai viena programa: ji pati priima svečius per telnet,
piešia ekranus ir saugo duomenis. Pirminis kodas atviras:
[github.com/schors/velourbbs](https://github.com/schors/velourbbs).

## Bibliotekos

Be Go standartinės bibliotekos, Velour turi tik tris tiesiogines priklausomybes:

- `go-toml` skaito nustatymus ir ekranus. Ji nurodo nežinomo rakto eilutę ir stulpelį,
  todėl `velourbbs check` parodo klaidos vietą.
- `displaywidth` skaido tekstą į grafemas ir žino kiekvienos plotį. Ji turi Unicode 17.
- `goldmark` skaito Markdown: CommonMark ir GFM lenteles. Ji neturi savų priklausomybių.

## Savas kodas

Visa kita Velour daro pati, ir kiekvienas sprendimas turi priežastį:

- Telnet. Go bibliotekos nesuteikia serveriui NAWS ir TTYPE su ribotais buferiais.
  Savas analizatorius turi apie 200 eilučių.
- PROXY protokolas. Biblioteka turi API v0.x, o jos 2026 metų leidimai
  pakeitė numatytąjį elgesį. Savas 1 ir 2 versijų priėmėjas turi apie 200 eilučių.
- Išvestis į terminalą. `tcell` piešia visą ekraną iš karto, o Velour rodo puslapius
  su eilute "Toliau". Be to, `tcell` neatpažįsta terminalo prisijungimo metu.
- ANSI meno analizatorius. Programai reikia tik SGR stilių, o importui - apie 15 sekų.
  Pilnas terminalo analizatorius būtų didelė priklausomybė.
- Ekranų sąlygos. Jų gramatika turi 7 taisykles. Išraiškų kalba
  pridėtų aritmetiką ir funkcijas, kurių formatas neleidžia.

## Duomenys

Stoties duomenys yra viename JSON Lines tekstiniame faile: vienas įrašas kiekvienoje eilutėje. Sistemos operatorius jį skaito ir taiso teksto redaktoriumi.
Velour nenaudoja SQLite: jos failui skaityti reikia atskiros programos,
o SQLite prideda prie programos 6 MB.

## Gyvas ekranas

Ekranas gyvena, kol laukia klavišo, be jokio scenarijaus.
Gyvas laukas keičiasi savo vietoje kas sekundę. Likusio ekrano Velour iš naujo nepiešia.
Taip laikrodis tiksi apatinėje meniu ir skyrių eilutėje. Ekrane "Apie stotį"
taip pat keičiasi ryšio trukmė, atviri seansai ir stoties veikimo trukmė.

Efektai animuoja ekrano dalis. Jų yra šeši: žvaigždės mirga, žvaigždės atsiranda ir dingsta,
žvakė plazda, raidė sužiba, šviesa bėga eilute, trenkia žaibas.
Ši stotis naudoja bėgančią šviesą `sweep`: ji bėga per logotipą įėjimo ekrane
ir per liniją po meniu antrašte.
Velour siunčia į terminalą tik pasikeitusias ekrano dalis, o tarp kadrų praeina ne mažiau kaip 50 ms.

## ANSI menas

ANSI menas piešiamas Moebius, icy_draw ir PabloDraw. Šie redaktoriai paprastai rašo CP437 failus.
`velourbbs art import` vieną kartą paverčia redaktoriaus failą į UTF-8.
Serveris skaito tik UTF-8 ir niekada neskaito CP437.
Ekranas įterpia meną tarp teksto.
