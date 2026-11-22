# 1. Programa

Tau reikia Go 1.27.3 ar naujesnės versijos. Velour BBS pirminio kodo saugykloje paleisk:

```
go build ./cmd/velourbbs
sudo install velourbbs /usr/local/bin/
```

Pirmoji komanda įrašo failą `velourbbs`. Antroji komanda nukopijuoja jį
į `/usr/local/bin/velourbbs`.

Kompiliuok git kopijoje, kurioje yra laidos žyma ir nėra pakeitimų, pavyzdžiui, po `git checkout v1.2.0`.
Tada laukas `bbs.version` rodo žymą. Kiekvienas kitas kompiliavimas duoda `dev`.
Visos taisyklės: [session-fields](../../../specs/session-fields.md), 3.4 skyrius.

## Rezultatas

`velourbbs` be argumentų išveda naudojimo aprašą:

```
usage:
  velourbbs art import <input> <output>
  velourbbs check [--size <cols>x<rows>]...
  velourbbs fields [--sample]
  velourbbs preview <name> [--size <cols>x<rows>] [--sample <set>] [--set <path>=<value>]... [--keys <keys>] [--view]
  velourbbs serve
```
