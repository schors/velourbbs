# Go ir bibliotekos

Velour parašyta Go kalba. Tai viena programa be cgo: ji priima svečius per telnet,
piešia ekranus, saugo duomenis ir tikrina sistemos operatoriaus failus.
Pirminis kodas atviras: [github.com/schors/velourbbs](https://github.com/schors/velourbbs).

## Kodėl Go

Kalba pasirinkta 2026 m. lapkričio 12 d. Go buvo palyginta su Rust, ir keturi peržiūrėtojai
pasirinko Go: Go programuotojas, Rust programuotojas, terminalų ir protokolų inžinierius
ir saugumo inžinierius. Palyginimas yra `docs/specs/host-language.md`.

- Seansai. Kiekvienas seansas yra atskira goroutine. Kodas skaito ir rašo blokuojančiais kvietimais,
  o laiką riboja `net.Conn` terminai. Rust kalboje seansas yra tokio užduotis su `.await`.
- SSH ateičiai. Go turi `golang.org/x/crypto/ssh` iš Go projekto.
  Rust turi `russh` iš nedidelės komandos.
- Surinkimas. Go surenka statinę programą be cgo, greitai ir su mažu priklausomybių grafu.
  Rust renka lėčiau, o priklausomybių `build.rs` ir procedūriniai makrosai vykdomi surinkimo metu.
- Terminalas. Čia kalbos lygios: meno analizatorius, langelių modelis, pakeitimų išvestis
  ir klavišų dekoderis abiem atvejais yra savas kodas.

## Kur Rust stipresnė

- Lua moduliams. Rust kalboje `mlua` nustato atminties ribą kiekvienai Lua būsenai.
  `gopher-lua` Go kalbai tokios ribos neturi. Lua moduliai kol kas tik planuose.
- Paruoštos dalys terminalo analizei: `vte` ir `vt100`.

## Bibliotekos

Be Go standartinės bibliotekos, Velour turi tris tiesiogines priklausomybes:

- `go-toml/v2` skaito nustatymus ir ekranus. Ji palyginta su `BurntSushi/toml`.
  Pasirinkta `go-toml`: ji nurodo nežinomo rakto eilutę ir stulpelį,
  todėl `velourbbs check` parodo klaidos vietą.
- `displaywidth` skaido tekstą į grafemas ir duoda kiekvienos plotį pagal Unicode 17.
  Ji palyginta su `uniseg` (Unicode 15) ir `go-runewidth` (matuoja runas, ne grafemas).
- `goldmark/v2` skaito Markdown: CommonMark 0.31.2 ir GFM lenteles.
  Ji neturi savų priklausomybių. 2 versija jauna, ir pataisa gali sugadinti API,
  todėl versija užfiksuota `go.mod`.

Žurnalą rašo `log/slog`, komandas analizuoja `flag`, duomenis koduoja `encoding/json`.

## Savas kodas vietoj bibliotekų

- Telnet. Go bibliotekos nesuteikia serveriui NAWS ir TTYPE su ribotais buferiais.
  Savas analizatorius turi apie 200 eilučių.
- PROXY protocol. `go-proxyproto` turi 0.x versijos API, o jos 2026 metų leidimai
  pakeitė numatytąjį elgesį. Savas 1 ir 2 versijų priėmėjas turi apie 200 eilučių.
- Išvestis į terminalą. `tcell` piešia visą ekraną kaip programa, o Velour rodo
  puslapius su eilute "Toliau". Be to, `tcell` neapklausia terminalo prisijungimo metu.
- ANSI menas. Serveriui reikia tik SGR stilių, o importui - apie 15 komandų.
  Pilnas terminalo analizatorius `charmbracelet/x/ansi` - didelė priklausomybė.
- Ekranų sąlygos. Jų gramatika turi 7 taisykles. `expr` pridėtų aritmetiką
  ir funkcijas, kurių formatas neturi.
- Socket activation. Kodas skaito du aplinkos kintamuosius ir kviečia `net.FileListener`:
  apie 30 eilučių vietoj `go-systemd`.

## Ribos

- Savas kodas - tai savi testai ir sava priežiūra. Kiekvienas sprendimas įrašytas
  savo specifikacijos skyriuje "Libraries".
- Go standartinė biblioteka ir `golang.org/x/text` neskaido teksto į grafemas.
  Todėl plotį duoda trečiosios šalies `displaywidth`.
