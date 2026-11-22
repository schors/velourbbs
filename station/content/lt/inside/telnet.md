# Telnet

## Ką daro

Velour perjungia klientą į simbolių režimą ir nustato terminalo dydį ir tipą.

## Kaip

Telnet siunčia komandas duomenų sraute. Komanda prasideda baitu IAC (255)
([RFC 854](https://www.rfc-editor.org/rfc/rfc854)). Parinktys įjungiamos komandomis
`WILL`, `WONT`, `DO` ir `DONT`. Parinkties duomenys eina tarp `IAC SB` ir `IAC SE`.

Iškart po prisijungimo Velour vienu rašymu siunčia:

- `WILL ECHO` ([RFC 857](https://www.rfc-editor.org/rfc/rfc857)).
  Aidą daro serveris, ir klientas nustoja pats spausdinti įvestį.
- `WILL SGA` ir `DO SGA` ([RFC 858](https://www.rfc-editor.org/rfc/rfc858)).
  Kartu su ECHO tai simbolių režimas: klientas siunčia kiekvieną klavišą iš karto, ne eilutę po Enter.
- `WILL BINARY` ir `DO BINARY` ([RFC 856](https://www.rfc-editor.org/rfc/rfc856)).
  8 bitų duomenys abiem kryptimis: taip praeina UTF-8 baitai.
- `DO NAWS` ([RFC 1073](https://www.rfc-editor.org/rfc/rfc1073)): klientas praneša lango dydį.
- `DO TTYPE` ([RFC 1091](https://www.rfc-editor.org/rfc/rfc1091)): klientas praneša terminalo tipą.
  Kai klientas atsako `WILL TTYPE`, Velour vieną kartą siunčia `IAC SB TTYPE SEND IAC SE`.

Velour saugo kiekvienos parinkties būseną kiekvienai pusei, pagal Q metodą
iš [RFC 1143](https://www.rfc-editor.org/rfc/rfc1143). Į užklausą, kuri sutampa
su dabartine būsena, ji neatsako. Todėl derybos neužsiciklina.
Parinktis, kurių nėra sąraše, Velour atmeta: `DONT` į `WILL`, `WONT` į `DO`.

Telnet analizė vyksta prieš UTF-8 dekoderį ir klavišų dekoderį:
IAC gali perskirti kelių baitų simbolį arba valdymo seką.
`IAC IAC` duoda baitą 255. Į AYT Velour atsako BEL. IP ir BRK baigia seansą.

## Ką gauna

- Dydį. NAWS duoda plotį ir aukštį dviem 16 bitų skaičiais.
  Velour juos riboja: 20-255 stulpelių ir 10-100 eilučių. Be NAWS dydis yra 80x24.
  Naujas NAWS seanso metu pakeičia dydį nuo kito ekrano.
- Tipą. Pirmas TTYPE pavadinimas mažosiomis raidėmis tampa `term.type`. Be pavadinimo - `ansi`.
  Klientai vadina save skirtingai: PuTTY - `XTERM`, Tera Term - `xterm`,
  inetutils telnet - `$TERM` didžiosiomis raidėmis, Windows telnet.exe - `ANSI`.
- Enter. CR LF, CR NUL, pavienis CR ir pavienis LF duoda vieną Enter.
  inetutils telnet simbolių režime siunčia CR NUL.

Išvestyje Velour visada siunčia CR LF ir padvigubina kiekvieną baitą 255.

## Ribos

- Vienos parinkties duomenys - ne daugiau kaip 256 baitai. Jei jie ilgesni arba nesibaigia
  per 10 sekundžių, Velour uždaro ryšį.
- Parinkties CHARSET ([RFC 2066](https://www.rfc-editor.org/rfc/rfc2066)) nėra.
  Koduotę Velour tikrina terminalo apklausa.
- Klientas, kuris nepriima simbolių režimo, nepalaikomas:
  jis siunčia visą eilutę po Enter.
- Telnet analizatorius - savas kodas, apie 200 eilučių. Go bibliotekos nesuteikia serveriui
  NAWS ir TTYPE su ribotais buferiais.
