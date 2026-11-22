# Terminalai

## Ryšys

Dabar svečiai jungiasi per telnet. Velour veikia su PuTTY, Tera Term ir telnet
UTF-8 terminale: Windows Terminal, xterm, tmux, iTerm2, Ghostty.
Su kitais klientais Velour veikia be garantijų.

## Ką Velour nustato pati

Prisijungimo metu Velour klausia terminalo ir iš atsakymų nustato:

- terminalo dydį: nuo 20 iki 255 stulpelių ir nuo 10 iki 100 eilučių;
- terminalo tipą;
- spalvų gylį: 16, 256 arba 16 milijonų spalvų;
- koduotę.

Velour pakeičia spalvą, kurios terminalas neturi, artimiausia spalva.
Taip vienas ekranas atrodo teisingai kiekvienu spalvų gyliu.

## Tik UTF-8

Velour rašo tekstą tik UTF-8. UTF-8 turi kirilicą, lotynišką abėcėlę su lietuviškomis raidėmis,
linijas ir ANSI meno blokus iš karto. 8 bitų koduotė turi tik 256 simbolius, ir kiekviena turi savo rinkinį.
Pavyzdžiui, CP437 neturi nei kirilicos, nei lietuviškų raidžių. Su viena koduote trys kalbos gyvena vienoje stotyje,
o ekranai, tekstai ir ANSI menas turi vieną pavidalą.

Velour koduotę tikrina taip: ji išspausdina raidę `â` ir klausia, kur yra žymeklis.
UTF-8 koduotėje žymeklis pasislenka vienu langeliu, 8 bitų koduotėje - dviem.
Jei žymeklis pasislinko dviem langeliais, Velour paaiškina, kaip pakeisti koduotę,
ir užveria ryšį. Jei terminalas neatsako, Velour laiko, kad tai UTF-8.

## Kas veikia ne visur

- SyncTERM ir NetRunner neturi UTF-8. Jie gauna paaiškinimą ir neįeina.
- Windows telnet.exe veikia be garantijų: tik 16 spalvų,
  o ANSI menas ir linijos gali subyrėti.
- MUD klientai, pavyzdžiui Mudlet, neveikia: jie neišvalo ekrano ir neperkelia žymeklio.

## Kas toliau

- SSH.
- telnets: telnet per TLS.

Sprendimo dėl CP437 terminalų ir terminalų be ANSI dar nėra.
