# Lektion 23 – Zoner

Fas 4 börjar: från en graf av rum till en värld av rutor à la *Caves of Qud*.
Varje rum blir en **zon** – ett rutnät du går omkring i, ruta för ruta. Gå ut
genom en öppning i kanten så är du i nästa rums zon. Rumskartan från lektion
10 lever vidare som *världskarta*.

```
####.####        # vägg    . golv     ~ vatten
#.......#        @ du      * en sak   c en varelse
#..~*@..#
#.......#        "n" tar nu ETT steg norrut.
#########        "go north" hoppar fortfarande till nästa rum.
```

En zon skickas till grafiska klienter **en gång**. Därefter skickar servern
bara det som *ändrats* – en `delta`-frame – eftersom en zon på 31×13 rutor
med alla varelser, skickad tio gånger i sekunden till varje spelare, vore
slöseri.

## Mål

- Modellera en 2D-värld: ett rutnät (`[][]byte`), positioner, grannar, kanter.
- Hålla spellogik och positioner i synk överallt där något flyttar sig.
- Räkna ut skillnader: vad som har tillkommit, flyttats och försvunnit sedan sist.
- Utöka ett protokoll med två nya frames utan att gamla klienter går sönder.

## Go-koncept

### Ett rutnät

```go
type Pos struct{ X, Y int }          // X växer österut, Y söderut
type Zone struct {
	W, H  int
	Tiles [][]byte                   // Tiles[y][x] – raden först!
}
func (z *Zone) At(p Pos) byte        // utanför zonen är allt vägg
```

`Pos` är en liten struct med två int – den är **jämförbar**, så den kan vara
nyckel i en map (`map[Pos]bool`) och jämföras med `==`. Det kommer du att
använda mycket i nästa lektioner.

### Lat initiering

Ett rums zon skapas första gången någon frågar efter den:

```go
func (w *World) zoneOf(r *Room) *Zone {
	if r.zone == nil {
		r.zone = ... // ParseZone(r.ZoneRows) eller en genererad
	}
	return r.zone
}
```

### Skillnader

Varje spelare minns vad klienten har fått veta (`p.seen`, en map från
entitetens ID till `Entity`). Efter varje kommando och varje tick:

```go
now := s.entities(p.Room)
// finns i now men inte i seen       -> added
// finns i båda, men olika (!=)      -> moved
// finns i seen men inte i now       -> removed
```

`Entity` är en struct av strängar och int, så `old != e` jämför alla fält.

## Kontraktet

**Rutor** (konstanter i `zone.go`): `.` golv, `#` vägg, `"` gräs, `T` träd,
`~` vatten, `^` berg, `+` dörr, `<` trappa upp, `>` trappa ner. `Walkable`:
golv, gräs, dörr, trappor. `tileMeanings` är förklaringen (legend).

**Världsfilen:**
- ett rum får ha `"zone": ["####.####", ...]` – en handgjord zon (minst 3×3,
  lika långa rader, kända rutor – annars `MsgErrZoneRows`); varje utgång
  behöver en öppning (en gångbar kantruta på den sidan, eller en trappa för
  upp/ner – annars `MsgErrZoneExit`);
- `"zone_size": [w, h]` – storleken på genererade zoner (standard 31×13).

**Rum utan handgjord zon** får en enkel: väggar runt ett golv, en öppning mitt
på varje kant som har en utgång, `<` på `(w/2, h/2-1)` för upp och `>` på
`(w/2, h/2+1)` för ner.

**Positioner:** `Player.Pos`, `NPC.Pos`, `Item.Pos`.
- Nya spelare, spelare som vaknat efter döden och återställda spelare: den
  lediga rutan närmast mitten (`freeNear`).
- Genom en utgång: ett steg innanför öppningen du kom in genom (`arrival`);
  upp och ner: på den andra trappan.
- Saker placeras nära mitten; en sak du släpper hamnar där du står.

