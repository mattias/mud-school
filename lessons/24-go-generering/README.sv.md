# Lektion 24 – Procedurgenerering

Världen har 21 rum, och varje zon är en rektangel med väggar. Att rita 21
zoner för hand är tråkigt; att rita 600 är omöjligt. Spel som *Caves of Qud*,
*Brogue* och *Dwarf Fortress* genererar sina kartor med små algoritmer – och
från ett **frö** (seed), så att samma värld kan återskapas exakt.

```
###############.###############     TTTTTTTTTTTTTTT.TTTTTTTTTTTTTTT
###############.###############     T"""""""""TT""""""""""""""TTTTT
#########....##..#....#####.###     T"""""""""""T"""""""""""""TTTTT
########......##.......##....##     T"""""""""T"""""""""""""""TTTTT
###..##......................##     .""""""""TTTT"TTT""""""""TTTTTT
##...........................##     T""""""T""TTTTTTTTT"""""TTTTTTT
###.........###>.......###..###     T"""""""""TTTTT""""""""""TTTTTT
        en grotta (cellautomat)              en skog (brus)
```

## Mål

- Implementera tre klassiska algoritmer: **cellautomater**, **värdebrus**
  (value noise) och **den fulle mannens vandring** (drunkard's walk).
- Använda en slumpgenerator med frö så att allt kan återskapas – och testas.
- Garantera en egenskap hos genererat innehåll (*man kan alltid gå från en
  utgång till en annan*) med en flödesfyllnad (flood fill).

## Go-koncept och algoritmer

### Frön

```go
rng := rand.New(rand.NewPCG(seed, hashString(r.ID)))
```

Varje rum får en egen ström av slumptal, bestämd av världens frö *och*
rummets ID (hashat med `hash/fnv`). Byt frö så ändras hela världen; behåll
det så är världen identisk – på varje dator.

### Cellautomater (grottor)

Fyll zonen med brus – 45 % vägg – och tillämpa en enkel regel några gånger:
*en ruta blir vägg om minst 5 av de 9 rutorna runt den (den själv inräknad)
är väggar*. Ensamma väggar försvinner, ensamma hål sluts, och det som blir
kvar ser ut som grottor.

```
före:       efter ett steg:
#####       #####
#.#.#       ##.##
#...#   ->  ##.##
#.#.#       ##.##
#####       #####
```

### Värdebrus (skogar, fält, sjöar)

Slumptal per ruta ser ut som myrornas krig. **Brus** är slump som varierar
*mjukt*: slumpvärden på ett grovt rutnät (gittret, *lattice*), utjämnade
mellan punkterna. Närliggande punkter får närliggande värden, så träd växer i
dungar och sjöar får stränder.

```go
// a, b, c, d: gittervärdena runt (x, y); tx, ty: smooth(bråkdelen)
top := a + (b-a)*tx
bottom := c + (d-c)*tx
return top + (bottom-top)*ty
```

*Fraktalt* brus lägger ihop flera lager (oktaver): stora former plus finare
detaljer. Sedan bestämmer en tröskel rutan: `n > 0.5` → ett träd.

### Den fulle mannens vandring (gångar)

Genererade kartor har en otäck vana: hörn man inte når, eller en utgång
inmurad i berget. Lösningen: hitta det största öppna området (flood fill),
och från varje utgång som inte ligger i det, gräv en vinglig gång – 70 % av
stegen mot målet, 30 % vart som helst. Rutor som ändå inte går att nå fylls
igen, så att ingen någonsin placeras i en instängd ficka.

## Kontraktet

- `GenerateZone(r *Room, w, h int, seed uint64) *Zone` – efter terräng:
  `cave` (vägg/golv) och `mountain` (berg/golv): cellautomater; `forest`,
  `field`, `water`: brus; `town`, `inside`, `road`: ett enkelt golv. Sedan en
  ram av terrängens kantruta, öppningarna för utgångarna (`openExits`) och
  `connectZone`.
- Samma rum och frö → exakt samma zon; ett annat frö eller rum → en annan.
- I varje genererad zon går **varje gångbar ruta** att nå från varje öppning,
  och kanten är stängd utom vid öppningarna.
- `caStep(grid [][]bool) [][]bool` – 4-5-regeln (rutor utanför rutnätet räknas som väggar).
- `ValueNoise(seed, x, y)` – i `[0,1)`, samma för samma indata, mjukt.
- `reachable(z, start) map[Pos]bool` – flood fill.
- `drunkardWalk(z, from, to, rng, floor)` – förbinder `from` och `to` och rör
  aldrig kanten.
- Världsfilen får `"seed": 42`; rum utan handgjord zon använder
  `GenerateZone` med världens frö och zonstorlek.

## Uppgift

`gen.go` är given utom `caStep`, `ValueNoise`, `FractalNoise`, `reachable`,
`drunkardWalk` och `connectZone`. Lägg sedan till `Seed` i `World` och
`worldFile` och låt `makeZone` anropa `GenerateZone`.

Gå omkring: `./mudschool run server`, `./mudschool connect`, `go west` några
gånger, `zone`. Prova `-world` med en egen världsfil med ett annat frö.

## Fundera på

- Varför blandas rummets ID in i fröet? Vad skulle hända utan det?
- Testerna genererar 160 zoner och kontrollerar att de hänger ihop. Varför är
  det ett bättre test än att jämföra med en bild av den "rätta" zonen?
- Vad händer med världen om du ändrar `caStep` en liten aning? Med spelarnas sparade positioner?
- `connectZone` fyller igen rutor man inte når. Kan det någonsin ta bort den
  enda vägen till något? Varför inte?

## Utmaningar

- Rum och korridorer (BSP) för `inside` och `town`: dela zonen i två, om och
  om igen, lägg ett rum i varje del, förbind dem.
- Riktigt **simplex**-brus eller Perlins gradientbrus i stället för
  värdebrus – jämför hur de ser ut.
- Olika frön för olika delar av världen, så att en "gruva" kan genereras om
  varje dag medan staden förblir densamma.
