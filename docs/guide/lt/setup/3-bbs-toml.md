# 3. `bbs.toml`

`bbs.toml` laiko keturias lenteles: `[bbs]` turi BBS pavadinimą,
`[telnet]` turi telnet serverio nustatymus, `[limits]` turi ryšių ribas ir laiko limitus,
o `[store]` turi duomenų failą.
Kiekvienas raktas neprivalomas. Nenustatytas raktas turi numatytąją reikšmę. Tuščias failas duoda numatytąsias reikšmes.

## Pavadinimas

`[bbs]` raktas `name` yra BBS pavadinimas. Ekranas jį rodo lauke `bbs.name`.
Numatytoji reikšmė – `"Kind Brothers BBS"`.

## Adresas

`[telnet]` raktas `listen` yra serverio TCP adresas: `<host>:<port>`. Numatytoji reikšmė – `":2323"`.

- Tuščias mazgas: serveris klausosi visais savo adresais: `":2323"`.
- Vienas adresas: serveris klausosi tik šiuo adresu, pavyzdžiui, `"192.0.2.1:2323"`.
  Naudok jį, kai serveris turi daugiau nei vieną adresą, o BBS turi naudoti tik vieną.

`velourbbs check` tikrina tik `listen` formą ir prievadą.
`velourbbs serve` tikrina adresą paleidimo metu. Jei serveris neturi
šio adreso, `velourbbs serve` sustoja.

## Ryšių ribos

Šie raktai yra lentelėje `[limits]`.

| Raktas | Numatytoji reikšmė | Ką jis daro | Kada jį keisti |
|---|---|---|---|
| `max_conns` | 64 | Didžiausias ryšių skaičius. | Vienu metu jungiasi daugiau naudotojų. Tada pakeisk ir 5 skyriaus ribas. |
| `max_conns_per_addr` | 4 | Didžiausias ryšių iš vieno adreso skaičius. | Daug naudotojų dalijasi vienu adresu, pavyzdžiui, už NAT. |
| `max_connects_per_min` | 10 | Didžiausias prisijungimų iš vieno adreso skaičius per 60 s. | Klientas daug kartų jungiasi iš naujo. |
| `allow` | `[]` | Adresai ir prefiksai be adreso ribų. | Vienas adresas turi turėti daugiau ryšių, pavyzdžiui, žinomas NAT. |

- „Vienas adresas“ yra vienas IPv4 adresas arba vienas IPv6 /64 prefiksas.
- `max_conns` taikomas ir `allow` adresams.
- Ryšys virš ribos gauna eilutę `Too many connections. Try again later.`
  Tada serveris jį uždaro.

Pavyzdys:

```toml
[limits]
allow = ["192.0.2.10", "2001:db8::/48"]
```

## PROXY protokolas

`[telnet]` raktas `proxy` laiko tarpinių serverių, kurie siunčia PROXY protokolo antraštę, adresus ir prefiksus.
Numatytoji reikšmė – `[]`: serveris neskaito jokios antraštės.
Nustatyk jį, kai tarpinis serveris, pavyzdžiui, HAProxy, siunčia ryšius į BBS.
Tada žurnalas ir ribos naudoja kliento adresą, o ne tarpinio serverio adresą.

- Serveris skaito antraštę tik iš `proxy` adresų.
- Ryšys iš `proxy` adreso be tinkamos antraštės užsidaro be išvesties.
  Todėl, jei `proxy` laiko `127.0.0.1`, telnet klientas tame pačiame serveryje negali prisijungti
  prie `127.0.0.1`.
- Tarpinis serveris turi siųsti antraštę prisijungimo metu. AWS NLB pagal numatymą laukia
  kliento duomenų. Tada tai neveikia.

Pavyzdys:

```toml
[telnet]
proxy = ["127.0.0.1"]
```

## Laiko limitai

`probe_timeout` yra lentelėje `[telnet]`. `idle_timeout` ir `output_timeout` yra lentelėje `[limits]`.

| Raktas | Numatytoji reikšmė | Ką jis daro | Kada jį keisti |
|---|---|---|---|
| `idle_timeout` | 600 | Sekundės be klavišo, kol serveris atjungia. | Naudotojai skaito ilgus ekranus. |
| `probe_timeout` | 5 | Sekundės nuo prisijungimo, kol serveris sužino terminalą. | Klientų ryšys lėtas. |
| `output_timeout` | 60 | Sekundės, per kurias rašymas turi baigtis. Kitaip serveris atjungia. | Klientų ryšys lėtas. |

## Duomenų failas

`[store]` raktas `path` yra duomenų failas. BBS laiko jame savo duomenis tarp seansų.
Numatytoji reikšmė – `"velourbbs.jsonl"`. Santykinis kelias yra santykinis BBS katalogo atžvilgiu.

`velourbbs serve` sukuria failą. Failo katalogas turi egzistuoti.
Palik numatytąją reikšmę: tada failas yra BBS kataloge, ir tavo naudotojas gali į jį rašyti.

## Failo pavyzdys

Šis failas turi numatytąsias reikšmes. Jis nustato kiekvieną raktą, išskyrus `allow` ir `proxy`:

```toml
[bbs]
name = "Kind Brothers BBS"

[telnet]
listen = ":2323"
probe_timeout = 5

[limits]
max_conns = 64
max_conns_per_addr = 4
max_connects_per_min = 10
idle_timeout = 600
output_timeout = 60

[store]
path = "velourbbs.jsonl"
```

Įrašyk jį į `/srv/velourbbs/bbs.toml`. Tada pakeisk reikšmes, kurių tau reikia.
Po kiekvieno pakeitimo paleisk `velourbbs check`.

Visos taisyklės: [bbs-toml](../../../specs/bbs-toml.md), 3 skyrius,
ir [data-store](../../../specs/data-store.md).

## Rezultatas

`velourbbs check` nieko neišveda.