**Kommandon:**
- `north`, `south`, `east`, `west`, `northeast`, … (`n`, `ne`, …): ett steg.
  En vägg, ett träd, vatten, … ger `MsgBlocked`. Går du ut ur zonen genom en
  öppning åt ett håll som har en utgång kommer du till nästa rum.
- `up`/`down`: bara på trappan (`MsgNoStairs`).
- `go <riktning>`: direkt till nästa rum, som förut (`MsgGoWhere`, `MsgCantGo`).
- `zone`: zonen igen.

**Frames** (se PROTOCOL.md):
- `zone` – `{"room", "width", "height", "rows", "legend", "entities", "self"}`;
  en entitet är `{"id", "kind", "name", "glyph", "x", "y"}` med ID:n `p:Anna`,
  `n:cat`, `i:lamp`, sorterna `player`/`creature`/`item`, tecknen `@`,
  NPC:ns `glyph` (standard: första bokstaven i ID:t) och sakens `glyph` (standard `*`).
- `delta` – `{"room", "added", "moved", "removed"}`; `removed` är en lista med ID:n.
- Grafiska klienter får en zone-frame vid varje `look` (och alltså i varje
  nytt rum), och en delta efter varje kommando och varje tick om något ändrats.
- Telnet får zonen ritad som text efter varje steg (`renderZone`, given).

## Uppgift

1. **Fälten:** `Room.ZoneRows` (json `zone`) och `Room.zone`;
   `World.ZoneSize` och `worldFile.ZoneSize` (json `zone_size`); `Pos` på
   `Player`, `NPC` och `Item`; `Glyph` (json `glyph,omitempty`) på `NPC` och
   `Item`; `Player.seenRoom` och `Player.seen`; `Frame.Zone` och
   `Frame.Delta`; `case "zone"` i `encodeText`. Nu kompilerar de nya filerna.
2. **`zone.go`**: `ParseZone`, `opening`, `arrival`, `plainZone`,
   `openExits`, `zoneOf`, `validateZone` (anropa den från `Validate`).
3. **`sync.go`**: `entities`, `zoneData`, `sendZone`, `syncAll`, `syncPlayer`.
4. **Förflyttning:** döp om `moveCommand` till `travel(p, dir)` och sätt
   `p.Pos`; lägg till `step`, trappkommandona, `go` och `zone`; registrera
   stegriktningarna och alias `ne`, `nw`, `se`, `sw`. Flykt använder `travel`.
5. **Överallt där något får ett rum ska det också få en position:**
   `addPlayer`, `restore`, `die`, `moveNPC`, `respawnCreatures`, `AddNPCs`,
   `AddItems`, `cmdDrop` och `removePlayer` (hjälpfunktionerna `placeItem`
   och `taken` är bra att ha).
6. **Synk:** `s.syncAll()` efter `dispatch`, i slutet av `onTick`, efter
   inloggning och efter utloggning. `look` skickar `sendZone` till grafiska klienter.

Det är många ställen – kompilatorn och testerna hittar dem du glömmer. Gå
omkring i telnet när du är klar: `./mudschool connect`, `n`, `e`, `zone`.

## Fundera på

- `Tiles[y][x]` – varför raden först? Vad skulle `Tiles[x][y]` betyda för `Rows()`?
- Vad kostar en delta jämfört med en hel zon? Räkna byten för ett steg.
- `syncAll` jämför *alla* entiteter i varje spelares rum efter varje
  kommando. När skulle det bli för långsamt, och vad skulle du göra då?
- Resten av spelet tänker fortfarande i rum (strid, `get`, `say`). Är det ett
  problem? Vad skulle ändras om `get` krävde att man står på saken?

## Utmaningar

- Dörrar: ett `+` som är stängt (inte gångbart) tills du gör `open`.
- Låt varelser ta ett slumpvis steg inne i zonen då och då.
- Visa zonen i `look` för telnetspelare också, bredvid beskrivningen.
