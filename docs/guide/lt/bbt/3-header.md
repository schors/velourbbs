# 3 užduotis. Galvutė

Kiekvienas „Lempos“ ekranas prasideda tomis pačiomis 5 eilutėmis: tai „galvutė“.
Parašyk jas vieną kartą fragmente ir įterpk fragmentą į kiekvieną ekraną.

- Fragmentas yra `screens/_<name>.bbt`. Jis neturi priekinės dalies.
- Eilutė `{>_<name>}` įterpia fragmentą: BBS įdeda fragmento eilutes į jos vietą.
  Šioje eilutėje negali būti kito teksto.
- Po įterpimo galioja tas pats stilius, kuris buvo prieš įterpimą.

Visos taisyklės: [bbt-format](../../../specs/bbt-format.md), 3.3 ir 3.4 skyriai.

Faile `screens/main.bbt` pakeisk 1–5 eilutes viena eilute:

```
+++
[[item]]
key = "G"
do = "logoff"
+++
{>_head}
       Labas vakaras. Čia tylu ir šilta.
```

Perkelk tas 5 eilutes į fragmentą:

`examples/lamp/lt/screens/_head.bbt`:
```

  {!line}{~─}{!}  

       {!title}Lempa{!}  {!orn}☼ {!dim}BBS prie lempos šviesos{!}

```

## Rezultatas

- `velourbbs check` nieko neišveda.
- `velourbbs preview main` išveda tą patį, ką ir 2 užduotyje. Kad palygintum, paleisk:

```
velourbbs preview main > task3.txt
diff task2.txt task3.txt
```

`diff` nieko neišveda. Sistemoje Windows naudok `fc task2.txt task3.txt`.
