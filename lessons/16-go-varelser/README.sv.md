# Lektion 16 – Varelser

Mudby vaknar till liv: en svart katt stryker runt på torget, krögaren Greta
pratar på krogen, en kråka hoppar längs skogsbrynet – och någonstans i
Mörkskogen finns en varg. Varelser (*NPC:er*, "non-player characters") agerar
på egen hand, så de lever i spelloopens `onTick`.

Vad en varelse gör beskrivs av **beteenden**: små värden som implementerar ett
gränssnitt. Världsfilen säger vilka varje varelse har.

```json
{"id": "cat", "name": "en svart katt", "room": "torget",
 "behaviors": [{"type": "wander", "chance": 0.03},
               {"type": "talk", "chance": 0.02, "lines": ["Mjau."]}]}
```

## Mål

- Designa med ett litet gränssnitt (`Behavior`) och flera implementationer.
- Bygga värden från data: en "fabrik" som gör en `BehaviorSpec` till ett `Behavior`.
- Göra slump testbar: injicera en `*rand.Rand` med fast frö (seed).
- Återanvända den generiska `FindMatch` från lektion 15 för en ny typ.

## Go-koncept

### Ett gränssnitt med flera implementationer

```go
type Behavior interface {
    Act(s *Server, n *NPC)
}

type Wanderer struct{ Chance float64 }
func (b Wanderer) Act(s *Server, n *NPC) { ... }

type Talker struct{ Chance float64; Lines []string }
func (b Talker) Act(s *Server, n *NPC) { ... }
```

Ett test kan lägga till ett eget `Behavior` – titta på `countingBehavior` i
`mudschool_16_test.go`. Det är styrkan med gränssnitt: loopen vet inte och
bryr sig inte om vad ett beteende gör.

### Slump som går att testa

```go
rng := rand.New(rand.NewPCG(1, 2))  // math/rand/v2, ett fast frö
rng.Float64()                        // samma följd varje körning
rng.IntN(3)                          // 0, 1 eller 2
```

Servern tar `Options.Rand`. I drift används ett slumpat frö; testerna
använder ett fast. För att samma frö ska ge samma resultat måste allt ske i
en fast ordning – därför sorterar `npcsByID`, och `Wanderer` går igenom
`Directions` i ordning istället för att loopa över en map.

## Kontraktet

- `type NPC struct` med JSON-fälten `id`, `name`, `description`, `keywords`,
  `room`, `behaviors`, och (för lektion 18, oanvända än så länge) `hp`,
  `damage`, `attackable`. Under körning: `Location *Room`.
- `func (n *NPC) Matches(word string) bool` och `func (n *NPC) AddBehavior(b Behavior)`.
- `type Behavior interface { Act(s *Server, n *NPC) }`, `Wanderer`, `Talker`,
  `BehaviorSpec{Type, Chance, Lines}` med typerna `"wander"` och `"talk"`.
- `Room.NPCs []*NPC`, `World.NPCs map[string]*NPC`,
  `(w *World) AddNPCs(...)` och `(w *World) LoadNPCs(r)` (`{"npcs": [...]}`),
  med `MsgErrNPCFields`, `MsgErrDuplicateNPC`, `MsgErrNPCRoom`,
  `MsgErrBehavior`. En världsfil får innehålla `"npcs"`; `DefaultWorld` läser
  den nya `npcs.json`.
- `Options.Rand *rand.Rand` (`math/rand/v2`; nil = slumpat frö).
- Varje tick körs alla varelsers beteenden (i ID-ordning).
- `look` listar varelser efter spelarna i `MsgHere`. `examine` fungerar på
  varelser. Varelser som kommer och går använder `MsgLeaves`/`MsgArrives` med
  kind `move`; när de pratar används `MsgSays` med kind `say`.

## Uppgift

1. **`npc.go`** (ny): fyll i `Matches`, `Wanderer.Act`, `Talker.Act`,
   `moveNPC` och `AddNPCs`.
2. **`world.go`**/**`defaultworld.go`**: de nya fälten, `"npcs"` i
   världsfilen, `npcs.json` i standardvärlden.
3. **`server.go`**/**`loop.go`**: `Options.Rand`, `s.rng`, och `onTick` kör
   alla beteenden.
4. **`rooms.go`**/**`items.go`**: varelser i `look` och i `examine`.

## Testa själv

```
./mudschool run server -tick 20ms
```

Ett snabbt tick gör varelserna rastlösa. Stå på torget och vänta på katten.
Besök krogen.

## Fundera på

- Varför får `Wanderer.Act` servern och varelsen som argument, istället för
  att spara dem i `Wanderer`?
- Vad skulle hända med testerna med fast frö om `onTick` loopade direkt över
  mappen `s.world.NPCs`?
- Ska en varelse få vandra vart som helst? Hur skulle du hålla krögaren kvar på krogen?

## Utmaningar

- Ett `follow`-beteende: kråkan följer närmaste spelare.
- Dag och natt: vargen vandrar bara på natten (använd `GameTime`).
- `say` till en varelse: låt krögaren svara när någon säger "öl".
