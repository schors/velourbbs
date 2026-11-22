# Prisijungimas

## Ką daro

Velour priima TCP ryšį, nustato svečio adresą ir tikrina ribas.
Tik tada prasideda telnet.

## Kaip

### Prievadas

Linux sistemoje prievadą laiko systemd, ne Velour: tai socket activation.
systemd atidaro prievadą ir paleidžia Velour su kintamaisiais `LISTEN_PID` ir `LISTEN_FDS=1`.
Atviras prievadas ateina kaip deskriptorius 3, pagal `sd_listen_fds` taisykles
([systemd dokumentacija](https://www.freedesktop.org/software/systemd/man/latest/sd_listen_fds.html)).
Be systemd Velour pati klauso adreso `telnet.listen` iš `bbs.toml`.

### Adresas už tarpinio serverio

Jei prieš Velour stovi tarpinis serveris, svečio adresas ateina PROXY protocol antraštėje
([HAProxy specifikacija](https://www.haproxy.org/download/3.1/doc/proxy-protocol.txt)).
Velour skaito antraštę tik iš adresų, esančių `bbs.toml` sąraše `proxy`.

- 1 versija - teksto eilutė, pavyzdžiui `PROXY TCP4 <adresas> <adresas> <prievadas> <prievadas>`,
  ne ilgesnė nei 107 baitai.
- 2 versija - dvejetainė antraštė: 16 baitų ir iki 536 baitų adresų ir TLV.
  TLV Velour praleidžia.

Antraštę siunčia HAProxy, nginx, Traefik ir stunnel.

### Ryšių ribos

- `max_conns` - viso serverio ryšių skaičius.
- `max_conns_per_addr` - atviri ryšiai iš vieno adreso.
- `max_connects_per_min` - prisijungimai iš vieno adreso per 60 sekundžių.
  Velour juos skaičiuoja 1 sekundės žingsniais ir prisimena iki 10000 adresų.

Vienas adresas - tai vienas IPv4 adresas arba vienas IPv6 /64 tinklas.
IPv4 adresas IPv6 viduje (`::ffff:a.b.c.d`) laikomas IPv4 adresu.
Sąrašo `allow` adresams adreso ribos negalioja.

## Ką gauna

- Svečio adresą. Jį mato ribos, žurnalas ir laukas `session.ip` ekrane "Apie stotį".
- Antraštės `UNKNOWN` (1 versija), `LOCAL`, `UNSPEC` ir `UNIX` (2 versija) adreso neduoda.
  Tada Velour ima tarpinio serverio adresą, o adreso ribos negalioja.
- Ryšys virš ribos gauna vieną eilutę `Too many connections. Try again later.`
  ir užsidaro.

Prieš uždarydama ryšį, Velour uždaro savo rašymo pusę ir iki 1 sekundės skaito svečio įvestį.
Jei uždaryti ryšį su neperskaityta įvestimi, TCP siunčia nutraukimą,
ir klientas gali prarasti paskutinį tekstą.

## Ribos

- PROXY antraštė ir terminalo apklausa turi baigtis per `probe_timeout` nuo prisijungimo.
- Neteisinga antraštė arba jos nebuvimas uždaro ryšį be atsakymo.
- Tarpinis serveris turi atsiųsti antraštę iš karto. Velour rašo pirma, todėl tarpinis serveris,
  kuris laukia kliento duomenų, neveikia. Taip pagal nutylėjimą elgiasi AWS NLB.
- Dabar ryšys vyksta tik per telnet be šifravimo. SSH ir telnet per TLS - planuose.
