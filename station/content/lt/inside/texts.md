# Tekstai ir menas

## Ką daro

Velour rodo Markdown tekstus ir ANSI meną. Tekstus ji skaito tokius, kokie yra.
Meną sistemos operatorius vieną kartą paverčia serverio formatu komanda `velourbbs art import`.

## Kaip: tekstai

Markdown skaito biblioteka `goldmark`: CommonMark 0.31.2
([specifikacija](https://spec.commonmark.org/0.31.2/)) ir GFM lenteles
([GFM lentelės](https://github.github.com/gfm/#tables-extension-)).
Ji duoda dokumento medį be HTML, o Velour išdėsto jį langeliais:

- antraštės, pastraipos, sąrašai, citatos, kodas ir lentelės - kiekvienas savo temos stiliumi;
- ilga pastraipa keliama žodžiais į stulpelio plotį, kodo eilutė nukerpama ženklu `...`;
- nuoroda su adresu gauna išnašos numerį `[1]`, o adresai eina sąrašu teksto pabaigoje.

Tekstas dalijamas į puslapius pagal terminalo aukštį: `Tarpas` verčia pirmyn, `B` - atgal.
Skaitymo laikas - žodžių skaičius, padalytas iš 200.

Vadovai yra saugyklos `docs/guide/`. Stotis mato juos per simbolinę nuorodą
ir jų nekopijuoja. Todėl vadovas stotyje ir GitHub - tas pats failas.

## Kaip: menas

ANSI meną piešia Moebius 1.0.29, icy_draw 0.6.2 ir PabloDraw 3.3.14-beta.
Paprastai jie rašo `.ans` failą CP437 koduote su SAUCE įrašu pabaigoje: tai 128 baitai
su pločiu, autoriumi ir vėliavėlėmis ([SAUCE](https://www.acid.org/info/sauce/sauce.htm)).

`velourbbs art import` atkuria failą į langelių tinklelį, kaip tai padarytų terminalas:
žymeklio judesiai, perkėlimai, trynimai. Plotį ji ima iš SAUCE įrašo `TInfo1`,
be jo - 80. `TFlags` 0 bitas įjungia iCE spalvas: ryškų foną vietoj mirksėjimo.
Tada komanda įrašo tinklelį plokščiu pavidalu: BOM, eilutės UTF-8 ir tarp jų tik SGR stiliai.

Serveris skaito tik plokščią meną UTF-8. Jis niekada nevykdo žymeklio judesių
ir niekada neskaito CP437. Ekranas įterpia meną tarp teksto: `{>logo.ans}`.

## Ką gauna

- Tekstą, kuris atrodo vienodai esant 80 ir 132 stulpeliams ir bet kokiam spalvų gyliui.
- Meną, kurį serveris išveda kaip įprastus langelius: jis gali stovėti šalia teksto,
  o efektai gali atgaivinti jo žvaigždes ir žvakes.
- VGA meno spalvas Velour siunčia RGB, jei terminalo paletė ne VGA.
  Taip menas išlaiko dailininko spalvas.

## Ribos

- Nuorodos į kitus Markdown failus kol kas neatsidaro.
- Teksto front matter - tik TOML. YAML yra klaida.
- CP437 glifus LF, CR, 0x1A ir ESC vietose redaktoriai rašo kaip valdymo simbolius.
  Importas negali jų atkurti.
- Moebius "Export As UTF-8" ir icy_draw savi šriftai nepalaikomi.
- Animacijos nėra: tinklelis yra paskutinis kadras. Importas įspėja apie pirmą langelio perrašymą.
- Atkūrimas rašo ne daugiau kaip 4080000 langelių: 16 tinklelių po 255 stulpelius ir 1000 eilučių.
