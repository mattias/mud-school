# Lektion 08 – Rum

En MUD är en värld av **rum** förbundna med **utgångar**. Matematiskt är det
en *graf*: rummen är noder, utgångarna är riktade kanter. I den här lektionen
får spelarna en plats i världen, kan `look`:a sig omkring och gå
`north`/`south`/... – och `say` hörs bara i samma rum.

```
              [Gästrummet]
                   | upp/ner
 [Templet] — [Torget] — [Smedjan]        ← byn Mudby
      ↑         |  ↑
   väster   [Stadsporten]
```

## Mål

- Arbeta med **pekare**: alla spelare i samma rum pekar på *samma* `*Room`.
- Skapa structs med struct-literals: `&Room{ID: "torg", ...}`.
- Skriva en **variadisk** funktion: `NewWorld(start string, rooms ...*Room)`.
- Använda **closures** för att skapa kommandon: `moveCommand("north")`.
- Bygga text effektivt med `strings.Builder`.

## Go-koncept

### Pekare

```go
r := &Room{Name: "Torget"} // r är en *Room – en pekare
p.Room = r                 // p och r pekar på samma rum
p.Room == other.Room       // jämför ADRESSER: står de i samma rum?
```

### Closures som "kommer ihåg"

```go
func moveCommand(dir string) CommandFunc {
    return func(s *Server, p *Player, cmd Command) {
        // dir finns kvar här, långt efter att moveCommand returnerat
    }
}

for _, dir := range Directions {
    s.Register(dir, moveCommand(dir))   // sex kommandon, en funktion
}
```

## Uppgift

1. **`world.go`** (ny fil): skriv `NewWorld` och `World.Room`.
   **`defaultworld.go`** (ny fil, färdig) innehåller `DefaultWorld` – byn
   Mudby. Lägg gärna till fler rum!
2. **`player.go`**: ge `Player` ett fält `Room *Room`. Det skyddas av `s.mu`
   (läs och skriv det bara när du håller låset).
3. **`server.go`**:
   - ge `Server` ett fält `world *World`,
   - `NewServerWithWorld(w *World) *Server` – och låt `NewServer()` anropa
     den med `DefaultWorld()`,
   - placera nya spelare i startrummet, och kör `look` direkt efter
     `MsgWelcomeName`.
4. **Kommandon** (lägg dem gärna i en ny fil `rooms.go`):
   - `look` skriver rummets namn, beskrivning, utgångar och vilka andra som
     är där:

     ```
     Torget                     ← Name
     Ett torg.                  ← Description
     Utgångar: norr, öster      ← MsgExits
     Här finns: Bertil          ← MsgHere
     ```

     Utgångarna i ordningen i `Directions`, med namnen i `directionNames`
     (i `messages.go`) och åtskilda med `MsgListSep`. Saknas utgångar:
     `MsgNoExits`. Raden `MsgHere` (andra spelare, sorterade) skrivs bara om
     någon annan är där.
   - `north`, `east`, `south`, `west`, `up`, `down`: finns ingen utgång,
     skriv `MsgCantGo`. Annars: de i gamla rummet ser `MsgLeaves` med
     `directionTravel[dir]` (`Anna går norrut.`), de i nya ser `MsgArrives`
     (`Anna kommer in.`), och Anna själv får det nya rummets `look`.
   - `say` hörs nu bara i samma rum. Nytt kommando `shout` hörs överallt:
     `MsgYouShout` / `MsgShouts` (`Du ropar: hej` / `Anna ropar: hej`).

## En konvention för lås

När allt fler funktioner behöver `s.mu` blir det lätt att råka låsa två
gånger (Gos mutex är **inte** reentrant – det blir en deadlock). Ett vanligt
knep är att låta funktioner som *förutsätter* att låset redan är taget sluta
på `Locked`:

```go
func cmdLook(s *Server, p *Player, cmd Command) {
    s.mu.Lock()
    defer s.mu.Unlock()
    p.Send(s.describeRoomLocked(p))   // låser INTE själv
}
```

## Testa själv

Logga in med två klienter, gå runt, prata, ropa.

## Fundera på

- Varför är `Room.Exits` en `map[string]string` (riktning → ID) och inte
  `map[string]*Room`? (Tips: tänk på nästa lektion, när världen ska läsas
  från en fil.)
- Utgångar är enkelriktade. Hur skulle du upptäcka en utgång `north` som
  saknar en matchande `south` tillbaka?

## Utmaningar

- `go <riktning>` och svenska riktningar (`norr`, `n`, ...).
- `exits` som visar vart varje utgång leder: `norr – Krogen`.
